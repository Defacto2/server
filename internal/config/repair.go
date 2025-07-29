package config

// Package file repair.go contains the repair functions for assets and downloads.

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Defacto2/archive/rezip"
	"github.com/Defacto2/helper"
	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/config/fixarc"
	"github.com/Defacto2/server/internal/config/fixarj"
	"github.com/Defacto2/server/internal/config/fixlha"
	"github.com/Defacto2/server/internal/config/fixzip"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"
)

const (
	unid      = "00000000-0000-0000-0000-000000000000" // common universal unique identifier example
	cfid      = "00000000-0000-0000-0000000000000000"  // coldfusion uuid example
	syncthing = ".stfolder"                            // syncthing directory name
)

// Archives checks the download directory for any legacy and obsolete archives.
// Obsolete archives are those that use a legacy compression method that is not supported
// by Go or JS libraries used by the website.
func (c *Config) Archives(ctx context.Context, exec boil.ContextExecutor, sl *slog.Logger) error { //nolint:cyclop,funlen,gocognit
	if err := argspanic(ctx, exec, sl); err != nil {
		return err
	}
	d := time.Now()
	artifacts := []string{}
	var err error
	extra := dir.Directory(c.AbsExtra)
	zipWalker := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%w: %s", err, path)
		}
		uid := fixzip.Check(ctx, path, extra, d, artifacts...)
		if uid == "" || fixzip.Invalid(ctx, path) {
			return nil
		}
		ra := Rearchiving{Source: path, UID: uid, Destination: extra}
		if err := Zip.ReArchive(ctx, sl, ra); err != nil {
			return fmt.Errorf("zip repair and re-archive: %w", err)
		}
		return nil
	}
	lhaWalker := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%w: %s", err, path)
		}
		uid := fixlha.Check(extra, d, artifacts...)
		if uid == "" || fixlha.Invalid(ctx, path) {
			return nil
		}
		ra := Rearchiving{Source: path, UID: uid, Destination: extra}
		if err := LHA.ReArchive(ctx, sl, ra); err != nil {
			return fmt.Errorf("lha/lzh repair and re-archive: %w", err)
		}
		return nil
	}
	arcWalker := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%w: %s", err, path)
		}
		uid := fixarc.Check(ctx, path, extra, d, artifacts...)
		if uid == "" || fixarc.Invalid(ctx, path) {
			return nil
		}
		ra := Rearchiving{Source: path, UID: uid, Destination: extra}
		if err := Arc.ReArchive(ctx, sl, ra); err != nil {
			return fmt.Errorf("arc repair and re-archive: %w", err)
		}
		return nil
	}
	arjWalker := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%w: %s", err, path)
		}
		uid := fixarj.Check(extra, d, artifacts...)
		if uid == "" || fixarj.Invalid(ctx, path) {
			return nil
		}
		ra := Rearchiving{Source: path, UID: uid, Destination: extra}
		if err := Arj.ReArchive(ctx, sl, ra); err != nil {
			return fmt.Errorf("arj repair and re-archive: %w", err)
		}
		return nil
	}

	download := dir.Directory(c.AbsDownload.String())
	for repair := range slices.Values(repairs()) {
		if err := repair.lookPath(); err != nil {
			sl.Error("archives "+repair.String(), slog.Any("error", err))
			continue
		}
		artifacts, err = repair.artifacts(ctx, exec, sl)
		if err != nil {
			sl.Error("archives "+repair.String(), slog.Any("error", err))
			continue
		}
		switch repair {
		case Zip:
			err = filepath.WalkDir(download.Path(), zipWalker)
		case LHA:
			err = filepath.WalkDir(download.Path(), lhaWalker)
		case Arc:
			err = filepath.WalkDir(download.Path(), arcWalker)
		case Arj:
			err = filepath.WalkDir(download.Path(), arjWalker)
		}
		if err != nil {
			sl.Error("archives walk directory",
				slog.Any("error", err),
				slog.String("path", download.Path()))
		}
	}
	sl.Info("completed uuid archive checks",
		slog.Float64("seconds taken", time.Since(d).Seconds()))
	return nil
}

// Repair is a type of archive for the re-archive and recompress methods.
type Repair int

