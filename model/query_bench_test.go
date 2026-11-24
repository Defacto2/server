package model_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
)

// Run all index benchmarks with:
// go test -run=^$ -bench=^BenchmarkIndex -benchmem -count=3 ./model

func BenchmarkIndexSearchDescriptions(b *testing.B) {
	db := testutil.DB(b)
	sl := logs.Discard()
	ctx := b.Context()
	terms := []string{"defacto", "razor 1911"}

	b.ResetTimer()
	for range b.N {
		if _, err := model.OnlyDescriptions(ctx, sl, db, terms); err != nil {
			b.Fatal(err)
		}
		var sum model.Summary
		if err := sum.ByDescription(ctx, db, terms); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIndexSearchFilenames(b *testing.B) {
	db := testutil.DB(b)
	ctx := b.Context()
	terms := []string{"razor", "nfo"}

	b.ResetTimer()
	for range b.N {
		if _, err := model.OnlyFilenames(ctx, db, terms); err != nil {
			b.Fatal(err)
		}
		var sum model.Summary
		if err := sum.ByFilename(ctx, db, terms); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIndexCategoryList(b *testing.B) {
	db := testutil.DB(b)
	ctx := b.Context()
	var nfo model.Nfo
	var intro model.IntroMsDos

	b.ResetTimer()
	for range b.N {
		if _, err := nfo.List(ctx, db, 1, 25); err != nil {
			b.Fatal(err)
		}
		if _, err := intro.List(ctx, db, 1, 25); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIndexReleasersWhere(b *testing.B) {
	db := testutil.DB(b)
	ctx := b.Context()

	b.ResetTimer()
	for range b.N {
		if _, err := model.ReleasersWhere(ctx, db, "razor-1911"); err != nil {
			b.Fatal(err)
		}
		var sum model.Summary
		if err := sum.ByReleaser(ctx, db, "razor-1911"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIndexOneByUUID(b *testing.B) {
	db := testutil.DB(b)
	ctx := b.Context()
	rec, err := model.OneFile(ctx, db, 1)
	if err != nil {
		b.Fatal(err)
	}
	uid := strings.TrimSpace(rec.UUID.String)

	b.ResetTimer()
	for range b.N {
		if _, err := model.OneByUUID(ctx, db, true, uid); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIndexUUIDPrefix(b *testing.B) {
	db := testutil.DB(b)
	ctx := b.Context()
	rec, err := model.OneFile(ctx, db, 1)
	if err != nil {
		b.Fatal(err)
	}
	prefix := strings.ToLower(strings.TrimSpace(rec.UUID.String))[:8]

	b.ResetTimer()
	for range b.N {
		if _, err := model.OnlyUniqueIDsWithPrefixes(ctx, db, nil, nil, []string{prefix}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIndexDedupeLookups(b *testing.B) {
	db := testutil.DB(b)
	ctx := b.Context()
	rec, err := model.OneFile(ctx, db, 1)
	if err != nil {
		b.Fatal(err)
	}
	hash := strings.TrimSpace(rec.FileIntegrityStrong.String)

	b.ResetTimer()
	for range b.N {
		if hash != "" {
			if _, err := model.ExistHash(ctx, db, hash); err != nil {
				b.Fatal(err)
			}
		}
		if _, _, err := model.OneDemozoo(ctx, db, 1000); err != nil {
			b.Fatal(err)
		}
		if _, _, err := model.OnePouet(ctx, db, 1000); err != nil {
			b.Fatal(err)
		}
	}
}
