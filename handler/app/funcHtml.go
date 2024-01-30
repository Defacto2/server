package app

import (
	"fmt"
	"html/template"
	"net/url"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/tags"
	"github.com/volatiletech/null/v8"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	link template.HTML = `<svg class="bi" aria-hidden="true">` +
		`<use xlink:href="/bootstrap-icons.svg#link"></use></svg>`
	wiki template.HTML = `<svg class="bi" aria-hidden="true">` +
		`<use xlink:href="/bootstrap-icons.svg#arrow-right-short"></use></svg>`
	typeErr   = "error: received an invalid type to "
	textamiga = "textamiga"
)

const (
	gif  = ".gif"
	jpeg = ".jpeg"
	jpg  = ".jpg"
	png  = ".png"
	webp = ".webp"
)

var (
	ErrLinkID   = fmt.Errorf("the id value cannot be a negative number")
	ErrLinkType = fmt.Errorf("the id value is an invalid type")
)

// archives returns a list of archive file extensions supported by this web application.
func archives() []string {
	return []string{".zip", ".rar", ".7z", ".tar", ".lha", ".lzh", ".arc", ".arj", ".ace", ".tar"}
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

// SafeHTML returns a string as a template.HTML type.
// This is intended to be used to prevent HTML escaping.
func SafeHTML(s string) template.HTML {
	return template.HTML(s)
}

// Screenshot returns a picture elment with screenshots for the given uuid.
func (web Web) Screenshot(uuid, desc string) template.HTML {
	fw := filepath.Join(web.Import.PreviewDir, fmt.Sprintf("%s%s", uuid, webp))
	fp := filepath.Join(web.Import.PreviewDir, fmt.Sprintf("%s%s", uuid, png))
	fj := filepath.Join(web.Import.PreviewDir, fmt.Sprintf("%s%s", uuid, jpg))
	webp := strings.Join([]string{config.StaticOriginal(), fmt.Sprintf("%s%s", uuid, webp)}, "/")
	png := strings.Join([]string{config.StaticOriginal(), fmt.Sprintf("%s%s", uuid, png)}, "/")
	jpg := strings.Join([]string{config.StaticOriginal(), fmt.Sprintf("%s%s", uuid, jpg)}, "/")
	alt := strings.ToLower(desc) + " screenshot"
	w, p, j := false, false, false
	if helper.IsStat(fw) {
		w = true
	}
	if helper.IsStat(fp) {
		p = true
	}
	if !p {
		// fallback to jpg on the odd chance that the png is missing
		if helper.IsStat(fj) {
			j = true
		}
	}
	class := "rounded mx-auto d-block img-fluid"
	if w && p {
		elm := "<picture>" +
			fmt.Sprintf("<source srcset=\"%s\" type=\"image/webp\" />", webp) +
			string(img(png, alt, class, "")) +
			"</picture>"
		return template.HTML(elm)
	}
	elm := ""
	if w {
		return img(webp, alt, class, "")
	}
	if p {
		return img(png, alt, class, "")
	}
	if j {
		return img(jpg, alt, class, "")
	}
	return template.HTML(elm)
}

// Thumb returns a HTML image tag or picture element for the given uuid.
// The uuid is the filename of the thumbnail image without an extension.
// The desc is the description of the image.
func (web Web) Thumb(uuid, desc string, bottom bool) template.HTML {
	fw := filepath.Join(web.Import.ThumbnailDir, fmt.Sprintf("%s.webp", uuid))
	fp := filepath.Join(web.Import.ThumbnailDir, fmt.Sprintf("%s.png", uuid))
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
	const style = "min-height:5em;max-height:20em;"
	class := "card-img-bottom"
	if !bottom {
		class = "card-img-top"
	}
	if !w && !p {
		s := "<img src=\"\" loading=\"lazy\" alt=\"thumbnail placeholder\"" +
			" class=\"" + class + " placeholder\" style=\"" + style + "\" />"
		return template.HTML(s)
	}
	if w && p {
		elm := "<picture class=\"" + class + "\">" +
			fmt.Sprintf("<source srcset=\"%s\" type=\"image/webp\" />", webp) +
			string(img(png, alt, class, style)) +
			"</picture>"
		return template.HTML(elm)
	}
	elm := ""
	if w {
		return img(webp, alt, class, style)
	}
	if p {
		return img(png, alt, class, style)
	}
	return template.HTML(elm)
}

// ImageSample returns a HTML image tag for the given uuid.
func (web Web) ImageSample(uuid string) template.HTML {
	const (
		png  = png
		webp = webp
	)
	ext, name, src := "", "", ""
	for _, ext = range []string{webp, png} {
		name = filepath.Join(web.Import.PreviewDir, fmt.Sprintf("%s%s", uuid, ext))
		src = strings.Join([]string{config.StaticOriginal(), fmt.Sprintf("%s%s", uuid, ext)}, "/")
		if helper.IsStat(name) {
			break
		}
	}
	hash, err := helper.IntegrityFile(name)
	if err != nil {
		return template.HTML(err.Error())
	}
	return template.HTML(fmt.Sprintf("<img src=\"%s?%s\" loading=\"lazy\" "+
		"class=\"img-fluid\" alt=\"%s sample\" integrity=\"%s\" />",
		src, hash, ext, hash))
}

// ThumbSample returns a HTML image tag for the given uuid.
func (web Web) ThumbSample(uuid string) template.HTML {
	const (
		png  = png
		webp = webp
	)
	ext, name, src := "", "", ""
	for _, ext = range []string{webp, png} {
		name = filepath.Join(web.Import.ThumbnailDir, fmt.Sprintf("%s%s", uuid, ext))
		src = strings.Join([]string{config.StaticThumb(), fmt.Sprintf("%s%s", uuid, ext)}, "/")
		if helper.IsStat(name) {
			break
		}
	}
	hash, err := helper.IntegrityFile(name)
	if err != nil {
		return template.HTML(err.Error())
	}
	return template.HTML(fmt.Sprintf("<img src=\"%s?%s\" loading=\"lazy\" "+
		"class=\"img-fluid\" alt=\"%s sample\" integrity=\"%s\" />",
		src, hash, ext, hash))
}

// img returns a HTML image tag.
func img(src, alt, class, style string) template.HTML {
	return template.HTML(fmt.Sprintf("<img src=\"%s\" loading=\"lazy\" alt=\"%s\" class=\"%s\" style=\"%s\" />",
		src, alt, class, style))
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

func names(s string) string {
	switch s {
	case "bbs", "ftp":
		return "file"
	}
	return s
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

func desc(p, s, y, m string) string {
	if p == "" && s == "" {
		return "An unknown release."
	}
	x := tags.Humanize(tags.TagByURI(p), tags.TagByURI(s))
	x = helper.Capitalize(x)
	// x := HumanizeDescription(p, s)
	if m != "" && y != "" {
		x = fmt.Sprintf("%s published in <span class=\"text-nowrap\">%s, %s</a>", x, m, y)
	} else if y != "" {
		x = fmt.Sprintf("%s published in %s", x, y)
	}
	return x + "."
}

// DownloadB returns a human readable string of the file size.
func DownloadB(i any) template.HTML {
	var s string
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
		return template.HTML(fmt.Sprintf("%sDownloadB: %s", typeErr, reflect.TypeOf(i).String()))
	}
	elm := fmt.Sprintf(" <small class=\"text-body-secondary\">(%s)</small>", s)
	return template.HTML(elm)
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

// LinkRelFast returns the groups associated with a release and a link to each group.
// It is a faster version of LinkRelrs and should be used with the templates that have large lists of group names.
func LinkRelFast(a, b any) template.HTML {
	return LinkRelrs(true, a, b)
}

// LinkRelrs returns the groups associated with a release and a link to each group.
func LinkRels(a, b any) template.HTML {
	return LinkRelrs(false, a, b)
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

// linkRelr returns a link to the named group page.
func linkRelr(name string) (string, error) {
	href, err := url.JoinPath("/", "g", helper.Slug(name))
	if err != nil {
		return "", fmt.Errorf("name %q could not be made into a valid url: %w", name, err)
	}
	return href, nil
}

// LinkPage creates a URL anchor element to link to the file page for the record.
func LinkPage(id any) template.HTML {
	s, err := linkID(id, "f")
	if err != nil {
		return template.HTML(err.Error())
	}
	return template.HTML(fmt.Sprintf(`<a class="card-link" href="%s">About</a>`, s))
}

// LinkHref creates a URL path to link to the file page for the record.
func LinkHref(id any) (string, error) {
	return linkID(id, "f")
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

// RecordRels returns the groups associated with a release and joins them with a plus sign.
func RecordRels(a, b any) template.HTML {
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
	return template.HTML(s)
}

// TagSel returns a HTML option tag with the selected attribute if the check value matches the option value.
func TagSel(check, option any) template.HTML {
	x, s := "", ""
	switch val := check.(type) {
	case string:
		x = reflect.ValueOf(val).String()
	case null.String:
		if val.Valid {
			x = val.String
		}
	}
	switch val := option.(type) {
	case string:
		s = reflect.ValueOf(val).String()
	case null.String:
		if val.Valid {
			s = val.String
		}
	}
	x = strings.TrimSpace(x)
	s = strings.TrimSpace(s)
	if x != "" && x == s {
		return template.HTML(fmt.Sprintf("<option value=\"%s\" selected>", s))
	}
	return template.HTML(fmt.Sprintf("<option value=\"%s\">", s))
}

// TagBrief returns a small summary of the tag.
func TagBrief(tag string) template.HTML {
	t := tags.TagByURI(tag)
	s := tags.Infos()[t]
	return template.HTML(s)
}

// TagWithOS returns a small summary of the tag with the operating system.
func TagWithOS(os, tag string) template.HTML {
	p, t := tags.TagByURI(os), tags.TagByURI(tag)
	s := tags.Humanize(p, t)
	return template.HTML(s)
}