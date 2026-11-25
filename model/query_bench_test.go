package model_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
)

// Run all index benchmarks with:
// go test -run=^$ -bench=^BenchmarkIndex -benchmem -count=3 ./model

func BenchmarkIndexSearchDescriptions(b *testing.B) {
	db := testutil.DB(b)
	sl := logs.Discard()
	ctx := b.Context()
	terms := []string{"defacto", "razor 1911"}

	for b.Loop() {
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

	for b.Loop() {
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

	for b.Loop() {
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

	for b.Loop() {
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

	for b.Loop() {
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

	for b.Loop() {
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

	for b.Loop() {
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

func BenchmarkIndexCategoryRare(b *testing.B) {
	db := testutil.DB(b)
	ctx := b.Context()
	var java model.Java
	var console model.Console
	var mac model.Macos
	var trial model.TrialCrackme

	for b.Loop() {
		if _, err := java.List(ctx, db, 1, 25); err != nil {
			b.Fatal(err)
		}
		if _, err := console.List(ctx, db, 1, 25); err != nil {
			b.Fatal(err)
		}
		if _, err := mac.List(ctx, db, 1, 25); err != nil {
			b.Fatal(err)
		}
		if _, err := trial.List(ctx, db, 1, 25); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIndexWriteInsert(b *testing.B) {
	db := testutil.DB(b)
	ctx := b.Context()
	b.Cleanup(func() {
		_, _ = db.ExecContext(context.WithoutCancel(ctx),
			`SELECT setval('"files_id_seq"', COALESCE((SELECT MAX(id) FROM files), 1))`)
	})

	i := 0
	for b.Loop() {
		i++
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			b.Fatal(err)
		}
		f := models.File{
			UUID:                null.StringFrom(fmt.Sprintf("01900000-0000-7000-8000-%012d", i)),
			Filename:            null.StringFrom(fmt.Sprintf("razor1911-bench-%d.zip", i)),
			RecordTitle:         null.StringFrom("Razor 1911 Benchmark Release Title"),
			Comment:             null.StringFrom("A typical scene release description mentioning defacto2, dos cracktro, and nfo artwork."),
			Section:             null.StringFrom("demo"),
			Platform:            null.StringFrom("dos"),
			GroupBrandFor:       null.StringFrom("RAZOR 1911"),
			GroupBrandBy:        null.StringFrom("DEFACTO2"),
			DateIssuedYear:      null.Int16From(1996),
			DateIssuedMonth:     null.Int16From(7),
			DateIssuedDay:       null.Int16From(4),
			FileIntegrityStrong: null.StringFrom(fmt.Sprintf("%096x", i)),
			WebIDDemozoo:        null.Int64From(999000 + int64(i)),
			WebIDPouet:          null.Int64From(888000 + int64(i)),
		}
		if err := f.Insert(ctx, tx, boil.Infer()); err != nil {
			_ = tx.Rollback()
			b.Fatal(err)
		}
		if err := tx.Rollback(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIndexWriteUpdateHOT(b *testing.B) {
	db := testutil.DB(b)
	ctx := b.Context()

	for b.Loop() {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			b.Fatal(err)
		}
		if err := model.UpdateEmulateRunProgram(ctx, tx, 1, "BENCH.EXE"); err != nil {
			_ = tx.Rollback()
			b.Fatal(err)
		}
		if err := tx.Rollback(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIndexWriteUpdateIndexed(b *testing.B) {
	db := testutil.DB(b)
	ctx := b.Context()
	wl := boil.Whitelist(
		models.FileColumns.Updatedat,
		models.FileColumns.Filename,
		models.FileColumns.RecordTitle,
		models.FileColumns.Comment,
		models.FileColumns.Section,
		models.FileColumns.Platform,
		models.FileColumns.GroupBrandFor,
	)

	i := 0
	for b.Loop() {
		i++
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			b.Fatal(err)
		}
		f := models.File{
			ID:            1,
			Filename:      null.StringFrom(fmt.Sprintf("razor-edited-%d.nfo", i)),
			RecordTitle:   null.StringFrom("Edited Benchmark Title"),
			Comment:       null.StringFrom("Updated scene release comment referencing razor 1911 and defacto2 archive."),
			Section:       null.StringFrom("releaseinformation"),
			Platform:      null.StringFrom("text"),
			GroupBrandFor: null.StringFrom("RAZOR 1911"),
		}
		if _, err := f.Update(ctx, tx, wl); err != nil {
			_ = tx.Rollback()
			b.Fatal(err)
		}
		if err := tx.Rollback(); err != nil {
			b.Fatal(err)
		}
	}
}
