package font

import _ "embed"

//go:embed CascadiaMono.ttf
var cascadiaMono []byte

// CascadiaMono returns the embedded Cascadia Mono TrueType font data.
func CascadiaMono() []byte {
	return cascadiaMono
}
