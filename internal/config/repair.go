package config

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

const (
	uuid = "00000000-0000-0000-0000-000000000000" // common universal unique identifier example
	cfid = "00000000-0000-0000-0000000000000000"  // coldfusion uuid example
)

// RepairFS, on startup check the file system directories for any invalid or unknown files.
// If any are found, they are removed without warning.
func (c Config) RepairFS(z *zap.SugaredLogger) error {
	if z == nil {
		return ErrZap
	}
	dirs := []string{c.PreviewDir, c.ThumbnailDir}
	p, t := 0, 0
	for _, dir := range dirs {
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			name := info.Name()
			if info.IsDir() {
				return directories(name, path, dir)
			}
			switch dir {
			case c.PreviewDir:
				if filepath.Ext(name) != ".webp" {
					p++
				}
			case c.ThumbnailDir:
				if filepath.Ext(name) != ".webp" {
					t++
				}
			}
			return images(name, path)
		})
		if err != nil {
			return err
		}
		switch dir {
		case c.PreviewDir:
			z.Infof("The preview directory contains, %d images: %s", p, dir)
		case c.ThumbnailDir:
			z.Infof("The thumb directory contains, %d images: %s", t, dir)
		}
	}
	return c.downloadFS(z)
}

func (c Config) downloadFS(z *zap.SugaredLogger) error {
	dir := c.DownloadDir
	if _, err := os.Stat(dir); err != nil {
		var ignore error
		return ignore //nolint:nilerr
	}
	d := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		name := info.Name()
		if filepath.Ext(name) == "" {
			d++
		}
		if info.IsDir() {
			return directories(name, path, dir)
		}
		return downloads(name, path)
	})
	if err != nil {
		return err
	}
	z.Infof("The downloads directory contains, %d files: %s", d, dir)
	return nil
}

func directories(name, path, dir string) error {
	const st = ".stfolder" // st is a syncthing directory
	switch name {
	case filepath.Base(dir):
		// skip the root directory
	case st:
		defer os.RemoveAll(path)
	default:
		fmt.Fprintln(os.Stderr, "unknown dir:", path)
	}
	return nil // always skip
}

func downloads(name, path string) error {
	l := len(name)
	switch filepath.Ext(name) {
	case ".chiptune", ".txt":
		return nil
	case ".zip":
		if l != len(uuid)+4 && l != len(cfid)+4 {
			rm(name, "remove", path)
		}
		return nil
	default:
		if l != len(uuid) && l != len(cfid) {
			rm(name, "unknown", path)
		}
	}
	return nil
}

func images(name, path string) error {
	const (
		png  = ".png"    // png file extension
		webp = ".webp"   // webp file extension
		lpng = len(png)  // length of png file extension
		lweb = len(webp) // length of webp file extension
	)
	ext := filepath.Ext(name)
	l := len(name)
	switch ext {
	case png:
		if l != len(uuid)+lpng && l != len(cfid)+lpng {
			rm(name, "remove", path)
		}
	case webp:
		if l != len(uuid)+lweb && l != len(cfid)+lweb {
			rm(name, "remove", path)
		}
	default:
		rm(name, "unknown", path)
	}
	return nil
}

func rm(name, info, path string) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", info, name)
	defer os.Remove(path)
}
