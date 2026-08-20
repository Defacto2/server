package app_test

import (
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/nalgeon/be"
)

func TestArtifact(t *testing.T) {
	t.Parallel()

	dir := app.Dirs{}
	err := dir.Artifact(t.Context(), nil, echoCtx(t), nil)
	be.Err(t, err)
}

func TestFileMissingErr(t *testing.T) {
	t.Parallel()

	err := app.FileMissingErr(nil, echoCtx(t), "", nil)
	be.Err(t, err)
}

func TestForbiddenErr(t *testing.T) {
	t.Parallel()

	err := app.ForbiddenErr(nil, echoCtx(t), "", nil)
	be.Err(t, err)
}

func TestInternalErr(t *testing.T) {
	t.Parallel()

	err := app.InternalErr(nil, echoCtx(t), "", nil)
	be.Err(t, err)
}

func TestStatusErr(t *testing.T) {
	t.Parallel()

	err := app.StatusErr(nil, echoCtx(t), -1, "")
	be.Err(t, err)
}

func TestSortContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "empty input",
			content:  "",
			expected: []string{""},
		},
		{
			name:     "single file",
			content:  "file.txt",
			expected: []string{"file.txt"},
		},
		{
			name:     "files and directories mixed",
			content:  "file1.txt\ndir/\nfile2.txt\nsubdir/\nfile3.exe",
			expected: []string{"file3.exe", "file1.txt", "file2.txt"},
		},
		{
			name:     "no directories",
			content:  "a.txt\nb.exe\nc.doc",
			expected: []string{"c.doc", "b.exe", "a.txt"},
		},
		{
			name:     "only directories",
			content:  "dir1/\ndir2/\ndir3/",
			expected: []string{},
		},
		{
			name:     "case insensitive sorting",
			content:  "Z.txt\nA.txt\nb.exe\nC.exe",
			expected: []string{"b.exe", "C.exe", "A.txt", "Z.txt"},
		},
		{
			name:     "sorted by extension then name",
			content:  "zebra.doc\napple.txt\nbanana.doc\ncar.txt",
			expected: []string{"banana.doc", "zebra.doc", "apple.txt", "car.txt"},
		},
		{
			name:     "filter out all directories",
			content:  "a/\nb/\nc/\nd/",
			expected: []string{},
		},
		{
			name:     "mixed depths - shallow first",
			content:  "deep/nested/file.txt\nshallow.txt\nmiddle/file.txt",
			expected: []string{"shallow.txt", "middle/file.txt", "deep/nested/file.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sorted := app.SortContent(tt.content)
			be.Equal(t, len(sorted), len(tt.expected))

			for i, got := range sorted {
				be.Equal(t, got, tt.expected[i])
			}

			// verify no entries end with "/" (directories filtered)
			for _, entry := range sorted {
				if len(entry) == 0 {
					continue
				}
				be.True(t, entry[len(entry)-1] != '/')
			}
		})
	}
}
