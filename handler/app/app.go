// Package app handles the routes and views for the Defacto2 website.
package app

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/app/internal/filerecord"
	"github.com/Defacto2/server/handler/app/internal/simple"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/null/v8"
	"github.com/bengarrett/bbs"
	"github.com/labstack/echo/v4"
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
	ErrClaims   = errors.New("no sub id in the claims playload")
	ErrCode     = errors.New("the http status code is not valid")
	ErrCxt      = errors.New("the server could not create a context")
	ErrData     = errors.New("cache data is invalid or corrupt")
	ErrDB       = errors.New("database connection is nil")
	ErrExtract  = errors.New("unknown extractor value")
	ErrLinkType = errors.New("the id value is an invalid type")
	ErrMisMatch = errors.New("token mismatch")
	ErrNegative = errors.New("value cannot be a negative number")
	ErrSession  = errors.New("no sub id in session")
	ErrTarget   = errors.New("target not found")
	ErrTmpl     = errors.New("the server could not render the html template for this page")
	ErrType     = errors.New("value is the wrong type")
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
// For example if the name is "ben", write is "ben" and code is "bianca,ben" then
// the following would return:
//
//	"Writer and programmer attributions"
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

// Brief returns a human readable brief description of the combined platform and section.
// For example providing "windows" and "intro" would return:
//
//	"a Windows intro"
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

// Day returns a string representation of the day number, a value between 1 and 31.
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
//
// For example providing "windows", "intro", 1990 and 1 would return:
//
//	"a Windows intro published in Jan, 1990."
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
	const s = "Last updated"
	return simple.Updated(t, s)
}

// LinkDownload creates a URL to link to the file download of the record.
// The id needs to be a valid integer.
// If the security alert is not empty, then a strikethrough warning is returned.
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
	return template.HTML(fmt.Sprintf(`<a class="card-link" href="%s">Download</a>`, s))
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
	if err != nil || p.Scheme == "" {
		// if href is not a valid URL, then it is a relative path to the site.
		return template.HTML("")
	}
	return arrowLink
}

// LinkPage creates a URL anchor element to link to the file page for the record.
// The id needs to be a valid integer. For example providing 1 would return:
//
//	<a class="card-link" href="/f/9b1c6">Artifact</a>
//
// The keyboard shortcut is "kboard" and is used to link to the file page with the keyboard focus.
// It can be left empty to not include the keyboard shortcut.
func LinkPage(id, kboard any) template.HTML {
	if id == nil {
		return ""
	}
	s, err := simple.LinkID(id, "f")
	if err != nil {
		return template.HTML(err.Error())
	}
	kb, valid := kboard.(int64)
	if !valid {
		return template.HTML(fmt.Sprintf(`<a class="card-link" href="%s">Artifact</a>`, s))
	}
	keypress := strconv.FormatInt(kb, 10)
	return template.HTML(fmt.Sprintf(`<a data-bs-toggle="tooltip" data-bs-title="control + alt + %s" `+
		`id="artifact-card-link-%s" class="card-link" href="%s">Artifact</a>`,
		keypress, keypress, s))
}

