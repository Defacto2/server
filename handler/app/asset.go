package app

// Package file asset.go contains the public facing JS, CSS, SVG, font files and the executable WASM file paths.
// The paths are relative to the public/ facing directory and are intended for use in HTML link and script elements.
// All assets listed here must be embedded into the binary using the "go:embed" directive.

// Asset is a relative path to a public facing CSS, JS or WASM file.
type Asset int

const (
	Bootstrap5      Asset = iota // Bootstrap is the path to the minified Bootstrap 5.3 CSS file.
	Bootstrap5JS                 // BootstrapJS is the path to the minified Bootstrap 5.3 JS file.
	BootstrapIcons               // BootstrapIcons is the path to the custom Bootstrap Icons SVG sprites file.
	DosboxJS                     // DosboxJS is the js-dos v6 default variant compiled with emscripten.
	DosboxWasm                   // DosboxWasm is the js-dos v6 WASM binary file.
	EditArtifact                 // EditArtifact is the path to the minified Artifact Editor JS file.
	EditAssets                   // EditAssets is the path to the minified Editor assets JS file.
	EditForApproval              // EditForApproval is the path to the minified Editor for-approval JS file.
	Htmx                         // Htmx is the path to the minified htmx AJAX JS file.
	HtmxRespTargets              // Htmx is the path to the minified response targets extension file.
	Jsdos6JS                     // Jsdos6JS is the path to the minified js-dos v6 JS file.
	Layout                       // Layout is the path to the minified layout CSS file.
	LayoutJS                     // LayoutJS is the path to the minified layout JS file.
	Pouet                        // Pouet is the path to the minified Pouet JS file.
	Readme                       // Readme is the path to the minified Readme JS file.
	Uploader                     // Uploader is the path to the minified Uploader JS file.
)

// Paths are a map of the public facing CSS, JS and WASM files.
type Paths map[Asset]string

// Hrefs returns the relative path of the public facing CSS, JS and WASM files.
// The strings are intended for href attributes in HTML link elements and
// the src attribute in HTML script elements.
func Hrefs() *Paths {
	return &Paths{
		Bootstrap5:      "/css/bootstrap.min.css",
		Bootstrap5JS:    "/js/bootstrap.bundle.min.js",
		BootstrapIcons:  "/svg/bootstrap-icons.svg",
		DosboxJS:        "/js/wdosbox.js",
		DosboxWasm:      "/js/wdosbox.wasm",
		EditArtifact:    "/js/editor-artifact.min.js",
		EditAssets:      "/js/editor-assets.min.js",
		EditForApproval: "/js/editor-forapproval.min.js",
		Htmx:            "/js/htmx.min.js",
		HtmxRespTargets: "/js/htmx-response-targets.min.js",
		Jsdos6JS:        "/js/js-dos.js",
		Layout:          "/css/layout.min.css",
		LayoutJS:        "/js/layout.min.js",
		Pouet:           "/js/votes-pouet.min.js",
		Readme:          "/js/readme.min.js",
		Uploader:        "/js/uploader.min.js",
	}
}

// Names returns the absolute path of the public facing CSS, JS and WASM files
// relative to the [embed.FS] root.
func Names() *Paths {
	const public = "public"
	hrefs := Hrefs()
	paths := make(Paths, len(*hrefs))
	for key, href := range *hrefs {
		paths[key] = public + href
	}
	return &paths
}

// Fonts are a map of the public facing font files.
type Fonts map[Font]string

// Font is a relative path to a public facing font file.
type Font int

const (
	VGA8             Font = iota // VGA8 is the path to the IBM VGA 8px font file.
	VGA8Woff                     // VGA8Woff is the path to the IBM VGA 8px legacy WOFF format font file.
	VGA8TT                       // VGA8TT is the path to the IBM VGA 8px legacy TrueType format font file.
	A1200                        // A1200 is the path to the Topaz Plus font file.
	A1200Woff                    // A1200Woff is the path to the Topaz Plus legacy WOFF format font file.
	A1200TT                      // A1200TT is the path to the Topaz Plus legacy TrueType format font file.
	CascadiaMono                 // CascadiaMono is the path to the Cascadia Mono font file.
	CascadiaMonoWoff             // CascadiaMonoWoff is the path to the Cascadia Mono WOFF format font file.
	CascadiaMonoTT               // CascadiaMonoTT is the path to the Cascadia Mono TrueType format font file.
)

// FontNames returns the absolute path of the public facing font files
// relative to the embed.FS root.
func FontNames() *Fonts {
	const public = "public/font"
	hrefs := FontRefs()
	paths := make(Fonts, len(*hrefs))
	for key, href := range *hrefs {
		paths[key] = public + href
	}
	return &paths
}

// FontRefs returns the relative path of the public facing font files.
// The strings are intended for href attributes in HTML link elements and
// the src attribute in HTML script elements.
func FontRefs() *Fonts {
	return &Fonts{
		VGA8:             "/pxplus_ibm_vga8.woff2",
		VGA8Woff:         "/pxplus_ibm_vga8.woff",
		VGA8TT:           "/pxplus_ibm_vga8.ttf",
		A1200:            "/topazplus_a1200.woff2",
		A1200Woff:        "/topazplus_a1200.woff",
		A1200TT:          "/topazplus_a1200.ttf",
		CascadiaMono:     "/CascadiaMono.woff2",
		CascadiaMonoWoff: "/CascadiaMono.woff",
		CascadiaMonoTT:   "/CascadiaMono.ttf",
	}
}
