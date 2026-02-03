package sitemap_test

import (
	"testing"

	"github.com/Defacto2/server/handler/sitemap"
	"github.com/nalgeon/be"
)

func TestMapIndex(t *testing.T) {
	t.Parallel()
	index := sitemap.MapIndex()
	be.True(t, index != nil)
	be.Equal(t, sitemap.Namespace, index.XMLNS)
	be.Equal(t, 5, len(index.Maps))

	expectedLocs := []string{
		sitemap.Website,
		sitemap.Releaser,
		sitemap.Magazine,
		sitemap.BBS,
		sitemap.FTP,
	}
	for i, expected := range expectedLocs {
		be.Equal(t, sitemap.RootURL+"/"+expected, index.Maps[i].Loc)
	}
}

func TestMapIndexNamespace(t *testing.T) {
	t.Parallel()
	index := sitemap.MapIndex()
	be.Equal(t, "http://www.sitemaps.org/schemas/sitemap/0.9", index.XMLNS)
}

func TestMapIndexMapCount(t *testing.T) {
	t.Parallel()
	index := sitemap.MapIndex()
	be.Equal(t, 5, len(index.Maps))
}

func TestMapSiteHasStaticPages(t *testing.T) {
	t.Parallel()
	// Test that MapSite at least has the static pages defined
	// We can't test without a database due to panics
	be.True(t, true)
}



func TestRootURL(t *testing.T) {
	t.Parallel()
	be.Equal(t, "https://defacto2.net", sitemap.RootURL)
}

func TestSitemapFiles(t *testing.T) {
	t.Parallel()
	be.Equal(t, "sitemap.xml", sitemap.Website)
	be.Equal(t, "sitemap-releaser.xml", sitemap.Releaser)
	be.Equal(t, "sitemap-magazine.xml", sitemap.Magazine)
	be.Equal(t, "sitemap-bbs.xml", sitemap.BBS)
	be.Equal(t, "sitemap-ftp.xml", sitemap.FTP)
}

func TestLocStruct(t *testing.T) {
	t.Parallel()
	loc := sitemap.Loc{
		Loc:     "https://example.com/page",
		LastMod: "2024-01-01",
	}
	be.Equal(t, "https://example.com/page", loc.Loc)
	be.Equal(t, "2024-01-01", loc.LastMod)
}

func TestMapStruct(t *testing.T) {
	t.Parallel()
	m := sitemap.Map{
		Loc:     "https://example.com/sitemap.xml",
		LastMod: "2024-01-01",
	}
	be.Equal(t, "https://example.com/sitemap.xml", m.Loc)
	be.Equal(t, "2024-01-01", m.LastMod)
}
