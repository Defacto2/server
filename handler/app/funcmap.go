package app

// Package file funcmap.go contains the custom template functions for the web framework.
// The functions are used by the HTML templates to format data.

import (
	"crypto/sha512"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/url"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/Defacto2/server/pkg/config"
	"github.com/Defacto2/server/pkg/fmts"
	"github.com/Defacto2/server/pkg/helper"
	"github.com/Defacto2/server/pkg/initialism"
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

	link  = `<svg class="bi" aria-hidden="true"><use xlink:href="/bootstrap-icons.svg#link"></use></svg>`
	merge = `<svg class="bi" aria-hidden="true" fill="currentColor"><use xlink:href="/bootstrap-icons.svg#forward"></use></svg>`
	wiki  = `<svg class="bi" aria-hidden="true"><use xlink:href="/bootstrap-icons.svg#arrow-right-short"></use></svg>`

	typeErr = "error: received an invalid type to "
)

// TemplateFuncMap are a collection of mapped functions that can be used in a template.
func (c Configuration) TemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"attribute":      Attribute,
		"describe":       Describe,
		"fmtByte":        FmtByte,
		"fmtByteCnt":     FmtByteCnt,
		"fmtByteName":    FmtByteName,
		"fmtDay":         FmtDay,
		"fmtInitalism":   initialism.Join,
		"fmtMonth":       FmtMonth,
		"fmtPrefix":      FmtPrefix,
		"lastUpdated":    LastUpdated,
		"linkDownload":   LinkDownload,
		"linkPage":       LinkPage,
		"linkRemote":     LinkRemote,
		"linkRelrs":      LinkRelrs,
		"linkScnr":       LinkScnr,
		"linkWiki":       LinkWiki,
		"logoText":       LogoText,
		"mod3":           Mod3,
		"safeHTML":       SafeHTML,
		"sizeOfDL":       SizeOfDL,
		"subTitle":       SubTitle,
		"thumb":          c.Thumb,
		"trimSiteSuffix": TrimSiteSuffix,
		"databaseDown": func() bool {
			return c.DatbaseErr
		},
		"fmtURI": func(uri string) string {
			return fmts.Name(uri)
		},
		"logo": func() string {
			return string(*c.Brand)
		},
		"mergeIcon": func() string {
			return merge
		},
		"msdos": func() template.HTML {
			return "<span class=\"text-nowrap\">MS Dos</span>"
		},
		"sriBootCSS": func() string {
			return c.Subresource.BootstrapCSS
		},
		"sriBootJS": func() string {
			return c.Subresource.BootstrapJS
		},
		"sriFA": func() string {
			return c.Subresource.FontAwesome
		},
		"sriLayout": func() string {
			return c.Subresource.LayoutCSS
		},
	}
}

func Attribute(write, code, art, music, name string) string {
	name = strings.ToLower(name)
	w, c, a, m :=
		strings.Split(strings.ToLower(write), ","),
		strings.Split(strings.ToLower(code), ","),
		strings.Split(strings.ToLower(art), ","),
		strings.Split(strings.ToLower(music), ",")
	if len(w) == 0 && len(c) == 0 && len(a) == 0 && len(m) == 0 {
		return ""
	}
	if name == "" {
		return ""
	}
	match := []string{}
	if slices.Contains(w, name) {
		match = append(match, "writer")
	}
	if slices.Contains(c, name) {
		match = append(match, "programmer")
	}
	if slices.Contains(a, name) {
		match = append(match, "artist")
	}
	if slices.Contains(m, name) {
		match = append(match, "musician")
	}
	if len(match) == 0 {
		all := []string{write, code, art, music}
		return fmt.Sprintf("error: %q, %s", name, strings.Join(all, ","))
	}
	match[0] = helper.Capitalize(match[0])
	if len(match) == 1 {
		return match[0] + " attribution"
	}
	if len(match) == 2 {
		return strings.Join(match, " and ") + " attributions"
	}
	last := len(match) - 1
	match[last] = "and " + match[last]
	return strings.Join(match, ", ") + " attributions"
}

// Thumb returns a HTML image tag or picture element for the given uuid.
// The uuid is the filename of the thumbnail image without an extension.
// The desc is the description of the image.
func (c Configuration) Thumb(uuid, desc string) template.HTML {
	fw := filepath.Join(c.Import.ThumbnailDir, fmt.Sprintf("%s.webp", uuid))
	fp := filepath.Join(c.Import.ThumbnailDir, fmt.Sprintf("%s.png", uuid))
	webp := strings.Join([]string{config.StaticThumb(), fmt.Sprintf("%s.webp", uuid)}, "/")
	png := strings.Join([]string{config.StaticThumb(), fmt.Sprintf("%s.png", uuid)}, "/")
	alt := strings.ToLower(desc) + " thumbnail"
	w, p := false, false
	if helper.IsStat(fw) {
		w = true
	}
	if helper.IsStat(fp) {
		p = true
	}
	const style = "min-height:10em;max-height:20em;"
	if !w && !p {
		return template.HTML("<img src=\"\" alt=\"thumbnail placeholder\" class=\"card-img-top placeholder\" style=\"" + style + "\" />")
	}
	if w && p {
		elm := "<picture class=\"card-img-top\">" +
			fmt.Sprintf("<source srcset=\"%s\" type=\"image/webp\" />", webp) +
			string(img(png, alt, style)) +
			"</picture>"
		return template.HTML(elm)
	}
	elm := ""
	if w {
		return img(webp, alt, style)
	}
	if p {
		return img(png, alt, style)
	}
	return template.HTML(elm)
}

