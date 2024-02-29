package archive_test

// It is highly recommended to run these tests with -race flag to detect
// race conditions.
//
// go test -timeout 30s -count 5 -race github.com/Defacto2/server/internal/archive

// TODO: TEST ALL assets for extract all files using an array of file names.

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

func filenames() []string {
	return []string{
		"ARJ310.ARJ",
		"LHA114.LZH",
		"PKZ80A1.ZIP",
		"PKZ80A2.ZIP",
		"PKZ80A3.ZIP",
		"PKZ80A4.ZIP",
		"PKZ80B1.ZIP",
		"PKZ80B2.ZIP",
		"PKZ80B3.ZIP",
		"PKZ80B4.ZIP",
		"PKZ90A1.ZIP",
		"PKZ90A2.ZIP",
		"PKZ90A3.ZIP",
		"PKZ90A4.ZIP",
		"PKZ90B1.ZIP",
		"PKZ90B2.ZIP",
		"PKZ90B3.ZIP",
		"PKZ90B4.ZIP",
		"PKZ110.ZIP",
		"PKZ110EI.ZIP",
		"PKZ110ES.ZIP",
		"PKZ110EX.ZIP",
		"PKZ204E0.ZIP",
		"PKZ204EF.ZIP",
		"PKZ204EN.ZIP",
		"PKZ204ES.ZIP",
		"PKZ204EX.ZIP",
		"RAR624.RAR",
		"TAR135.TAR",
		"TEST.tar.xz",
	}
}

func TestContent(t *testing.T) {
	t.Parallel()

	const wantFiles = 15

	files, err := archive.List("", "")
	assert.Error(t, err)
	assert.Empty(t, files)

	files, err = archive.List(td(""), "")
	assert.Error(t, err)
	assert.Empty(t, files)

	// test the unsupported file types
	files, err = archive.List(td("ARC521P.ARC"), "ARC521P.ARC")
	assert.Error(t, err)
	assert.Empty(t, files)
	files, err = archive.List(td("TEST.7z"), "TEST.7z")
	assert.Error(t, err)
	assert.Empty(t, files)

	// test the tar.gz handler
	finename := "TAR135.TAR.GZ"
	files, err = archive.List(td("TAR135.GZ"), finename)
	assert.NoError(t, err)
	assert.Len(t, files, wantFiles)

	// test unicode filename
	files, err = archive.List(td("τεχτƒιℓε.zip"), "τεχτƒιℓε.zip")
	assert.NoError(t, err)
	assert.Len(t, files, 1)

	// test all the supported files
	for _, name := range filenames() {
		files, err = archive.List(td(name), name)
		assert.NoError(t, err)
		assert.Len(t, files, wantFiles)
	}
}

func TestExtractAll(t *testing.T) {
	t.Parallel()
	err := archive.Extract("", "", "")
	assert.Error(t, err)

	err = archive.Extract(td(""), "", "")
	assert.Error(t, err)

	err = archive.Extract(td(""), os.TempDir(), "")
	assert.Error(t, err)

	err = archive.Extract(td("PKZ204EX.ZIP"), os.TempDir(), "")
	assert.Error(t, err)

	err = archive.Extract(td("PKZ204EX.ZIP"), os.TempDir(), "test.exe")
	assert.Error(t, err)

	tmp, err := os.MkdirTemp("", "testextractall-")
	assert.NoError(t, err)
	defer os.RemoveAll(tmp)

	err = archive.Extract(td("PKZ204EX.ZIP"), tmp, "PKZ204EX.ZIP")
	assert.NoError(t, err)

	tmp1, err := os.MkdirTemp("", "testextractall1-")
	assert.NoError(t, err)
	defer os.RemoveAll(tmp1)

	name := "ARJ310.ARJ"
	err = archive.Extract(td(name), tmp1, name)
	assert.NoError(t, err)
	count, err := helper.Count(tmp1)
	assert.NoError(t, err)
	assert.Equal(t, 15, count)
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
	x := archive.Content{}
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
	x := archive.Content{}
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
	x := archive.Content{}
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
	x := archive.Content{}
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

	name = "LHA114.LZH"
	magic, err = archive.MagicExt(td(name))
	assert.NoError(t, err)
	assert.Equal(t, ".lha", magic)
}
