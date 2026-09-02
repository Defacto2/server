package postgres_test

import (
	"log/slog"
	"testing"

	"github.com/Defacto2/server/internal/postgres"
)

func BenchmarkVersionString(b *testing.B) {
	v := postgres.Version("PostgreSQL 13.8 on x86_64-pc-linux-gnu")
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = v.String()
		}
	})
}

func BenchmarkConnectionValidate(b *testing.B) {
	logger := slog.Default()
	conn := postgres.Connection{URL: "postgres://localhost:5432/test"}
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = conn.Validate(logger)
		}
	})
}

func BenchmarkRoles(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.Roles()
		}
	})
}

func BenchmarkReleasersAlphabetical(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.ReleasersAlphabetical()
		}
	})
}

func BenchmarkBBSsAlphabetical(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.BBSsAlphabetical()
		}
	})
}

func BenchmarkMagazinesAlphabetical(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.MagazinesAlphabetical()
		}
	})
}

func BenchmarkReleasersProlific(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.ReleasersProlific()
		}
	})
}

func BenchmarkReleasersOldest(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.ReleasersOldest()
		}
	})
}

func BenchmarkSceners(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.Sceners()
		}
	})
}

func BenchmarkWriters(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.Writers()
		}
	})
}

func BenchmarkArtists(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.Artists()
		}
	})
}

func BenchmarkCoders(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.Coders()
		}
	})
}

func BenchmarkMusicians(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.Musicians()
		}
	})
}

func BenchmarkSetUpper(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.SetUpper("releaser")
		}
	})
}

func BenchmarkSetFilesize0(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.SetFilesize0()
		}
	})
}

func BenchmarkSumSection(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.SumSection()
		}
	})
}

func BenchmarkSumGroup(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.SumGroup()
		}
	})
}

func BenchmarkSumPlatform(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.SumPlatform()
		}
	})
}

func BenchmarkSummary(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.Summary()
		}
	})
}

func BenchmarkReleasers(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = postgres.Releasers()
		}
	})
}

func BenchmarkScenerSQL(b *testing.B) {
	input := []string{"ABC", "ati", "Grim"}
	b.Run("", func(b *testing.B) {
		for _, name := range input {
			for range b.N {
				_, _ = postgres.ScenerSQL(name)
			}
		}
	})
}

func BenchmarkSimilarToReleaser(b *testing.B) {
	input := []string{"Razor", "Amiga", "Test"}
	b.Run("", func(b *testing.B) {
		for range b.N {
			_, _ = postgres.SimilarToReleaser(input...)
		}
	})
}

func BenchmarkSimilarToMagazine(b *testing.B) {
	input := []string{"Reality Check", "NWR"}
	b.Run("", func(b *testing.B) {
		for range b.N {
			_, _ = postgres.SimilarToMagazine(input...)
		}
	})
}

func BenchmarkSimilarToExact(b *testing.B) {
	input := []string{"Razor 1911", "Defacto", "Lotus"}
	b.Run("", func(b *testing.B) {
		for range b.N {
			_, _ = postgres.SimilarToExact(input...)
		}
	})
}
