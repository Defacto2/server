//nolint:gochecknoglobals
package app

// Package file template.go contains the template functions for the application.

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"io/fs"
	"maps"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/handler/form"
	"github.com/Defacto2/server/handler/internal/filerecord"
	"github.com/Defacto2/server/handler/internal/simple"
	"github.com/Defacto2/server/handler/pouet"
	"github.com/Defacto2/server/handler/releaser"
	"github.com/Defacto2/server/handler/releaser/lism"
	"github.com/Defacto2/server/handler/releaser/name"
	"github.com/Defacto2/server/handler/tidbit"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/tags"
	"github.com/aarondl/null/v8"
)

const (
	favicon   = "/image/layout/defacto2-ascii.png"
	linkClass = "text-nowrap link-offset-2 link-underline link-underline-opacity-25"
)

// WebApp is the configuration and status of the web application templates.
type WebApp struct {
	Public      fs.FS         // Public facing files.
	View        fs.FS         // Views are Go templates.
	Subresource SRI           // SRI are the Subresource Integrity hashes for the layout.
	Version     string        // Version is the current version of the app.
	Brand       []byte        // Brand contains to the Defacto2 ASCII logo.
	Environment config.Config // Environment configurations from the host system environment.
	RecordCount int64         // RecordCount is the total number of records in the database.
}

// Templates returns a map of the templates used by the route.
func (wa *WebApp) Templates(ctx context.Context, db *sql.DB) (map[string]*template.Template, error) {
	const format = "templates mapper %s: %w"
	if err := nils.Check(ctx, db); err != nil {
		return nil, fmt.Errorf(format, "check", err)
	}

	if err := wa.Subresource.Verify(wa.Public); err != nil {
		return nil, fmt.Errorf(format, "verify", err)
	}

	templates := make(map[string]*template.Template)
	for key, name := range wa.Pages() {
		templates[key] = wa.parseFS(ctx, db, name)
	}

	return templates, nil
}

const (
	artifactTmpl     = "artifact.tmpl"
	artifactsTmpl    = "artifacts.tmpl"
	categoriesTmpl   = "categories.tmpl"
	releaserTmpl     = "releaser.tmpl"
	releaseryearTmpl = "releaseryear.tmpl"
	scenerTmpl       = "scener.tmpl"
	websitesTmpl     = "websites.tmpl"
)

type filename string // filename is the name of the template file in the view directory.

type Page map[string]filename

var pages = Page{
	"api-info":      "apiinfo.tmpl",
	"apps":          "apps.tmpl",
	"areacodes":     "areacodes.tmpl",
	"artifact":      artifactTmpl,
	"artifacts":     artifactsTmpl,
	"bbs":           releaserTmpl,
	"bbs-year":      releaseryearTmpl,
	"brokentexts":   "brokentexts.tmpl",
	"categories":    categoriesTmpl,
	"configs":       "configurations.tmpl",
	"coder":         scenerTmpl,
	"compression":   "compression.tmpl",
	"ftp":           releaserTmpl,
	"fixers":        "fixers.tmpl",
	"fixes":         "fixes.tmpl",
	"history":       "history.tmpl",
	"index":         "index.tmpl",
	"interview":     "interview.tmpl",
	"magazine":      releaseryearTmpl,
	"magazine-az":   releaserTmpl,
	"new":           "new.tmpl",
	"releaser":      releaserTmpl,
	"releaser-year": releaseryearTmpl,
	"routes":        "routes.tmpl",
	"scener":        scenerTmpl,
	"searchhtmx":    "searchhtmx.tmpl",
	"searchpost":    "searchpost.tmpl",
	"signin":        "signin.tmpl",
	"signout":       "signout.tmpl",
	"status":        "status.tmpl",
	"terms":         "terms.tmpl",
	"thanks":        "thanks.tmpl",
	"thescene":      "thescene.tmpl",
	"titles":        "titles.tmpl",
	websites:        websitesTmpl,
}

