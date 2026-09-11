package app_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

func c(t *testing.T) *echo.Context {
	t.Helper()

	e := echo.New()
	r := httptest.NewRequestWithContext(
		t.Context(), http.MethodPost, "/", strings.NewReader("{}"))
	r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w := httptest.NewRecorder()

	return e.NewContext(r, w)
}

func TestArtifact(t *testing.T) {
	t.Parallel()

	ds := app.Dirs{
		ID:  1,
		URI: "9b1c6",
	}

	sl := logs.Discard()
	db := testutil.DB(t)

	err := ds.Artifact(sl, c(t), db)
	be.Err(t, err)
}

func TestStatusErr(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	err := app.StatusErr(sl, c(t), 1, "x")
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
			expected: nil,
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
