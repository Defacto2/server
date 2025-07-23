package demozoo_test

import (
	"errors"
	"testing"

	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/internal/tags"
	"github.com/nalgeon/be"
)

// Set to true to test against the remote servers.
const testRemoteServers = false

func TestDemozoo_Get(t *testing.T) {
	t.Parallel()
	prod := demozoo.Production{}
	_, err := prod.Get(-1)
	be.Err(t, err)
	ok := errors.Is(err, demozoo.ErrID)
	be.True(t, ok)
	if !testRemoteServers {
		return
	}

	_, err = prod.Get(1)
	be.Err(t, err, nil)
	ok = errors.Is(err, demozoo.ErrSuccess)
	be.True(t, ok)
}

func TestFind(t *testing.T) {
	t.Parallel()
	prod := demozoo.Find("defacto2")
	want := demozoo.GroupID(10000)
	be.Equal(t, want, prod)

	prod = demozoo.Find("notfound")
	be.Equal(t, prod, demozoo.GroupID(0))
}

func TestExternalLinks(t *testing.T) {
	t.Parallel()
	d := demozoo.Production{}
	be.Equal(t, 0, d.PouetProd())

	d.ExternalLinks = append(d.ExternalLinks, struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	}{
		LinkClass: "class1",
		URL:       "http://example.com/1",
	})
	be.Equal(t, 0, d.PouetProd())

	d.ExternalLinks = append(d.ExternalLinks, struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	}{
		LinkClass: "PouetProduction",
		URL:       "http://example.com/1",
	})
	be.Equal(t, 0, d.PouetProd())

	d.ExternalLinks = append(d.ExternalLinks, struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	}{
		LinkClass: "PouetProduction",
		URL:       "https://www.pouet.net/prod.php?which=71562",
	})
	be.Equal(t, 71562, d.PouetProd())
	be.Equal(t, d.GithubRepo(), "")

	d.ExternalLinks = append(d.ExternalLinks, struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	}{
		LinkClass: "GithubRepo",
		URL:       "https://github.com/Defacto2/server",
	})
	be.Equal(t, "/Defacto2/server", d.GithubRepo())
	be.Equal(t, d.YouTubeVideo(), "")
	d.ExternalLinks = append(d.ExternalLinks, struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	}{
		LinkClass: "YoutubeVideo",
		URL:       "https://www.youtube.com/watch?v=x6QrKsBOERA",
	})
	be.Equal(t, "x6QrKsBOERA", d.YouTubeVideo())
}

func TestUnmarshal(t *testing.T) {
	t.Parallel()
	prod := demozoo.Production{}
	err := prod.Unmarshal(nil)
	be.Err(t, err, nil)
}

func TestSuperType(t *testing.T) {
	t.Parallel()
	prod := demozoo.Production{}
	x, y := prod.SuperType()
	const want tags.Tag = -1
	be.Equal(t, want, x)
	be.Equal(t, want, y)
}

func TestReleased(t *testing.T) {
	t.Parallel()
	prod := demozoo.Production{}
	y, m, d := prod.Released()
	const want int16 = 0
	be.Equal(t, want, y)
	be.Equal(t, want, m)
	be.Equal(t, want, d)
}

func TestGroups(t *testing.T) {
	t.Parallel()
	prod := demozoo.Production{}
	a, b := prod.Groups()
	be.Equal(t, a, "")
	be.Equal(t, b, "")
}

func TestSite(t *testing.T) {
	t.Parallel()
	s := demozoo.Site("")
	be.Equal(t, s, "")
	s = demozoo.Site("the cool bbs")
	be.Equal(t, "cool BBS", s)
	s = demozoo.Site("Cool BBS")
	be.Equal(t, "Cool BBS", s)
}

func TestReleasers(t *testing.T) {
	t.Parallel()
	prod := demozoo.Production{}
	a, b, c, d := prod.Releasers()
	be.Equal(t, len(a), 0)
	be.Equal(t, len(b), 0)
	be.Equal(t, len(c), 0)
	be.Equal(t, len(d), 0)
}

func TestCategory(t *testing.T) {
	t.Parallel()
	c := demozoo.TextC.String()
	be.Equal(t, "text", c)
	c = demozoo.CodeC.String()
	be.Equal(t, "code", c)
	c = demozoo.GraphicsC.String()
	be.Equal(t, "graphics", c)
	c = demozoo.MusicC.String()
	be.Equal(t, "music", c)
	c = demozoo.MagazineC.String()
	be.Equal(t, "magazine", c)
}
