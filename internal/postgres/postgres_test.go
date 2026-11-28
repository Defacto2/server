package postgres_test

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

// checked in Sep 26, test coverage was average at just under 40%

func TestConnValidate(t *testing.T) {
	t.Parallel()
	be.Err(t, (postgres.Connection{}).Validate(nil))

	tests := []struct {
		name      string
		url       string
		wantWarn  bool
		forbidStr string
	}{
		{ //nolint:gosec
			name:     "valid postgres URL",
			url:      "postgres://testuser:testpass@localhost:5432/testdb",
			wantWarn: false,
		},
		{ //nolint:gosec
			name:     "valid postgresql URL",
			url:      "postgresql://testuser:testpass@localhost:5432/testdb",
			wantWarn: false,
		},
		{
			name:     "empty URL",
			url:      "",
			wantWarn: true,
		},
		{
			name:     "invalid URL scheme",
			url:      "mysql://localhost/db",
			wantWarn: true,
		},
		{
			name:      "malformed URL hides password",
			url:       "postgres://user:supersecret@localhost:badport/db",
			wantWarn:  true,
			forbidStr: "supersecret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			sl := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

			conn := postgres.Connection{URL: tt.url}
			err := conn.Validate(sl)
			be.Err(t, err, nil)

			out := buf.String()
			if tt.wantWarn {
				be.True(t, len(out) > 0)
			} else {
				be.Equal(t, out, "")
			}
			if tt.forbidStr != "" {
				be.True(t, !strings.Contains(out, tt.forbidStr))
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		t.Setenv("D2_DATABASE_URL", "")
		be.Err(t, os.Unsetenv("D2_DATABASE_URL"), nil)

		conn, err := postgres.New()
		be.Err(t, err, nil)
		be.Equal(t, conn.URL, postgres.DefaultURL)
	})

	t.Run("env", func(t *testing.T) {
		const custom = "postgres://user:pass@db.example:5432/app" //nolint:gosec
		t.Setenv("D2_DATABASE_URL", custom)

		conn, err := postgres.New()
		be.Err(t, err, nil)
		be.Equal(t, conn.URL, custom)
	})
}

func TestVersion(t *testing.T) {
	t.Parallel()
	var ver postgres.Version
	got := ver.Query(nil)
	be.Err(t, got, nil)

	parsed := postgres.Version("PostgreSQL 16.2 on x86_64-pc-linux-gnu")
	be.Equal(t, parsed.String(), "and using PostgreSQL 16.2")
	nonFloat := postgres.Version("PostgreSQL beta on linux")
	be.Equal(t, nonFloat.String(), "PostgreSQL beta on linux")
	short := postgres.Version("short")
	be.Equal(t, short.String(), "short")

	db := testutil.DB(t)
	got = ver.Query(db)
	be.Err(t, got, nil)
	be.True(t, strings.Contains(ver.String(), "PostgreSQL"))
}

func TestConnections(t *testing.T) {
	t.Parallel()

	_, _, err := postgres.Connections(t.Context(), nil)
	be.Err(t, err)

	db := testutil.DB(t)

	x, y, err := postgres.Connections(t.Context(), db)
	be.Err(t, err, nil)
	be.True(t, x > 0)
	be.True(t, y > 0)
	be.Equal(t, db.Stats().MaxOpenConnections, 25)
}

func TestScenerSQL(t *testing.T) {
	t.Parallel()
	clause, args := postgres.ScenerSQL("ace")
	be.Equal(t, len(args), 2)
	be.Equal(t, args[0], "ACE")
	be.Equal(t, args[1], "ACE")
	be.True(t, strings.Contains(clause, `LIKE '%,'||$2||',%' ESCAPE '\'`))
	be.True(t, !strings.Contains(clause, `LIKE '%'||`))

	_, hypArgs := postgres.ScenerSQL("razor-1911")
	be.Equal(t, len(hypArgs), 2)
	be.Equal(t, hypArgs[0], "RAZOR 1911")
	be.Equal(t, hypArgs[1], "RAZOR 1911")

	_, invalidArgs := postgres.ScenerSQL("bad#name")
	be.Equal(t, len(invalidArgs), 2)
	be.Equal(t, invalidArgs[0], "")
	be.Equal(t, invalidArgs[1], "")
}

func TestSimilarTo(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		q, args := postgres.SimilarToReleaser()
		be.Equal(t, q, postgres.SQL(""))
		be.Equal(t, len(args), 0)

		q, args = postgres.SimilarToMagazine()
		be.Equal(t, q, postgres.SQL(""))
		be.Equal(t, len(args), 0)

		q, args = postgres.SimilarToExact()
		be.Equal(t, q, postgres.SQL(""))
		be.Equal(t, len(args), 0)
	})

	t.Run("releaser", func(t *testing.T) {
		t.Parallel()
		q, args := postgres.SimilarToReleaser(" razor ", `100%_a\b`)
		be.True(t, strings.Contains(string(q), `releaser LIKE $1 ESCAPE '\' OR releaser LIKE $2 ESCAPE '\'`))
		be.Equal(t, len(args), 2)
		be.Equal(t, args[0], "RAZOR%")
		be.Equal(t, args[1], `100\%\_A\\B%`)
	})

	t.Run("magazine", func(t *testing.T) {
		t.Parallel()
		q, args := postgres.SimilarToMagazine("nwr")
		be.True(t, strings.Contains(string(q), `releaser LIKE $1 ESCAPE '\'`))
		be.True(t, strings.Contains(string(q), `section = 'magazine'`))
		be.Equal(t, len(args), 1)
		be.Equal(t, args[0], "NWR%")
	})

	t.Run("exact", func(t *testing.T) {
		t.Parallel()
		q, args := postgres.SimilarToExact(" razor 1911 ", "defacto")
		be.True(t, strings.Contains(string(q), "releaser IN ($1, $2)"))
		be.Equal(t, len(args), 2)
		be.Equal(t, args[0], "RAZOR 1911")
		be.Equal(t, args[1], "DEFACTO")
	})
}

