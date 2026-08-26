package querymod

// Package querymod contains functions that return a null.String type for use in SQL queries.

import (
	"github.com/Defacto2/server/internal/tags"
	"github.com/aarondl/null/v8"
)

// uris is a cached reference to the tags.URIs() map.
//
//nolint:gochecknoglobals
var uris = tags.URIs()

// funcs that begin with S are for the section column.

func SAdvert() null.String {
	return null.String{String: uris[tags.ForSale], Valid: true}
}

func SAnnouncement() null.String {
	return null.String{String: uris[tags.Announcement], Valid: true}
}

func SAppleII() null.String {
	return null.String{String: uris[tags.AppleII], Valid: true}
}

func SAtariST() null.String {
	return null.String{String: uris[tags.AtariST], Valid: true}
}

func SBbs() null.String {
	return null.String{String: uris[tags.BBS], Valid: true}
}

func SBrand() null.String {
	return null.String{String: uris[tags.Logo], Valid: true}
}

func SDemo() null.String {
	return null.String{String: uris[tags.Demo], Valid: true}
}

func SDrama() null.String {
	return null.String{String: uris[tags.Drama], Valid: true}
}

func SFtp() null.String {
	return null.String{String: uris[tags.Ftp], Valid: true}
}

func SHack() null.String {
	return null.String{String: uris[tags.GameHack], Valid: true}
}

func SHowTo() null.String {
	return null.String{String: uris[tags.Guide], Valid: true}
}

func SInstall() null.String {
	return null.String{String: uris[tags.Install], Valid: true}
}

func SIntro() null.String {
	return null.String{String: uris[tags.Intro], Valid: true}
}

func SJobAdvert() null.String {
	return null.String{String: uris[tags.Job], Valid: true}
}

func SMag() null.String {
	return null.String{String: uris[tags.Mag], Valid: true}
}

func SNews() null.String {
	return null.String{String: uris[tags.News], Valid: true}
}

func SNfo() null.String {
	return null.String{String: uris[tags.Nfo], Valid: true}
}

func SNfoTool() null.String {
	return null.String{String: uris[tags.NfoTool], Valid: true}
}

func SPack() null.String {
	return null.String{String: uris[tags.Pack], Valid: true}
}

func SProof() null.String {
	return null.String{String: uris[tags.Proof], Valid: true}
}

func SRestrict() null.String {
	return null.String{String: uris[tags.Restrict], Valid: true}
}

func SStandard() null.String {
	return null.String{String: uris[tags.Rule], Valid: true}
}

func STakedown() null.String {
	return null.String{String: uris[tags.Bust], Valid: true}
}

func STool() null.String {
	return null.String{String: uris[tags.Tool], Valid: true}
}

// Methods that begin with P are for the platform column.

func PAnsi() null.String {
	return null.String{String: uris[tags.ANSI], Valid: true}
}

func PConsole() null.String {
	return null.String{String: uris[tags.Console], Valid: true}
}

func PDatabase() null.String {
	return null.String{String: uris[tags.DataB], Valid: true}
}

func PDos() null.String {
	return null.String{String: uris[tags.DOS], Valid: true}
}

func PHtml() null.String {
	return null.String{String: uris[tags.Markup], Valid: true}
}

func PImage() null.String {
	return null.String{String: uris[tags.Image], Valid: true}
}

func PLinux() null.String {
	return null.String{String: uris[tags.Linux], Valid: true}
}

func PJava() null.String {
	return null.String{String: uris[tags.Java], Valid: true}
}

func PMac() null.String {
	return null.String{String: uris[tags.Mac], Valid: true}
}

func PMusic() null.String {
	return null.String{String: uris[tags.Audio], Valid: true}
}

func PPCBoard() null.String {
	return null.String{String: uris[tags.PCB], Valid: true}
}

func PPdf() null.String {
	return null.String{String: uris[tags.PDF], Valid: true}
}

func PScript() null.String {
	return null.String{String: uris[tags.PHP], Valid: true}
}

func PText() null.String {
	return null.String{String: uris[tags.Text], Valid: true}
}

func PTextAmiga() null.String {
	return null.String{String: uris[tags.TextAmiga], Valid: true}
}

func PVideo() null.String {
	return null.String{String: uris[tags.Video], Valid: true}
}

func PWindows() null.String {
	return null.String{String: uris[tags.Windows], Valid: true}
}
