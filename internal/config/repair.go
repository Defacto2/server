package config

// Package file repair.go contains the repair functions for assets and downloads.

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

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
)

// RepairFS, on startup check the file system directories for any invalid or unknown files.
// If any are found, they are removed without warning.
func (c Config) RepairFS(logger *zap.SugaredLogger) error {
	if logger == nil {
		return ErrZap
	}
	dirs := []string{c.AbsPreview, c.AbsThumbnail}
	p, t := 0, 0
	for _, dir := range dirs {
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return fmt.Errorf("walk: %w", err)
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
			return RemoveImage(name, path)
		})
		if err != nil {
			return fmt.Errorf("filepath.Walk: %w", err)
		}
		switch dir {
		case c.AbsPreview:
			logger.Infof("The preview directory contains, %d images: %s", p, dir)
		case c.AbsThumbnail:
			logger.Infof("The thumb directory contains, %d images: %s", t, dir)
		}
	}
	return DownloadFS(logger, c.AbsDownload)
}

// DownloadFS, on startup check the download directory for any invalid or unknown files.
func DownloadFS(logger *zap.SugaredLogger, dir string) error {
	if _, err := os.Stat(dir); err != nil {
		var exit error
		return exit //nolint:nilerr
	}
	count := 0
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("filepath.WalkDir: %w", err)
		}
		name := d.Name()
		if d.IsDir() {
			return RemoveDir(name, path, dir)
		}
		if err = RemoveDownload(name, path); err != nil {
			return fmt.Errorf("RemoveDownload: %w", err)
		}
		if filepath.Ext(name) == "" {
			count++
		}
		return RenameDownload(name, path)
	})
	if err != nil {
		return fmt.Errorf("filepath.WalkDir: %w", err)
	}
	if logger != nil {
		logger.Infof("The downloads directory contains, %d files: %s", count, dir)
	}
	return nil
}

// RenameDownload, rename the download file if the basename uses an invalid coldfusion uuid.
func RenameDownload(basename, absPath string) error {
	st, err := os.Stat(absPath)
	if err != nil {
		return nil
	}
	if st.IsDir() {
		return fmt.Errorf("%w: %s", ErrIsDir, absPath)
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
func RemoveDownload(basename, path string) error {
	st, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("os.Stat: %w", err)
	}
	if st.IsDir() {
		return fmt.Errorf("%w: %s", ErrIsDir, path)
	}
	const filedownload = ""
	ext := filepath.Ext(basename)
	switch ext {
	case filedownload, ".chiptune", ".txt", ".zip":
		return nil
	default:
		remove(basename, "remove invalid ext", path)
	}
	return nil
}

// RemoveImage, check the image files for invalid names and extensions.
// If any are found, they are removed without warning.
// Basename must be the name of the file with a valid file extension.
//
// Valid file extensions are .png and .webp, and basename must be a
// valid uuid or cfid with the correct length.
func RemoveImage(basename, path string) error {
	st, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("os.Stat: %w", err)
	}
	if st.IsDir() {
		return fmt.Errorf("%w: %s", ErrIsDir, path)
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
			remove(basename, "remove invalid uuid", path)
			return nil //nolint:nilerr
		}
	}
	switch ext {
	case png, webp:
		return nil
	default:
		remove(basename, "remove invalid ext", path)
	}
	return nil
}

func remove(name, info, path string) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", info, name)
	defer os.Remove(path)
}

func rename(oldpath, info, newpath string) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", info, oldpath)
	defer os.Rename(oldpath, newpath)
}