const (
	Zip Repair = iota // ZIP and PKZip archives
	LHA               // LHA and LZH archives
	Arc               // ARC archives
	Arj               // ARJ archives
)

func repairs() []Repair {
	return []Repair{Zip, LHA, Arc, Arj}
}

func (r Repair) String() string {
	return [...]string{"zip", "lha", "arc", "arj"}[r]
}

// Rearchiving are the source and destination arguments required
// by the ReArchive Repair method.
type Rearchiving struct {
	Source      string        // Source is the file extracted to a temporary directory and re-compressed.
	UID         string        // UID is the destination filename using a universal unique ID naming syntax.
	Destination dir.Directory // Destination is the directory to save the re-compressed file.
}

// ReArchive the file using the specified compression method.
// The original ra.Source file is not removed.
func (r Repair) ReArchive(ctx context.Context, sl *slog.Logger, ra Rearchiving) error {
	if sl == nil {
		return ErrNoSlog
	}
	if ra.Source == "" || ra.UID == "" {
		return fmt.Errorf("rearchive %s %w: %q %q", r, ErrNoPath, ra.Source, ra.UID)
	}
	if err := ra.Destination.IsDir(); err != nil {
		return fmt.Errorf("rearchive %s %w: %q", r, err, ra.Destination)
	}
	tmp, err := os.MkdirTemp(helper.TmpDir(), "rearchive-")
	if err != nil {
		return fmt.Errorf("rearchive mkdir temp %w: %s", err, ra.Source)
	}
	defer func() {
		err := os.RemoveAll(tmp)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	}()
	extractCmd, extractArg := "", ""
	switch r {
	case Zip:
		extractCmd, extractArg = command.HWZip, "extract"
	case LHA:
		extractCmd, extractArg = command.Lha, "xf"
	case Arc:
		extractCmd, extractArg = command.Arc, "x"
	case Arj:
		extractCmd, extractArg = command.Zip7, "x"
	}
	ctx1min, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx1min, extractCmd, extractArg, ra.Source)
	cmd.Dir = tmp
	if stdoutStderr, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("rearchive run %w: %s: dump: %q",
			err, ra.Source, stdoutStderr)
	}
	c, err := helper.Count(tmp)
	if err != nil {
		return fmt.Errorf("rearchive tmp count %w: %s", err, tmp)
	}
	sl.Info("rearchive",
		slog.String("count", fmt.Sprintf("rezipped %d files for %s", c, ra.UID)),
		slog.String("tmp", tmp))
	_, err = os.Stat(tmp)
	if err != nil {
		return fmt.Errorf("rearchive tmp stat %w: %s", err, tmp)
	}
	basename := ra.UID + ".zip"
	tmpArc := filepath.Join(helper.TmpDir(), basename)
	if written, err := rezip.CompressDir(tmp, tmpArc); err != nil {
		return fmt.Errorf("rearchive dir %w: %s", err, tmp)
	} else if written == 0 {
		return nil
	}
	finalArc := ra.Destination.Join(basename)
	if err = helper.RenameCrossDevice(tmpArc, finalArc); err != nil {
		defer func() {
			if err := os.RemoveAll(tmpArc); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
			}
		}()
		return fmt.Errorf("rearchive rename %w: %s", err, tmpArc)
	}
	st, err := os.Stat(finalArc)
	if err != nil {
		return fmt.Errorf("rearchive zip stat %w: %s", err, finalArc)
	}
	sl.Info("rearchive",
		slog.String("new", "extra deflated zipfile created"),
		slog.Int("size as bytes", int(st.Size())),
		slog.String("path", finalArc))
	return nil
}

func (r Repair) lookPath() error {
	switch r {
	case Zip:
		if _, err := exec.LookPath(command.HWZip); err != nil {
			return fmt.Errorf("cannot find hwzip executable: %w", err)
		}
	case LHA:
		if _, err := exec.LookPath(command.Lha); err != nil {
			return fmt.Errorf("cannot find lha executable: %w", err)
		}
	case Arc:
		if _, err := exec.LookPath(command.Arc); err != nil {
			return fmt.Errorf("cannot find arc executable: %w", err)
		}
	case Arj:
		if _, err := exec.LookPath(command.Zip7); err != nil {
			return fmt.Errorf("cannot find arj executable: %w", err)
		}
	default:
	}
	return nil
}