func img(src, alt, style string) template.HTML {
	return template.HTML(fmt.Sprintf("<img src=\"%s\" alt=\"%s\" class=\"card-img-top\" style=\"%s\" />", src, alt, style))
}

// Describe returns a human readable description of a release.
// Based on the platform, section, year and month.
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
		m = helper.ShortMonth(int(i))
	case null.Int16:
		if val.Valid {
			m = helper.ShortMonth(int(val.Int16))
		}
	default:
		return template.HTML(fmt.Sprintf("%s %s %s", typeErr, "describe", month))
	}

	if p == "" && s == "" {
		return template.HTML("An unknown release")
	}
	x := tags.Humanize(tags.TagByURI(p), tags.TagByURI(s))
	x = helper.Capitalize(x)
	//x := HumanizeDescription(p, s)
	if m != "" && y != "" {
		x = fmt.Sprintf("%s published in <span class=\"text-nowrap\">%s, %s</a>", x, m, y)
	} else if y != "" {
		x = fmt.Sprintf("%s published in %s", x, y)
	}
	return template.HTML(x + ".")
}

// FmtByte returns a human readable string of the byte count.
func FmtByte(b any) string {
	switch val := b.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		return helper.ByteCount(i)
	default:
		return fmt.Sprintf("%sByteFmt: %s", typeErr, reflect.TypeOf(b).String())
	}
}

// FmtByteCnt returns a human readable string of the file count and bytes.
func FmtByteCnt(c, b any) template.HTML {
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
		s = fmt.Sprintf("%s <small>(%s)</small>", s, helper.ByteCount(i))
	default:
		s = fmt.Sprintf("%sByteFmt: %s", typeErr, reflect.TypeOf(b).String())
		return template.HTML(s)
	}
	return template.HTML(s)
}

// FmtByteName returns a human readable string of the byte count with a named description.
func FmtByteName(name string, c, b any) template.HTML {
	s := ""
	switch val := c.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		if i != 1 {
			name = fmt.Sprintf("%ss", name)
		}
		p := message.NewPrinter(language.English)
		s = p.Sprintf("%d", i)
	default:
		s = fmt.Sprintf("%sByteFmt: %s", typeErr, reflect.TypeOf(c).String())
		return template.HTML(s)
	}
	switch val := b.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		s = fmt.Sprintf("%s %s <small>(%s)</small>", s, name, helper.ByteCount(i))
	default:
		s = fmt.Sprintf("%sByteFmt: %s", typeErr, reflect.TypeOf(b).String())
		return template.HTML(s)
	}
	return template.HTML(s)
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

// FmtYears returns a string of the years if they are different.
// If they are the same, it returns a singular year.
func FmtYears(a, b int) string {
	if a == b {
		return fmt.Sprintf("the year %d", a)
	}
	if b-a == 1 {
		return fmt.Sprintf("the years %d and %d", a, b)
	}
	return fmt.Sprintf("the years %d - %d", a, b)
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

// LinkDownload creates a URL to link to the file download of the record.
func LinkDownload(id any) template.HTML {
	s, err := linkID(id, "d")
	if err != nil {
		return template.HTML(err.Error())
	}
	return template.HTML(fmt.Sprintf(`<a class="card-link" href="%s">Download</a>`, s))
}

// LinkRelrs returns the groups associated with a release and a link to each group.
func LinkRelrs(a, b any) template.HTML {
	const class = "text-nowrap"
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
		ref, err := LinkRelr(av)
		if err != nil {
			return template.HTML(fmt.Sprintf("error: %s", err))
		}
		prime = fmt.Sprintf(`<a class="%s" href="%s">%s</a>`, class, ref, fmts.Name(helper.Slug(av)))
	}
	if bv != "" {
		ref, err := LinkRelr(bv)
		if err != nil {
			return template.HTML(fmt.Sprintf("error: %s", err))
		}
		second = fmt.Sprintf(`<a class="%s" href="%s">%s</a>`, class, ref, fmts.Name(helper.Slug(bv)))
	}
	if prime != "" && second != "" {
		s = fmt.Sprintf("%s<br>+ %s", prime, second)
	} else if prime != "" {
		s = prime
	} else if second != "" {
		s = second
	}
	return template.HTML(s)
}

