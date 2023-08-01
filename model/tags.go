package model

import (
	"github.com/Defacto2/server/pkg/tags"
	"github.com/volatiletech/null/v8"
)

func ansi() null.String {
	return null.String{String: tags.URIs()[tags.ANSI], Valid: true}
}

func bbs() null.String {
	return null.String{String: tags.URIs()[tags.BBS], Valid: true}
}

func demo() null.String {
	return null.String{String: tags.URIs()[tags.Demo], Valid: true}
}

func dos() null.String {
	return null.String{String: tags.URIs()[tags.DOS], Valid: true}
}

func install() null.String {
	return null.String{String: tags.URIs()[tags.Install], Valid: true}
}

func intro() null.String {
	return null.String{String: tags.URIs()[tags.Intro], Valid: true}
}

func linux() null.String {
	return null.String{String: tags.URIs()[tags.Linux], Valid: true}
}

func java() null.String {
	return null.String{String: tags.URIs()[tags.Java], Valid: true}
}

func mag() null.String {
	return null.String{String: tags.URIs()[tags.Mag], Valid: true}
}

func mac() null.String {
	return null.String{String: tags.URIs()[tags.Mac], Valid: true}
}

func nfo() null.String {
	return null.String{String: tags.URIs()[tags.Nfo], Valid: true}
}

func nfoTool() null.String {
	return null.String{String: tags.URIs()[tags.NfoTool], Valid: true}
}

func proof() null.String {
	return null.String{String: tags.URIs()[tags.Proof], Valid: true}
}

func script() null.String {
	return null.String{String: tags.URIs()[tags.PHP], Valid: true}
}

func text() null.String {
	return null.String{String: tags.URIs()[tags.Text], Valid: true}
}

func windows() null.String {
	return null.String{String: tags.URIs()[tags.Windows], Valid: true}
}
