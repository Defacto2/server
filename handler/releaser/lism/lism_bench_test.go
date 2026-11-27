package lism_test

import (
	"slices"
	"testing"

	"github.com/Defacto2/server/handler/releaser/lism"
)

func benchPaths() []lism.Path {
	ins := lism.All()
	paths := make([]lism.Path, 0, len(ins)+1)
	for k := range ins {
		paths = append(paths, k)
	}
	slices.Sort(paths)
	return append(paths, "nonexistent-group")
}

func benchInitialisms() []string {
	ins := lism.All()
	vals := make([]string, 0, len(ins)+1)
	for _, vs := range ins {
		vals = append(vals, vs...)
	}
	slices.Sort(vals)
	return append(vals, "NONEXISTENT")
}

func BenchmarkExist(b *testing.B) {
	paths := benchPaths()
	for b.Loop() {
		for _, p := range paths {
			_ = lism.Exist(p)
		}
	}
}

func BenchmarkFind(b *testing.B) {
	vals := benchInitialisms()
	for b.Loop() {
		for _, v := range vals {
			_ = lism.Find(v)
		}
	}
}

func BenchmarkFindOne(b *testing.B) {
	vals := benchInitialisms()
	for b.Loop() {
		for _, v := range vals {
			_ = lism.FindOne(v)
		}
	}
}

func BenchmarkInitialism(b *testing.B) {
	paths := benchPaths()
	for b.Loop() {
		for _, p := range paths {
			_ = lism.Initialism(p)
		}
	}
}

func BenchmarkString(b *testing.B) {
	paths := benchPaths()
	for b.Loop() {
		for _, p := range paths {
			_ = lism.String(p)
		}
	}
}
