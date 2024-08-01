// Package str provides functions for handling string or integer input data.
package str

import (
	"errors"
	"fmt"
	"html/template"
	"image"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/handler/app/internal/exts"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/magicnumber"
	"github.com/dustin/go-humanize"
	"github.com/h2non/filetype"
	"github.com/volatiletech/null/v8"
)

var (
	ErrLinkType = errors.New("the id value is an invalid type")
	ErrNegative = errors.New("value cannot be a negative number")
)

const (
	avif        = ".avif"
	gif         = ".gif"
	jpeg        = ".jpeg"
	jpg         = ".jpg"
	png         = ".png"
	webp        = ".webp"
	textamiga   = "textamiga"
	typeErr     = "error: received an invalid type to "
	closeAnchor = "</a>"
	noFile      = "file not found"
	YYYYMMDD    = "2006-Jan-02"
)

// ArtifactSrc returns a URL to an artifact asset with an cache busting hash.
// The named dir is the directory where the asset is stored, the unid is the unique identifier of the asset
// and the ext is the file extension of the expected asset.
func AssetSrc(abs, dir, unid, ext string) string {
	ext = strings.ToLower(ext)
	name := filepath.Join(dir, unid+ext)
	hash, err := helper.IntegrityFile(name)
	if err != nil {
		return err.Error()
	}
	root := ""
	switch abs {
	case config.Prev:
		root = config.StaticOriginal()
	case config.Thumb:
		root = config.StaticThumb()
	}
	src := strings.Join([]string{root, unid + ext}, "/")
	return fmt.Sprintf("%s?%s", src, hash)
}

// BytesHuman returns the file size for the file record.
func BytesHuman(i int64) string {
	if i == 0 {
		return "(n/a)"
	}
	return humanize.Bytes(uint64(i))
}

// DemozooGetLink returns a HTML link to the Demozoo download links.
func DemozooGetLink(filename, filesize, demozoo, unid any) template.HTML {
	if val, valExists := filename.(null.String); valExists {
		fileExists := val.Valid && val.String != ""
		if fileExists {
			return ""
		}
	}
	if val, valExists := filesize.(null.Int64); valExists {
		fileExists := val.Valid && val.Int64 > 0
		if fileExists {
			return ""
		}
	}
	var zooID int64
	if val, valExists := demozoo.(null.Int64); valExists {
		if !val.Valid || val.Int64 == 0 {
			return ""
		}
		zooID = val.Int64
	}
	if zooID == 0 {
		return ""
	}
	var uID string
	if val, valExists := unid.(null.String); valExists {
		if val.Valid && val.String == "" {
			return ""
		}
		uID = val.String
	}
	if uID == "" {
		return ""
	}
	// s := fmt.Sprintf(`, <a href="" name="editorGetDemozoo"`+
	// 	` data-id="%d" data-uid="%s" id=btn"%s">Use demozoo assets</a>`, zooID, uID, uID)
	return template.HTML(`clone the demozoo assets`)
}

// DownloadB returns a human readable string of the file size.
func DownloadB(i any) template.HTML {
	var s string
	switch val := i.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		s = helper.ByteCount(i)
		s = fmt.Sprintf("(%s)", s)
	case null.Int64:
		if !val.Valid {
			return " <small class=\"text-danger-emphasis\">(n/a)</small>"
		}
		s = BytesHuman(val.Int64)
	default:
		return template.HTML(fmt.Sprintf("%sDownloadB: %s", typeErr, reflect.TypeOf(i).String()))
	}
	elm := fmt.Sprintf(" <small class=\"text-body-secondary\">%s</small>", s)
	return template.HTML(elm)
}

// ImageSample returns a HTML image tag for the given unid.
func ImageSample(unid, previewDir string) template.HTML {
	ext, name, src := "", "", ""
	for _, ext = range []string{avif, webp, png} {
		name = filepath.Join(previewDir, unid+ext)
		src = strings.Join([]string{config.StaticOriginal(), unid + ext}, "/")
		if helper.Stat(name) {
			break
		}
	}
	hash, err := helper.IntegrityFile(name)
	if err != nil {
		return template.HTML(`<div class="card-body">No preview image file</div>`)
	}
	return template.HTML(fmt.Sprintf("<img src=\"%s?%s\" loading=\"lazy\" "+
		"class=\"p-2 img-fluid rounded mx-auto d-block\" alt=\"%s sample\" integrity=\"%s\" />",
		src, hash, ext, hash))
}

