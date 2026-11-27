package name_test

import (
	"slices"
	"testing"

	"github.com/Defacto2/server/handler/releaser/name"
)

func benchPaths() []name.Path {
	all := name.All()
	paths := make([]name.Path, 0, len(all)+2)
	for k := range all {
		paths = append(paths, k)
	}
	slices.Sort(paths)
	return append(paths, "unlisted-group", "invalid/path")
}

func benchNames() []string {
	all := name.All()
	vals := make([]string, 0, len(all)+1)
	for _, v := range all {
		vals = append(vals, v)
	}
	slices.Sort(vals)
	return append(vals, "Razor 1911 Demo & Skillion")
}

func BenchmarkClean(b *testing.B) {
	names := benchNames()
	keep := func(c byte) bool {
		return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == ' '
	}
	for b.Loop() {
		for _, s := range names {
			_ = name.Clean(s, keep)
		}
	}
}

func BenchmarkFind(b *testing.B) {
	names := benchNames()
	for b.Loop() {
		for _, s := range names {
			_ = name.Find(s)
		}
	}
}

func BenchmarkHumanize(b *testing.B) {
	paths := benchPaths()
	for b.Loop() {
		for _, p := range paths {
			_, _ = name.Humanize(p)
		}
	}
}

func BenchmarkObfuscate(b *testing.B) {
	names := benchNames()
	for b.Loop() {
		for _, s := range names {
			_ = name.Obfuscate(s)
		}
	}
}

func BenchmarkString(b *testing.B) {
	paths := benchPaths()
	for b.Loop() {
		for _, p := range paths {
			_ = name.String(p)
		}
	}
}

func BenchmarkValid(b *testing.B) {
	paths := benchPaths()
	for b.Loop() {
		for _, p := range paths {
			_ = name.Valid(p)
		}
	}
}
