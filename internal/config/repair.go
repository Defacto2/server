package config

// Package file repair.go contains the repair functions for assets and downloads.

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Defacto2/server/internal/helper"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	unid = "00000000-0000-0000-0000-000000000000" // common universal unique identifier example
	cfid = "00000000-0000-0000-0000000000000000"  // coldfusion uuid example
)

var (
	ErrIsDir = errors.New("is directory")
	ErrEmpty = errors.New("empty path or name")
)

// RepairFS, on startup check the file system directories for any invalid or unknown files.
// If any are found, they are removed without warning.
func (c Config) RepairFS(logger *zap.SugaredLogger) error {
	if logger == nil {
		return ErrZap
	}
	backupDir := c.AbsOrphaned
	if st, err := os.Stat(backupDir); err != nil {
		return fmt.Errorf("repair fs backup directory %w: %s", err, backupDir)
	} else if !st.IsDir() {
		return fmt.Errorf("repair fs backup directory %w: %s", ErrNotDir, backupDir)
	}
	dirs := []string{c.AbsPreview, c.AbsThumbnail}
	p, t := 0, 0
	for _, dir := range dirs {
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return fmt.Errorf("repair fs walk path %w: %s", err, path)
			}
			name := d.Name()
			if d.IsDir() {
				return RemoveDir(name, path, dir)
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
			return fmt.Errorf("repair fs walk directory %w: %s", err, dir)
		}
		switch dir {
		case c.AbsPreview:
			containsInfo(logger, "preview", p)
		case c.AbsThumbnail:
			containsInfo(logger, "thumb", t)
		}
	}
	return DownloadFS(logger, c.AbsDownload, c.AbsOrphaned)
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
func DownloadFS(logger *zap.SugaredLogger, srcDir, destDir string) error {
	if srcDir == "" || destDir == "" {
		return fmt.Errorf("download fs %w: %s %s", ErrEmpty, srcDir, destDir)
	}

	count := 0
	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("download fs walk path %w: %s", err, path)
		}
		name := d.Name()
		if d.IsDir() {
			return RemoveDir(name, path, srcDir)
		}
		if err = RemoveDownload(name, path, destDir); err != nil {
			return fmt.Errorf("RemoveDownload: %w", err)
		}
		if filepath.Ext(name) == "" {
			count++
		}
		return RenameDownload(name, path)
	})
	if err != nil {
		return fmt.Errorf("download fs walk directory %w: %s", err, srcDir)
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
	const syncthing = ".stfolder"
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
func RemoveDownload(basename, path, destDir string) error {
	if basename == "" || path == "" || destDir == "" {
		return fmt.Errorf("remove download %w: %s %s %s", ErrEmpty, basename, path, destDir)
	}
	const filedownload = ""
	ext := filepath.Ext(basename)
	switch ext {
	case filedownload, ".chiptune", ".txt", ".zip":
		return nil
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
		dest := filepath.Join(destDir, fmt.Sprintf("%s_%s", name, now))
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
