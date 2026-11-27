package releaser_test

import (
	"slices"
	"testing"

	"github.com/Defacto2/server/handler/releaser"
	"github.com/Defacto2/server/handler/releaser/lism"
)

func benchPaths() []string {
	ins := lism.All()
	paths := make([]string, 0, len(ins))
	for k := range ins {
		paths = append(paths, string(k))
	}
	slices.Sort(paths)
	return paths
}

func benchNames() []string {
	paths := benchPaths()
	names := make([]string, len(paths))
	for i, p := range paths {
		names[i] = releaser.Humanize(p)
	}
	return names
}

func benchIndexed() []string {
	paths := benchPaths()
	indexed := make([]string, len(paths))
	for i, p := range paths {
		indexed[i] = releaser.Index(p)
	}
	return indexed
}

func BenchmarkCell(b *testing.B) {
	names := benchNames()
	for b.Loop() {
		for _, n := range names {
			_ = releaser.Cell(n)
		}
	}
}

func BenchmarkClean(b *testing.B) {
	names := benchNames()
	for b.Loop() {
		for _, n := range names {
			_ = releaser.Clean(n)
		}
	}
}

func BenchmarkHumanize(b *testing.B) {
	paths := benchPaths()
	for b.Loop() {
		for _, p := range paths {
			_ = releaser.Humanize(p)
		}
	}
}

func BenchmarkIndex(b *testing.B) {
	paths := benchPaths()
	for b.Loop() {
		for _, p := range paths {
			_ = releaser.Index(p)
		}
	}
}

func BenchmarkLink(b *testing.B) {
	paths := benchPaths()
	for b.Loop() {
		for _, p := range paths {
			_ = releaser.Link(p)
		}
	}
}

func BenchmarkObfuscate(b *testing.B) {
	names := benchNames()
	for b.Loop() {
		for _, n := range names {
			_ = releaser.Obfuscate(n)
		}
	}
}

func BenchmarkTitle(b *testing.B) {
	indexed := benchIndexed()
	for b.Loop() {
		for _, s := range indexed {
			_ = releaser.Title(s)
		}
	}
}
