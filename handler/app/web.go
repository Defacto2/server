package app

// Package file funcmap.go contains the custom template functions for the web framework.
// The functions are used by the HTML templates to format data.

import (
	"embed"
	"html/template"
	"strings"
	"time"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/releaser/initialism"
	"github.com/Defacto2/releaser/name"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/helper"
	"go.uber.org/zap"
)

// Web is the configuration and status of the web app.
// Rename to app or template?
type Web struct {
	Brand       *[]byte            // Brand points to the Defacto2 ASCII logo.
	Import      *config.Config     // Import configurations from the host system environment.
	Logger      *zap.SugaredLogger // Logger is the zap sugared logger.
	Subresource SRI                // SRI are the Subresource Integrity hashes for the layout.
	Version     string             // Version is the current version of the app.
	Public      embed.FS           // Public facing files.
	View        embed.FS           // Views are Go templates.
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
	return template.FuncMap{
		"az": func() template.HTML {
			return template.HTML(`<small><small class=\"fw-lighter\">A-Z</small></small>`)
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
				return template.HTML("<button class=\"btn btn-outline-secondary\" type=\"button\" disabled>")
			}
			return template.HTML("<button id=\"recordLMBtn\" class=\"btn btn-outline-secondary\" type=\"button\" " +
				"data-bs-toggle=\"tooltip\" data-bs-title=\"Apply the file last modified date\">")
		},
		"recordOnline": func(b bool) template.HTML {
			if b {
				return template.HTML("<input class=\"form-check-input\"" +
					" name=\"online\" type=\"checkbox\" role=\"switch\" id=\"recordOnline\" checked>")
			}
			return template.HTML(("<input class=\"form-check-input\"" +
				" name=\"online\" type=\"checkbox\" role=\"switch\" id=\"recordOnline\">"))
		},
		"recordReadme": func(b bool) template.HTML {
			if b {
				return template.HTML("<input class=\"form-check-input\"" +
					" name=\"hide-readme\" type=\"checkbox\" role=\"switch\" id=\"edHideMe\" checked>")
			}
			return template.HTML(("<input class=\"form-check-input\"" +
				" name=\"hide-readme\" type=\"checkbox\" role=\"switch\" id=\"edHideMe\">"))
		},
	}
}

