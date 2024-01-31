// Package app handles the routes and views for the Defacto2 website.
package app

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/tags"
	"github.com/volatiletech/null/v8"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	// Welcome is the default logo monospace text,
	// each side contains 20 whitespace characters.
	// The welcome to defacto2 text is 19 characters long.
	// The letter 'O' of TO is the center of the text.
	Welcome = `:                    ` +
		`·· WELCOME TO DEFACTO2 ··` +
		`                    ·`
	// SessionName is the name given to the session cookie.
	SessionName = "d2_op"
)

var (
	ErrCategory = errors.New("unknown file category")
	ErrClaims   = errors.New("no sub id in the claims playload")
	ErrCode     = errors.New("the http status code is not valid")
	ErrCxt      = errors.New("the server could not create a context")
	ErrData     = errors.New("cache data is invalid or corrupt")
	ErrDB       = errors.New("database connection is nil")
	ErrExtract  = errors.New("unknown extractor value")
	ErrLinkID   = errors.New("the id value cannot be a negative number")
	ErrLinkType = errors.New("the id value is an invalid type")
	ErrMisMatch = errors.New("token mismatch")
	ErrSession  = errors.New("no sub id in session")
	ErrTarget   = errors.New("target not found")
	ErrTmpl     = errors.New("the server could not render the html template for this page")
	ErrUser     = errors.New("unknown user")
	ErrZap      = errors.New("the zap logger cannot be nil")
)

// Caching are values that are used throughout the app or layouts.
var Caching = Cache{} //nolint:gochecknoglobals

// Attribute returns a formatted string of the roles for the given scener name.
func Attribute(write, code, art, music, name string) string {
	name = strings.ToLower(name)
	w, c, a, m := strings.Split(strings.ToLower(write), ","),
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
	const and = 2
	if len(match) == and {
		return strings.Join(match, " and ") + " attributions"
	}
	last := len(match) - 1
	match[last] = "and " + match[last]
	return strings.Join(match, ", ") + " attributions"
}

// Brief returns a human readable brief description of a release.
// Based on the platform, section, year and month.
func Brief(plat, sect any) template.HTML {
	p, s := "", ""
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
	if p == "" && s == "" {
		return template.HTML("an unknown release")
	}
	x := tags.Humanize(tags.TagByURI(p), tags.TagByURI(s))
	// x = helper.Capitalize(x)
	return template.HTML(x + ".")
}

// ByteFile returns a human readable string of the file count and bytes.
func ByteFile(cnt, bytes any) template.HTML {
	var s string
	switch val := cnt.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		p := message.NewPrinter(language.English)
		s = p.Sprintf("%d", i)
	default:
		s = fmt.Sprintf("%sByteFile: %s", typeErr, reflect.TypeOf(cnt).String())
		return template.HTML(s)
	}
	switch val := bytes.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		s = fmt.Sprintf("%s <small>(%s)</small>", s, helper.ByteCount(i))
	default:
		s = fmt.Sprintf("%sByteFile: %s", typeErr, reflect.TypeOf(bytes).String())
		return template.HTML(s)
	}
	return template.HTML(s)
}

// ByteFileS returns a human readable string of the byte count with a named description.
func ByteFileS(name string, cnt, bytes any) template.HTML {
	var s string
	switch val := cnt.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		name = names(name)
		if i != 1 {
			name += "s"
		}
		p := message.NewPrinter(language.English)
		s = p.Sprintf("%d", i)
	default:
		s = fmt.Sprintf("%sByteFileS: %s", typeErr, reflect.TypeOf(cnt).String())
		return template.HTML(s)
	}
	switch val := bytes.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		s = fmt.Sprintf("%s %s <small>(%s)</small>", s, name, helper.ByteCount(i))
	default:
		s = fmt.Sprintf("%sByteFileS: %s", typeErr, reflect.TypeOf(bytes).String())
		return template.HTML(s)
	}
	return template.HTML(s)
}

