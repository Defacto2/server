package app_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/nalgeon/be"
)

func TestHrefs(t *testing.T) {
	t.Parallel()

	paths := app.Hrefs()
	be.Equal(t, len(paths), app.Assets)

	for _, s := range paths {
		got := strings.HasPrefix(s, "/js/") ||
			strings.HasPrefix(s, "/css/") ||
			strings.HasPrefix(s, "/svg/")
		be.True(t, got)
	}

	names := app.Names()
	be.Equal(t, len(names), app.Assets)
	for _, s := range names {
		be.True(t, strings.HasPrefix(s, "public/"))
	}
	be.Equal(t, names[0], "public/css/bootstrap.min.css")
}

func TestFonts(t *testing.T) {
	t.Parallel()

	fonts := app.FontRefs()
	be.Equal(t, len(fonts), app.Fonts)

	for _, s := range fonts {
		got := strings.HasPrefix(s, "/topazplus_a1200.") ||
			strings.HasPrefix(s, "/MicroKnightPlus_v1.0.") ||
			strings.HasPrefix(s, "/CascadiaMono.") ||
			strings.HasPrefix(s, "/Ac437_IBM_")
		be.True(t, got)
	}

	names := app.FontNames()
	be.Equal(t, len(names), app.Fonts)
	for _, s := range names {
		be.True(t, strings.HasPrefix(s, "public/font/"))
	}
	be.Equal(t, names[0], "public/font/topazplus_a1200.woff2")
}
