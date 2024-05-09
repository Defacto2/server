package app

// Fonts are a map of the public facing font files.
type Fonts map[Font]string

// Font is a relative path to a public facing font file.
type Font int

const (
	VGA8         Font = iota // VGA8 is the path to the IBM VGA 8px font file.
	VGA8Woff                 // VGA8Woff is the path to the IBM VGA 8px legacy WOFF format font file.
	VGA8TT                   // VGA8TT is the path to the IBM VGA 8px legacy TrueType format font file.
	A1200                    // A1200 is the path to the Topaz Plus font file.
	A1200Woff                // A1200Woff is the path to the Topaz Plus legacy WOFF format font file.
	A1200TT                  // A1200TT is the path to the Topaz Plus legacy TrueType format font file.
	OpenSans                 // OpenSans is the path to the Open Sans font file.
	OpenSansWoff             // OpenSansWoff is the path to the Open Sans WOFF format font file.
	OpenSansTT               // OpenSansTT is the path to the Open Sans TrueType format font file.
)

// Names returns the absolute path of the public facing font files
// relative to the embed.FS root.
func FontNames() Fonts {
	const public = "public/font"
	href := FontRefs()
	paths := make(Fonts, len(href))
	for k, v := range href {
		paths[k] = public + v
	}
	return paths
}

// FontRefs returns the relative path of the public facing font files.
// The strings are intended for href attributes in HTML link elements and
// the src attribute in HTML script elements.
func FontRefs() Fonts {
	return Fonts{
		VGA8:         "/pxplus_ibm_vga8.woff2",
		VGA8Woff:     "/pxplus_ibm_vga8.woff",
		VGA8TT:       "/pxplus_ibm_vga8.ttf",
		A1200:        "/topazplus_a1200.woff2",
		A1200Woff:    "/topazplus_a1200.woff",
		A1200TT:      "/topazplus_a1200.ttf",
		OpenSans:     "/opensans_variable.woff2",
		OpenSansWoff: "/opensans_variable.woff",
		OpenSansTT:   "/opensans_variable.ttf",
	}
}
