package app

// Asset is a relative path to a public facing CSS, JS or WASM file.
type Asset int

const (
	Bootstrap   Asset = iota // Bootstrap is the path to the minified Bootstrap 5.3 CSS file.
	BootstrapJS              // BootstrapJS is the path to the minified Bootstrap 5.3 JS file.
	Editor                   // Editor is the path to the minified Editor JS file.
	EditAssets               // EditAssets is the path to the minified Editor assets JS file.
	EditArchive              // EditArchive is the path to the minified Editor archive JS file.
	FontAwesome              // FontAwesome is the path to the minified Font Awesome 3 JS file.
	JSDosUI                  // JSDosUI is the path to the minified JS DOS user-interface JS file.
	JSDosW                   // JSDosW is the JS DOS default variant compiled with emscripten.
	JSDosWasm                // JSDOSWasm is the JS DOS WASM binary file.
	Layout                   // Layout is the path to the minified layout CSS file.
	Pouet                    // Pouet is the path to the minified Pouet JS file.
	Readme                   // Readme is the path to the minified Readme JS file.
	RESTPouet                // RESTPouet is the path to the minified Pouet REST JS file.
	RESTZoo                  // RESTZoo is the path to the minified Demozoo REST JS file.
	Uploader                 // Uploader is the path to the minified Uploader JS file.
)

// Paths are a map of the public facing CSS, JS and WASM files.
type Paths map[Asset]string

// Hrefs returns the relative path of the public facing CSS, JS and WASM files.
// The strings are intended for href attributes in HTML link elements and
// the src attribute in HTML script elements.
func Hrefs() Paths {
	// note, the js-dos (JS DOS v6) are minified files,
	// help: https://js-dos.com/6.22/examples/?arkanoid
	return Paths{
		Bootstrap:   "/css/bootstrap.min.css",
		BootstrapJS: "/js/bootstrap.bundle.min.js",
		Editor:      "/js/editor.min.js",
		EditAssets:  "/js/editor-assets.min.js",
		EditArchive: "/js/editor-archive.min.js",
		FontAwesome: "/js/fontawesome.min.js",
		JSDosW:      "/js/wdosbox.js",
		JSDosWasm:   "/js/wdosbox.wasm",
		JSDosUI:     "/js/js-dos.js",
		Layout:      "/css/layout.min.css",
		Pouet:       "/js/pouet.min.js",
		Readme:      "/js/readme.min.js",
		RESTPouet:   "/js/rest-pouet.min.js",
		RESTZoo:     "/js/rest-zoo.min.js",
		Uploader:    "/js/uploader.min.js",
	}
}

// Names returns the absolute path of the public facing CSS, JS and WASM files
// relative to the embed.FS root.
func Names() Paths {
	const public = "public"
	href := Hrefs()
	// iterate and return HRefs() with the public prefix
	paths := make(Paths, len(href))
	for k, v := range href {
		paths[k] = public + v
	}
	return paths
}
