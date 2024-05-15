// Package app handles the routes and views for the Defacto2 website.
package app

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"cmp"
	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"slices"
)

const (
	// Welcome is the default logo monospace text,
	// each side contains 20 whitespace characters.
	// The welcome to defacto2 text is 19 characters long.
	// The letter 'O' of TO is the center of the text.
	Welcome = `:                    ` +
		`·· WELCOME TO DEFACTO2 ··` +
		`                    ·`
)

var (
	ErrCategory = errors.New("unknown artifacts categories")
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
	ErrVal      = errors.New("value is empty")
	ErrZap      = errors.New("the zap logger cannot be nil")
)

func errVal(name string) template.HTML {
	return template.HTML(fmt.Sprintf("error, %s: %s", ErrVal, name))
}

const (
	attr  = " attributions"
	br    = "<br>"
	div1  = "</div>"
	sect0 = "<section>"
	sect1 = "</section>"
	ul0   = "<ul>"
	ul1   = "</ul>"
	fzip  = ".zip"
)

// Caching are values that are used throughout the app or layouts.
var Caching = Cache{} //nolint:gochecknoglobals

// Records caches the database record count.
func (c *Cache) Records(i int) {
	c.RecordCount = i
}

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
		return strings.Join(match, " and ") + attr
	}
	last := len(match) - 1
	match[last] = "and " + match[last]
	return strings.Join(match, ", ") + attr
}

// Brief returns a human readable brief description of a release.
// Based on the platform, section, year and month.
func Brief(platform, section any) string {
	p, s := "", ""
	switch val := platform.(type) {
	case string:
		p = val
	case null.String:
		if val.Valid {
			p = val.String
		}
	default:
		s := fmt.Sprintf("%s %s %T", typeErr, "describe", platform)
		return s
	}
	p = strings.TrimSpace(p)
	switch val := section.(type) {
	case string:
		s = val
	case null.String:
		if val.Valid {
			s = val.String
		}
	default:
		s := fmt.Sprintf("%s %s %T", typeErr, "describe", section)
		return s
	}
	s = strings.TrimSpace(s)
	if p == "" && s == "" {
		return "an unknown release"
	}
	x := tags.Humanize(tags.TagByURI(p), tags.TagByURI(s)) + "."
	return x
}

// ByteFile returns a human readable string of the file count and bytes.
func ByteFile(cnt, bytes any) template.HTML {
	var s string
	switch val := cnt.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		p := message.NewPrinter(language.English)
		s = p.Sprintf("%d", i)
	default:
		s = fmt.Sprintf("%sByteFile: %s", typeErr, reflect.TypeOf(cnt).String())
		return template.HTML(s)
	}
	switch val := bytes.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		s = fmt.Sprintf("%s <small>(%s)</small>", s, helper.ByteCount(i))
	default:
		s = fmt.Sprintf("%sByteFile: %s", typeErr, reflect.TypeOf(bytes).String())
		return template.HTML(s)
	}
	return template.HTML(s)
}

// ByteFileS returns a human readable string of the byte count with a named description.
func ByteFileS(name string, count, bytes any) template.HTML {
	var s string
	switch val := count.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		name = names(name)
		if i != 1 {
			name += "s"
		}
		p := message.NewPrinter(language.English)
		s = p.Sprintf("%d", i)
	default:
		s = fmt.Sprintf("%sByteFileS: %s", typeErr, reflect.TypeOf(count).String())
		return template.HTML(s)
	}
	switch val := bytes.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		s = fmt.Sprintf("%s %s <small>(%s)</small>", s, name, helper.ByteCount(i))
	default:
		s = fmt.Sprintf("%sByteFileS: %s", typeErr, reflect.TypeOf(bytes).String())
		return template.HTML(s)
	}
	return template.HTML(s)
}

// Day returns a string of the day number from the day d number between 1 and 31.
func Day(d any) string {
	var s string
	switch val := d.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		if i == 0 {
			return ""
		}
		if i < 0 || i > 31 {
			s = fmt.Sprintf(" error: day out of range %d", i)
			return s
		}
		s = fmt.Sprintf(" %d", i)
	default:
		s = fmt.Sprintf("%sDay: %s", typeErr, reflect.TypeOf(d).String())
	}
	return s
}

