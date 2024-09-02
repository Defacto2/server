package config

// Package file repair.go contains the repair functions for assets and downloads.

import (
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

var ErrEmpty = errors.New("empty path or name")

// Archives, on startup check the download directory for any legacy and obsolete archives.
// Obsolete archives are those that use a legacy compression method that is not supported
// by Go or JS libraries used by the website.
func (c Config) Archives(ctx context.Context, ce boil.ContextExecutor) error { //nolint:cyclop,funlen,gocognit
	if ce == nil {
		return nil
	}
	tick := time.Now()
	downloadDir, logger := c.AbsDownload, helper.Logger(ctx)
	artifacts := []string{}
	var err error

	zipWalker := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%w: %s", err, path)
		}
		uid := fixzip.Check(ctx, path, c.AbsExtra, d, artifacts...)
		if uid == "" || fixzip.Invalid(ctx, path) {
			return nil
		}
		if err := Zip.rearchive(ctx, path, c.AbsExtra, uid); err != nil {
			return fmt.Errorf("zip repair and re-archive: %w", err)
		}
		return nil
	}
	lhaWalker := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%w: %s", err, path)
		}
		uid := fixlha.Check(c.AbsExtra, d, artifacts...)
		if uid == "" || fixlha.Invalid(ctx, path) {
			return nil
		}
		if err := LHA.rearchive(ctx, path, c.AbsExtra, uid); err != nil {
			return fmt.Errorf("lha/lzh repair and re-archive: %w", err)
		}
		return nil
	}
	arcWalker := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%w: %s", err, path)
		}
		uid := fixarc.Check(ctx, path, c.AbsExtra, d, artifacts...)
		if uid == "" || fixarc.Invalid(ctx, path) {
			return nil
		}
		if err := Arc.rearchive(ctx, path, c.AbsExtra, uid); err != nil {
			return fmt.Errorf("arc repair and re-archive: %w", err)
		}
		return nil
	}
	arjWalker := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%w: %s", err, path)
		}
		uid := fixarj.Check(c.AbsExtra, d, artifacts...)
		if uid == "" || fixarj.Invalid(ctx, path) {
			return nil
		}
		if err := Arj.rearchive(ctx, path, c.AbsExtra, uid); err != nil {
			return fmt.Errorf("arj repair and re-archive: %w", err)
		}
		return nil
	}

	for _, r := range repairs() {
		if err := r.lookPath(); err != nil {
			logger.Errorf("repair %s archives: %s", r.String(), err)
			continue
		}
		artifacts, err = r.artifacts(ctx, ce, logger)
		if err != nil {
			logger.Errorf("repair %s archives: %s", r.String(), err)
			continue
		}
		switch r {
		case Zip:
			err = filepath.WalkDir(downloadDir, zipWalker)
		case LHA:
			err = filepath.WalkDir(downloadDir, lhaWalker)
		case Arc:
			err = filepath.WalkDir(downloadDir, arcWalker)
		case Arj:
			err = filepath.WalkDir(downloadDir, arjWalker)
		}
		if err != nil {
			logger.Errorf("walk directory %s: %s", err, downloadDir)
		}
	}
	logger.Infof("Completed UUID archive checks in %s", time.Since(tick))
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

