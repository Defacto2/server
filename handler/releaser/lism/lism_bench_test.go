//nolint:testpackage
package lism

import (
	"maps"
	"slices"
	"strings"
	"testing"
)

// Benchmarks for various Match patterns, to run:
// go test -bench=Benchmark -benchmem

func BenchmarkMatch00(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, s := range testMatches {
			_ = Match00(s)
		}
	}
}

func BenchmarkMatch01(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, s := range testMatches {
			_ = Match01(s)
		}
	}
}

func BenchmarkMatch02(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, s := range testMatches {
			_ = Match02(s)
		}
	}
}

func BenchmarkMarch03(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, s := range testMatches {
			_ = Match03(s)
		}
	}
}

// Benchmarks for init ranges

func BenchmarkInit00(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		Init00()
	}
}

func BenchmarkInit01(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		Init01()
	}
}

var testMatches = [...]string{cpc2001, uart, core, crue, "nso", epix, neua, tlf, iirg, quick, scud, nasa, quick}

func Match00(s string) []Path {
	var partials []Path
	for partial, values := range maps.All(initialisms) {
		for value := range slices.Values(values) {
			if strings.EqualFold(value, s) {
				partials = append(partials, partial)
			}
		}
	}
	return partials
}

func Match01(s string) []Path {
	p := Path(s)
	if _, ok := initialisms[p]; ok {
		return []Path{p}
	}
	return nil
}

func Match02(s string) []Path {
	return pathsIndex[strings.ToLower(s)]
}

func Match03(s string) Path {
	return pathIndex[strings.ToLower(s)]
}

func init() {
	pathsIndex = make(map[string][]Path)
	for partial, values := range initialisms {
		for _, value := range values {
			key := strings.ToLower(value)
			pathsIndex[key] = append(pathsIndex[key], partial)
		}
	}
}

func Init00() {
	pathIndex = make(map[string]Path, len(initialisms)*2)
	for uri, values := range initialisms {
		for _, value := range values {
			pathIndex[strings.ToLower(value)] = uri
		}
	}
}

func Init01() {
	pathsIndex = make(map[string][]Path)
	for uri, values := range initialisms {
		for _, value := range values {
			key := strings.ToLower(value)
			pathsIndex[key] = append(pathsIndex[key], uri)
		}
	}
}
