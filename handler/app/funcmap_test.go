package app_test

import (
	"embed"
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/stretchr/testify/assert"
)

const (
	exampleURL  = "https://example.com"
	exampleWiki = "/some/wiki/page"
)

func TestExternalLink(t *testing.T) {
	t.Parallel()
	x := app.ExternalLink("", "")
	assert.Contains(t, x, "error:")
	x = app.ExternalLink(exampleURL, "")
	assert.Contains(t, x, "error:")
	x = app.ExternalLink(exampleURL, "Example")
	assert.Contains(t, x, exampleURL)
}

func TestWikiLink(t *testing.T) {
	t.Parallel()
	x := app.WikiLink("", "")
	assert.Contains(t, x, "error:")
	x = app.WikiLink(exampleWiki, "")
	assert.Contains(t, x, "error:")
	x = app.WikiLink(exampleWiki, "Example")
	assert.Contains(t, x, exampleWiki)
}

func TestIntegrity(t *testing.T) {
	t.Parallel()
	x, err := app.Integrity("", embed.FS{})
	assert.Error(t, err)
	assert.Empty(t, x)
}

func TestIntegrityBytes(t *testing.T) {
	t.Parallel()
	x := app.IntegrityBytes(nil)
	assert.Equal(t, "sha384-OLBgp1GsljhM2TJ+sbHjaiH9txEUvgdDTAzHv2P24donTt6/529l+9Ua0vFImLlb", x)
	x = app.IntegrityBytes([]byte("hello world"))
	assert.Equal(t, "sha384-/b2OdaZ/KfcBpOBAOF4uI5hjA+oQI5IRr5B/y7g1eLPkF8txzmRu/QgZ3YwIjeG9", x)
}

func TestLogoText(t *testing.T) {
	t.Parallel()
	x := app.LogoText("")
	assert.Equal(t, app.Welcome, x)
	x = app.LogoText("X")
	assert.Equal(t, app.Welcome, x)
	x = app.LogoText("XY")
	assert.Equal(t, app.Welcome, x)
}