// ImageSampleStat returns true if the image sample file exists and is not a 0 byte file.
func ImageSampleStat(unid, previewDir string) bool {
	name := ""
	for _, ext := range []string{avif, webp, png} {
		name = filepath.Join(previewDir, unid+ext)
		if helper.Stat(name) {
			break
		}
	}
	st, err := os.Stat(name)
	if err != nil {
		return false
	}
	return st.Size() > 0
}

// ImageXY returns the image file size and dimensions.
func ImageXY(name string) [2]string {
	switch filepath.Ext(strings.ToLower(name)) {
	case ".jpg", ".jpeg", ".gif", ".png", ".webp":
	default:
		st, err := os.Stat(name)
		if err != nil {
			return [2]string{err.Error(), ""}
		}
		return [2]string{humanize.Comma(st.Size()), ""}
	}
	reader, err := os.Open(name)
	if err != nil {
		return [2]string{err.Error(), ""}
	}
	defer reader.Close()
	st, err := reader.Stat()
	if err != nil {
		return [2]string{err.Error(), ""}
	}
	config, _, err := image.DecodeConfig(reader)
	if err != nil {
		return [2]string{err.Error(), ""}
	}
	return [2]string{humanize.Comma(st.Size()), fmt.Sprintf("%dx%d", config.Width, config.Height)}
}

// LinkID creates a URL to link to the record.
// The id is obfuscated to prevent direct linking.
// The elem is the element to link to, such as 'f' for file or 'd' for download.
func LinkID(id any, elem string) (string, error) {
	var i int64
	switch val := id.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i = reflect.ValueOf(val).Int()
		if i <= 0 {
			return "", fmt.Errorf("app link id %w: %d", ErrNegative, i)
		}
	default:
		return "", fmt.Errorf("app link id %w: %s", ErrLinkType, reflect.TypeOf(id).String())
	}
	href, err := url.JoinPath("/", elem, helper.ObfuscateID(i))
	if err != nil {
		return "", fmt.Errorf("app link id %d could not be made into a valid url: %w", i, err)
	}
	return href, nil
}

// LinkPreviewTip returns a tooltip to describe the preview link.
func LinkPreviewTip(name, platform string) string {
	if name == "" {
		return ""
	}
	platform = strings.TrimSpace(platform)
	ext := strings.ToLower(filepath.Ext(name))
	switch {
	case slices.Contains(exts.Archives(), ext):
		// this case must always be first
		return ""
	case platform == textamiga, platform == "text":
		return "Read this as text"
	case slices.Contains(exts.Documents(), ext):
		return "Read this as text"
	case slices.Contains(exts.Images(), ext):
		return "View this as an image or photo"
	case slices.Contains(exts.Media(), ext):
		return "Play this as media"
	}
	return ""
}

// LinkRelations returns a collection of HTML anchor links that point to artifacts.
//
// The val string is a list of artifact descriptions and their URL ID separated by a semicolon ";".
// Multiple artifact entries are separated by a pipe "|".
//
// For example, "NFO;9f1c2|Intro;a92116e".
func LinkRelations(val string) template.HTML {
	links := strings.Split(val, "|")
	hrefs := []string{}
	const expected = 2
	for _, link := range links {
		s := strings.Split(link, ";")
		if len(s) != expected {
			continue
		}
		name := s[0]
		id := s[1]
		ref := `<a href="/f/` + id + `">` + name + closeAnchor
		if key := helper.DeObfuscate(id); key == "" || key == id {
			ref = fmt.Sprintf("%s ❌ link /f/%s is an invalid download path.", ref, id)
		}
		hrefs = append(hrefs, ref)
	}
	html := strings.Join(hrefs, " + ")
	return template.HTML(html)
}

// LinkRelr returns a link to the named group page.
func LinkRelr(name string) (string, error) {
	href, err := url.JoinPath("/", "g", helper.Slug(name))
	if err != nil {
		return "", fmt.Errorf("name %q could not be made into a valid url: %w", name, err)
	}
	return href, nil
}