// Describe returns a human readable description of a release.
// Based on the platform, section, year and month.
func Describe(platform, section, year, month any) template.HTML {
	const tmpl = "describe"
	p, s, y, m := "", "", "", ""
	switch val := platform.(type) {
	case string:
		p = val
	case null.String:
		if val.Valid {
			p = val.String
		}
	default:
		return template.HTML(fmt.Sprintf("%s %s %s", typeErr, tmpl, platform))
	}
	p = strings.TrimSpace(p)
	switch val := section.(type) {
	case string:
		s = val
	case null.String:
		if val.Valid {
			s = val.String
		}
	default:
		return template.HTML(fmt.Sprintf("%s %s %s", typeErr, tmpl, section))
	}
	s = strings.TrimSpace(s)
	switch val := year.(type) {
	case int, int8, int16, int32, int64:
		y = fmt.Sprintf("%v", val)
	case null.Int16:
		if val.Valid {
			y = strconv.Itoa(int(val.Int16))
		}
	default:
		return template.HTML(fmt.Sprintf("%s %s %s", typeErr, tmpl, year))
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
		return template.HTML(fmt.Sprintf("%s %s %s", typeErr, tmpl, month))
	}
	return template.HTML(desc(p, s, y, m))
}

// GlobTo returns the path to the template file.
func GlobTo(name string) string {
	const pathSeparator = "/"
	return strings.Join([]string{"view", "app", name}, pathSeparator)
}

// LastUpdated returns a string of the time since the given time t.
// The time is formatted as "Last updated 1 hour ago".
// If the time is not valid, an empty string is returned.
func LastUpdated(t any) string {
	if t == nil {
		return ""
	}
	const s = "Last updated"
	return Updated(t, s)
}

// LinkDownload creates a URL to link to the file download of the record.
func LinkDownload(id any, uri string) template.HTML {
	if id == nil {
		return ""
	}
	s, err := linkID(id, "d")
	if err != nil {
		return template.HTML(err.Error())
	}
	if uri != "" {
		return template.HTML(`<s class="card-link text-warning-emphasis" data-bs-toggle="tooltip" ` +
			`data-bs-title="Use the link to access this file download">Download</s>`)
	}
	return template.HTML(fmt.Sprintf(`<a class="card-link" href="%s">Download</a>`, s))
}

// LinkHref creates a URL path to link to the file page for the record.
func LinkHref(id any) (string, error) {
	if id == nil {
		return "", fmt.Errorf("id is nil, %w", ErrLinkID)
	}
	return linkID(id, "f")
}

// LinkInterview returns a SVG arrow icon to indicate an interview link hosted on an external website.
func LinkInterview(href string) template.HTML {
	if href == "" {
		return errVal("href")
	}
	p, err := url.Parse(href)
	if err != nil || p.Scheme == "" {
		// if href is not a valid URL, then it is a relative path to the site.
		return template.HTML("")
	}
	return arrowLink
}

// LinkPage creates a URL anchor element to link to the file page for the record.
func LinkPage(id any) template.HTML {
	if id == nil {
		return ""
	}
	s, err := linkID(id, "f")
	if err != nil {
		return template.HTML(err.Error())
	}
	return template.HTML(fmt.Sprintf(`<a class="card-link" href="%s">Artifact</a>`, s))
}

