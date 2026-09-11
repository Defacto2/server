// Package app handles the routes and views for the Defacto2 website.
package app

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/internal/filerecord"
	"github.com/Defacto2/server/handler/internal/simple"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/null/v8"
	"github.com/bengarrett/bbs"
	"github.com/labstack/echo/v5"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	// Welcome is the default logo monospace text,
	// each side contains 20 whitespace characters.
	// The welcome to defacto2 text is 19 characters long.
	// The letter O of the word "TO" is the center of the text.
	Welcome = `:                    ` +
		`·· WELCOME TO DEFACTO2 ··` +
		`                    ·`
)

var (
	ErrArtifact               = errors.New("app: artifact response is nil")
	ErrClaims                 = errors.New("app: no sub id in the claims payload")
	ErrCorrupt                = errors.New("app: cache data is invalid or corrupt")
	ErrDownload               = errors.New("app: cannot stat the downloaded file")
	ErrMisMatch               = errors.New("app: token mismatch")
	ErrNegative               = errors.New("app: value cannot be a negative number")
	ErrSession                = errors.New("app: no sub id in session")
	ErrStatus                 = errors.New("app: http status code is not valid")
	ErrType                   = errors.New("app: wrong type of value")
	ErrUser                   = errors.New("app: unknown user")
	ErrValue                  = errors.New("app: value is empty")
	ErrMissingObfuscatedID    = errors.New("app: missing obfuscated ID")
	ErrInvalidObfuscatedID    = errors.New("app: invalid obfuscated ID")
	ErrFileNotFound           = errors.New("app: file not found")
	ErrInvalidFilenamePattern = errors.New("app: filename does not match numeric suffix pattern")
)

func errVal(name string) template.HTML {
	return template.HTML("error, " + ErrValue.Error() + ": " + name)
}

const (
	pathSeparator               = "/"
	attr                        = " attributions"
	br                          = "<br>"
	div1                        = "</div>"
	sect0                       = "<section>"
	sect1                       = "</section>"
	ul0                         = "<ul>"
	ul1                         = "</ul>"
	lessThan                    = "<"
	ltEntity                    = "&lt;"
	clearScreen                 = "@CLS@"
	typeErr                     = "error: received an invalid type to "
	arrowLink     template.HTML = `<svg class="bi" aria-hidden="true">` +
		`<use xlink:href="/svg/bootstrap-icons.svg#arrow-right"></use></svg>`
	arrowRight template.HTML = `<use xlink:href="/svg/bootstrap-icons.svg#arrow-right"/></svg>`
	svg        template.HTML = `<svg class="bi text-black" width="16" height="16" fill="currentColor" viewBox="0 0 16 16">`
)

const (
	bb        = "bbs"
	ca        = "California"
	cc        = "CC BY-SA 4.0"
	flt       = "Fairlight"
	gns       = "Genesis"
	pd        = "Public Domain"
	rz        = "RZR"
	rzr       = "Razor 1911"
	thg       = "The Humble Guys"
	alpha     = "alphabetically"
	byYear    = "by year"
	canonical = "canonical"
	demo      = "demo"
	files     = "files"
	magazine  = "magazine"
	ordrby    = "orderBy"
	pubs      = "pubs"
	records   = "records"
	search    = "search"
	text      = "text"
	textamiga = "textamiga"
	websites  = "websites"
	years     = "years"
)

// Caching are values that are used throughout the app or layouts.
var Caching = Cache{RecordCount: 0} //nolint:gochecknoglobals

// Records caches the database record count.
func (c *Cache) Records(i int64) {
	c.RecordCount = i
}

