package dir_test

import (
	"testing"

	"github.com/Defacto2/server/internal/dir"
)

func BenchmarkJoin(b *testing.B) {
	d := dir.Directory("/tmp")
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = d.Join("testfile.zip")
		}
	})
}

func BenchmarkJoinNested(b *testing.B) {
	d := dir.Directory("/var/lib/defacto2/downloads")
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = d.Join("archive/subfolder/file.zip")
		}
	})
}

func BenchmarkIsDir(b *testing.B) {
	d := dir.Directory("/tmp")
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = d.IsDir()
		}
	})
}

func BenchmarkPath(b *testing.B) {
	d := dir.Directory("/tmp/test")
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = d.Path()
		}
	})
}

func BenchmarkPaths(b *testing.B) {
	dirs := []dir.Directory{"/tmp", "/var/lib", "/home/user"}
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = dir.Paths(dirs...)
		}
	})
}
