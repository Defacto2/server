package app_test

import (
	"strings"
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
	x := app.LinkRemote("", "")
	assert.Contains(t, x, "error:")
	x = app.LinkRemote(exampleURL, "")
	assert.Contains(t, x, "error:")
	x = app.LinkRemote(exampleURL, "Example")
	assert.Contains(t, x, exampleURL)
}

func TestWikiLink(t *testing.T) {
	t.Parallel()
	x := app.LinkWiki("", "")
	assert.Contains(t, x, "error:")
	x = app.LinkWiki(exampleWiki, "")
	assert.Contains(t, x, "error:")
	x = app.LinkWiki(exampleWiki, "Example")
	assert.Contains(t, x, exampleWiki)
}

func TestLogoText(t *testing.T) {
	t.Parallel()
	const want1 = "      :                             ·· X ··                             ·"
	const want2 = "      :                             ·· XY ··                            ·"
	const want3 = "      :                            ·· XYZ ··                            ·"
	const wantR = "      : ·· I'M MEANT TO BE WRITING AT THIS MOMENT. WHAT I MEAN IS, I ·· ·"
	x := app.LogoText("")
	want := strings.Repeat(" ", 7) + app.Welcome
	assert.Equal(t, want, x)
	x = app.LogoText("X")
	assert.Equal(t, want1, x)
	x = app.LogoText("XY")
	assert.Equal(t, want2, x)
	x = app.LogoText("xyz")
	assert.Equal(t, want3, x)
	const rand = "I'm meant to be writing at this moment. " +
		"What I mean is, I'm meant to be writing something else at this moment."
	x = app.LogoText(rand)
	assert.Equal(t, wantR, x)
}
