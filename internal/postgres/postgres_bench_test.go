package postgres_test

import (
	"log/slog"
	"testing"

	"github.com/Defacto2/server/internal/postgres"
)

func BenchmarkVersionString(b *testing.B) {
	v := postgres.Version("PostgreSQL 13.8 on x86_64-pc-linux-gnu")
	for b.Loop() {
		_ = v.String()
	}
}

func BenchmarkConnectionValidate(b *testing.B) {
	logger := slog.New(slog.DiscardHandler)
	conn := postgres.Connection{URL: "postgres://localhost:5432/test"}
	for b.Loop() {
		_ = conn.Validate(logger)
	}
}

func BenchmarkRoles(b *testing.B) {
	for b.Loop() {
		_ = postgres.Roles()
	}
}

func BenchmarkReleasersAlphabetical(b *testing.B) {
	for b.Loop() {
		_ = postgres.ReleasersAlphabetical()
	}
}

func BenchmarkBBSsAlphabetical(b *testing.B) {
	for b.Loop() {
		_ = postgres.BBSsAlphabetical()
	}
}

func BenchmarkMagazinesAlphabetical(b *testing.B) {
	for b.Loop() {
		_ = postgres.MagazinesAlphabetical()
	}
}

func BenchmarkReleasersProlific(b *testing.B) {
	for b.Loop() {
		_ = postgres.ReleasersProlific()
	}
}

func BenchmarkReleasersOldest(b *testing.B) {
	for b.Loop() {
		_ = postgres.ReleasersOldest()
	}
}

func BenchmarkSceners(b *testing.B) {
	for b.Loop() {
		_ = postgres.Sceners()
	}
}

func BenchmarkWriters(b *testing.B) {
	for b.Loop() {
		_ = postgres.Writers()
	}
}

func BenchmarkArtists(b *testing.B) {
	for b.Loop() {
		_ = postgres.Artists()
	}
}

func BenchmarkCoders(b *testing.B) {
	for b.Loop() {
		_ = postgres.Coders()
	}
}

func BenchmarkMusicians(b *testing.B) {
	for b.Loop() {
		_ = postgres.Musicians()
	}
}

func BenchmarkSetUpper(b *testing.B) {
	for b.Loop() {
		_ = postgres.SetUpper("releaser")
	}
}

func BenchmarkSetFilesize0(b *testing.B) {
	for b.Loop() {
		_ = postgres.SetFilesize0()
	}
}

func BenchmarkSumSection(b *testing.B) {
	for b.Loop() {
		_ = postgres.SumSection()
	}
}

func BenchmarkSumGroup(b *testing.B) {
	for b.Loop() {
		_ = postgres.SumGroup()
	}
}

func BenchmarkSumPlatform(b *testing.B) {
	for b.Loop() {
		_ = postgres.SumPlatform()
	}
}

func BenchmarkSummary(b *testing.B) {
	for b.Loop() {
		_ = postgres.Summary()
	}
}

func BenchmarkReleasers(b *testing.B) {
	for b.Loop() {
		_ = postgres.Releasers()
	}
}

func BenchmarkScenerSQL(b *testing.B) {
	input := [...]string{"ABC", "ati", "Grim"}
	for b.Loop() {
		for _, name := range input {
			_, _ = postgres.ScenerSQL(name)
		}
	}
}

func BenchmarkSimilarToReleaser(b *testing.B) {
	input := []string{"Razor", "Amiga", "Test"}
	for b.Loop() {
		_, _ = postgres.SimilarToReleaser(input...)
	}
}

func BenchmarkSimilarToMagazine(b *testing.B) {
	input := []string{"Reality Check", "NWR"}
	for b.Loop() {
		_, _ = postgres.SimilarToMagazine(input...)
	}
}

func BenchmarkSimilarToExact(b *testing.B) {
	input := []string{"Razor 1911", "Defacto", "Lotus"}
	for b.Loop() {
		_, _ = postgres.SimilarToExact(input...)
	}
}
