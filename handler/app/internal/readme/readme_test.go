package readme_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/app/internal/readme"
	"github.com/stretchr/testify/assert"
)

func TestSortContent(t *testing.T) {
	tests := []struct {
		content  []string
		expected []string
	}{
		{
			content:  nil,
			expected: nil,
		},
		{
			content: []string{
				"dir1/file1",
				"dir2/file2",
				"dir1/subdir/file3",
				"file4",
			},
			expected: []string{
				"file4",
				"dir1/file1",
				"dir2/file2",
				"dir1/subdir/file3",
			},
		},
		{
			content: []string{
				"dir1/file1",
				"dir1/subdir/file2",
				"dir2/file3",
				"file4",
			},
			expected: []string{
				"file4",
				"dir1/file1",
				"dir2/file3",
				"dir1/subdir/file2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.content, ","), func(t *testing.T) {
			// Make a copy of the original content
			originalContent := make([]string, len(tt.content))
			copy(originalContent, tt.content)

			// Sort the content using the SortContent function
			sortedContent := readme.SortContent(tt.content...)

			for i, v := range sortedContent {
				assert.Equal(t, tt.expected[i], v)
			}
		})
	}
}

func TestReadmeSuggest(t *testing.T) {
	tests := []struct {
		filename string
		group    string
		content  []string
		expected string
	}{
		{
			filename: "file1",
			group:    "group1",
			content: []string{
				"file1.nfo",
				"file1.txt",
				"file1.unp",
				"file1.doc",
			},
			expected: "file1.nfo",
		},
		{
			filename: "file2",
			group:    "group2",
			content: []string{
				"file.diz",
				"file.asc",
				"file.1st",
				"group2.dox",
			},
			expected: "group2.dox",
		},
		{
			filename: "file3",
			group:    "group3",
			content: []string{
				"file3.nfo",
				"file.txt",
				"file30.unp",
				"file3x.doc",
				"filex3.diz",
				"file3.asc",
				"file3.1st",
				"file3.dox",
			},
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
			result := readme.Suggest(tt.filename, tt.group, tt.content...)
			assert.Equal(t, tt.expected, result)
		})
	}
}