// Pages returns a map of the template names and their corresponding filenames.
func (wa *WebApp) Pages() Page {
	// To embed a template within one of these .tmpl pages,
	// use the parseFS() func found later in this file.
	//
	// Embed file example:
	// {{- define "abc" }}<h1>Hi!</h1>{{- end}}
	//
	// Template page usage:
	// {{ template "abc" }}
	//
	// Or to pass a maximum of one value:
	// {{ template "abc" $myVar }}
	return pages
}

func (wa *WebApp) Layout(name filename) []string {
	return []string{
		GlobTo("layout.tmpl"),
		GlobTo("modal.tmpl"),
		GlobTo("option_os.tmpl"),
		GlobTo("option_tag.tmpl"),
		GlobTo(string(name)),
		GlobTo("pagination.tmpl"),
		GlobTo("opengraph.tmpl"),
	}
}

// Funcs are a collection of mapped functions that can be used in a template.
//
// The "fmtURI" function is not performant for large lists,
// instead use "fmtRangeURI" in TemplateStrings().
func (wa *WebApp) Funcs() template.FuncMap {
	return funcMap
}

var funcMap = template.FuncMap{
	"add":                helper.Add1,
	"attribute":          Attribute,
	"brief":              Brief,
	"capitalize":         helper.Capitalize,
	"describe":           Describe,
	"downloadB":          simple.DownloadInBytes,
	"byteBytes":          ByteBytes,
	"byteFile":           ByteFile,
	"byteFileS":          ByteFileS,
	"demozooGetLink":     simple.DemozooGetLink,
	"fmtDay":             Day,
	"fmtMonth":           Month,
	"fmtPrefix":          Prefix,
	"fmtRoles":           helper.FmtSlice,
	"fmtURI":             releaser.Link,
	"hasSuffix":          HasSuffix,
	"lastUpdated":        LastUpdated,
	"linkDownload":       LinkDownload,
	"linkHref":           LinkHref,
	"linkInterview":      LinkInterview,
	"linkPage":           LinkPage,
	"linkPreview":        LinkPreview,
	"linkRemote":         LinkRemote,
	"linkRemoteTip":      LinkRemoteTip,
	"linkRunApp":         LinkRunApp,
	"linkRelrs":          LinkRels,
	"linkScnr":           LinkScnr,
	"linkScnrs":          LinkScnrs,
	"linkSVG":            filerecord.LinkSVG,
	"linkWiki":           LinkWiki,
	"linkWikiTip":        LinkWikiTip,
	"logoText":           LogoText,
	"mask":               Mask,
	"musicMod":           MusicModule,
	"jsdosUsage":         filerecord.JsdosUsage,
	"recordInfoOSTag":    TagWithOS,
	"recordLinkPreviews": LinkPreviews,
	"recordTagInfo":      TagBrief,
	"safeBBS":            SafeBBS,
	"safeDocument":       SafeDocument,
	"safeHTML":           SafeHTML,
	"safeJS":             SafeJS,
	"safety":             Safety,
	"slugify":            helper.Slug,
	"stripSup":           StripSup,
	"subTitle":           SubTitle,
	"tagOption":          TagOption,
	"tidbitMissing":      tidbit.Missing,
	"toLower":            strings.ToLower,
	"trimSpace":          TrimSpace,
	"websiteIcon":        WebsiteIcon,
	"urlEncode":          URLEncode,
	"yearRange":          yearRange,
}

func (wa *WebApp) FuncDB(ctx context.Context, db *sql.DB) template.FuncMap {
	if ctx == nil || db == nil {
		return template.FuncMap{}
	}

	return template.FuncMap{
		"classification": func(s, p string) string {
			count, _ := form.HumanizeCount(ctx, db, s, p)
			return string(count)
		},
		"classificationStr": func(s, p string) string {
			return form.HumanizeCountStr(ctx, db, s, p)
		},
	}
}

