// Package simple provides functions for handling string or integer input data.
package simple

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"hash/fnv"
	"html/template"
	"image"
	_ "image/gif"  // gif format decoder
	_ "image/jpeg" // jpeg format decoder
	_ "image/png"  // png format decoder
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/handler/releaser"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/extensions"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/tags"
	"github.com/aarondl/null/v8"
	"github.com/dustin/go-humanize"
	"github.com/h2non/filetype"
	_ "golang.org/x/image/webp" // webp format decoder
)

var (
	ErrLinkType = errors.New("the id value is an invalid type")
	ErrName     = errors.New("name is an empty string")
	ErrNegative = errors.New("value cannot be a negative number")
)

const (
	avif        = ".avif"
	jpg         = ".jpg"
	png         = ".png"
	webp        = ".webp"
	textamiga   = "textamiga"
	typeErr     = "error: received an invalid type to "
	closeAnchor = "</a>"
	noFile      = "file not found"
	YYYYMMDD    = "2006-Jan-02"
)

// AssetSrc returns a URL to an artifact asset with an cache busting hash.
// The named dir is the directory where the asset is stored,
// the unid is the unique identifier of the asset,
// and the ext is the file extension of the expected asset.
func AssetSrc(abs, dir, unid, ext string) string {
	ext = strings.ToLower(ext)
	name := filepath.Join(dir, unid+ext)
	integrity, err := helper.IntegrityFile(name)
	if err != nil {
		return err.Error()
	}
	root := ""
	switch abs {
	case config.AbsPreview:
		root = config.StaticOriginal()
	case config.AbsThumbnail:
		root = config.StaticThumb()
	}
	src := strings.Join([]string{root, unid + ext}, "/")
	return fmt.Sprintf("%s?%s", src, integrity)
}

// BytesHuman returns the file size for the file record.
func BytesHuman(i int64) string {
	if i == 0 {
		return "(n/a)"
	}
	return humanize.Bytes(uint64(math.Abs(float64(i))))
}

// CleanFname runs the string such as a filename through a HTML template
// to remove any possible XSS problems such as < > characters.
func CleanFname(s string) (string, error) {
	const format = "simple clean fname %s tmpl: %w"
	if s == "" {
		return "", nil
	}
	// template placeholder
	type TemplateData struct {
		Fname string
	}
	tmpl, err := template.New("cleanTmpl").Parse(`{{.Fname}}`)
	if err != nil {
		return "", fmt.Errorf(format, "new", err)
	}
	data := TemplateData{Fname: s}
	var wr bytes.Buffer
	err = tmpl.Execute(&wr, data)
	if err != nil {
		return "", fmt.Errorf(format, "execute", err)
	}
	return wr.String(), nil
}

// CleanHTML removes all HTML tags from content, returning plain text.
func CleanHTML(html string) string {
	if html == "" {
		return html
	}

	// First, handle <q> tags specially - convert to quoted text (non-greedy)
	re := regexp.MustCompile(`<q\b[^>]*>(.*?)<\/q>`)
	html = re.ReplaceAllString(html, `"$1"`)

	// Convert common HTML entities to regular characters
	html = strings.ReplaceAll(html, "&amp;", "&")
	html = strings.ReplaceAll(html, "&lt;", "<")
	html = strings.ReplaceAll(html, "&gt;", ">")

	// Remove all HTML tags and replace with single space
	re = regexp.MustCompile(`<[^>]*>`)
	result := re.ReplaceAllString(html, " ")

	// Fix common spacing issues
	// Remove spaces before punctuation
	re = regexp.MustCompile(`\s+([.,;:!?])`)
	result = re.ReplaceAllString(result, "${1}")

	// Remove spaces after opening parentheses and before closing parentheses
	re = regexp.MustCompile(`\(\s+`)
	result = re.ReplaceAllString(result, "(")
	re = regexp.MustCompile(`\s+\)`)
	result = re.ReplaceAllString(result, ")")

	// Add space after punctuation if missing (but not if already there)
	re = regexp.MustCompile(`([.!?])(\w)`)
	result = re.ReplaceAllString(result, "${1} ${2}")

	// Handle &nbsp; by converting to single space (preserves intent without double spacing)
	result = strings.ReplaceAll(result, "&nbsp;", " ")

	// Clean up all multiple spaces
	re = regexp.MustCompile(`[\s\n\r\t]+`)
	result = re.ReplaceAllString(result, " ")

	return strings.TrimSpace(result)
}

