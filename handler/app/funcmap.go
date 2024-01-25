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
	hrefs := Hrefs()
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
			return web.Subresource.Bootstrap
		},
		"sriBootJS": func() string {
			return web.Subresource.BootstrapJS
		},
		"sriFA": func() string {
			return web.Subresource.FontAwesome
		},
		"sriJSDos": func() string {
			return web.Subresource.JSDosUI
		},
		"sriJSWDos": func() string {
			return web.Subresource.JSDosW
		},
		"sriLayout": func() string {
			return web.Subresource.Layout
		},
		"sriPouet": func() string {
			return web.Subresource.Pouet
		},
		"sriReadme": func() string {
			return web.Subresource.Readme
		},
		"sriUploader": func() string {
			return web.Subresource.Uploader
		},
		"cssBoot": func() string {
			return hrefs[Bootstrap]
		},
		"cssLayout": func() string {
			return hrefs[Layout]
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
		"jsBoot": func() string {
			return hrefs[BootstrapJS]
		},
		"jsDos": func() string {
			return hrefs[JSDosUI]
		},
		"jsWDos": func() string {
			return hrefs[JSDosW]
		},
		"jsFA": func() string {
			return hrefs[FontAwesome]
		},
		"jsPouet": func() string {
			return hrefs[Pouet]
		},
		"jsReadme": func() string {
			return hrefs[Readme]
		},
		"version": func() string {
			return web.Version
		},
		"uploader": func() string {
			return hrefs[Uploader]
		},
		"editorForm": func() string {
			return hrefs[Editor]
		},
		"sriEditorForm": func() string {
			return web.Subresource.Editor
		},
		"editorAssetsForm": func() string {
			return hrefs[EditAssets]
		},
		"sriEditorAssetsForm": func() string {
			return web.Subresource.EditAssets
		},
		"editorArchiveForm": func() string {
			return hrefs[EditArchive]
		},
		"sriEditorArchiveForm": func() string {
			return web.Subresource.EditArchive
		},
		"restPouet": func() string {
			return hrefs[RESTPouet]
		},
		"sriRestPouet": func() string {
			return web.Subresource.RESTPouet
		},
		"restZoo": func() string {
			return hrefs[RESTZoo]
		},
		"sriRestZoo": func() string {
			return web.Subresource.RESTZoo
		},
		// "recordReleasers": RecordRels,
		"tagSel": TagSel,
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
		"recordLastMod": func(b bool) template.HTML {
			if b {
				// tooltips do not work on disabled buttons
				return template.HTML("<button class=\"btn btn-outline-secondary\" type=\"button\" disabled>")
			}
			return template.HTML("<button id=\"recordLMBtn\" class=\"btn btn-outline-secondary\" type=\"button\" " +
				"data-bs-toggle=\"tooltip\" data-bs-title=\"Apply the file last modified date\">")
		},
		"recordImgSample":   web.ImageSample,
		"recordThumbSample": web.ThumbSample,
		"recordInfoOSTag":   TagWithOS,
		"recordTagInfo":     TagBrief,
	}
}
