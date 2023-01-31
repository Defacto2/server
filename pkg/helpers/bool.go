package helpers

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// bool.go are funcs that return a boolean.

// Finds returns true if the name is found in the collection of names.
func Finds(name string, names ...string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}

// IsStat stats the named file or directory to confirm it exists on the system.
func IsStat(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return true
}

// IsDay returns true if the i value can be used as a day time value.
func IsDay(i int) bool {
	const maxDay = 31
	if i > 0 && i <= maxDay {
		return true
	}
	return false
}

// IsYear returns true  if the i value is greater than 1969
// or equal to the current year.
func IsYear(i int) bool {
	const unix = 1970
	now := time.Now().Year()
	if i >= unix && i <= now {
		return true
	}
	return false
}

// IsApp returns true if the named file uses a Windows application filename.
func IsApp(name string) bool {
	s := []string{com, exe}
	return IsExt(name, s...)
}

// IsApp returns true if the named file uses a common compressed or archived filename.
func IsArchive(name string) bool {
	s := []string{z7, arc, ark, arj, cab, gz, lha, lzh, rar, tar, tgz, zip}
	return IsExt(name, s...)
}

// IsDocument returns true if the named file uses a common document or text filename.
func IsDocument(name string) bool {
	s := []string{fst, asc, ans, cpt, diz, doc, dox, me, nfo, pcb, pdf, txt, unp}
	return IsExt(name, s...)
}

// IsImage returns true if the named file uses a common image or photo filename.
func IsImage(name string) bool {
	s := []string{bmp, gif, ico, iff, jpg, jpeg, lbm, png, pcx}
	return IsExt(name, s...)
}

// IsHTML returns true if the named file uses a HTML markup filename.
func IsHTML(name string) bool {
	s := []string{htm, html}
	return IsExt(name, s...)
}

// IsImage returns true if the named file uses a common digital audio filename.
func IsAudio(name string) bool {
	s := []string{au, fla, mla, m2a, mid, midi, mp1, mp2, mp3, mpa, mpga, mpeg, ogg, snd, wav, wave, wma}
	return IsExt(name, s...)
}

// IsImage returns true if the named file uses a common tracker music filename.
func IsTune(name string) bool {
	s := []string{it, mod, s3m, xm}
	return IsExt(name, s...)
}

// IsImage returns true if the named file uses a common video filename.
func IsVideo(name string) bool {
	s := []string{avi, divx, flv, gt, mov, m4a, m4v, mp4, swf, rm, ram, wmv, xvid}
	return IsExt(name, s...)
}

// IsExt returns true if the file extension of the named file is found in the collection of extensions.
func IsExt(name string, extensions ...string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return Finds(ext, extensions...)
}