func (r Repair) artifacts(ctx context.Context, exec boil.ContextExecutor, sl *slog.Logger) ([]string, error) {
	var files models.FileSlice
	var err error
	switch r {
	case Zip:
		files, err = fixzip.Files(ctx, exec)
	case LHA:
		files, err = fixlha.Files(ctx, exec)
	case Arc:
		files, err = fixarc.Files(ctx, exec)
	case Arj:
		files, err = fixarj.Files(ctx, exec)
	}
	if err != nil {
		return nil, fmt.Errorf("artifacts %s files, %w", r.String(), err)
	}
	size := len(files)
	s := fmt.Sprintf("%s %s", tags.DOS, r)
	sl.Info("repair",
		slog.String("archives", s),
		slog.Int("count", size))
	artifacts := make([]string, size)
	for i, f := range files {
		if !f.UUID.Valid || f.UUID.String == "" {
			continue
		}
		artifacts[i] = f.UUID.String
	}
	artifacts = slices.Clip(artifacts)
	slices.Sort(artifacts)
	return artifacts, nil
}

// Assets on startup checks the file system directories for any invalid or unknown files.
// These specifically match the base filename against the UUID column in the database.
// When there is no matching UUID, the file is considered orphaned and these are moved
// to the orphaned directory without warning.
//
// There are no checks on the 3 directories that get scanned.
func (c *Config) Assets(ctx context.Context, exec boil.ContextExecutor, sl *slog.Logger) error {
	if err := argspanic(ctx, exec, sl); err != nil {
		return err
	}
	d := time.Now()
	mods := []qm.QueryMod{}
	mods = append(mods, qm.Select("uuid"))
	mods = append(mods, qm.WithDeleted())
	files, err := models.Files(mods...).All(ctx, exec)
	if err != nil {
		return fmt.Errorf("config repair select all uuids: %w", err)
	}
	size := len(files)
	sl.Info("assets", slog.Int("checking uuid count", size))
	artifacts := make([]string, size)
	for i, f := range files {
		if !f.UUID.Valid || f.UUID.String == "" {
			continue
		}
		artifacts[i] = f.UUID.String
	}
	artifacts = slices.Clip(artifacts)
	slices.Sort(artifacts)

	dirs := []string{string(c.AbsDownload.String()), string(c.AbsPreview.String()), c.AbsThumbnail.String()}
	counters := make([]int, len(dirs))
	var wg sync.WaitGroup
	wg.Add(len(dirs))
	orphaned := dir.Directory(c.AbsOrphaned)
	for i, dir := range dirs {
		go func(dir string) {
			defer wg.Done()
			err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return fmt.Errorf("walk path %w: %s", err, path)
				}
				if d.IsDir() {
					return nil
				}
				counters[i]++
				uid := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
				if _, found := slices.BinarySearch(artifacts, uid); !found {
					unknownAsset(sl, path, d.Name(), uid, orphaned)
				}
				return nil
			})
			if err != nil {
				sl.Error("assets", slog.String("walk directory", dir), slog.Any("error", err))
			}
		}(dir)
	}

	wg.Wait()
	sum := 0
	for val := range slices.Values(counters) {
		sum += val
	}
	sl.Info("assets", slog.String("done", "checked files for uuids"),
		slog.Int("files checked", sum), slog.Int("uuids", size), slog.Duration("time taken", time.Since(d)))
	return nil
}

// unknownAsset logs a warning message for an unknown asset file.
func unknownAsset(sl *slog.Logger, oldpath, name, uid string, orphaned dir.Directory) {
	sl.Warn("unknown file",
		slog.String("issue", "no matching artifact in the database for the found file"),
		slog.String("uuid", uid), slog.String("filename", name))
	defer func() {
		now := time.Now().Format("2006-01-02_15-04-05")
		dest := orphaned.Join(fmt.Sprintf("%s_%s", now, name))
		if err := helper.RenameCrossDevice(oldpath, dest); err != nil {
			sl.Error("unknown file",
				slog.String("issue", "could not move the file to the orphaned directory"),
				slog.String("source path", oldpath), slog.String("destination path", dest),
				slog.Any("error", err))
		}
	}()
}

