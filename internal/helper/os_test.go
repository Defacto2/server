package helper_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Defacto2/server/internal/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDataFileCount = 61

func TestCount(t *testing.T) {
	dir, err := filepath.Abs("../../assets/testdata")
	require.NoError(t, err)

	i, err := helper.Count("")
	require.Error(t, err)
	assert.Equal(t, 0, i)

	i, err = helper.Count("nosuchfile")
	require.Error(t, err)
	assert.Equal(t, 0, i)

	i, err = helper.Count(dir)
	require.NoError(t, err)
	assert.Equal(t, testDataFileCount, i)
}

func TestDuplicate(t *testing.T) {
	dir, err := filepath.Abs("../../assets/testdata")
	require.NoError(t, err)

	r, err := helper.Duplicate(dir, "")
	require.Error(t, err)
	assert.Empty(t, r)
	r, err = helper.Duplicate(dir, dir)
	require.Error(t, err)
	assert.Empty(t, r)

	file, err := filepath.Abs("../../assets/testdata/uncompress/TEST.NFO")
	require.NoError(t, err)

	r, err = helper.Duplicate(file, "")
	require.Error(t, err)
	assert.Empty(t, r)

	r, err = helper.Duplicate("", file)
	require.Error(t, err)
	assert.Empty(t, r)

	r, err = helper.Duplicate(file, file)
	require.Error(t, err)
	assert.Empty(t, r)

	dest, err := os.MkdirTemp(os.TempDir(), "test_duplicate")
	require.NoError(t, err)
	defer os.RemoveAll(dest)

	r, err = helper.Duplicate(file, dest)
	require.Error(t, err)
	assert.Empty(t, r)

	dest = filepath.Join(dest, "TEST.NFO")
	written, err := helper.Duplicate(file, dest)
	require.NoError(t, err)
	assert.Equal(t, written, int64(13))
}

func TestFiles(t *testing.T) {
	dir, err := filepath.Abs("../../assets/testdata")
	require.NoError(t, err)

	r, err := helper.Files("")
	require.Error(t, err)
	assert.Empty(t, r)

	r, err = helper.Files("nosuchfile")
	require.Error(t, err)
	assert.Empty(t, r)

	r, err = helper.Files(dir)
	require.NoError(t, err)
	assert.Len(t, r, testDataFileCount)
}

func TestLines(t *testing.T) {
	i, err := helper.Lines("")
	require.Error(t, err)
	assert.Equal(t, 0, i)

	i, err = helper.Lines("nosuchfile")
	require.Error(t, err)
	assert.Equal(t, 0, i)

	i, err = helper.Lines(td(""))
	require.Error(t, err)
	assert.Equal(t, 0, i)

	i, err = helper.Lines(td("TEST.BMP"))
	require.Error(t, err)
	assert.Equal(t, 0, i)

	i, err = helper.Lines(td("PKZ80A1.TXT"))
	require.NoError(t, err)
	assert.Equal(t, 175, i)
}

func td(name string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("runtime.Caller failed")
	}
	d := filepath.Join(filepath.Dir(file), "../..")
	x := filepath.Join(d, "assets", "testdata", name)
	return x
}

func TestRenameFile(t *testing.T) {
	const name = "test_rename_file"

	err := helper.RenameFile("", "")
	require.Error(t, err)

	dir, err := os.MkdirTemp(os.TempDir(), "test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	err = helper.RenameFile(dir, "")
	require.ErrorIs(t, err, helper.ErrFilePath)

	abs := filepath.Join(dir, name)
	err = helper.Touch(abs)
	require.NoError(t, err)

	err = helper.RenameFile(abs, "")
	require.Error(t, err)

	err = helper.RenameFile(abs, abs)
	require.Error(t, err)

	err = helper.RenameFile(abs, abs+"~")
	require.NoError(t, err)
}

func TestRenameFileOW(t *testing.T) {
	const name = "test_rename_file"

	err := helper.RenameFileOW("", "")
	require.Error(t, err)

	dir, err := os.MkdirTemp(os.TempDir(), "test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	err = helper.RenameFileOW(dir, "")
	require.ErrorIs(t, err, helper.ErrFilePath)

	abs := filepath.Join(dir, name)
	err = helper.Touch(abs)
	require.NoError(t, err)

	err = helper.RenameFileOW(abs, "")
	require.Error(t, err)

	err = helper.RenameFileOW(abs, abs)
	require.Error(t, err)

	err = helper.RenameFileOW(abs, abs+"~")
	require.NoError(t, err)
}

func TestRenameCrossDevice(t *testing.T) {
	const name = "test_rename_file"

	err := helper.RenameCrossDevice("", "")
	require.Error(t, err)

	dir, err := os.MkdirTemp(os.TempDir(), "test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	err = helper.RenameCrossDevice(dir, "")
	require.Error(t, err)

	abs := filepath.Join(dir, name)
	err = helper.Touch(abs)
	require.NoError(t, err)

	err = helper.RenameCrossDevice(abs, "")
	require.Error(t, err)

	err = helper.RenameCrossDevice(abs, abs+"~")
	require.NoError(t, err)
}

func TestSize(t *testing.T) {
	const name = "test_rename_file"
	const none = int64(-1)
	data := []byte("Hello, World!")

	i := helper.Size("")
	assert.Equal(t, none, i)

	i = helper.Size("nosuchfile")
	assert.Equal(t, none, i)

	dir, err := os.MkdirTemp(os.TempDir(), "test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	abs := filepath.Join(dir, name)
	x, err := helper.TouchW(abs, data...)
	require.NoError(t, err)

	i = helper.Size(abs)
	assert.Equal(t, int64(x), i)
}

func TestStrongIntegrity(t *testing.T) {
	const name = "test_strong_integrity"
	const expected = "5485cc9b3365b4305dfb4e8337e0a598a574f8242bf17289e0" +
		"dd6c20a3cd44a089de16ab4ab308f63e44b1170eb5f515"
	data := []byte("Hello, World!")

	s, err := helper.StrongIntegrity("")
	require.Error(t, err)
	assert.Empty(t, s)

	s, err = helper.StrongIntegrity("nosuchfile")
	require.Error(t, err)
	assert.Empty(t, s)

	dir, err := os.MkdirTemp(os.TempDir(), "test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	abs := filepath.Join(dir, name)
	_, err = helper.TouchW(abs, data...)
	require.NoError(t, err)

	s, err = helper.StrongIntegrity(abs)
	require.NoError(t, err)
	assert.Equal(t, expected, s)
}
