package dir_test

import (
	"testing"

	"github.com/Defacto2/server/internal/dir"
)

// Benchmark Join function (now without redundant filepath.Clean)
func BenchmarkJoin(b *testing.B) {
	d := dir.Directory("/tmp")
	for i := 0; i < b.N; i++ {
		_ = d.Join("testfile.zip")
	}
}

// Benchmark Join with nested paths
func BenchmarkJoinNested(b *testing.B) {
	d := dir.Directory("/var/lib/defacto2/downloads")
	for i := 0; i < b.N; i++ {
		_ = d.Join("archive/subfolder/file.zip")
	}
}

// Benchmark IsDir validation
func BenchmarkIsDir(b *testing.B) {
	d := dir.Directory("/tmp")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = d.IsDir()
	}
}

// Benchmark Path() method
func BenchmarkPath(b *testing.B) {
	d := dir.Directory("/tmp/test")
	for i := 0; i < b.N; i++ {
		_ = d.Path()
	}
}

// Benchmark Paths() converting multiple directories
func BenchmarkPaths(b *testing.B) {
	dirs := []dir.Directory{"/tmp", "/var/lib", "/home/user"}
	for i := 0; i < b.N; i++ {
		_ = dir.Paths(dirs...)
	}
}
