package app_test

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

func TestHrefs(t *testing.T) {
	t.Parallel()

	fsys := testutil.ProjectFS(t)
	paths := app.Hrefs()
	names := app.Names()
	be.Equal(t, len(paths), app.Assets)
	be.Equal(t, len(names), app.Assets)

	seen := make(map[string]bool, app.Assets)
	for i := range app.Assets {
		key := app.Asset(i)
		href, ok := paths[key]
		be.True(t, ok)
		be.True(t, strings.HasPrefix(href, "/js/") ||
			strings.HasPrefix(href, "/css/") ||
			strings.HasPrefix(href, "/svg/"))
		be.True(t, !seen[href])
		seen[href] = true

		name := names[key]
		be.Equal(t, name, "public"+href)
		info, err := fs.Stat(fsys, name)
		be.Err(t, err, nil)
		be.True(t, info.Size() > 0)
	}
	be.Equal(t, names[app.Bootstrap5], "public/css/bootstrap.min.css")
}

func TestFonts(t *testing.T) {
	t.Parallel()

	fsys := testutil.ProjectFS(t)
	fonts := app.FontRefs()
	names := app.FontNames()
	be.Equal(t, len(fonts), app.Fonts)
	be.Equal(t, len(names), app.Fonts)

	seen := make(map[string]bool, app.Fonts)
	for i := range app.Fonts {
		key := app.Font(i)
		ref, ok := fonts[key]
		be.True(t, ok)
		be.True(t, strings.HasPrefix(ref, "/topazplus_a1200.") ||
			strings.HasPrefix(ref, "/MicroKnightPlus_v1.0.") ||
			strings.HasPrefix(ref, "/CascadiaMono.") ||
			strings.HasPrefix(ref, "/Ac437_IBM_"))
		be.True(t, !seen[ref])
		seen[ref] = true

		name := names[key]
		be.Equal(t, name, "public/font"+ref)
		info, err := fs.Stat(fsys, name)
		be.Err(t, err, nil)
		be.True(t, info.Size() > 0)
	}
	be.Equal(t, names[app.A1200], "public/font/topazplus_a1200.woff2")
}
