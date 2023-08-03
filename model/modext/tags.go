package modext

import (
	"github.com/Defacto2/server/pkg/tags"
	"github.com/volatiletech/null/v8"
)

func PAnsi() null.String {
	return null.String{String: tags.URIs()[tags.ANSI], Valid: true}
}

func SBbs() null.String {
	return null.String{String: tags.URIs()[tags.BBS], Valid: true}
}

func SDemo() null.String {
	return null.String{String: tags.URIs()[tags.Demo], Valid: true}
}

func PDos() null.String {
	return null.String{String: tags.URIs()[tags.DOS], Valid: true}
}

func SInstall() null.String {
	return null.String{String: tags.URIs()[tags.Install], Valid: true}
}

func SIntro() null.String {
	return null.String{String: tags.URIs()[tags.Intro], Valid: true}
}

func PLinux() null.String {
	return null.String{String: tags.URIs()[tags.Linux], Valid: true}
}

func PJava() null.String {
	return null.String{String: tags.URIs()[tags.Java], Valid: true}
}

func SMag() null.String {
	return null.String{String: tags.URIs()[tags.Mag], Valid: true}
}

func PMac() null.String {
	return null.String{String: tags.URIs()[tags.Mac], Valid: true}
}

func PNfo() null.String {
	return null.String{String: tags.URIs()[tags.Nfo], Valid: true}
}

func nfoTool() null.String {
	return null.String{String: tags.URIs()[tags.NfoTool], Valid: true}
}

func SProof() null.String {
	return null.String{String: tags.URIs()[tags.Proof], Valid: true}
}

func PScript() null.String {
	return null.String{String: tags.URIs()[tags.PHP], Valid: true}
}

func PText() null.String {
	return null.String{String: tags.URIs()[tags.Text], Valid: true}
}

func PWindows() null.String {
	return null.String{String: tags.URIs()[tags.Windows], Valid: true}
}
