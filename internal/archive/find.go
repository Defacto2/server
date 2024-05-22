package archive

// Package file archive/find.go contains the filename search and matching functions.

import (
	"cmp"
	"path/filepath"
	"slices"
	"strings"
)

// Finds are a collection of matched filenames and their usability ranking.
type Finds map[string]Usability

// BestMatch returns the most usable filename from a collection of finds.
func (f Finds) BestMatch() string {
	if len(f) == 0 {
		return ""
	}
	type match struct {
		Filename  string
		Usability Usability
	}
	matches := make([]match, len(f))
	i := 0
	for k, v := range f {
		matches[i] = match{k, v}
		i++
	}
	slices.SortStableFunc(matches, func(a, b match) int {
		return cmp.Compare(a.Usability, b.Usability)
	})
	for _, m := range matches {
		return m.Filename // return first result
	}
	return ""
}

const (
	diz = ".diz"
	nfo = ".nfo"
	txt = ".txt"
)

// Readme returns the best matching scene text README or NFO file from a collection of files.
// The filename is the name of the archive file, and the files are the list of files in the archive.
// Note the filename matches are case-insensitive as many handled file archives are
// created on Windows FAT32, NTFS or MS-DOS FAT16 file systems.
func Readme(filename string, files ...string) string {
	f := make(Finds)
	for _, file := range files {
		name := strings.ToLower(file)
		base := strings.ToLower(strings.TrimSuffix(filename, filepath.Ext(filename)))
		ext := strings.ToLower(filepath.Ext(name))
		switch ext {
		case diz, nfo, txt:
			// okay
		default:
			continue
		}
		f = matchs(file, name, base, f)
	}
	return f.BestMatch()
}

func matchs(file, name, base string, f Finds) Finds {
	ext := strings.ToLower(filepath.Ext(name))
	switch {
	case name == base+nfo:
		// [archive name].nfo
		f[file] = Lvl1
	case name == base+txt:
		// [archive name].txt
		f[file] = Lvl2
	case ext == nfo:
		// [random].nfo
		f[file] = Lvl3
	case name == "file_id.diz":
		// BBS file description
		f[file] = Lvl4
	case name == base+diz:
		// [archive name].diz
		f[file] = Lvl5
	case name == txt:
		// [random].txt
		f[file] = Lvl6
	case name == diz:
		// [random].diz
		f[file] = Lvl7
	default:
		// currently lacking is [group name].nfo and [group name].txt priorities
	}
	return f
}

// Usability of search, filename pattern matches.
type Usability uint

const (
	// Lvl1 is the highest usability.
	Lvl1 Usability = iota + 1
	Lvl2
	Lvl3
	Lvl4
	Lvl5
	Lvl6
	Lvl7
	Lvl8
	Lvl9 // Lvl9 is the least usable.
)
