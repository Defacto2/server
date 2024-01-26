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
		"cssBoot": func() string {
			return hrefs[Bootstrap]
		},
		"cssLayout": func() string {
			return hrefs[Layout]
		},
		"editorArchiveForm": func() string {
			return hrefs[EditArchive]
		},
		"editorAssetsForm": func() string {
			return hrefs[EditAssets]
		},
		"editorForm": func() string {
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
		"fmtFastURI": func(s string) string {
			x, err := name.Humanize(name.Path(s))
			if err != nil {
				return err.Error()
			}
			return helper.Capitalize(x)
		},
		"fmtName": func(s string) string {
			return helper.Capitalize(strings.ToLower(s))
		},
		"initialisms": func(s string) string {
			return initialism.Join(initialism.Path(s))
		},
		"jsBoot": func() string {
			return hrefs[BootstrapJS]
		},
		"jsDos": func() string {
			return hrefs[JSDosUI]
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
		"jsWDos": func() string {
			return hrefs[JSDosW]
		},
		"logo": func() string {
			return string(*web.Brand)
		},
		"restPouet": func() string {
			return hrefs[RESTPouet]
		},
		"restZoo": func() string {
			return hrefs[RESTZoo]
		},
		"sriBootCSS": func() string {
			return web.Subresource.Bootstrap
		},
		"sriBootJS": func() string {
			return web.Subresource.BootstrapJS
		},
		"sriEditorArchiveForm": func() string {
			return web.Subresource.EditArchive
		},
		"sriEditorAssetsForm": func() string {
			return web.Subresource.EditAssets
		},
		"sriEditorForm": func() string {
			return web.Subresource.Editor
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
		"sriRestPouet": func() string {
			return web.Subresource.RESTPouet
		},
		"sriRestZoo": func() string {
			return web.Subresource.RESTZoo
		},
		"sriUploader": func() string {
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
		"fmtURI":            releaser.Link, // this is not performant for large lists, instead use fmtFastURI in TemplateStrings()
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
		"websiteIcon":       WebsiteIcon,
	}
	return funcMap
}
