// Package simple provides functions for handling string or integer input data.
package simple

import (
	"errors"
	"fmt"
	"html/template"
	"image"
	_ "image/gif"  // gif format decoder
	_ "image/jpeg" // jpeg format decoder
	_ "image/png"  // png format decoder
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/extensions"
	"github.com/Defacto2/server/internal/tags"
	"github.com/dustin/go-humanize"
	"github.com/h2non/filetype"
	"github.com/volatiletech/null/v8"
	_ "golang.org/x/image/webp" // webp format decoder
)

var (
	ErrLinkType = errors.New("the id value is an invalid type")
	ErrName     = errors.New("name is an empty string")
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

// AssetSrc returns a URL to an artifact asset with an cache busting hash.
// The named dir is the directory where the asset is stored,
// the unid is the unique identifier of the asset,
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
	return humanize.Bytes(uint64(math.Abs(float64(i))))
}

// DemozooGetLink returns a HTML link to the Demozoo download links.
// The filename and filesize are used to determine if the file exists.
// The demozoo is the ID for the production on Demozoo.
// The unid is the unique identifier for the file record.
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
	return template.HTML(`clone the demozoo assets`)
}

// DownloadB returns a human readable string of the file size.
// The i value must be an integer or a null.Int64.
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
// The preview is the directory where the preview images are stored.
func ImageSample(unid string, preview dir.Directory) template.HTML {
	ext, name, src := "", "", ""
	exts := []string{avif, webp, png}
	for ext = range slices.Values(exts) {
		name = preview.Join(unid + ext)
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
// The preview is the directory where the preview images are stored.
func ImageSampleStat(unid string, preview dir.Directory) bool {
	name := ""
	exts := []string{avif, webp, png}
	for ext := range slices.Values(exts) {
		name = preview.Join(unid + ext)
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
func ImageXY(name string) [2]string {
	zero := [2]string{"0", ""}
	switch filepath.Ext(strings.ToLower(name)) {
	case ".jpg", ".jpeg", ".gif", ".png", ".webp":
	default:
		st, err := os.Stat(name)
		// open /mnt/volume_sfo3_01/assets/images000/ca6cf279-3758-4e1e-8e8b-f60871e877be.jpg: no such file or directoryB
		if errors.Is(err, os.ErrNotExist) {
			return zero
		} else if err != nil {
			return [2]string{err.Error(), ""}
		}
		return [2]string{humanize.Comma(st.Size()), ""}
	}
	reader, err := os.Open(name)
	if errors.Is(err, os.ErrNotExist) {
		return zero
	} else if err != nil {
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
	return [2]string{
		humanize.Comma(st.Size()),
		fmt.Sprintf("%dx%d", config.Width, config.Height),
	}
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
			ref = fmt.Sprintf("%s ❌ link /f/%s is an invalid download path.", ref, id)
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
	sign := magicnumber.Find(r)
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
//
// The returned path should be removed after use.
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
func Releasers(prime, second string, magazine bool) template.HTML {
	var s string
	switch {
	case magazine && prime != "" && second != "":
		s = fmt.Sprintf("%s <small>published by</small> %s", second, prime)
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

// Screenshot returns a image elment with screenshots for the given unid.
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

	srcW := strings.Join([]string{config.StaticOriginal(), unid + webp}, separator)
	srcP := strings.Join([]string{config.StaticOriginal(), unid + png}, separator)
	srcJ := strings.Join([]string{config.StaticOriginal(), unid + jpg}, separator)
	srcA := strings.Join([]string{config.StaticOriginal(), unid + avif}, separator)

	sizeW := helper.Size(preview.Join(unid + webp))
	sizeP := helper.Size(preview.Join(unid + png))
	sizeJ := helper.Size(preview.Join(unid + jpg))
	sizeA := helper.Size(preview.Join(unid + avif))

	hashW, _ := helper.IntegrityFile(preview.Join(unid + webp))
	hashP, _ := helper.IntegrityFile(preview.Join(unid + png))
	hashJ, _ := helper.IntegrityFile(preview.Join(unid + jpg))
	hashA, _ := helper.IntegrityFile(preview.Join(unid + avif))

	usePicture := (sizeA > 0 || sizeW > 0) && (sizeJ > 0 || sizeP > 0)
	if usePicture {
		elm := template.HTML("<picture>")
		switch {
		case sizeA > 0:
			elm += template.HTML(fmt.Sprintf("<source srcset=\"%s?%s\""+
				" type=\"image/avif\" integrity=\"%s\" />", srcA, hashA, hashA))
		case sizeW > 0:
			elm += template.HTML(fmt.Sprintf("<source srcset=\"%s?%s\""+
				" type=\"image/webp\" integrity=\"%s\" />", srcW, hashW, hashW))
		}
		// the <picture> element is used to provide multiple sources for an image.
		// if no <img> element is provided, the <picture> element won't be rendered by the browser.
		useSmallerJpg := sizeJ > 0 && sizeJ < sizeP
		switch {
		case useSmallerJpg:
			elm += img(srcJ, alt, hashJ)
		case sizeP > 0:
			elm += img(srcP, alt, hashP)
		default:
			elm += img(srcJ, alt, hashJ)
		}
		return elm + "</picture>"
	}
	switch {
	case sizeA > 0:
		return img(srcA, alt, hashA)
	case sizeW > 0:
		return img(srcW, alt, hashW)
	case sizeJ > 0:
		return img(srcJ, alt, hashJ)
	case sizeP > 0:
		return img(srcP, alt, hashP)
	}
	return ""
}

// img returns a HTML image tag.
func img(src, alt, integrity string) template.HTML {
	return template.HTML(fmt.Sprintf("<img src=\"%s?%s\" loading=\"lazy\" alt=\"%s\""+
		" class=\"rounded mx-auto d-block img-fluid\" integrity=\"%s\" />",
		src, integrity, alt, integrity))
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
	fw := thumbnail.Join(unid + webp)
	fp := thumbnail.Join(unid + png)
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
	if !w && !p {
		return template.HTML("<!-- no thumbnail found -->")
	}
	const style = "max-height:400px;"
	class := "m-2 img-fluid rounded mx-auto d-block"
	if !bottom {
		class = "card-img-top"
	}
	if w && p {
		elm := "<picture class=\"" + class + "\">" +
			fmt.Sprintf("<source srcset=\"%s\" type=\"image/webp\" />", webp) +
			string(thumb(png, alt, class, style)) +
			"</picture>"
		return template.HTML(elm)
	}
	src := webp
	if p {
		src = png
	}
	return thumb(src, alt, class, style)
}

// img returns a HTML image tag.
func thumb(src, alt, class, style string) template.HTML {
	return template.HTML(fmt.Sprintf("<img src=\"%s\" loading=\"lazy\" alt=\"%s\" class=\"%s\" style=\"%s\" />",
		src, alt, class, style))
}

// ThumbSample returns a HTML image tag for the given unid.
// The unid is the filename of the thumbnail image without an extension.
// The thumbDir is the directory where the thumbnail images are stored.
func ThumbSample(unid string, thumbnail dir.Directory) template.HTML {
	ext, name, src := "", "", ""
	exts := []string{avif, webp, png}
	for ext = range slices.Values(exts) {
		name = thumbnail.Join(unid + ext)
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
	switch val := t.(type) {
	case null.Time:
		if !val.Valid {
			return ""
		}
		x := helper.TimeDistance(val.Time, time.Now(), false)
		if x == justnow {
			return s + " just now"
		}
		return s + " " + x + " ago"
	case time.Time:
		x := helper.TimeDistance(val, time.Now(), false)
		if x == justnow {
			return s + " just now"
		}
		return s + " " + x + " ago"
	default:
		return fmt.Sprintf("%supdated: %s", typeErr, reflect.TypeOf(t).String())
	}
}
