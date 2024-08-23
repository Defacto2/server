package rezip_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Defacto2/server/internal/archive/rezip"
	"github.com/Defacto2/server/internal/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func td(name string) string {
	_, file, _, usable := runtime.Caller(0)
	if !usable {
		panic("runtime.Caller failed")
	}
	d := filepath.Join(filepath.Dir(file), "../../")
	x := filepath.Join(d, "testdata", name)
	return x
}

func TestCompress(t *testing.T) {
	t.Parallel()

	src := td("TEST.EXE")

	dir, err := os.MkdirTemp(helper.TmpDir(), "unzip_test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	dest := filepath.Join(dir, "zip_test.zip")

	st, err := os.Stat(src)
	require.NoError(t, err)

	n, err := rezip.Compress(src, dest)
	require.NoError(t, err)

	assert.Equal(t, int64(n), st.Size())

	// confirm the zip file is smaller than the total size of the files
	st, err = os.Stat(dest)
	require.NoError(t, err)
	assert.Less(t, st.Size(), int64(n))

	// confirm command fails when the file already exists
	n, err = rezip.Compress(src, dest)
	require.Error(t, err)
	require.Zero(t, n)

	// confirm command fails when the dest is a directory
	n, err = rezip.Compress(src, dir)
	require.Error(t, err)
	require.Zero(t, n)
}

func TestCompressDir(t *testing.T) {
	t.Parallel()

	srcDir := td("")
	dir, err := os.MkdirTemp(helper.TmpDir(), "unzip_test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	dest := filepath.Join(dir, "unzip_test.zip")

	n, err := rezip.CompressDir(srcDir, dest)
	require.NoError(t, err)

	const fiveMB = 5 * 1024 * 1024
	assert.Greater(t, n, int64(fiveMB))

	// confirm the zip file is smaller than the total size of the files
	st, err := os.Stat(dest)
	require.NoError(t, err)
	assert.Less(t, st.Size(), n)
}

func TestUnzip(t *testing.T) {
	t.Parallel()

	src := td("PKZ80A1.ZIP")
	err := rezip.Test(src)
	require.NoError(t, err)

	src = td("ARJ310.ARJ")
	err = rezip.Test(src)
	require.Error(t, err)
}
