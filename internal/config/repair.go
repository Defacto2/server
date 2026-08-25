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
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/Defacto2/archive/rezip"
	"github.com/Defacto2/helper"
	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/command/option"
	"github.com/Defacto2/server/internal/config/fixarc"
	"github.com/Defacto2/server/internal/config/fixarj"
	"github.com/Defacto2/server/internal/config/fixlha"
	"github.com/Defacto2/server/internal/config/fixzip"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"
)

const (
	Timeout = 1 * time.Minute
)

const (
	unid      = "00000000-0000-0000-0000-000000000000" // common universal unique identifier example
	cfid      = "00000000-0000-0000-0000000000000000"  // coldfusion uuid example
	syncthing = ".stfolder"                            // syncthing directory name
)

// RepairArchives checks the download directory for any legacy and obsolete archives.
// Obsolete archives are those that use a legacy compression method
// that is not supported by Go or JS libraries used by the website.
func (c *Config) RepairArchives(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor) error {
	const format = "config archives repair: %w"
	if err := nils.Check(ctx, sl, exec); err != nil {
		return fmt.Errorf(format, err)
	}

	start := time.Now()
	download := dir.Directory(c.AbsDownload.String())

	s := repairs()
	for repair := range slices.Values(s[:]) {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf(format, err)
		}

		if err := repair.lookups(); err != nil {
			sl.Error("archives "+repair.String(), slog.Any("error", err))
			continue
		}

		artifacts, err := repair.artifacts(ctx, sl, exec)
		if err != nil {
			sl.Error("archives "+repair.String(), slog.Any("error", err))
			continue
		}

		if err := c.walkAndRepair(ctx, sl, repair, artifacts); err != nil {
			sl.Error("Archives directory walk",
				slog.Any("error", err),
				slog.String("path", download.Path()),
				slog.String("format", repair.String()))
		}
	}

	sl.Info("Archives check", slog.String("task", "Time taken"),
		slog.Duration("time", time.Since(start).Round(time.Millisecond)))

	return nil
}