var (
	hrefs = Hrefs()

	staticFuncMap = template.FuncMap{
		"bootstrap5":      func() string { return hrefs[Bootstrap5] },
		"bootstrap5JS":    func() string { return hrefs[Bootstrap5JS] },
		"bootstrapIcons":  func() string { return hrefs[BootstrapIcons] },
		"canvasAnsi":      func() string { return hrefs[ContentBinary] },
		"canvasReadme":    func() string { return hrefs[ContentText] },
		"demozooSanity":   func() string { return strconv.Itoa(demozoo.Sanity) },
		"chiptunePlayer":  func() string { return hrefs[ChiptunePlayer] },
		"editArtifact":    func() string { return hrefs[EditArtifact] },
		"editAssets":      func() string { return hrefs[EditAssets] },
		"editForApproval": func() string { return hrefs[EditForApproval] },
		"exampleDay":      func() string { return time.Now().Format("2") },
		"exampleMonth":    func() string { return time.Now().Format("1") },
		"exampleYear":     func() string { return time.Now().Format("2006") },
		"fmtName":         func(s string) string { return helper.Capitalize(strings.ToLower(s)) },
		"fmtRangeURI": func(s string) string {
			x, err := name.Humanize(name.Path(s))
			if err != nil {
				return err.Error()
			}
			return helper.Titleize(x)
		},
		"htmx":            func() string { return hrefs[Htmx] },
		"htmxRespTargets": func() string { return hrefs[HtmxRespTargets] },
		"initialisms":     func(s string) string { return lism.String(lism.Path(s)) },
		"indexJS":         func() string { return hrefs[IndexJS] },
		"jsdos6JS":        func() string { return hrefs[Jsdos6JS] },
		"dosboxJS":        func() string { return hrefs[DosboxJS] },
		"layout":          func() string { return hrefs[Layout] },
		"layoutJS":        func() string { return hrefs[LayoutJS] },
		"pouet":           func() string { return hrefs[Pouet] },
		"pouetSanity":     func() string { return strconv.Itoa(pouet.Sanity) },
		"tagGameHack":     func() string { return tags.GameHack.String() },
		"tagInstall":      func() string { return tags.Install.String() },
		"tagWindows":      func() string { return tags.Windows.String() },
		"tagDOS":          func() string { return tags.DOS.String() },
		"tagLogo":         func() string { return tags.Logo.String() },
		"tagProof":        func() string { return tags.Proof.String() },
		"tagText":         func() string { return tags.Text.String() },
		"tagTextAmiga":    func() string { return tags.TextAmiga.String() },
		"uploader":        func() string { return hrefs[Uploader] },
		"sub":             func(start, end int) int { return end - start },
	}
)

func (wa *WebApp) FuncStatic() template.FuncMap {
	return staticFuncMap
}

