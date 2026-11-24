//nolint:paralleltest
package postgres_test

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

// checked in Sep 26, test coverage was average at just under 40%

func TestDefaultURL(t *testing.T) {
	be.True(t, strings.HasPrefix(postgres.DefaultURL, "postgres://"))
	be.True(t, strings.Contains(postgres.DefaultURL, "localhost"))
	be.True(t, strings.Contains(postgres.DefaultURL, "defacto2_ps"))
}

func TestDriverName(t *testing.T) {
	be.Equal(t, "pgx", postgres.DriverName)
}

func TestProtocol(t *testing.T) {
	be.Equal(t, "postgres", postgres.Protocol)
}

func TestErrEnvValue(t *testing.T) {
	be.True(t, len(postgres.ErrEnvValue.Error()) > 0)
	be.True(t, strings.Contains(postgres.ErrEnvValue.Error(), "environment"))
}

func TestConnValidate(t *testing.T) {
	sl := slog.Default()

	tests := []struct {
		name string
		url  string
		want bool
	}{
		{ //nolint:gosec
			name: "valid URL",
			url:  "postgres://testuser:testpass@localhost:5432/testdb",
			want: false,
		},
		{
			name: "empty URL",
			url:  "",
			want: false,
		},
		{
			name: "invalid URL scheme",
			url:  "mysql://localhost/db",
			want: false,
		},
		{
			name: "malformed URL",
			url:  "ht!tp://[invalid",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := postgres.Connection{URL: tt.url}
			err := conn.Validate(sl)

			if tt.want {
				be.Err(t, err)
			} else {
				be.Err(t, err, nil)
			}
		})
	}
}

func TestNew(t *testing.T) {
	conn, got := postgres.New()
	be.Err(t, got, nil)
	be.True(t, len(conn.URL) > 0)
	// Should use default URL when no env var is set
	be.True(t, strings.HasPrefix(conn.URL, "postgres://") || conn.URL == postgres.DefaultURL)
}

func TestConnection(t *testing.T) {
	conn := postgres.Connection{URL: "postgres://test"}
	be.Equal(t, conn.URL, "postgres://test")
}

func TestVersion(t *testing.T) {
	var ver postgres.Version
	got := ver.Query(nil)
	be.Err(t, got, nil)

	db := testutil.DB(t)
	got = ver.Query(db)
	be.Err(t, got, nil)
	be.True(t, strings.Contains(ver.String(), "PostgreSQL"))
}

func TestConnections(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)

	_, _, err := postgres.Connections(t.Context(), nil)
	be.Err(t, err)

	x, y, err := postgres.Connections(t.Context(), db)
	be.Err(t, err, nil)
	be.True(t, x > 0)
	be.True(t, y > 0)
	be.Equal(t, db.Stats().MaxOpenConnections, 25)
}

func TestScenerSQL(t *testing.T) {
	clause, args := postgres.ScenerSQL("ace")
	be.Equal(t, len(args), 2)
	be.Equal(t, args[0], "ACE")
	be.Equal(t, args[1], "ACE")
	be.True(t, strings.Contains(clause, `LIKE '%,'||$2||',%' ESCAPE '\'`))
	be.True(t, !strings.Contains(clause, `LIKE '%'||`))
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
