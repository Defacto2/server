package readme_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/readme"
	"github.com/nalgeon/be"
)

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
			result := readme.Suggest(tt.filename, tt.group, strings.Join(tt.content, "\n"))
			be.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveCtrls(t *testing.T) {
	t.Parallel()
	p := []byte("a\x1b[1;cabc")
	r := readme.RemoveCtrls(p)
	be.Equal(t, []byte("aabc"), r)
}

func TestIncompatibleANSI(t *testing.T) {
	t.Parallel()
	b, err := readme.MatchANSI(nil)
	be.Err(t, err, nil)
	be.True(t, !b)
	r := strings.NewReader("a\x1b[1;cabc")
	b, err = readme.MatchANSI(r)
	be.Err(t, err, nil)
	be.True(t, b)
	r = strings.NewReader("a\x1b[Acabc")
	b, err = readme.MatchANSI(r)
	be.Err(t, err, nil)
	be.True(t, b)
}