// FuncClosure returns a map of closures that return converted type or modified strings.
func (wa *WebApp) FuncClosure() template.FuncMap {
	return template.FuncMap{
		"logo": func() string { return string(wa.Brand) },
		"recordImgSampleStat": func(unid string) bool {
			return simple.ImageSampleStat(unid, dir.Directory(wa.Environment.AbsPreview))
		},
		"recordImgSample": func(unid string) template.HTML {
			return simple.ImageSample(unid, dir.Directory(wa.Environment.AbsPreview))
		},
		"recordThumbSample": func(unid string) template.HTML {
			return simple.ThumbSample(unid, dir.Directory(wa.Environment.AbsThumbnail))
		},
		"screenshot": func(unid, desc string) template.HTML {
			return simple.Screenshot(unid, desc, dir.Directory(wa.Environment.AbsPreview))
		},
		"sri_bootstrap5":      func() string { return wa.Subresource.Bootstrap5 },
		"sri_bootstrap5JS":    func() string { return wa.Subresource.Bootstrap5JS },
		"sri_bootstrapIcons":  func() string { return wa.Subresource.BootstrapIcons },
		"sri_canvasAnsi":      func() string { return wa.Subresource.CanvasAnsi },
		"sri_canvasReadme":    func() string { return wa.Subresource.CanvasReadme },
		"sri_chiptunePlayer":  func() string { return wa.Subresource.ChiptunePlayer },
		"sri_editArtifact":    func() string { return wa.Subresource.EditArtifact },
		"sri_editAssets":      func() string { return wa.Subresource.EditAssets },
		"sri_editForApproval": func() string { return wa.Subresource.EditForApproval },
		"sri_htmx":            func() string { return wa.Subresource.Htmx },
		"sri_htmxRespTargets": func() string { return wa.Subresource.HtmxRespTargets },
		"sri_indexJS":         func() string { return wa.Subresource.IndexJS },
		"sri_jsdos6JS":        func() string { return wa.Subresource.Jsdos6JS },
		"sri_dosboxJS":        func() string { return wa.Subresource.DosboxJS },
		"sri_layout":          func() string { return wa.Subresource.Layout },
		"sri_layoutJS":        func() string { return wa.Subresource.LayoutJS },
		"sri_pouet":           func() string { return wa.Subresource.Pouet },
		"sri_uploader":        func() string { return wa.Subresource.Uploader },
		"version":             func() string { return wa.Version },
		"thumb": func(unid, desc string, bottom bool) template.HTML {
			return simple.Thumb(unid, desc, dir.Directory(wa.Environment.AbsThumbnail), bottom)
		},
		"recordPreviewSrc": func(unid, ext string) string {
			return simple.AssetSrc(config.AbsPreview, wa.Environment.AbsPreview.String(), unid, ext)
		},
		"recordThumbnailSrc": func(unid, ext string) string {
			return simple.AssetSrc(config.AbsThumbnail, wa.Environment.AbsThumbnail.String(), unid, ext)
		},
		"og_image": wa.ogImage,
	}
}

func yearRange(start, end int) []int {
	const epoch = 1980
	if start < epoch {
		start = epoch
	}
	now := time.Now().Year()
	if end > now {
		end = now
	}

	// dont include start or end range years in the results
	start++
	years := make([]int, end-start)

	for i := range years {
		years[i] = start + i
	}

	return years
}

const (
	htmlAZ        = template.HTML(`<small><small class="fw-lighter">A-Z</small></small>`)
	htmlYear      = template.HTML(`<small><small class="fw-lighter">YEARS</small></small>`)
	htmlMSDos     = template.HTML(`<span class="text-nowrap">MS Dos</span>`)
	htmlMergeIcon = template.HTML(`<svg class="bi" aria-hidden="true" fill="currentColor">` +
		`<use xlink:href="/svg/bootstrap-icons.svg#forward"></use></svg>`)
)

var elementFuncs = template.FuncMap{
	"az":          func() template.HTML { return htmlAZ },
	"year":        func() template.HTML { return htmlYear },
	"msdos":       func() template.HTML { return htmlMSDos },
	"mergeIcon":   func() template.HTML { return htmlMergeIcon },
	"radioPublic": radioPublic,
	"radioHidden": radioHidden,
}

// FuncElem returns a map of functions that return HTML elements.
func (wa *WebApp) FuncElem() template.FuncMap {
	return elementFuncs
}

// FuncMap returns a map of all the template functions.
func (wa *WebApp) FuncMap(ctx context.Context, db *sql.DB) template.FuncMap {
	if db == nil {
		return nil
	}

	dst := wa.Funcs()

	src := wa.FuncDB(ctx, db)
	maps.Copy(dst, src)

	src = wa.FuncStatic()
	maps.Copy(dst, src)

	src = wa.FuncClosure()
	maps.Copy(dst, src)

	src = wa.FuncElem()
	maps.Copy(dst, src)

	return dst
}

