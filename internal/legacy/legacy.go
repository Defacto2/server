// Package legacy provides a single source of truth for MS-DOS era legacy
// archive formats shared across js-dos and artifact repacking helpers.
//
//nolint:gochecknoglobals
package legacy

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/Defacto2/magicnumber"
)

// archives lists the DOS and BBS era archive signatures that are
// handled as legacy formats for repacking and emulation.
//
// Even though the ZIP Deflate method dates back to 1991,
// it is skipped here as it is not a legacy method. And while
// formats such as RAR v1.3 date to 1993, the .RAR file extension
// is almost always associated with Windows, so it is not included here.
var archives = [...]magicnumber.Signature{
	magicnumber.ARChiveSEA,
	magicnumber.ArchiveRobertJung,
	magicnumber.LibraryUtility,
	magicnumber.NoGatePAK,
	magicnumber.PKWAREZipImplode,
	magicnumber.PKWAREZipReduce,
	magicnumber.PKWAREZipShrink,
	magicnumber.YoshiLHA,
}

var (
	exts   = initExts()
	titles = initTitles()
)

func initExts() []string {
	exts := magicnumber.Exts()
	s := make([]string, 0, len(archives))
	for _, sig := range archives {
		for _, ext := range exts[sig] {
			if !slices.Contains(s, ext) {
				s = append(s, ext)
			}
		}
	}
	return s
}

func initTitles() []string {
	s := make([]string, 0, len(archives))
	for _, sig := range archives {
		s = append(s, sig.Title())
	}
	return s
}

// ArchiveExt reports whether filename has a known MS-DOS archive extension.
func ArchiveExt(filename string) bool {
	if filename == "" {
		return false
	}
	ext := strings.ToLower(filepath.Ext(filename))
	return slices.Contains(exts, ext)
}

// Archive reports whether magic is a magicnumber title string for a
// legacy MS-DOS archive format that should be repacked as Deflate zip.
func Archive(magic any) bool {
	val, ok := magic.(string)
	if !ok {
		return false
	}
	return slices.Contains(titles, val)
}