// walkAndRepair executes filepath.WalkDir using format-specific validation hooks.
func (c *Config) walkAndRepair(ctx context.Context, sl *slog.Logger, repair Repair, artifacts []string) error {
	root := c.AbsDownload.String()
	extra := dir.Directory(c.AbsExtra)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%w: %s", err, path)
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf("walk path ctx error: %w", ctxErr)
		}

		// format specific checks
		var uid string
		var invalid bool

		switch repair {
		case Zip:
			uid = fixzip.Check(sl, path, extra, d, artifacts...)
			invalid = uid == "" || fixzip.Invalid(ctx, sl, path)
		case LHA:
			uid = fixlha.Check(sl, extra, d, artifacts...)
			invalid = uid == "" || fixlha.Invalid(ctx, sl, path)
		case Arc:
			uid = fixarc.Check(sl, path, extra, d, artifacts...)
			invalid = uid == "" || fixarc.Invalid(ctx, sl, path)
		case Arj:
			uid = fixarj.Check(extra, d, artifacts...)
			invalid = uid == "" || fixarj.Invalid(ctx, sl, path)
		}

		if invalid {
			return nil
		}

		// re-archive the legacy file
		ra := Repack{Source: path, UID: uid, Destination: extra}
		if err := repair.NewArchive(ctx, sl, ra); err != nil {
			return fmt.Errorf("%s repair and re-archive: %w", repair.String(), err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("walk and repair: %w", err)
	}

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

func repairs() [4]Repair {
	return [4]Repair{Zip, LHA, Arc, Arj}
}

func (r Repair) String() string {
	return [...]string{"zip", "lha", "arc", "arj"}[r]
}

// Repack are the source and destination arguments required by the ReArchive Repair method.
type Repack struct {
	Source      string        // Source is the file extracted to a temporary directory and re-compressed.
	UID         string        // UID is the destination filename using a universal unique ID naming syntax.
	Destination dir.Directory // Destination is the directory to save the re-compressed file.
}

// NewArchive extracts the source archive file and then repackages (rearchives) it to a DEFLATE zip archive.
// The original source file is always kept.
func (r Repair) NewArchive(ctx context.Context, sl *slog.Logger, ra Repack) error { //nolint:funlen
	const format = "config repair and repack %s: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, "check", err)
	}
	if ra.Source == "" || ra.UID == "" {
		return fmt.Errorf(format, "source or uid are missing", ErrNoPath)
	}
	if err := ra.Destination.IsDir(); err != nil {
		return fmt.Errorf(format, "destination is not a directory", err)
	}

	// resolve extraction command per format
	extractCmd, extractArg, err := r.extractConfig()
	if err != nil {
		return fmt.Errorf(format, "extract config", err)
	}

	// prepare temporary directory
	root, err := dir.MkdirTemp("newarc")
	if err != nil {
		return fmt.Errorf(format, "make temp dir", err)
	}
	defer func() {
		if rmErr := dir.RemoveAll(root); rmErr != nil {
			sl.Error("repack failed to remove temp dir", slog.Any("error", rmErr))
		}
	}()

	// extract source archive
	timeout, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	cr := command.Runner{
		Timeout:    1 * Timeout,
		Log:        sl,
		WorkingDir: root,
	}
	arg := option.Join(extractArg, ra.Source)
	out, err := cr.Run(timeout, extractCmd, arg...)
	if err != nil {
		return fmt.Errorf(format, "run command error: "+string(out), err)
	}

	// inspect extracted files
	files, err := helper.Count(root)
	if err != nil {
		return fmt.Errorf(format, "tmp count", err)
	}

	const msg = "Repacked archive"
	sl.Info(msg, slog.String("entry", ra.UID),
		slog.Int("file(s)", files), slog.String("tmp", root))

	// compress extracted directory into standard ZIP
	name := ra.UID + ".zip"
	dest := filepath.Join(root, name)

	written, err := rezip.CompressDir(root, dest)
	if err != nil {
		return fmt.Errorf(format, "compress dir", err)
	}
	if written == 0 {
		return nil
	}

	// move temporary archive to destination
	newpath := ra.Destination.Join(name)
	if err := helper.RenameCrossDevice(dest, newpath); err != nil {
		return fmt.Errorf(format, "rename archive", err)
	}

	st, err := os.Stat(newpath)
	if err != nil {
		return fmt.Errorf(format, "stat archive", err)
	}

	sl.Info(msg, slog.String("Re-archive", "Contemporary 'deflate' zip archive created"),
		slog.Int64("bytes", st.Size()), slog.String("path", newpath))

	return nil
}

// extractConfig maps the Repair type to its extraction command and argument.
func (r Repair) extractConfig() (cmd, arg string, err error) { //nolint:nonamedreturns
	switch r {
	case Zip:
		return command.HWZip, "extract", nil
	case LHA:
		return command.Lha, "xf", nil
	case Arc:
		return command.Arc, "x", nil
	case Arj:
		return command.Zip7, "x", nil
	default:
		return "", "", fmt.Errorf("repair format %v: %w", r, ErrFormat)
	}
}

func (r Repair) lookups() error {
	const format = "cannot find %s exec: %w"
	switch r {
	case Zip:
		if _, err := command.Lookup(command.HWZip); err != nil {
			return fmt.Errorf(format, "hwzip", err)
		}
	case LHA:
		if _, err := command.Lookup(command.Lha); err != nil {
			return fmt.Errorf(format, "lha", err)
		}
	case Arc:
		if _, err := command.Lookup(command.Arc); err != nil {
			return fmt.Errorf(format, "arc", err)
		}
	case Arj:
		if _, err := command.Lookup(command.Zip7); err != nil {
			return fmt.Errorf(format, "7zz", err)
		}
	default:
	}
	return nil
}

