package ext_test

import (
	"testing"

	"github.com/Defacto2/server/internal/ext"
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
		{"double.ext.avi", "movie", assert.Equal},
		{"a web site .htm", "generic", assert.Equal},
		{"👾.mp3", "sound2", assert.Equal},
	}
	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			tt.assertion(t, tt.expect, ext.IconName(tt.name))
		})
	}
}

func TestIconName(t *testing.T) {
	s := ext.IconName("myfile")
	assert.Equal(t, "", s)
	s = ext.IconName("myimage.png")
	assert.Equal(t, ext.Pic, s)
	s = ext.IconName("double.ext.avi")
	assert.Equal(t, ext.Vid, s)
	s = ext.IconName("a web site .htm")
	assert.Equal(t, ext.Htm, s)
	s = ext.IconName("👾.mp3")
	assert.Equal(t, ext.Sfx, s)
	s = ext.IconName("archive.rar")
	assert.Equal(t, ext.Zip, s)
	s = ext.IconName("archive.zip")
	assert.Equal(t, ext.Zip, s)
	s = ext.IconName("program.exe")
	assert.Equal(t, ext.App, s)
	s = ext.IconName("document.txt")
	assert.Equal(t, ext.Doc, s)
	s = ext.IconName("document.mod")
	assert.Equal(t, ext.Sfx, s)
}
