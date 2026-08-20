//nolint:gochecknoglobals
package readme_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/readme"
	"github.com/nalgeon/be"
)

var content1 = [...]string{
	"file1.nfo",
	"file1.txt",
	"file1.unp",
	"file1.doc",
}

var content2 = [...]string{
	"file.diz",
	"file.asc",
	"file.1st",
	"group2.dox",
}

var content3 = [...]string{
	"file3.nfo",
	"file.txt",
	"file30.unp",
	"file3x.doc",
	"filex3.diz",
	"file3.asc",
	"file3.1st",
	"file3.dox",
}

func Test_Handler_ReadmeSuggest(t *testing.T) {
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
			content:  content1[:],
			expected: "file1.nfo",
		},
		{
			filename: "file2",
			group:    "group2",
			content:  content2[:],
			expected: "group2.dox",
		},
		{
			filename: "file3",
			group:    "group3",
			content:  content3[:],
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