func (r Repair) artifacts(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor) ([]string, error) {
	const format = "repair artifacts %s: %w"
	if err := nils.Check(ctx, sl, exec); err != nil {
		return nil, fmt.Errorf(format, "check", err)
	}

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
	default:
		return nil, fmt.Errorf(format, "unsupported repair format "+r.String(), ErrFormat)
	}
	if err != nil {
		return nil, fmt.Errorf(format, r.String(), err)
	}

	sl.Info("Repair artifact", slog.String("era", "MS-DOS"),
		slog.String("format", r.String()), slog.Int("count", len(files)))

	artifacts := make([]string, 0, len(files))
	for _, f := range files {
		if f.UUID.Valid && f.UUID.String != "" {
			artifacts = append(artifacts, f.UUID.String)
		}
	}

	slices.Sort(artifacts)
	return slices.Clip(artifacts), nil
}

// Assets on startup checks the file system directories for any invalid or unknown files.
// These specifically match the base filename against the UUID column in the database.
// When there is no matching UUID, the file is considered orphaned and these are moved
// to the orphaned directory without warning.
//
// There are no checks on the 3 directories that get scanned.
func (c *Config) Assets(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor) error { //nolint:funlen
	const format = "repair assets %s: %w"
	if err := nils.Check(ctx, sl, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	start := time.Now()

	mods := []qm.QueryMod{qm.Select("uuid"), qm.WithDeleted()}
	files, err := models.Files(mods...).All(ctx, exec)
	if err != nil {
		return fmt.Errorf(format, "select all uuids", err)
	}

	count := len(files)
	const msg = "Repair assets"
	sl.Info(msg, slog.String("task", "Check UUID count"), slog.Int("result", count))

	artifacts := make([]string, 0, count)
	for _, f := range files {
		if f.UUID.Valid && f.UUID.String != "" {
			artifacts = append(artifacts, f.UUID.String)
		}
	}
	slices.Sort(artifacts)
	artifacts = slices.Clip(artifacts)

	dirs := [...]string{c.AbsDownload.String(), c.AbsPreview.String(), c.AbsThumbnail.String()}
	orphaned := dir.Directory(c.AbsOrphaned)

	var (
		wg          sync.WaitGroup
		totalChecks atomic.Int64
	)

	for _, targetDir := range dirs {
		if targetDir == "" {
			continue
		}

		wg.Add(1)
		go func(root string) {
			defer wg.Done()

			var localCount int64
			walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return fmt.Errorf(format, "walk path "+path, walkErr)
				}
				if ctxErr := ctx.Err(); ctxErr != nil {
					return fmt.Errorf("walk path ctx error: %w", ctxErr)
				}
				if d.IsDir() {
					return nil
				}

				localCount++
				uid := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
				if _, found := slices.BinarySearch(artifacts, uid); !found {
					unknownAsset(sl, path, d.Name(), uid, orphaned)
				}
				return nil
			})

			totalChecks.Add(localCount)

			if walkErr != nil {
				sl.Error(msg, slog.String("walk_directory", root), slog.Any("error", walkErr))
			}
		}(targetDir)
	}

	wg.Wait()

	sl.Info(msg,
		slog.String("task", "Time taken"), slog.Int64("checks", totalChecks.Load()),
		slog.Int("uuids", count), slog.Duration("time", time.Since(start).Round(time.Millisecond)),
	)

	return nil
}

// unknownAsset logs a warning message for an unknown asset file and moves it to the orphaned directory.
func unknownAsset(sl *slog.Logger, oldpath, name, uid string, orphaned dir.Directory) {
	const msg = "unknown file"

	if sl == nil {
		sl = slog.Default()
	}

	sl.Warn(msg, slog.String("issue", "no matching artifact in the database for the found file"),
		slog.String("uuid", uid), slog.String("filename", name),
	)

	// Format timestamp with sub-second precision to avoid collisions across concurrent goroutines
	now := time.Now().Format("2006-01-02_15-04-05.000000")
	dest := orphaned.Join(fmt.Sprintf("%s_%s_%s", now, uid, name))

	if err := helper.RenameCrossDevice(oldpath, dest); err != nil {
		sl.Error(msg,
			slog.String("issue", "could not move the file to the orphaned directory"),
			slog.String("source_path", oldpath),
			slog.String("destination_path", dest),
			slog.Any("error", err),
		)
	}
}

