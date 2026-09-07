package areacode_test

import (
	"testing"

	"github.com/Defacto2/server/handler/areacode"
)

// Use the following command to run all:
// go test -bench=Benchmark -benchmem

var (
	checks  = [...]areacode.NAN{0, 999, 911, 555, 200, 303, 404, 505, 655, 211, 888, 800, 801, 705}
	lookup  = [...]string{"0", "999", "911", "555", "200", "303", "404", "505", "655", "211", "888", "800", "801", "705"}
	regions = [...]string{"unknown", "ALBERTA", "new york", "Utah", "ohio", "new"}
	inputs  = [...]any{"NYC", 212, "NY", areacode.NAN(415), "california", 9999, "ON", "FooBar", 312}
)

func BenchmarkValid(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		for _, n := range checks {
			_ = n.Valid()
		}
	}
}

func BenchmarkHTML(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		for _, n := range checks {
			_ = n.HTML() // this runs both HTML() and Valid()
		}
	}
}

func BenchmarkLookupNAN(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		for _, n := range checks {
			_ = areacode.Lookup(n)
		}
	}
}

func BenchmarkLookupStr(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		for _, s := range lookup {
			_ = areacode.Lookup(s)
		}
	}
}

func BenchmarkQueryNAN(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		for _, n := range checks {
			_ = areacode.Query(n)
		}
	}
}

func BenchmarkQueryStr(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		for _, s := range lookup {
			_ = areacode.Query(s)
		}
	}
}

func BenchmarkQueries(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		for _, s := range lookup {
			_ = areacode.Queries(s)
		}
	}
}

func BenchmarkRegionNAN(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		for _, n := range checks {
			_ = areacode.RegionByCode(n)
		}
	}
}

func BenchmarkRegionByName(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		for _, s := range regions {
			_ = areacode.RegionByName(s)
		}
	}
}

func BenchmarkRegionContains(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		for _, s := range lookup {
			_ = areacode.RegionContains(s)
		}
	}
}

func BenchmarkRegionByAbbr(b *testing.B) {
	b.ReportAllocs()

	abbrs := areacode.Abbreviations()
	for b.Loop() {
		for _, abbr := range abbrs {
			_, _ = areacode.RegionByAbbr(abbr)
		}
	}
}

func BenchmarkLookups(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		_ = areacode.Lookups(inputs)
	}
}
