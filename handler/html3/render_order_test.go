package html3_test

import (
	"testing"

	"github.com/Defacto2/server/handler/html3"
	"github.com/stretchr/testify/assert"
)

func TestClauses(t *testing.T) {
	tests := []string{
		html3.NameAsc,
		html3.NameDes,
		html3.PublAsc,
		html3.PublDes,
		html3.PostAsc,
		html3.PostDes,
		html3.SizeAsc,
		html3.SizeDes,
		html3.DescAsc,
		html3.DescDes,
	}
	for i, s := range tests {
		assert.Equal(t, int(html3.Clauses(s)), i)
	}
	assert.Equal(t, int(html3.Clauses("")),
		int(html3.Clauses(html3.NameAsc)), "default should be name asc")
}

func TestSorter(t *testing.T) {
	tests := []string{
		html3.NameAsc,
		html3.NameDes,
		html3.PublAsc,
		html3.PublDes,
		html3.PostAsc,
		html3.PostDes,
		html3.SizeAsc,
		html3.SizeDes,
		html3.DescAsc,
		html3.DescDes,
	}
	for _, s := range tests {
		switch s {
		case html3.NameAsc:
			assert.Equal(t, html3.Sorter(s)[string(html3.Name)], "D")
		case html3.NameDes:
			assert.Equal(t, html3.Sorter(s)[string(html3.Name)], "A")
		case html3.PublAsc:
			assert.Equal(t, html3.Sorter(s)[string(html3.Publish)], "D")
		case html3.PublDes:
			assert.Equal(t, html3.Sorter(s)[string(html3.Publish)], "A")
		case html3.PostAsc:
			assert.Equal(t, html3.Sorter(s)[string(html3.Posted)], "D")
		case html3.PostDes:
			assert.Equal(t, html3.Sorter(s)[string(html3.Posted)], "A")
		case html3.SizeAsc:
			assert.Equal(t, html3.Sorter(s)[string(html3.Size)], "D")
		case html3.SizeDes:
			assert.Equal(t, html3.Sorter(s)[string(html3.Size)], "A")
		case html3.DescAsc:
			assert.Equal(t, html3.Sorter(s)[string(html3.Desc)], "D")
		case html3.DescDes:
			assert.Equal(t, html3.Sorter(s)[string(html3.Desc)], "A")
		}

	}
}
