package app

// Package file template.go contains the template functions for the application.

import (
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"maps"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/releaser"
	"github.com/Defacto2/releaser/initialism"
	"github.com/Defacto2/releaser/name"
	"github.com/Defacto2/server/handler/app/internal/filerecord"
	"github.com/Defacto2/server/handler/app/internal/simple"
	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/handler/form"
	"github.com/Defacto2/server/handler/pouet"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/tags"
	"github.com/aarondl/null/v8"
)

const (
	closeAnchor = "</a>"
	input       = "<input class=\"form-check-input\""
	radio       = `<input type="radio" class="btn-check" name="artifact-editor-record"`
)

// Templ is the configuration and status of the web application templates.
type Templ struct {
	Public      embed.FS      // Public facing files.
	View        embed.FS      // Views are Go templates.
	Subresource SRI           // SRI are the Subresource Integrity hashes for the layout.
	Version     string        // Version is the current version of the app.
	Brand       []byte        // Brand contains to the Defacto2 ASCII logo.
	Environment config.Config // Environment configurations from the host system environment.
	RecordCount int           // RecordCount is the total number of records in the database.
}

// Templates returns a map of the templates used by the route.
func (t *Templ) Templates(db *sql.DB) (map[string]*template.Template, error) {
	if err := t.Subresource.Verify(t.Public); err != nil {
		return nil, fmt.Errorf("app templates verify, %w", err)
	}
	tmpls := make(map[string]*template.Template)
	for key, name := range *t.Pages() {
		tmpl := t.parseFS(db, name)
		tmpls[key] = tmpl
	}
	return tmpls, nil
}

const (
	artifactTmpl   = "artifact.tmpl"
	artifactsTmpl  = "artifacts.tmpl"
	categoriesTmpl = "categories.tmpl"
	releaserTmpl   = "releaser.tmpl"
	scenerTmpl     = "scener.tmpl"
	websitesTmpl   = "websites.tmpl"
)

type filename string // filename is the name of the template file in the view directory.

type Page map[string]filename

// Pages returns a map of the template names and their corresponding filenames.
func (t *Templ) Pages() *Page {
	return &Page{
		"areacodes":     "areacodes.tmpl",
		"artifact":      artifactTmpl,
		"artifacts":     artifactsTmpl,
		"bbs":           releaserTmpl,
		"bbs-year":      "releaseryear.tmpl",
		"categories":    categoriesTmpl,
		"configs":       "configurations.tmpl",
		"coder":         scenerTmpl,
		"ftp":           releaserTmpl,
		"history":       "history.tmpl",
		"index":         "index.tmpl",
		"interview":     "interview.tmpl",
		"magazine":      "releaseryear.tmpl",
		"magazine-az":   releaserTmpl,
		"new":           "new.tmpl",
		"releaser":      releaserTmpl,
		"releaser-year": "releaseryear.tmpl",
		"scener":        scenerTmpl,
		"searchhtmx":    "searchhtmx.tmpl",
		"searchpost":    "searchpost.tmpl",
		"signin":        "signin.tmpl",
		"signout":       "signout.tmpl",
		"status":        "status.tmpl",
		"thanks":        "thanks.tmpl",
		"thescene":      "thescene.tmpl",
		"titles":        "titles.tmpl",
		"websites":      websitesTmpl,
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
		"downloadB":          simple.DownloadB,
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
		"jsdosUsage":         filerecord.JsdosUsage,
		"recordInfoOSTag":    TagWithOS,
		"recordLinkPreviews": LinkPreviews,
		"recordTagInfo":      TagBrief,
		"safeBBS":            SafeBBS,
		"safeHTML":           SafeHTML,
		"safeJS":             SafeJS,
		"slugify":            helper.Slug,
		"subTitle":           SubTitle,
		"tagOption":          TagOption,
		"trimSpace":          TrimSpace,
		"websiteIcon":        WebsiteIcon,
		"urlEncode":          URLEncode,
	}
}

