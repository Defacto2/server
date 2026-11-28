package ext_test

import (
	"testing"

	"github.com/Defacto2/server/handler/html3/ext"
	"github.com/nalgeon/be"
)

func TestIconName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		label  string
		name   string
		expect string
	}{
		{"empty", "", ""},
		{"no extension", "myfile", ""},
		{"dot only", ".", ""},
		{"unknown ext", "file.unknown", ""},
		{"image png", "myimage.png", ext.Pic},
		{"uppercase image", "MYIMAGE.JPG", ext.Pic},
		{"video double ext", "double.ext.avi", ext.Vid},
		{"html space", "a web site .htm", ext.Htm},
		{"audio unicode", "👾.mp3", ext.Sfx},
		{"tune mod", "document.mod", ext.Sfx},
		{"archive rar", "archive.rar", ext.Zip},
		{"archive zip", "archive.zip", ext.Zip},
		{"archive tar.gz", "archive.tar.gz", ext.Zip},
		{"archive tgz", "archive.tgz", ext.Zip},
		{"app exe", "program.exe", ext.App},
		{"uppercase app", "PROGRAM.COM", ext.App},
		{"doc txt", "document.txt", ext.Doc},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			t.Parallel()

			be.Equal(t, ext.IconName(tt.name), tt.expect)
		})
	}
}

func TestIsPredicates(t *testing.T) {
	t.Parallel()

	be.True(t, ext.IsApp("run.exe"))
	be.True(t, ext.IsApp("RUN.COM"))
	be.True(t, !ext.IsApp("readme.txt"))

	be.True(t, ext.IsArchive("pack.zip"))
	be.True(t, ext.IsArchive("PACK.7Z"))
	be.True(t, !ext.IsArchive("pack.txt"))

	be.True(t, ext.IsDocument("release.nfo"))
	be.True(t, ext.IsDocument("FILE_ID.DIZ"))
	be.True(t, !ext.IsDocument("demo.exe"))

	be.True(t, ext.IsImage("art.pcx"))
	be.True(t, ext.IsImage("PIC.PNG"))
	be.True(t, !ext.IsImage("song.mp3"))

	be.True(t, ext.IsHTML("index.html"))
	be.True(t, ext.IsHTML("INDEX.HTM"))
	be.True(t, !ext.IsHTML("index.txt"))

	be.True(t, ext.IsAudio("track.mp3"))
	be.True(t, ext.IsAudio("TRACK.FLAC"))
	be.True(t, ext.IsAudio("track.m4a"))
	be.True(t, !ext.IsAudio("track.mod"))

	be.True(t, ext.IsTune("song.xm"))
	be.True(t, ext.IsTune("SONG.S3M"))
	be.True(t, !ext.IsTune("song.mp3"))

	be.True(t, ext.IsVideo("clip.mp4"))
	be.True(t, ext.IsVideo("CLIP.AVI"))
	be.True(t, ext.IsVideo("clip.qt"))
	be.True(t, !ext.IsVideo("track.m4a"))
	be.True(t, !ext.IsVideo("clip.png"))

	be.True(t, ext.IsExt("CUSTOM.FOO", ".foo", ".bar"))
	be.True(t, !ext.IsExt("noext", ".foo", ".bar"))
	be.True(t, !ext.IsExt("", ".foo"))
}