// DemozooGetLink returns a HTML link to the Demozoo download links.
// The filename and filesize are used to determine if the file exists.
// The demozoo is the ID for the production on Demozoo.
// The unid is the unique identifier for the file record.
func DemozooGetLink(filename, filesize, demozoo, unid any) template.HTML {
	if s, ok := filename.(null.String); ok {
		exist := s.Valid && s.String != ""
		if exist {
			return ""
		}
	}
	if i, ok := filesize.(null.Int64); ok {
		exist := i.Valid && i.Int64 > 0
		if exist {
			return ""
		}
	}
	var demozooID int64
	if i, ok := demozoo.(null.Int64); ok {
		if !i.Valid || i.Int64 == 0 {
			return ""
		}
		demozooID = i.Int64
	}
	if demozooID == 0 {
		return ""
	}
	var uniqueID string
	if s, ok := unid.(null.String); ok {
		if s.Valid && s.String == "" {
			return ""
		}
		uniqueID = s.String
	}
	if uniqueID == "" {
		return ""
	}
	return template.HTML(`clone the demozoo assets`)
}

// DownloadInBytes returns a human readable string of the file size.
// The value must be an integer or a null.Int64.
func DownloadInBytes(v any) template.HTML {
	var value string
	switch val := v.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		b := reflect.ValueOf(val).Int()
		value = fmt.Sprintf("(%s)", helper.ByteCount(b))
	case null.Int64:
		if !val.Valid {
			return ` <small class="text-danger-emphasis">(n/a)</small>`
		}
		value = BytesHuman(val.Int64)
	default:
		const format = "%sDownloadB: %s"
		return template.HTML(fmt.Sprintf(format, typeErr, reflect.TypeOf(v).String()))
	}
	const format = ` <small class="text-body-secondary">%s</small>`
	return template.HTML(fmt.Sprintf(format, value))
}

// Hash creates a stable hash ID from a string and returns it as base64.
func Hash(s string) string {
	h := fnv.New64a()
	h.Write([]byte(s))
	src := h.Sum(nil)
	// Use URLEncoding to avoid special characters
	return base64.URLEncoding.EncodeToString(src)
}

// ImageSample returns a HTML image tag for the given unid.
// The preview is the directory where the preview images are stored.
func ImageSample(unid string, preview dir.Directory) template.HTML {
	alt, name, src := "", "", ""
	exts := []string{avif, webp, png, jpg}
	for alt = range slices.Values(exts) {
		name = preview.Join(unid + alt)
		src = strings.Join([]string{config.StaticOriginal(), unid + alt}, "/")
		if helper.Stat(name) {
			break
		}
	}
	integrity, err := helper.IntegrityFile(name)
	if err != nil {
		return template.HTML(`<div class="card-body">No preview image file</div>`)
	}
	const format = `<img src="%s?%s" loading="lazy" class="%s" alt="%s sample" integrity="%s" />`
	const class = `p-2 img-fluid rounded mx-auto d-block`
	return template.HTML(fmt.Sprintf(format, src, integrity, class, alt, integrity))
}

// ImageSampleStat returns true if the image sample file exists and is not a 0 byte file.
// The preview is the directory where the preview images are stored.
func ImageSampleStat(unid string, preview dir.Directory) bool {
	exts := []string{avif, webp, png}
	const minimum = 60
	for ext := range slices.Values(exts) {
		name := preview.Join(unid + ext)
		st, err := os.Stat(name)
		if err != nil {
			continue // any errors inc file not found are okay, continue with the next extension
		}
		if st.Size() > minimum {
			return true
		}
	}
	return false
}

