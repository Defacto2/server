package config

// Package file repair.go contains the repair functions for assets and downloads.

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
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
	"github.com/google/uuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/zap"
)

const (
	unid      = "00000000-0000-0000-0000-000000000000" // common universal unique identifier example
	cfid      = "00000000-0000-0000-0000000000000000"  // coldfusion uuid example
	syncthing = ".stfolder"                            // syncthing directory name
)

var (
	ErrCE    = errors.New("nil context executor")
	ErrEmpty = errors.New("empty path or name")
)

// Archives, on startup check the download directory for any legacy and obsolete archives.
// Obsolete archives are those that use a legacy compression method that is not supported
// by Go or JS libraries used by the website.
func (c *Config) Archives(ctx context.Context, exec boil.ContextExecutor) error { //nolint:cyclop,funlen,gocognit
	if exec == nil {
		return fmt.Errorf("config repair archives %w", ErrCE)
	}
	d := time.Now()
	logger := helper.Logger(ctx)
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
		if err := Zip.ReArchive(ctx, path, uid, extra); err != nil {
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
		if err := LHA.ReArchive(ctx, path, uid, extra); err != nil {
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
		if err := Arc.ReArchive(ctx, path, uid, extra); err != nil {
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
		if err := Arj.ReArchive(ctx, path, uid, extra); err != nil {
			return fmt.Errorf("arj repair and re-archive: %w", err)
		}
		return nil
	}

	download := dir.Directory(c.AbsDownload)
	for repair := range slices.Values(repairs()) {
		if err := repair.lookPath(); err != nil {
			logger.Errorf("repair %s archives: %s", repair.String(), err)
			continue
		}
		artifacts, err = repair.artifacts(ctx, exec, logger)
		if err != nil {
			logger.Errorf("repair %s archives: %s", repair.String(), err)
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
			logger.Errorf("walk directory %s: %s", err, download.Path())
		}
	}
	logger.Infof("Completed UUID archive checks in %.1fs", time.Since(d).Seconds())
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

func (r Repair) artifacts(ctx context.Context, exec boil.ContextExecutor, logger *zap.SugaredLogger) ([]string, error) {
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
	logger.Infof("Check %d %s %s archives", size, tags.DOS.String(), r.String())
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

// ReArchive, re-archive the file using the specified compression method.
// The source file is extracted to a temporary directory, then re-compressed
// and saved to the destination directory using the uid as the new named file.
// The original src file is not removed.
func (r Repair) ReArchive(ctx context.Context, src, uid string, dest dir.Directory) error {
	if src == "" || uid == "" {
		return fmt.Errorf("rearchive %s %w: %q %q", r, ErrEmpty, src, uid)
	}
	if err := dest.IsDir(); err != nil {
		return fmt.Errorf("rearchive %s %w: %q", r, err, dest)
	}
	logger := helper.Logger(ctx)
	tmp, err := os.MkdirTemp(helper.TmpDir(), "rearchive-")
	if err != nil {
		return fmt.Errorf("rearchive mkdir temp %w: %s", err, src)
	}
	defer os.RemoveAll(tmp)

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
	cmd := exec.CommandContext(ctx1min, extractCmd, extractArg, src)
	cmd.Dir = tmp
	if stdoutStderr, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("rearchive run %w: %s: dump: %q",
			err, src, stdoutStderr)
	}
	c, err := helper.Count(tmp)
	if err != nil {
		return fmt.Errorf("rearchive tmp count %w: %s", err, tmp)
	}
	logger.Infof("Rezipped %d files for %s found in: %s", c, uid, tmp)
	_, err = os.Stat(tmp)
	if err != nil {
		return fmt.Errorf("rearchive tmp stat %w: %s", err, tmp)
	}

	basename := uid + ".zip"
	tmpArc := filepath.Join(helper.TmpDir(), basename)
	if written, err := rezip.CompressDir(tmp, tmpArc); err != nil {
		return fmt.Errorf("rearchive dir %w: %s", err, tmp)
	} else if written == 0 {
		return nil
	}

	finalArc := dest.Join(basename)
	if err = helper.RenameCrossDevice(tmpArc, finalArc); err != nil {
		defer os.RemoveAll(tmpArc)
		return fmt.Errorf("rearchive rename %w: %s", err, tmpArc)
	}

	st, err := os.Stat(finalArc)
	if err != nil {
		return fmt.Errorf("rearchive zip stat %w: %s", err, finalArc)
	}
	logger.Infof("Extra deflated zipfile created %d bytes: %s", st.Size(), finalArc)
	return nil
}

// Assets, on startup check the file system directories for any invalid or unknown files.
// These specifically match the base filename against the UUID column in the database.
// When there is no matching UUID, the file is considered orphaned and these are moved
// to the orphaned directory without warning.
//
// There are no checks on the 3 directories that get scanned.
func (c *Config) Assets(ctx context.Context, exec boil.ContextExecutor) error {
	if exec == nil {
		return fmt.Errorf("config repair assets %w", ErrCE)
	}
	d := time.Now()
	logger := helper.Logger(ctx)
	mods := []qm.QueryMod{}
	mods = append(mods, qm.Select("uuid"))
	mods = append(mods, qm.WithDeleted())
	files, err := models.Files(mods...).All(ctx, exec)
	if err != nil {
		return fmt.Errorf("config repair select all uuids: %w", err)
	}
	size := len(files)
	logger.Infof("Check %d UUIDs", size)
	artifacts := make([]string, size)
	for i, f := range files {
		if !f.UUID.Valid || f.UUID.String == "" {
			continue
		}
		artifacts[i] = f.UUID.String
	}
	artifacts = slices.Clip(artifacts)
	slices.Sort(artifacts)

	dirs := []string{c.AbsDownload, c.AbsPreview, c.AbsThumbnail}
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
					unknownAsset(logger, path, d.Name(), uid, orphaned)
				}
				return nil
			})
			if err != nil {
				logger.Errorf("walk directory %w: %s", err, dir)
			}
		}(dir)
	}

	wg.Wait()
	sum := 0
	for val := range slices.Values(counters) {
		sum += val
	}
	logger.Infof("Checked %d files for %d UUIDs in %.1fs", sum, size, time.Since(d).Seconds())
	return nil
}

