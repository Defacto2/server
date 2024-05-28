package app

// Package file template.go contains the template functions for the application.

import (
	"embed"
	"fmt"
	"html"
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
	"github.com/Defacto2/server/internal/form"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/pouet"
	"github.com/volatiletech/null/v8"
)

const closeAnchor = "</a>"

// Templ is the configuration and status of the web application templates.
type Templ struct {
	Brand       []byte        // Brand contains to the Defacto2 ASCII logo.
	Environment config.Config // Environment configurations from the host system environment.
	Public      embed.FS      // Public facing files.
	View        embed.FS      // Views are Go templates.
	RecordCount int           // RecordCount is the total number of records in the database.
	Subresource SRI           // SRI are the Subresource Integrity hashes for the layout.
	Version     string        // Version is the current version of the app.
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
	var uID string
	if val, valExists := unid.(null.String); valExists {
		if val.Valid && val.String == "" {
			return ""
		}
		uID = val.String
	}
	if uID == "" || zooID == 0 {
		return "no id provided"
	}
	s := fmt.Sprintf("<button type=\"button\" "+
		"class=\"btn btn-outline-primary me-2\" name=\"editorGetDemozoo\" "+
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

// LinkSample returns a collection of HTML formatted links for the artifact editor.
func LinkSamples(youtube, demozoo, pouet, colors16, github, rels, sites string) string {
	links := LinkPreviews(youtube, demozoo, pouet, colors16, github, rels, sites)
	for i, link := range links {
		links[i] = html.EscapeString(link)
	}
	return strings.Join(links, "<br>")
}

// LinkPreviews returns a slice of HTML formatted links for the artifact editor.
func LinkPreviews(youtube, demozoo, pouet, colors16, github, rels, sites string) []string {
	rel := func(url string) string {
		return `<a href="https://` + url + `">` + url + closeAnchor
	}

	links := []string{}
	if youtube != "" {
		links = append(links, rel("youtube.com/watch?v="+youtube))
	}
	if demozoo != "" {
		links = append(links, rel("demozoo.org/productions/"+demozoo))
	}
	if pouet != "" {
		links = append(links, rel("pouet.net/prod.php?which="+pouet))
	}
	if colors16 != "" {
		links = append(links, rel("16colo.rs/"+colors16))
	}
	if github != "" {
		links = append(links, rel("github.com/"+github))
	}
	if rels != "" {
		links = append(links, string(LinkRelations(rels)))
	}
	if sites != "" {
		links = append(links, string(LinkSites(sites)))
	}
	return links
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

// ImageSample returns a HTML image tag for the given unid.
func (web Templ) ImageSample(unid string) template.HTML {
	ext, name, src := "", "", ""
	for _, ext = range []string{webp, png} {
		name = filepath.Join(web.Environment.PreviewDir, unid+ext)
		src = strings.Join([]string{config.StaticOriginal(), unid + ext}, "/")
		if helper.Stat(name) {
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

// Screenshot returns a picture elment with screenshots for the given unid.
// The unid is the filename of the screenshot image without an extension.
// The desc is the description of the image used for the alt attribute in the img tag.
// Supported formats are webp, png, jpg and avif.
func (web Templ) Screenshot(unid, desc string) template.HTML {
	const separator = "/"
	class := "rounded mx-auto d-block img-fluid"
	alt := strings.ToLower(desc) + " screenshot"

	srcW := strings.Join([]string{config.StaticOriginal(), unid + webp}, separator)
	srcP := strings.Join([]string{config.StaticOriginal(), unid + png}, separator)
	srcJ := strings.Join([]string{config.StaticOriginal(), unid + jpg}, separator)
	srcA := strings.Join([]string{config.StaticOriginal(), unid + avif}, separator)

	sizeA := helper.Size(filepath.Join(web.Environment.PreviewDir, unid+avif))
	sizeJ := helper.Size(filepath.Join(web.Environment.PreviewDir, unid+jpg))
	sizeP := helper.Size(filepath.Join(web.Environment.PreviewDir, unid+png))
	sizeW := helper.Size(filepath.Join(web.Environment.PreviewDir, unid+webp))

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
func (web Templ) TemplateFuncMap() template.FuncMap {
	funcMap := web.TemplateFuncs()
	for k, v := range web.TemplateClosures() {
		funcMap[k] = v
	}
	for k, v := range web.TemplateElms() {
		funcMap[k] = v
	}
	return funcMap
}

const (
	input = "<input class=\"form-check-input\""
	radio = `<input type="radio" class="btn-check" name="artifact-editor-record"`
)

// TemplateElms returns a map of functions that return HTML elements.
func (web Templ) TemplateElms() template.FuncMap {
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
		"recordLastMod": recordLastMod,
		"radioPublic":   radioPublic,
		"radioHidden":   radioHidden,
		"recordOnline":  recordOnline,
		"recordReadme":  recordReadme,
	}
}

func recordLastMod(b bool) template.HTML {
	if b { // tooltips do not work on disabled buttons
		return template.HTML("<button id=\"recordLMBtn\" class=\"btn btn-outline-secondary\" type=\"button\" " +
			"data-bs-toggle=\"tooltip\" data-bs-title=\"No last modification date found\" disabled>")
	}
	return template.HTML("<button id=\"recordLMBtn\" class=\"btn btn-outline-secondary\" type=\"button\" " +
		"data-bs-toggle=\"tooltip\" data-bs-title=\"Apply the file last modified date\">")
}

func radioPublic(b bool) template.HTML {
	const htmx = ` hx-post="/editor/online/true"
	hx-include="[name='artifact-editor-key']"`
	if b {
		return template.HTML(radio +
			htmx + ` id="artifact-editor-public" autocomplete="off" checked>`)
	}
	return template.HTML(radio +
		htmx + ` id="artifact-editor-public" autocomplete="off">`)
}

func radioHidden(b bool) template.HTML {
	const htmx = ` hx-post="/editor/online/false"
	hx-include="[name='artifact-editor-key']"`
	if !b {
		return template.HTML(radio +
			htmx + ` id="artifact-editor-hidden" autocomplete="off" checked>`)
	}
	return template.HTML(radio +
		htmx + ` id="artifact-editor-hidden" autocomplete="off">`)
}

func recordOnline(b bool) template.HTML {
	if b {
		return template.HTML(input +
			" name=\"online\" type=\"checkbox\" role=\"switch\" id=\"recordOnline\" checked>")
	}
	return template.HTML((input +
		" name=\"online\" type=\"checkbox\" role=\"switch\" id=\"recordOnline\">"))
}

func recordReadme(b bool) template.HTML {
	if b {
		return template.HTML(input +
			" name=\"hide-readme\" type=\"checkbox\" role=\"switch\" id=\"edHideMe\" checked>")
	}
	return template.HTML((input +
		" name=\"hide-readme\" type=\"checkbox\" role=\"switch\" id=\"edHideMe\">"))
}

// TemplateClosures returns a map of closures that return converted type or modified strings.
func (web Templ) TemplateClosures() template.FuncMap { //nolint:funlen
	hrefs := Hrefs()
	return template.FuncMap{
		"artifactEditor": func() string {
			return hrefs[ArtifactEditor]
		},
		"bootstrap5": func() string {
			return hrefs[Bootstrap5]
		},
		"bootstrap5JS": func() string {
			return hrefs[Bootstrap5JS]
		},
		"classification": func(s, p string) string {
			count, _ := form.HumanizeAndCount(s, p)
			return string(count)
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
			return string(web.Brand)
		},
		"pouet": func() string {
			return hrefs[Pouet]
		},
		"pouetSanity": func() string {
			return strconv.Itoa(pouet.Sanity)
		},
		"readme": func() string {
			return hrefs[Readme]
		},
		"sri_artifactEditor": func() string {
			return web.Subresource.ArtifactEditor
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
		"tagOption": TagOption,
		"uploader": func() string {
			return hrefs[Uploader]
		},
		"titleize": helper.Titleize,
		"version": func() string {
			return web.Version
		},
	}
}

// TemplateFuncs are a collection of mapped functions that can be used in a template.
//
// The "fmtURI" function is not performant for large lists,
// instead use "fmtRangeURI" in TemplateStrings().
func (web Templ) TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"add":                helper.Add1,
		"attribute":          Attribute,
		"brief":              Brief,
		"describe":           Describe,
		"downloadB":          DownloadB,
		"byteFile":           ByteFile,
		"byteFileS":          ByteFileS,
		"demozooGetLink":     DemozooGetLink,
		"fmtDay":             Day,
		"fmtMonth":           Month,
		"fmtPrefix":          Prefix,
		"fmtRoles":           helper.FmtSlice,
		"fmtURI":             releaser.Link,
		"lastUpdated":        LastUpdated,
		"linkDownload":       LinkDownload,
		"linkHref":           LinkHref,
		"linkInterview":      LinkInterview,
		"linkPage":           LinkPage,
		"linkPreview":        LinkPreview,
		"linkRemote":         LinkRemote,
		"linkRelrs":          LinkRelFast,
		"linkScnr":           LinkScnr,
		"linkSVG":            LinkSVG,
		"linkWiki":           LinkWiki,
		"logoText":           LogoText,
		"mimeMagic":          MimeMagic,
		"recordImgSample":    web.ImageSample,
		"recordThumbSample":  web.ThumbSample,
		"recordInfoOSTag":    TagWithOS,
		"recordLinkPreviews": LinkSamples,
		"recordTagInfo":      TagBrief,
		"safeHTML":           SafeHTML,
		"safeJS":             SafeJS,
		"screenshot":         web.Screenshot,
		"slugify":            helper.Slug,
		"subTitle":           SubTitle,
		"thumb":              web.Thumb,
		"trimSiteSuffix":     TrimSiteSuffix,
		"trimSpace":          TrimSpace,
		"websiteIcon":        WebsiteIcon,
	}
}

// Templates returns a map of the templates used by the route.
func (web *Templ) Templates() (map[string]*template.Template, error) {
	if err := web.Subresource.Verify(web.Public); err != nil {
		return nil, fmt.Errorf("web.Subresource.Verify: %w", err)
	}
	tmpls := make(map[string]*template.Template)
	for k, name := range templates() {
		tmpl := web.tmpl(name)
		tmpls[k] = tmpl
	}
	return tmpls, nil
}

// Thumb returns a HTML image tag or picture element for the given unid.
// The unid is the filename of the thumbnail image without an extension.
// The desc is the description of the image.
func (web Templ) Thumb(unid, desc string, bottom bool) template.HTML {
	fw := filepath.Join(web.Environment.ThumbnailDir, unid+webp)
	fp := filepath.Join(web.Environment.ThumbnailDir, unid+png)
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

// ThumbSample returns a HTML image tag for the given unid.
func (web Templ) ThumbSample(unid string) template.HTML {
	const (
		png  = png
		webp = webp
	)
	ext, name, src := "", "", ""
	for _, ext = range []string{webp, png} {
		name = filepath.Join(web.Environment.ThumbnailDir, unid+ext)
		src = strings.Join([]string{config.StaticThumb(), unid + ext}, "/")
		if helper.Stat(name) {
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
func (web Templ) tmpl(name filename) *template.Template {
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
	offline := web.RecordCount < 1
	files = dbTmpls(config.ReadMode, offline, files...)
	// append any additional and embedded templates
	switch name {
	case "artifact.tmpl":
		files = artifactTmpls(config.ReadMode, files...)
	case "categories.tmpl":
		files = append(files, GlobTo("categoriesmore.tmpl"))
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
		GlobTo("artifactinfo.tmpl"),
		GlobTo("artifactjsdos.tmpl"),
		GlobTo("artifactzip.tmpl"))
	if lock {
		return append(files,
			GlobTo("artifactedit_null.tmpl"),
			GlobTo("artifactlock_null.tmpl"))
	}
	return append(files,
		GlobTo("artifactfile.tmpl"),
		GlobTo("artifactedit.tmpl"),
		GlobTo("artifactfooter.tmpl"),
		GlobTo("artifactlock.tmpl"))
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
		"artifacts":     "artifacts.tmpl",
		"bbs":           releaser,
		"bbs-year":      "releaseryear.tmpl",
		"categories":    "categories.tmpl",
		"coder":         scener,
		"ftp":           releaser,
		"history":       "history.tmpl",
		"interview":     "interview.tmpl",
		"magazine":      "releaseryear.tmpl",
		"magazine-az":   releaser,
		"reader":        "reader_todo.tmpl",
		"releaser":      releaser,
		"releaser-year": "releaseryear.tmpl",
		"scener":        scener,
		"searchhtmx":    "searchhtmx.tmpl",
		"searchpost":    "searchpost.tmpl",
		"signin":        "signin.tmpl",
		"signout":       "signout.tmpl",
		"status":        "status.tmpl",
		"thanks":        "thanks.tmpl",
		"thescene":      "thescene.tmpl",
		"websites":      "websites.tmpl",
	}
}

func dbTmpls(lock, offline bool, files ...string) []string {
	if offline || lock {
		return append(files,
			GlobTo("layoutup_null.tmpl"),
			GlobTo("layoutjsup_null.tmpl"),
			GlobTo("uploader_null.tmpl"),
		)
	}
	return append(files,
		GlobTo("layoutup.tmpl"),
		GlobTo("layoutjsup.tmpl"),
		GlobTo("uploader.tmpl"),
		GlobTo("uploaderhtmx.tmpl"),
	)
}

func lockTmpls(lock bool, files ...string) []string {
	if lock {
		return append(files,
			GlobTo("layoutlock_null.tmpl"),
			GlobTo("layoutjs_null.tmpl"),
		)
	}
	return append(files,
		GlobTo("layoutlock.tmpl"),
		GlobTo("layoutjs.tmpl"),
	)
}
