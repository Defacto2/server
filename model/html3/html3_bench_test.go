package html3_test

import (
	"testing"
	"time"

	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model/html3"
	"github.com/aarondl/null/v8"
)

// Use the following command to run all:
// go test -bench=Benchmark -benchmem

func BenchmarkOrderString(b *testing.B) {
	for b.Loop() {
		_ = html3.NameAsc.String()
	}
}

func BenchmarkOrderStringAll(b *testing.B) {
	orders := [...]html3.Order{
		html3.NameAsc, html3.NameDes,
		html3.PublAsc, html3.PublDes,
		html3.PostAsc, html3.PostDes,
		html3.SizeAsc, html3.SizeDes,
		html3.DescAsc, html3.DescDes,
	}
	for b.Loop() {
		for _, o := range orders {
			_ = o.String()
		}
	}
}

func BenchmarkCreated(b *testing.B) {
	f := &models.File{
		Createdat: null.TimeFrom(time.Date(2022, time.December, 31, 0, 0, 0, 0, time.UTC)),
	}
	for b.Loop() {
		_ = html3.Created(f)
	}
}

func BenchmarkPublished(b *testing.B) {
	f := &models.File{
		DateIssuedYear:  null.Int16From(1995),
		DateIssuedMonth: null.Int16From(6),
		DateIssuedDay:   null.Int16From(15),
	}
	for b.Loop() {
		_ = html3.Published(f)
	}
}

func BenchmarkLeadStr(b *testing.B) {
	for b.Loop() {
		_ = html3.LeadStr(20, "test")
	}
}

func BenchmarkIcon(b *testing.B) {
	f := &models.File{
		Filename: null.StringFrom("sample.txt"),
	}
	for b.Loop() {
		_ = html3.Icon(f)
	}
}