// RepairAssets on startup check the file system directories for any invalid or unknown files.
// If any are found, they are removed without warning.
func (c *Config) RepairAssets(ctx context.Context, exec boil.ContextExecutor, sl *slog.Logger) error {
	if err := argspanic(ctx, exec, sl); err != nil {
		return err
	}
	backup := dir.Directory(c.AbsOrphaned)
	if st, err := os.Stat(backup.Path()); err != nil {
		return fmt.Errorf("repair backup directory %w: %s", err, backup.Path())
	} else if !st.IsDir() {
		return fmt.Errorf("repair backup directory %w: %s", ErrNotDir, backup.Path())
	}
	if err := c.ImageDirs(sl); err != nil {
		return fmt.Errorf("repair the images directories %w", err)
	}
	src := dir.Directory(c.AbsDownload)
	extra := dir.Directory(c.AbsExtra)
	if err := DownloadDir(sl, src, backup, extra); err != nil {
		return fmt.Errorf("repair the download directory %w", err)
	}
	if err := c.Assets(ctx, exec, sl); err != nil {
		return fmt.Errorf("repair assets %w", err)
	}
	if err := c.Archives(ctx, exec, sl); err != nil {
		return fmt.Errorf("repair archives %w", err)
	}
	if err := c.Previews(ctx, exec, sl); err != nil {
		return fmt.Errorf("repair previews %w", err)
	}
	if err := c.MagicNumbers(ctx, exec, sl); err != nil {
		return fmt.Errorf("repair magics %w", err)
	}
	if err := c.TextFiles(ctx, exec, sl); err != nil {
		return fmt.Errorf("repair textfiles %w", err)
	}
	return nil
}

// TextFiles on startup check the extra directory for any readme text files that are duplicates of the diz text files.
func (c *Config) TextFiles(ctx context.Context, exec boil.ContextExecutor, sl *slog.Logger) error {
	uuids, err := model.UUID(ctx, exec)
	if err != nil {
		return fmt.Errorf("config %w", err)
	}
	dupes := 0
	for val := range slices.Values(uuids) {
		name := filepath.Join(c.AbsExtra.String(), val.UUID.String)
		diz := name + ".diz"
		txt := name + ".txt"
		dizF, err := os.Stat(diz)
		if err != nil || dizF.Size() == 0 {
			continue
		}
		txtF, err := os.Stat(txt)
		if err != nil || txtF.Size() == 0 {
			continue
		}
		if dizF.Size() != txtF.Size() {
			continue
		}
		dizSI, err := helper.StrongIntegrity(diz)
		if err != nil {
			continue
		}
		txtSI, err := helper.StrongIntegrity(txt)
		if err != nil {
			continue
		}
		if identical := dizSI == txtSI; !identical {
			continue
		}
		dupes++
		dupe, err := Remove(diz, txt)
		if err != nil {
			sl.Error("remove",
				slog.String("problem", "cannot remove file duplicates"),
				slog.String("file_id.diz", diz),
				slog.String("readme text", txt))
			continue
		}
		sl.Info("remove",
			slog.String("success", "removed duplicate file_id.diz = readme text"),
			slog.String("filename", dupe))
	}
	if dupes > 0 {
		sl.Info("remove",
			slog.String("duplicates", "discovered text files that are duplicates of the file_id.diz files"),
			slog.Int("finds", dupes))
	}
	return nil
}

// Remove either the named diz or txt file that are idential duplicates.
// The file deleted depends on if the pair look to be a FILE_ID.DIZ or a longer form text file.
//
// If successful, the basename of the file removed is returned.
func Remove(diz, txt string) (string, error) {
	file, err := os.Open(diz)
	if err != nil {
		return "", fmt.Errorf("remove open %w: %s", err, diz)
	}
	defer func() { _ = file.Close() }()
	if !FileID(file) {
		if err := os.Remove(diz); err != nil {
			return "", fmt.Errorf("remove diz %w: %q", err, diz)
		}
		return filepath.Base(diz), nil
	}
	if err := os.Remove(txt); err != nil {
		return "", fmt.Errorf("remove readme %w: %q", err, txt)
	}
	return filepath.Base(txt), nil
}