// LinkPreview creates a URL to link to the file record in tab, to use as a preview.
// The preview link will only show with compatible file types based on their extension.
func LinkPreview(id any, name, platform string) template.HTML {
	if id == nil || name == "" {
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
//
// A list of supported file types: https://developer.mozilla.org/en-US/docs/Web/Media/Formats/Image_types
func LinkPreviewHref(id any, name, platform string) string {
	if id == nil || name == "" {
		return ""
	}
	platform = strings.TrimSpace(platform)
	// supported formats
	// https://developer.mozilla.org/en-US/docs/Web/Media/Formats/Image_types
	ext := strings.ToLower(filepath.Ext(name))
	switch {
	case slices.Contains(archives(), ext):
		// this must always be first
		return ""
	case platform == textamiga, platform == "text":
		break
	case slices.Contains(documents(), ext):
		break
	case slices.Contains(images(), ext):
		break
	case slices.Contains(media(), ext):
		break
	default:
		return ""
	}
	s, err := linkID(id, "v")
	if err != nil {
		return fmt.Sprint("error: ", err)
	}
	return s
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
		// this case must always be first
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
	if a == nil || b == nil {
		return ""
	}
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
	if av == "" && bv != "" {
		av = bv
		bv = ""
	}

	var prime, second string
	var err error
	if av != "" {
		prime, err = makeLink(av, class, performant)
		if err != nil {
			return template.HTML(fmt.Sprintf("error: %s", err))
		}
	}
	if bv != "" {
		second, err = makeLink(bv, class, performant)
		if err != nil {
			return template.HTML(fmt.Sprintf("error: %s", err))
		}
	}
	return relHTML(prime, second)
}

func makeLink(name, class string, performant bool) (string, error) {
	ref, err := linkRelr(name)
	if err != nil {
		return "", fmt.Errorf("linkRelr: %w", err)
	}
	x := helper.Capitalize(strings.ToLower(name))
	title := x
	if !performant {
		title = releaser.Link(helper.Slug(name))
	}
	s := fmt.Sprintf(`<a class="%s" href="%s">%s</a>`, class, ref, title)
	if x != "" && title == "" {
		s = "error: could not link group"
	}
	return s, nil
}

// LinkRelrs returns the groups associated with a release and a link to each group.
func LinkRels(a, b any) template.HTML {
	if a == nil || b == nil {
		return ""
	}
	return LinkRelrs(false, a, b)
}

// LinkRemote returns a HTML link with an embedded SVG icon to an external website.
func LinkRemote(href, name string) template.HTML {
	if href == "" {
		return errVal("href")
	}
	if name == "" {
		return errVal("name")
	}
	a := fmt.Sprintf(`<a class="dropdown-item icon-link icon-link-hover link-light" href="%s">%s %s</a>`,
		href, name, arrowLink)
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

// LinkSVG returns an right-arrow SVG icon.
func LinkSVG() template.HTML {
	return arrowLink
}

// LinkWiki returns a HTML link with an embedded SVG icon to the Defacto2 wiki on GitHub.
func LinkWiki(uri, name string) template.HTML {
	if uri == "" {
		return errVal("uri")
	}
	if name == "" {
		return errVal("name")
	}
	href, err := url.JoinPath("https://github.com/Defacto2/defacto2.net/wiki/", uri)
	if err != nil {
		return template.HTML(err.Error())
	}
	a := fmt.Sprintf(`<a class="dropdown-item icon-link icon-link-hover link-light" href="%s">%s %s</a>`,
		href, name, arrowLink)
	return template.HTML(a)
}

// LogoText returns a string of text padded with spaces to center it in the logo.
func LogoText(s string) string {
	const spaces = 6
	indent := strings.Repeat(" ", spaces)
	if s == "" {
		return indent + Welcome
	}
	// odd returns true if the given integer is odd.
	odd := func(i int) bool {
		return i%2 != 0
	}
	const padder = " ·· "
	const wl, pl = len(Welcome), len(padder)
	const limit = wl - (pl + pl) - 3
	s = strings.ToUpper(s)

	truncateStr := len(s) > limit
	if truncateStr {
		return fmt.Sprintf("%s:%s%s%s·",
			indent, padder, s[:limit], padder)
	}
	styled := fmt.Sprintf("%s%s%s", padder, s, padder)
	if !odd(len(s)) {
		styled = fmt.Sprintf(" %s%s%s", padder, s, padder)
	}
	const split = 2
	padding := (wl / split) - (len(styled) / split) - split
	text := fmt.Sprintf(":%s%s%s·",
		strings.Repeat(" ", padding),
		styled,
		strings.Repeat(" ", padding))
	return indent + text
}

// MimeMagic overrides some of the default linux file (magic) results.
// This is intended to be used to provide a more human readable description.
func MimeMagic(s string) string {
	x := strings.ToLower(s)
	const zipV1 = "zip archive data, at least v1.0 to extract"
	if strings.Contains(x, zipV1) {
		return "Obsolete zip archive, implode or shrink"
	}
	const zipV2 = "zip archive data, at least v2.0 to extract"
	if strings.Contains(x, zipV2) {
		return "Zip archive"
	}
	return s
}

// Mod returns true if the given integer is a multiple of the given max integer.
func Mod(i any, max int) bool {
	if max == 0 {
		return false
	}
	switch val := i.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
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
func Month(m any) string {
	if m == nil {
		return ""
	}
	var s string
	switch val := m.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		if i == 0 {
			return ""
		}
		if i < 0 || i > 12 {
			s = fmt.Sprintf(" error: month out of range %d", i)
			return s
		}
		s = " " + time.Month(i).String()
	default:
		s = fmt.Sprintf("%sFmtMonth: %s", typeErr, reflect.TypeOf(m).String())
	}
	return s
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
			fzip, ".arc", ".arj", ".ace", ".lha", ".lzh", ".7z", ".tar", ".gz", ".bz2", ".xz", ".z",
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

// Prefix returns a string prefixed with a space.
func Prefix(s string) string {
	if s == "" {
		return ""
	}
	return s + " "
}

// RecordsSub returns the records for the artifacts category URI.
func RecordsSub(uri string) string { //nolint:cyclop
	const ignore = -1
	switch Match(uri) {
	case advert:
		return tags.Humanizes(ignore, tags.ForSale)
	case announcement:
		return tags.Humanizes(ignore, tags.Announcement)
	case ansi:
		return tags.Humanizes(tags.ANSI, ignore)
	case ansiBrand:
		return tags.Humanizes(tags.ANSI, tags.Logo)
	case ansiBBS:
		return tags.Humanizes(tags.ANSI, tags.BBS)
	case ansiFTP:
		return tags.Humanizes(tags.ANSI, tags.Ftp)
	case ansiNfo:
		return tags.Humanizes(tags.ANSI, tags.Nfo)
	case ansiPack:
		return tags.Humanizes(tags.ANSI, tags.Pack)
	case bbs:
		return tags.Humanizes(ignore, tags.BBS)
	case bbsImage:
		return tags.Humanizes(tags.Image, tags.BBS)
	case bbstro:
		return tags.Humanizes(tags.DOS, tags.BBS)
	case bbsText:
		return tags.Humanizes(tags.Text, tags.BBS)
	case database:
		return tags.Humanizes(ignore, tags.DataB)
	case demoscene:
		return tags.Humanizes(ignore, tags.Demo)
	case drama:
		return tags.Humanizes(ignore, tags.Drama)
	case ftp:
		return tags.Humanizes(ignore, tags.Ftp)
	case hack:
		return tags.Humanizes(ignore, tags.GameHack)
	}
	return recordsSub0(uri)
}

func recordsSub0(uri string) string {
	const ignore = -1
	switch Match(uri) {
	case htm:
		return uri
	case howTo:
		return tags.Humanizes(ignore, tags.Guide)
	case imageFile:
		return tags.Humanizes(tags.Image, ignore)
	case imagePack:
		return tags.Humanizes(tags.Image, tags.Pack)
	case installer:
		return tags.Humanizes(ignore, tags.Install)
	case intro:
		return tags.Humanizes(ignore, tags.Intro)
	case linux:
		return tags.Humanizes(tags.Linux, ignore)
	case java:
		return tags.Humanizes(tags.Java, ignore)
	case jobAdvert:
		return tags.Humanizes(ignore, tags.Job)
	case macos:
		return tags.Humanizes(tags.Mac, ignore)
	case msdosPack:
		return tags.Humanizes(tags.DOS, tags.Pack)
	case music:
		return tags.Humanizes(tags.Audio, ignore)
	case newsArticle:
		return tags.Humanizes(ignore, tags.News)
	case nfo:
		return tags.Humanizes(ignore, tags.Nfo)
	case nfoTool:
		return tags.Humanizes(ignore, tags.NfoTool)
	}
	return recordsSub1(uri)
}

func recordsSub1(uri string) string { //nolint:cyclop
	const ignore = -1
	switch Match(uri) {
	case standards:
		return tags.Humanizes(ignore, tags.Rule)
	case script:
		return tags.Humanizes(tags.PHP, ignore)
	case introMsdos:
		return tags.Humanizes(tags.DOS, tags.Intro)
	case introWindows:
		return tags.Humanizes(tags.Windows, tags.Intro)
	case magazine:
		return tags.Humanizes(ignore, tags.Mag)
	case msdos:
		return tags.Humanizes(tags.DOS, ignore)
	case pdf:
		return tags.Humanizes(tags.PDF, ignore)
	case proof:
		return tags.Humanizes(ignore, tags.Proof)
	case restrict:
		return tags.Humanizes(ignore, tags.Restrict)
	case takedown:
		return tags.Humanizes(ignore, tags.Bust)
	case text:
		return tags.Humanizes(tags.Text, ignore)
	case textAmiga:
		return tags.Humanizes(tags.TextAmiga, ignore)
	case textApple2:
		return tags.Humanizes(tags.Text, tags.AppleII)
	case textAtariST:
		return tags.Humanizes(tags.Text, tags.AtariST)
	case textPack:
		return tags.Humanizes(tags.Text, tags.Pack)
	case tool:
		return tags.Humanizes(ignore, tags.Tool)
	case trialCrackme:
		return tags.Humanizes(tags.Windows, tags.Job)
	case video:
		return tags.Humanizes(tags.Video, ignore)
	case windows:
		return tags.Humanizes(tags.Windows, ignore)
	case windowsPack:
		return tags.Humanizes(tags.Windows, tags.Pack)
	default:
		return "unknown uri"
	}
}

// ReadmeSuggest returns a suggested readme file name for the record.
// It prioritizes the filename and group name with a priority extension,
// such as ".nfo", ".txt", etc. If no priority extension is found,
// it will return the first textfile in the content list.
//
// The filename should be the name of the file archive artifact.
// The group should be a name or common abbreviation of the group that
// released the artifact. The content should be a list of files contained
// in the artifact.
//
// This is a port of the CFML function, variables.findTextfile found in File.cfc.
func ReadmeSuggest(filename, group string, content ...string) string {
	finds := Readmes(content...)
	if len(finds) == 1 {
		return finds[0]
	}
	finds = SortContent(finds...)

	// match either the filename or the group name with a priority extension
	// e.g. .nfo, .txt, .unp, .doc
	base := filepath.Base(filename)
	for _, ext := range priority() {
		for _, name := range finds {
			if strings.EqualFold(base+ext, name) {
				return name
			}
			if strings.EqualFold(group+ext, name) {
				return name
			}
		}
	}
	const matchFileID = "file_id.diz"
	for _, name := range finds {
		if strings.EqualFold(matchFileID, name) {
			return name
		}
	}
	// match either the filename or the group name with a candidate extension
	for _, ext := range candidate() {
		for _, name := range finds {
			if strings.EqualFold(base+ext, name) {
				return name
			}
			if strings.EqualFold(group+ext, name) {
				return name
			}
		}
	}
	// match any finds that use a priority extension
	for _, name := range finds {
		s := strings.ToLower(name)
		ext := filepath.Ext(s)
		if slices.Contains(priority(), ext) {
			return name
		}
	}
	// match the first file in the list
	for _, name := range finds {
		return name
	}
	return ""
}

// RecordRels returns the groups associated with a release and joins them with a plus sign.
func RecordRels(a, b any) string {
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
	av = strings.TrimSpace(av)
	bv = strings.TrimSpace(bv)
	switch {
	case av != "" && bv != "":
		s = strings.Join([]string{av, bv}, " + ")
	case av != "":
		s = av
	case bv != "":
		s = bv
	}
	return s
}

func RecordReleasers(a, b any) [2]string {
	av, bv := "", ""
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
	av = strings.TrimSpace(av)
	bv = strings.TrimSpace(bv)
	switch {
	case av != "" && bv != "":
		return [2]string{av, bv}
	case bv != "":
		return [2]string{bv, ""}
	case av != "":
		return [2]string{av, ""}
	}
	return [2]string{}
}

// SafeHTML returns a string as a template.HTML type.
// This is intended to be used to prevent HTML escaping.
func SafeHTML(s string) template.HTML {
	return template.HTML(s)
}

// SafeJS returns a string as a template.JS type.
// This is intended to be used to prevent JS escaping.
func SafeJS(s string) template.JS {
	return template.JS(s)
}

// SortContent sorts the content list by the number of slashes in each string.
// It prioritizes strings with fewer slashes (i.e., closer to the root).
// If the number of slashes is the same, it sorts alphabetically.
func SortContent(content ...string) []string {
	const windowsPath = "\\"
	const pathSeparator = "/"
	slices.SortFunc(content, func(a, b string) int {
		a = strings.ReplaceAll(a, windowsPath, pathSeparator)
		b = strings.ReplaceAll(b, windowsPath, pathSeparator)
		aCount := strings.Count(a, pathSeparator)
		bCount := strings.Count(b, pathSeparator)
		if aCount != bCount {
			return aCount - bCount
		}
		return cmp.Compare(strings.ToLower(a), strings.ToLower(b))
	})
	return content
}

// SubTitle returns a secondary element with the record title.
func SubTitle(section null.String, s any) template.HTML {
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
	if strings.TrimSpace(strings.ToLower(section.String)) == "magazine" {
		if i, err := strconv.Atoi(val); err == nil {
			val = fmt.Sprintf("Issue %d", i)
		}
	}
	cls := "card-subtitle mb-2 text-body-secondary fs-6"
	elem := fmt.Sprintf("<h3 class=\"%s\">%s</h3>", cls, val)
	return template.HTML(elem)
}

// TagBrief returns a small summary of the tag.
func TagBrief(tag string) string {
	t := tags.TagByURI(tag)
	s := tags.Infos()[t]
	return s
}

// TagOption returns a HTML option tag with the selected attribute if the selected matches the value.
func TagOption(selected, value any) template.HTML {
	sel, val := "", ""
	switch i := selected.(type) {
	case string:
		sel = reflect.ValueOf(i).String()
	case null.String:
		if i.Valid {
			sel = i.String
		}
	}
	switch i := value.(type) {
	case string:
		val = reflect.ValueOf(i).String()
	case null.String:
		if i.Valid {
			val = i.String
		}
	}
	sel = strings.TrimSpace(sel)
	val = strings.TrimSpace(val)
	if sel != "" && sel == val {
		return template.HTML(fmt.Sprintf("<option value=\"%s\" selected>", val))
	}
	return template.HTML(fmt.Sprintf("<option value=\"%s\">", val))
}

// TagWithOS returns a small summary of the tag with the operating system.
func TagWithOS(os, tag string) string {
	p, t := tags.TagByURI(os), tags.TagByURI(tag)
	s := tags.Humanize(p, t)
	return s
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

// TrimSpace returns a string with all leading and trailing whitespace removed.
func TrimSpace(a any) string {
	if a == nil {
		return ""
	}
	switch val := a.(type) {
	case string:
		return strings.TrimSpace(val)
	case null.String:
		if val.Valid {
			return strings.TrimSpace(val.String)
		}
		return ""
	default:
		return fmt.Sprintf("%sTrimSpace: %s", typeErr, reflect.TypeOf(a).String())
	}
}

// Updated returns a string of the time since the given time t.
// The time is formatted as "Last updated 1 hour ago".
// If the time is not valid, an empty string is returned.
func Updated(t any, s string) string {
	if t == nil {
		return ""
	}
	if s == "" {
		s = "Time"
	}
	switch val := t.(type) {
	case null.Time:
		if !val.Valid {
			return ""
		}
		return fmt.Sprintf("%s %s ago", s, helper.TimeDistance(val.Time, time.Now(), true))
	case time.Time:
		return fmt.Sprintf("%s %s ago", s, helper.TimeDistance(val, time.Now(), true))
	default:
		return fmt.Sprintf("%supdated: %s", typeErr, reflect.TypeOf(t).String())
	}
}

// websiteIcon returns a Bootstrap icon name for the given website url.
func WebsiteIcon(url string) template.HTML {
	icon := websiteIcon(url)
	if icon == "arrow-right" {
		return `<svg class="bi" aria-hidden="true"><use xlink:href="/bootstrap-icons.svg#arrow-right"></use></svg>`
	}
	return template.HTML(fmt.Sprintf(`<svg class="bi" aria-hidden="true">`+
		`<use xlink:href="/bootstrap-icons.svg#%s"></use></svg>`, icon))
}

func websiteIcon(url string) string {
	switch {
	case strings.Contains(url, "archive.org"):
		return "bank2"
	case strings.Contains(url, "reddit.com"):
		return "reddit"
	case strings.Contains(url, "yalebooks.yale.edu"),
		strings.Contains(url, "explodingthephone.com"),
		strings.Contains(url, "punctumbooks"):
		return "book"
	case strings.Contains(url, "youtube.com"):
		return "youtube"
	case strings.Contains(url, "vimeo.com"):
		return "vimeo"
	case strings.Contains(url, "slashdot.org"):
		return "slash-circle"
	}
	return "arrow-right"
}

// YMDEdit handles the post submission for the Year, Month, Day selection fields.
func YMDEdit(c echo.Context) error {
	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	r, err := model.Edit(f.ID)
	if err != nil {
		return fmt.Errorf("model.EditFind: %w", err)
	}
	y := model.ValidY(f.Year)
	m := model.ValidM(f.Month)
	d := model.ValidD(f.Day)
	if err = model.UpdateYMD(int64(f.ID), y, m, d); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// Cache contains database values that are used throughout the app or layouts.
type Cache struct {
	RecordCount int // The total number of file records in the database.
}

// SRI are the Subresource Integrity hashes for the layout.
type SRI struct {
	ArtifactEditor  string // Artifact Editor JS verification hash.
	Bootstrap5      string // Bootstrap CSS verification hash.
	Bootstrap5JS    string // Bootstrap JS verification hash.
	EditAssets      string // Editor Assets JS verification hash.
	EditArchive     string // Editor Archive JS verification hash.
	EditForApproval string // Editor For Approval JS verification hash.
	FA5Pro          string // Font Awesome Pro v5 verification hash.
	Jsdos6JS        string // js-dos v6 verification hash.
	DosboxJS        string // DOSBox Emscripten verification hash.
	Layout          string // Layout CSS verification hash.
	LayoutJS        string // Layout JS verification hash.
	Pouet           string // Pouet JS verification hash.
	Readme          string // Readme JS verification hash.
	Uploader        string // Uploader JS verification hash.
	Htmx            string // htmx JS verification hash.
}

// Verify checks the integrity of the embedded CSS and JS files.
// These are required for Subresource Integrity (SRI) verification in modern browsers.
func (s *SRI) Verify(fs embed.FS) error { //nolint:funlen
	names := Names()
	var err error
	s.ArtifactEditor, err = helper.Integrity(names[ArtifactEditor], fs)
	if err != nil {
		return fmt.Errorf("helper.Integrity: %w", err)
	}
	s.Bootstrap5, err = helper.Integrity(names[Bootstrap5], fs)
	if err != nil {
		return fmt.Errorf("helper.Integrity: %w", err)
	}
	s.Bootstrap5JS, err = helper.Integrity(names[Bootstrap5JS], fs)
	if err != nil {
		return fmt.Errorf("helper.Integrity: %w", err)
	}
	s.LayoutJS, err = helper.Integrity(names[LayoutJS], fs)
	if err != nil {
		return fmt.Errorf("helper.Integrity: %w", err)
	}
	s.EditAssets, err = helper.Integrity(names[EditAssets], fs)
	if err != nil {
		return fmt.Errorf("helper.Integrity: %w", err)
	}
	s.EditArchive, err = helper.Integrity(names[EditArchive], fs)
	if err != nil {
		return fmt.Errorf("helper.Integrity: %w", err)
	}
	s.EditForApproval, err = helper.Integrity(names[EditForApproval], fs)
	if err != nil {
		return fmt.Errorf("helper.Integrity: %w", err)
	}
	s.FA5Pro, err = helper.Integrity(names[FA5Pro], fs)
	if err != nil {
		return fmt.Errorf("helper.Integrity: %w", err)
	}
	s.Jsdos6JS, err = helper.Integrity(names[Jsdos6JS], fs)
	if err != nil {
		return fmt.Errorf("helper.Integrity: %w", err)
	}
	s.DosboxJS, err = helper.Integrity(names[DosboxJS], fs)
	if err != nil {
		return fmt.Errorf("helper.Integrity: %w", err)
	}
	s.Layout, err = helper.Integrity(names[Layout], fs)
	if err != nil {
		return fmt.Errorf("helper.Integrity: %w", err)
	}
	s.Pouet, err = helper.Integrity(names[Pouet], fs)
	if err != nil {
		return fmt.Errorf("helper.Integrity: %w", err)
	}
	s.Readme, err = helper.Integrity(names[Readme], fs)
	if err != nil {
		return fmt.Errorf("helper.Integrity: %w", err)
	}
	s.Uploader, err = helper.Integrity(names[Uploader], fs)
	if err != nil {
		return fmt.Errorf("helper.Integrity: %w", err)
	}
	s.Htmx, err = helper.Integrity(names[Htmx], fs)
	if err != nil {
		return fmt.Errorf("helper.Integrity: %w", err)
	}
	return nil
}

// badRequest returns a JSON response with a 400 status code,
// the server cannot or will not process the request due to something that is perceived to be a client error.
func badRequest(c echo.Context, err error) error {
	return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request " + err.Error()})
}

func desc(p, s, y, m string) string {
	if p == "" && s == "" {
		return "An unknown release."
	}
	x := tags.Humanize(tags.TagByURI(p), tags.TagByURI(s))
	x = helper.Capitalize(x)
	if m != "" && y != "" {
		x = fmt.Sprintf("%s published in <span class=\"text-nowrap\">%s, %s</a>", x, m, y)
	} else if y != "" {
		x = fmt.Sprintf("%s published in %s", x, y)
	}
	return x + "."
}

// fileInfo is a helper function for Files that returns the page title, h1 title and lead text.
func fileInfo(uri string) (string, string, string) {
	var logo, h1sub, lead string
	switch uri {
	case newUploads.String():
		logo = "new uploads"
		h1sub = "the new uploads"
		lead = "These are the recent file artifacts that have been submitted to Defacto2."
	case newUpdates.String():
		logo = "new changes"
		h1sub = "the new changes"
		lead = "These are the recent file artifacts that have been modified or submitted on Defacto2."
	case forApproval.String():
		logo = "new uploads"
		h1sub = "edit the new uploads"
		lead = "These are the recent file artifacts that have been submitted for approval on Defacto2."
	case deletions.String():
		logo = "deletions"
		h1sub = "edit the (hidden) deletions"
		lead = "These are the file artifacts that have been removed from Defacto2."
	case unwanted.String():
		logo = "unwanted releases"
		h1sub = "edit the unwanted software releases"
		lead = "These are the file artifacts that have been marked as potential unwanted software " +
			"or containing viruses on Defacto2."
	case oldest.String():
		logo = "oldest releases"
		h1sub = "the oldest releases"
		lead = "These are the earliest, historical file artifacts in the collection."
	case newest.String():
		logo = "newest releases"
		h1sub = "the newest releases"
		lead = "These are the most recent file artifacts in the collection."
	default:
		s := RecordsSub(uri)
		h1sub = s
		logo = s
	}
	return logo, h1sub, lead
}

// linkID creates a URL to link to the record.
// The id is obfuscated to prevent direct linking.
// The elem is the element to link to, such as 'f' for file or 'd' for download.
func linkID(id any, elem string) (string, error) {
	var i int64
	switch val := id.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i = reflect.ValueOf(val).Int()
		if i <= 0 {
			return "", fmt.Errorf("%w: %d", ErrLinkID, i)
		}
	default:
		return "", fmt.Errorf("%w: %s", ErrLinkType, reflect.TypeOf(id).String())
	}
	href, err := url.JoinPath("/", elem, helper.ObfuscateID(i))
	if err != nil {
		return "", fmt.Errorf("id %d could not be made into a valid url: %w", i, err)
	}
	return href, nil
}

// linkRelr returns a link to the named group page.
func linkRelr(name string) (string, error) {
	href, err := url.JoinPath("/", "g", helper.Slug(name))
	if err != nil {
		return "", fmt.Errorf("name %q could not be made into a valid url: %w", name, err)
	}
	return href, nil
}

func names(s string) string {
	switch s {
	case "bbs", "ftp":
		return "file"
	}
	return s
}

// relHTML returns a HTML links for the primary and secondary group names.
func relHTML(prime, second string) template.HTML {
	var s string
	switch {
	case prime != "" && second != "":
		s = fmt.Sprintf("%s <strong>+</strong><br>%s", prime, second)
	case prime != "":
		s = prime
	case second != "":
		s = second
	default:
		return ""
	}
	return template.HTML(s)
}

// Form is the form data for the editor.
type Form struct {
	Target   string `query:"target"`   // Target is the name of the file to extract from the zip archive.
	Value    string `query:"value"`    // Value is the value of the form input field to change.
	Platform string `query:"platform"` // Platform is the platform of the release.
	Tag      string `query:"tag"`      // Tag is the tag of the release.
	ID       int    `query:"id"`       // ID is the auto incrementing database id of the record.
	Year     int16  `query:"year"`     // Year is the year of the release.
	Month    int16  `query:"month"`    // Month is the month of the release.
	Day      int16  `query:"day"`      // Day is the day of the release.
	Online   bool   `query:"online"`   // Online is the record online and public toggle.
	Readme   bool   `query:"readme"`   // Readme hides the readme textfile from the artifact page.
}

const (
	avif                    = ".avif"
	gif                     = ".gif"
	jpeg                    = ".jpeg"
	jpg                     = ".jpg"
	png                     = ".png"
	textamiga               = "textamiga"
	typeErr                 = "error: received an invalid type to "
	webp                    = ".webp"
	arrowLink template.HTML = `<svg class="bi" aria-hidden="true">` +
		`<use xlink:href="/bootstrap-icons.svg#arrow-right"></use>` +
		`</svg>`
)

// archives returns a list of archive file extensions supported by this web application.
func archives() []string {
	return []string{fzip, ".rar", ".7z", ".tar", ".lha", ".lzh", ".arc", ".arj", ".ace", ".tar"}
}

// documents returns a list of document file extensions that can be read as text in the browser.
func documents() []string {
	return []string{
		".txt", ".nfo", ".diz", ".asc", ".lit", ".rtf", ".doc", ".docx",
		".pdf", ".unp", ".htm", ".html", ".xml", ".json", ".csv",
	}
}

// images returns a list of image file extensions that can be displayed in the browser.
func images() []string {
	return []string{".avif", gif, jpg, jpeg, ".jfif", png, ".svg", webp, ".bmp", ".ico"}
}

// media returns a list of [media file extensions] that can be played in the browser.
//
// [media file extensions]: https://developer.mozilla.org/en-US/docs/Web/Media/Formats
func media() []string {
	return []string{".mpeg", ".mp1", ".mp2", ".mp3", ".mp4", ".ogg", ".wmv"}
}
