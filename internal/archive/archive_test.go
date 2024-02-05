package archive_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Defacto2/server/internal/archive"
	"github.com/Defacto2/server/internal/helper"
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
	assert.Error(t, err)
	assert.Empty(t, files)
	assert.Empty(t, name)

	files, name, err = archive.Content(td(""), "")
	assert.Error(t, err)
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
	assert.Error(t, err)
	assert.Empty(t, files)
	assert.Empty(t, name)

	// test a legacy shrunk archive
	files, name, err = archive.Content(td("PKZ80A1.ZIP"), "PKZ80A1.ZIP")
	assert.Nil(t, err)
	assert.Len(t, files, 15)
	assert.Equal(t, "PKZ80A1.ZIP", name)

	// test an unsupported 7z file
	files, name, err = archive.Content(td("TEST.7z"), "TEST.7z")
	assert.Error(t, err)
	assert.Empty(t, files)
	assert.Empty(t, name)

	// test a xz archive
	files, name, err = archive.Content(td("TEST.tar.xz"), "TEST.tar.xz")
	assert.Nil(t, err)
	assert.Len(t, files, 15)
	assert.Equal(t, "TEST.tar.xz", name)

	// test an unsupported lha archive
	files, name, err = archive.Content(td("LHA114.LZH"), "LHA114.LZH")
	assert.Error(t, err)
	assert.Empty(t, files)
	assert.Empty(t, name)

	// test non-latin text
	files, name, err = archive.Content(td("τεχτƒιℓε.zip"), "τεχτƒιℓε.zip")
	assert.Nil(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, "τεχτƒιℓε.zip", name)
}

func TestExtractAll(t *testing.T) {
	t.Parallel()
	err := archive.ExtractAll("", "", "")
	assert.Error(t, err)

	err = archive.ExtractAll(td(""), "", "")
	assert.Error(t, err)

	err = archive.ExtractAll(td(""), os.TempDir(), "")
	assert.Error(t, err)

	err = archive.ExtractAll(td("PKZ204EX.ZIP"), os.TempDir(), "")
	assert.Error(t, err)

	err = archive.ExtractAll(td("PKZ204EX.ZIP"), os.TempDir(), "test.exe")
	assert.Error(t, err)

	tmp, err := os.MkdirTemp("", "testextractall-")
	assert.NoError(t, err)

	err = archive.ExtractAll(td("PKZ204EX.ZIP"), tmp, "PKZ204EX.ZIP")
	assert.NoError(t, err)

	defer os.RemoveAll(tmp)
}

func TestExtract(t *testing.T) {
	t.Parallel()
	err := archive.Extract("", "", "", "")
	assert.Error(t, err)

	err = archive.Extract(td(""), "", "", "")
	assert.Error(t, err)

	// Extract(src, dst, filename, target string) error {
	err = archive.Extract(td(""), os.TempDir(), "", "")
	assert.Error(t, err)

	err = archive.Extract(td("PKZ204EX.ZIP"), os.TempDir(), "", "")
	assert.Error(t, err)

	err = archive.Extract(td("PKZ204EX.ZIP"), os.TempDir(), "", "test.exe")
	assert.Error(t, err)

	tmp, err := os.MkdirTemp("", "test-")
	assert.NoError(t, err)

	err = archive.Extract(td("PKZ204EX.ZIP"), tmp, "PKZ204EX.ZIP", "")
	assert.NoError(t, err)

	err = archive.Extract(td("PKZ204EX.ZIP"), tmp, "PKZ204EX.ZIP", "test.me")
	assert.NoError(t, err)
	st, err := os.Stat(filepath.Join(tmp, "TEST.ME"))
	assert.Error(t, err)
	assert.Nil(t, st)

	tmp1, err := os.MkdirTemp("", "test-")
	assert.NoError(t, err)

	err = archive.Extract(td("PKZ204EX.ZIP"), tmp1, "PKZ204EX.ZIP", "TEST.ME")
	assert.NoError(t, err)
	st, err = os.Stat(filepath.Join(tmp1, "TEST.ME"))
	assert.NoError(t, err)
	assert.Greater(t, st.Size(), int64(0))

	defer os.RemoveAll(tmp)
	defer os.RemoveAll(tmp1)
}

