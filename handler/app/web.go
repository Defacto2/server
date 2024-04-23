package app

// Package file web.go contains the custom template functions for the web framework.
// The functions are used by the HTML templates to format data.

import (
	"embed"
	"fmt"
	"html/template"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/releaser/initialism"
	"github.com/Defacto2/releaser/name"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/demozoo"
	"github.com/Defacto2/server/internal/helper"
	"github.com/volatiletech/null/v8"
)

// Web is the configuration and status of the web app.
// Rename to app or template?
type Web struct {
	Brand       *[]byte        // Brand contains to the Defacto2 ASCII logo.
	Environment *config.Config // Environment configurations from the host system environment.
	Public      embed.FS       // Public facing files.
	View        embed.FS       // Views are Go templates.
	Subresource SRI            // SRI are the Subresource Integrity hashes for the layout.
	Version     string         // Version is the current version of the app.
}

// DemozooGetLink returns a HTML link to the Demozoo download links.
func DemozooGetLink(filename, filesize, demozoo, uuid any) template.HTML {
	if val, ok := filename.(null.String); ok {
		if val.Valid && val.String != "" {
			return ""
		}
	}
	if val, ok := filesize.(null.Int64); ok {
		if val.Valid && val.Int64 > 0 {
			return ""
		}
	}
	var zooID int64
	if val, ok := demozoo.(null.Int64); ok {
		if !val.Valid || val.Int64 == 0 {
			return ""
		}
		zooID = val.Int64
	}
	var uID string
	if val, ok := uuid.(null.String); ok {
		if val.Valid && val.String == "" {
			return ""
		}
		uID = val.String
	}
	s := fmt.Sprintf("<button type=\"button\" class=\"btn btn-outline-primary me-2\" name=\"editorGetDemozoo\" "+
		"data-id=\"%d\" data-uid=\"%s\" id=btn\"%s\">GET from Demozoo</button>", zooID, uID, uID)
	return template.HTML(s)
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
			return "(n/a)"
		}
		s = artifactByteCount(val.Int64)
	default:
		return template.HTML(fmt.Sprintf("%sDownloadB: %s", typeErr, reflect.TypeOf(i).String()))
	}
	elm := fmt.Sprintf(" <small class=\"text-body-secondary\">%s</small>", s)
	return template.HTML(elm)
}

