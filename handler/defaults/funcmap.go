package defaults

import (
	"crypto/sha512"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/url"
	"strings"
)

const (
	wiki = `<svg class="bi" aria-hidden="true"><use xlink:href="bootstrap-icons.svg#arrow-right-short"></use></svg>`
	link = `<svg class="bi" aria-hidden="true"><use xlink:href="bootstrap-icons.svg#link"></use></svg>`
)

// Helper functions for the TemplateFuncMap var.

// Integrity
func Integrity(name string, fs embed.FS) (string, error) {
	b, err := fs.ReadFile(name)
	if err != nil {
		return "", err
	}
	sum := sha512.Sum384(b)
	b64 := base64.StdEncoding.EncodeToString(sum[:])
	return fmt.Sprintf("sha384-%s", b64), nil
}

func ExternalLink(href, name string) template.HTML {
	if href == "" {
		return "error: href is empty"
	}
	if name == "" {
		return "error: name is empty"
	}

	return template.HTML(fmt.Sprintf(`<a class="dropdown-item icon-link icon-link-hover" href="%s">%s %s</a>`, href, name, link))
}

func WikiLink(uri, name string) template.HTML {
	if uri == "" {
		return "error: href is empty"
	}
	if name == "" {
		return "error: name is empty"
	}
	href, err := url.JoinPath("https://github.com/Defacto2/defacto2.net/wiki/", uri)
	if err != nil {
		return template.HTML(err.Error())
	}
	return template.HTML(fmt.Sprintf(`<a class="dropdown-item icon-link icon-link-hover" href="%s">%s %s</a>`, href, name, wiki))
}

// TODO: if pad is an odd value, include a blinking cursor
func LogoText(s string) string {
	const (
		welcome = ":                  ·· WELCOME TO DEFACTO2█ ··                     ·"
		padder  = " ·· "
	)
	if s == "" {
		return welcome
	}
	const w, p = len(welcome), len(padder)
	const max = w - 3 - (p * 2)
	l := len(s)
	if l > max {
		return fmt.Sprintf(":%s%s%s·", padder, s[:max], padder)
	}

	// ":                  ·· WELCOME TO DEFACTO2 ··                      ·"
	ns := fmt.Sprintf("%s%s%s", padder, s, padder)
	pad := (w / 2) - (len(ns) / 2) - (p / 2)
	return fmt.Sprintf(":%s%s", strings.Repeat(" ", pad), ns)
}
