package querymod

// Package querymod contains functions that return a null.String type for use in SQL queries.

import (
	"github.com/Defacto2/server/internal/tags"
	"github.com/aarondl/null/v8"
)

// uris is a cached reference to the tags.URIs() map.
var uris map[tags.Tag]string

func getURIs() map[tags.Tag]string {
	if uris == nil {
		uris = tags.URIs()
	}
	return uris
}

// funcs that begin with S are for the section column.

func SAdvert() null.String {
	return null.String{String: getURIs()[tags.ForSale], Valid: true}
}

func SAnnouncement() null.String {
	return null.String{String: getURIs()[tags.Announcement], Valid: true}
}

func SAppleII() null.String {
	return null.String{String: getURIs()[tags.AppleII], Valid: true}
}

func SAtariST() null.String {
	return null.String{String: getURIs()[tags.AtariST], Valid: true}
}

func SBbs() null.String {
	return null.String{String: getURIs()[tags.BBS], Valid: true}
}

func SBrand() null.String {
	return null.String{String: getURIs()[tags.Logo], Valid: true}
}

func SDemo() null.String {
	return null.String{String: getURIs()[tags.Demo], Valid: true}
}

func SDrama() null.String {
	return null.String{String: getURIs()[tags.Drama], Valid: true}
}

func SFtp() null.String {
	return null.String{String: getURIs()[tags.Ftp], Valid: true}
}

func SHack() null.String {
	return null.String{String: getURIs()[tags.GameHack], Valid: true}
}

func SHowTo() null.String {
	return null.String{String: getURIs()[tags.Guide], Valid: true}
}

func SInstall() null.String {
	return null.String{String: getURIs()[tags.Install], Valid: true}
}

func SIntro() null.String {
	return null.String{String: getURIs()[tags.Intro], Valid: true}
}

func SJobAdvert() null.String {
	return null.String{String: getURIs()[tags.Job], Valid: true}
}

func SMag() null.String {
	return null.String{String: getURIs()[tags.Mag], Valid: true}
}

func SNews() null.String {
	return null.String{String: getURIs()[tags.News], Valid: true}
}

func SNfo() null.String {
	return null.String{String: getURIs()[tags.Nfo], Valid: true}
}

func SNfoTool() null.String {
	return null.String{String: getURIs()[tags.NfoTool], Valid: true}
}

func SPack() null.String {
	return null.String{String: getURIs()[tags.Pack], Valid: true}
}

func SProof() null.String {
	return null.String{String: getURIs()[tags.Proof], Valid: true}
}

func SRestrict() null.String {
	return null.String{String: getURIs()[tags.Restrict], Valid: true}
}

func SStandard() null.String {
	return null.String{String: getURIs()[tags.Rule], Valid: true}
}

func STakedown() null.String {
	return null.String{String: getURIs()[tags.Bust], Valid: true}
}

func STool() null.String {
	return null.String{String: getURIs()[tags.Tool], Valid: true}
}

// funcs that begin with P are for the platform column.

func PAnsi() null.String {
	return null.String{String: getURIs()[tags.ANSI], Valid: true}
}

func PDatabase() null.String {
	return null.String{String: getURIs()[tags.DataB], Valid: true}
}

func PDos() null.String {
	return null.String{String: getURIs()[tags.DOS], Valid: true}
}

func PHtml() null.String {
	return null.String{String: getURIs()[tags.Markup], Valid: true}
}

func PImage() null.String {
	return null.String{String: getURIs()[tags.Image], Valid: true}
}

func PLinux() null.String {
	return null.String{String: getURIs()[tags.Linux], Valid: true}
}

func PJava() null.String {
	return null.String{String: getURIs()[tags.Java], Valid: true}
}

func PMac() null.String {
	return null.String{String: getURIs()[tags.Mac], Valid: true}
}

func PMusic() null.String {
	return null.String{String: getURIs()[tags.Audio], Valid: true}
}

func PPdf() null.String {
	return null.String{String: getURIs()[tags.PDF], Valid: true}
}

func PScript() null.String {
	return null.String{String: getURIs()[tags.PHP], Valid: true}
}

func PText() null.String {
	return null.String{String: getURIs()[tags.Text], Valid: true}
}

func PTextAmiga() null.String {
	return null.String{String: getURIs()[tags.TextAmiga], Valid: true}
}

func PVideo() null.String {
	return null.String{String: getURIs()[tags.Video], Valid: true}
}

func PWindows() null.String {
	return null.String{String: getURIs()[tags.Windows], Valid: true}
}