// assets on startup check the file system directories for any invalid or unknown files.
// If any are found, they are removed without warning.
func (c *Config) assets(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor) error {
	const format = "config assets repair %s: %w"
	if err := nils.Check(ctx, sl, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	backup := dir.Directory(c.AbsOrphaned)
	if err := backup.Check(sl); err != nil {
		return fmt.Errorf(format, "backup directory", err)
	}

	steps := [...]struct {
		name string
		fn   func() error
	}{
		{"images directories", func() error { return c.ImageDirs(sl) }},
		{"download directory", func() error {
			return DownloadDir(sl, dir.Directory(c.AbsDownload), backup, dir.Directory(c.AbsExtra))
		}},
		{"assets", func() error { return c.Assets(ctx, sl, exec) }},
		{"archives", func() error { return c.RepairArchives(ctx, sl, exec) }},
		{"previews", func() error { return c.Previews(ctx, sl, exec) }},
		{"magic numbers", func() error { return c.MagicNumbers(ctx, sl, exec) }},
		{"textfiles", func() error { return c.TextFiles(ctx, sl, exec) }},
	}

	for _, step := range steps {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf(format, "execute", err)
		}

		if err := step.fn(); err != nil {
			return fmt.Errorf(format, step.name, err)
		}
	}

	return nil
}

// TextFiles on startup check the extra directory for any readme text files that are duplicates of the diz text files.
func (c *Config) TextFiles(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor) error { //nolint:funlen
	const format = "repair text files %s: %w"
	if err := nils.Check(ctx, sl, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	uuids, err := model.UUID(ctx, exec)
	if err != nil {
		return fmt.Errorf(format, "model uuid", err)
	}

	extraDir := c.AbsExtra.String()
	dupes := 0

	const msg = "Fix textfile"
	for val := range slices.Values(uuids) {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf("%s: %w", msg, ctxErr)
		}

		if !val.UUID.Valid || val.UUID.String == "" {
			continue
		}

		basePath := filepath.Join(extraDir, val.UUID.String)
		diz := basePath + ".diz"
		txt := basePath + ".txt"

		dizF, err := os.Stat(diz)
		if err != nil || dizF.Size() == 0 {
			continue
		}
		txtF, err := os.Stat(txt)
		if err != nil || txtF.Size() == 0 {
			continue
		}

		// match exact file sizes before performing hash checks
		if dizF.Size() != txtF.Size() {
			continue
		}

		// cryptographic hashes to confirm duplication
		dizSI, err := helper.StrongIntegrity(diz)
		if err != nil {
			continue
		}
		txtSI, err := helper.StrongIntegrity(txt)
		if err != nil {
			continue
		}
		if dizSI != txtSI {
			continue
		}

		dupes++

		// remove duplicate file
		dupe, err := Remove(diz, txt)
		if err != nil {
			sl.Error(msg,
				slog.String("problem", "Cannot remove file duplicates"),
				slog.String("file_id.diz", diz), slog.String("readme_text", txt),
				slog.Any("error", err),
			)
			continue
		}

		sl.Info(msg, slog.String("success", "Removed duplicate text: fileid == readme"),
			slog.String("filename", dupe),
		)
	}

	if dupes > 0 {
		sl.Info(msg, slog.String("duplicates", "Discovered text duplicates"), slog.Int("finds", dupes))
	}

	return nil
}

// Remove deletes either the named diz or txt file from an identical pair.
// The file deleted depends on whether the pair is identified as a FILE_ID.DIZ or a longer text file.
// If successful, the base name of the removed file is returned.
func Remove(diz, txt string) (string, error) {
	const format = "remove diz text %s: %w"
	descriptor, err := isFileID(diz)
	if err != nil {
		return "", fmt.Errorf(format, "inspect", err)
	}

	if !descriptor {
		if err := os.Remove(diz); err != nil {
			return "", fmt.Errorf(format, "remove diz", err)
		}
		return filepath.Base(diz), nil
	}

	if err := os.Remove(txt); err != nil {
		return "", fmt.Errorf(format, "remove readme", err)
	}

	return filepath.Base(txt), nil
}

