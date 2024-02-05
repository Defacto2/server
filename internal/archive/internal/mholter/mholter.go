// Package mholter handles interactions with the mholt archiver package. This should
// be used as the primary means of interacting with file archives before using
// external archive programs.
package mholter

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver"
)

var (
	ErrNoCustom = errors.New("does not support customization")
	ErrWalkrFmt = errors.New("format specified by archive filename is not a walker format")
	ErrPanic    = errors.New("mholt panic with the archive")
)

// Configure interface for the archiver,
// a cross-platform, multi-format archive utility and Go library.
func Configure(f any) error {
	tarCfg := &archiver.Tar{
		OverwriteExisting:      true,
		MkdirAll:               true,
		ImplicitTopLevelFolder: false,
		ContinueOnError:        false,
	}
	switch v := f.(type) {
	case *archiver.Rar:
		// options: https://pkg.go.dev/github.com/mholt/archiver?tab=doc#Rar
		v.OverwriteExisting = true
		v.MkdirAll = true
		v.ImplicitTopLevelFolder = false
		v.ContinueOnError = false
	case *archiver.Tar:
		// options: https://pkg.go.dev/github.com/mholt/archiver?tab=doc#Tar
		v.OverwriteExisting = true
		v.MkdirAll = true
		v.ImplicitTopLevelFolder = false
		v.ContinueOnError = false
	case *archiver.TarBz2:
		v.Tar = tarCfg
	case *archiver.TarGz:
		v.Tar = tarCfg
	case *archiver.TarLz4:
		v.Tar = tarCfg
	case *archiver.TarSz:
		v.Tar = tarCfg
	case *archiver.TarXz:
		v.Tar = tarCfg
	case *archiver.Zip:
		// options: https://pkg.go.dev/github.com/mholt/archiver?tab=doc#Zip
		v.OverwriteExisting = true
		v.MkdirAll = true
		v.SelectiveCompression = true
		v.ImplicitTopLevelFolder = false
		v.ContinueOnError = false
	case *archiver.Gz,
		*archiver.Bz2,
		*archiver.Lz4,
		*archiver.Snappy,
		*archiver.Xz:
		return nil
	default:
		return fmt.Errorf("arc configure %v: %w", f, ErrNoCustom)
	}
	return nil
}

// Walkr calls walkFn for each file within the given archive file.
// The archive format is chosen implicitly.
// Archiver relies on the filename extension to determine which
// decompression format to use, which must be supplied using filename.
func Walkr(src, filename string, walkFn archiver.WalkFunc) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%w %s while extracting %s: %v", ErrPanic, filepath.Base(src), filename, e)
		}
	}()
	name := strings.ToLower(filename)
	a, err := archiver.ByExtension(name)
	if err != nil {
		return fmt.Errorf("arc walkr byextension %q: %w", name, err)
	}
	w, ok := a.(archiver.Walker)
	if !ok {
		return fmt.Errorf("arc walkr %s (%T): %w", name, a, ErrWalkrFmt)
	}
	if err := w.Walk(src, walkFn); err != nil {
		return fmt.Errorf("arc walkr %q: %w", filepath.Base(src), err)
	}
	return nil
}
