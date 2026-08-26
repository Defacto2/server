package postgres_test

import (
	"log/slog"
	"testing"

	"github.com/Defacto2/server/internal/postgres"
)

// BenchmarkVersionString benchmarks the Version.String method.
func BenchmarkVersionString(b *testing.B) {
	v := postgres.Version("PostgreSQL 13.8 on x86_64-pc-linux-gnu")
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = v.String()
		}
	})
}

// BenchmarkConnectionValidate benchmarks the Connection.Validate method.
func BenchmarkConnectionValidate(b *testing.B) {
	logger := slog.Default()
	conn := postgres.Connection{URL: "postgres://localhost:5432/test"}
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = conn.Validate(logger)
		}
	})
}

// BenchmarkRoles benchmarks the Roles function.
func BenchmarkRoles(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.Roles()
		}
	})
}

// BenchmarkReleasersAlphabetical benchmarks the ReleasersAlphabetical function.
func BenchmarkReleasersAlphabetical(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.ReleasersAlphabetical()
		}
	})
}

// BenchmarkBBSsAlphabetical benchmarks the BBSsAlphabetical function.
func BenchmarkBBSsAlphabetical(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.BBSsAlphabetical()
		}
	})
}

// BenchmarkMagazinesAlphabetical benchmarks the MagazinesAlphabetical function.
func BenchmarkMagazinesAlphabetical(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.MagazinesAlphabetical()
		}
	})
}

// BenchmarkReleasersProlific benchmarks the ReleasersProlific function.
func BenchmarkReleasersProlific(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.ReleasersProlific()
		}
	})
}

// BenchmarkReleasersOldest benchmarks the ReleasersOldest function.
func BenchmarkReleasersOldest(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.ReleasersOldest()
		}
	})
}

// BenchmarkSceners benchmarks the Sceners function.
func BenchmarkSceners(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.Sceners()
		}
	})
}

// BenchmarkWriters benchmarks the Writers function.
func BenchmarkWriters(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.Writers()
		}
	})
}

// BenchmarkArtists benchmarks the Artists function.
func BenchmarkArtists(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.Artists()
		}
	})
}

// BenchmarkCoders benchmarks the Coders function.
func BenchmarkCoders(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.Coders()
		}
	})
}

// BenchmarkMusicians benchmarks the Musicians function.
func BenchmarkMusicians(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.Musicians()
		}
	})
}

// BenchmarkSetUpper benchmarks the SetUpper function.
func BenchmarkSetUpper(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.SetUpper("releaser")
		}
	})
}

// BenchmarkSetFilesize0 benchmarks the SetFilesize0 function.
func BenchmarkSetFilesize0(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.SetFilesize0()
		}
	})
}

// BenchmarkSumSection benchmarks the SumSection function.
func BenchmarkSumSection(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.SumSection()
		}
	})
}

// BenchmarkSumGroup benchmarks the SumGroup function.
func BenchmarkSumGroup(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.SumGroup()
		}
	})
}

// BenchmarkSumPlatform benchmarks the SumPlatform function.
func BenchmarkSumPlatform(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.SumPlatform()
		}
	})
}

// BenchmarkSummary benchmarks the Summary function.
func BenchmarkSummary(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.Summary()
		}
	})
}

// BenchmarkReleasers benchmarks the Releasers function.
func BenchmarkReleasers(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.Releasers()
		}
	})
}

// BenchmarkScenerSQL benchmarks the ScenerSQL function.
func BenchmarkScenerSQL(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_, _ = postgres.ScenerSQL("John Doe")
		}
	})
}

// BenchmarkSimilarToReleaser benchmarks the SimilarToReleaser function.
func BenchmarkSimilarToReleaser(b *testing.B) {
	input := []string{"Lotus", "Amiga", "Test"}
	b.Run("", func(b *testing.B) {
		for range b.N {
			_, _ = postgres.SimilarToReleaser(input...)
		}
	})
}

// BenchmarkSimilarToMagazine benchmarks the SimilarToMagazine function.
func BenchmarkSimilarToMagazine(b *testing.B) {
	input := []string{"Amiga World", "PC Zone"}
	b.Run("", func(b *testing.B) {
		for range b.N {
			_, _ = postgres.SimilarToMagazine(input...)
		}
	})
}

// BenchmarkSimilarToExact benchmarks the SimilarToExact function.
func BenchmarkSimilarToExact(b *testing.B) {
	input := []string{"Breadbox", "Apogee", "Lotus"}
	b.Run("", func(b *testing.B) {
		for range b.N {
			_, _ = postgres.SimilarToExact(input...)
		}
	})
}
