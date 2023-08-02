package app

// Helper functions for the TemplateFuncMap var.

import (
	"crypto/sha512"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/Defacto2/server/pkg/helpers"
	"github.com/Defacto2/server/pkg/tags"
	"github.com/bengarrett/cfw"
	"github.com/volatiletech/null/v8"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	// Welcome is the default logo monospace text,
	// each side contains 20 whitespace characters.
	// The welcome to defacto2 text is 19 characters long.
	// The letter 'O' of TO is the center of the text.
	Welcome = ":                    ·· WELCOME TO DEFACTO2 ··                    ·"

	// wiki and link are SVG icons.
	wiki  = `<svg class="bi" aria-hidden="true"><use xlink:href="/bootstrap-icons.svg#arrow-right-short"></use></svg>`
	link  = `<svg class="bi" aria-hidden="true"><use xlink:href="/bootstrap-icons.svg#link"></use></svg>`
	merge = `<svg class="bi" aria-hidden="true" fill="currentColor"><use xlink:href="/bootstrap-icons.svg#forward"></use></svg>`

	typeErr = "error: received an invalid type to "
)

// TemplateFuncMap are a collection of mapped functions that can be used in a template.
func (c Configuration) TemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"byteFmt":      ByteFormat,
		"cntByteFmt":   CountByteFormat,
		"describe":     Describe,
		"externalLink": ExternalLink,
		"fmtDay":       FmtDay,
		"fmtMonth":     FmtMonth,
		"fmtPrefix":    FmtPrefix,
		"lastUpdated":  LastUpdated,
		"linkDL":       IDDownload,
		"linkGroups":   GroupsLink,
		"linkPage":     IDPage,
		"logoText":     LogoText,
		"mod3":         Mod3,
		"safeHTML":     SafeHTML,
		"sizeOfDL":     SizeOfDL,
		"subTitle":     SubTitle,
		"wikiLink":     WikiLink,
		"databaseDown": func() bool {
			return c.DatbaseErr
		},
		"logo": func() string {
			return string(*c.Brand)
		},
		"mergeIcon": func() string {
			return merge
		},
		"sriBootstrapCSS": func() string {
			return c.Subresource.BootstrapCSS
		},
		"sriBootstrapJS": func() string {
			return c.Subresource.BootstrapJS
		},
		"sriFontAwesome": func() string {
			return c.Subresource.FontAwesome
		},
		"sriLayoutCSS": func() string {
			return c.Subresource.LayoutCSS
		},
	}
}

func Describe(plat, sect, year, month any) template.HTML {
	p, s, y, m := "", "", "", ""
	switch val := plat.(type) {
	case string:
		p = val
	case null.String:
		if val.Valid {
			p = val.String
		}
	default:
		return template.HTML(fmt.Sprintf("%s %s %s", typeErr, "describe", plat))
	}
	p = strings.TrimSpace(p)
	switch val := sect.(type) {
	case string:
		s = val
	case null.String:
		if val.Valid {
			s = val.String
		}
	default:
		return template.HTML(fmt.Sprintf("%s %s %s", typeErr, "describe", sect))
	}
	s = strings.TrimSpace(s)
	switch val := year.(type) {
	case int, int8, int16, int32, int64:
		y = fmt.Sprintf("%v", val)
	case null.Int16:
		if val.Valid {
			y = fmt.Sprintf("%v", val.Int16)
		}
	default:
		return template.HTML(fmt.Sprintf("%s %s %s", typeErr, "describe", year))
	}
	switch val := month.(type) {
	case int, int8, int16, int32, int64:
		i := reflect.ValueOf(val).Int()
		m = helpers.ShortMonth(int(i))
	case null.Int16:
		if val.Valid {
			m = helpers.ShortMonth(int(val.Int16))
		}
	default:
		return template.HTML(fmt.Sprintf("%s %s %s", typeErr, "describe", month))
	}

	if p == "" && s == "" {
		return template.HTML("An unknown release")
	}
	x := HumanizeDescription(p, s)
	if m != "" && y != "" {
		x = fmt.Sprintf("%s published in <span class=\"text-nowrap\">%s, %s</a>", x, m, y)
	} else if y != "" {
		x = fmt.Sprintf("%s published in %s", x, y)
	}
	return template.HTML(x + ".")
}