func (wa *WebApp) ogImage(unid any) string {
	val, ok := unid.(string)
	if !ok {
		return favicon
	}
	if val == "" {
		return favicon
	}

	return simple.OpenGraphImg(val,
		dir.Directory(wa.Environment.AbsPreview),
		dir.Directory(wa.Environment.AbsThumbnail))
}

func (wa *WebApp) artifact(lock bool, files ...string) []string {
	files = append(
		files,
		GlobTo("artifactinfo.tmpl"),
		GlobTo("artifactjsdos.tmpl"),
	)

	if lock {
		return append(
			files,
			GlobTo("artifactedit_null.tmpl"),
			GlobTo("artifacteditjsdos_null.tmpl"),
			GlobTo("artifactlock_null.tmpl"),
		)
	}

	return append(
		files,
		GlobTo("artifactfile.tmpl"),
		GlobTo("artifactedit.tmpl"),
		GlobTo("artifacteditjsdos.tmpl"),
		GlobTo("artifactfooter.tmpl"),
		GlobTo("artifactlock.tmpl"),
	)
}

func (wa *WebApp) locked(lock bool, files ...string) []string {
	if lock {
		return append(
			files,
			GlobTo("layoutlock_null.tmpl"),
			GlobTo("layoutjs_null.tmpl"),
		)
	}

	return append(
		files,
		GlobTo("layoutlock.tmpl"),
		GlobTo("layoutjs.tmpl"),
	)
}

func (wa *WebApp) lockLayout(lock bool, files ...string) []string {
	if lock {
		return append(
			files,
			GlobTo("layoutup_null.tmpl"),
			GlobTo("layoutjsup_null.tmpl"),
			GlobTo("uploader_null.tmpl"),
		)
	}

	return append(
		files,
		GlobTo("layoutup.tmpl"),
		GlobTo("layoutjsup.tmpl"),
		GlobTo("uploader.tmpl"),
		GlobTo("uploader_modal.tmpl"),
	)
}

// parseFS returns a layout template for the given named view.
// Note that the name is relative to the view/defaults directory.
func (wa *WebApp) parseFS(ctx context.Context, db *sql.DB, name filename) *template.Template {
	if db == nil {
		return nil
	}

	files := wa.Layout(name)
	config := wa.Environment
	readonly := bool(config.ReadOnly)

	files = wa.locked(readonly, files...)
	files = wa.lockLayout(readonly, files...)

	// append any additional and embedded templates
	switch name {
	case artifactTmpl:
		files = wa.artifact(readonly, files...)
	case artifactsTmpl:
		files = append(files, GlobTo("artifactsedit.tmpl"))
	case categoriesTmpl:
		files = append(files, GlobTo("categoriesmore.tmpl"))
	case releaseryearTmpl:
		files = append(files, GlobTo("releasertimeline.tmpl"))
	case websitesTmpl:
		const individualWebsite = "website.tmpl"
		files = append(files, GlobTo(individualWebsite))
	}

	funcMap := wa.FuncMap(ctx, db)
	if funcMap == nil {
		return nil
	}

	return template.Must(template.New("").Funcs(funcMap).ParseFS(
		wa.View, files...),
	)
}

// radio name value must be the same for all inputs.
// 	id: artifact-editor-public
// 	id: artifact-editor-hidden

const (
	radio         = `<input type="radio" class="btn-check" name="artifact-editor-record"`
	radiopubicChk = template.HTML(
		radio + ` hx-patch="/editor/online/true" hx-include="[name='artifact-editor-key']" ` +
			`id="artifact-editor-public" autocomplete="off" checked>`,
	)
	radiopubic = template.HTML(
		radio + ` hx-patch="/editor/online/true" hx-include="[name='artifact-editor-key']" ` +
			`id="artifact-editor-public" autocomplete="off">`,
	)
	radiohideChk = template.HTML(
		radio + ` hx-patch="/editor/online/false" hx-include="[name='artifact-editor-key']" ` +
			`id="artifact-editor-hidden" autocomplete="off" checked>`,
	)
	radiohide = template.HTML(
		radio + ` hx-patch="/editor/online/false" hx-include="[name='artifact-editor-key']" ` +
			`id="artifact-editor-hidden" autocomplete="off">`,
	)
)