// Day returns a string of the day number from the day d number between 1 and 31.
func Day(d any) template.HTML {
	var s string
	switch val := d.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		if i == 0 {
			return ""
		}
		if i < 0 || i > 31 {
			s = fmt.Sprintf(" error: day out of range %d", i)
			return template.HTML(s)
		}
		s = fmt.Sprintf(" %d", i)
	default:
		s = fmt.Sprintf("%sDay: %s", typeErr, reflect.TypeOf(d).String())
	}
	return template.HTML(s)
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
	return template.HTML(desc(p, s, y, m))
}

// GlobTo returns the path to the template file.
func GlobTo(name string) string {
	// The path is relative to the embed.FS root and must not use the OS path separator.
	return strings.Join([]string{"view", "app", name}, "/")
}

// LastUpdated returns a string of the time since the given time t.
// The time is formatted as "Last updated 1 hour ago".
// If the time is not valid, an empty string is returned.
func LastUpdated(t any) string {
	const s = "Last updated"
	return Updated(t, s)
}

// LinkDownload creates a URL to link to the file download of the record.
func LinkDownload(id any, alertURL string) template.HTML {
	s, err := linkID(id, "d")
	if err != nil {
		return template.HTML(err.Error())
	}
	if alertURL != "" {
		return template.HTML(`<s class="card-link text-warning-emphasis" data-bs-toggle="tooltip" ` +
			`data-bs-title="Use the About link to access this file download">Download</s>`)
	}
	return template.HTML(fmt.Sprintf(`<a class="card-link" href="%s">Download</a>`, s))
}

// LinkHref creates a URL path to link to the file page for the record.
func LinkHref(id any) (string, error) {
	return linkID(id, "f")
}

// LinkPage creates a URL anchor element to link to the file page for the record.
func LinkPage(id any) template.HTML {
	s, err := linkID(id, "f")
	if err != nil {
		return template.HTML(err.Error())
	}
	return template.HTML(fmt.Sprintf(`<a class="card-link" href="%s">About</a>`, s))
}

// LinkPreview creates a URL to link to the file record in tab, to use as a preview.
// The preview link will only show with compatible file types based on their extension.
func LinkPreview(id any, name, platform string) template.HTML {
	if name == "" {
		return template.HTML("")
	}
	s := LinkPreviewHref(id, name, platform)
	if s == "" {
		return template.HTML("")
	}
	elm := fmt.Sprintf(`&nbsp; <a class="card-link" href="%s">Preview</a>`, s)
	return template.HTML(elm)
}

// LinkPreviewHref creates a URL path to link to the file record in tab, to use as a preview.
func LinkPreviewHref(id any, name, platform string) template.HTML {
	if name == "" {
		return template.HTML("")
	}
	platform = strings.TrimSpace(platform)
	// supported formats
	// https://developer.mozilla.org/en-US/docs/Web/Media/Formats/Image_types
	ext := strings.ToLower(filepath.Ext(name))
	switch {
	case slices.Contains(archives(), ext):
		// this must always be first
		return template.HTML("")
	case platform == textamiga, platform == "text":
		break
	case slices.Contains(documents(), ext):
		break
	case slices.Contains(images(), ext):
		break
	case slices.Contains(media(), ext):
		break
	default:
		return template.HTML("")
	}
	s, err := linkID(id, "v")
	if err != nil {
		return template.HTML(err.Error())
	}
	return template.HTML(s)
}

// LinkPreviewTip returns a tooltip to describe the preview link.
func LinkPreviewTip(name, platform string) string {
	if name == "" {
		return ""
	}
	platform = strings.TrimSpace(platform)
	ext := strings.ToLower(filepath.Ext(name))
	switch {
	case slices.Contains(archives(), ext):
		// this must always be first
		return ""
	case platform == textamiga, platform == "text":
		return "Read this as text"
	case slices.Contains(documents(), ext):
		return "Read this as text"
	case slices.Contains(images(), ext):
		return "View this as an image or photo"
	case slices.Contains(media(), ext):
		return "Play this as media"
	}
	return ""
}

// LinkRelFast returns the groups associated with a release and a link to each group.
// It is a faster version of LinkRelrs and should be used with the templates that have large lists of group names.
func LinkRelFast(a, b any) template.HTML {
	return LinkRelrs(true, a, b)
}

