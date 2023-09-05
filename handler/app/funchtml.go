package app

import (
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
	"github.com/Defacto2/server/pkg/tags"
	"github.com/volatiletech/null/v8"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	link template.HTML = `<svg class="bi" aria-hidden="true">` +
		`<use xlink:href="/bootstrap-icons.svg#link"></use></svg>`
	wiki template.HTML = `<svg class="bi" aria-hidden="true">` +
		`<use xlink:href="/bootstrap-icons.svg#arrow-right-short"></use></svg>`
	typeErr = "error: received an invalid type to "
)

var (
	archives  = []string{".zip", ".rar", ".7z", ".tar", ".lha", ".lzh", ".arc", ".arj", ".ace", ".tar"}
	documents = []string{".txt", ".nfo", ".diz", ".asc", ".lit", ".rtf", ".doc", ".docx", ".pdf", ".unp", ".htm", ".html", ".xml", ".json", ".csv"}
	images    = []string{".avif", ".gif", ".jpg", ".jpeg", ".jfif", ".png", ".svg", ".webp", ".bmp", ".ico"}
	media     = []string{".mpeg", ".mp1", ".mp2", ".mp3", ".mp4", ".ogg", ".wmv"}
)

// SafeHTML returns a string as a template.HTML type.
// This is intended to be used to prevent HTML escaping.
func SafeHTML(s string) template.HTML {
	return template.HTML(s)
}

// Screenshots returns a picture elment with screenshots for the given uuid.
func (web Web) Screenshot(uuid, desc string) template.HTML {
	fw := filepath.Join(web.Import.ScreenshotsDir, fmt.Sprintf("%s.webp", uuid))
	fp := filepath.Join(web.Import.ScreenshotsDir, fmt.Sprintf("%s.png", uuid))
	webp := strings.Join([]string{config.StaticOriginal(), fmt.Sprintf("%s.webp", uuid)}, "/")
	png := strings.Join([]string{config.StaticOriginal(), fmt.Sprintf("%s.png", uuid)}, "/")
	alt := strings.ToLower(desc) + " screenshot"
	w, p := false, false
	if helper.IsStat(fw) {
		w = true
	}
	if helper.IsStat(fp) {
		p = true
	}
	class := "rounded mx-auto d-block img-fluid"
	if !w && !p {
		return template.HTML("<img src=\"\" loading=\"lazy\" alt=\"screenshot placeholder\" class=\"" + class + "\" />")
	}
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
		return template.HTML("<img src=\"\" loading=\"lazy\" alt=\"thumbnail placeholder\" class=\"" + class + " placeholder\" style=\"" + style + "\" />")
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
		return template.HTML("An unknown release")
	}
	x := tags.Humanize(tags.TagByURI(p), tags.TagByURI(s))
	//x = helper.Capitalize(x)
	return template.HTML(x + ".")
}

// ByteCount returns a human readable string of the byte count.
func ByteCount(b any) template.HTML {
	s := ""
	switch val := b.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		s = helper.ByteCount(i)
	default:
		s = fmt.Sprintf("%sByteFmt: %s", typeErr, reflect.TypeOf(b).String())
	}
	return template.HTML(s)
}

// ByteFile returns a human readable string of the file count and bytes.
func ByteFile(cnt, bytes any) template.HTML {
	s := ""
	switch val := cnt.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		p := message.NewPrinter(language.English)
		s = p.Sprintf("%d", i)
	default:
		s = fmt.Sprintf("%sByteFmt: %s", typeErr, reflect.TypeOf(cnt).String())
		return template.HTML(s)
	}
	switch val := bytes.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		s = fmt.Sprintf("%s <small>(%s)</small>", s, helper.ByteCount(i))
	default:
		s = fmt.Sprintf("%sByteFmt: %s", typeErr, reflect.TypeOf(bytes).String())
		return template.HTML(s)
	}
	return template.HTML(s)
}