// FileID will return true if there are less than 10 lines of text
// and the maximum width of each line is no more than 45 characters.
// This is not a guarantee of a [FILE_ID.DIZ] but it is true for many situations.
//
// [FILE_ID.DIZ]: http://www.textfiles.com/computers/fileid.txt
func FileID(r io.Reader) bool {
	scanner := bufio.NewScanner(r)
	const (
		maximumLines = 10
		maximumWidth = 45
	)
	lines := 0
	for scanner.Scan() {
		lines++
		if lines > maximumLines {
			return false
		}
		line := scanner.Text()
		if len(line) > maximumWidth {
			return false
		}
	}
	return true
}

// MagicNumbers checks the magic numbers of the artifacts and replaces any missing or
// legacy values with the current method of detection. Previous detection methods were
// done using the `file` command line utility, which is a bit to verbose for our needs.
func (c *Config) MagicNumbers(ctx context.Context, exec boil.ContextExecutor, sl *slog.Logger) error {
	if err := argspanic(ctx, exec, sl); err != nil {
		return err
	}
	tick := time.Now()
	r := model.Artifacts{}
	magics, err := r.ByMagicErr(ctx, exec, false)
	if err != nil {
		return fmt.Errorf("magicnumbers %w", err)
	}
	const large = 1000
	if len(magics) > large && sl != nil {
		sl.Warn("magic numbers",
			slog.String("issue", "there are a large number of artifacts to check, it could take a while"),
			slog.Int("task count", len(magics)))
	}
	count := 0
	for val := range slices.Values(magics) {
		name := filepath.Join(string(c.AbsDownload), val.UUID.String)
		r, err := os.Open(name)
		if err != nil {
			_ = r.Close()
			continue
		}
		magic := magicnumber.Find(r)
		count++
		_ = model.UpdateMagic(ctx, exec, val.ID, magic.Title())
		_ = r.Close()
	}
	if count == 0 || sl == nil {
		return nil
	}
	sl.Info("magic numbers complete",
		slog.String("success", ""),
		slog.Int("values update", count),
		slog.Duration("time taken", time.Since(tick)))
	return nil
}

// Previews on startup check the preview directory for any unnecessary preview images such as textfile artifacts.
func (c *Config) Previews(ctx context.Context, exec boil.ContextExecutor, sl *slog.Logger) error {
	if err := argspanic(ctx, exec, sl); err != nil {
		return err
	}
	r := model.Artifacts{}
	artifacts, err := r.ByTextPlatform(ctx, exec)
	if err != nil {
		return fmt.Errorf("nopreview %w", err)
	}
	var count, totals int64
	for val := range slices.Values(artifacts) {
		png := filepath.Join(c.AbsPreview.String(), val.UUID.String) + ".png"
		st, err := os.Stat(png)
		if err != nil {
			_, _ = fmt.Fprintln(io.Discard, err)
			continue
		}
		_ = os.Remove(png)
		count++
		totals += st.Size()
	}
	for val := range slices.Values(artifacts) {
		webp := filepath.Join(c.AbsPreview.String(), val.UUID.String) + ".webp"
		st, err := os.Stat(webp)
		if err != nil {
			_, _ = fmt.Fprintln(io.Discard, err)
			continue
		}
		_ = os.Remove(webp)
		count++
		totals += st.Size()
	}
	if count == 0 {
		return nil
	}
	sl.Info("previews",
		slog.String("success", "erased textfile previews"),
		slog.Int64("count", count), slog.String("totaling", helper.ByteCountFloat(totals)))
	return nil
}

// ImageDirs on startup check the image directories for any invalid or unknown files.
func (c *Config) ImageDirs(sl *slog.Logger) error {
	if sl == nil {
		return ErrNoSlog
	}
	backup := dir.Directory(c.AbsOrphaned.String())
	dirs := []string{c.AbsPreview.String(), c.AbsThumbnail.String()}
	if err := removeSub(dirs...); err != nil {
		return fmt.Errorf("remove subdirectories %w", err)
	}
	// remove any invalid files
	p, t := 0, 0
	for dir := range slices.Values(dirs) {
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return fmt.Errorf("walk path %w: %s", err, path)
			}
			name := d.Name()
			if d.IsDir() {
				return nil
			}
			switch dir {
			case c.AbsPreview.String():
				if filepath.Ext(name) == ".png" {
					p++
				}
			case c.AbsThumbnail.String():
				if filepath.Ext(name) == ".png" {
					t++
				}
			}
			return RemoveImage(name, path, backup)
		})
		if err != nil {
			return fmt.Errorf("walk directory %w: %s", err, dir)
		}
		switch dir {
		case c.AbsPreview.String():
			containsInfo(sl, "preview", p)
		case c.AbsThumbnail.String():
			containsInfo(sl, "thumb", t)
		}
	}
	return nil
}

