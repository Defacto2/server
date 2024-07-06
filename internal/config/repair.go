package config

// Package file repair.go contains the repair functions for assets and downloads.

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Defacto2/server/internal/archive/pkzip"
	"github.com/Defacto2/server/internal/archive/rezip"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
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

// zipfiles returns all the DOS platform artifacts using a .zip extension filename.
func zipfiles(ctx context.Context, ce boil.ContextExecutor) (models.FileSlice, error) {
	mods := []qm.QueryMod{}
	mods = append(mods, qm.Select("uuid"))
	mods = append(mods, qm.Where("platform = ?", tags.DOS.String()))
	mods = append(mods, qm.Where("filename ILIKE ?", "%.zip"))
	mods = append(mods, qm.WithDeleted())
	files, err := models.Files(mods...).All(ctx, ce)
	if err != nil {
		return nil, fmt.Errorf("select all uuids: %w", err)
	}
	return files, nil
}

// requireRearchive returns the UUID of the zipped file if it requires re-archiving because it uses a
// legacy compression method that is not supported by Go or JS libraries.
//
// Rearchived UUID named files are moved to the extra directory and are given a .zip extension.
func requireRearchive(ctx context.Context, path, extra string, d fs.DirEntry, artifacts ...string) string {
	if d.IsDir() {
		return ""
	}
	if ext := filepath.Ext(strings.ToLower(d.Name())); ext != ".zip" && ext != "" {
		return ""
	}
	uid := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
	if _, found := slices.BinarySearch(artifacts, uid); !found {
		return ""
	}
	logger := helper.Logger(ctx)
	extraZip := filepath.Join(extra, uid+".zip")
	if f, err := os.Stat(extraZip); err == nil && !f.IsDir() {
		return ""
	}
	methods, err := pkzip.Methods(path)
	if err != nil {
		logger.Errorf("%s: %s", err, path)
		return ""
	}
	for _, method := range methods {
		if !method.Zip() {
			return uid
		}
	}
	return ""
}

// hwZipTest returns true if the zip file fails the hwzip list command.
// The path is the path to the zip file.
func hwZipTest(ctx context.Context, path string) bool {
	logger := helper.Logger(ctx)
	z, err := exec.Command(command.HWZip, "list", path).Output()
	if err != nil {
		logger.Errorf("hwzip list %s: %s", err, path)
		return true
	}
	if !strings.Contains(string(z), "Failed to parse ") {
		return true
	}
	return false

}

// hwCompress uses the [hwzip application] by Hans Wennborg to extract zip archives
// using antiquated compression methods that are not supported by Go, JS or other
// Linux utilities. The extracted files are then re-archived using Go and moved
// to the extra directory with a .zip extension.
//
// [hwzip application]: https://www.hanshq.net/zip.html
func hwCompress(ctx context.Context, path, extra, uid string) error {
	logger := helper.Logger(ctx)
	tmp, err := os.MkdirTemp(os.TempDir(), "defacto2-rezip-")
	if err != nil {
		return fmt.Errorf("hwcompress mkdir temp %w: %s", err, path)
	}
	defer os.RemoveAll(tmp)

	cmd := exec.Command(command.HWZip, "extract", path)
	cmd.Dir = tmp
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("hwzip run %w: %s", err, path)
	}

	c, err := helper.Count(tmp)
	if err != nil {
		return fmt.Errorf("hwcompress tmp count %w: %s", err, tmp)
	}
	logger.Infof("Rezipped %d files for %s found in: %s", c, uid, tmp)
	_, err = os.Stat(tmp)
	if err != nil {
		return fmt.Errorf("hwcompress tmp stat %w: %s", err, tmp)
	}

	basename := uid + ".zip"
	tmpZip := filepath.Join(os.TempDir(), basename)
	if written, err := rezip.CompressDir(tmp, tmpZip); err != nil {
		return fmt.Errorf("hwcompress compress dir %w: %s", err, tmp)
	} else if written == 0 {
		return nil
	}

	finalZip := filepath.Join(extra, basename)
	if err = helper.RenameCrossDevice(tmpZip, finalZip); err != nil {
		defer os.RemoveAll(tmpZip)
		return fmt.Errorf("hwcompress rename %w: %s", err, tmpZip)
	}

	st, err := os.Stat(finalZip)
	if err != nil {
		return fmt.Errorf("hwcompress zip stat %w: %s", err, finalZip)
	}
	logger.Infof("Extra deflated zipfile created %d bytes: %s", st.Size(), finalZip)
	return nil
}

// ImplodedZips checks the DOS platform artifacts for any zip files that require re-archiving.
// These are identified by the use of a legacy compression method that is not supported by Go or JS libraries.
// The re-archived files are stored the extra directory and can be used by js-dos and other tools.
func (c Config) ImplodedZips(ctx context.Context, ce boil.ContextExecutor) error {
	if ce == nil {
		return nil
	}
	tick := time.Now()
	logger := helper.Logger(ctx)

	if _, err := exec.LookPath(command.HWZip); err != nil {
		return fmt.Errorf("cannot find hwzip executable: %w", err)
	}
	files, err := zipfiles(ctx, ce)
	if err != nil {
		return fmt.Errorf("config pkzips zipfiles, %w", err)
	}
	size := len(files)
	logger.Infof("Checking %d %s UUIDs", size, tags.DOS.String())
	artifacts := make([]string, size)
	for i, f := range files {
		if !f.UUID.Valid || f.UUID.String == "" {
			continue
		}
		artifacts[i] = f.UUID.String
	}
	artifacts = slices.Clip(artifacts)
	slices.Sort(artifacts)

	sum := 0
	dir := c.AbsDownload
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk path %w: %s", err, path)
		}
		uid := requireRearchive(ctx, path, c.AbsExtra, d, artifacts...)
		if uid == "" {
			return nil
		}
		if !hwZipTest(ctx, path) {
			return nil
		}
		if err := hwCompress(ctx, path, c.AbsExtra, uid); err != nil {
			logger.Errorf("zip repair and re-archive: %s", err)
			return nil
		}
		sum++
		return nil
	})
	if err != nil {
		logger.Errorf("walk directory %w: %s", err, dir)
	}
	if sum == 0 {
		logger.Infof("No files were re-archived")
		return nil
	}
	logger.Infof("Checked %d files for %d UUIDs in %s", sum, size, time.Since(tick))
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
	if err := c.ImplodedZips(ctx, exec); err != nil {
		return fmt.Errorf("repair imploded zips %w", err)
	}
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

	newname, _ := helper.CFToUUID(rawname)
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
			filename, _ = helper.CFToUUID(filename)
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