// LinkRelrs returns the groups associated with a release and a link to each group.
// The performant flag will use the group name instead of the much slower group slug formatter.
func LinkRelrs(performant bool, a, b any) template.HTML {
	const class = "text-nowrap link-offset-2 link-underline link-underline-opacity-25"
	var av, bv string
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
	av, bv = strings.TrimSpace(av), strings.TrimSpace(bv)
	prime, second := "", ""
	if av == "" && bv == "" {
		return template.HTML("error: unknown group")
	}
	if av != "" {
		ref, err := linkRelr(av)
		if err != nil {
			return template.HTML(fmt.Sprintf("error: %s", err))
		}
		x := helper.Capitalize(strings.ToLower(av))
		if !performant {
			x = releaser.Link(helper.Slug(av))
		}
		prime = fmt.Sprintf(`<a class="%s" href="%s">%s</a>`, class, ref, x)
	}
	if bv != "" {
		ref, err := linkRelr(bv)
		if err != nil {
			return template.HTML(fmt.Sprintf("error: %s", err))
		}
		x := helper.Capitalize(strings.ToLower(bv))
		if !performant {
			x = releaser.Link(helper.Slug(bv))
		}
		second = fmt.Sprintf(`<a class="%s" href="%s">%s</a>`, class, ref, x)
	}
	s := relHTML(prime, second)
	return template.HTML(s)
}

// LinkRelrs returns the groups associated with a release and a link to each group.
func LinkRels(a, b any) template.HTML {
	return LinkRelrs(false, a, b)
}

// LinkRemote returns a HTML link with an embedded SVG icon to an external website.
func LinkRemote(href, name string) template.HTML {
	if href == "" {
		return "error: href is empty"
	}
	if name == "" {
		return "error: name is empty"
	}
	a := fmt.Sprintf(`<a class="dropdown-item icon-link icon-link-hover link-light" href="%s">%s %s</a>`,
		href, name, link)
	return template.HTML(a)
}

