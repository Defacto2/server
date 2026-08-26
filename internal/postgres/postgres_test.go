package postgres_test

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/nalgeon/be"
)

// TestDefaultURL verifies the default connection URL format.
func TestDefaultURL(t *testing.T) {
	be.True(t, strings.HasPrefix(postgres.DefaultURL, "postgres://"))
	be.True(t, strings.Contains(postgres.DefaultURL, "localhost"))
	be.True(t, strings.Contains(postgres.DefaultURL, "defacto2_ps"))
}

// TestDriverName verifies the driver name is correct.
func TestDriverName(t *testing.T) {
	be.Equal(t, "pgx", postgres.DriverName)
}

// TestProtocol verifies the protocol name is correct.
func TestProtocol(t *testing.T) {
	be.Equal(t, "postgres", postgres.Protocol)
}

// TestErrEnvValue verifies the error value is defined.
func TestErrEnvValue(t *testing.T) {
	be.True(t, len(postgres.ErrEnvValue.Error()) > 0)
	be.True(t, strings.Contains(postgres.ErrEnvValue.Error(), "environment"))
}

// TestConnectionValidate tests the Connection.Validate method.
func TestConnectionValidate(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name        string
		url         string
		shouldError bool
	}{
		{ //nolint:gosec
			name:        "valid URL",
			url:         "postgres://testuser:testpass@localhost:5432/testdb",
			shouldError: false,
		},
		{
			name:        "empty URL",
			url:         "",
			shouldError: false,
		},
		{
			name:        "invalid URL scheme",
			url:         "mysql://localhost/db",
			shouldError: false, // Validate logs warnings but doesn't error
		},
		{
			name:        "malformed URL",
			url:         "ht!tp://[invalid",
			shouldError: false, // Invalid URL format, but Validate handles gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := postgres.Connection{URL: tt.url}
			err := conn.Validate(logger)

			if tt.shouldError {
				be.True(t, err != nil)
			} else {
				be.Equal(t, nil, err)
			}
		})
	}
}

// TestConnectionValidateNilLogger tests Validate with nil logger.
func TestConnectionValidateNilLogger(t *testing.T) {
	conn := postgres.Connection{URL: "postgres://localhost"}
	err := conn.Validate(nil)
	be.True(t, err != nil)
}

// TestNew tests the New connection initialization.
func TestNew(t *testing.T) {
	conn, err := postgres.New()

	be.Equal(t, nil, err)
	be.True(t, len(conn.URL) > 0)
	// Should use default URL when no env var is set
	be.True(t, strings.HasPrefix(conn.URL, "postgres://") || conn.URL == postgres.DefaultURL)
}

// TestConnectionStruct tests the Connection struct fields.
func TestConnectionStruct(t *testing.T) {
	conn := postgres.Connection{URL: "postgres://test"}
	be.Equal(t, "postgres://test", conn.URL)
}

// TestVersionQuery tests Version with nil database.
func TestVersionQuery_NilDB(t *testing.T) {
	var v postgres.Version
	err := v.Query(nil)
	be.Equal(t, nil, err)
}
