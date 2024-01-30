package app_test

import (
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/stretchr/testify/assert"
)

func TestAsset(t *testing.T) {
	t.Parallel()

	x, y := app.Bootstrap, app.Uploader
	assert.Equal(t, app.Asset(0), x)
	assert.Equal(t, app.Asset(14), y)

	hrefs := app.Hrefs()
	for i, href := range hrefs {
		assert.NotEmpty(t, href)
		switch i {
		case 0, 9:
			ext := href[len(href)-8:]
			assert.Equal(t, ".min.css", ext)
		case 6, 7, 8:
		default:
			ext := href[len(href)-7:]
			assert.Equal(t, ".min.js", ext)
		}
	}
}

func TestNames(t *testing.T) {
	t.Parallel()

	x := app.Names()
	assert.Equal(t, "public/css/bootstrap.min.css", x[0])
}

func TestFontRefs(t *testing.T) {
	t.Parallel()

	x := app.FontRefs()
	assert.Equal(t, "/pxplus_ibm_vga8.woff2", x[app.VGA8])

	n := app.FontNames()
	assert.Equal(t, "public/font/pxplus_ibm_vga8.woff2", n[app.VGA8])
}

func TestGlobTo(t *testing.T) {
	t.Parallel()

	x := app.GlobTo("file.css")
	assert.Equal(t, "view/app/file.css", x)
}

func TestTemplates(t *testing.T) {
	t.Parallel()

	w := app.Web{}
	_, err := w.Templates()
	assert.Error(t, err)
}
