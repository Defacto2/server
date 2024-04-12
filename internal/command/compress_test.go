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
	err := command.ExtractOne(nil, "", "", "", "")
	require.Error(t, err)

	src, dst, hint, name := "", "", "", ""
	err = command.ExtractOne(logr(), src, dst, hint, name)
	require.Error(t, err)

	src = td("PKZ80A1.TXT")
	err = command.ExtractOne(logr(), src, "", "", "")
	require.Error(t, err)

	src = td("PKZ80A1.ZIP")
	err = command.ExtractOne(logr(), src, "", ".zip", "")
	require.Error(t, err)

	err = command.ExtractOne(logr(), src, "", ".zip", "TEST.ASC")
	require.Error(t, err)

	dst = os.TempDir()
	err = command.ExtractOne(logr(), src, dst, ".zip", "TEST.ASC")
	require.Error(t, err)

	src = td("PKZ204EX.ZIP")
	dst, _ = filepath.Abs(filepath.Join(os.TempDir(), "PKZ204EX.ZIP"))
	err = command.ExtractOne(logr(), src, dst, ".zip", "TEST.ASC")
	require.NoError(t, err)

	ok := helper.IsFile(dst)
	assert.True(t, ok)
	defer os.Remove(dst)

	src = td("PKZ80A1.ZIP")
	dst, _ = filepath.Abs(filepath.Join(os.TempDir(), "PKZ80A1.ZIP"))
	err = command.ExtractOne(logr(), src, dst, ".zip", "TEST.ASC")
	require.NoError(t, err)

	ok = helper.IsFile(dst)
	assert.True(t, ok)
	defer os.Remove(dst)

	src = td("ARC521P.ARC")
	dst, _ = filepath.Abs(filepath.Join(os.TempDir(), "ARC521P.ARC"))
	err = command.ExtractOne(logr(), src, dst, ".arc", "TEST.JPG")
	require.NoError(t, err)

	ok = helper.IsFile(dst)
	assert.True(t, ok)
	defer os.Remove(dst)

	src = td("ARJ310.ARJ")
	dst, _ = filepath.Abs(filepath.Join(os.TempDir(), "ARJ310.ARJ"))
	err = command.ExtractOne(logr(), src, dst, ".arj", "TEST.JPEG")
	require.NoError(t, err)

	ok = helper.IsFile(dst)
	assert.True(t, ok)
	defer os.Remove(dst)

	src = td("RAR624.RAR")
	dst, _ = filepath.Abs(filepath.Join(os.TempDir(), "RAR624.RAR"))
	err = command.ExtractOne(logr(), src, dst, ".rar", "TEST.JPG")
	require.NoError(t, err)

	ok = helper.IsFile(dst)
	assert.True(t, ok)
	defer os.Remove(dst)

	src = td("TAR135.TAR")
	dst, _ = filepath.Abs(filepath.Join(os.TempDir(), "TAR135.TAR"))
	err = command.ExtractOne(logr(), src, dst, ".tar", "TEST.JPG")
	require.NoError(t, err)

	ok = helper.IsFile(dst)
	assert.True(t, ok)
	defer os.Remove(dst)

	src = td("TAR135.GZ")
	dst, _ = filepath.Abs(filepath.Join(os.TempDir(), "TAR135.GZ"))
	err = command.ExtractOne(logr(), src, dst, ".gz", "TEST.JPG")
	require.NoError(t, err)

	ok = helper.IsFile(dst)
	assert.True(t, ok)
	defer os.Remove(dst)
}

func Test_ArjExitStatus(t *testing.T) {
	t.Parallel()
	s := command.ArjExitStatus(nil)
	assert.Equal(t, "", s)
	cmd := exec.Command("arj", "throwawayarg")
	err := cmd.Run()
	require.Error(t, err)
	s = command.ArjExitStatus(err)
	assert.Equal(t, "user error, bad command line parameters", s)
}

func Test_UnRarExitStatus(t *testing.T) {
	t.Parallel()
	s := command.UnRarExitStatus(nil)
	assert.Equal(t, "", s)
	cmd := exec.Command("unrar", "throwawayarg")
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
	ok := helper.IsFile("000000ABCDE.webp")
	assert.True(t, ok)
	err = os.Remove("000000ABCDE.webp")
	require.NoError(t, err)

	err = dir.ExtractAnsiLove(logr(),
		src, ".zip", "000000ABCDE", "nosuchfile")
	require.Error(t, err)
}

func Test_ExtractImage(t *testing.T) {
	t.Parallel()
	prev, err := os.MkdirTemp(os.TempDir(), "preview")
	require.NoError(t, err)
	thumb, err := os.MkdirTemp(os.TempDir(), "thumb")
	require.NoError(t, err)
	dl, err := os.MkdirTemp(os.TempDir(), "download")
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

		ok := helper.IsFile(op)
		assert.True(t, ok)
		ok = helper.IsFile(ot)
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
		ok := helper.IsFile(op)
		assert.True(t, ok)
		ok = helper.IsFile(ot)
		assert.True(t, ok)
		err = os.Remove(op)
		require.NoError(t, err)
		err = os.Remove(ot)
		require.NoError(t, err)
	}
}

func Test_LosslessScreenshot(t *testing.T) {
	t.Parallel()
	prev, err := os.MkdirTemp(os.TempDir(), "preview")
	require.NoError(t, err)
	thumb, err := os.MkdirTemp(os.TempDir(), "thumb")
	require.NoError(t, err)
	dl, err := os.MkdirTemp(os.TempDir(), "download")
	require.NoError(t, err)
	dir := command.Dirs{
		Download:  dl,    // this prefixes to UUID
		Preview:   prev,  // this is the output dest
		Thumbnail: thumb, // this is the cropped output dest
	}
	imgs := []string{"TEST.BMP", "TEST.GIF", "TEST.JPG", "TEST.PCX", "TEST.PNG"}
	for _, name := range imgs {
		fp := tduncompress(name)
		err = dir.LosslessScreenshot(logr(), fp, "000000ABCDE")
		require.NoError(t, err)
	}

	err = dir.LosslessScreenshot(logr(), "", "")
	require.Error(t, err)
}
