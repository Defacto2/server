package app

// Package file funcmap.go contains the custom template functions for the web framework.
// The functions are used by the HTML templates to format data.

import (
	"html/template"
	"strings"
	"time"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/releaser/initialism"
	"github.com/Defacto2/releaser/name"
	"github.com/Defacto2/server/internal/helper"
)

const (
	merge template.HTML = `<svg class="bi" aria-hidden="true" fill="currentColor">` +
		`<use xlink:href="/bootstrap-icons.svg#forward"></use></svg>`
	msDos template.HTML = `<span class="text-nowrap">MS Dos</span>`
)

// TemplateFuncMap are a collection of mapped functions that can be used in a template.
func (web Web) TemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"add":            helper.Add1,
		"attribute":      Attribute,
		"brief":          Brief,
		"describe":       Describe,
		"downloadB":      DownloadB,
		"byteFile":       ByteFile,
		"byteFileS":      ByteFileS,
		"fmtDay":         Day,
		"fmtMonth":       Month,
		"fmtPrefix":      Prefix,
		"fmtRoles":       helper.FmtSlice,
		"fmtURI":         releaser.Link, // this is not performant for large lists, use fmtFastURI instead
		"lastUpdated":    LastUpdated,
		"linkDownload":   LinkDownload,
		"linkHref":       LinkHref,
		"linkPage":       LinkPage,
		"linkPreview":    LinkPreview,
		"linkRemote":     LinkRemote,
		"linkRelrs":      LinkRelFast,
		"linkScnr":       LinkScnr,
		"linkWiki":       LinkWiki,
		"logoText":       LogoText,
		"mimeMagic":      MimeMagic,
		"safeHTML":       SafeHTML,
		"screenshot":     web.Screenshot,
		"subTitle":       SubTitle,
		"thumb":          web.Thumb,
		"trimSiteSuffix": TrimSiteSuffix,
		"websiteIcon":    WebsiteIcon,
		// these closures should only return simple values
		"initialisms": func(s string) string {
			return initialism.Join(initialism.Path(s))
		},
		"fmtName": func(s string) string {
			return helper.Capitalize(strings.ToLower(s))
		},
		"fmtFastURI": func(s string) string {
			x, err := name.Humanize(name.Path(s))
			if err != nil {
				return err.Error()
			}
			return helper.Capitalize(x)
		},
		"logo": func() string {
			return string(*web.Brand)
		},
		"mergeIcon": func() template.HTML {
			return merge
		},
		"msdos": func() template.HTML {
			return msDos
		},
		"sriBootCSS": func() string {
			return web.Subresource.BootstrapCSS
		},
		"sriBootJS": func() string {
			return web.Subresource.BootstrapJS
		},
		"sriFA": func() string {
			return web.Subresource.FontAwesome
		},
		"sriLayout": func() string {
			return web.Subresource.LayoutCSS
		},
		"sriPouet": func() string {
			return web.Subresource.PouetJS
		},
		"sriReadme": func() string {
			return web.Subresource.ReadmeJS
		},
		"sriJSDos": func() string {
			return web.Subresource.JSDos
		},
		"sriJSWDos": func() string {
			return web.Subresource.JSWDos
		},
		"cssBoot": func() string {
			return BootCSS
		},
		"cssLayout": func() string {
			return LayoutCSS
		},
		"jsBoot": func() string {
			return BootJS
		},
		"jsFA": func() string {
			return FAJS
		},
		"jsPouet": func() string {
			return PouetJS
		},
		"jsReadme": func() string {
			return ReadmeJS
		},
		"jsDos": func() string {
			return JSDos
		},
		"jsWDos": func() string {
			return JSWDos
		},
		"version": func() string {
			return web.Version
		},
		"exampleYear": func() string {
			return time.Now().Format("2006")
		},
		"exampleMonth": func() string {
			return time.Now().Format("1")
		},
		"exampleDay": func() string {
			return time.Now().Format("2")
		},
		"uploader": func() string {
			return UploaderJS
		},
		"sriUploader": func() string {
			return web.Subresource.UploaderJS
		},
	}
}