// FuncClosures returns a map of closures that return converted type or modified strings.
func (t *Templ) FuncClosures(db *sql.DB) *template.FuncMap { //nolint:funlen
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
		"classification": func(s, p string) string {
			count, _ := form.HumanizeCount(db, s, p)
			return string(count)
		},
		"classificationStr": func(s, p string) string {
			return form.HumanizeCountStr(db, s, p)
		},
		"demozooSanity": func() string {
			return strconv.Itoa(demozoo.Sanity)
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
			return string(t.Brand)
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
		"sri_readme": func() string {
			return t.Subresource.Readme
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
		"tagGameHack":  func() string { return tags.GameHack.String() },
		"tagInstall":   func() string { return tags.Install.String() },
		"tagWindows":   func() string { return tags.Windows.String() },
		"tagDOS":       func() string { return tags.DOS.String() },
		"tagLogo":      func() string { return tags.Logo.String() },
		"tagProof":     func() string { return tags.Proof.String() },
		"tagText":      func() string { return tags.Text.String() },
		"tagTextAmiga": func() string { return tags.TextAmiga.String() },
		"thumb": func(unid, desc string, bottom bool) template.HTML {
			return simple.Thumb(unid, desc, dir.Directory(t.Environment.AbsThumbnail), bottom)
		},
		"recordPreviewSrc": func(unid, ext string) string {
			return simple.AssetSrc(config.Prev, t.Environment.AbsPreview, unid, ext)
		},
		"recordThumbnailSrc": func(unid, ext string) string {
			return simple.AssetSrc(config.Thumb, t.Environment.AbsThumbnail, unid, ext)
		},
	}
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
func (t *Templ) FuncMap(db *sql.DB) *template.FuncMap {
	funcs := t.Funcs()
	maps.Copy(funcs, *t.FuncClosures(db))
	maps.Copy(funcs, *t.Elements())
	return &funcs
}

func (t *Templ) artifact(lock bool, files ...string) []string {
	files = append(files,
		GlobTo("artifactinfo.tmpl"),
		GlobTo("artifactjsdos.tmpl"),
	)
	if lock {
		return append(files,
			GlobTo("artifactedit_null.tmpl"),
			GlobTo("artifacteditjsdos_null.tmpl"),
			GlobTo("artifactlock_null.tmpl"),
		)
	}
	return append(files,
		GlobTo("artifactfile.tmpl"),
		GlobTo("artifactedit.tmpl"),
		GlobTo("artifacteditjsdos.tmpl"),
		GlobTo("artifactfooter.tmpl"),
		GlobTo("artifactlock.tmpl"),
	)
}

func (t *Templ) locked(lock bool, files ...string) []string {
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

func (t *Templ) lockLayout(lock bool, files ...string) []string {
	if lock {
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
		GlobTo("uploader_modal.tmpl"),
	)
}

// parseFS returns a layout template for the given named view.
// Note that the name is relative to the view/defaults directory.
func (t *Templ) parseFS(db *sql.DB, name filename) *template.Template {
	files := t.Layout(name)
	config := t.Environment
	files = t.locked(config.ReadOnly, files...)
	files = t.lockLayout(config.ReadOnly, files...)
	// append any additional and embedded templates
	switch name {
	case artifactTmpl:
		files = t.artifact(config.ReadOnly, files...)
	case artifactsTmpl:
		files = append(files, GlobTo("artifactsedit.tmpl"))
	case categoriesTmpl:
		files = append(files, GlobTo("categoriesmore.tmpl"))
	case websitesTmpl:
		const individualWebsite = "website.tmpl"
		files = append(files, GlobTo(individualWebsite))
	}
	return template.Must(template.New("").Funcs(
		*t.FuncMap(db)).ParseFS(t.View, files...))
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
	const htmx = ` hx-patch="/editor/online/true"
	hx-include="[name='artifact-editor-key']"`
	if b {
		return template.HTML(radio +
			htmx + ` id="artifact-editor-public" autocomplete="off" checked>`)
	}
	return template.HTML(radio +
		htmx + ` id="artifact-editor-public" autocomplete="off">`)
}

func radioHidden(b bool) template.HTML {
	const htmx = ` hx-patch="/editor/online/false"
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
	var av, bv string
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

	av, bv = strings.TrimSpace(av), strings.TrimSpace(bv)
	if av == "" && bv != "" {
		av = bv
		bv = ""
	}

	var prime, second string
	var err error
	if av != "" {
		prime, err = simple.MakeLink(av, class, performant)
		if err != nil {
			return template.HTML(fmt.Sprintf("error: %s", err))
		}
	}
	if bv != "" {
		second, err = simple.MakeLink(bv, class, performant)
		if err != nil {
			return template.HTML(fmt.Sprintf("error: %s", err))
		}
	}
	return simple.Releasers(prime, second, magazine)
}
