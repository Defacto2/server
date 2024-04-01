package app

// Asset is a relative path to a public facing CSS, JS or WASM file.
type Asset int

const (
	Bootstrap5      Asset = iota // Bootstrap is the path to the minified Bootstrap 5.3 CSS file.
	Bootstrap5JS                 // BootstrapJS is the path to the minified Bootstrap 5.3 JS file.
	DosboxJS                     // DosboxJS is the js-dos v6 default variant compiled with emscripten.
	DosboxWasm                   // DosboxWasm is the js-dos v6 WASM binary file.
	Editor                       // Editor is the path to the minified Editor JS file.
	EditAssets                   // EditAssets is the path to the minified Editor assets JS file.
	EditArchive                  // EditArchive is the path to the minified Editor archive JS file.
	EditForApproval              // EditForApproval is the path to the minified Editor for-approval JS file.
	FA5Pro                       // FA5Pro is the path to the minified Font Awesome Pro v5 JS file.
	Htmx                         // Htmx is the path to the minified htmx AJAX JS file.
	Jsdos6JS                     // Jsdos6JS is the path to the minified js-dos v6 JS file.
	Layout                       // Layout is the path to the minified layout CSS file.
	Pouet                        // Pouet is the path to the minified Pouet JS file.
	Readme                       // Readme is the path to the minified Readme JS file.
	Uploader                     // Uploader is the path to the minified Uploader JS file.
)

// Paths are a map of the public facing CSS, JS and WASM files.
type Paths map[Asset]string

// Hrefs returns the relative path of the public facing CSS, JS and WASM files.
// The strings are intended for href attributes in HTML link elements and
// the src attribute in HTML script elements.
func Hrefs() Paths {
	return Paths{
		Bootstrap5:      "/css/bootstrap.min.css",
		Bootstrap5JS:    "/js/bootstrap.bundle.min.js",
		DosboxJS:        "/js/wdosbox.js",
		DosboxWasm:      "/js/wdosbox.wasm",
		Editor:          "/js/editor.min.js",
		EditAssets:      "/js/editor-assets.min.js",
		EditArchive:     "/js/editor-archive.min.js",
		EditForApproval: "/js/editor-forapproval.min.js",
		FA5Pro:          "/js/fontawesome.min.js",
		Htmx:            "/js/htmx.min.js",
		Jsdos6JS:        "/js/js-dos.js",
		Layout:          "/css/layout.min.css",
		Pouet:           "/js/pouet.min.js",
		Readme:          "/js/readme.min.js",
		Uploader:        "/js/uploader.min.js",
	}
}

// Names returns the absolute path of the public facing CSS, JS and WASM files
// relative to the [embed.FS] root.
func Names() Paths {
	const public = "public"
	href := Hrefs()
	paths := make(Paths, len(href))
	for k, v := range href {
		paths[k] = public + v
	}
	return paths
}