func isFileID(path string) (bool, error) {
	r, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("is file id: %w", err)
	}
	defer r.Close()

	return FileID(r), nil
}

// FileID will return true if there are less than 10 lines of text
// and the maximum width of each line is no more than 45 characters.
// This is not a guarantee of a [FILE_ID.DIZ] but it is true for many situations.
//
// [FILE_ID.DIZ]: http://www.textfiles.com/computers/fileid.txt
func FileID(r io.Reader) bool {
	if r == nil {
		return false
	}

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

		if utf8.RuneCountInString(scanner.Text()) > maximumWidth {
			return false
		}
	}

	if err := scanner.Err(); err != nil {
		return false
	}
	return lines > 0
}

// MagicNumbers checks the magic numbers of the artifacts and replaces any missing or
// legacy values with the current method of detection. Previous detection methods were
// done using the `file` command line utility, which is a bit to verbose for our needs.
// MagicNumbers checks the magic numbers of the artifacts and replaces any missing or
// legacy values with the current method of detection.
func (c *Config) MagicNumbers(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor) error {
	const format = "config magic numbers %s: %w"
	if err := nils.Check(ctx, sl, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	tick := time.Now()
	r := model.Artifacts{Bytes: 0, Count: 0, MinYear: 0, MaxYear: 0}

	magics, err := r.ByMagicErr(ctx, exec, false)
	if err != nil {
		return fmt.Errorf(format, "models file slice", err)
	}

	const msg = "magic numbers"
	const large = 1000
	if len(magics) > large && sl != nil {
		sl.Warn(msg+", there are a large number of artifacts to check, it could take a while",
			slog.Int("task_count", len(magics)))
	}

	downloadDir := c.AbsDownload.String()
	count := 0

	for val := range slices.Values(magics) {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf(format, "context", ctxErr)
		}

		if !val.UUID.Valid || val.UUID.String == "" {
			continue
		}

		name := filepath.Join(downloadDir, val.UUID.String)
		magicTitle, err := inspectMagic(name)
		if err != nil {
			continue
		}
		if err := model.UpdateMagic(ctx, exec, val.ID, magicTitle); err != nil {
			if sl != nil {
				sl.Error(msg+" failed to update magic number", slog.Int64("id", val.ID), slog.Any("error", err))
			}
			continue
		}

		count++
	}

	if count > 0 && sl != nil {
		sl.Info(msg, slog.Int("values_update", count),
			slog.Duration("time", time.Since(tick).Round(time.Millisecond)))
	}

	return nil
}

// inspectMagic safely opens a file, extracts its magic number title, and guarantees handle closure.
func inspectMagic(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("inspect magic: %w", err)
	}
	defer file.Close()

	magic := magicnumber.Find(file)
	return magic.Title(), nil
}

// Previews on startup check the preview directory for any unnecessary preview images such as textfile artifacts.
func (c *Config) Previews(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor) error {
	const format = "config previews %s: %w"
	if err := nils.Check(ctx, sl, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	r := model.Artifacts{Bytes: 0, Count: 0, MinYear: 0, MaxYear: 0}
	artifacts, err := r.ByTextPlatform(ctx, exec)
	if err != nil {
		return fmt.Errorf(format, "models file slice", err)
	}

	var count, totals int64
	previewDir := c.AbsPreview.String()
	exts := [...]string{".png", ".webp"}

	for val := range slices.Values(artifacts) {
		// 1. Respect cancellation during disk cleanup
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf(format, "context", ctxErr)
		}

		if !val.UUID.Valid || val.UUID.String == "" {
			continue
		}

		basePath := filepath.Join(previewDir, val.UUID.String)
		for _, ext := range exts {
			file := basePath + ext
			st, err := os.Stat(file)
			if err != nil {
				continue
			}

			if err := os.Remove(file); err != nil {
				if sl != nil {
					sl.Error("previews could not remove unwanted file",
						slog.String("path", file), slog.Any("error", err))
				}
				continue
			}

			count++
			totals += st.Size()
		}
	}

	if count == 0 {
		return nil
	}

	if sl != nil {
		sl.Info("Erased textfile previews", slog.Int64("count", count), slog.String("sum", helper.ByteCountFloat(totals)))
	}

	return nil
}

