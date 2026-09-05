package readme_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/readme"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

func TestReadmeSuggest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		filename string
		group    string
		content  []string
		expected string
	}{
		{
			filename: "file1",
			group:    "group1",
			content:  testutil.Content1[:],
			expected: "file1.nfo",
		},
		{
			filename: "file2",
			group:    "group2",
			content:  testutil.Content2[:],
			expected: "group2.dox",
		},
		{
			filename: "file3",
			group:    "group3",
			content:  testutil.Content3[:],
			expected: "file3.nfo",
		},
		{
			filename: "file4",
			group:    "group4",
			content:  []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.filename+"_"+tt.group, func(t *testing.T) {
			t.Parallel()

			entries := strings.Join(tt.content, "\n")
			got := readme.Suggest(tt.filename, tt.group, entries)
			be.Equal(t, got, tt.expected)
		})
	}
}
