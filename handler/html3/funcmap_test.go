package html3_test

import (
	"math"
	"testing"

	"github.com/Defacto2/server/handler/html3"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
	"go.uber.org/zap"
)

func z() *zap.SugaredLogger {
	return zap.NewExample().Sugar()
}
func TestFile_Description(t *testing.T) {
	type fields struct {
		Title    string
		GroupBy  string
		Section  string
		Platform string
	}
	const (
		x = "Hello world"
		g = "Defacto2"
		s = "intro"
		p = "dos"
		m = "magazine"
	)
	tests := []struct {
		name      string
		fields    fields
		expect    string
		assertion assert.ComparisonAssertionFunc
	}{
		{"empty", fields{}, "", assert.Exactly},
		{"only title", fields{Title: x}, "", assert.Exactly},
		{"req group", fields{Title: x, Platform: p}, "", assert.Exactly},
		{"default", fields{x, g, "", ""}, "Hello world from Defacto2.", assert.Exactly},
		{"with platform", fields{x, g, "", p}, "Hello world from Defacto2 for Dos.", assert.Exactly},
		{"no title", fields{"", g, "", p}, "A release from Defacto2 for Dos.", assert.Exactly},
		{"magazine", fields{"1", g, m, p}, "Defacto2 issue 1 for Dos.", assert.Exactly},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := html3.File{
				Title:    tt.fields.Title,
				GroupBy:  tt.fields.GroupBy,
				Section:  tt.fields.Section,
				Platform: tt.fields.Platform,
			}
			t.Run(tt.name, func(t *testing.T) {
				tt.assertion(t, tt.expect, f.Description())
			})
		})
	}
}

func TestDescription(t *testing.T) {
	empty := null.String{}
	s := html3.Description(empty, empty, empty, empty)
	assert.Empty(t, s)
	s = html3.Description(
		null.StringFrom("intro"),
		null.StringFrom("dos"),
		null.StringFrom("Defacto2"),
		null.StringFrom("Hello world"))
	assert.Equal(t, "Hello world from Defacto2 for Dos.", s)
}

func TestFileHref(t *testing.T) {
	s := html3.FileHref(nil, 0)
	assert.Equal(t, "zap logger is nil", s)
	s = html3.FileHref(z(), 0)
	assert.Equal(t, "/html3/d/0", s)
}

func TestFileLinkPad(t *testing.T) {
	n := null.String{}
	s := html3.FileLinkPad(0, n)
	assert.Equal(t, "", s)
	s = html3.FileLinkPad(20, null.StringFrom("file"))
	assert.Equal(t, "                ", s)
}

func TestFormattings(t *testing.T) {
	assert.Equal(t, html3.File{Filename: ""}.FileLinkPad(0), "", "empty")
	assert.Equal(t, html3.File{Filename: ""}.FileLinkPad(4), "    ", "4 pads")
	assert.Equal(t, html3.File{Filename: "file"}.FileLinkPad(6), "  ", "2 pads")
	assert.Equal(t, html3.File{Filename: "file.txt"}.FileLinkPad(6), "", "too big")
	assert.Equal(t, html3.File{Size: 0}.LeadFS(0), "0B", "0 size")
	assert.Equal(t, html3.File{Size: 1}.LeadFS(3), " 1B", "1 size")
	assert.Equal(t, html3.File{Size: 1000}.LeadFS(0), "1000B", "1000 size")
	assert.Equal(t, html3.File{Size: 1024}.LeadFS(0), "1k", "1k size")
	assert.Equal(t, html3.File{Size: int64(math.Pow(1024, 2))}.LeadFS(0), "1M", "1MB size")
	assert.Equal(t, html3.File{Size: int64(math.Pow(1024, 3))}.LeadFS(0), "1G", "1GB size")
	assert.Equal(t, html3.File{Size: int64(math.Pow(1024, 4))}.LeadFS(0), "1T", "1TB size")
	assert.Equal(t, html3.LeadInt(0, 1), "1")
	assert.Equal(t, html3.LeadInt(1, 1), "1")
	assert.Equal(t, html3.LeadInt(3, 1), "  1")
	assert.True(t, html3.File{Platform: "java"}.IsOS())
	assert.Equal(t, html3.File{Platform: "java"}.OS(), " for Java")
}