func HumanizeDescription(p, s string) string {
	x := ""
	if p == "" {
		x = fmt.Sprintf("A %s", s)
	}
	if s == "" {
		if IsOS(p) {
			x = fmt.Sprintf("A release for %s", p)
		} else {
			x = fmt.Sprintf("A %s file", p)
		}
	}
	// if x == "" && IsSwap(p) {
	// 	x = fmt.Sprintf("A %s %s", tags.NameByURI(s), tags.NameByURI(p))
	// }
	if x == "" && p == tags.Text.String() && s == tags.Nfo.String() {
		x = "A scene release text file"
	}
	if x == "" && IsOS(p) {
		x = fmt.Sprintf("A %s for %s", tags.NameByURI(s), tags.NameByURI(p))
	}
	if x == "" {
		x = fmt.Sprintf("A %s %s", tags.NameByURI(s), tags.NameByURI(p))
	}
	return x
}

func IsSwap(platform string) bool {
	s := []string{tags.Text.String(), tags.TextAmiga.String()}
	apps := s[:]
	plat := strings.TrimSpace(strings.ToLower(platform))
	return helpers.Finds(plat, apps...)
}

// IsOS returns true if the platform matches Windows, macOS, Linux, MS-DOS or Java.
func IsOS(platform string) bool {
	s := tags.OSTags()
	apps := s[:]
	plat := strings.TrimSpace(strings.ToLower(platform))
	return helpers.Finds(plat, apps...)
}

func SubTitle(s any) template.HTML {
	val := ""
	switch v := s.(type) {
	case string:
		val = v
	case null.String:
		if !v.Valid {
			return ""
		}
		val = v.String
	}
	if val == "" {
		return ""
	}
	elem := fmt.Sprintf("<h6 class=\"card-subtitle mb-2 text-body-secondary\">%s</h6>", val)
	return template.HTML(elem)
}

func GroupsLink(a, b any) template.HTML {
	av, bv, s := "", "", ""
	switch val := a.(type) {
	case string:
		av = reflect.ValueOf(val).String()
	case null.String:
		if val.Valid {
			av = val.String
		}
	}
	switch val := b.(type) {
	case string:
		bv = reflect.ValueOf(val).String()
	case null.String:
		if val.Valid {
			bv = val.String
		}
	}
	prime, second, s := "", "", ""
	if av == "" && bv == "" {
		return template.HTML("error: unknown group")
	}
	if av != "" {
		ref, err := GroupLink(av)
		if err != nil {
			return template.HTML(fmt.Sprintf("error: %s", err))
		}
		prime = fmt.Sprintf(`<a href="%s">%s</a>`, ref, av)
	}
	if bv != "" {
		ref, err := GroupLink(bv)
		if err != nil {
			return template.HTML(fmt.Sprintf("error: %s", err))
		}
		second = fmt.Sprintf(`<a href="%s">%s</a>`, ref, bv)
	}
	if prime != "" && second != "" {
		s = fmt.Sprintf("%s + %s", prime, second)
	} else if prime != "" {
		s = prime
	} else if second != "" {
		s = second
	}
	return template.HTML(s)
}

func GroupLink(name string) (string, error) {
	href, err := url.JoinPath("/", "g", helpers.Slug(name))
	if err != nil {
		return "", fmt.Errorf("name %q could not be made into a valid url: %s", name, err)
	}
	return href, nil
}