func TestARJ(t *testing.T) {
	t.Parallel()
	const name = "ARJ310.ARJ"
	x := archive.Extractor{}
	err := x.ARJ()
	assert.Error(t, err)

	x = archive.Extractor{
		Source: td(name),
	}
	err = x.ARJ()
	assert.Error(t, err)

	tmp1, err := os.MkdirTemp("", "testarj-")
	assert.NoError(t, err)
	defer os.RemoveAll(tmp1)

	x = archive.Extractor{
		Source:      td(name),
		Destination: tmp1,
	}
	err = x.ARJ()
	assert.NoError(t, err)
	st, err := os.Stat(filepath.Join(tmp1, "TEST.ME"))
	assert.NoError(t, err)
	assert.Greater(t, st.Size(), int64(0))

	count, err := helper.Count(tmp1)
	assert.NoError(t, err)
	assert.Equal(t, 15, count)

	tmp2, err := os.MkdirTemp("", "testarj3files-")
	assert.NoError(t, err)
	defer os.RemoveAll(tmp2)
	x = archive.Extractor{
		Source:      td(name),
		Destination: tmp2,
	}
	err = x.ARJ("TEST.ME", "TEST.TXT", "TEST.EXE")
	assert.NoError(t, err)

	count, err = helper.Count(tmp2)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestLHA(t *testing.T) {
	t.Parallel()
	const name = "LHA114.LZH"

	x := archive.Extractor{}
	err := x.LHA()
	assert.Error(t, err)

	x = archive.Extractor{
		Source: td(name),
	}
	err = x.LHA()
	assert.Error(t, err)

	tmp1, err := os.MkdirTemp("", "testlzh-")
	assert.NoError(t, err)
	defer os.RemoveAll(tmp1)

	x = archive.Extractor{
		Source:      td(name),
		Destination: tmp1,
	}
	err = x.LHA()
	assert.NoError(t, err)
	st, err := os.Stat(filepath.Join(tmp1, "TEST.ME"))
	assert.NoError(t, err)
	assert.Greater(t, st.Size(), int64(0))
	count, err := helper.Count(tmp1)
	assert.NoError(t, err)
	assert.Equal(t, 15, count)

	tmp2, err := os.MkdirTemp("", "testlzhfiles-")
	assert.NoError(t, err)
	defer os.RemoveAll(tmp2)
	x = archive.Extractor{
		Source:      td(name),
		Destination: tmp2,
	}
	err = x.LHA("TEST.ME", "TEST.TXT", "TEST.EXE")
	assert.NoError(t, err)

	count, err = helper.Count(tmp2)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestZip(t *testing.T) {
	t.Parallel()
	const name = "PKZ80A4.ZIP"

	x := archive.Extractor{}
	err := x.Zip()
	assert.Error(t, err)

	x = archive.Extractor{
		Source: td(name),
	}
	err = x.Zip()
	assert.Error(t, err)

	tmp1, err := os.MkdirTemp("", "testtar-")
	assert.NoError(t, err)
	defer os.RemoveAll(tmp1)

	x = archive.Extractor{
		Source:      td(name),
		Destination: tmp1,
	}
	err = x.Zip()
	assert.NoError(t, err)
	st, err := os.Stat(filepath.Join(tmp1, "TEST.ME"))
	assert.NoError(t, err)
	assert.Greater(t, st.Size(), int64(0))
	count, err := helper.Count(tmp1)
	assert.NoError(t, err)
	assert.Equal(t, 15, count)

	tmp2, err := os.MkdirTemp("", "testtarfiles-")
	assert.NoError(t, err)
	defer os.RemoveAll(tmp2)
	x = archive.Extractor{
		Source:      td(name),
		Destination: tmp2,
	}
	err = x.Zip("TEST.ME", "TEST.TXT", "TEST.EXE")
	assert.NoError(t, err)

	count, err = helper.Count(tmp2)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestBodyARJ(t *testing.T) {
	t.Parallel()
	const name = "ARJ310.ARJ"
	x := archive.Contents{}
	err := x.ARJ("")
	assert.Error(t, err)

	err = x.ARJ(name)
	assert.Error(t, err)

	err = x.ARJ(td(name))
	assert.NoError(t, err)
	assert.Len(t, x.Files, 15)
}

func TestBodyLHA(t *testing.T) {
	t.Parallel()
	const name = "LHA114.LZH"
	x := archive.Contents{}
	err := x.LHA("")
	assert.Error(t, err)

	err = x.LHA(name)
	assert.Error(t, err)

	err = x.LHA(td(name))
	assert.NoError(t, err)
	assert.Len(t, x.Files, 15)
}

func TestBodyRar(t *testing.T) {
	t.Parallel()
	const name = "RAR624.RAR"
	x := archive.Contents{}
	err := x.Rar("")
	assert.Error(t, err)

	err = x.Rar(name)
	assert.Error(t, err)

	err = x.Rar(td(name))
	assert.NoError(t, err)
	assert.Len(t, x.Files, 15)
}

func TestBodyZip(t *testing.T) {
	t.Parallel()
	const name = "PKZ80A4.ZIP"
	x := archive.Contents{}
	err := x.Zip("")
	assert.Error(t, err)

	err = x.Zip(name)
	assert.Error(t, err)

	err = x.Zip(td(name))
	assert.NoError(t, err)
	assert.Len(t, x.Files, 15)
}

func TestMagicExt(t *testing.T) {
	t.Parallel()
	magic, err := archive.MagicExt("")
	assert.Error(t, err)
	assert.Empty(t, magic)

	name := "PKZ80A4.ZIP"
	magic, err = archive.MagicExt(td(name))
	assert.NoError(t, err)
	assert.Equal(t, ".zip", magic)

	// name := "LHA114.LZH"
	// magic, err := archive.MagicExt(td(name))
	// assert.NoError(t, err)
	// assert.Equal(t, ".lha", magic)
}