// unknownAsset logs a warning message for an unknown asset file.
func unknownAsset(logger *zap.SugaredLogger, oldpath, name, uid string, orphaned dir.Directory) {
	logger.Warnf("Unknown file: %s, no matching artifact for UUID: %q", name, uid)
	defer func() {
		now := time.Now().Format("2006-01-02_15-04-05")
		dest := orphaned.Join(fmt.Sprintf("%s_%s", now, name))
		if err := helper.RenameCrossDevice(oldpath, dest); err != nil {
			logger.Errorf("could not move orphaned artifact asset for %q: %s", name, err)
		}
	}()
}

// RepairAssets, on startup check the file system directories for any invalid or unknown files.
// If any are found, they are removed without warning.
func (c *Config) RepairAssets(ctx context.Context, exec boil.ContextExecutor) error {
	if exec == nil {
		return fmt.Errorf("config repair assets %w", ErrCE)
	}
	logger := helper.Logger(ctx)
	backup := dir.Directory(c.AbsOrphaned)
	if st, err := os.Stat(backup.Path()); err != nil {
		return fmt.Errorf("repair backup directory %w: %s", err, backup.Path())
	} else if !st.IsDir() {
		return fmt.Errorf("repair backup directory %w: %s", ErrNotDir, backup.Path())
	}
	if err := c.ImageDirs(logger); err != nil {
		return fmt.Errorf("repair the images directories %w", err)
	}
	src := dir.Directory(c.AbsDownload)
	extra := dir.Directory(c.AbsExtra)
	if err := DownloadDir(logger, src, backup, extra); err != nil {
		return fmt.Errorf("repair the download directory %w", err)
	}
	if err := c.Assets(ctx, exec); err != nil {
		return fmt.Errorf("repair assets %w", err)
	}
	if err := c.Archives(ctx, exec); err != nil {
		return fmt.Errorf("repair archives %w", err)
	}
	if err := c.Previews(ctx, exec, logger); err != nil {
		return fmt.Errorf("repair previews %w", err)
	}
	if err := c.MagicNumbers(ctx, exec, logger); err != nil {
		return fmt.Errorf("repair magics %w", err)
	}
	if err := c.TextFiles(ctx, exec, logger); err != nil {
		return fmt.Errorf("repair textfiles %w", err)
	}
	return nil
}