func (r Repair) artifacts(ctx context.Context, ce boil.ContextExecutor, logger *zap.SugaredLogger) ([]string, error) {
	var files models.FileSlice
	var err error
	switch r {
	case Zip:
		files, err = fixzip.Files(ctx, ce)
	case LHA:
		files, err = fixlha.Files(ctx, ce)
	case Arc:
		files, err = fixarc.Files(ctx, ce)
	case Arj:
		files, err = fixarj.Files(ctx, ce)
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

func (r Repair) rearchive(ctx context.Context, path, extra, uid string) error {
	logger := helper.Logger(ctx)
	tmp, err := os.MkdirTemp(helper.TmpDir(), "rearchive-")
	if err != nil {
		return fmt.Errorf("rearchive mkdir temp %w: %s", err, path)
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
	cmd := exec.Command(extractCmd, extractArg, path)
	cmd.Dir = tmp
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("rearchive run %w: %s", err, path)
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

	finalArc := filepath.Join(extra, basename)
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
func (c Config) Assets(ctx context.Context, ce boil.ContextExecutor) error {
	if ce == nil {
		return nil
	}
	tick := time.Now()
	logger := helper.Logger(ctx)
	mods := []qm.QueryMod{}
	mods = append(mods, qm.Select("uuid"))
	mods = append(mods, qm.WithDeleted())
	files, err := models.Files(mods...).All(ctx, ce)
	if err != nil {
		return fmt.Errorf("config repair select all uuids: %w", err)
	}
	size := len(files)
	logger.Infof("Checking %d UUIDs", size)
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
					unknownAsset(logger, path, c.AbsOrphaned, d.Name(), uid)
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
	for _, count := range counters {
		sum += count
	}
	logger.Infof("Checked %d files for %d UUIDs in %s", sum, size, time.Since(tick))
	return nil
}

// unknownAsset logs a warning message for an unknown asset file.
func unknownAsset(logger *zap.SugaredLogger, oldpath, orphanedDir, name, uid string) {
	logger.Warnf("Unknown file: %s, no matching artifact for UUID: %q", name, uid)
	defer func() {
		now := time.Now().Format("2006-01-02_15-04-05")
		dest := filepath.Join(orphanedDir, fmt.Sprintf("%s_%s", now, name))
		if err := helper.RenameCrossDevice(oldpath, dest); err != nil {
			logger.Errorf("could not move orphaned artifact asset for %q: %s", name, err)
		}
	}()
}

// RepairAssets, on startup check the file system directories for any invalid or unknown files.
// If any are found, they are removed without warning.
func (c Config) RepairAssets(ctx context.Context, exec boil.ContextExecutor) error {
	logger := helper.Logger(ctx)
	backupDir := c.AbsOrphaned
	if st, err := os.Stat(backupDir); err != nil {
		return fmt.Errorf("repair backup directory %w: %s", err, backupDir)
	} else if !st.IsDir() {
		return fmt.Errorf("repair backup directory %w: %s", ErrNotDir, backupDir)
	}
	if err := ImageDirs(logger, c); err != nil {
		return fmt.Errorf("repair the images directories %w", err)
	}
	if err := DownloadDir(logger, c.AbsDownload, c.AbsOrphaned, c.AbsExtra); err != nil {
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
	return nil
}

// MagicNumbers checks the magic numbers of the artifacts and replaces any missing or
// legacy values with the current method of detection. Previous detection methods were
// done using the `file` command line utility, which is a bit to verbose for our needs.
func (c Config) MagicNumbers(ctx context.Context, ce boil.ContextExecutor, logger *zap.SugaredLogger) error {
	tick := time.Now()
	r := model.Artifacts{}
	magics, err := r.ByMagicErr(ctx, ce, false)
	if err != nil {
		return fmt.Errorf("magicnumbers %w", err)
	}
	const large = 1000
	if len(magics) > large && logger != nil {
		logger.Warnf("Checking %d magic number values for artifacts, this could take a while", len(magics))
	}
	count := 0
	for _, v := range magics {
		name := filepath.Join(c.AbsDownload, v.UUID.String)
		r, err := os.Open(name)
		if err != nil {
			_ = r.Close()
			continue
		}
		magic := magicnumber.Find(r)
		count++
		_ = model.UpdateMagic(ctx, ce, v.ID, magic.Title())
		_ = r.Close()
	}
	if count == 0 || logger == nil {
		return nil
	}
	logger.Infof("Updated %d magic number values for artifacts in %s", count, time.Since(tick))
	return nil
}

// Previews, on startup check the preview directory for any unnecessary preview images such as textfile artifacts.
func (c Config) Previews(ctx context.Context, ce boil.ContextExecutor, logger *zap.SugaredLogger) error {
	r := model.Artifacts{}
	artifacts, err := r.ByTextPlatform(ctx, ce)
	if err != nil {
		return fmt.Errorf("nopreview %w", err)
	}
	var count, totals int64
	for _, v := range artifacts {
		png := filepath.Join(c.AbsPreview, v.UUID.String) + ".png"
		st, err := os.Stat(png)
		if err != nil {
			fmt.Fprintln(io.Discard, err)
			continue
		}
		_ = os.Remove(png)
		count++
		totals += st.Size()
	}
	for _, v := range artifacts {
		webp := filepath.Join(c.AbsPreview, v.UUID.String) + ".webp"
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
func ImageDirs(logger *zap.SugaredLogger, c Config) error {
	backupDir := c.AbsOrphaned
	dirs := []string{c.AbsPreview, c.AbsThumbnail}
	if err := removeSub(dirs...); err != nil {
		return fmt.Errorf("remove subdirectories %w", err)
	}
	// remove any invalid files
	p, t := 0, 0
	for _, dir := range dirs {
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
			return RemoveImage(name, path, backupDir)
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
	for _, dir := range dirs {
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
func DownloadDir(logger *zap.SugaredLogger, srcDir, destDir, extraDir string) error {
	if srcDir == "" || destDir == "" || extraDir == "" {
		return fmt.Errorf("%w: %s %s", ErrEmpty, srcDir, destDir)
	}
	count := 0
	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk path %w: %s", err, path)
		}
		name := d.Name()
		if d.IsDir() {
			return RemoveDir(name, path, srcDir)
		}
		if err = RemoveDownload(name, path, destDir, extraDir); err != nil {
			return fmt.Errorf("remove download: %w", err)
		}
		if filepath.Ext(name) == "" {
			count++
		}
		return RenameDownload(name, path)
	})
	if err != nil {
		return fmt.Errorf("walk directory %w: %s", err, srcDir)
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
func RemoveDownload(basename, path, destDir, extraDir string) error {
	if basename == "" || path == "" || destDir == "" || extraDir == "" {
		return fmt.Errorf("remove download %w: %s %s %s %s",
			ErrEmpty, basename, path, destDir, extraDir)
	}
	const filedownload = ""
	ext := filepath.Ext(basename)
	switch ext {
	case filedownload:
		return nil
	case ".txt", ".zip", ".chiptune":
		rename(path, "rename valid ext", filepath.Join(extraDir, basename))
	default:
		remove(basename, "remove invalid ext", path, destDir)
	}
	return nil
}

// RemoveImage, check the image files for invalid names and extensions.
// If any are found, they are moved to the destDir without warning.
// Basename must be the name of the file with a valid file extension.
//
// Valid file extensions are .png and .webp, and basename must be a
// valid uuid or cfid with the correct length.
func RemoveImage(basename, path, destDir string) error {
	if basename == "" || path == "" || destDir == "" {
		return fmt.Errorf("remove image %w: %s %s %s", ErrEmpty, basename, path, destDir)
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
		}
		if err := uuid.Validate(filename); err != nil {
			remove(basename, "remove invalid uuid", path, destDir)
			return nil //nolint:nilerr
		}
	}
	switch ext {
	case png, webp:
		return nil
	default:
		remove(basename, "remove invalid ext", path, destDir)
	}
	return nil
}

// remove the file without warning.
func remove(name, info, path, destDir string) {
	w := os.Stderr
	fmt.Fprintf(w, "%s: %s\n", info, name)
	defer func() {
		now := time.Now().Format("2006-01-02_15-04-05")
		dest := filepath.Join(destDir, fmt.Sprintf("%s_%s", now, name))
		err := helper.RenameCrossDevice(path, dest)
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