// LinkRunApp creates a URL anchor element to link to the artifact page to launch the js-dos emulator.
// The id needs to be a valid integer. For example providing 1 would return:
//
//	&nbsp; &nbsp; <a class="card-link" href="/f/9b1c6#runapp">Run app</a>
func LinkRunApp(id any) template.HTML {
	if id == nil {
		return ""
	}
	s, err := simple.LinkID(id, "f")
	if err != nil {
		return template.HTML(err.Error())
	}
	return template.HTML(fmt.Sprintf(`&nbsp; &nbsp; <a class="card-link" href="%s#runapp">Run app</a>`, s))
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
	s := filerecord.LinkPreviewHref(id, name, platform)
	if s == "" {
		return template.HTML("")
	}
	elm := fmt.Sprintf(`&nbsp; <a class="card-link" href="%s">Preview</a>`, s)
	return template.HTML(elm)
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
	a := fmt.Sprintf(`<a class="dropdown-item icon-link icon-link-hover link-light" href="%s">%s %s</a>`,
		href, name, arrowLink)
	return template.HTML(a)
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
	a := fmt.Sprintf(`<a class="dropdown-item icon-link icon-link-hover link-light" `+
		`data-bs-toggle="tooltip" data-bs-title="%s" href="%s">%s %s</a>`,
		tooltip, href, name, arrowLink)
	return template.HTML(a)
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
	links := []string{}
	vals := strings.Split(names, ",")
	cls := "link-dark link-offset-2 link-offset-3-hover link-underline " +
		"link-underline-opacity-0 link-underline-opacity-75-hover"
	for val := range slices.Values(vals) {
		val = strings.TrimSpace(val)
		if val == "" {
			continue
		}
		scnr, err := LinkScnr(val)
		if err != nil {
			_, _ = fmt.Fprint(io.Discard, err)
			continue
		}
		linkr := fmt.Sprintf(`<a class="%s" href="%s">%s</a>`, cls, scnr, val)
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
	var href string
	href, err := url.JoinPath(wikiBase, uri)
	if err != nil {
		return template.HTML(err.Error())
	}
	if strings.HasPrefix(uri, "#") {
		href = fmt.Sprintf("%s%s", wikiBase, uri)
	}
	a := fmt.Sprintf(`<a class="dropdown-item icon-link icon-link-hover link-light" href="%s">%s %s</a>`,
		href, name, arrowLink)
	return template.HTML(a)
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
	var href string
	href, err := url.JoinPath(wikiBase, uri)
	if err != nil {
		return template.HTML(err.Error())
	}
	if strings.HasPrefix(uri, "#") {
		href = fmt.Sprintf("%s%s", wikiBase, uri)
	}
	a := fmt.Sprintf(`<a class="dropdown-item icon-link icon-link-hover link-light" `+
		`data-bs-toggle="tooltip" data-bs-title="%s" href="%s">%s %s</a>`,
		tooltip, href, name, arrowLink)
	return template.HTML(a)
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

// Month returns a short string of the month.
// If the month number is not a valid then an empty string is returned.
// For example providing 1 would return:
//
//	"Jan"
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

// Prefix returns a string prefixed with a space.
func Prefix(s string) string {
	if s == "" {
		return ""
	}
	return s + " "
}

// RecordRels returns the groups associated with a release and joins them using a plus sign.
// For example providing "Group 1" and "Group 2" would return:
//
//	"Group 1 + Group 2"
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

// SafeBBS returns a string as a template.HTML type to prevent HTML escaping in the template,
// but also removes any PCBoard sequences from the string.
//
// If any value is not a valid string then an empty string is returned.
func SafeBBS(a any) template.HTML {
	const lessThan = "<"
	const ltEntity = "&lt;"
	switch val := a.(type) {
	case string:
		b := []byte(val)
		b = bytes.ReplaceAll(b, []byte(lessThan), []byte(ltEntity))
		if !bbs.IsPCBoard(b) {
			return SafeHTML(string(b))
		}
		s := string(RemovePCBoard(b))
		clear(b)
		return SafeHTML(s)
	default:
		return template.HTML("")
	}
}

// RemovePCBoard removes any PCBoard sequences from the byte slice.
func RemovePCBoard(b []byte) []byte {
	re := regexp.MustCompile(bbs.PCBoardRe)
	return re.ReplaceAll(b, []byte(""))
}

// SafeHTML returns a string as a template.HTML type to prevent HTML escaping in the template.
func SafeHTML(s string) template.HTML {
	return template.HTML(s)
}

// SafeJS returns a string as a template.JS type to prevent JavaScript escaping in the template.
func SafeJS(s string) template.JS {
	return template.JS(s)
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
	val := ""
	switch v := title.(type) {
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
	fs := "fs-6"
	if large {
		fs = "fs-5"
	}
	cls := "card-subtitle mb-2 text-body-secondary " + fs
	elem := fmt.Sprintf("<h3 class=\"%s\">%s</h3>", cls, val)
	return template.HTML(elem)
}

// TagBrief returns a small summary of the tag.
// For example providing "interview" would return:
//
//	"Conversations with the personalities of The Scene"
func TagBrief(tag string) string {
	t := tags.TagByURI(tag)
	s := tags.Infos()[t]
	return s
}

// TagOption returns a HTML option tag with a "selected" attribute if the s matches the value.
// For example providing "interview" and "interview" would return:
//
//	`<option value="interview" selected>`
func TagOption(s, value any) template.HTML {
	sel, val := "", ""
	switch i := s.(type) {
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
// If either the os or tags are unknown then a message is returned.
// For example providing "dos" and "magazine" would return:
//
//	"a Dos magazine"
func TagWithOS(os, tag string) string {
	p, t := tags.TagByURI(os), tags.TagByURI(tag)
	s := tags.Humanize(p, t)
	return s
}

// TrimSiteSuffix returns a string with the last 4 characters removed if they are " FTP" or " BBS".
// For example providing "My super FTP" would return:
//
//	"My super"
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
// If the value is a null.String then the value is checked for validity.
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
		return fmt.Sprintf("%s trim site suffix: %s", typeErr, reflect.TypeOf(a).String())
	}
}

// URLEncode returns a URL encoded string from the given string.
// This can be used to pass filenames as URL parameters.
func URLEncode(a any) string {
	if a == nil {
		return ""
	}
	switch val := a.(type) {
	case string:
		return url.QueryEscape(val)
	case null.String:
		if val.Valid {
			return url.QueryEscape(val.String)
		}
		return ""
	default:
		return fmt.Sprintf("%s url encode: %s", typeErr, reflect.TypeOf(a).String())
	}
}

// WebsiteIcon returns a Bootstrap icon name for the given website url.
// For example if the url contains "archive.org" then the Bootstrap icon "bank2" svg icon is returned.
func WebsiteIcon(url string) template.HTML {
	icon := websiteIcon(url)
	const svg = `<svg class="bi text-black" width="16" height="16" fill="currentColor" viewBox="0 0 16 16">`
	if icon == "arrow-right" {
		html := svg + `<use xlink:href="/svg/bootstrap-icons.svg#arrow-right"/></svg>`
		return template.HTML(html)
	}
	html := svg + fmt.Sprintf(`<use xlink:href="/svg/bootstrap-icons.svg#%s"/></svg>`, icon)
	return template.HTML(html)
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
func YMDEdit(c echo.Context, db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("ymdedit: %w", ErrDB)
	}
	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ymdedit begin tx %w", err)
	}
	r, err := model.One(ctx, tx, true, f.ID)
	if err != nil {
		return fmt.Errorf("ymdedit model one %w", err)
	}
	y := model.ValidY(f.Year)
	m := model.ValidM(f.Month)
	d := model.ValidD(f.Day)
	if err = model.UpdateYMD(ctx, tx, int64(f.ID), y, m, d); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("ymdedit commit %w", err)
	}
	return c.JSON(http.StatusOK, r)
}

// Cache contains database values that are used throughout the app or layouts,
// but do not change frequently enough to warrant a database query on every page load.
type Cache struct {
	RecordCount int // The total number of file records in the database.
}

// SRI are the Subresource Integrity hashes for the layout.
type SRI struct {
	Bootstrap5      string // Bootstrap CSS verification hash.
	Bootstrap5JS    string // Bootstrap JS verification hash.
	BootstrapIcons  string // Bootstrap Icons SVG verification hash.
	EditArtifact    string // Artifact Editor JS verification hash.
	EditAssets      string // Editor Assets JS verification hash.
	EditForApproval string // Editor For Approval JS verification hash.
	Jsdos6JS        string // js-dos v6 verification hash.
	DosboxJS        string // DOSBox Emscripten verification hash.
	Layout          string // Layout CSS verification hash.
	LayoutJS        string // Layout JS verification hash.
	Pouet           string // Pouet JS verification hash.
	Readme          string // Readme JS verification hash.
	Uploader        string // Uploader JS verification hash.
	Htmx            string // htmx JS verification hash.
	HtmxRespTargets string // htmx response targets extension JS verification hash.
}

// Verify checks the integrity of the embedded CSS and JS files.
// These are required for Subresource Integrity (SRI) verification in modern browsers.
// The fs is the embedded file system that contains the public facing file assets.
func (s *SRI) Verify(fs embed.FS) error { //nolint:funlen
	names := *Names()
	var err error
	name := names[Bootstrap5]
	s.Bootstrap5, err = helper.Integrity(name, fs)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	name = names[Bootstrap5JS]
	s.Bootstrap5JS, err = helper.Integrity(name, fs)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	name = names[BootstrapIcons]
	s.BootstrapIcons, err = helper.Integrity(name, fs)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	name = names[LayoutJS]
	s.LayoutJS, err = helper.Integrity(name, fs)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	name = names[EditArtifact]
	s.EditArtifact, err = helper.Integrity(name, fs)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	name = names[EditAssets]
	s.EditAssets, err = helper.Integrity(name, fs)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	name = names[EditForApproval]
	s.EditForApproval, err = helper.Integrity(name, fs)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	name = names[Jsdos6JS]
	s.Jsdos6JS, err = helper.Integrity(name, fs)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	name = names[DosboxJS]
	s.DosboxJS, err = helper.Integrity(name, fs)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	name = names[Layout]
	s.Layout, err = helper.Integrity(name, fs)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	name = names[Pouet]
	s.Pouet, err = helper.Integrity(name, fs)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	name = names[Readme]
	s.Readme, err = helper.Integrity(name, fs)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	name = names[Uploader]
	s.Uploader, err = helper.Integrity(name, fs)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	name = names[Htmx]
	s.Htmx, err = helper.Integrity(name, fs)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	name = names[HtmxRespTargets]
	s.HtmxRespTargets, err = helper.Integrity(name, fs)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

// badRequest returns a JSON response with a 400 status code,
// the server cannot or will not process the request due to something that is perceived to be a client error.
func badRequest(c echo.Context, err error) error {
	return c.JSON(http.StatusBadRequest,
		map[string]string{"error": "bad request " + err.Error()})
}

func desc(p, s, y, m string) string {
	if p == "" && s == "" {
		return "An unknown release."
	}
	x := tags.Humanize(tags.TagByURI(p), tags.TagByURI(s))
	x = helper.Capitalize(x)
	if m != "" && y != "" {
		x = fmt.Sprintf("%s published in <span class=\"text-nowrap\">%s, %s</span>", x, m, y)
	} else if y != "" {
		x = fmt.Sprintf("%s published in %s", x, y)
	}
	return x + "."
}

func names(s string) string {
	switch s {
	case "bbs", "ftp":
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
		`<use xlink:href="/svg/bootstrap-icons.svg#arrow-right"></use></svg>`
)
