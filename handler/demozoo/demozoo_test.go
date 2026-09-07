package demozoo_test

import (
	"testing"

	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/internal/tags"
	"github.com/nalgeon/be"
)

// checked in Sep 26, test coverage was fine at under 40%

// Set to true to test against the remote servers.
const testRemoteServers = false

func TestDemozoo_Get(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	prod := demozoo.Production{}
	_, got := prod.Get(ctx, -1)
	be.Err(t, got, demozoo.ErrID)

	if testRemoteServers {
		_, got = prod.Get(ctx, 1)
		be.Err(t, got, demozoo.ErrSuccess)
	}
}

func TestFind(t *testing.T) {
	t.Parallel()

	got := demozoo.Find("defacto2")
	wants := demozoo.GroupID(10000)
	be.Equal(t, got, wants)

	got = demozoo.Find("notfound")
	wants = demozoo.GroupID(0)
	be.Equal(t, got, wants)
}

func TestExternalLinks(t *testing.T) {
	t.Parallel()

	d := demozoo.Production{}
	be.Equal(t, d.PouetProd(), 0)

	d.ExternalLinks = append(d.ExternalLinks, struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	}{
		LinkClass: "class1",
		URL:       "http://example.com/1",
	})
	be.Equal(t, d.PouetProd(), 0)

	d.ExternalLinks = append(d.ExternalLinks, struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	}{
		LinkClass: "PouetProduction",
		URL:       "http://example.com/1",
	})
	be.Equal(t, d.PouetProd(), 0)

	d.ExternalLinks = append(d.ExternalLinks, struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	}{
		LinkClass: "PouetProduction",
		URL:       "https://www.pouet.net/prod.php?which=71562",
	})
	be.Equal(t, d.PouetProd(), 71562)
	be.Equal(t, d.GithubRepo(), "")

	d.ExternalLinks = append(d.ExternalLinks, struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	}{
		LinkClass: "GithubRepo",
		URL:       "https://github.com/Defacto2/server",
	})
	be.Equal(t, d.GithubRepo(), "/Defacto2/server")
	be.Equal(t, d.YouTubeVideo(), "")

	d.ExternalLinks = append(d.ExternalLinks, struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	}{
		LinkClass: "YoutubeVideo",
		URL:       "https://www.youtube.com/watch?v=x6QrKsBOERA",
	})
	be.Equal(t, d.YouTubeVideo(), "x6QrKsBOERA")
}

func TestUnmarshal(t *testing.T) {
	t.Parallel()

	prod := demozoo.Production{}
	got := prod.Unmarshal(nil)
	be.Err(t, got, nil)
}

func TestSuperType(t *testing.T) {
	t.Parallel()

	prod := demozoo.Production{}
	x, y := prod.SuperType()

	const wants tags.Tag = -1
	be.Equal(t, x, wants)
	be.Equal(t, y, wants)
}

func TestReleased(t *testing.T) {
	t.Parallel()

	prod := demozoo.Production{}
	y, m, d := prod.Released()

	const wants int16 = 0
	be.Equal(t, y, wants)
	be.Equal(t, m, wants)
	be.Equal(t, d, wants)
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
	be.Equal(t, s, "cool BBS")

	s = demozoo.Site("Cool BBS")
	be.Equal(t, s, "Cool BBS")
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
	be.Equal(t, c, "text")

	c = demozoo.CodeC.String()
	be.Equal(t, c, "code")

	c = demozoo.GraphicsC.String()
	be.Equal(t, c, "graphics")

	c = demozoo.MusicC.String()
	be.Equal(t, c, "music")

	c = demozoo.MagazineC.String()
	be.Equal(t, c, "magazine")
}