// LinkScnr returns a link to the named scener page.
func LinkScnr(name string) (string, error) {
	href, err := url.JoinPath("/", "p", helper.Slug(name))
	if err != nil {
		return "", fmt.Errorf("name %q could not be made into a valid url: %w", name, err)
	}
	return href, nil
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
	a := fmt.Sprintf(`<a class="dropdown-item icon-link icon-link-hover link-light" href="%s">%s %s</a>`,
		href, name, wiki)
	return template.HTML(a)
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

// MimeMagic overrides some of the default linux file (magic) results.
// This is intended to be used to provide a more human readable description.
func MimeMagic(s string) template.HTML {
	x := strings.ToLower(s)
	const zipV1 = "zip archive data, at least v1.0 to extract"
	if strings.Contains(x, zipV1) {
		return template.HTML("Obsolete zip archive, implode or shrink")
	}
	const zipV2 = "zip archive data, at least v2.0 to extract"
	if strings.Contains(x, zipV2) {
		return template.HTML("Zip archive")
	}
	return template.HTML(s)
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

// Month returns a string of the month name from the month m number between 1 and 12.
func Month(m any) template.HTML {
	var s string
	switch val := m.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		if i == 0 {
			return ""
		}
		if i < 0 || i > 12 {
			s = fmt.Sprintf(" error: month out of range %d", i)
			return template.HTML(s)
		}
		s = " " + time.Month(i).String()
	default:
		s = fmt.Sprintf("%sFmtMonth: %s", typeErr, reflect.TypeOf(m).String())
	}
	return template.HTML(s)
}

// OptionsAnsiLove returns a list of possible text or ANSI files in the archive content.
// In general, all file extensions are valid except for well known archives and executables.
// Due to the CP/M and DOS platform, 8.3 filename limitations, the file extension is not always reliable.
func OptionsAnsiLove(zipContent string) template.HTML {
	list := strings.Split(zipContent, "\n")
	s := ""
	for _, v := range list {
		x := strings.TrimSpace(strings.ToLower(v))
		switch filepath.Ext(x) {
		case ".com", ".exe", ".dll", gif, png, jpg, jpeg, webp, ".bmp",
			".ico", ".avi", ".mpg", ".mpeg", ".mp1", ".mp2", ".mp3", ".mp4", ".ogg", ".wmv",
			".zip", ".arc", ".arj", ".ace", ".lha", ".lzh", ".7z", ".tar", ".gz", ".bz2", ".xz", ".z",
			".───", ".──-", ".-", ".--", ".---":
			continue
		}
		s += fmt.Sprintf("<option>%s</option>", v)
	}
	return template.HTML(s)
}

// OptionsPreview returns a list of preview images or textfiles in the archive content.
func OptionsPreview(zipContent string) template.HTML {
	list := strings.Split(zipContent, "\n")
	s := ""
	for _, v := range list {
		x := strings.TrimSpace(strings.ToLower(v))
		switch filepath.Ext(x) {
		case gif, png, jpg, jpeg, webp, ".bmp":
			s += fmt.Sprintf("<option>%s</option>", v)
		}
	}
	return template.HTML(s)
}

// OptionsReadme returns a list of readme and known textfiles in the archive content.
func OptionsReadme(zipContent string) template.HTML {
	list := strings.Split(zipContent, "\n")
	s := ""
	for _, v := range list {
		x := strings.TrimSpace(strings.ToLower(v))
		switch filepath.Ext(x) {
		case ".txt", ".nfo", ".diz", ".me", ".asc", ".doc":
			s += fmt.Sprintf("<option>%s</option>", v)
			continue
		}
		x = strings.ToLower(v)
		if strings.Contains(x, "readme") {
			s += fmt.Sprintf("<option>%s</option>", v)
			continue
		}
	}
	return template.HTML(s)
}

// Cache contains database values that are used throughout the app or layouts.
type Cache struct {
	RecordCount int // The total number of file records in the database.
}

// SRI are the Subresource Integrity hashes for the layout.
type SRI struct {
	Bootstrap   string // Bootstrap CSS verification hash.
	BootstrapJS string // Bootstrap JS verification hash.
	Editor      string // Editor JS verification hash.
	EditAssets  string // Editor Assets JS verification hash.
	EditArchive string // Editor Archive JS verification hash.
	FontAwesome string // Font Awesome verification hash.
	JSDosUI     string // JS DOS verification hash.
	JSDosW      string // JS DOS emscripten verification hash.
	Layout      string // Layout CSS verification hash.
	Pouet       string // Pouet JS verification hash.
	Readme      string // Readme JS verification hash.
	RESTPouet   string // Pouet REST JS verification hash.
	RESTZoo     string // Demozoo REST JS verification hash.
	Uploader    string // Uploader JS verification hash.
}

// Verify checks the integrity of the embedded CSS and JS files.
// These are required for Subresource Integrity (SRI) verification in modern browsers.
func (s *SRI) Verify(fs embed.FS) error {
	names := Names()
	var err error
	s.Bootstrap, err = helper.Integrity(names[Bootstrap], fs)
	if err != nil {
		return err
	}
	s.BootstrapJS, err = helper.Integrity(names[BootstrapJS], fs)
	if err != nil {
		return err
	}
	s.Editor, err = helper.Integrity(names[Editor], fs)
	if err != nil {
		return err
	}
	s.EditAssets, err = helper.Integrity(names[EditAssets], fs)
	if err != nil {
		return err
	}
	s.EditArchive, err = helper.Integrity(names[EditArchive], fs)
	if err != nil {
		return err
	}
	s.FontAwesome, err = helper.Integrity(names[FontAwesome], fs)
	if err != nil {
		return err
	}
	s.JSDosUI, err = helper.Integrity(names[JSDosUI], fs)
	if err != nil {
		return err
	}
	s.JSDosW, err = helper.Integrity(names[JSDosW], fs)
	if err != nil {
		return err
	}
	s.Layout, err = helper.Integrity(names[Layout], fs)
	if err != nil {
		return err
	}
	s.Pouet, err = helper.Integrity(names[Pouet], fs)
	if err != nil {
		return err
	}
	s.Readme, err = helper.Integrity(names[Readme], fs)
	if err != nil {
		return err
	}
	s.RESTPouet, err = helper.Integrity(names[RESTPouet], fs)
	if err != nil {
		return err
	}
	s.RESTZoo, err = helper.Integrity(names[RESTZoo], fs)
	if err != nil {
		return err
	}
	s.Uploader, err = helper.Integrity(names[Uploader], fs)
	if err != nil {
		return err
	}
	return nil
}