// LinkSites returns a collection of HTML anchor links that point to websites.
//
// The val string is a list of website descriptions and their URL ID separated by a semicolon ";".
// Multiple website entries are separated by a pipe "|".
//
// For example, "Site;example.com|Documentation;example.com/doc".
func LinkSites(val string) template.HTML {
	links := strings.Split(val, "|")
	hrefs := []string{}
	const expected = 2
	for _, link := range links {
		s := strings.Split(link, ";")
		if len(s) != expected {
			continue
		}
		name, id := s[0], s[1]
		ref := `<a href="https://` + id + `">` + name + closeAnchor
		hrefs = append(hrefs, ref)
	}
	html := strings.Join(hrefs, " + ")
	return template.HTML(html)
}

// MakeLink returns a HTML anchor link to the named group page.
// When the performant flag is false, the link will apply additional typography to the group name.
// But this should not be used for large lists of links as it will significantly slow down the page rendering.
//
// For example supplying the name "tport"
//   - with performant false will return a link displaying "tPORt"
//   - with performant true will return a link displaying "Tport"
func MakeLink(name, class string, performant bool) (string, error) {
	ref, err := LinkRelr(name)
	if err != nil {
		return "", fmt.Errorf("app make link %w", err)
	}
	capt := helper.Capitalize(strings.ToLower(name))
	title := capt
	if !performant {
		title = releaser.Link(helper.Slug(name))
	}
	s := fmt.Sprintf(`<a class="%s" href="%s">%s</a>`, class, ref, title)
	if capt != "" && title == "" {
		s = "error: could not link group"
	}
	return s, nil
}

// MagicAsTitle returns the magic number description for the named file.
func MagicAsTitle(name string) string {
	r, err := os.Open(name)
	if err != nil {
		return noFile
	}
	defer r.Close()
	sign, err := magicnumber.Find(r)
	if err != nil {
		return err.Error()
	}
	return sign.Title()
}

// MIME returns the [MIME type] for the named file.
//
// [MIME type]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types
func MIME(name string) string {
	file, err := os.Open(name)
	if err != nil {
		return noFile
	}
	defer file.Close()

	const sample = 512
	head := make([]byte, sample)
	_, err = file.Read(head)
	if err != nil {
		return err.Error()
	}

	kind, err := filetype.Match(head)
	if err != nil {
		return err.Error()
	}
	if kind != filetype.Unknown {
		return kind.MIME.Value
	}

	return http.DetectContentType(head)
}

// MkContent makes and/or returns a distinct directory path in the temp directory
// that is used to extract the contents of the content of the file download archive.
// To make the directory distinct it is prefixed with the basename of the src file.
func MkContent(src string) string {
	if src == "" {
		return ""
	}
	path, err := helper.MkContent(src)
	if err != nil {
		return err.Error()
	}
	return path
}

