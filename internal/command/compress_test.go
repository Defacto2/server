package command_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func tduncompress(name string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("runtime.Caller failed")
	}
	d := filepath.Join(filepath.Dir(file), "../..")
	x := filepath.Join(d, "assets", "testdata", "uncompress", name)
	return x
}

func Test_ExtractOne(t *testing.T) {
	t.Parallel()
	err := command.ExtractFile(nil, "", "", "", "")
	require.Error(t, err)

	src, dst, hint, name := "", "", "", ""
	err = command.ExtractFile(logr(), src, dst, hint, name)
	require.Error(t, err)

	src = td("PKZ80A1.TXT")
	err = command.ExtractFile(logr(), src, "", "", "")
	require.Error(t, err)

	src = td("PKZ80A1.ZIP")
	err = command.ExtractFile(logr(), src, "", ".zip", "")
	require.Error(t, err)

	err = command.ExtractFile(logr(), src, "", ".zip", "TEST.ASC")
	require.Error(t, err)

	dst = helper.TmpDir()
	err = command.ExtractFile(logr(), src, dst, ".zip", "TEST.ASC")
	require.Error(t, err)

	src = td("PKZ204EX.ZIP")
	dst, _ = filepath.Abs(filepath.Join(helper.TmpDir(), "PKZ204EX.ZIP"))
	err = command.ExtractFile(logr(), src, dst, ".zip", "TEST.ASC")
	require.NoError(t, err)

	ok := helper.File(dst)
	assert.True(t, ok)
	defer os.Remove(dst)

	src = td("PKZ80A1.ZIP")
	dst, _ = filepath.Abs(filepath.Join(helper.TmpDir(), "PKZ80A1.ZIP"))
	err = command.ExtractFile(logr(), src, dst, ".zip", "TEST.ASC")
	require.NoError(t, err)

	ok = helper.File(dst)
	assert.True(t, ok)
	defer os.Remove(dst)

	src = td("ARC521P.ARC")
	dst, _ = filepath.Abs(filepath.Join(helper.TmpDir(), "ARC521P.ARC"))
	err = command.ExtractFile(logr(), src, dst, ".arc", "TEST.JPG")
	require.NoError(t, err)

	ok = helper.File(dst)
	assert.True(t, ok)
	defer os.Remove(dst)

	src = td("ARJ310.ARJ")
	dst, _ = filepath.Abs(filepath.Join(helper.TmpDir(), "ARJ310.ARJ"))
	err = command.ExtractFile(logr(), src, dst, ".arj", "TEST.JPEG")
	require.NoError(t, err)

	ok = helper.File(dst)
	assert.True(t, ok)
	defer os.Remove(dst)

	src = td("RAR624.RAR")
	dst, _ = filepath.Abs(filepath.Join(helper.TmpDir(), "RAR624.RAR"))
	err = command.ExtractFile(logr(), src, dst, ".rar", "TEST.JPG")
	require.NoError(t, err)

	ok = helper.File(dst)
	assert.True(t, ok)
	defer os.Remove(dst)

	src = td("TAR135.TAR")
	dst, _ = filepath.Abs(filepath.Join(helper.TmpDir(), "TAR135.TAR"))
	err = command.ExtractFile(logr(), src, dst, ".tar", "TEST.JPG")
	require.NoError(t, err)

	ok = helper.File(dst)
	assert.True(t, ok)
	defer os.Remove(dst)

	src = td("TAR135.GZ")
	dst, _ = filepath.Abs(filepath.Join(helper.TmpDir(), "TAR135.GZ"))
	err = command.ExtractFile(logr(), src, dst, ".gz", "TEST.JPG")
	require.NoError(t, err)

	ok = helper.File(dst)
	assert.True(t, ok)
	defer os.Remove(dst)
}

func Test_ArjExitStatus(t *testing.T) {
	t.Parallel()
	s := command.ArjExitStatus(nil)
	assert.Equal(t, "", s)
	const name = command.Arj
	cmd := exec.Command(name, "throwawayarg")
	err := cmd.Run()
	require.Error(t, err)
	s = command.ArjExitStatus(err)
	assert.Equal(t, "user error, bad command line parameters", s)
}

func Test_UnRarExitStatus(t *testing.T) {
	t.Parallel()
	s := command.UnRarExitStatus(nil)
	assert.Equal(t, "", s)
	const name = command.Unrar
	cmd := exec.Command(name, "throwawayarg")
	err := cmd.Run()
	require.Error(t, err)
	s = command.UnRarExitStatus(err)
	assert.Equal(t, "wrong command line option", s)
}

