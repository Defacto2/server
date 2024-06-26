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

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/magicnumber/pkzip"
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

var (
	ErrCtxLog = errors.New("context logger is invalid")
	ErrIsDir  = errors.New("is directory")
	ErrEmpty  = errors.New("empty path or name")
)

func zipfiles(ctx context.Context, ce boil.ContextExecutor) (models.FileSlice, error) {
	mods := []qm.QueryMod{}
	mods = append(mods, qm.Select("uuid"))
	mods = append(mods, qm.Where("platform = ?", tags.DOS.String()))
	mods = append(mods, qm.Where("filename ILIKE ?", "%.zip"))
	files, err := models.Files(mods...).All(ctx, ce)
	if err != nil {
		return nil, fmt.Errorf("select all uuids: %w", err)
	}
	return files, nil
}

func walkPath(ctx context.Context, path, extra string, d fs.DirEntry, artifacts ...string) string {
	if d.IsDir() {
		return ""
	}
	uid := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
	if _, found := slices.BinarySearch(artifacts, uid); !found {
		return ""
	}
	logger := helper.Logger(ctx)
	if f, err := os.Stat(filepath.Join(extra, uid)); err == nil && !f.IsDir() {
		logger.Info("Found extra file:", uid)
		return ""
	}
	methods, err := pkzip.Methods(path)
	if err != nil {
		logger.Errorf("pkzip methods %w: %s", err, path)
		return ""
	}
	usable := true
	for _, method := range methods {
		if !method.Zip() {
			usable = false
			break
		}
	}
	if usable {
		return ""
	}
	return uid
}

func zipPassTest(ctx context.Context, path string) bool {
	logger := helper.Logger(ctx)
	_, err := exec.Command("/usr/bin/unzip", "-t", path).Output()
	if err != nil {
		diag := pkzip.ExitStatus(err)
		switch diag {
		case pkzip.Normal, pkzip.Warning:
			// normal or warnings are fine
			return true
		case pkzip.CompressionMethod:
			// cannot do anything about this
			return true
		}
		logger.Errorf("unzip -t %s: %s", diag, path)
	}
	return false
}

func zipReArchive(ctx context.Context, path, extra, uid string) error {
	logger := helper.Logger(ctx)
	tmp, err := os.MkdirTemp(os.TempDir(), "defacto2-rezip-")
	if err != nil {
		return fmt.Errorf("os.MkdirTemp %w: %s", err, path)
	}
	defer os.RemoveAll(tmp)
	err = exec.Command("/usr/bin/unzip", path, "-d", tmp).Run()
	if err != nil {
		return fmt.Errorf("unzip -o %w: %s", err, path)
	}
	c, err := helper.Count(tmp)
	if err != nil {
		return fmt.Errorf("helper.Count %w: %s", err, tmp)
	}
	logger.Infof("Rezipped %d files for %s found in: %s", c, uid, tmp)
	st, err := os.Stat(tmp)
	if err != nil {
		return fmt.Errorf("os.Stat %w: %s", err, tmp)
	}

	// TODO rezip to new archive

	// TODO helper duplicate

	logger.Infof("Rezipped %d bytes for %s found in: %s", st.Size(), uid, tmp)
	// unzip [-Z] [-cflptTuvz[abjnoqsCDKLMUVWX$/:^]] file[.zip] [file(s) ...]  [-x xfile(s) ...] [-d exdir]
	// /usr/bin/unzip path -d os.MkdirAll(filepath.Join(c.AbsExtra, uid), 0755)
	return nil
}

func (c Config) pkzips(ctx context.Context, ce boil.ContextExecutor) error { //nolint:funlen,gocognit
	if ce == nil {
		return nil
	}
	tick := time.Now()
	logger := helper.Logger(ctx)

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
		uid := walkPath(ctx, path, c.AbsExtra, d, artifacts...)
		if uid == "" {
			return nil
		}
		if zipPassTest(ctx, path) {
			return nil
		}
		if err := zipReArchive(ctx, path, c.AbsExtra, uid); err != nil {
			logger.Errorf("zip repair and re-archive %w", err)
			return nil
		}
		sum++
		return nil
	})
	if err != nil {
		logger.Errorf("walk directory %w: %s", err, dir)
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
func (c Config) assets(ctx context.Context, ce boil.ContextExecutor) error {
	if ce == nil {
		return nil
	}
	tick := time.Now()
	logger := helper.Logger(ctx)
	mods := []qm.QueryMod{}
	mods = append(mods, qm.Select("uuid"))
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

// RepairFS, on startup check the file system directories for any invalid or unknown files.
// If any are found, they are removed without warning.
func (c Config) RepairFS(ctx context.Context, exec boil.ContextExecutor) error {
	logger := helper.Logger(ctx)
	backupDir := c.AbsOrphaned
	if st, err := os.Stat(backupDir); err != nil {
		return fmt.Errorf("repair fs backup directory %w: %s", err, backupDir)
	} else if !st.IsDir() {
		return fmt.Errorf("repair fs backup directory %w: %s", ErrNotDir, backupDir)
	}
	if err := ImagesFS(logger, c); err != nil {
		return fmt.Errorf("repair fs images %w", err)
	}
	if err := DownloadFS(logger, c.AbsDownload, c.AbsOrphaned, c.AbsExtra); err != nil {
		return fmt.Errorf("repair fs downloads %w", err)
	}
	if err := c.assets(ctx, exec); err != nil {
		return fmt.Errorf("repair fs assets %w", err)
	}
	if err := c.pkzips(ctx, exec); err != nil {
		return fmt.Errorf("repair fs pkzips %w", err)
	}
	return nil
}

func ImagesFS(logger *zap.SugaredLogger, c Config) error {
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

func containsInfo(logger *zap.SugaredLogger, name string, count int) {
	if logger == nil {
		return
	}
	if MinimumFiles > count {
		logger.Warnf("The %s directory contains %d files, which is less than the minimum of %d", name, count, MinimumFiles)
		return
	}
	logger.Infof("The %s directory contains %d files", name, count)
}

// DownloadFS, on startup check the download directory for any invalid or unknown files.
func DownloadFS(logger *zap.SugaredLogger, srcDir, destDir, extraDir string) error {
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

func rename(oldpath, info, newpath string) {
	w := os.Stderr
	fmt.Fprintf(w, "%s: %s\n", info, oldpath)
	defer func() {
		err := os.Rename(oldpath, newpath)
		if err != nil {
			fmt.Fprintf(w, "defer repair file rename: %s\n", err)
		}
	}()
}
