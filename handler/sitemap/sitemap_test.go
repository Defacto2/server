package sitemap_test

import (
	"testing"

	"github.com/Defacto2/server/handler/sitemap"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

// checked in Sep 26, test coverage was great at around 85%+

func TestMapIndex(t *testing.T) {
	t.Parallel()

	index := sitemap.MapIndex()
	be.True(t, index != nil)
	be.Equal(t, sitemap.Namespace, index.XMLNS)
	be.Equal(t, 5, len(index.Maps))

	expectedLocs := [...]string{
		sitemap.Website,
		sitemap.Releaser,
		sitemap.Magazine,
		sitemap.BBS,
		sitemap.FTP,
	}
	for i, expected := range expectedLocs {
		be.Equal(t, sitemap.RootURL+"/"+expected, index.Maps[i].Loc)
	}

	be.Equal(t, index.XMLNS, "http://www.sitemaps.org/schemas/sitemap/0.9")
	be.Equal(t, len(index.Maps), 5)
	be.Equal(t, sitemap.RootURL, "https://defacto2.net")
	be.Equal(t, sitemap.Website, "sitemap.xml")
	be.Equal(t, sitemap.Releaser, "sitemap-releaser.xml")
	be.Equal(t, sitemap.Magazine, "sitemap-magazine.xml")
	be.Equal(t, sitemap.BBS, "sitemap-bbs.xml")
	be.Equal(t, sitemap.FTP, "sitemap-ftp.xml")
}

func TestMapSite(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	sl := logs.Discard()

	sitemap := sitemap.MapSite(t.Context(), sl, db)
	locs := len(sitemap.Locs)
	t.Log(locs)
	be.True(t, locs > 20)
}

func TestMapReleaser(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	sl := logs.Discard()

	sitemap := sitemap.MapReleaser(t.Context(), sl, db)
	locs := len(sitemap.Locs)
	t.Log(locs)
	be.True(t, locs > 20)
}

func TestMapMagazine(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	sl := logs.Discard()

	sitemap := sitemap.MapMagazine(t.Context(), sl, db)
	locs := len(sitemap.Locs)
	t.Log(locs)
	be.True(t, locs > 2)
}

func TestMapBBS(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	sl := logs.Discard()

	sitemap := sitemap.MapBBS(t.Context(), sl, db)
	locs := len(sitemap.Locs)
	t.Log(locs)
	be.True(t, locs > 2)
}

func TestMapFTP(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	sl := logs.Discard()

	sitemap := sitemap.MapFTP(t.Context(), sl, db)
	locs := len(sitemap.Locs)
	t.Log(locs)
	be.True(t, locs > 2)
}