// ImageXY returns the named image filesize and dimensions as a styled string array.
// The dimensions are returned as a string in the format "width x height".
// If the file does not exist, an empty string array is returned.
//
// For example, the returned values are:
//
//	["4,163", "500x500"]
//
// However, if the file does not exist, the returned values are:
//
//	["0", ""]
func ImageXY(sl *slog.Logger, name string) [2]string {
	const msg = "simple image size and dimension"
	if sl == nil {
		sl = logs.Discard()
	}

	notfound := [2]string{"0", ""}
	invalid := func(err error) [2]string {
		sl.Info(msg+" caused an error",
			slog.String("name", name), slog.Any("error", err))
		return [2]string{err.Error(), ""}
	}

	// open /mnt/volume/assets/images000/ca6cf279-3758-4e1e-8e8b-f60871e877be.jpg: no such file or directoryB
	switch filepath.Ext(strings.ToLower(name)) {
	case ".jpg", ".jpeg", ".gif", ".png", ".webp":
	// extension is okay, so do nothing
	default:
		// lookup unique file names without file extensions
		st, err := os.Stat(name)
		if errors.Is(err, os.ErrNotExist) {
			return notfound
		}
		if err != nil {
			return invalid(err)
		}
		// return the file size but without any image dimensions
		return [2]string{
			humanize.Comma(st.Size()),
			"",
		}
	}

	// open files with a known image file extension
	r, err := os.Open(name)
	if errors.Is(err, os.ErrNotExist) {
		return notfound
	}
	if err != nil {
		return invalid(err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			sl.Info(msg+" cannot close openned file",
				slog.String("name", name), slog.Any("error", err))
		}
	}()
	st, err := r.Stat()
	if err != nil {
		return invalid(err)
	}
	c, _, err := image.DecodeConfig(r)
	if err != nil {
		return invalid(err)
	}
	const format = `%dx%d`
	return [2]string{
		humanize.Comma(st.Size()),
		fmt.Sprintf(format, c.Width, c.Height),
	}
}

// LinkID creates a URL to link to the record.
// The id is obfuscated to prevent direct linking.
// The elem is the element to link to, such as 'f' for file or 'd' for download.
func LinkID(id any, elem string) (string, error) {
	const format = "app link id %d%s: %w"
	var i int64
	switch val := id.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i = reflect.ValueOf(val).Int()
		if i <= 0 {
			return "", fmt.Errorf(format, i, "", ErrNegative)
		}
	default:
		return "", fmt.Errorf(format, i, reflect.TypeOf(id).String(), ErrLinkType)
	}
	href, err := url.JoinPath("/", elem, helper.ObfuscateID(i))
	if err != nil {
		return "", fmt.Errorf(format, i, "could not be made into a valid url", err)
	}
	return href, nil
}

