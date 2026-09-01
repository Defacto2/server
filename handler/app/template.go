package app

// Package file template.go contains the template functions for the application.

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"io/fs"
	"maps"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/app/internal/filerecord"
	"github.com/Defacto2/server/handler/app/internal/simple"
	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/handler/form"
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
	closeAnchor = "</a>"
	input       = `<input class="form-check-input"`
	radio       = `<input type="radio" class="btn-check" name="artifact-editor-record"`
)

// Templ is the configuration and status of the web application templates.
type Templ struct {
	Public      fs.FS         // Public facing files.
	View        fs.FS         // Views are Go templates.
	Subresource SRI           // SRI are the Subresource Integrity hashes for the layout.
	Version     string        // Version is the current version of the app.
	Brand       []byte        // Brand contains to the Defacto2 ASCII logo.
	Environment config.Config // Environment configurations from the host system environment.
	RecordCount int64         // RecordCount is the total number of records in the database.
}

// Templates returns a map of the templates used by the route.
func (t *Templ) Templates(ctx context.Context, db *sql.DB) (map[string]*template.Template, error) {
	const format = "templates mapper %s: %w"
	if err := nils.Check(ctx, db); err != nil {
		return nil, fmt.Errorf(format, "check", err)
	}

	if err := t.Subresource.Verify(t.Public); err != nil {
		return nil, fmt.Errorf(format, "verify", err)
	}

	tmpls := make(map[string]*template.Template)
	for key, name := range *t.Pages() {
		tmpl := t.parseFS(ctx, db, name)
		tmpls[key] = tmpl
	}

	return tmpls, nil
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

// Pages returns a map of the template names and their corresponding filenames.
func (t *Templ) Pages() *Page {
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

	return &Page{
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
}

func (t *Templ) Layout(name filename) []string {
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
func (t *Templ) Funcs() template.FuncMap {
	return template.FuncMap{
		"add":                helper.Add1,
		"attribute":          Attribute,
		"brief":              Brief,
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
		"trimSpace":          TrimSpace,
		"websiteIcon":        WebsiteIcon,
		"urlEncode":          URLEncode,
	}
}

// FuncClosures returns a map of closures that return converted type or modified strings.
func (t *Templ) FuncClosures(ctx context.Context, db *sql.DB) *template.FuncMap { //nolint:funlen
	if db == nil {
		return nil
	}

	hrefs := *Hrefs()
	return &template.FuncMap{
		"bootstrap5": func() string {
			return hrefs[Bootstrap5]
		},
		"bootstrap5JS": func() string {
			return hrefs[Bootstrap5JS]
		},
		"bootstrapIcons": func() string {
			return hrefs[BootstrapIcons]
		},
		"capitalize": helper.Capitalize,
		"canvasAnsi": func() string {
			return hrefs[ContentBinary]
		},
		"canvasReadme": func() string {
			return hrefs[ContentText]
		},
		"classification": func(s, p string) string {
			count, _ := form.HumanizeCount(ctx, db, s, p)
			return string(count)
		},
		"classificationStr": func(s, p string) string {
			return form.HumanizeCountStr(ctx, db, s, p)
		},
		"demozooSanity": func() string {
			return strconv.Itoa(demozoo.Sanity)
		},
		"chiptunePlayer": func() string {
			return hrefs[ChiptunePlayer]
		},
		"editArtifact": func() string {
			return hrefs[EditArtifact]
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
		"htmx": func() string {
			return hrefs[Htmx]
		},
		"htmxRespTargets": func() string {
			return hrefs[HtmxRespTargets]
		},
		"initialisms": func(s string) string {
			return lism.String(lism.Path(s))
		},
		"indexJS": func() string {
			return hrefs[IndexJS]
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
			return string(t.Brand)
		},
		"pouet": func() string {
			return hrefs[Pouet]
		},
		"pouetSanity": func() string {
			return strconv.Itoa(pouet.Sanity)
		},
		"recordImgSampleStat": func(unid string) bool {
			return simple.ImageSampleStat(unid, dir.Directory(t.Environment.AbsPreview))
		},
		"recordImgSample": func(unid string) template.HTML {
			return simple.ImageSample(unid, dir.Directory(t.Environment.AbsPreview))
		},
		"recordThumbSample": func(unid string) template.HTML {
			return simple.ThumbSample(unid, dir.Directory(t.Environment.AbsThumbnail))
		},
		"screenshot": func(unid, desc string) template.HTML {
			return simple.Screenshot(unid, desc, dir.Directory(t.Environment.AbsPreview))
		},
		"sri_bootstrap5": func() string {
			return t.Subresource.Bootstrap5
		},
		"sri_bootstrap5JS": func() string {
			return t.Subresource.Bootstrap5JS
		},
		"sri_bootstrapIcons": func() string {
			return t.Subresource.BootstrapIcons
		},
		"sri_canvasAnsi": func() string {
			return t.Subresource.CanvasAnsi
		},
		"sri_canvasReadme": func() string {
			return t.Subresource.CanvasReadme
		},
		"sri_chiptunePlayer": func() string {
			return t.Subresource.ChiptunePlayer
		},
		"sri_editArtifact": func() string {
			return t.Subresource.EditArtifact
		},
		"sri_editAssets": func() string {
			return t.Subresource.EditAssets
		},
		"sri_editForApproval": func() string {
			return t.Subresource.EditForApproval
		},
		"sri_htmx": func() string {
			return t.Subresource.Htmx
		},
		"sri_htmxRespTargets": func() string {
			return t.Subresource.HtmxRespTargets
		},
		"sri_indexJS": func() string {
			return t.Subresource.IndexJS
		},
		"sri_jsdos6JS": func() string {
			return t.Subresource.Jsdos6JS
		},
		"sri_dosboxJS": func() string {
			return t.Subresource.DosboxJS
		},
		"sri_layout": func() string {
			return t.Subresource.Layout
		},
		"sri_layoutJS": func() string {
			return t.Subresource.LayoutJS
		},
		"sri_pouet": func() string {
			return t.Subresource.Pouet
		},
		"sri_uploader": func() string {
			return t.Subresource.Uploader
		},
		"toLower": strings.ToLower,
		"uploader": func() string {
			return hrefs[Uploader]
		},
		"version": func() string {
			return t.Version
		},
		"tagGameHack":   func() string { return tags.GameHack.String() },
		"tagInstall":    func() string { return tags.Install.String() },
		"tagWindows":    func() string { return tags.Windows.String() },
		"tagDOS":        func() string { return tags.DOS.String() },
		"tagLogo":       func() string { return tags.Logo.String() },
		"tagProof":      func() string { return tags.Proof.String() },
		"tagText":       func() string { return tags.Text.String() },
		"tagTextAmiga":  func() string { return tags.TextAmiga.String() },
		"tidbitMissing": tidbit.Missing,
		"thumb": func(unid, desc string, bottom bool) template.HTML {
			return simple.Thumb(unid, desc, dir.Directory(t.Environment.AbsThumbnail), bottom)
		},
		"recordPreviewSrc": func(unid, ext string) string {
			return simple.AssetSrc(config.AbsPreview, t.Environment.AbsPreview.String(), unid, ext)
		},
		"recordThumbnailSrc": func(unid, ext string) string {
			return simple.AssetSrc(config.AbsThumbnail, t.Environment.AbsThumbnail.String(), unid, ext)
		},
		"og_image":  t.ogImage,
		"yearRange": yearRange,
		"sub": func(start, end int) int {
			return end - start
		},
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

// Elements returns a map of functions that return HTML elements.
func (t *Templ) Elements() *template.FuncMap {
	return &template.FuncMap{
		"az": func() template.HTML {
			return template.HTML(`<small><small class="fw-lighter">A-Z</small></small>`)
		},
		"year": func() template.HTML {
			return template.HTML(`<small><small class="fw-lighter">YEARS</small></small>`)
		},
		"mergeIcon": func() template.HTML {
			return template.HTML(`<svg class="bi" aria-hidden="true" fill="currentColor">` +
				`<use xlink:href="/svg/bootstrap-icons.svg#forward"></use></svg>`)
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

// FuncMap returns a map of all the template functions.
func (t *Templ) FuncMap(ctx context.Context, db *sql.DB) *template.FuncMap {
	if db == nil {
		return nil
	}

	src := t.FuncClosures(ctx, db)
	if src == nil {
		return nil
	}

	dst := t.Funcs()
	maps.Copy(dst, *src)

	src = t.Elements()
	if src == nil {
		return nil
	}
	maps.Copy(dst, *src)

	return &dst
}

func (t *Templ) ogImage(unid any) string {
	const favicon = "/image/layout/defacto2-ascii.png"

	val, ok := unid.(string)
	if !ok {
		return favicon
	}
	if val == "" {
		return favicon
	}

	return simple.OpenGraphImg(val,
		dir.Directory(t.Environment.AbsPreview),
		dir.Directory(t.Environment.AbsThumbnail))
}

func (t *Templ) artifact(lock bool, files ...string) []string {
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

func (t *Templ) locked(lock bool, files ...string) []string {
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

func (t *Templ) lockLayout(lock bool, files ...string) []string {
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
func (t *Templ) parseFS(ctx context.Context, db *sql.DB, name filename) *template.Template {
	if db == nil {
		return nil
	}

	files := t.Layout(name)
	config := t.Environment
	readonly := bool(config.ReadOnly)

	files = t.locked(readonly, files...)
	files = t.lockLayout(readonly, files...)

	// append any additional and embedded templates
	switch name {
	case artifactTmpl:
		files = t.artifact(readonly, files...)
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

	funcMap := t.FuncMap(ctx, db)
	if funcMap == nil {
		return nil
	}

	return template.Must(template.New("").Funcs(
		*funcMap,
	).ParseFS(t.View, files...))
}

func recordLastMod(b bool) template.HTML {
	const id = `recordLMBtn`
	const class = `btn btn-outline-secondary`
	const button = `button`

	if b {
		// tooltips do not work on disabled buttons
		const title = `No last modification date found`
		return template.HTML(`<button id="` + id + `" class="` + class + `" type="` + button + `" ` +
			`data-bs-toggle="tooltip" data-bs-title="` + title + `" disabled>`)
	}

	const title = `Apply the file last modified date`
	return template.HTML(`<button id="` + id + `" class="` + class + `" type="` + button + `" ` +
		`data-bs-toggle="tooltip" data-bs-title="` + title + `">`)
}

func radioPublic(b bool) template.HTML {
	const patch = `/editor/online/true`
	const include = `[name='artifact-editor-key']`
	const id = `artifact-editor-public`
	const htmx = ` hx-patch="` + patch + `"	hx-include="` + include + `" id="` + id +
		`" autocomplete="off"`

	if b {
		return template.HTML(radio + htmx + ` checked>`)
	}

	return template.HTML(radio + htmx + `>`)
}

func radioHidden(b bool) template.HTML {
	const patch = `/editor/online/false`
	const include = `[name='artifact-editor-key']`
	const id = `artifact-editor-hidden`
	const htmx = ` hx-patch="` + patch + `"	hx-include="` + include + `" id="` + id +
		`" autocomplete="off"`

	if !b {
		return template.HTML(radio + htmx + ` checked>`)
	}

	return template.HTML(radio + htmx + `>`)
}

func recordOnline(b bool) template.HTML {
	const htm = ` name="online" type="checkbox" role="switch" id="recordOnline"`

	if b {
		return template.HTML(input + htm + ` checked>`)
	}

	return template.HTML((input + htm + `>`))
}

func recordReadme(b bool) template.HTML {
	const htm = ` name="hide-readme" type="checkbox" role="switch" id="edHideMe"`

	if b {
		return template.HTML(input + htm + ` checked>`)
	}

	return template.HTML((input + htm + `>`))
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
	const class = "text-nowrap link-offset-2 link-underline link-underline-opacity-25"

	var x, y string
	switch i := a.(type) {
	case string:
		x = reflect.ValueOf(i).String()
	case null.String:
		if i.Valid {
			x = i.String
		}
	}
	switch i := b.(type) {
	case string:
		y = reflect.ValueOf(i).String()
	case null.String:
		if i.Valid {
			y = i.String
		}
	}

	x = strings.TrimSpace(x)
	y = strings.TrimSpace(y)
	if x == "" && y != "" {
		x = y
		y = ""
	}

	const format = "error: %s"
	var prime, second string
	var err error
	if x != "" {
		prime, err = simple.MakeLink("1", x, class, performant)
		if err != nil {
			return template.HTML(fmt.Sprintf(format, err))
		}
	}
	if y != "" {
		second, err = simple.MakeLink("2", y, class, performant)
		if err != nil {
			return template.HTML(fmt.Sprintf(format, err))
		}
	}

	return simple.Releasers(prime, second, magazine)
}

func Mask(s string) string {
	return string(helper.MaskTerm([]byte(s)...))
}
