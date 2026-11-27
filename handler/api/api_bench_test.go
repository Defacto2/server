package api_test

import (
	"database/sql"
	"testing"

	"github.com/Defacto2/server/handler/api"
)

func BenchmarkApiMarkup(b *testing.B) {
	const html = `<div class="content">
		<p class="lead">This is a <strong>test</strong> with <a href="https://example.com" class="link" id="test">links</a> and <span style="color: red;">formatting</span>.</p>
		<p>Another paragraph with <a name="anchor">anchor</a> and <data-info="test">data attributes</data-info>.</p>
	</div>`

	b.ReportAllocs()
	for b.Loop() {
		_ = api.Markup(html)
	}
}

// BenchmarkCategoriesAPI benchmarks the CategoriesAPI with realistic stats calculation.
func BenchmarkCategoriesAPI(b *testing.B) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		b.Skipf("Could not create database connection: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			b.Fatal(err)
		}
	}()

	if err := db.PingContext(b.Context()); err != nil {
		b.Skipf("Could not ping database: %v", err)
	}

	b.ReportAllocs()
	for b.Loop() {
		_, c := newRequest(b, apiv1+"/categories")
		_ = api.CategoriesAPI(c, db)
	}
}