func TestIndexes(t *testing.T) {
	t.Parallel()

	_, err := postgres.Indexes(t.Context(), nil)
	be.Err(t, err)

	idx := postgres.IndexStat{
		Method:     "gin",
		BlocksHit:  95,
		BlocksRead: 5,
		Scans:      1234567,
		Definition: "CREATE INDEX idx_files_desc_fts ON public.files USING gin (to_tsvector('english'::regconfig, COALESCE(record_title, ''::character varying)::text)) WHERE (deletedat IS NULL)",
	}
	be.Equal(t, idx.CacheHitPct(), "95.0%")
	be.Equal(t, idx.LastScanString(), "never")
	be.Equal(t, idx.ScansFormatted(), "1,234,567")
	be.Equal(t, idx.ExprSummary(), "(to_tsvector('english', COALESCE(record_title, ''))) WHERE (deletedat IS NULL)")

	summary := postgres.TableIndexStats{SeqScan: 1, IdxScan: 9, Updates: 10, HotUpdates: 8}
	be.Equal(t, summary.IdxScanPct(), "90.0%")
	be.Equal(t, summary.HotUpdatePct(), "80.0%")
	be.Equal(t, summary.LastVacuumString(), "never")
	be.Equal(t, summary.LastAnalyzeString(), "never")
	be.Equal(t, summary.StatsResetString(), "")

	zeroIdx := postgres.IndexStat{}
	be.Equal(t, zeroIdx.CacheHitPct(), "-")

	ts := time.Date(2026, time.January, 2, 15, 4, 0, 0, time.UTC)
	zeroIdx.LastScan = &ts
	be.Equal(t, zeroIdx.LastScanString(), "2026-01-02 15:04")

	zeroSummary := postgres.TableIndexStats{}
	be.Equal(t, zeroSummary.IdxScanPct(), "-")
	be.Equal(t, zeroSummary.HotUpdatePct(), "-")
	zeroSummary.LastVacuum = &ts
	zeroSummary.LastAnalyze = &ts
	zeroSummary.StatsReset = &ts
	be.Equal(t, zeroSummary.LastVacuumString(), "2026-01-02 15:04")
	be.Equal(t, zeroSummary.LastAnalyzeString(), "2026-01-02 15:04")
	be.Equal(t, zeroSummary.StatsResetString(), "2026-01-02 15:04")

	db := testutil.DB(t)
	got, err := postgres.Indexes(t.Context(), db)
	be.Err(t, err, nil)
	be.Equal(t, got.TableName, "files")
	be.True(t, got.TableSize != "")
	be.True(t, got.IndexesSize != "")
	be.True(t, got.TotalSize != "")
	be.True(t, len(got.Indexes) > 0)
	for _, item := range got.Indexes {
		be.True(t, item.Name != "")
		be.True(t, item.Method != "")
		be.True(t, item.Size != "")
		be.True(t, item.Definition != "")
	}
}