// SizeOfDL returns a human readable string of the file size.
func SizeOfDL(i any) template.HTML {
	s := ""
	switch val := i.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		s = helpers.ByteCount(i)
	case null.Int64:
		if !val.Valid {
			return ""
		}
		s = helpers.ByteCount(val.Int64)
	default:
		return template.HTML(fmt.Sprintf("%sSizeOfDL: %s", typeErr, reflect.TypeOf(i).String()))
	}
	return template.HTML(fmt.Sprintf(" <small class=\"text-body-secondary\">(%s)</small>", s))
}

// ByteFormat returns a human readable string of the byte count.
func ByteFormat(b any) string {
	switch val := b.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		return helpers.ByteCount(i)
	default:
		return fmt.Sprintf("%sByteFmt: %s", typeErr, reflect.TypeOf(b).String())
	}
}

// CountByteFormat returns a human readable string of the file count and bytes.
func CountByteFormat(c, b any) template.HTML {
	s := ""
	switch val := c.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		p := message.NewPrinter(language.English)
		s = p.Sprintf("%d", i)
	default:
		s = fmt.Sprintf("%sByteFmt: %s", typeErr, reflect.TypeOf(c).String())
		return template.HTML(s)
	}
	switch val := b.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		s = fmt.Sprintf("%s <small>(%s)</small>", s, helpers.ByteCount(i))
	default:
		s = fmt.Sprintf("%sByteFmt: %s", typeErr, reflect.TypeOf(b).String())
		return template.HTML(s)
	}
	return template.HTML(s)
}

// ExternalLink returns a HTML link with an embedded SVG icon to an external website.
func ExternalLink(href, name string) template.HTML {
	if href == "" {
		return "error: href is empty"
	}
	if name == "" {
		return "error: name is empty"
	}

	return template.HTML(fmt.Sprintf(`<a class="dropdown-item icon-link icon-link-hover" href="%s">%s %s</a>`, href, name, link))
}

// IDDownload creates a URL to link to the file download of the record.
func IDDownload(id any) template.HTML {
	s, err := IDHref(id, "d")
	if err != nil {
		return template.HTML(err.Error())
	}
	return template.HTML(fmt.Sprintf(`<a class="card-link" href="%s">Download</a>`, s))
}

// IDPage creates a URL to link to the file page for the record.
func IDPage(id any) template.HTML {
	s, err := IDHref(id, "f")
	if err != nil {
		return template.HTML(err.Error())
	}
	return template.HTML(fmt.Sprintf(`<a class="card-link" href="%s">About</a>`, s))
}

// IDHref creates a URL to link to the record.
// The id is obfuscated to prevent direct linking.
// The elem is the element to link to, such as 'f' for file or 'd' for download.
func IDHref(id any, elem string) (string, error) {
	i := int64(0)
	switch val := id.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i = reflect.ValueOf(val).Int()
		if i <= 0 {
			return "", fmt.Errorf("negative id %d", i)
		}
	default:
		return "", fmt.Errorf("%s %s", typeErr, reflect.TypeOf(id).String())
	}
	href, err := url.JoinPath("/", elem, helpers.Obfuscate(i))
	if err != nil {
		return "", fmt.Errorf("id %d could not be made into a valid url: %s", i, err)
	}
	return href, nil
}

// FmtDate returns a string of the date in the format YYYY-MM-DD.
func FmtDate(d any) string {
	switch val := d.(type) {
	case time.Time:
		return val.Format("2006-01-02")
	default:
		return fmt.Sprintf("%sFmtDate: %s", typeErr, reflect.TypeOf(d).String())
	}
}

// FmtDateTime returns a string of the date and time in the format YYYY-MM-DD HH:MM:SS.
func FmtDateTime(d any) string {
	switch val := d.(type) {
	case time.Time:
		return val.Format("2006-01-02 15:04:05")
	default:
		return fmt.Sprintf("%sFmtDateTime: %s", typeErr, reflect.TypeOf(d).String())
	}
}

