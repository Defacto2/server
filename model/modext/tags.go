package modext

import (
	"github.com/Defacto2/server/pkg/tags"
	"github.com/volatiletech/null/v8"
)

// funcs that begin with S are for the section column.

func SAdvert() null.String {
	return null.String{String: tags.URIs()[tags.ForSale], Valid: true}
}

func SAnnouncement() null.String {
	return null.String{String: tags.URIs()[tags.Announcement], Valid: true}
}

func SAppleII() null.String {
	return null.String{String: tags.URIs()[tags.AppleII], Valid: true}
}

func SAtariST() null.String {
	return null.String{String: tags.URIs()[tags.AtariST], Valid: true}
}

func SBbs() null.String {
	return null.String{String: tags.URIs()[tags.BBS], Valid: true}
}

func SBrand() null.String {
	return null.String{String: tags.URIs()[tags.Logo], Valid: true}
}

func SDemo() null.String {
	return null.String{String: tags.URIs()[tags.Demo], Valid: true}
}

func SDrama() null.String {
	return null.String{String: tags.URIs()[tags.Drama], Valid: true}
}

func SFtp() null.String {
	return null.String{String: tags.URIs()[tags.Ftp], Valid: true}
}

func SHack() null.String {
	return null.String{String: tags.URIs()[tags.GameHack], Valid: true}
}

func SHowTo() null.String {
	return null.String{String: tags.URIs()[tags.Guide], Valid: true}
}

func SInstall() null.String {
	return null.String{String: tags.URIs()[tags.Install], Valid: true}
}

func SIntro() null.String {
	return null.String{String: tags.URIs()[tags.Intro], Valid: true}
}

func SJobAdvert() null.String {
	return null.String{String: tags.URIs()[tags.Job], Valid: true}
}

func SMag() null.String {
	return null.String{String: tags.URIs()[tags.Mag], Valid: true}
}

func SNews() null.String {
	return null.String{String: tags.URIs()[tags.News], Valid: true}
}

func SNfo() null.String {
	return null.String{String: tags.URIs()[tags.Nfo], Valid: true}
}

func SNfoTool() null.String {
	return null.String{String: tags.URIs()[tags.NfoTool], Valid: true}
}

func SPack() null.String {
	return null.String{String: tags.URIs()[tags.Pack], Valid: true}
}

func SProof() null.String {
	return null.String{String: tags.URIs()[tags.Proof], Valid: true}
}

func SRestrict() null.String {
	return null.String{String: tags.URIs()[tags.Restrict], Valid: true}
}

func SStandard() null.String {
	return null.String{String: tags.URIs()[tags.Rule], Valid: true}
}

func STakedown() null.String {
	return null.String{String: tags.URIs()[tags.Bust], Valid: true}
}

func STool() null.String {
	return null.String{String: tags.URIs()[tags.Tool], Valid: true}
}

// funcs that begin with P are for the platform column.

func PAnsi() null.String {
	return null.String{String: tags.URIs()[tags.ANSI], Valid: true}
}

func PDatabase() null.String {
	return null.String{String: tags.URIs()[tags.DataB], Valid: true}
}

func PDos() null.String {
	return null.String{String: tags.URIs()[tags.DOS], Valid: true}
}

func PHtml() null.String {
	return null.String{String: tags.URIs()[tags.Markup], Valid: true}
}

func PImage() null.String {
	return null.String{String: tags.URIs()[tags.Image], Valid: true}
}

func PLinux() null.String {
	return null.String{String: tags.URIs()[tags.Linux], Valid: true}
}

func PJava() null.String {
	return null.String{String: tags.URIs()[tags.Java], Valid: true}
}

func PMac() null.String {
	return null.String{String: tags.URIs()[tags.Mac], Valid: true}
}

func PMusic() null.String {
	return null.String{String: tags.URIs()[tags.Audio], Valid: true}
}

func PPdf() null.String {
	return null.String{String: tags.URIs()[tags.PDF], Valid: true}
}

func PScript() null.String {
	return null.String{String: tags.URIs()[tags.PHP], Valid: true}
}

func PText() null.String {
	return null.String{String: tags.URIs()[tags.Text], Valid: true}
}

func PTextAmiga() null.String {
	return null.String{String: tags.URIs()[tags.TextAmiga], Valid: true}
}

func PVideo() null.String {
	return null.String{String: tags.URIs()[tags.Video], Valid: true}
}

func PWindows() null.String {
	return null.String{String: tags.URIs()[tags.Windows], Valid: true}
}
