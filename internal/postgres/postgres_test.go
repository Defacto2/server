package postgres

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/nalgeon/be"
)

// TestDefaultURL verifies the default connection URL format.
func TestDefaultURL(t *testing.T) {
	be.True(t, strings.HasPrefix(DefaultURL, "postgres://"))
	be.True(t, strings.Contains(DefaultURL, "localhost"))
	be.True(t, strings.Contains(DefaultURL, "defacto2_ps"))
}

// TestDriverName verifies the driver name is correct.
func TestDriverName(t *testing.T) {
	be.Equal(t, "pgx", DriverName)
}

// TestProtocol verifies the protocol name is correct.
func TestProtocol(t *testing.T) {
	be.Equal(t, "postgres", Protocol)
}

// TestErrEnvValue verifies the error value is defined.
func TestErrEnvValue(t *testing.T) {
	be.True(t, len(ErrEnvValue.Error()) > 0)
	be.True(t, strings.Contains(ErrEnvValue.Error(), "environment"))
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
			conn := Connection{URL: tt.url}
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
	conn := Connection{URL: "postgres://localhost"}
	err := conn.Validate(nil)
	be.True(t, err != nil)
}

// TestNew tests the New connection initialization.
func TestNew(t *testing.T) {
	conn, err := New()

	be.Equal(t, nil, err)
	be.True(t, len(conn.URL) > 0)
	// Should use default URL when no env var is set
	be.True(t, strings.HasPrefix(conn.URL, "postgres://") || conn.URL == DefaultURL)
}

// TestConnectionStruct tests the Connection struct fields.
func TestConnectionStruct(t *testing.T) {
	conn := Connection{URL: "postgres://test"}
	be.Equal(t, "postgres://test", conn.URL)
}

// TestVersionQuery tests Version with nil database.
func TestVersionQuery_NilDB(t *testing.T) {
	var v Version
	err := v.Query(nil)
	be.Equal(t, nil, err)
}

// BenchmarkVersionString benchmarks the Version.String method.
func BenchmarkVersionString(b *testing.B) {
	v := Version("PostgreSQL 13.8 on x86_64-pc-linux-gnu")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.String()
	}
}

// BenchmarkColumns benchmarks the Columns function.
func BenchmarkColumns(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Columns()
	}
}

// BenchmarkStat benchmarks the Stat function.
func BenchmarkStat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Stat()
	}
}

// BenchmarkConnectionValidate benchmarks the Connection.Validate method.
func BenchmarkConnectionValidate(b *testing.B) {
	logger := slog.Default()
	conn := Connection{URL: "postgres://localhost:5432/test"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = conn.Validate(logger)
	}
}

// BenchmarkRoles benchmarks the Roles function.
func BenchmarkRoles(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Roles()
	}
}

// BenchmarkReleasersAlphabetical benchmarks the ReleasersAlphabetical function.
func BenchmarkReleasersAlphabetical(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ReleasersAlphabetical()
	}
}

// BenchmarkBBSsAlphabetical benchmarks the BBSsAlphabetical function.
func BenchmarkBBSsAlphabetical(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = BBSsAlphabetical()
	}
}

// BenchmarkMagazinesAlphabetical benchmarks the MagazinesAlphabetical function.
func BenchmarkMagazinesAlphabetical(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = MagazinesAlphabetical()
	}
}

// BenchmarkReleasersProlific benchmarks the ReleasersProlific function.
func BenchmarkReleasersProlific(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ReleasersProlific()
	}
}

// BenchmarkReleasersOldest benchmarks the ReleasersOldest function.
func BenchmarkReleasersOldest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ReleasersOldest()
	}
}

// BenchmarkSceners benchmarks the Sceners function.
func BenchmarkSceners(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Sceners()
	}
}

// BenchmarkWriters benchmarks the Writers function.
func BenchmarkWriters(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Writers()
	}
}

// BenchmarkArtists benchmarks the Artists function.
func BenchmarkArtists(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Artists()
	}
}

// BenchmarkCoders benchmarks the Coders function.
func BenchmarkCoders(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Coders()
	}
}

// BenchmarkMusicians benchmarks the Musicians function.
func BenchmarkMusicians(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Musicians()
	}
}

// BenchmarkSetUpper benchmarks the SetUpper function.
func BenchmarkSetUpper(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = SetUpper("releaser")
	}
}

// BenchmarkSetFilesize0 benchmarks the SetFilesize0 function.
func BenchmarkSetFilesize0(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = SetFilesize0()
	}
}

// BenchmarkSumSection benchmarks the SumSection function.
func BenchmarkSumSection(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = SumSection()
	}
}

// BenchmarkSumGroup benchmarks the SumGroup function.
func BenchmarkSumGroup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = SumGroup()
	}
}

// BenchmarkSumPlatform benchmarks the SumPlatform function.
func BenchmarkSumPlatform(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = SumPlatform()
	}
}

// BenchmarkSummary benchmarks the Summary function.
func BenchmarkSummary(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Summary()
	}
}

// BenchmarkReleasers benchmarks the Releasers function.
func BenchmarkReleasers(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Releasers()
	}
}