// LinkPreviewTip returns a tooltip to describe the preview link.
// The name is the filename of the file to preview and does not require path information.
// The platform is the platform or format of the file.
func LinkPreviewTip(name, platform string) string {
	if name == "" {
		return ""
	}
	platform = strings.TrimSpace(platform)
	ext := strings.ToLower(filepath.Ext(name))
	switch {
	case slices.Contains(extensions.Archive(), ext):
		// this case must always be first
		return ""
	case platform == tags.Markup.String():
		return "Read this as HTML"
	case platform == textamiga, platform == tags.Text.String():
		return "Read this as text"
	case slices.Contains(extensions.Document(), ext):
		return "Read this as text"
	case slices.Contains(extensions.Image(), ext):
		return "View this as an image or photo"
	case slices.Contains(extensions.Media(), ext):
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
	for link := range slices.Values(links) {
		s := strings.Split(link, ";")
		if len(s) != expected {
			continue
		}
		name := s[0]
		id := s[1]
		ref := `<a href="/f/` + id + `">` + name + closeAnchor
		if key := helper.DeObfuscate(id); key == "" || key == id {
			const format = "%s ❌ link /f/%s is an invalid download path."
			ref = fmt.Sprintf(format, ref, id)
		}
		hrefs = append(hrefs, ref)
	}
	html := strings.Join(hrefs, " + ")
	return template.HTML(html)
}

// LinkRelr returns a link to the named group page.
//
// Providing the name "a group" will return "/g/a-group".
func LinkRelr(name string) (string, error) {
	if name == "" {
		return "", ErrName
	}
	href, err := url.JoinPath("/", "g", helper.Slug(name))
	if err != nil {
		const format = "name %q could not be made into a valid url: %w"
		return "", fmt.Errorf(format, name, err)
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
	for link := range slices.Values(links) {
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
func MakeLink(id, name, class string, performant bool) (string, error) {
	href, err := LinkRelr(name)
	if err != nil {
		return "", fmt.Errorf("app make link %w", err)
	}
	capt := helper.Capitalize(strings.ToLower(name))
	value := capt
	if !performant {
		value = releaser.Link(helper.Slug(name))
	}
	const format = `<a id="named-group-page-%s" class="%s" href="%s">%s</a>`
	s := fmt.Sprintf(format, id, class, href, value)
	if capt != "" && value == "" {
		s = "error: could not link group"
	}
	return s, nil
}

// MagicAsTitle returns the magic number description for the named file.
func MagicAsTitle(sl *slog.Logger, name string) string {
	const msg = "simple magic as title"
	if sl == nil {
		sl = logs.Discard()
	}
	r, err := os.Open(name)
	if err != nil {
		sl.Info(msg+" could not open named file",
			slog.String("name", name), slog.Any("error", err))
		return noFile
	}
	defer func() {
		if err := r.Close(); err != nil {
			sl.Info(msg+" could not close named file",
				slog.String("name", name), slog.Any("error", err))
		}
	}()
	sign := magicnumber.Find(r)
	return sign.Title()
}

// MIME returns the [MIME type] for the named file.
//
// [MIME type]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types
func MIME(sl *slog.Logger, name string) string {
	const msg = "simple mime type lookup"
	if sl == nil {
		sl = logs.Discard()
	}
	file, err := os.Open(name)
	if err != nil {
		sl.Info(msg+" could not open file",
			slog.String("name", name), slog.Any("error", err))
		return noFile
	}
	defer func() {
		if err := file.Close(); err != nil {
			sl.Info(msg+" could not close file",
				slog.String("name", name), slog.Any("error", err))
		}
	}()

	const sample = 512
	head := make([]byte, sample)
	_, err = file.Read(head)
	if err != nil {
		sl.Info(msg+" could not read a sample of file",
			slog.String("name", name), slog.Int("sample size", sample),
			slog.Any("error", err))
		return err.Error()
	}

	kind, err := filetype.Match(head)
	if err != nil {
		sl.Info(msg+" could not match file type",
			slog.String("name", name), slog.Any("error", err))
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
//
// The returned path should be removed after use.
func MkContent(sl *slog.Logger, src string) string {
	const msg = "simple make content"
	if sl == nil {
		sl = logs.Discard()
	}
	if src == "" {
		return ""
	}
	path, err := helper.MkContent(src)
	if err != nil {
		sl.Info(msg+" caused an error",
			slog.String("src", src), slog.Any("error", err))
		return err.Error()
	}
	return path
}

// Releasers returns a HTML links for the primary and secondary group names.
func Releasers(prime, second string, magazine bool) template.HTML {
	var s string
	switch {
	case magazine && prime != "" && second != "":
		const format = `%s <small>published by</small> %s`
		s = fmt.Sprintf(format, second, prime)
	case prime != "" && second != "":
		const format = `%s <strong class="text-secondary">+</strong> %s`
		s = fmt.Sprintf(format, prime, second)
	case prime != "":
		s = prime
	case second != "":
		s = second
	default:
		return ""
	}
	return template.HTML(s)
}

// OpenGraphImg returns a URI for a thumbnail that is intended
// to be used in the 'og:image' content metadata attribute.
func OpenGraphImg(unid string, preview, thumbnail dir.Directory) string {
	name, src := ogImage(unid, config.StaticOriginal(), preview)
	integrity, err := helper.IntegrityFile(name)
	if err != nil {
		name, src := ogImage(unid, config.StaticThumb(), thumbnail)
		integrity, err := helper.IntegrityFile(name)
		if err != nil {
			return "/image/layout/defacto2-ascii.png"
		}
		return src + "?" + integrity
	}
	return src + "?" + integrity
}

func ogImage(unid, path string, dd dir.Directory) (string, string) {
	ext, name, src := "", "", ""
	exts := []string{avif, webp, png}
	for ext = range slices.Values(exts) {
		name = dd.Join(unid + ext)
		src = strings.Join([]string{path, unid + ext}, "/")
		if helper.Stat(name) {
			break
		}
	}
	return name, src
}

// ReleaserPair returns the primary and secondary releaser groups as two strings.
func ReleaserPair(a, b any) [2]string {
	r1, r2 := "", ""
	switch val := a.(type) {
	case string:
		r1 = reflect.ValueOf(val).String()
	case null.String:
		if val.Valid {
			r1 = val.String
		}
	}
	switch val := b.(type) {
	case string:
		r2 = reflect.ValueOf(val).String()
	case null.String:
		if val.Valid {
			r2 = val.String
		}
	}
	r1 = strings.TrimSpace(r1)
	r2 = strings.TrimSpace(r2)
	switch {
	case r1 != "" && r2 != "":
		return [2]string{r1, r2}
	case r2 != "":
		return [2]string{r2, ""}
	case r1 != "":
		return [2]string{r1, ""}
	}
	return [2]string{}
}

// Screenshot returns a image element with screenshots for the given unid.
// If a webp or avif image is available, and a legacy png or jpg image is available,
// a picture element is used to provide multiple sources for the image. Otherwise,
// a single img element is used.
//
// The unid is the filename of the screenshot image without an extension.
// The desc is the description of the image used for the alt attribute in the img tag.
// The preview is the directory where the preview images are stored.
//
// Supported formats are webp, png, jpg and avif.
func Screenshot(unid, desc string, preview dir.Directory) template.HTML {
	const separator = "/"
	alt := strings.ToLower(desc) + " screenshot"
	// source links
	srcWeb := strings.Join([]string{config.StaticOriginal(), unid + webp}, separator)
	srcPng := strings.Join([]string{config.StaticOriginal(), unid + png}, separator)
	srcJpg := strings.Join([]string{config.StaticOriginal(), unid + jpg}, separator)
	srcAvi := strings.Join([]string{config.StaticOriginal(), unid + avif}, separator)
	// image file sizes
	sizeWeb := helper.Size(preview.Join(unid + webp))
	sizePng := helper.Size(preview.Join(unid + png))
	sizeJpg := helper.Size(preview.Join(unid + jpg))
	sizeAvi := helper.Size(preview.Join(unid + avif))
	// image file integrity hash values
	integrityWeb, _ := helper.IntegrityFile(preview.Join(unid + webp))
	integrityPng, _ := helper.IntegrityFile(preview.Join(unid + png))
	integrityJpg, _ := helper.IntegrityFile(preview.Join(unid + jpg))
	integrityAvi, _ := helper.IntegrityFile(preview.Join(unid + avif))

	usePicture := (sizeAvi > 0 || sizeWeb > 0) && (sizeJpg > 0 || sizePng > 0)
	if usePicture {
		var elm string
		switch {
		case sizeAvi > 0:
			const format = `<source srcset="%s?%s" type="image/avif" integrity="%s" />`
			elm += fmt.Sprintf(format, srcAvi, integrityAvi, integrityAvi)
		case sizeWeb > 0:
			const format = `<source srcset="%s?%s" type="image/webp" integrity="%s" />`
			elm += fmt.Sprintf(format, srcWeb, integrityWeb, integrityWeb)
		}
		// the <picture> element is used to provide multiple sources for an image.
		// if no <img> element is provided, the <picture> element won't be rendered by the browser.
		useSmallerJpg := sizeJpg > 0 && sizeJpg < sizePng
		switch {
		case useSmallerJpg:
			elm += img(srcJpg, alt, integrityJpg)
		case sizePng > 0:
			elm += img(srcPng, alt, integrityPng)
		default:
			elm += img(srcJpg, alt, integrityJpg)
		}
		return template.HTML(`<picture>` + elm + `</picture>`)
	}
	var elm string
	switch {
	case sizeAvi > 0:
		elm = img(srcAvi, alt, integrityAvi)
	case sizeWeb > 0:
		elm = img(srcWeb, alt, integrityWeb)
	case sizeJpg > 0:
		elm = img(srcJpg, alt, integrityJpg)
	case sizePng > 0:
		elm = img(srcPng, alt, integrityPng)
	default:
		elm = ""
	}
	return template.HTML(elm)
}

// img returns a HTML image tag.
func img(src, alt, integrity string) string {
	const format = `<img src="%s?%s" loading="lazy" alt="%s" class="rounded mx-auto d-block img-fluid" integrity="%s" />`
	return fmt.Sprintf(format, src, integrity, alt, integrity)
}

// StatHumanize returns the last modified date, size in bytes and size formatted
// of the named file.
// If the file does not exist, the string "file not found" is returned.
//
// An example of the returned values are:
//
//	"2024-Sep-03", "4,163", "4.2 kB"
func StatHumanize(name string) (string, string, string) {
	st, err := os.Stat(name)
	if err != nil {
		return noFile, noFile, noFile
	}
	u := uint64(math.Abs(float64(st.Size())))
	return st.ModTime().Format(YYYYMMDD),
		humanize.Comma(st.Size()),
		humanize.Bytes(u)
}

// Thumb returns a HTML image tag or picture element for the given unid.
// The unid is the filename of the thumbnail image without an extension.
// The desc is the description of the image.
// The thumbnail is the directory where the thumbnail images are stored.
// The bottom flag is true if the image should be displayed at the bottom of the container element.
func Thumb(unid, desc string, thumbnail dir.Directory, bottom bool) template.HTML {
	srcsetW := strings.Join([]string{config.StaticThumb(), unid + webp}, "/")
	srcsetP := strings.Join([]string{config.StaticThumb(), unid + png}, "/")
	alt := strings.ToLower(desc) + " thumbnail"
	w, p := false, false
	name := thumbnail.Join(unid + webp)
	if helper.Stat(name) {
		w = true
	}
	name = thumbnail.Join(unid + png)
	if helper.Stat(name) {
		p = true
	}
	if !w && !p {
		const comment = `<!-- no thumbnail found -->`
		return template.HTML(comment)
	}
	const style = "max-height:400px;"
	class := "card-img-bottom" // m-2 img-fluid rounded mx-auto d-block"
	if !bottom {
		class = "card-img-top"
	}
	if w && p {
		const format = `<source srcset="%s" type="image/webp" />`
		source := fmt.Sprintf(format, srcsetW) + string(imgTag(srcsetP, alt, class, style))
		return template.HTML(`<picture class="` + class + `">` + source + `</picture>`)
	}
	src := srcsetW
	if p {
		src = srcsetP
	}
	return imgTag(src, alt, class, style)
}

// imgTag returns a HTML image element.
func imgTag(src, alt, class, style string) template.HTML {
	const format = `<img src="%s" loading="lazy" alt="%s" class="%s" style="%s" />`
	return template.HTML(fmt.Sprintf(format, src, alt, class, style))
}

// ThumbSample returns a HTML image tag for the given unid.
// The unid is the filename of the thumbnail image without an extension.
// The thumbDir is the directory where the thumbnail images are stored.
func ThumbSample(unid string, thumbnail dir.Directory) template.HTML {
	alt, name, src := "", "", ""
	exts := []string{avif, webp, png}
	for alt = range slices.Values(exts) {
		name = thumbnail.Join(unid + alt)
		src = strings.Join([]string{config.StaticThumb(), unid + alt}, "/")
		if helper.Stat(name) {
			break
		}
	}
	integrity, err := helper.IntegrityFile(name)
	if err != nil {
		return template.HTML(`<div class="card-body">No thumbnail picture file</div>`)
	}
	const format = `<img src="%s?%s" loading="lazy" class="%s" alt="%s sample" integrity="%s" />`
	const class = `p-2 img-fluid rounded mx-auto d-block`
	return template.HTML(fmt.Sprintf(format, src, integrity, class, alt, integrity))
}

// Updated returns a string of the time since the given time t.
// If the time is not valid, an empty string is returned.
// An example of the returned string is:
//
//	"Time 1 day ago"
func Updated(t any, s string) string {
	if t == nil {
		return ""
	}
	if s == "" {
		s = "Time"
	}
	justnow := "less than a minute"
	const seconds = false
	switch val := t.(type) {
	case null.Time:
		if !val.Valid {
			return ""
		}
		x := helper.TimeDistance(val.Time, time.Now(), seconds)
		if x == justnow {
			return s + " just now"
		}
		return s + " " + x + " ago"
	case time.Time:
		x := helper.TimeDistance(val, time.Now(), seconds)
		if x == justnow {
			return s + " just now"
		}
		return s + " " + x + " ago"
	default:
		const format = `%supdated: %s`
		return fmt.Sprintf(format, typeErr, reflect.TypeOf(t).String())
	}
}
