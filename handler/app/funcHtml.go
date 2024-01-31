package app

import (
	"fmt"
	"html/template"
	"net/url"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/tags"
	"github.com/volatiletech/null/v8"
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

func names(s string) string {
	switch s {
	case "bbs", "ftp":
		return "file"
	}
	return s
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