// ImageSample returns a HTML image tag for the given uuid.
func (web Web) ImageSample(uuid string) template.HTML {
	ext, name, src := "", "", ""
	for _, ext = range []string{webp, png} {
		name = filepath.Join(web.Environment.PreviewDir, uuid+ext)
		src = strings.Join([]string{config.StaticOriginal(), uuid + ext}, "/")
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

// Screenshot returns a picture elment with screenshots for the given uuid.
// The uuid is the filename of the screenshot image without an extension.
// The desc is the description of the image used for the alt attribute in the img tag.
// Supported formats are webp, png, jpg and avif.
func (web Web) Screenshot(uuid, desc string) template.HTML {
	const separator = "/"
	class := "rounded mx-auto d-block img-fluid"
	alt := strings.ToLower(desc) + " screenshot"

	srcW := strings.Join([]string{config.StaticOriginal(), uuid + webp}, separator)
	srcP := strings.Join([]string{config.StaticOriginal(), uuid + png}, separator)
	srcJ := strings.Join([]string{config.StaticOriginal(), uuid + jpg}, separator)
	srcA := strings.Join([]string{config.StaticOriginal(), uuid + avif}, separator)

	sizeA := helper.Size(filepath.Join(web.Environment.PreviewDir, uuid+avif))
	sizeJ := helper.Size(filepath.Join(web.Environment.PreviewDir, uuid+jpg))
	sizeP := helper.Size(filepath.Join(web.Environment.PreviewDir, uuid+png))
	sizeW := helper.Size(filepath.Join(web.Environment.PreviewDir, uuid+webp))

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

// TemplateFuncMap returns a map of all the template functions.
func (web Web) TemplateFuncMap() template.FuncMap {
	funcMap := web.TemplateFuncs()
	for k, v := range web.TemplateClosures() {
		funcMap[k] = v
	}
	for k, v := range web.TemplateElms() {
		funcMap[k] = v
	}
	return funcMap
}

// TemplateElms returns a map of functions that return HTML elements.
func (web Web) TemplateElms() template.FuncMap {
	const input = "<input class=\"form-check-input\""
	return template.FuncMap{
		"az": func() template.HTML {
			return template.HTML(`<small><small class="fw-lighter">A-Z</small></small>`)
		},
		"year": func() template.HTML {
			return template.HTML(`<small><small class="fw-lighter">YEARS</small></small>`)
		},
		"mergeIcon": func() template.HTML {
			return template.HTML(`<svg class="bi" aria-hidden="true" fill="currentColor">` +
				`<use xlink:href="/bootstrap-icons.svg#forward"></use></svg>`)
		},
		"msdos": func() template.HTML {
			return template.HTML(`<span class="text-nowrap">MS Dos</span>`)
		},
		"recordLastMod": func(b bool) template.HTML {
			if b {
				// tooltips do not work on disabled buttons
				return template.HTML("<button id=\"recordLMBtn\" class=\"btn btn-outline-secondary\" type=\"button\" " +
					"data-bs-toggle=\"tooltip\" data-bs-title=\"No last modification date found\" disabled>")
			}
			return template.HTML("<button id=\"recordLMBtn\" class=\"btn btn-outline-secondary\" type=\"button\" " +
				"data-bs-toggle=\"tooltip\" data-bs-title=\"Apply the file last modified date\">")
		},
		"recordOnline": func(b bool) template.HTML {
			if b {
				return template.HTML(input +
					" name=\"online\" type=\"checkbox\" role=\"switch\" id=\"recordOnline\" checked>")
			}
			return template.HTML((input +
				" name=\"online\" type=\"checkbox\" role=\"switch\" id=\"recordOnline\">"))
		},
		"recordReadme": func(b bool) template.HTML {
			if b {
				return template.HTML(input +
					" name=\"hide-readme\" type=\"checkbox\" role=\"switch\" id=\"edHideMe\" checked>")
			}
			return template.HTML((input +
				" name=\"hide-readme\" type=\"checkbox\" role=\"switch\" id=\"edHideMe\">"))
		},
	}
}

// TemplateClosures returns a map of closures that return converted type or modified strings.
func (web Web) TemplateClosures() template.FuncMap {
	hrefs := Hrefs()
	return template.FuncMap{
		"bootstrap5": func() string {
			return hrefs[Bootstrap5]
		},
		"bootstrap5JS": func() string {
			return hrefs[Bootstrap5JS]
		},
		"demozooSanity": func() string {
			return strconv.Itoa(demozoo.Sanity)
		},
		"editArchive": func() string {
			return hrefs[EditArchive]
		},
		"editAssets": func() string {
			return hrefs[EditAssets]
		},
		"editForApproval": func() string {
			return hrefs[EditForApproval]
		},
		"editor": func() string {
			return hrefs[Editor]
		},
		"exampleDay": func() string {
			return time.Now().Format("2")
		},
		"exampleMonth": func() string {
			return time.Now().Format("1")
		},
		"exampleYear": func() string {
			return time.Now().Format("2006")
		},
		"fmtName": func(s string) string {
			return helper.Capitalize(strings.ToLower(s))
		},
		"fmtRangeURI": func(s string) string {
			x, err := name.Humanize(name.Path(s))
			if err != nil {
				return err.Error()
			}
			return helper.Titleize(x)
		},
		"fa5Pro": func() string {
			return hrefs[FA5Pro]
		},
		"htmx": func() string {
			return hrefs[Htmx]
		},
		"initialisms": func(s string) string {
			return initialism.Join(initialism.Path(s))
		},
		"jsdos6JS": func() string {
			return hrefs[Jsdos6JS]
		},
		"dosboxJS": func() string {
			return hrefs[DosboxJS]
		},
		"layout": func() string {
			return hrefs[Layout]
		},
		"layoutJS": func() string {
			return hrefs[LayoutJS]
		},
		"logo": func() string {
			return string(*web.Brand)
		},
		"pouet": func() string {
			return hrefs[Pouet]
		},
		"readme": func() string {
			return hrefs[Readme]
		},
		"sri_bootstrap5": func() string {
			return web.Subresource.Bootstrap5
		},
		"sri_bootstrap5JS": func() string {
			return web.Subresource.Bootstrap5JS
		},
		"sri_editArchive": func() string {
			return web.Subresource.EditArchive
		},
		"sri_editAssets": func() string {
			return web.Subresource.EditAssets
		},
		"sri_editForApproval": func() string {
			return web.Subresource.EditForApproval
		},
		"sri_editor": func() string {
			return web.Subresource.Editor
		},
		"sri_fa5Pro": func() string {
			return web.Subresource.FA5Pro
		},
		"sri_htmx": func() string {
			return web.Subresource.Htmx
		},
		"sri_jsdos6JS": func() string {
			return web.Subresource.Jsdos6JS
		},
		"sri_dosboxJS": func() string {
			return web.Subresource.DosboxJS
		},
		"sri_layout": func() string {
			return web.Subresource.Layout
		},
		"sri_layoutJS": func() string {
			return web.Subresource.LayoutJS
		},
		"sri_pouet": func() string {
			return web.Subresource.Pouet
		},
		"sri_readme": func() string {
			return web.Subresource.Readme
		},
		"sri_uploader": func() string {
			return web.Subresource.Uploader
		},
		"tagSel": TagSel,
		"uploader": func() string {
			return hrefs[Uploader]
		},
		"version": func() string {
			return web.Version
		},
	}
}

// TemplateFuncs are a collection of mapped functions that can be used in a template.
//
// The "fmtURI" function is not performant for large lists,
// instead use "fmtRangeURI" in TemplateStrings().
func (web Web) TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"add":               helper.Add1,
		"attribute":         Attribute,
		"brief":             Brief,
		"describe":          Describe,
		"downloadB":         DownloadB,
		"byteFile":          ByteFile,
		"byteFileS":         ByteFileS,
		"demozooGetLink":    DemozooGetLink,
		"fmtDay":            Day,
		"fmtMonth":          Month,
		"fmtPrefix":         Prefix,
		"fmtRoles":          helper.FmtSlice,
		"fmtURI":            releaser.Link,
		"lastUpdated":       LastUpdated,
		"linkDownload":      LinkDownload,
		"linkHref":          LinkHref,
		"linkInterview":     LinkInterview,
		"linkPage":          LinkPage,
		"linkPreview":       LinkPreview,
		"linkRemote":        LinkRemote,
		"linkRelrs":         LinkRelFast,
		"linkScnr":          LinkScnr,
		"linkSVG":           LinkSVG,
		"linkWiki":          LinkWiki,
		"logoText":          LogoText,
		"mimeMagic":         MimeMagic,
		"recordImgSample":   web.ImageSample,
		"recordThumbSample": web.ThumbSample,
		"recordInfoOSTag":   TagWithOS,
		"recordTagInfo":     TagBrief,
		"safeHTML":          SafeHTML,
		"safeJS":            SafeJS,
		"screenshot":        web.Screenshot,
		"slugify":           helper.Slug,
		"subTitle":          SubTitle,
		"thumb":             web.Thumb,
		"trimSiteSuffix":    TrimSiteSuffix,
		"trimSpace":         TrimSpace,
		"websiteIcon":       WebsiteIcon,
	}
}

// Templates returns a map of the templates used by the route.
func (web *Web) Templates() (map[string]*template.Template, error) {
	if err := web.Subresource.Verify(web.Public); err != nil {
		return nil, err
	}
	tmpls := make(map[string]*template.Template)
	for k, name := range templates() {
		tmpl := web.tmpl(name)
		tmpls[k] = tmpl
	}
	return tmpls, nil
}

// Thumb returns a HTML image tag or picture element for the given uuid.
// The uuid is the filename of the thumbnail image without an extension.
// The desc is the description of the image.
func (web Web) Thumb(uuid, desc string, bottom bool) template.HTML {
	fw := filepath.Join(web.Environment.ThumbnailDir, uuid+webp)
	fp := filepath.Join(web.Environment.ThumbnailDir, uuid+png)
	webp := strings.Join([]string{config.StaticThumb(), uuid + webp}, "/")
	png := strings.Join([]string{config.StaticThumb(), uuid + png}, "/")
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
	if w {
		return img(webp, alt, class, style)
	}
	if p {
		return img(png, alt, class, style)
	}
	return ""
}

// ThumbSample returns a HTML image tag for the given uuid.
func (web Web) ThumbSample(uuid string) template.HTML {
	const (
		png  = png
		webp = webp
	)
	ext, name, src := "", "", ""
	for _, ext = range []string{webp, png} {
		name = filepath.Join(web.Environment.ThumbnailDir, uuid+ext)
		src = strings.Join([]string{config.StaticThumb(), uuid + ext}, "/")
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

// Web tmpl returns a layout template for the given named view.
// Note that the name is relative to the view/defaults directory.
func (web Web) tmpl(name filename) *template.Template {
	files := []string{
		GlobTo("layout.tmpl"),
		GlobTo("modal.tmpl"),
		GlobTo("option_os.tmpl"),
		GlobTo("option_tag.tmpl"),
		GlobTo(string(name)),
		GlobTo("pagination.tmpl"),
	}
	config := web.Environment
	files = lockTmpls(config.ReadMode, files...)
	// append any additional and embedded templates
	switch name {
	case "artifact.tmpl":
		files = artifactTmpls(config.ReadMode, files...)
	case "file.tmpl":
		files = append(files, GlobTo("file_expand.tmpl"))
	case "websites.tmpl":
		const individualWebsite = "website.tmpl"
		files = append(files, GlobTo(individualWebsite))
	}
	return template.Must(template.New("").Funcs(
		web.TemplateFuncMap()).ParseFS(web.View, files...))
}

type filename string // filename is the name of the template file in the view directory.

func artifactTmpls(lock bool, files ...string) []string {
	files = append(files,
		GlobTo("artifact_table.tmpl"),
		GlobTo("artifact_jsdos6.tmpl"),
		GlobTo("artifact_editor_archive.tmpl"))
	if lock {
		return append(files,
			GlobTo("artifact_editor_null.tmpl"),
			GlobTo("artifact_editor_table_null.tmpl"),
			GlobTo("artifact_table_switch_null.tmpl"))
	}
	return append(files,
		GlobTo("artifact_editor.tmpl"),
		GlobTo("artifact_editorHtmx.tmpl"),
		GlobTo("artifact_editor_table.tmpl"),
		GlobTo("artifact_table_switch.tmpl"))
}

// img returns a HTML image tag.
func img(src, alt, class, style string) template.HTML {
	return template.HTML(fmt.Sprintf("<img src=\"%s\" loading=\"lazy\" alt=\"%s\" class=\"%s\" style=\"%s\" />",
		src, alt, class, style))
}

func templates() map[string]filename {
	const releaser, scener = "releaser.tmpl", "scener.tmpl"
	return map[string]filename{
		"index":         "index.tmpl",
		"artifact":      "artifact.tmpl",
		"bbs":           releaser,
		"bbs-year":      "releaser_year.tmpl",
		"coder":         scener,
		"file":          "file.tmpl",
		"files":         "files.tmpl",
		"ftp":           releaser,
		"history":       "history.tmpl",
		"interview":     "interview.tmpl",
		"magazine":      "releaser_year.tmpl",
		"magazine-az":   releaser,
		"reader":        "reader.tmpl",
		"releaser":      releaser,
		"releaser-year": "releaser_year.tmpl",
		"scener":        scener,
		"searchHtmx":    "searchHtmx.tmpl",
		"searchList":    "searchList.tmpl",
		"searchPost":    "searchPost.tmpl",
		"signin":        "signin.tmpl",
		"signout":       "signout.tmpl",
		"status":        "status.tmpl",
		"thanks":        "thanks.tmpl",
		"thescene":      "the_scene.tmpl",
		"websites":      "websites.tmpl",
	}
}

func lockTmpls(lock bool, files ...string) []string {
	if lock {
		return append(files,
			GlobTo("layout_editor_null.tmpl"),
			GlobTo("layout_editorJS_null.tmpl"),
			GlobTo("layout_uploader_null.tmpl"),
			GlobTo("layout_uploaderJS_null.tmpl"),
			GlobTo("uploader_null.tmpl"))
	}
	return append(files,
		GlobTo("layout_editor.tmpl"),
		GlobTo("layout_editorJS.tmpl"),
		GlobTo("layout_uploader.tmpl"),
		GlobTo("layout_uploaderJS.tmpl"),
		GlobTo("uploader.tmpl"),
		GlobTo("uploaderHtmx.tmpl"))
}