// TextFiles, on startup check the extra directory for any readme text files that are duplicates of the diz text files.
func (c *Config) TextFiles(ctx context.Context, exec boil.ContextExecutor, logger *zap.SugaredLogger) error {
	if exec == nil {
		return fmt.Errorf("config %w", ErrCE)
	}
	uuids, err := model.UUID(ctx, exec)
	if err != nil {
		return fmt.Errorf("config %w", err)
	}
	dupes := 0
	for val := range slices.Values(uuids) {
		name := filepath.Join(c.AbsExtra, val.UUID.String)
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
			logger.Errorf("Could not remove duplicate file_id.diz = readme files: %s %s", diz, txt)
			continue
		}
		logger.Infoln("Removed duplicate file_id.diz = readme file:", dupe)
	}
	if dupes > 0 {
		logger.Infof("Found %d text files that are duplicate texts: file_id.diz = readme", dupes)
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
	defer file.Close()
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
func (c *Config) MagicNumbers(ctx context.Context, exec boil.ContextExecutor, logger *zap.SugaredLogger) error {
	if exec == nil {
		return fmt.Errorf("config repair magic numbers %w", ErrCE)
	}
	tick := time.Now()
	r := model.Artifacts{}
	magics, err := r.ByMagicErr(ctx, exec, false)
	if err != nil {
		return fmt.Errorf("magicnumbers %w", err)
	}
	const large = 1000
	if len(magics) > large && logger != nil {
		logger.Warnf("Check %d magic number values for artifacts, this could take a while", len(magics))
	}
	count := 0
	for val := range slices.Values(magics) {
		name := filepath.Join(c.AbsDownload, val.UUID.String)
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
	if count == 0 || logger == nil {
		return nil
	}
	logger.Infof("Updated %d magic number values for artifacts in %s", count, time.Since(tick))
	return nil
}

// Previews, on startup check the preview directory for any unnecessary preview images such as textfile artifacts.
func (c *Config) Previews(ctx context.Context, exec boil.ContextExecutor, logger *zap.SugaredLogger) error {
	if exec == nil {
		return fmt.Errorf("config repair previews %w", ErrCE)
	}
	r := model.Artifacts{}
	artifacts, err := r.ByTextPlatform(ctx, exec)
	if err != nil {
		return fmt.Errorf("nopreview %w", err)
	}
	var count, totals int64
	for val := range slices.Values(artifacts) {
		png := filepath.Join(c.AbsPreview, val.UUID.String) + ".png"
		st, err := os.Stat(png)
		if err != nil {
			fmt.Fprintln(io.Discard, err)
			continue
		}
		_ = os.Remove(png)
		count++
		totals += st.Size()
	}
	for val := range slices.Values(artifacts) {
		webp := filepath.Join(c.AbsPreview, val.UUID.String) + ".webp"
		st, err := os.Stat(webp)
		if err != nil {
			fmt.Fprintln(io.Discard, err)
			continue
		}
		_ = os.Remove(webp)
		count++
		totals += st.Size()
	}
	if count == 0 || logger == nil {
		return nil
	}
	logger.Infof("Erased %d textfile preview images, totaling %s", count, helper.ByteCountFloat(totals))
	return nil
}

// ImageDirs, on startup check the image directories for any invalid or unknown files.
func (c *Config) ImageDirs(logger *zap.SugaredLogger) error {
	backup := dir.Directory(c.AbsOrphaned)
	dirs := []string{c.AbsPreview, c.AbsThumbnail}
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
			case c.AbsPreview:
				if filepath.Ext(name) == ".png" {
					p++
				}
			case c.AbsThumbnail:
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
		case c.AbsPreview:
			containsInfo(logger, "preview", p)
		case c.AbsThumbnail:
			containsInfo(logger, "thumb", t)
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
func containsInfo(logger *zap.SugaredLogger, name string, count int) {
	if logger == nil {
		return
	}
	if MinimumFiles > count {
		logger.Warnf("The %s directory contains %d files, which is less than the minimum of %d",
			name, count, MinimumFiles)
		return
	}
	logger.Infof("The %s directory contains %d files", name, count)
}

// DownloadDir, on startup check the download directory for any invalid or unknown files.
func DownloadDir(logger *zap.SugaredLogger, src, dest, extra dir.Directory) error {
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
	containsInfo(logger, "downloads", count)
	return nil
}

// RenameDownload, rename the download file if the basename uses an invalid coldfusion uuid.
func RenameDownload(basename, absPath string) error {
	if basename == "" || absPath == "" {
		return fmt.Errorf("rename download %w: %s %s", ErrEmpty, basename, absPath)
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

// RemoveDir, check the directory for invalid names.
// If any are found, they are printed to stderr.
// Any directory that matches the name ".stfolder" is removed.
func RemoveDir(name, path, root string) error {
	if name == "" || path == "" || root == "" {
		return fmt.Errorf("remove directory %w: %s %s %s", ErrEmpty, name, path, root)
	}
	rootDir := filepath.Base(root)
	switch name {
	case rootDir:
		return nil
	case syncthing:
		defer os.RemoveAll(path)
	default:
		fmt.Fprintln(os.Stderr, "unknown dir:", path)
		return nil
	}
	return nil
}

// RemoveDownload, check the download files for invalid names and extensions.
// If any are found, they are removed without warning.
// Basename must be the name of the file with a valid file extension.
//
// Valid file extensions are none, .chiptune, .txt, and .zip.
func RemoveDownload(basename, path string, backup, extra dir.Directory) error {
	if basename == "" || path == "" {
		return fmt.Errorf("remove download %w: %s %s", ErrEmpty, basename, path)
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

// RemoveImage, check the image files for invalid names and extensions.
// If any are found, they are moved to the destDir without warning.
// Basename must be the name of the file with a valid file extension.
//
// Valid file extensions are .png and .webp, and basename must be a
// valid uuid or cfid with the correct length.
func RemoveImage(basename, path string, backup dir.Directory) error {
	if basename == "" || path == "" {
		return fmt.Errorf("remove image %w: %s %s", ErrEmpty, basename, path)
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
	fmt.Fprintf(w, "%s: %s\n", info, name)
	defer func() {
		now := time.Now().Format("2006-01-02_15-04-05")
		newpath := backup.Join(fmt.Sprintf("%s_%s", now, name))
		err := helper.RenameCrossDevice(path, newpath)
		if err != nil {
			fmt.Fprintf(w, "defer repair file remove: %s\n", err)
		}
	}()
}

// rename the file without warning.
func rename(oldpath, info, newpath string) {
	w := os.Stderr
	fmt.Fprintf(w, "%s: %s\n", info, oldpath)
	defer func() {
		if err := helper.RenameCrossDevice(oldpath, newpath); err != nil {
			fmt.Fprintf(w, "defer repair file rename: %s\n", err)
		}
	}()
}