// ImageDirs on startup check the image directories for any invalid or unknown files.
func (c *Config) ImageDirs(sl *slog.Logger) error {
	const msg = "image directories"
	if err := nils.Check(sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	backup := dir.Directory(c.AbsOrphaned.String())
	dirs := []string{c.AbsPreview.String(), c.AbsThumbnail.String()}
	if err := removeSub(sl, dirs...); err != nil {
		return fmt.Errorf("%s remove subdirectories %w", msg, err)
	}
	// remove any invalid files
	p, t := 0, 0
	for dir := range slices.Values(dirs) {
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return fmt.Errorf("%s walk path %w: %s", msg, err, path)
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
			return RemoveImage(sl, name, path, backup)
		})
		if err != nil {
			return fmt.Errorf("%s walk directory %w: %s", msg, err, dir)
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
func removeSub(sl *slog.Logger, dirs ...string) error {
	if sl == nil {
		sl = logs.Discard()
	}
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
				return RemoveDir(sl, name, path, dir)
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
	const msg = "contains info"
	if err := nils.Check(sl); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	s := "" //nolint:wastedassign
	switch strings.ToLower(name) {
	case "thumb":
		s = " thumbnails"
	case "preview":
		s = " previews"
	case "downloads":
		s = " artifact downloads"
	default:
		s = name
	}
	if MinimumFiles > count {
		sl.Warn("File"+s,
			slog.String("issue", "The directory contains too few files"),
			slog.Int("count", count), slog.Int("minimum", MinimumFiles))
		return
	}
	sl.Info("File"+s,
		slog.Int("count", count))
}

// DownloadDir on startup check the download directory for any invalid or unknown files.
func DownloadDir(sl *slog.Logger, src, dest, extra dir.Directory) error {
	const msg = "download directory"
	if err := nils.Check(sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	if err := src.Check(sl); err != nil {
		return fmt.Errorf("%s %w: %s", msg, err, src)
	}
	if err := dest.Check(sl); err != nil {
		return fmt.Errorf("%s %w: %s", msg, err, dest)
	}
	if err := extra.Check(sl); err != nil {
		return fmt.Errorf("%s %w: %s", msg, err, extra)
	}
	count := 0
	err := filepath.WalkDir(src.Path(), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%s walk path %w: %s", msg, err, path)
		}
		name := d.Name()
		if d.IsDir() {
			return RemoveDir(sl, name, path, src.Path())
		}
		if err = RemoveDownload(sl, name, path, dest, extra); err != nil {
			return fmt.Errorf("%s remove download: %w", msg, err)
		}
		if filepath.Ext(name) == "" {
			count++
		}
		return RenameDownload(sl, name, path)
	})
	if err != nil {
		return fmt.Errorf("%s walk directory %w: %s", msg, err, src.Path())
	}
	containsInfo(sl, "downloads", count)
	return nil
}

// RenameDownload rename the download file if the basename uses an invalid coldfusion uuid.
func RenameDownload(sl *slog.Logger, basename, absPath string) error {
	const msg = "rename download"
	if err := nils.Check(sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	if basename == "" || absPath == "" {
		return fmt.Errorf("%s %w: %s %s", msg, ErrNoPath, basename, absPath)
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
		return fmt.Errorf("%s uuid validate %q: %w", msg, newname, err)
	}
	dir := filepath.Dir(absPath)
	oldpath := filepath.Join(dir, basename)
	newpath := filepath.Join(dir, newname+ext)
	rename(sl, oldpath, "renamed invalid cfid", newpath)
	return nil
}

// RemoveDir check the directory for invalid names.
// If any are found, they are printed to stderr.
// Any directory that matches the name ".stfolder" is removed.
func RemoveDir(sl *slog.Logger, name, path, root string) error {
	const msg = "repair remove directory"
	if err := nils.Check(sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	if name == "" || path == "" || root == "" {
		return fmt.Errorf("%s: %w: %s %s %s", msg, ErrNoPath, name, path, root)
	}
	rootDir := filepath.Base(root)
	switch name {
	case rootDir:
		return nil
	case syncthing:
		defer func() {
			err := os.RemoveAll(path)
			sl.Error(msg, slog.Any("error", err))
		}()
	default:
		sl.Error(msg, slog.String("unknown_path", path))
		return nil
	}
	return nil
}

// RemoveDownload checks the download files for invalid names and extensions.
// If any are found, they are removed without warning.
// Basename must be the name of the file with a valid file extension.
//
// Valid file extensions are none, .chiptune, .txt, and .zip.
func RemoveDownload(sl *slog.Logger, basename, path string, backup, extra dir.Directory) error {
	const msg = "remove download"
	if err := nils.Check(sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	if basename == "" || path == "" {
		return fmt.Errorf("%s %w: %s %s", msg, ErrNoPath, basename, path)
	}
	const filedownload = ""
	ext := filepath.Ext(basename)
	switch ext {
	case filedownload:
		return nil
	case ".txt", ".zip", ".chiptune":
		rename(sl, path, "rename valid ext", extra.Join(basename))
	default:
		remove(sl, basename, "remove invalid ext", path, backup)
	}
	return nil
}

// RemoveImage checks the image files for invalid names and extensions.
// If any are found, they are moved to the destDir without warning.
// Basename must be the name of the file with a valid file extension.
//
// Valid file extensions are .png and .webp, and basename must be a
// valid uuid or cfid with the correct length.
func RemoveImage(sl *slog.Logger, basename, path string, backup dir.Directory) error {
	const msg = "remove image"
	if err := backup.Check(sl); err != nil {
		return fmt.Errorf("%s: %w: %s", msg, err, backup)
	}
	if basename == "" || path == "" {
		return fmt.Errorf("%s %w: %s %s", msg, ErrNoPath, basename, path)
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
				rename(sl, path, "rename cfid "+ext, filepath.Join(newpath, filename+ext))
				return nil
			}
		}
		if err := uuid.Validate(filename); err != nil {
			remove(sl, basename, "remove invalid uuid image", path, backup)
			return nil //nolint:nilerr
		}
	}
	switch ext {
	case png, webp:
		return nil
	default:
		remove(sl, basename, "remove invalid uuid ext", path, backup)
	}
	return nil
}

// remove the file without warning.
func remove(sl *slog.Logger, name, info, path string, backup dir.Directory) {
	const msg = "Remove file"
	if sl == nil {
		sl = logs.Discard()
	}
	sl.Info(msg, slog.String("name", name), slog.String("detail", info))
	defer func() {
		now := time.Now().Format("2006-01-02_15-04-05")
		newpath := backup.Join(fmt.Sprintf("%s_%s", now, name))
		err := helper.RenameCrossDevice(path, newpath)
		if err != nil {
			sl.Error(msg, slog.String("name", name), slog.String("detail", info), slog.Any("error", err))
		}
	}()
}

// rename the file without warning.
func rename(sl *slog.Logger, oldpath, info, newpath string) {
	const msg = "Rename or move file"
	if sl == nil {
		sl = logs.Discard()
	}
	sl.Info(msg, slog.String("original_path", oldpath), slog.String("new_path", newpath), slog.String("detail", info))
	defer func() {
		if err := helper.RenameCrossDevice(oldpath, newpath); err != nil {
			sl.Error(msg, slog.String("original_path", oldpath), slog.String("new_path", newpath),
				slog.String("detail", info), slog.Any("error", err))
		}
	}()
}
