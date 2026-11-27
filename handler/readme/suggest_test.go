package readme_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/readme"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

func TestSuggest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		group    string
		content  []string
		expected string
	}{
		{
			name:     "filename match",
			filename: "file1",
			group:    "group1",
			content:  testutil.Content1[:],
			expected: "file1.nfo",
		},
		{
			name:     "group match",
			filename: "file2",
			group:    "group2",
			content:  testutil.Content2[:],
			expected: "group2.dox",
		},
		{
			name:     "filename priority match",
			filename: "file3",
			group:    "group3",
			content:  testutil.Content3[:],
			expected: "file3.nfo",
		},
		{
			name:     "empty",
			filename: "file4",
			group:    "group4",
			content:  []string{},
			expected: "",
		},
		{
			name:     "single entry",
			filename: "demo.zip",
			group:    "grp",
			content:  []string{"only.diz", "game.exe"},
			expected: "only.diz",
		},
		{
			name:     "archive path and ext",
			filename: "dir/file1.zip",
			group:    "group1",
			content:  testutil.Content1[:],
			expected: "file1.nfo",
		},
		{
			name:     "case insensitive",
			filename: "file1.zip",
			group:    "group1",
			content:  []string{"other.txt", "FILE1.NFO"},
			expected: "FILE1.NFO",
		},
		{
			name:     "priority fallback",
			filename: "archive.zip",
			group:    "grp",
			content:  []string{"info.diz", "sub/notes.txt"},
			expected: "sub/notes.txt",
		},
		{
			name:     "candidate fallback",
			filename: "archive.zip",
			group:    "grp",
			content:  []string{"second.asc", "first.diz"},
			expected: "first.diz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			entries := strings.Join(tt.content, "\n")
			got := readme.Suggest(tt.filename, tt.group, entries)
			be.Equal(t, got, tt.expected)
		})
	}
}

func TestSortList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		compact bool
		entries string
		want    []string
	}{
		{
			name:    "empty",
			compact: true,
			entries: "",
			want:    nil,
		},
		{
			name:    "compact filters non-text",
			compact: true,
			entries: "game.exe\nreadme.txt\ninfo.nfo\nfile.diz",
			want:    []string{"info.nfo", "readme.txt", "file.diz"},
		},
		{
			name:    "non-compact keeps all",
			compact: false,
			entries: "game.exe\nreadme.txt\ninfo.nfo\nfile.diz",
			want:    []string{"info.nfo", "readme.txt", "file.diz", "game.exe"},
		},
		{
			name:    "subdir depth",
			compact: true,
			entries: "a/b/deep.nfo\nsub/mid.nfo\nroot.diz",
			want:    []string{"root.diz", "sub/mid.nfo", "a/b/deep.nfo"},
		},
		{
			name:    "banned and blank lines",
			compact: true,
			entries: "\n  \nscene.org\nSCENE.ORG.TXT\nwe-will.sue\nwww.acid.org\nkeep.nfo\n",
			want:    []string{"keep.nfo"},
		},
		{
			name:    "crlf entries",
			compact: true,
			entries: "file.diz\r\nreadme.nfo\r\n",
			want:    []string{"readme.nfo", "file.diz"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := readme.SortList(tt.compact, tt.entries)
			be.Equal(t, got, tt.want)
		})
	}
}
