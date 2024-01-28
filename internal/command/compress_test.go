package command_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/helper"
	"github.com/stretchr/testify/assert"
)

func Test_ExtractOne(t *testing.T) {
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
	assert.Nil(t, err)

	ok := helper.IsFile(dst)
	assert.True(t, ok)
	defer os.Remove(dst)

	src, _ = filepath.Abs("testdata/PKZ80A1.ZIP")
	dst, _ = filepath.Abs(filepath.Join(os.TempDir(), "PKZ80A1.ZIP"))
	err = command.ExtractOne(z(), src, dst, ".zip", "TEST.ASC")
	assert.Nil(t, err)

	ok = helper.IsFile(dst)
	assert.True(t, ok)
	defer os.Remove(dst)

	src, _ = filepath.Abs("testdata/ARC521P.ARC")
	dst, _ = filepath.Abs(filepath.Join(os.TempDir(), "ARC521P.ARC"))
	err = command.ExtractOne(z(), src, dst, ".arc", "TEST.JPG")
	assert.Nil(t, err)

	ok = helper.IsFile(dst)
	assert.True(t, ok)
	defer os.Remove(dst)

	src, _ = filepath.Abs("testdata/ARJ310.ARJ")
	dst, _ = filepath.Abs(filepath.Join(os.TempDir(), "ARJ310.ARJ"))
	err = command.ExtractOne(z(), src, dst, ".arj", "TEST.JPEG")
	assert.Nil(t, err)

	ok = helper.IsFile(dst)
	assert.True(t, ok)
	defer os.Remove(dst)

	src, _ = filepath.Abs("testdata/RAR624.RAR")
	dst, _ = filepath.Abs(filepath.Join(os.TempDir(), "RAR624.RAR"))
	err = command.ExtractOne(z(), src, dst, ".rar", "TEST.JPG")
	assert.Nil(t, err)

	ok = helper.IsFile(dst)
	assert.True(t, ok)
	defer os.Remove(dst)

	src, _ = filepath.Abs("testdata/TAR135.TAR")
	dst, _ = filepath.Abs(filepath.Join(os.TempDir(), "TAR135.TAR"))
	err = command.ExtractOne(z(), src, dst, ".tar", "TEST.JPG")
	assert.Nil(t, err)

	ok = helper.IsFile(dst)
	assert.True(t, ok)
	defer os.Remove(dst)

	src, _ = filepath.Abs("testdata/TAR135.GZ")
	dst, _ = filepath.Abs(filepath.Join(os.TempDir(), "TAR135.GZ"))
	err = command.ExtractOne(z(), src, dst, ".gz", "TEST.JPG")
	assert.Nil(t, err)

	ok = helper.IsFile(dst)
	assert.True(t, ok)
	defer os.Remove(dst)
}

func Test_ArjExitStatus(t *testing.T) {
	s := command.ArjExitStatus(nil)
	assert.Equal(t, "", s)
	cmd := exec.Command("arj", "throwawayarg")
	err := cmd.Run()
	assert.Error(t, err)
	s = command.ArjExitStatus(err)
	assert.Equal(t, "user error, bad command line parameters", s)
}

func Test_UnRarExitStatus(t *testing.T) {
	s := command.UnRarExitStatus(nil)
	assert.Equal(t, "", s)
	cmd := exec.Command("unrar", "throwawayarg")
	err := cmd.Run()
	assert.Error(t, err)
	s = command.UnRarExitStatus(err)
	assert.Equal(t, "wrong command line option", s)
}
