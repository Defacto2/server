package readme_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/readme"
	"github.com/nalgeon/be"
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
				be.Equal(t, tt.expected[i], v)
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
			be.Equal(t, tt.expected, result)
		})
	}
}

func TestRead(t *testing.T) {
	t.Parallel()
	p, _, err := readme.Read(nil, "", "")
	be.Err(t, err)
	be.True(t, len(p) == 0)
}

func TestRemoveCtrls(t *testing.T) {
	t.Parallel()
	p := []byte("a\x1b[1;cabc")
	r := readme.RemoveCtrls(p)
	be.Equal(t, []byte("aabc"), r)
}

func TestIncompatibleANSI(t *testing.T) {
	t.Parallel()
	b, err := readme.IncompatibleANSI(nil)
	be.Err(t, err, nil)
	be.True(t, !b)
	r := strings.NewReader("a\x1b[1;cabc")
	b, err = readme.IncompatibleANSI(r)
	be.Err(t, err, nil)
	be.True(t, !b)
	r = strings.NewReader("a\x1b[Acabc")
	b, err = readme.IncompatibleANSI(r)
	be.Err(t, err, nil)
	be.True(t, b)
}
