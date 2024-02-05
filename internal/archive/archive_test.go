package archive_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Defacto2/server/internal/archive"
	"github.com/stretchr/testify/assert"
)

func td(name string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("runtime.Caller failed")
	}
	d := filepath.Join(filepath.Dir(file), "../..")
	x := filepath.Join(d, "assets", "testdata", name)
	return x

}

func TestContent(t *testing.T) {
	t.Parallel()
	files, name, err := archive.Content("", "")
	assert.NotNil(t, err)
	assert.Empty(t, files)
	assert.Empty(t, name)

	files, name, err = archive.Content(td(""), "")
	assert.NotNil(t, err)
	assert.Empty(t, files)
	assert.Empty(t, name)

	// test a deflated zip file
	files, name, err = archive.Content(td("PKZ204EX.ZIP"), "PKZ204EX.ZIP")
	assert.Nil(t, err)
	assert.Len(t, files, 15)
	assert.Equal(t, "PKZ204EX.ZIP", name)

	// test the tar handler
	files, name, err = archive.Content(td("TAR135.TAR"), "TAR135.TAR")
	assert.Nil(t, err)
	assert.Len(t, files, 15)
	assert.Equal(t, "TAR135.TAR", name)

	// test the rar handler
	files, name, err = archive.Content(td("RAR624.RAR"), "RAR624.RAR")
	assert.Nil(t, err)
	assert.Len(t, files, 15)
	assert.Equal(t, "RAR624.RAR", name)

	// test the tar.gz handler
	files, name, err = archive.Content(td("TAR135.GZ"), "TAR135.TAR.GZ")
	assert.Nil(t, err)
	assert.Len(t, files, 15)
	assert.Equal(t, "TAR135.TAR.GZ", name)

	// test an arj file
	files, name, err = archive.Content(td("ARJ310.ARJ"), "ARJ310.ARJ")
	assert.Nil(t, err)
	assert.Len(t, files, 15)
	assert.Equal(t, "ARJ310.ARJ", name)

	// test an unsupported arc file
	files, name, err = archive.Content(td("ARC521P.ARC"), "ARC521P.ARC")
	assert.NotNil(t, err)
	assert.Empty(t, files)
	assert.Empty(t, name)

	// test a legacy shrunk archive
	files, name, err = archive.Content(td("PKZ80A1.ZIP"), "PKZ80A1.ZIP")
	assert.Nil(t, err)
	assert.Len(t, files, 15)
	assert.Equal(t, "PKZ80A1.ZIP", name)

	// test an unsupported 7z file
	files, name, err = archive.Content(td("TEST.7z"), "TEST.7z")
	assert.NotNil(t, err)
	assert.Empty(t, files)
	assert.Empty(t, name)

	// test a xz archive
	files, name, err = archive.Content(td("TEST.tar.xz"), "TEST.tar.xz")
	assert.Nil(t, err)
	assert.Len(t, files, 15)
	assert.Equal(t, "TEST.tar.xz", name)
}
