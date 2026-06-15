// Package releaser provides string functions for cleaning and reformatting
// the names of release groups and partial URL paths.
package releaser

import (
	"maps"
	"slices"
	"strings"

	"github.com/Defacto2/server/handler/releaser/fix"
	"github.com/Defacto2/server/handler/releaser/initialism"
	"github.com/Defacto2/server/handler/releaser/name"
)

// initialisms are a cache of that greatly improves benchmark performance.
var initialisms = initialism.Initialisms() //nolint:gochecknoglobals

// specials are a cache of that greatly improves benchmark performance.
var specials = name.Special() //nolint:gochecknoglobals

// Cell formats the string to be used as a cell in a database table.
//
//   - The removal of duplicate spaces
//   - The removal of excess whitespace
//   - If found "The " prefix from BBS and FTP named sites
//   - The stripping of incompatible characters
//
// Compatible characters include: A-Z a-z À-Ö Ø-ö ø-ÿ 0-9 - , &
//
// Example:
//
//	Cell("  Defacto2  demo  group.") = "DEFACTO2 DEMO GROUP"
//	Cell("the x bbs") = "X BBS"
//	Cell("defacto2.net") = "DEFACTO2NET"
//	Cell("TDT / TRSi") = "TDT TRSI"
//	Cell("TDT,TRSi") = "TDT, TRSI"
func Cell(s string) string {
	x := fix.StripChars(s)
	x = fix.StripStart(x)
	x = strings.TrimSpace(x)
	x = fix.TrimThe(x)
	x = fix.TrimSP(x)
	return fix.Cell(x)
}

// Clean fixes the malformed string and applies title case formatting.
// It does not apply any name deobfuscations such as initials or abbreviations,
// as it only stylizes the string.
//
//   - The removal of duplicate spaces
//   - The removal of excess whitespace
//   - If found "The " prefix from BBS and FTP named sites
//   - The stripping of incompatible characters
//
// Compatible characters include: A-Z a-z À-Ö Ø-ö ø-ÿ 0-9 - , &
//
// Example:
//
//	Clean("  Defacto2  demo  group.") = "Defacto2 Demo Group"
//	Clean("the x bbs") = "X BBS"
//	Clean("The X Ftp") = "X FTP"
//	Clean("tdt / trsi") = "Tdt Trsi" // behaves as a single group
//	Clean("tdt,trsi") = "Tdt, TRSi"  // behaves as two groups
func Clean(s string) string {
	x := fix.StripChars(s)
	x = fix.StripStart(x)
	x = strings.TrimSpace(x)
	x = fix.TrimThe(x)
	x = fix.TrimSP(x)
	return fix.Format(x)
}

// Humanize deobfuscates the URL path and returns the formatted, human-readable group name.
// The path is expected to be in the format of a URL path without the scheme or domain.
// If the URL path contains invalid characters then an empty string is returned.
//
// Example:
//
//	Humanize("defacto2") = "Defacto2"
//	Humanize("razor-1911-demo") = "Razor 1911 Demo"
//	Humanize("razor-1911-demo-ampersand-skillion") = "Razor 1911 Demo & Skillion"
//	Humanize("north-american-pirate_phreak-association") = "North American Pirate-Phreak Association"
//	Humanize("coop") = "TDT / TRSi"
//	Humanize("united-software-association*fairlight") =
//		"United Software Association + Fairlight PC Division" // special name
//	Humanize("razor-1911-demo*trsi") = "Razor 1911 Demo, TRSi"
//	Humanize("razor-1911-demo#trsi") = "" // invalid # character
func Humanize(path string) string {
	p := name.Path(strings.ToLower(path))
	if special := p.String(); special != "" {
		return special
	}
	s, err := name.Humanize(p)
	if err != nil {
		return ""
	}
	return Clean(s)
}

// Index deobfuscates the URL path and applies [releaser.Humanize] so that it can
// be stored in a database table as a releaser key and index in the database table.
func Index(path string) string {
	p := name.Path(strings.ToLower(path))
	s, err := name.Humanize(p)
	if err != nil {
		return ""
	}
	return strings.ToUpper(s)
}

// Link deobfuscates the URL path and applies [releaser.Humanize].
// In addition, the humanized name is formatted to be used as a link description.
// If the URL path contains invalid characters then an empty string is returned.
//
// Example:
//
//	Link("razor-1911-demo*trsi") = "Razor 1911 Demo + TRSi"
//	Link("class*paradigm*razor-1911") = "Class + Paradigm + Razor 1911"
//	Link("united-software-association*fairlight") = "United Software Association + Fairlight PC Division"
func Link(path string) string {
	s := Humanize(path)
	return strings.ReplaceAll(s, ", ", " + ")
}

// Obfuscate cleans and formats the string for use as a URL path.
// The string is expected to be a release group name or an known initialism, acronym or special name.
//
// Beware that initialisms and acronyms often are not unique and an unexpected URL may be returned.
//
// Example:
//
//	Obfuscate("ACiD Productions") = "acid-productions"
//	Obfuscate("Razor 1911 Demo & Skillion") = "razor-1911-demo-ampersand-skillion"
//	Obfuscate("TDU-Jam!") = "tdu_jam"
//	Obfuscate("The 12AM BBS.") = "12am-bbs"
//
// Examples using unique, known initialisms:
//
//	Obfuscate("fltdox") = "fairlight-dox"
//	Obfuscate("tdt") = "the-dream-team"
//
// Examples using special names:
//
//	Obfuscate("TDT / TRSi") = "coop"
//	Obfuscate("United Software Association + Fairlight PC Division") = "united-software-association*fairlight"
func Obfuscate(s string) string {
	x := fix.StripStart(s)
	x = strings.TrimSpace(x)
	for uri, special := range maps.All(*specials) {
		if strings.EqualFold(x, special) {
			return string(uri)
		}
	}
	for uri, values := range maps.All(*initialisms) {
		for value := range slices.Values(values) {
			if strings.EqualFold(x, value) {
				return string(uri)
			}
		}
	}
	x = fix.StripChars(x)
	x = fix.TrimThe(x)
	x = fix.TrimSP(x)
	c := name.Obfuscate(x)
	return string(c)
}

// Title formats the string to be used as a title or the basis for a LIKE SQL query.
// Any known initialisms, acronyms or special names are deobfuscated.
//
// Example:
//
//	Title("razor 1911") = "Razor 1911"
//	Title("_.=[   RaZoR 1911   ]=._") = "Razor 1911"
//	Title("COOP") = "TDT / TRSi"
//	Title("tdt / trsi") = "TDT / TRSi"
//	Title("nappa") = "North American Pirate-Phreak Association"
func Title(s string) string {
	x := fix.StripStart(s)
	x = strings.TrimSpace(x)
	for _, special := range *specials {
		if strings.EqualFold(x, special) {
			return special
		}
	}
	for uri, values := range maps.All(*initialisms) {
		for value := range slices.Values(values) {
			if strings.EqualFold(x, value) {
				return Humanize(string(uri))
			}
		}
	}
	x = fix.StripChars(x)
	x = fix.TrimThe(x)
	x = fix.TrimSP(x)
	c := name.Obfuscate(x)
	return Humanize(string(c))
}