// Releasers returns a HTML links for the primary and secondary group names.
func Releasers(prime, second string) template.HTML {
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

// ReleaserPair returns the primary and secondary releaser groups as two strings.
func ReleaserPair(a, b any) [2]string {
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

// Screenshot returns a picture elment with screenshots for the given unid.
// The unid is the filename of the screenshot image without an extension.
// The desc is the description of the image used for the alt attribute in the img tag.
// Supported formats are webp, png, jpg and avif.
func Screenshot(unid, desc, previewDir string) template.HTML {
	const separator = "/"
	class := "rounded mx-auto d-block img-fluid"
	alt := strings.ToLower(desc) + " screenshot"

	srcW := strings.Join([]string{config.StaticOriginal(), unid + webp}, separator)
	srcP := strings.Join([]string{config.StaticOriginal(), unid + png}, separator)
	srcJ := strings.Join([]string{config.StaticOriginal(), unid + jpg}, separator)
	srcA := strings.Join([]string{config.StaticOriginal(), unid + avif}, separator)

	sizeA := helper.Size(filepath.Join(previewDir, unid+avif))
	sizeJ := helper.Size(filepath.Join(previewDir, unid+jpg))
	sizeP := helper.Size(filepath.Join(previewDir, unid+png))
	sizeW := helper.Size(filepath.Join(previewDir, unid+webp))

	useLegacyJpg := sizeJ > 0 && sizeJ < sizeA && sizeJ < sizeP && sizeJ < sizeW
	if useLegacyJpg {
		return img(srcJ, alt, class, "")
	}
	useLegacyPng := sizeP > 0 && sizeP < sizeA && sizeP < sizeW
	if useLegacyPng {
		return img(srcP, alt, class, "")
	}
	useModernFmts := sizeA > 0 || sizeW > 0
	if useModernFmts {
		elm := template.HTML("<picture>")
		if sizeA > 0 {
			elm += template.HTML(fmt.Sprintf("<source srcset=\"%s\" type=\"image/avif\" />", srcA))
		}
		if sizeW > 0 {
			elm += template.HTML(fmt.Sprintf("<source srcset=\"%s\" type=\"image/webp\" />", srcW))
		}
		if sizeJ > 0 && sizeJ < sizeP {
			elm += img(srcJ, alt, class, "")
		} else if sizeP > 0 {
			elm += img(srcP, alt, class, "")
		}
		elm += "</picture>"
		return elm
	}
	if sizeJ > 0 {
		return img(srcJ, alt, class, "")
	}
	if sizeP > 0 {
		return img(srcP, alt, class, "")
	}
	return ""
}

// img returns a HTML image tag.
func img(src, alt, class, style string) template.HTML {
	return template.HTML(fmt.Sprintf("<img src=\"%s\" loading=\"lazy\" alt=\"%s\" class=\"%s\" style=\"%s\" />",
		src, alt, class, style))
}

// StatHumanize returns the last modified date, size in bytes and size formatted
// of the named file.
func StatHumanize(name string) (string, string, string) {
	stat, err := os.Stat(name)
	if err != nil {
		return noFile, noFile, noFile
	}
	return stat.ModTime().Format(YYYYMMDD),
		humanize.Comma(stat.Size()),
		humanize.Bytes(uint64(stat.Size()))
}

// Thumb returns a HTML image tag or picture element for the given unid.
// The unid is the filename of the thumbnail image without an extension.
// The desc is the description of the image.
func Thumb(unid, desc, thumbDir string, bottom bool) template.HTML {
	fw := filepath.Join(thumbDir, unid+webp)
	fp := filepath.Join(thumbDir, unid+png)
	webp := strings.Join([]string{config.StaticThumb(), unid + webp}, "/")
	png := strings.Join([]string{config.StaticThumb(), unid + png}, "/")
	alt := strings.ToLower(desc) + " thumbnail"
	w, p := false, false
	if helper.Stat(fw) {
		w = true
	}
	if helper.Stat(fp) {
		p = true
	}
	const style = "max-height:400px;"
	class := "m-2 img-fluid rounded mx-auto d-block"
	if !bottom {
		class = "card-img-top"
	}
	if !w && !p {
		return template.HTML("<!-- no thumbnail found -->")
	}
	if w && p {
		elm := "<picture class=\"" + class + "\">" +
			fmt.Sprintf("<source srcset=\"%s\" type=\"image/webp\" />", webp) +
			string(img(png, alt, class, style)) +
			"</picture>"
		return template.HTML(elm)
	}
	if w {
		return img(webp, alt, class, style)
	}
	if p {
		return img(png, alt, class, style)
	}
	return ""
}

// ThumbSample returns a HTML image tag for the given unid.
func ThumbSample(unid, thumbDir string) template.HTML {
	ext, name, src := "", "", ""
	for _, ext = range []string{avif, webp, png} {
		name = filepath.Join(thumbDir, unid+ext)
		src = strings.Join([]string{config.StaticThumb(), unid + ext}, "/")
		if helper.Stat(name) {
			break
		}
	}
	hash, err := helper.IntegrityFile(name)
	if err != nil {
		return template.HTML(`<div class="card-body">No thumbnail picture file</div>`)
	}
	return template.HTML(fmt.Sprintf("<img src=\"%s?%s\" loading=\"lazy\" "+
		"class=\"p-2 img-fluid rounded mx-auto d-block\" alt=\"%s sample\" integrity=\"%s\" />",
		src, hash, ext, hash))
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
