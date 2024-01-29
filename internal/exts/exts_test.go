package exts_test

import (
	"testing"

	"github.com/Defacto2/server/internal/exts"
	"github.com/stretchr/testify/assert"
)

func TestIcon(t *testing.T) {
	tests := []struct {
		name      string
		expect    string
		assertion assert.ComparisonAssertionFunc
	}{
		{"", "", assert.Equal},
		{"myfile", "", assert.Equal},
		{"myimage.png", "image2", assert.Equal},
		{"double.exts.avi", "movie", assert.Equal},
		{"a web site .htm", "generic", assert.Equal},
		{"👾.mp3", "sound2", assert.Equal},
	}
	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			tt.assertion(t, tt.expect, exts.IconName(tt.name))
		})
	}
}

func TestIconName(t *testing.T) {
	s := exts.IconName("myfile")
	assert.Equal(t, "", s)
	s = exts.IconName("myimage.png")
	assert.Equal(t, exts.Pic, s)
	s = exts.IconName("double.exts.avi")
	assert.Equal(t, exts.Vid, s)
	s = exts.IconName("a web site .htm")
	assert.Equal(t, exts.Htm, s)
	s = exts.IconName("👾.mp3")
	assert.Equal(t, exts.Sfx, s)
	s = exts.IconName("archive.rar")
	assert.Equal(t, exts.Zip, s)
	s = exts.IconName("archive.zip")
	assert.Equal(t, exts.Zip, s)
	s = exts.IconName("program.exe")
	assert.Equal(t, exts.App, s)
	s = exts.IconName("document.txt")
	assert.Equal(t, exts.Doc, s)
	s = exts.IconName("document.mod")
	assert.Equal(t, exts.Sfx, s)
}