// removeSub removes any subdirectories found in the specified directories.
func removeSub(dirs ...string) error {
	for dir := range slices.Values(dirs) {
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return fmt.Errorf("walk path %w: %s", err, path)
			}
			name := d.Name()
			if d.IsDir() {
				return RemoveDir(name, path, dir)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("walk directory %w: %s", err, dir)
		}
	}
	return nil
}

// containsInfo logs the number of files found in the directory.
func containsInfo(sl *slog.Logger, name string, count int) {
	if sl == nil {
		return
	}
	if MinimumFiles > count {
		sl.Warn(name+" images",
			slog.String("issue", "directory contains too few files"),
			slog.Int("count", count), slog.Int("minimum", MinimumFiles))
		return
	}
	sl.Info(name+" images",
		slog.Int("file count", count))
}

// DownloadDir on startup check the download directory for any invalid or unknown files.
func DownloadDir(sl *slog.Logger, src, dest, extra dir.Directory) error {
	if sl == nil {
		return ErrNoSlog
	}
	if err := src.Check(); err != nil {
		return fmt.Errorf("download directory %w: %s", err, src)
	}
	if err := dest.Check(); err != nil {
		return fmt.Errorf("download directory %w: %s", err, dest)
	}
	if err := extra.Check(); err != nil {
		return fmt.Errorf("download directory %w: %s", err, extra)
	}
	count := 0
	err := filepath.WalkDir(src.Path(), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk path %w: %s", err, path)
		}
		name := d.Name()
		if d.IsDir() {
			return RemoveDir(name, path, src.Path())
		}
		if err = RemoveDownload(name, path, dest, extra); err != nil {
			return fmt.Errorf("remove download: %w", err)
		}
		if filepath.Ext(name) == "" {
			count++
		}
		return RenameDownload(name, path)
	})
	if err != nil {
		return fmt.Errorf("walk directory %w: %s", err, src.Path())
	}
	containsInfo(sl, "downloads", count)
	return nil
}

// RenameDownload rename the download file if the basename uses an invalid coldfusion uuid.
func RenameDownload(basename, absPath string) error {
	if basename == "" || absPath == "" {
		return fmt.Errorf("rename download %w: %s %s", ErrNoPath, basename, absPath)
	}

	ext := filepath.Ext(basename)
	rawname, found := strings.CutSuffix(basename, ext)
	if !found {
		return nil
	}
	const cflen = len(cfid) // coldfusion uuid length
	if len(rawname) != cflen {
		return nil
	}

	newname, _ := helper.CfUUID(rawname)
	if err := uuid.Validate(newname); err != nil {
		return fmt.Errorf("uuid.Validate %q: %w", newname, err)
	}
	dir := filepath.Dir(absPath)
	oldpath := filepath.Join(dir, basename)
	newpath := filepath.Join(dir, newname+ext)

	rename(oldpath, "renamed invalid cfid", newpath)
	return nil
}

// RemoveDir check the directory for invalid names.
// If any are found, they are printed to stderr.
// Any directory that matches the name ".stfolder" is removed.
func RemoveDir(name, path, root string) error {
	if name == "" || path == "" || root == "" {
		return fmt.Errorf("remove directory %w: %s %s %s", ErrNoPath, name, path, root)
	}
	rootDir := filepath.Base(root)
	switch name {
	case rootDir:
		return nil
	case syncthing:
		defer func() {
			err := os.RemoveAll(path)
			_, _ = fmt.Fprintln(os.Stderr, err)
		}()
	default:
		fmt.Fprintln(os.Stderr, "unknown dir:", path)
		return nil
	}
	return nil
}

