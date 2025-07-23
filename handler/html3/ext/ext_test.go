package ext_test

import (
	"testing"

	"github.com/Defacto2/server/handler/html3/ext"
	"github.com/nalgeon/be"
)

func TestIcon(t *testing.T) {
	tests := []struct {
		name   string
		expect string
	}{
		{"", ""},
		{"myfile", ""},
		{"myimage.png", "image2"},
		{"double.ext.avi", "movie"},
		{"a web site .htm", "generic"},
		{"👾.mp3", "sound2"},
	}
	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			be.Equal(t, ext.IconName(tt.name), tt.expect)
		})
	}
}

func TestIconName(t *testing.T) {
	s := ext.IconName("myfile")
	be.Equal(t, s, "")
	s = ext.IconName("myimage.png")
	be.Equal(t, ext.Pic, s)
	s = ext.IconName("double.ext.avi")
	be.Equal(t, ext.Vid, s)
	s = ext.IconName("a web site .htm")
	be.Equal(t, ext.Htm, s)
	s = ext.IconName("👾.mp3")
	be.Equal(t, ext.Sfx, s)
	s = ext.IconName("archive.rar")
	be.Equal(t, ext.Zip, s)
	s = ext.IconName("archive.zip")
	be.Equal(t, ext.Zip, s)
	s = ext.IconName("program.exe")
	be.Equal(t, ext.App, s)
	s = ext.IconName("document.txt")
	be.Equal(t, ext.Doc, s)
	s = ext.IconName("document.mod")
	be.Equal(t, ext.Sfx, s)
}