// Attribute returns a formatted string of the roles for the given scener name.
// For example if the name is "ben", write is "ben" and code is "bianca,ben" then
// the following would return:
//
//	"Writer and programmer attributions"
func Attribute(write, code, art, music, name string) string {
	if name == "" {
		return ""
	}

	const sep = `,`
	w, c, a, m := strings.Split(strings.ToLower(write), sep),
		strings.Split(strings.ToLower(code), sep),
		strings.Split(strings.ToLower(art), sep),
		strings.Split(strings.ToLower(music), sep)
	if len(w) == 0 && len(c) == 0 && len(a) == 0 && len(m) == 0 {
		return ""
	}

	name = strings.ToLower(name)
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
		return fmt.Sprintf("error: %q, %s", name, strings.Join(all, sep))
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

// Brief returns a human readable brief description of the combined platform and section.
// For example providing "windows" and "intro" would return:
//
//	"a Windows intro"
func Brief(platform, section any) string {
	p, ok := parseValS(platform)
	if !ok {
		return fmt.Sprintf("%s describe %T", typeErr, platform)
	}
	s, ok := parseValS(section)
	if !ok {
		return fmt.Sprintf("%s describe %T", typeErr, section)
	}
	if p == "" && s == "" {
		return "an unknown release"
	}

	return tags.Humanize(tags.TagByURI(p), tags.TagByURI(s)) + "."
}

// ByteBytes returns both the bytes and a human readable string of the bytes.
func ByteBytes(bytes any) template.HTML {
	n, ok := parseValI(bytes)
	if !ok {
		return template.HTML(fmt.Sprintf("%sByteBytes: %s",
			typeErr, reflect.TypeOf(bytes).String()))
	}

	return template.HTML(helper.ByteCountFloat(int64(n)) +
		` <small>(` + strconv.Itoa(n) + `B)</small>`)
}

// ByteFile returns a human readable string of the file count and bytes.
func ByteFile(count, bytes any) template.HTML {
	return ByteFileS("", count, bytes)
}

// ByteFileS returns a human readable string of the byte count with a named description.
func ByteFileS(name string, count, bytes any) template.HTML {
	n, ok := parseValI(count)
	if !ok {
		return template.HTML(fmt.Sprintf("%sByteFile cnt: %s",
			typeErr, reflect.TypeOf(count).String()))
	}

	p := message.NewPrinter(language.English)
	s := p.Sprint(strconv.Itoa(n))

	b, ok := parseValI64(bytes)
	if !ok {
		return template.HTML(fmt.Sprintf("%sByteFile bytes: %s",
			typeErr, reflect.TypeOf(bytes).String()))
	}

	name = names(name)
	const size = 2
	if n < size {
		return template.HTML(s + ` ` + name + ` <small>(` + helper.ByteCountFloat(b) + `)</small>`)
	}
	name += "s"

	return template.HTML(s + ` ` + name + ` <small>(` + helper.ByteCountFloat(b) + `)</small>`)
}

// Day returns a string representation of the day number, a value between 1 and 31.
func Day(d any) string {
	n, ok := parseValI(d)
	if !ok {
		return fmt.Sprintf("%sDay: %s", typeErr, reflect.TypeOf(d).String())
	}

	if n < 0 || n > 31 {
		return " error: day out of range " + strconv.Itoa(n)
	}

	return " " + strconv.Itoa(n)
}

// Describe returns a human readable description of a release.
// Based on the platform, section, year and month.
//
// For example providing "windows", "intro", 1990 and 1 would return:
//
//	"a Windows intro published in Jan, 1990."
func Describe(platform, section, year, month any) template.HTML {
	const tmpl = "describe"

	p, ok := parseValS(platform)
	if !ok {
		return template.HTML(fmt.Sprintf("%s %s %s", typeErr, tmpl, platform))
	}
	p = strings.TrimSpace(p)

	s, ok := parseValS(section)
	if !ok {
		return template.HTML(fmt.Sprintf("%s %s %s", typeErr, tmpl, section))
	}
	s = strings.TrimSpace(s)

	n, ok := parseValI(year)
	if !ok {
		return template.HTML(fmt.Sprintf("%s %s %s", typeErr, tmpl, year))
	}
	y := strconv.Itoa(n)

	n, ok = parseValI(month)
	if !ok {
		return template.HTML(fmt.Sprintf("%s %s %s", typeErr, tmpl, month))
	}
	m := helper.ShortMonth(n)

	return template.HTML(desc(p, s, y, m))
}

// GlobTo returns the path to the template file.
func GlobTo(name string) string {
	return strings.Join([]string{"view", "app", name}, pathSeparator)
}

// HasSuffix returns true if the string s ends with the suffix.
func HasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// LastUpdated returns a string of the time since the given time t.
// The time is formatted as "Last updated 1 hour ago".
// If the time is not valid, an empty string is returned.
func LastUpdated(t any) string {
	if t == nil {
		return ""
	}

	return simple.Updated(t, "Last updated")
}

// LinkDownload creates a URL to link to the file download of the record.
// The id needs to be a valid integer.
// If the security alert is not empty, then a strike-through warning is returned.
// For example providing 1 and an empty security alert would return:
//
//	<a class="card-link" href="/d/9b1c6">Download</a>
func LinkDownload(id any, securityAlert string) template.HTML {
	if id == nil {
		return ""
	}

	s, err := simple.LinkID(id, "d")
	if err != nil {
		return template.HTML(err.Error())
	}

	if securityAlert != "" {
		return template.HTML(`<s class="card-link text-warning-emphasis" data-bs-toggle="tooltip" ` +
			`data-bs-title="Use the link to access this file download">Download</s>`)
	}

	return template.HTML(`<a class="card-link" href="` + s + `" rel="nofollow">Download</a>`)
}

// LinkHref creates a URL path to link to the file page for the record.
// The id needs to be a valid integer.
func LinkHref(id any) (string, error) {
	if id == nil {
		return "", fmt.Errorf("id is nil, %w", ErrNegative)
	}

	return simple.LinkID(id, "f") //nolint:wrapcheck
}

// LinkInterview returns a SVG arrow icon to indicate an interview link hosted on an external website.
// If the href is not a valid URL then an empty string is returned.
func LinkInterview(href string) template.HTML {
	if href == "" {
		return errVal("href")
	}

	p, err := url.Parse(href)
	if err != nil || (p != nil && p.Scheme == "") {
		// if href is not a valid URL, then it is a relative path to the site.
		return template.HTML("")
	}
	return arrowLink
}

// LinkPage creates a URL anchor element to link to the file page for the record.
// The id needs to be a valid integer.
// The keyboard shortcut is "kboard" and is used to link to the file page with the keyboard focus.
// It can be left empty to not include the keyboard shortcut.
func LinkPage(id, kboard any) template.HTML {
	if id == nil {
		return ""
	}

	href, err := simple.LinkID(id, "f")
	if err != nil {
		return template.HTML(err.Error())
	}

	n, ok := parseValI64(kboard)
	if !ok {
		return template.HTML(`<a class="card-link" href="` + href + `" rel="nofollow">Artifact</a>`)
	}

	keypress := strconv.FormatInt(n, 10)
	return template.HTML(`<a data-bs-toggle="tooltip" data-bs-title="control + alt + ` + keypress + `" ` +
		`id="artifact-card-link-` + keypress + `" class="card-link" href="` + href + `" rel="nofollow">Artifact</a>`)
}

// LinkRunApp creates a URL anchor element to link to the artifact page to launch the js-dos emulator.
// The id needs to be a valid integer. For example providing 1 would return:
//
//	&nbsp; &nbsp; <a class="card-link" href="/f/9b1c6#runapp">Run app</a>
func LinkRunApp(id any) template.HTML {
	if id == nil {
		return ""
	}

	href, err := simple.LinkID(id, "f")
	if err != nil {
		return template.HTML(err.Error())
	}

	return template.HTML(`&nbsp; &nbsp; <a class="card-link" href="` + href + `#runapp" rel="nofollow">Run app</a>`)
}

// LinkPreview creates a URL to link to the file record in-tab to use as a preview.
// The preview link will only show with compatible file types based on the platform and filename extension.
// The id needs to be a valid integer, the name is the filename and the platform is the platform of the release.
// Any invalid values will return an empty string.
//
// For example providing 1, "readme.txt" and "text" would return:
//
//	&nbsp; <a class="card-link" href="/v/9b1c6">Preview</a>
//
// But providing 1, "file.zip" and "text" would return an empty string.
func LinkPreview(id any, name, platform string) template.HTML {
	if id == nil || name == "" {
		return template.HTML("")
	}

	href := filerecord.LinkPreviewHref(id, name, platform)
	if href == "" {
		return template.HTML("")
	}

	return template.HTML(`&nbsp; <a class="card-link" href="` + href + `">Preview</a>`)
}

// LinkRemote returns a HTML link with an embedded SVG icon to an external website.
// There are no checks for the href or name values other than they are not empty.
func LinkRemote(href, name string) template.HTML {
	if href == "" {
		return errVal("href")
	}
	if name == "" {
		return errVal("name")
	}

	return template.HTML(`<a class="dropdown-item icon-link icon-link-hover link-light" href="` + href + `">` +
		name + ` ` + string(arrowLink) + `</a>`)
}

// LinkRemoteTip returns a HTML link with an embedded SVG icon to an external website.
// If the href or name values are empty then an error message is returned.
// If the tooltip is empty then LinkRemote is returned.
func LinkRemoteTip(href, name, tooltip string) template.HTML {
	if href == "" {
		return errVal("href")
	}
	if name == "" {
		return errVal("name")
	}
	if tooltip == "" {
		return LinkRemote(href, name)
	}

	return template.HTML(`<a class="dropdown-item icon-link icon-link-hover link-light" ` +
		`data-bs-toggle="tooltip" data-bs-title="` + tooltip +
		`" href="` + href + `">` + name + ` ` + string(arrowLink) + `</a>`)
}

// LinkScnr returns a link to the named scener page.
// If the name is empty then an empty string is returned with no error.
// An example of providing "some scener" would return:
//
//	"/p/some-scener", nil
func LinkScnr(name string) (string, error) {
	if name == "" {
		return "", nil
	}

	href, err := url.JoinPath("/", "p", helper.Slug(name))
	if err != nil {
		return "", fmt.Errorf("name %q could not be made into a valid url: %w", name, err)
	}

	return href, nil
}

// LinkScnrs returns a list of links to the named scener pages.
// Multiple names can be provided as a comma separated string.
// If the name is empty then an empty string is returned with no error.
func LinkScnrs(names string) template.HTML {
	vals := strings.Split(names, ",")
	links := make([]string, 0, len(vals))
	for _, val := range vals {
		val = strings.TrimSpace(val)
		if val == "" {
			continue
		}
		scnr, err := LinkScnr(val)
		if err != nil {
			discard(err)
			continue
		}
		linkr := `<a class="link-dark link-offset-2 link-offset-3-hover link-underline ` +
			`link-underline-opacity-0 link-underline-opacity-75-hover" href="` + scnr + `">` + val + `</a>`
		links = append(links, linkr)
	}

	return template.HTML(strings.Join(links, ", "))
}

const wikiBase = "https://github.com/Defacto2/defacto2.net/wiki"

// LinkWiki returns a HTML link with an embedded SVG icon to the Defacto2 wiki on GitHub.
// The uri must be a valid URI path to a wiki page and the name must not be empty.
func LinkWiki(uri, name string) template.HTML {
	if uri == "" {
		return errVal("uri")
	}
	if name == "" {
		return errVal("name")
	}

	href, err := url.JoinPath(wikiBase, uri)
	if err != nil {
		return template.HTML(err.Error())
	}
	if strings.HasPrefix(uri, "#") {
		href = wikiBase + uri
	}

	return template.HTML(`<a class="dropdown-item icon-link icon-link-hover link-light" href="` + href + `">` +
		name + ` ` + string(arrowLink) + `</a>`)
}

// LinkWikiTip returns a HTML link with an embedded SVG icon to the Defacto2 wiki on GitHub.
// The uri must be a valid URI path to a wiki page and the name must not be empty.
// If the tooltip is empty then LinkWiki is returned.
func LinkWikiTip(uri, name, tooltip string) template.HTML {
	if uri == "" {
		return errVal("uri")
	}
	if name == "" {
		return errVal("name")
	}
	if tooltip == "" {
		return LinkWiki(uri, name)
	}

	href, err := url.JoinPath(wikiBase, uri)
	if err != nil {
		return template.HTML(err.Error())
	}

	if strings.HasPrefix(uri, "#") {
		href = wikiBase + uri
	}
	return template.HTML(`<a class="dropdown-item icon-link icon-link-hover link-light" ` +
		`data-bs-toggle="tooltip" data-bs-title="` + tooltip +
		`" href="` + href + `">` + name + ` ` + string(arrowLink) + `</a>`)
}

// LogoText returns a string of text padded with spaces to center it in the logo.
// If the string is empty then the default logo text is returned.
// The text is converted to uppercase and truncated if it is longer than the limit.
// An example of providing "abc" would return:
//
//	"      :                            ·· ABC ··                            ·"
func LogoText(s string) string {
	const spaces = 6
	indent := strings.Repeat(" ", spaces)
	if s == "" {
		return indent + Welcome
	}

	// odd returns true if the given number is odd.
	odd := func(n int) bool {
		return n%2 != 0
	}

	const padder = " ·· "
	const wl = len(Welcome)
	const pl = len(padder)
	const limit = wl - (pl + pl) - 3

	s = strings.ToUpper(s)
	truncate := len(s) > limit
	if truncate {
		return indent + ":" + padder + s[:limit] + padder + "·"
	}

	styled := padder + s + padder
	if !odd(len(s)) {
		styled = " " + styled
	}

	const split = 2
	count := (wl / split) - (len(styled) / split) - split
	pad := strings.Repeat(" ", count)
	return indent + ":" + pad + styled + pad + "·"
}

// MarkAll surrounds all occurrences of highlight in the string with <mark> elements.
func MarkAll(highlight, s string) string { // TODO: test against emoji
	if highlight == "" || s == "" {
		return s
	}

	substr := strings.ToLower(highlight)
	var builder strings.Builder
	const size = 13
	builder.Grow(len(s) + size)

	pos := 0
	for {
		// search the remaining text
		remainer := strings.ToLower(s[pos:])
		matchIdx := strings.Index(remainer, substr)

		if matchIdx == -1 {
			builder.WriteString(s[pos:])
			return builder.String()
		}

		start := pos + matchIdx
		end := start + len(substr)
		builder.WriteString(s[pos:start])
		builder.WriteString("<mark>")
		builder.WriteString(s[start:end])
		builder.WriteString("</mark>")

		// move cursor forward
		pos = end
	}
}

// Month returns a short string of the month.
// If the month number is not a valid then an empty string is returned.
// For example providing 1 would return:
//
//	"Jan"
func Month(m any) string {
	if m == nil {
		return ""
	}

	n, ok := parseValI(m)
	if !ok {
		return ""
	}

	if n < 1 || n > 12 {
		return ""
	}

	return " " + time.Month(n).String()
}

// MusicModule returns true if the magic string indicates a music file.
// Only tracker music is valid, MIDI, MP3, return false.
func MusicModule(magic any) bool {
	s, ok := parseValS(magic)
	if !ok {
		return false
	}

	patterns := [...]string{
		"Extended Module",
		"Multi-Track Module",
		"Impulse Tracker",
		"ProTracker",
		"Tracker music",
		"Module music",
		"MOD music",
		"S3M music",
		"IT music",
		"XM music",
	}

	s = strings.ToLower(s)
	for _, pattern := range patterns {
		if strings.Contains(s, strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// Prefix returns a string prefixed with a space.
func Prefix(s string) string {
	if s == "" {
		return ""
	}
	return " " + s
}

// RecordRels returns the groups associated with a release and joins them using a plus sign.
// For example providing "Group 1" and "Group 2" would return:
//
//	"Group 1 + Group 2"
func RecordRels(a, b any) string {
	x, _ := parseValS(a)
	x = strings.TrimSpace(x)
	y, _ := parseValS(b)
	y = strings.TrimSpace(y)

	switch {
	case x == "" && y == "":
		return ""
	case x != "" && y != "":
		return strings.Join([]string{x, y}, " + ")
	case x != "":
		return x
	case y != "":
		return y
	default:
		return ""
	}
}

// SafeBBS returns a string as a template.HTML type to prevent HTML escaping in the template.
// If PCBoard or Renegard color codes are discovered,
// these will be converted into italic elements containing custom color classes.
//
// Note: See [internal.filerecord.ForceSimpleText] for manual override options based on
// the SHA3 filehash value.
//
// If any value is not a valid string then an empty string is returned.
func SafeBBS(a any) template.HTML {
	val, ok := parseValS(a)
	if !ok {
		return ""
	}

	src := []byte(val)

	// Check for and strip RTF formatting first, before any other processing
	if simple.RTF(src) {
		// Strip RTF formatting if detected
		src = simple.StripRTF(src)
	}

	// remove any html elements false positives
	rene := bbs.IsRenegade(src)
	pcb := bbs.IsPCBoard(src)
	if !rene && !pcb {
		// return plain text, which also needs replacements
		src = bytes.ReplaceAll(src, []byte(lessThan), []byte(ltEntity))
		return SafeHTML(string(src))
	}

	// build stylized text
	src = bytes.ReplaceAll(src, []byte(clearScreen), []byte("\n"))

	var buf bytes.Buffer
	if pcb {
		if err := bbs.PCBoardHTML(&buf, src...); err != nil {
			return template.HTML(fmt.Sprintf("PCBoard conversion error: %v", err))
		}
	}

	if rene {
		if err := bbs.RenegadeHTML(&buf, src...); err != nil {
			return template.HTML(fmt.Sprintf("Renegade conversion error: %v", err))
		}
	}

	// return the stylized text
	return SafeHTML(buf.String())
}

// SafeDocument returns a string as a template.HTML type to prevent HTML escaping in the template.
// To avoid false positives, this does not handle PCBoard or Renegard color codes.
//
// If any value is not a valid string then an empty string is returned.
func SafeDocument(a any) template.HTML {
	val, ok := parseValS(a)
	if !ok {
		return ""
	}

	src := []byte(val)

	// Check for and strip RTF formatting first, before any other processing
	if simple.RTF(src) {
		src = simple.StripRTF(src)
	}

	// remove any html elements false positives
	src = bytes.ReplaceAll(src, []byte(lessThan), []byte(ltEntity))
	return SafeHTML(string(src))
}

var rePCBoard = regexp.MustCompile(bbs.PCBoardRe)

// RemovePCBoard removes any PCBoard sequences from the byte slice.
func RemovePCBoard(b []byte) []byte {
	return rePCBoard.ReplaceAll(b, []byte(""))
}

// SafeHTML returns a string as a template.HTML type to prevent HTML escaping in the template.
func SafeHTML(s string) template.HTML {
	return template.HTML(s)
}

// SafeJS returns a string as a template.JS type to prevent JavaScript escaping in the template.
func SafeJS(s string) template.JS {
	return template.JS(s)
}

// Safety returns true if SafeDocument should be used instead of SafeBBS.
func Safety(platform, section any) bool {
	p, ok := parseValS(platform)
	if !ok {
		return false
	}
	s, ok := parseValS(section)
	if !ok {
		return false
	}

	switch s {
	case "internaldocument", magazine:
		return true
	}
	return p == "placeholder"
}

// SubTitle returns a secondary element with the record title.
// If the section is "magazine" and the title is a number then it is prefixed with "Issue".
// For example providing "magazine" and 1 would return:
//
//	`<h3 class="card-subtitle mb-2 text-body-secondary fs-6">Issue 1</h3>`
//
// Otherwise providing "text" and "Some Cool Stuff" would return:
//
//	`<h3 class="card-subtitle mb-2 text-body-secondary fs-6">Some Cool Stuff</h3>`
func SubTitle(section null.String, title any, large bool) template.HTML {
	val, ok := parseValS(title)
	if !ok || val == "" {
		return ""
	}

	if strings.TrimSpace(strings.ToLower(section.String)) == magazine {
		if i, err := strconv.Atoi(val); err == nil {
			val = "Issue " + strconv.Itoa(i)
		}
	}

	fs := "fs-6"
	if large {
		fs = "fs-5"
	}
	return template.HTML(`<h3 class="card-subtitle mb-2 text-body-secondary ` + fs + `">` + val + `</h3>`)
}

// TagBrief returns a small summary of the tag.
// For example providing "interview" would return:
//
//	"Conversations with the personalities of The Scene"
func TagBrief(tag string) string {
	if tag == "" {
		return ""
	}
	return tags.Infos()[tags.TagByURI(tag)]
}

// TagOption returns a HTML option tag with a "selected" attribute if the s matches the value.
// For example providing "interview" and "interview" would return:
//
//	`<option value="interview" selected>`
func TagOption(s, value any) template.HTML {
	selected, ok := parseValS(s)
	if !ok {
		return ""
	}
	selected = strings.TrimSpace(selected)

	val, ok := parseValS(value)
	if !ok {
		return ""
	}
	val = strings.TrimSpace(val)

	if selected != "" && selected == val {
		return template.HTML(`<option value="` + val + `" selected>`)
	}

	return template.HTML(`<option value="` + val + `">`)
}

// TagWithOS returns a small summary of the tag with the operating system.
// If either the os or tags are unknown then a message is returned.
// For example providing "dos" and "magazine" would return:
//
//	"a Dos magazine"
func TagWithOS(os, tag string) string {
	return tags.Humanize(
		tags.TagByURI(os),
		tags.TagByURI(tag),
	)
}

// TrimSiteSuffix returns a string with the last 4 characters removed if they are " FTP" or " BBS".
// For example providing "My super FTP" would return:
//
//	"My super"
func TrimSiteSuffix(s string) string {
	n := strings.ToLower(strings.TrimSpace(s))

	const chrs = 4
	count := len(s)
	if count < chrs {
		return s
	}
	switch n[count-chrs:] {
	case " ftp", " bbs":
		return s[:count-chrs]
	}

	return s
}

// TrimSpace returns a string with all leading and trailing whitespace removed.
// If the value is a null.String then the value is checked for validity.
func TrimSpace(a any) string {
	if a == nil {
		return ""
	}

	s, ok := parseValS(a)
	if !ok {
		return fmt.Sprintf("%s trim site suffix: %s", typeErr, reflect.TypeOf(a).String())
	}
	return strings.TrimSpace(s)
}

// URLEncode returns a URL encoded string from the given string.
// This can be used to pass filenames as URL parameters.
func URLEncode(a any) string {
	if a == nil {
		return ""
	}

	s, ok := parseValS(a)
	if !ok {
		return fmt.Sprintf("%s url encode: %s", typeErr, reflect.TypeOf(a).String())
	}
	return url.QueryEscape(s)
}

// WebsiteIcon returns a Bootstrap icon name for the given website url.
// For example if the url contains "archive.org" then the Bootstrap icon "bank2" svg icon is returned.
func WebsiteIcon(url string) template.HTML {
	if url == "" {
		return ""
	}

	icon := websiteIcon(url)

	if icon == "arrow-right" {
		return svg + arrowRight
	}

	return svg + template.HTML(`<use xlink:href="/svg/bootstrap-icons.svg#`+icon+`"/></svg>`)
}

func websiteIcon(url string) string {
	switch {
	case
		strings.Contains(url, "archive.org"),
		strings.Contains(url, "wayback.defacto2.net"):
		return "bank2"
	case strings.Contains(url, "reddit.com"):
		return "reddit"
	case
		strings.Contains(url, "yalebooks.yale.edu"),
		strings.Contains(url, "explodingthephone.com"),
		strings.Contains(url, "punctumbooks"):
		return "book"
	case strings.Contains(url, "youtube.com"):
		return "youtube"
	case strings.Contains(url, "vimeo.com"):
		return "vimeo"
	case strings.Contains(url, "slashdot.org"):
		return "slash-circle"
	case strings.Contains(url, "twitter"):
		return "twitter"
	case strings.Contains(url, "wikipedia.org"):
		return "wikipedia"
	case strings.Contains(url, "textfiles.com"):
		return "textfiles"
	default:
		return "arrow-right"
	}
}

var reSupElems = regexp.MustCompile(`<sup>.*?</sup>`)

// StripSup removes <sup>...</sup> tags from a string and returns the cleaned string.
// The sup tags are returned separately if present, otherwise empty string.
//
// Usage in templates: {{ $result := stripSup .Title }} returns a map with "text" and "sup" keys.
func StripSup(s string) (map[string]template.HTML, error) { // TODO: remove error
	clean := strings.TrimSpace(reSupElems.ReplaceAllString(s, ""))
	return map[string]template.HTML{
		text:  template.HTML(clean),
		"sup": template.HTML(reSupElems.FindString(s)),
	}, nil
}

// YMDEdit handles the post submission for the Year, Month, Day selection fields.
func YMDEdit(c *echo.Context, tx *sql.Tx) error {
	const format = "year month day edit %s: %w"
	if err := nils.Check(c, tx); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}

	ctx := c.Request().Context()
	key := f.ID
	r, err := model.One(ctx, tx, true, key)
	if err != nil {
		return fmt.Errorf(format, "model one", err)
	}

	y := model.ValidY(f.Year)
	m := model.ValidM(f.Month)
	d := model.ValidD(f.Day)
	if err = model.UpdateYMD(ctx, tx, int64(key), y, m, d); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(format, "tx commit", err)
	}

	return c.JSON(http.StatusOK, r)
}

// Cache contains database values that are used throughout the app or layouts,
// but do not change frequently enough to warrant a database query on every page load.
type Cache struct {
	RecordCount int64 // The total number of file records in the database.
}

// SRI are the Sub-Resource Integrity hashes for the layout.
type SRI struct {
	Bootstrap5      string // Bootstrap CSS verification hash.
	Bootstrap5JS    string // Bootstrap JS verification hash.
	BootstrapIcons  string // Bootstrap Icons SVG verification hash.
	CanvasAnsi      string // ANSI JS verification hash.
	CanvasReadme    string // Readme JS verification hash.
	ChiptunePlayer  string // Chiptune Player JS verification hash.
	EditArtifact    string // Artifact Editor JS verification hash.
	EditAssets      string // Editor Assets JS verification hash.
	EditForApproval string // Editor For Approval JS verification hash.
	IndexJS         string
	Jsdos6JS        string // js-dos v6 verification hash.
	DosboxJS        string // DOSBox Emscripten verification hash.
	Layout          string // Layout CSS verification hash.
	LayoutJS        string // Layout JS verification hash.
	Pouet           string // Pouet JS verification hash.
	Uploader        string // Uploader JS verification hash.
	Htmx            string // htmx JS verification hash.
	HtmxRespTargets string // htmx response targets extension JS verification hash.
}

// Verify checks the integrity of the embedded CSS and JS files.
// These are required for Subresource Integrity (SRI) verification in modern browsers.
// The fs is the embedded file system that contains the public facing file assets.
func (s *SRI) Verify(fsys fs.FS) error { //nolint:funlen
	if err := nils.Check(fsys); err != nil {
		return fmt.Errorf("sri verify: %w", err)
	}

	names := Names()

	var err error
	const format = "%s: %w"

	name := names[Bootstrap5]
	s.Bootstrap5, err = helper.Integrity(name, fsys)
	if err != nil {
		return fmt.Errorf(format, name, err)
	}

	name = names[Bootstrap5JS]
	s.Bootstrap5JS, err = helper.Integrity(name, fsys)
	if err != nil {
		return fmt.Errorf(format, name, err)
	}

	name = names[BootstrapIcons]
	s.BootstrapIcons, err = helper.Integrity(name, fsys)
	if err != nil {
		return fmt.Errorf(format, name, err)
	}

	name = names[ContentBinary]
	s.CanvasAnsi, err = helper.Integrity(name, fsys)
	if err != nil {
		return fmt.Errorf(format, name, err)
	}

	name = names[ContentText]
	s.CanvasReadme, err = helper.Integrity(name, fsys)
	if err != nil {
		return fmt.Errorf(format, name, err)
	}

	name = names[LayoutJS]
	s.LayoutJS, err = helper.Integrity(name, fsys)
	if err != nil {
		return fmt.Errorf(format, name, err)
	}

	name = names[ChiptunePlayer]
	s.ChiptunePlayer, err = helper.Integrity(name, fsys)
	if err != nil {
		return fmt.Errorf(format, name, err)
	}

	name = names[EditArtifact]
	s.EditArtifact, err = helper.Integrity(name, fsys)
	if err != nil {
		return fmt.Errorf(format, name, err)
	}

	name = names[EditAssets]
	s.EditAssets, err = helper.Integrity(name, fsys)
	if err != nil {
		return fmt.Errorf(format, name, err)
	}

	name = names[EditForApproval]
	s.EditForApproval, err = helper.Integrity(name, fsys)
	if err != nil {
		return fmt.Errorf(format, name, err)
	}

	name = names[IndexJS]
	s.IndexJS, err = helper.Integrity(name, fsys)
	if err != nil {
		return fmt.Errorf(format, name, err)
	}

	name = names[Jsdos6JS]
	s.Jsdos6JS, err = helper.Integrity(name, fsys)
	if err != nil {
		return fmt.Errorf(format, name, err)
	}

	name = names[DosboxJS]
	s.DosboxJS, err = helper.Integrity(name, fsys)
	if err != nil {
		return fmt.Errorf(format, name, err)
	}

	name = names[Layout]
	s.Layout, err = helper.Integrity(name, fsys)
	if err != nil {
		return fmt.Errorf(format, name, err)
	}

	name = names[Pouet]
	s.Pouet, err = helper.Integrity(name, fsys)
	if err != nil {
		return fmt.Errorf(format, name, err)
	}

	name = names[Uploader]
	s.Uploader, err = helper.Integrity(name, fsys)
	if err != nil {
		return fmt.Errorf(format, name, err)
	}

	name = names[Htmx]
	s.Htmx, err = helper.Integrity(name, fsys)
	if err != nil {
		return fmt.Errorf(format, name, err)
	}

	name = names[HtmxRespTargets]
	s.HtmxRespTargets, err = helper.Integrity(name, fsys)
	if err != nil {
		return fmt.Errorf(format, name, err)
	}

	return nil
}

// badRequest returns a JSON response with a 400 status code,
// the server cannot or will not process the request due to something that is perceived to be a client error.
func badRequest(c *echo.Context, err error) error {
	if err := nils.Check(c); err != nil {
		return fmt.Errorf("app bad request: %w", err)
	}

	const code = http.StatusBadRequest
	s := ""
	if err != nil {
		s = " " + err.Error()
	}
	return c.JSON(code, map[string]string{"error": "bad request" + s})
}

func desc(platform, section, year, month string) string {
	if platform == "" && section == "" {
		return "An unknown release."
	}

	x := tags.Humanize(tags.TagByURI(platform), tags.TagByURI(section))
	x = helper.Capitalize(x)

	if month != "" && year != "" {
		return x + ` published in <span class="text-nowrap">` + month + `, ` + year + `</span>.`
	}
	if year != "" {
		return x + ` published in ` + year + "."
	}
	return x + "."
}

func names(s string) string {
	switch s {
	case bb, "ftp":
		return "file"
	}
	return s
}

// Form is the form data for the editor.
type Form struct {
	Target   string `query:"target"`   // Target is the name of the file to extract from the zip archive.
	Value    string `query:"value"`    // Value of the form input field to change.
	Platform string `query:"platform"` // Platform of the release.
	Tag      string `query:"tag"`      // Tag of the release.
	ID       int    `query:"id"`       // ID is the auto incrementing database id of the record.
	Year     int16  `query:"year"`     // Year of the release.
	Month    int16  `query:"month"`    // Month of the release.
	Day      int16  `query:"day"`      // Day of the release.
	Online   bool   `query:"online"`   // Online is the record online and public toggle.
	Readme   bool   `query:"readme"`   // Readme hides the readme text file from the artifact page.
}