// TemplateClosures returns a map of closures that return converted type or modified strings.
func (web Web) TemplateClosures() template.FuncMap {
	hrefs := Hrefs()
	return template.FuncMap{
		"bootstrap": func() string {
			return hrefs[Bootstrap]
		},
		"bootstrapJS": func() string {
			return hrefs[BootstrapJS]
		},
		"editArchive": func() string {
			return hrefs[EditArchive]
		},
		"editAssets": func() string {
			return hrefs[EditAssets]
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
			return helper.Capitalize(x)
		},
		"fontAwesome": func() string {
			return hrefs[FontAwesome]
		},
		"initialisms": func(s string) string {
			return initialism.Join(initialism.Path(s))
		},
		"jsdosUI": func() string {
			return hrefs[JSDosUI]
		},
		"jsdosW": func() string {
			return hrefs[JSDosW]
		},
		"layout": func() string {
			return hrefs[Layout]
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
		"restPouet": func() string {
			return hrefs[RESTPouet]
		},
		"restZoo": func() string {
			return hrefs[RESTZoo]
		},
		"sri_bootstrap": func() string {
			return web.Subresource.Bootstrap
		},
		"sri_bootstrapJS": func() string {
			return web.Subresource.BootstrapJS
		},
		"sri_editArchive": func() string {
			return web.Subresource.EditArchive
		},
		"sri_editAssets": func() string {
			return web.Subresource.EditAssets
		},
		"sri_editor": func() string {
			return web.Subresource.Editor
		},
		"sri_fontAwesome": func() string {
			return web.Subresource.FontAwesome
		},
		"sri_jsdosUI": func() string {
			return web.Subresource.JSDosUI
		},
		"sri_jsdosW": func() string {
			return web.Subresource.JSDosW
		},
		"sri_layout": func() string {
			return web.Subresource.Layout
		},
		"sri_pouet": func() string {
			return web.Subresource.Pouet
		},
		"sri_readme": func() string {
			return web.Subresource.Readme
		},
		"sri_restPouet": func() string {
			return web.Subresource.RESTPouet
		},
		"sri_restZoo": func() string {
			return web.Subresource.RESTZoo
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
func (web Web) TemplateFuncs() template.FuncMap {
	funcMap := template.FuncMap{
		"add":               helper.Add1,
		"attribute":         Attribute,
		"brief":             Brief,
		"describe":          Describe,
		"downloadB":         DownloadB,
		"byteFile":          ByteFile,
		"byteFileS":         ByteFileS,
		"fmtDay":            Day,
		"fmtMonth":          Month,
		"fmtPrefix":         Prefix,
		"fmtRoles":          helper.FmtSlice,
		"fmtURI":            releaser.Link, // this is not performant for large lists, instead use fmtRangeURI in TemplateStrings()
		"lastUpdated":       LastUpdated,
		"linkDownload":      LinkDownload,
		"linkHref":          LinkHref,
		"linkPage":          LinkPage,
		"linkPreview":       LinkPreview,
		"linkRemote":        LinkRemote,
		"linkRelrs":         LinkRelFast,
		"linkScnr":          LinkScnr,
		"linkWiki":          LinkWiki,
		"logoText":          LogoText,
		"mimeMagic":         MimeMagic,
		"recordImgSample":   web.ImageSample,
		"recordThumbSample": web.ThumbSample,
		"recordInfoOSTag":   TagWithOS,
		"recordTagInfo":     TagBrief,
		"safeHTML":          SafeHTML,
		"screenshot":        web.Screenshot,
		"subTitle":          SubTitle,
		"thumb":             web.Thumb,
		"trimSiteSuffix":    TrimSiteSuffix,
		"trimSpace":         TrimSpace,
		"websiteIcon":       WebsiteIcon,
	}
	return funcMap
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
	config := web.Import
	files = uploaderTmpls(config.ReadMode, files...)
	// append any additional and embedded templates
	switch name {
	case "about.tmpl":
		files = aboutTmpls(config.ReadMode, files...)
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

func templates() map[string]filename {
	const releaser, scener = "releaser.tmpl", "scener.tmpl"
	return map[string]filename{
		"index":       "index.tmpl",
		"about":       "about.tmpl",
		"bbs":         releaser,
		"coder":       scener,
		"file":        "file.tmpl",
		"files":       "files.tmpl",
		"ftp":         releaser,
		"history":     "history.tmpl",
		"interview":   "interview.tmpl",
		"magazine":    "releaser_year.tmpl",
		"magazine-az": releaser,
		"reader":      "reader.tmpl",
		"releaser":    releaser,
		"scener":      scener,
		"searchList":  "searchList.tmpl",
		"searchPost":  "searchPost.tmpl",
		"signin":      "signin.tmpl",
		"signout":     "signout.tmpl",
		"status":      "status.tmpl",
		"thanks":      "thanks.tmpl",
		"thescene":    "the_scene.tmpl",
		"websites":    "websites.tmpl",
	}
}

func uploaderTmpls(lock bool, files ...string) []string {
	if lock {
		return append(files,
			GlobTo("layout_editor_null.tmpl"),
			GlobTo("layout_uploader_null.tmpl"),
			GlobTo("uploader_null.tmpl"))
	}
	return append(files,
		GlobTo("layout_editor.tmpl"),
		GlobTo("layout_uploader.tmpl"),
		GlobTo("uploader.tmpl"))
}

func aboutTmpls(lock bool, files ...string) []string {
	files = append(files,
		GlobTo("about_table.tmpl"),
		GlobTo("about_jsdos.tmpl"),
		GlobTo("about_editor_archive.tmpl"))
	if lock {
		return append(files,
			GlobTo("about_editor_null.tmpl"),
			GlobTo("about_editor_table_null.tmpl"),
			GlobTo("about_table_switch_null.tmpl"))
	}
	return append(files,
		GlobTo("about_editor.tmpl"),
		GlobTo("about_editor_table.tmpl"),
		GlobTo("about_table_switch.tmpl"))
}
