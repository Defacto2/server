package dir_test

import (
	"testing"

	"github.com/Defacto2/server/internal/dir"
)

// Benchmark Join function (now without redundant filepath.Clean).
func BenchmarkJoin(b *testing.B) {
	d := dir.Directory("/tmp")
	for range b.N {
		_ = d.Join("testfile.zip")
	}
}

// Benchmark Join with nested paths.
func BenchmarkJoinNested(b *testing.B) {
	d := dir.Directory("/var/lib/defacto2/downloads")
	for range b.N {
		_ = d.Join("archive/subfolder/file.zip")
	}
}

// Benchmark IsDir validation.
func BenchmarkIsDir(b *testing.B) {
	d := dir.Directory("/tmp")
	b.ResetTimer()
	for range b.N {
		_ = d.IsDir()
	}
}

// Benchmark Path() method.
func BenchmarkPath(b *testing.B) {
	d := dir.Directory("/tmp/test")
	for range b.N {
		_ = d.Path()
	}
}

// Benchmark Paths() converting multiple directories.
func BenchmarkPaths(b *testing.B) {
	dirs := []dir.Directory{"/tmp", "/var/lib", "/home/user"}
	for range b.N {
		_ = dir.Paths(dirs...)
	}
}
