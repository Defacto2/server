package app_test

import (
	"database/sql"
	"testing"

	"github.com/Defacto2/server/handler/app"
)

func BenchmarkApiMarkup(b *testing.B) {
	html := `<div class="content">
		<p class="lead">This is a <strong>test</strong> with <a href="https://example.com" class="link" id="test">links</a> and <span style="color: red;">formatting</span>.</p>
		<p>Another paragraph with <a name="anchor">anchor</a> and <data-info="test">data attributes</data-info>.</p>
	</div>`

	b.Run("", func(b *testing.B) {
		for range b.N {
			app.APIMarkup(html)
		}
	})
}

// BenchmarkCategoriesAPI benchmarks the CategoriesAPI with realistic stats calculation.
func BenchmarkCategoriesAPI(b *testing.B) {
	w, c := newRequest(b, api+"/categories")
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

	b.ResetTimer()
	for b.Loop() {
		_ = app.CategoriesAPI(b.Context(), c, db)
		w.Body.Reset()
	}
}
