package demozoo_test

import (
	"testing"

	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/internal/tags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Set to true to test against the remote servers.
const testRemoteServers = false

func TestDemozoo_Get(t *testing.T) {
	t.Parallel()
	prod := demozoo.Production{}
	_, err := prod.Get(-1)
	require.Error(t, err)
	require.ErrorIs(t, err, demozoo.ErrID)

	if !testRemoteServers {
		return
	}

	_, err = prod.Get(1)
	require.NoError(t, err)
	require.ErrorIs(t, err, demozoo.ErrSuccess)
}

func TestFind(t *testing.T) {
	t.Parallel()
	prod := demozoo.Find("defacto2")
	want := demozoo.GroupID(10000)
	assert.Equal(t, want, prod)

	prod = demozoo.Find("notfound")
	assert.Equal(t, prod, demozoo.GroupID(0))
}

func TestExternalLinks(t *testing.T) {
	t.Parallel()
	d := demozoo.Production{}
	assert.Equal(t, 0, d.PouetProd())

	d.ExternalLinks = append(d.ExternalLinks, struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	}{
		LinkClass: "class1",
		URL:       "http://example.com/1",
	})
	assert.Equal(t, 0, d.PouetProd())

	d.ExternalLinks = append(d.ExternalLinks, struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	}{
		LinkClass: "PouetProduction",
		URL:       "http://example.com/1",
	})
	assert.Equal(t, 0, d.PouetProd())

	d.ExternalLinks = append(d.ExternalLinks, struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	}{
		LinkClass: "PouetProduction",
		URL:       "https://www.pouet.net/prod.php?which=71562",
	})
	assert.Equal(t, 71562, d.PouetProd())
	assert.Empty(t, d.GithubRepo())

	d.ExternalLinks = append(d.ExternalLinks, struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	}{
		LinkClass: "GithubRepo",
		URL:       "https://github.com/Defacto2/server",
	})
	assert.Equal(t, "/Defacto2/server", d.GithubRepo())

	assert.Empty(t, d.YouTubeVideo())
	d.ExternalLinks = append(d.ExternalLinks, struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	}{
		LinkClass: "YoutubeVideo",
		URL:       "https://www.youtube.com/watch?v=x6QrKsBOERA",
	})
	assert.Equal(t, "x6QrKsBOERA", d.YouTubeVideo())
}

func TestUnmarshal(t *testing.T) {
	t.Parallel()
	prod := demozoo.Production{}
	err := prod.Unmarshal(nil)
	require.NoError(t, err)
}

func TestSuperType(t *testing.T) {
	t.Parallel()
	prod := demozoo.Production{}
	x, y := prod.SuperType()
	const want tags.Tag = -1
	assert.Equal(t, want, x)
	assert.Equal(t, want, y)
}

func TestReleased(t *testing.T) {
	t.Parallel()
	prod := demozoo.Production{}
	y, m, d := prod.Released()
	const want int16 = 0
	assert.Equal(t, want, y)
	assert.Equal(t, want, m)
	assert.Equal(t, want, d)
}

func TestGroups(t *testing.T) {
	t.Parallel()
	prod := demozoo.Production{}
	a, b := prod.Groups()
	assert.Empty(t, a)
	assert.Empty(t, b)
}

func TestSite(t *testing.T) {
	t.Parallel()
	s := demozoo.Site("")
	assert.Empty(t, s)
	s = demozoo.Site("the cool bbs")
	assert.Equal(t, "cool BBS", s)
	s = demozoo.Site("Cool BBS")
	assert.Equal(t, "Cool BBS", s)
}

func TestReleasers(t *testing.T) {
	t.Parallel()
	prod := demozoo.Production{}
	a, b, c, d := prod.Releasers()
	assert.Empty(t, a)
	assert.Empty(t, b)
	assert.Empty(t, c)
	assert.Empty(t, d)
}

func TestCategory(t *testing.T) {
	t.Parallel()
	c := demozoo.TextC.String()
	assert.Equal(t, "text", c)
	c = demozoo.CodeC.String()
	assert.Equal(t, "code", c)
	c = demozoo.GraphicsC.String()
	assert.Equal(t, "graphics", c)
	c = demozoo.MusicC.String()
	assert.Equal(t, "music", c)
	c = demozoo.MagazineC.String()
	assert.Equal(t, "magazine", c)
}