func radioPublic(checked bool) template.HTML {
	if checked {
		return radiopubicChk
	}
	return radiopubic
}

func radioHidden(checked bool) template.HTML {
	if checked {
		return radiohideChk
	}
	return radiohide
}

// LinkPreviews returns a slice of HTML formatted links for the artifact editor.
func LinkPreviews(youtube, demozoo, pouet, colors16, github, rels, sites string) []string {
	if youtube == "" && demozoo == "" && pouet == "" && colors16 == "" && github == "" && rels == "" && sites == "" {
		return nil
	}

	rel := func(url string) string {
		return `<a href="https://` + url + `">` + url + `</a>`
	}

	links := make([]string, 0, 1) // there will be at least one link
	if youtube != "" {
		links = append(links, rel("youtube.com/watch?v="+youtube))
	}
	if demozoo != "" && demozoo != "0" {
		links = append(links, rel("demozoo.org/productions/"+demozoo))
	}
	if pouet != "" && pouet != "0" {
		links = append(links, rel("pouet.net/prod.php?which="+pouet))
	}
	if colors16 != "" {
		links = append(links, rel("16colo.rs/"+colors16))
	}
	if github != "" {
		links = append(links, rel("github.com/"+github))
	}
	if rels != "" {
		links = append(links, strings.Split(string(simple.LinkRelations(rels)), "+")...)
	}
	if sites != "" {
		links = append(links, strings.Split(string(simple.LinkSites(sites)), "+")...)
	}

	return links
}

// LinkRelrs returns the groups associated with a release and a link to each group.
func LinkRelrs(magazine bool, a, b any) template.HTML {
	if a == nil || b == nil {
		return ""
	}
	return LinkReleasers(false, magazine, a, b)
}

// LinkRels returns the groups associated with a release and a link to each group.
func LinkRels(a, b any) template.HTML {
	if a == nil || b == nil {
		return ""
	}
	return LinkReleasers(false, false, a, b)
}

// LinkRelsPerf returns the groups associated with a release and a link to each group.
// It is a faster version of LinkRels and can be used with the templates that have large lists of group names.
func LinkRelsPerf(a, b any) template.HTML {
	if a == nil || b == nil {
		return ""
	}
	return LinkReleasers(true, false, a, b)
}

// LinkReleasers returns the groups associated with a release and a link to each group.
// The performant flag will use the group name instead of the much slower group slug formatter.
func LinkReleasers(performant, magazine bool, a, b any) template.HTML {
	if a == nil && b == nil {
		return ""
	}

	x := toString(a)
	x = strings.TrimSpace(x)
	y := toString(b)
	y = strings.TrimSpace(y)

	if x == "" && y == "" {
		return ""
	}

	if x == "" {
		x = y
		y = ""
	}

	const format = "error: %s"

	var prime, second string
	var err error

	if x != "" {
		prime, err = simple.MakeLink("1", x, linkClass, performant)
		if err != nil {
			return template.HTML(fmt.Sprintf(format, err))
		}
	}
	if y != "" {
		second, err = simple.MakeLink("2", y, linkClass, performant)
		if err != nil {
			return template.HTML(fmt.Sprintf(format, err))
		}
	}

	return simple.Releasers(prime, second, magazine)
}

func Mask(s string) string {
	return string(helper.MaskTerm([]byte(s)...))
}

func toString(val any) string {
	if val == nil {
		return ""
	}

	switch v := val.(type) {
	case string:
		return v
	case null.String:
		if v.Valid {
			return v.String
		}
	case *string:
		if v != nil {
			return *v
		}
	case fmt.Stringer:
		return v.String()
	}

	return ""
}
