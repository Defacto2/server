package command_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/helper"
	"github.com/stretchr/testify/assert"
)

func Test_ExtractOne(t *testing.T) {
	t.Parallel()
	err := command.ExtractOne(nil, "", "", "", "")
	assert.Error(t, err)

	src, dst, hint, name := "", "", "", ""
	err = command.ExtractOne(z(), src, dst, hint, name)
	assert.Error(t, err)

	src, _ = filepath.Abs("testdata/PKZ80A1.TXT")
	err = command.ExtractOne(z(), src, "", "", "")
	assert.Error(t, err)

	src, _ = filepath.Abs("testdata/PKZ80A1.ZIP")
	err = command.ExtractOne(z(), src, "", "", "")
	assert.Error(t, err)

	err = command.ExtractOne(z(), src, "", ".zip", "")
	assert.Error(t, err)

	err = command.ExtractOne(z(), src, "", ".zip", "TEST.ASC")
	assert.Error(t, err)

	dst = os.TempDir()
	err = command.ExtractOne(z(), src, dst, ".zip", "TEST.ASC")
	assert.Error(t, err)

	src, _ = filepath.Abs("testdata/PKZ204EX.ZIP")
	dst, _ = filepath.Abs(filepath.Join(os.TempDir(), "PKZ204EX.ZIP"))
	err = command.ExtractOne(z(), src, dst, ".zip", "TEST.ASC")
	assert.NoError(t, err)

	ok := helper.IsFile(dst)
	assert.True(t, ok)
	defer os.Remove(dst)

	src, _ = filepath.Abs("testdata/PKZ80A1.ZIP")
	dst, _ = filepath.Abs(filepath.Join(os.TempDir(), "PKZ80A1.ZIP"))
	err = command.ExtractOne(z(), src, dst, ".zip", "TEST.ASC")
	assert.NoError(t, err)

	ok = helper.IsFile(dst)
	assert.True(t, ok)
	defer os.Remove(dst)

	src, _ = filepath.Abs("testdata/ARC521P.ARC")
	dst, _ = filepath.Abs(filepath.Join(os.TempDir(), "ARC521P.ARC"))
	err = command.ExtractOne(z(), src, dst, ".arc", "TEST.JPG")
	assert.NoError(t, err)

	ok = helper.IsFile(dst)
	assert.True(t, ok)
	defer os.Remove(dst)

	src, _ = filepath.Abs("testdata/ARJ310.ARJ")
	dst, _ = filepath.Abs(filepath.Join(os.TempDir(), "ARJ310.ARJ"))
	err = command.ExtractOne(z(), src, dst, ".arj", "TEST.JPEG")
	assert.NoError(t, err)

	ok = helper.IsFile(dst)
	assert.True(t, ok)
	defer os.Remove(dst)

	src, _ = filepath.Abs("testdata/RAR624.RAR")
	dst, _ = filepath.Abs(filepath.Join(os.TempDir(), "RAR624.RAR"))
	err = command.ExtractOne(z(), src, dst, ".rar", "TEST.JPG")
	assert.NoError(t, err)

	ok = helper.IsFile(dst)
	assert.True(t, ok)
	defer os.Remove(dst)

	src, _ = filepath.Abs("testdata/TAR135.TAR")
	dst, _ = filepath.Abs(filepath.Join(os.TempDir(), "TAR135.TAR"))
	err = command.ExtractOne(z(), src, dst, ".tar", "TEST.JPG")
	assert.NoError(t, err)

	ok = helper.IsFile(dst)
	assert.True(t, ok)
	defer os.Remove(dst)

	src, _ = filepath.Abs("testdata/TAR135.GZ")
	dst, _ = filepath.Abs(filepath.Join(os.TempDir(), "TAR135.GZ"))
	err = command.ExtractOne(z(), src, dst, ".gz", "TEST.JPG")
	assert.NoError(t, err)

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
	assert.Error(t, err)
	s = command.ArjExitStatus(err)
	assert.Equal(t, "user error, bad command line parameters", s)
}

func Test_UnRarExitStatus(t *testing.T) {
	t.Parallel()
	s := command.UnRarExitStatus(nil)
	assert.Equal(t, "", s)
	cmd := exec.Command("unrar", "throwawayarg")
	err := cmd.Run()
	assert.Error(t, err)
	s = command.UnRarExitStatus(err)
	assert.Equal(t, "wrong command line option", s)
}