// FmtDay returns a string of the day number from the day d number between 1 and 31.
func FmtDay(d any) string {
	switch val := d.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		if i == 0 {
			return ""
		}
		if i < 0 || i > 31 {
			return fmt.Sprintf(" error: day out of range %d", i)
		}
		return fmt.Sprintf(" %d", i)
	default:
		return fmt.Sprintf("%sFmtDay: %s", typeErr, reflect.TypeOf(d).String())
	}
}

// FmtMonth returns a string of the month name from the month m number between 1 and 12.
func FmtMonth(m any) string {
	switch val := m.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		if i == 0 {
			return ""
		}
		if i < 0 || i > 12 {
			return fmt.Sprintf(" error: month out of range %d", i)
		}
		return " " + time.Month(i).String()
	default:
		return fmt.Sprintf("%sFmtMonth: %s", typeErr, reflect.TypeOf(m).String())
	}
}

// FmtPrefix returns a string prefixed with a space.
func FmtPrefix(s string) string {
	if s == "" {
		return ""
	}
	return fmt.Sprintf("%s ", s)
}

// Integrity returns the sha384 hash of the named embed file.
// This is intended to be used for Subresource Integrity (SRI)
// verification with integrity attributes in HTML script and link tags.
func Integrity(name string, fs embed.FS) (string, error) {
	b, err := fs.ReadFile(name)
	if err != nil {
		return "", err
	}
	return IntegrityBytes(b), nil
}

// IntegrityBytes returns the sha384 hash of the given byte slice.
func IntegrityBytes(b []byte) string {
	sum := sha512.Sum384(b)
	b64 := base64.StdEncoding.EncodeToString(sum[:])
	return fmt.Sprintf("sha384-%s", b64)
}

// LastUpdated returns a string of the time since the given time t.
// The time is formatted as "Last updated 1 hour ago".
// If the time is not valid, an empty string is returned.
func LastUpdated(t any) string {
	switch val := t.(type) {
	case null.Time:
		if !val.Valid {
			return ""
		}
		return fmt.Sprintf("Last updated %s ago", cfw.TimeDistance(val.Time, time.Now(), true))
	case time.Time:
		return fmt.Sprintf("Last updated %s ago", cfw.TimeDistance(val, time.Now(), true))
	default:
		return fmt.Sprintf("%sLastUpdated: %s", typeErr, reflect.TypeOf(t).String())
	}
}

// LogoText returns a string of text padded with spaces to center it in the logo.
func LogoText(s string) string {
	indent := strings.Repeat(" ", 6)
	if s == "" {
		return indent + Welcome
	}

	// odd returns true if the given integer is odd.
	odd := func(i int) bool {
		return i%2 != 0
	}

	s = strings.ToUpper(s)

	const padder = " ·· "
	const wl, pl = len(Welcome), len(padder)
	const limit = wl - (pl + pl) - 3

	// Truncate the string to the limit.
	if len(s) > limit {
		return fmt.Sprintf("%s:%s%s%s·",
			indent, padder, s[:limit], padder)
	}

	styled := fmt.Sprintf("%s%s%s", padder, s, padder)
	if !odd(len(s)) {
		styled = fmt.Sprintf(" %s%s%s", padder, s, padder)
	}

	// Pad the string with spaces to center it.
	count := (wl / 2) - (len(styled) / 2) - 2

	text := fmt.Sprintf(":%s%s%s·",
		strings.Repeat(" ", count),
		styled,
		strings.Repeat(" ", count))
	return indent + text
}

// Mod returns true if the given integer is a multiple of the given max integer.
func Mod(i any, max int) bool {
	switch val := i.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		v := reflect.ValueOf(val).Int()
		return v%int64(max) == 0
	default:
		return false
	}
}

// Mod3 returns true if the given integer is a multiple of 3.
func Mod3(i any) bool {
	return Mod(i, 3)
}

// SafeHTML returns a string as a template.HTML type.
// This is intended to be used to prevent HTML escaping.
func SafeHTML(s string) template.HTML {
	return template.HTML(s)
}

// WikiLink returns a HTML link with an embedded SVG icon to the Defacto2 wiki on GitHub.
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
