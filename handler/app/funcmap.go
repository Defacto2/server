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
		"sriJSDos": func() string {
			return web.Subresource.JSDos
		},
		"sriJSWDos": func() string {
			return web.Subresource.JSWDos
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
		"sriUploader": func() string {
			return web.Subresource.UploaderJS
		},
		"cssBoot": func() string {
			return BootCSS
		},
		"cssLayout": func() string {
			return LayoutCSS
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
			return BootJS
		},
		"jsDos": func() string {
			return JSDos
		},
		"jsWDos": func() string {
			return JSWDos
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
		"version": func() string {
			return web.Version
		},
		"uploader": func() string {
			return UploaderJS
		},
		"editorForm": func() string {
			return EditorJS
		},
		"sriEditorForm": func() string {
			return web.Subresource.EditorJS
		},
		"editorAssetsForm": func() string {
			return EditorAssetsJS
		},
		"sriEditorAssetsForm": func() string {
			return web.Subresource.EditorAssetsJS
		},
		"editorArchiveForm": func() string {
			return EditorArchiveJS
		},
		"sriEditorArchiveForm": func() string {
			return web.Subresource.EditorArchiveJS
		},
		"restPouet": func() string {
			return RestPouetJS
		},
		"sriRestPouet": func() string {
			return web.Subresource.RestPouetJS
		},
		"restZoo": func() string {
			return RestZooJS
		},
		"sriRestZoo": func() string {
			return web.Subresource.RestZooJS
		},
		//"recordReleasers": RecordRels,
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
		"recordInfoOSTag":   InfoOSTag,
		"recordTagInfo":     TagInfoX,
	}
}
