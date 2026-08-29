package postgres_test

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

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
		{
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

	_, _, err := postgres.Connections(nil)
	be.Err(t, err)

	x, y, err := postgres.Connections(db)
	be.Err(t, err, nil)
	be.True(t, x > 0)
	be.True(t, y > 0)
}