func Test_ExtractAnsiLove(t *testing.T) {
	t.Parallel()
	dir := command.Dirs{}

	err := dir.ExtractAnsiLove(nil, "", "", "", "")
	assert.Error(t, err)

	err = dir.ExtractAnsiLove(z(), "", "", "", "")
	assert.Error(t, err)

	err = dir.ExtractAnsiLove(z(),
		"testdata/PKZ204EX.ZIP", ".zip", "000000ABCDE", "TEST.ANS")
	assert.NoError(t, err)
	ok := helper.IsFile("000000ABCDE.webp")
	assert.True(t, ok)
	err = os.Remove("000000ABCDE.webp")
	assert.NoError(t, err)

	err = dir.ExtractAnsiLove(z(),
		"testdata/PKZ204EX.ZIP", ".zip", "000000ABCDE", "nosuchfile")
	assert.Error(t, err)
}

func Test_ExtractImage(t *testing.T) {
	t.Parallel()
	prev, err := os.MkdirTemp(os.TempDir(), "preview")
	assert.NoError(t, err)
	thumb, err := os.MkdirTemp(os.TempDir(), "thumb")
	assert.NoError(t, err)
	dl, err := os.MkdirTemp(os.TempDir(), "download")
	assert.NoError(t, err)
	dir := command.Dirs{
		Download:  dl,    // this prefixes to UUID
		Preview:   prev,  // this is the output dest
		Thumbnail: thumb, // this is the cropped output dest
	}
	// intentional errors
	err = dir.ExtractImage(nil, "", "", "", "")
	assert.Error(t, err)
	err = dir.ExtractImage(z(), "", "", "", "")
	assert.Error(t, err)

	broken := []string{"nosuchfile", "TEST.ASC", "TEST.JPEG", "TEST.ANS"}
	for _, name := range broken {
		err = dir.ExtractImage(z(),
			"testdata/PKZ204EX.ZIP", ".zip", "000000ABCDE", name)
		assert.Error(t, err)
	}
	// cases that create webp files
	names := []string{"TEST.JPG", "TEST.GIF", "TEST.BMP"}
	op := filepath.Join(prev, "000000ABCDE.webp")
	ot := filepath.Join(thumb, "000000ABCDE.webp")
	for _, name := range names {
		//continue
		err = dir.ExtractImage(z(),
			"testdata/PKZ204EX.ZIP", ".zip", "000000ABCDE", name)
		assert.NoError(t, err)

		ok := helper.IsFile(op)
		assert.True(t, ok)
		ok = helper.IsFile(ot)
		assert.True(t, ok)
		err = os.Remove(op)
		assert.NoError(t, err)
		err = os.Remove(ot)
		assert.NoError(t, err)
	}
	// unique case that creates compressed PNG files
	{
		op = filepath.Join(prev, "000000ABCDE.png")
		ot = filepath.Join(thumb, "000000ABCDE.webp")
		name := "TEST.PNG"
		err = dir.ExtractImage(z(),
			"testdata/PKZ204EX.ZIP", ".zip", "000000ABCDE", name)
		assert.NoError(t, err)
		ok := helper.IsFile(op)
		assert.True(t, ok)
		ok = helper.IsFile(ot)
		assert.True(t, ok)
		err = os.Remove(op)
		assert.NoError(t, err)
		err = os.Remove(ot)
		assert.NoError(t, err)
	}
}

func Test_LosslessScreenshot(t *testing.T) {
	t.Parallel()
	prev, err := os.MkdirTemp(os.TempDir(), "preview")
	assert.NoError(t, err)
	thumb, err := os.MkdirTemp(os.TempDir(), "thumb")
	assert.NoError(t, err)
	dl, err := os.MkdirTemp(os.TempDir(), "download")
	assert.NoError(t, err)
	dir := command.Dirs{
		Download:  dl,    // this prefixes to UUID
		Preview:   prev,  // this is the output dest
		Thumbnail: thumb, // this is the cropped output dest
	}
	imgs := []string{"TEST.BMP", "TEST.GIF", "TEST.JPG", "TEST.PCX", "TEST.PNG"}
	for _, name := range imgs {
		fp := filepath.Join("testdata", "uncompress", name)
		err = dir.LosslessScreenshot(z(), fp, "000000ABCDE")
		fmt.Println(err)
		assert.NoError(t, err)
	}

	err = dir.LosslessScreenshot(nil, "", "")
	assert.Error(t, err)
}