// RemoveDownload checks the download files for invalid names and extensions.
// If any are found, they are removed without warning.
// Basename must be the name of the file with a valid file extension.
//
// Valid file extensions are none, .chiptune, .txt, and .zip.
func RemoveDownload(basename, path string, backup, extra dir.Directory) error {
	if basename == "" || path == "" {
		return fmt.Errorf("remove download %w: %s %s", ErrNoPath, basename, path)
	}
	const filedownload = ""
	ext := filepath.Ext(basename)
	switch ext {
	case filedownload:
		return nil
	case ".txt", ".zip", ".chiptune":
		rename(path, "rename valid ext", extra.Join(basename))
	default:
		remove(basename, "remove invalid ext", path, backup)
	}
	return nil
}

// RemoveImage checks the image files for invalid names and extensions.
// If any are found, they are moved to the destDir without warning.
// Basename must be the name of the file with a valid file extension.
//
// Valid file extensions are .png and .webp, and basename must be a
// valid uuid or cfid with the correct length.
func RemoveImage(basename, path string, backup dir.Directory) error {
	if basename == "" || path == "" {
		return fmt.Errorf("remove image %w: %s %s", ErrNoPath, basename, path)
	}
	if err := backup.Check(); err != nil {
		return fmt.Errorf("remove image %w: %s", err, backup)
	}
	const (
		png   = ".png"    // png file extension
		webp  = ".webp"   // webp file extension
		valid = len(unid) // valid uuid length
		cflen = len(cfid) // coldfusion uuid length
	)
	ext := filepath.Ext(basename)
	if filename, found := strings.CutSuffix(basename, ext); found {
		if len(filename) == cflen {
			filename, _ = helper.CfUUID(filename)
			newpath := filepath.Dir(path)
			switch ext {
			case png, webp:
				rename(path, "rename cfid "+ext, filepath.Join(newpath, filename+ext))
				return nil
			}
		}
		if err := uuid.Validate(filename); err != nil {
			remove(basename, "remove invalid uuid image", path, backup)
			return nil //nolint:nilerr
		}
	}
	switch ext {
	case png, webp:
		return nil
	default:
		remove(basename, "remove invalid uuid ext", path, backup)
	}
	return nil
}

// remove the file without warning.
func remove(name, info, path string, backup dir.Directory) {
	w := os.Stderr
	_, _ = fmt.Fprintf(w, "%s: %s\n", info, name)
	defer func() {
		now := time.Now().Format("2006-01-02_15-04-05")
		newpath := backup.Join(fmt.Sprintf("%s_%s", now, name))
		err := helper.RenameCrossDevice(path, newpath)
		if err != nil {
			_, _ = fmt.Fprintf(w, "defer repair file remove: %s\n", err)
		}
	}()
}

// rename the file without warning.
func rename(oldpath, info, newpath string) {
	w := os.Stderr
	_, _ = fmt.Fprintf(w, "%s: %s\n", info, oldpath)
	defer func() {
		if err := helper.RenameCrossDevice(oldpath, newpath); err != nil {
			_, _ = fmt.Fprintf(w, "defer repair file rename: %s\n", err)
		}
	}()
}

// TmpCleaner will remove any temporary directories created by this web applcation
// that are older than 3 days.
//
// This is a safety measure to ensure that the server does not run out of disk space.
func TmpCleaner() {
	const threeDays = 3 * 24 * time.Hour
	w := os.Stderr
	name := helper.TmpDir()
	dir, err := os.OpenRoot(name)
	if err != nil {
		_, _ = fmt.Fprintf(w, "repair tmp cleaner %s: %s", err, name)
		return
	}
	defer func() { _ = dir.Close() }()
	_ = fs.WalkDir(dir.FS(), ".", func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			_, _ = fmt.Fprint(io.Discard, err)
			return nil
		}
		if !d.IsDir() || !strings.HasPrefix(d.Name(), "artifact-content-") {
			return nil
		}
		inf, err := d.Info()
		if err != nil {
			_, _ = fmt.Fprintf(w, "repair tmp cleaner %s: %s", err, d.Name())
			return nil
		}
		if time.Since(inf.ModTime()) < threeDays {
			return nil
		}
		rmpath := filepath.Join(name, d.Name())
		if err := os.RemoveAll(rmpath); err != nil {
			_, _ = fmt.Fprintf(w, "repair tmp cleaner %s: %s", err, rmpath)
		}
		return nil
	})
}
