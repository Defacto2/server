// Package app handles the routes and views for the Defacto2 website.
package app

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/server/handler/app/internal/mf"
	"github.com/Defacto2/server/handler/app/internal/str"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
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
	return str.Updated(t, s)
}

// LinkDownload creates a URL to link to the file download of the record.
func LinkDownload(id any, uri string) template.HTML {
	if id == nil {
		return ""
	}
	s, err := str.LinkID(id, "d")
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
		return "", fmt.Errorf("id is nil, %w", ErrNegative)
	}
	return str.LinkID(id, "f") //nolint:wrapcheck
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
	s, err := str.LinkID(id, "f")
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
	s := mf.LinkPreviewHref(id, name, platform)
	if s == "" {
		return template.HTML("")
	}
	elm := fmt.Sprintf(`&nbsp; <a class="card-link" href="%s">Preview</a>`, s)
	return template.HTML(elm)
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

// websiteIcon returns a Bootstrap icon name for the given website url.
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
func YMDEdit(c echo.Context) error {
	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	ctx := context.Background()
	db, tx, err := postgres.ConnectTx()
	if err != nil {
		return fmt.Errorf("ymdedit connect %w", err)
	}
	defer db.Close()
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

// Cache contains database values that are used throughout the app or layouts.
type Cache struct {
	RecordCount int // The total number of file records in the database.
}

// SRI are the Subresource Integrity hashes for the layout.
type SRI struct {
	ArtifactEditor  string // Artifact Editor JS verification hash.
	Bootstrap5      string // Bootstrap CSS verification hash.
	Bootstrap5JS    string // Bootstrap JS verification hash.
	BootstrapIcons  string // Bootstrap Icons SVG verification hash.
	EditAssets      string // Editor Assets JS verification hash.
	EditArchive     string // Editor Archive JS verification hash.
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
func (s *SRI) Verify(fs embed.FS) error { //nolint:funlen
	names := Names()
	var err error
	name := names[ArtifactEditor]
	s.ArtifactEditor, err = helper.Integrity(name, fs)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	name = names[Bootstrap5]
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
	name = names[EditAssets]
	s.EditAssets, err = helper.Integrity(name, fs)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	name = names[EditArchive]
	s.EditArchive, err = helper.Integrity(name, fs)
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
		`<use xlink:href="/svg/bootstrap-icons.svg#arrow-right"></use></svg>`
)
