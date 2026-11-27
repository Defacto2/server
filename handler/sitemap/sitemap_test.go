package sitemap_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/xml"
	"log/slog"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/sitemap"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

var testLog = logs.Discard()

func TestMapIndex(t *testing.T) {
	t.Parallel()

	index := sitemap.MapIndex()
	if index == nil {
		t.Fatal("expected non-nil index")
	}
	be.Equal(t, index.XMLNS, sitemap.Namespace)

	expectedLocs := [...]string{
		sitemap.Website,
		sitemap.Releaser,
		sitemap.Magazine,
		sitemap.BBS,
		sitemap.FTP,
	}
	if len(index.Maps) != len(expectedLocs) {
		t.Fatalf("got %d maps, want %d", len(index.Maps), len(expectedLocs))
	}
	for i, expected := range expectedLocs {
		be.Equal(t, index.Maps[i].Loc, sitemap.RootURL+"/"+expected)
	}

	raw, err := xml.Marshal(index)
	be.Err(t, err, nil)
	out := string(raw)
	be.True(t, strings.HasPrefix(out, `<sitemapindex xmlns="`+sitemap.Namespace+`">`))
	be.True(t, strings.Contains(out, `<sitemap><loc>`+sitemap.RootURL+"/"+sitemap.Website+`</loc></sitemap>`))
	be.True(t, strings.HasSuffix(out, `</sitemapindex>`))
}

func TestOffline(t *testing.T) {
	t.Parallel()

	db, err := sql.Open("pgx", "")
	be.Err(t, err, nil)
	t.Cleanup(func() {
		_ = db.Close()
	})

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	tests := []struct {
		name     string
		fn       func(context.Context, *slog.Logger, *sql.DB) *sitemap.Sitemap
		wantLocs int
	}{
		{"site", sitemap.MapSite, 18},
		{"releaser", sitemap.MapReleaser, 0},
		{"magazine", sitemap.MapMagazine, 0},
		{"bbs", sitemap.MapBBS, 0},
		{"ftp", sitemap.MapFTP, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var logBuf bytes.Buffer
			sl := slog.New(slog.NewTextHandler(&logBuf, nil))

			sm := tt.fn(ctx, sl, db)
			if sm == nil {
				t.Fatal("expected non-nil sitemap")
			}
			be.Equal(t, sm.XMLNS, sitemap.Namespace)
			be.Equal(t, len(sm.Locs), tt.wantLocs)
			be.True(t, strings.Contains(logBuf.String(), "level=ERROR"))

			raw, err := xml.Marshal(sm)
			be.Err(t, err, nil)
			out := string(raw)
			be.True(t, strings.HasPrefix(out, `<urlset xmlns="`+sitemap.Namespace+`">`))
			if tt.wantLocs > 0 {
				be.True(t, strings.Contains(out, `<url><loc>`+sitemap.RootURL+`/`))
			}
		})
	}
}

func TestMapSite(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	sm := sitemap.MapSite(t.Context(), testLog, db)
	be.True(t, len(sm.Locs) > 20)
}

func TestMapReleaser(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	sm := sitemap.MapReleaser(t.Context(), testLog, db)
	be.True(t, len(sm.Locs) > 20)
}

func TestMapMagazine(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	sm := sitemap.MapMagazine(t.Context(), testLog, db)
	be.True(t, len(sm.Locs) > 2)
}

func TestMapBBS(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	sm := sitemap.MapBBS(t.Context(), testLog, db)
	be.True(t, len(sm.Locs) > 2)
}

func TestMapFTP(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	sm := sitemap.MapFTP(t.Context(), testLog, db)
	be.True(t, len(sm.Locs) > 2)
}
