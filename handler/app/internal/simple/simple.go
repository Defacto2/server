// Package simple provides functions for handling string or integer input data.
//
//nolint:gochecknoglobals,nonamedreturns
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
	"strconv"
	"strings"
	"sync"
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
	ErrLinkType = errors.New("simple: the id value is an invalid type")
	ErrName     = errors.New("simple: name is an empty string")
	ErrNegative = errors.New("simple: value cannot be a negative number")
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
	return src + "?" + integrity
}

// BytesHuman returns the file size for the file record.
func BytesHuman(i int64) string {
	if i == 0 {
		return "(n/a)"
	}
	return humanize.Bytes(uint64(math.Abs(float64(i))))
}

type templateData struct {
	Fname string
}

var templateClean = sync.OnceValue(func() *template.Template {
	return template.Must(template.New("cleanTmpl").Parse(`{{.Fname}}`))
})

// CleanFname runs the string such as a filename through a HTML template
// to remove any possible XSS problems such as < > characters.
func CleanFname(s string) (string, error) {
	if s == "" {
		return "", nil
	}

	var wr bytes.Buffer
	const size = 16
	wr.Grow(len(s) + size)

	if err := templateClean().Execute(&wr, templateData{Fname: s}); err != nil {
		return "", fmt.Errorf("simple clean fname execute: %w", err)
	}

	return wr.String(), nil
}

var (
	reQuoteElements   = regexp.MustCompile(`<q\b[^>]*>(.*?)<\/q>`)
	reHTMLElements    = regexp.MustCompile(`<[^>]*>`)
	reSpacePunct      = regexp.MustCompile(`\s+([.,;:!?])`)
	reParentheseOpen  = regexp.MustCompile(`\(\s+`)
	reParentheseClose = regexp.MustCompile(`\s+\)`)
	rePunctSpace      = regexp.MustCompile(`([.!?])(\w)`)
	reMultipleSpace   = regexp.MustCompile(`[\s\n\r\t]+`)
)

// CleanHTML removes all HTML tags from content, returning plain text.
func CleanHTML(html string) string {
	if html == "" {
		return html
	}

	// First, handle <q> tags specially - convert to quoted text (non-greedy)
	html = reQuoteElements.ReplaceAllString(html, `"$1"`)

	// Convert common HTML entities to regular characters
	html = strings.ReplaceAll(html, "&amp;", "&")
	html = strings.ReplaceAll(html, "&lt;", "<")
	html = strings.ReplaceAll(html, "&gt;", ">")

	// Remove all HTML tags and replace with single space
	src := reHTMLElements.ReplaceAllString(html, " ")

	// Fix common spacing issues
	// Remove spaces before punctuation
	src = reSpacePunct.ReplaceAllString(src, "${1}")

	// Remove spaces after opening parentheses and before closing parentheses
	src = reParentheseOpen.ReplaceAllString(src, "(")
	src = reParentheseClose.ReplaceAllString(src, ")")

	// Add space after punctuation if missing (but not if already there)
	src = rePunctSpace.ReplaceAllString(src, "${1} ${2}")

	// Handle &nbsp; by converting to single space (preserves intent without double spacing)
	src = strings.ReplaceAll(src, "&nbsp;", " ")

	// Clean up all multiple spaces
	return strings.TrimSpace(reMultipleSpace.ReplaceAllString(src, " "))
}