// LinkRelr returns a link to the named group page.
func LinkRelr(name string) (string, error) {
	href, err := url.JoinPath("/", "g", helper.Slug(name))
	if err != nil {
		return "", fmt.Errorf("name %q could not be made into a valid url: %s", name, err)
	}
	return href, nil
}

// LinkScnr returns a link to the named scener page.
func LinkScnr(name string) (string, error) {
	href, err := url.JoinPath("/", "p", helper.Slug(name))
	if err != nil {
		return "", fmt.Errorf("name %q could not be made into a valid url: %s", name, err)
	}
	return href, nil
}

// LinkRemote returns a HTML link with an embedded SVG icon to an external website.
func LinkRemote(href, name string) template.HTML {
	if href == "" {
		return "error: href is empty"
	}
	if name == "" {
		return "error: name is empty"
	}
	a := fmt.Sprintf(`<a class="dropdown-item icon-link icon-link-hover" href="%s">%s %s</a>`,
		href, name, link)
	return template.HTML(a)
}

// LinkPage creates a URL to link to the file page for the record.
func LinkPage(id any) template.HTML {
	s, err := linkID(id, "f")
	if err != nil {
		return template.HTML(err.Error())
	}
	return template.HTML(fmt.Sprintf(`<a class="card-link" href="%s">About</a>`, s))
}

// LinkWiki returns a HTML link with an embedded SVG icon to the Defacto2 wiki on GitHub.
func LinkWiki(uri, name string) template.HTML {
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
	a := fmt.Sprintf(`<a class="dropdown-item icon-link icon-link-hover" href="%s">%s %s</a>`,
		href, name, wiki)
	return template.HTML(a)
}

// linkID creates a URL to link to the record.
// The id is obfuscated to prevent direct linking.
// The elem is the element to link to, such as 'f' for file or 'd' for download.
func linkID(id any, elem string) (string, error) {
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
	href, err := url.JoinPath("/", elem, helper.Obfuscate(i))
	if err != nil {
		return "", fmt.Errorf("id %d could not be made into a valid url: %s", i, err)
	}
	return href, nil
}

// LogoText returns a string of text padded with spaces to center it in the logo.
func LogoText(s string) string {
	const spaces = 6
	indent := strings.Repeat(" ", spaces)
	if s == "" {
		return indent + Welcome
	}

	const padder = " ·· "
	const wl, pl = len(Welcome), len(padder)
	const limit = wl - (pl + pl) - 3
	// odd returns true if the given integer is odd.
	odd := func(i int) bool {
		return i%2 != 0
	}
	s = strings.ToUpper(s)

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
	const split = 2
	count := (wl / split) - (len(styled) / split) - split
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
	const max = 3
	return Mod(i, max)
}

// SafeHTML returns a string as a template.HTML type.
// This is intended to be used to prevent HTML escaping.
func SafeHTML(s string) template.HTML {
	return template.HTML(s)
}

// SizeOfDL returns a human readable string of the file size.
func SizeOfDL(i any) template.HTML {
	s := ""
	switch val := i.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		s = helper.ByteCount(i)
	case null.Int64:
		if !val.Valid {
			return ""
		}
		s = helper.ByteCount(val.Int64)
	default:
		return template.HTML(fmt.Sprintf("%sSizeOfDL: %s", typeErr, reflect.TypeOf(i).String()))
	}
	return template.HTML(fmt.Sprintf(" <small class=\"text-body-secondary\">(%s)</small>", s))
}

// SubTitle returns a secondary element with the record title.
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

// TrimSiteSuffix returns a string with the last 4 characters removed if they are " FTP" or " BBS".
func TrimSiteSuffix(s string) string {
	n := strings.ToLower(strings.TrimSpace(s))
	const chrs = 4
	if len(s) < chrs {
		return s
	}
	switch n[len(s)-chrs:] {
	case " ftp", " bbs":
		return s[:len(s)-chrs]
	}
	return s
}

// TODO: remove?
// HumanizeDescription returns a human readable description of a release.
// Based on the platform and section.
func HumanizeDescription(p, s string) string {
	x := ""

	if p == "" {
		x = fmt.Sprintf("A %s", s)
	}
	if s == "" {
		if isOS(p) {
			x = fmt.Sprintf("A release for %s", p)
		} else {
			x = fmt.Sprintf("A %s file", p)
		}
	}
	if x == "" && p == tags.Text.String() && s == tags.Nfo.String() {
		x = "A scene release text file"
	}
	if x == "" && isOS(p) {
		x = fmt.Sprintf("A %s for %s", tags.NameByURI(s), tags.NameByURI(p))
	}
	if x == "" {
		x = fmt.Sprintf("A %s %s", tags.NameByURI(s), tags.NameByURI(p))
	}
	return x
}

// isOS returns true if the platform matches Windows, macOS, Linux, MS-DOS or Java.
func isOS(platform string) bool {
	s := tags.OSTags()
	apps := s[:]
	plat := strings.TrimSpace(strings.ToLower(platform))
	return helper.Finds(plat, apps...)
}
