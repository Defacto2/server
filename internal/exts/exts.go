// Package exts contains common file extensions used by the file records.
package exts

import (
	"path/filepath"
	"strings"

	"github.com/Defacto2/server/internal/helper"
)

const (
	com = ".com"
	exe = ".exe"

	z7  = ".7z"
	arc = ".arc"
	ark = ".ark"
	arj = ".arj"
	cab = ".cab"
	gz  = ".gz"
	lha = ".lha"
	lzh = ".lzh"
	rar = ".rar"
	tar = ".tar"
	tgz = ".tar.gz"
	zip = ".zip"

	fst = ".1st"
	asc = ".asc"
	ans = ".ans"
	cpt = ".cap"
	diz = ".diz"
	doc = ".doc"
	dox = ".dox"
	me  = ".me"
	nfo = ".nfo"
	pcb = ".pcb"
	pdf = ".pdf"
	txt = ".txt"
	unp = ".unp"

	bmp  = ".bmp"
	gif  = ".gif"
	ico  = ".ico"
	iff  = ".iff"
	jpg  = ".jpg"
	jpeg = ".jpeg"
	lbm  = ".lbm"
	png  = ".png"
	pcx  = ".pcx"

	htm  = ".htm"
	htmx = ".html"

	au   = ".au"
	fla  = ".flac"
	mla  = ".m1a"
	m2a  = ".m2a"
	mid  = ".mid"
	midi = ".midi"
	mp1  = ".mp1"
	mp2  = ".mp2"
	mp3  = ".mp3"
	mpa  = ".mpa"
	mpga = ".mpga"
	mpeg = ".mpeg"
	ogg  = ".ogg"
	snd  = ".snd"
	wav  = ".wav"
	wave = ".wave"
	wma  = ".wma"

	it  = ".it"
	mod = ".mod"
	s3m = ".s3m"
	xm  = ".xm"

	avi  = ".avi"
	divx = ".divx"
	flv  = ".flv"
	gt   = ".gt"
	mov  = ".mov"
	m4a  = ".m4a"
	m4v  = ".m4v"
	mp4  = ".mp4"
	swf  = ".swf"
	rm   = ".rm"
	ram  = ".ram"
	wmv  = ".wmv"
	xvid = ".xvid"
)

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
	s := []string{htm, htmx}
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
	return helper.Finds(ext, extensions...)
}

// Icon names for the Apache2 .gif image files used as icons for the named file.
const (
	App = "comp2"
	Doc = "doc"
	Htm = "generic"
	Pic = "image2"
	Sfx = "sound2"
	Vid = "movie"
	Zip = "compressed"
)

// IconName returns the extensionless name of an Apache2 .gif image file to use as an icon
// for the named file.
func IconName(name string) string {
	n := strings.ToLower(filepath.Ext(name))
	switch {
	case IsArchive(n):
		return Zip
	case IsApp(n):
		return App
	case IsAudio(n):
		return Sfx
	case IsDocument(n):
		return Doc
	case IsHTML(n):
		return Htm
	case IsImage(n):
		return Pic
	case IsTune(n):
		return Sfx
	case IsVideo(n):
		return Vid
	}
	return ""
}