func Test_ExtractAnsiLove(t *testing.T) {
	t.Parallel()
	dir := command.Dirs{}

	err := dir.ExtractAnsiLove(nil, "", "", "", "")
	require.Error(t, err)

	err = dir.ExtractAnsiLove(logr(), "", "", "", "")
	require.Error(t, err)

	src := td("PKZ204EX.ZIP")
	err = dir.ExtractAnsiLove(logr(),
		src, ".zip", "000000ABCDE", "TEST.ANS")
	require.NoError(t, err)

	ok := helper.File("000000ABCDE.webp")
	assert.True(t, ok)
	err = os.Remove("000000ABCDE.webp")
	require.NoError(t, err)

	err = dir.ExtractAnsiLove(logr(),
		src, ".zip", "000000ABCDE", "nosuchfile")
	require.Error(t, err)
}

func Test_ExtractImage(t *testing.T) {
	t.Parallel()
	prev, err := os.MkdirTemp(helper.TmpDir(), "preview")
	require.NoError(t, err)
	thumb, err := os.MkdirTemp(helper.TmpDir(), "thumb")
	require.NoError(t, err)
	dl, err := os.MkdirTemp(helper.TmpDir(), "download")
	require.NoError(t, err)
	dir := command.Dirs{
		Download:  dl,    // this prefixes to UUID
		Preview:   prev,  // this is the output dest
		Thumbnail: thumb, // this is the cropped output dest
	}
	// intentional errors
	err = dir.ExtractImage(nil, "", "", "", "")
	require.Error(t, err)
	err = dir.ExtractImage(logr(), "", "", "", "")
	require.Error(t, err)

	broken := []string{"nosuchfile", "TEST.ASC", "TEST.JPEG", "TEST.ANS"}
	for _, name := range broken {
		src := td("PKZ204EX.ZIP")
		err = dir.ExtractImage(logr(), src, ".zip", "000000ABCDE", name)
		require.Error(t, err)
	}
	// cases that create webp files
	names := []string{"TEST.JPG", "TEST.GIF", "TEST.BMP"}
	op := filepath.Join(prev, "000000ABCDE.webp")
	ot := filepath.Join(thumb, "000000ABCDE.webp")
	for _, name := range names {
		src := td("PKZ204EX.ZIP")
		err = dir.ExtractImage(logr(), src, ".zip", "000000ABCDE", name)
		require.NoError(t, err)

		ok := helper.File(op)
		assert.True(t, ok)
		ok = helper.File(ot)
		assert.True(t, ok)
		err = os.Remove(op)
		require.NoError(t, err)
		err = os.Remove(ot)
		require.NoError(t, err)
	}
	// unique case that creates compressed PNG files
	{
		op = filepath.Join(prev, "000000ABCDE.png")
		ot = filepath.Join(thumb, "000000ABCDE.webp")
		name := "TEST.PNG"
		src := td("PKZ204EX.ZIP")
		err = dir.ExtractImage(logr(), src, ".zip", "000000ABCDE", name)
		require.NoError(t, err)
		ok := helper.File(op)
		assert.True(t, ok)
		ok = helper.File(ot)
		assert.True(t, ok)
		err = os.Remove(op)
		require.NoError(t, err)
		err = os.Remove(ot)
		require.NoError(t, err)
	}
}

func Test_PreviewPixels(t *testing.T) {
	t.Parallel()
	prev, err := os.MkdirTemp(helper.TmpDir(), "preview")
	require.NoError(t, err)
	thumb, err := os.MkdirTemp(helper.TmpDir(), "thumb")
	require.NoError(t, err)
	dl, err := os.MkdirTemp(helper.TmpDir(), "download")
	require.NoError(t, err)
	dir := command.Dirs{
		Download:  dl,    // this prefixes to UUID
		Preview:   prev,  // this is the output dest
		Thumbnail: thumb, // this is the cropped output dest
	}
	imgs := []string{"TEST.BMP", "TEST.GIF", "TEST.JPG", "TEST.PCX", "TEST.PNG"}
	for _, name := range imgs {
		fp := tduncompress(name)
		err = dir.PreviewPixels(logr(), fp, "000000ABCDE")
		require.NoError(t, err)
	}

	err = dir.PreviewPixels(logr(), "", "")
	require.Error(t, err)
}
