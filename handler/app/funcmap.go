package app

// Package file funcmap.go contains the custom template functions for the web framework.
// The functions are used by the HTML templates to format data.

import (
	"html/template"

	"github.com/Defacto2/server/pkg/fmts"
	"github.com/Defacto2/server/pkg/helper"
	"github.com/Defacto2/server/pkg/initialism"
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
		"fmtByte":        ByteCount,
		"fmtByteCnt":     ByteFile,
		"fmtByteName":    ByteFileS,
		"fmtDay":         Day,
		"fmtInitalism":   initialism.Join,
		"fmtMonth":       Month,
		"fmtPrefix":      Prefix,
		"fmtRoles":       helper.FmtSlice,
		"fmtURI":         fmts.Name,
		"lastUpdated":    LastUpdated,
		"linkDownload":   LinkDownload,
		"linkPage":       LinkPage,
		"linkPreview":    LinkPreview,
		"linkRemote":     LinkRemote,
		"linkRelrs":      LinkRelrs,
		"linkScnr":       LinkScnr,
		"linkWiki":       LinkWiki,
		"logoText":       LogoText,
		"mod3":           Mod3,
		"safeHTML":       SafeHTML,
		"screenshot":     web.Screenshot,
		"sizeOfDL":       DownloadB,
		"subTitle":       SubTitle,
		"thumb":          web.Thumb,
		"trimSiteSuffix": TrimSiteSuffix,
		"updated":        Updated,
		"websiteIcon":    WebsiteIcon,
		// these closures should only return simple values
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
	}
}
