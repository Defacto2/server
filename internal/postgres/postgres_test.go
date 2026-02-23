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
		{
			name:        "valid URL",
			url:         "postgres://user:pass@localhost:5432/db",
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

// BenchmarkVersionString benchmarks the Version.String method.
func BenchmarkVersionString(b *testing.B) {
	v := postgres.Version("PostgreSQL 13.8 on x86_64-pc-linux-gnu")
	b.ResetTimer()
	for range b.N {
		_ = v.String()
	}
}

// BenchmarkColumns benchmarks the Columns function.
func BenchmarkColumns(b *testing.B) {
	for range b.N {
		_ = postgres.Columns()
	}
}

// BenchmarkStat benchmarks the Stat function.
func BenchmarkStat(b *testing.B) {
	for range b.N {
		_ = postgres.Stat()
	}
}

// BenchmarkConnectionValidate benchmarks the Connection.Validate method.
func BenchmarkConnectionValidate(b *testing.B) {
	logger := slog.Default()
	conn := postgres.Connection{URL: "postgres://localhost:5432/test"}
	b.ResetTimer()
	for range b.N {
		_ = conn.Validate(logger)
	}
}

// BenchmarkRoles benchmarks the Roles function.
func BenchmarkRoles(b *testing.B) {
	for range b.N {
		_ = postgres.Roles()
	}
}

// BenchmarkReleasersAlphabetical benchmarks the ReleasersAlphabetical function.
func BenchmarkReleasersAlphabetical(b *testing.B) {
	for range b.N {
		_ = postgres.ReleasersAlphabetical()
	}
}

// BenchmarkBBSsAlphabetical benchmarks the BBSsAlphabetical function.
func BenchmarkBBSsAlphabetical(b *testing.B) {
	for range b.N {
		_ = postgres.BBSsAlphabetical()
	}
}

// BenchmarkMagazinesAlphabetical benchmarks the MagazinesAlphabetical function.
func BenchmarkMagazinesAlphabetical(b *testing.B) {
	for range b.N {
		_ = postgres.MagazinesAlphabetical()
	}
}

// BenchmarkReleasersProlific benchmarks the ReleasersProlific function.
func BenchmarkReleasersProlific(b *testing.B) {
	for range b.N {
		_ = postgres.ReleasersProlific()
	}
}

// BenchmarkReleasersOldest benchmarks the ReleasersOldest function.
func BenchmarkReleasersOldest(b *testing.B) {
	for range b.N {
		_ = postgres.ReleasersOldest()
	}
}

// BenchmarkSceners benchmarks the Sceners function.
func BenchmarkSceners(b *testing.B) {
	for range b.N {
		_ = postgres.Sceners()
	}
}

// BenchmarkWriters benchmarks the Writers function.
func BenchmarkWriters(b *testing.B) {
	for range b.N {
		_ = postgres.Writers()
	}
}

// BenchmarkArtists benchmarks the Artists function.
func BenchmarkArtists(b *testing.B) {
	for range b.N {
		_ = postgres.Artists()
	}
}

// BenchmarkCoders benchmarks the Coders function.
func BenchmarkCoders(b *testing.B) {
	for range b.N {
		_ = postgres.Coders()
	}
}

// BenchmarkMusicians benchmarks the Musicians function.
func BenchmarkMusicians(b *testing.B) {
	for range b.N {
		_ = postgres.Musicians()
	}
}

// BenchmarkSetUpper benchmarks the SetUpper function.
func BenchmarkSetUpper(b *testing.B) {
	for range b.N {
		_ = postgres.SetUpper("releaser")
	}
}

// BenchmarkSetFilesize0 benchmarks the SetFilesize0 function.
func BenchmarkSetFilesize0(b *testing.B) {
	for range b.N {
		_ = postgres.SetFilesize0()
	}
}

// BenchmarkSumSection benchmarks the SumSection function.
func BenchmarkSumSection(b *testing.B) {
	for range b.N {
		_ = postgres.SumSection()
	}
}

// BenchmarkSumGroup benchmarks the SumGroup function.
func BenchmarkSumGroup(b *testing.B) {
	for range b.N {
		_ = postgres.SumGroup()
	}
}

// BenchmarkSumPlatform benchmarks the SumPlatform function.
func BenchmarkSumPlatform(b *testing.B) {
	for range b.N {
		_ = postgres.SumPlatform()
	}
}

// BenchmarkSummary benchmarks the Summary function.
func BenchmarkSummary(b *testing.B) {
	for range b.N {
		_ = postgres.Summary()
	}
}

// BenchmarkReleasers benchmarks the Releasers function.
func BenchmarkReleasers(b *testing.B) {
	for range b.N {
		_ = postgres.Releasers()
	}
}