// ByteFileS returns a human readable string of the byte count with a named description.
func ByteFileS(name string, cnt, bytes any) template.HTML {
	s := ""
	switch val := cnt.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		if i != 1 {
			name = fmt.Sprintf("%ss", name)
		}
		p := message.NewPrinter(language.English)
		s = p.Sprintf("%d", i)
	default:
		s = fmt.Sprintf("%sByteFmt: %s", typeErr, reflect.TypeOf(cnt).String())
		return template.HTML(s)
	}
	switch val := bytes.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		s = fmt.Sprintf("%s %s <small>(%s)</small>", s, name, helper.ByteCount(i))
	default:
		s = fmt.Sprintf("%sByteFmt: %s", typeErr, reflect.TypeOf(bytes).String())
		return template.HTML(s)
	}
	return template.HTML(s)
}

// Day returns a string of the day number from the day d number between 1 and 31.
func Day(d any) template.HTML {
	s := ""
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

	if p == "" && s == "" {
		return template.HTML("An unknown release")
	}
	x := tags.Humanize(tags.TagByURI(p), tags.TagByURI(s))
	x = helper.Capitalize(x)
	// x := HumanizeDescription(p, s)
	if m != "" && y != "" {
		x = fmt.Sprintf("%s published in <span class=\"text-nowrap\">%s, %s</a>", x, m, y)
	} else if y != "" {
		x = fmt.Sprintf("%s published in %s", x, y)
	}
	return template.HTML(x + ".")
}

// DownloadB returns a human readable string of the file size.
func DownloadB(i any) template.HTML {
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
		return template.HTML(fmt.Sprintf("%sDownloadB: %s", typeErr, reflect.TypeOf(i).String()))
	}
	elm := fmt.Sprintf(" <small class=\"text-body-secondary\">(%s)</small>", s)
	return template.HTML(elm)
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
	av = strings.TrimSpace(av)
	bv = strings.TrimSpace(bv)
	prime, second, s := "", "", ""
	if av == "" && bv == "" {
		return template.HTML("error: unknown group")
	}
	if av != "" {
		ref, err := linkRelr(av)
		if err != nil {
			return template.HTML(fmt.Sprintf("error: %s", err))
		}
		prime = fmt.Sprintf(`<a class="%s" href="%s">%s</a>`, class, ref, fmts.Name(helper.Slug(av)))
	}
	if bv != "" {
		ref, err := linkRelr(bv)
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

// linkRelr returns a link to the named group page.
func linkRelr(name string) (string, error) {
	href, err := url.JoinPath("/", "g", helper.Slug(name))
	if err != nil {
		return "", fmt.Errorf("name %q could not be made into a valid url: %s", name, err)
	}
	return href, nil
}

// LinkPage creates a URL to link to the file page for the record.
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
	platform = strings.TrimSpace(platform)
	// supported formats
	// https://developer.mozilla.org/en-US/docs/Web/Media/Formats/Image_types
	ext := strings.ToLower(filepath.Ext(name))
	switch {
	case slices.Contains(archives, ext):
		// this must always be first
		return template.HTML("")
	case platform == "textamiga", platform == "text":
		break
	case slices.Contains(documents, ext):
		break
	case slices.Contains(images, ext):
		break
	case slices.Contains(media, ext):
		break
	default:
		return template.HTML("")
	}
	s, err := linkID(id, "v")
	if err != nil {
		return template.HTML(err.Error())
	}
	elm := fmt.Sprintf(`&nbsp; <a class="card-link" href="%s">Preview</a>`, s)
	return template.HTML(elm)
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
	href, err := url.JoinPath("/", elem, helper.ObfuscateID(i))
	if err != nil {
		return "", fmt.Errorf("id %d could not be made into a valid url: %s", i, err)
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
	s := ""
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
	cls := "card-subtitle mb-2 text-body-secondary"
	elem := fmt.Sprintf("<h6 class=\"%s\">%s</h6>", cls, val)
	return template.HTML(elem)
}
