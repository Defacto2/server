package readme_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/readme"
	"github.com/Defacto2/server/internal/testutil"
)

func BenchmarkNormalize(b *testing.B) {
	for _, tt := range testutil.ANSITests(b) {
		b.Run(strings.TrimSpace(tt.Name), func(b *testing.B) {
			b.SetBytes(int64(len(tt.Data)))
			buf := make([]byte, len(tt.Data))
			for b.Loop() {
				copy(buf, tt.Data)
				_ = readme.Normalize(buf)
			}
		})
	}
}

func BenchmarkSuggest(b *testing.B) {
	entries := []struct {
		name  string
		file  string
		group string
		list  string
	}{
		{"content1", "file1.zip", "group1", strings.Join(testutil.Content1[:], "\n")},
		{"content2", "file2.zip", "group2", strings.Join(testutil.Content2[:], "\n")},
		{"content3", "file3.zip", "group3", strings.Join(testutil.Content3[:], "\n")},
	}
	for _, tt := range entries {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				_ = readme.Suggest(tt.file, tt.group, tt.list)
			}
		})
	}
}

func BenchmarkSortList(b *testing.B) {
	entries := strings.Join(testutil.Content3[:], "\n") +
		"\ngame.exe\nscene.org\nsub/readme.txt\na/b/deep.nfo"
	for _, compact := range []bool{true, false} {
		name := "all"
		if compact {
			name = "compact"
		}
		b.Run(name, func(b *testing.B) {
			for b.Loop() {
				_ = readme.SortList(compact, entries)
			}
		})
	}
}