// DemozooGetLink returns a HTML link to the Demozoo download links.
// The filename and filesize are used to determine if the file exists.
// The demozoo is the ID for the production on Demozoo.
// The unid is the unique identifier for the file record.
func DemozooGetLink(filename, filesize, demozoo, unid any) template.HTML {
	if s, ok := filename.(null.String); ok {
		if exist := s.Valid && s.String != ""; exist {
			return ""
		}
	}
	if i, ok := filesize.(null.Int64); ok {
		if exist := i.Valid && i.Int64 > 0; exist {
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
	var bytes int64

	switch val := v.(type) {
	case int:
		bytes = int64(val)
	case int64:
		bytes = val
	case int32:
		bytes = int64(val)
	case int16:
		bytes = int64(val)
	case int8:
		bytes = int64(val)
	case uint:
		if val > math.MaxInt64 {
			bytes = math.MaxInt64
		} else {
			bytes = int64(val)
		}
	case uint64:
		if val > math.MaxInt64 {
			bytes = math.MaxInt64
		} else {
			bytes = int64(val)
		}
	case uint32:
		bytes = int64(val)
	case uint16:
		bytes = int64(val)
	case uint8:
		bytes = int64(val)
	case null.Int64:
		if !val.Valid {
			return ` <small class="text-danger-emphasis">(n/a)</small>`
		}
		bytes = val.Int64
	default:
		return template.HTML(fmt.Sprintf("%sDownloadB: %T", typeErr, v))
	}

	return template.HTML(` <small class="text-body-secondary">(` +
		helper.ByteCount(bytes) + `)</small>`)
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
	ext, name, src := "", "", ""

	exts := [...]string{avif, webp, png, jpg}
	for _, ext = range exts {
		name = preview.Join(unid + ext)
		src = strings.Join([]string{config.StaticOriginal(), unid + ext}, "/")
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
	return template.HTML(fmt.Sprintf(format, src, integrity, class, ext, integrity))
}

// ImageSampleStat returns true if the image sample file exists and is not a 0 byte file.
// The preview is the directory where the preview images are stored.
func ImageSampleStat(unid string, preview dir.Directory) bool {
	const minimum = 60

	exts := [...]string{avif, webp, png}
	for _, ext := range exts {
		name := preview.Join(unid + ext)

		st, err := os.Stat(name)
		if err != nil {
			continue // any errors, including the file not found are okay, continue on
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
		return [2]string{humanize.Comma(st.Size()), ""}
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

	return [2]string{
		humanize.Comma(st.Size()),
		strconv.Itoa(c.Width) + "x" + strconv.Itoa(c.Height),
	}
}

// LinkID creates a URL to link to the record.
// The id is obfuscated to prevent direct linking.
// The elem is the element to link to, such as 'f' for file or 'd' for download.
func LinkID(id any, elem string) (string, error) {
	const format = "app link id %d%s: %w"

	var i int64
	switch val := id.(type) {
	case
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
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
	case platform == textamiga,
		platform == tags.Text.String():
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
	const expected = 2
	hrefs := []string{}

	for link := range strings.SplitSeq(val, "|") {
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

	return template.HTML(strings.Join(hrefs, " + "))
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
	const expected = 2
	hrefs := []string{}

	for link := range strings.SplitSeq(val, "|") {
		s := strings.Split(link, ";")
		if len(s) != expected {
			continue
		}

		name, id := s[0], s[1]
		ref := `<a href="https://` + id + `">` + name + closeAnchor
		hrefs = append(hrefs, ref)
	}

	return template.HTML(strings.Join(hrefs, " + "))
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
	if sl == nil {
		sl = logs.Discard()
	}

	logInfo := func(s string, err error) {
		sl.Info("simple magic as title could not "+s+" named file",
			slog.String("name", name), slog.Any("error", err))
	}

	r, err := os.Open(name)
	if err != nil {
		logInfo("open", err)
		return noFile
	}

	sign := magicnumber.Find(r)

	if err := r.Close(); err != nil {
		logInfo("close", err)
	}

	return sign.Title()
}

// MIME returns the [MIME type] for the named file.
//
// [MIME type]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types
func MIME(sl *slog.Logger, name string) string {
	if sl == nil {
		sl = logs.Discard()
	}

	logInfo := func(s string, err error) {
		sl.Info("simple mime lookup could not "+s+" named file",
			slog.String("name", name), slog.Any("error", err))
	}

	file, err := os.Open(name)
	if err != nil {
		logInfo("open", err)
		return noFile
	}
	defer func() {
		if err := file.Close(); err != nil {
			logInfo("close", err)
		}
	}()

	const sample = 512
	data := make([]byte, sample)
	_, err = file.Read(data)
	if err != nil {
		logInfo("read sample of", err)
		return err.Error()
	}

	kind, err := filetype.Match(data)
	if err != nil {
		logInfo("match file type of", err)
		return err.Error()
	}
	if kind != filetype.Unknown {
		return kind.MIME.Value
	}

	return http.DetectContentType(data)
}

// MkdirStale makes and/or returns a distinct directory path in the system temporary
// directory that is used to extract the contents a file download archive.
//
// To make the directory distinct it is prefixed with the basename of the src file.
// The returned path should be removed after use.
func MkdirStale(sl *slog.Logger, src string) string {
	if src == "" {
		return ""
	}
	if sl == nil {
		sl = logs.Discard()
	}

	path, err := dir.MkdirStale(src)
	if err != nil {
		sl.Info("simple stale dir caused an error",
			slog.String("src", src), slog.Any("error", err))
		return err.Error()
	}

	return path
}

// Releasers returns a HTML links for the primary and secondary group names.
func Releasers(prime, second string, magazine bool) template.HTML {
	switch {
	case magazine && prime != "" && second != "":
		return template.HTML(second + ` <small>published by</small> ` + prime)
	case prime != "" && second != "":
		return template.HTML(prime + ` <strong class="text-secondary">+</strong> ` + second)
	case prime != "":
		return template.HTML(prime)
	case second != "":
		return template.HTML(second)
	default:
		return ""
	}
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

func ogImage(unid, path string, dd dir.Directory) (name string, src string) {
	exts := [...]string{avif, webp, png}
	for _, ext := range exts {
		name = dd.Join(unid + ext)
		src = strings.Join([]string{path, unid + ext}, "/")
		if helper.Stat(name) {
			return name, src
		}
	}

	return "", ""
}

// ReleaserPair returns the primary and secondary releaser groups as two strings.
func ReleaserPair(a, b any) [2]string {
	// releaser 1
	r1 := ""
	switch val := a.(type) {
	case string:
		r1 = reflect.ValueOf(val).String()
	case null.String:
		if val.Valid {
			r1 = val.String
		}
	}
	r1 = strings.TrimSpace(r1)

	// releaser 2
	r2 := ""
	switch val := b.(type) {
	case string:
		r2 = reflect.ValueOf(val).String()
	case null.String:
		if val.Valid {
			r2 = val.String
		}
	}
	r2 = strings.TrimSpace(r2)

	switch {
	case r1 != "" && r2 != "":
		return [2]string{r1, r2}
	case r2 != "":
		return [2]string{r2, ""}
	case r1 != "":
		return [2]string{r1, ""}
	default:
		return [2]string{}
	}
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

	img := func(src, alt, integrity string) string {
		return `<img src="` + src + `?` + integrity + `" loading="lazy" alt="` + alt +
			`" class="rounded mx-auto d-block img-fluid" integrity="` + integrity + `" />`
	}

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

	alt := strings.ToLower(desc) + " screenshot"
	var elm strings.Builder
	const size = 50
	elm.Grow(size)

	usePictureElm := (sizeAvi > 0 || sizeWeb > 0) && (sizeJpg > 0 || sizePng > 0)
	if usePictureElm {
		elm.WriteString(`<picture>`)
		switch {
		case sizeAvi > 0:
			elm.WriteString(`<source srcset="` + srcAvi + `?` + integrityAvi +
				`" type="image/avif" integrity="` + integrityAvi + `" />`)
		case sizeWeb > 0:
			elm.WriteString(`<source srcset="` + srcWeb + `?` + integrityWeb +
				`" type="image/webp" integrity="` + integrityWeb + `" />`)
		}
		// the <picture> element is used to provide multiple sources for an image.
		// if no <img> element is provided, the <picture> element won't be rendered by the browser.
		switch {
		case sizePng > 0 && (sizeJpg <= 0 || sizePng <= sizeJpg):
			elm.WriteString(img(srcPng, alt, integrityPng))
		default:
			elm.WriteString(img(srcJpg, alt, integrityJpg))
		}
		elm.WriteString(`</picture>`)
		return template.HTML(elm.String())
	}

	switch {
	case sizeAvi > 0:
		elm.WriteString(img(srcAvi, alt, integrityAvi))
	case sizeWeb > 0:
		elm.WriteString(img(srcWeb, alt, integrityWeb))
	case sizeJpg > 0:
		elm.WriteString(img(srcJpg, alt, integrityJpg))
	case sizePng > 0:
		elm.WriteString(img(srcPng, alt, integrityPng))
	default:
		// no screenshot
	}
	return template.HTML(elm.String())
}

// StatHumanize returns the last modified date, size in bytes and size formatted
// of the named file.
// If the file does not exist, the string "file not found" is returned.
//
// An example of the returned values are:
//
//	"2024-Sep-03", "4,163", "4.2 kB"
func StatHumanize(name string) (date string, bytes string, size string) {
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
	img := func(src, class string) string {
		alt := strings.ToLower(desc) + " thumbnail"
		return `<img src="` + src + `" loading="lazy" alt="` + alt +
			`" class="` + class + `" style="max-height:400px;" />`
	}

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

	class := "card-img-bottom" // m-2 img-fluid rounded mx-auto d-block"
	if !bottom {
		class = "card-img-top"
	}

	srcsetW := strings.Join([]string{config.StaticThumb(), unid + webp}, "/")
	srcsetP := strings.Join([]string{config.StaticThumb(), unid + png}, "/")
	if w && p {
		source := `<source srcset="` + srcsetW + `" type="image/webp" />` + img(srcsetP, class)
		return template.HTML(`<picture class="` + class + `">` + source + `</picture>`)
	}

	if p {
		return template.HTML(img(srcsetP, class))
	}

	return template.HTML(img(srcsetW, class))
}

// ThumbSample returns a HTML image tag for the given unid.
// The unid is the filename of the thumbnail image without an extension.
// The thumbDir is the directory where the thumbnail images are stored.
func ThumbSample(unid string, thumbnail dir.Directory) template.HTML {
	ext, name, src := "", "", ""

	exts := [...]string{avif, webp, png}
	for _, ext = range exts {
		name = thumbnail.Join(unid + ext)
		src = strings.Join([]string{config.StaticThumb(), unid + ext}, "/")
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
	return template.HTML(fmt.Sprintf(format, src, integrity, class, ext, integrity))
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

	const justnow = "less than a minute"
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
