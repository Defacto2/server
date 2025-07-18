package config_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/dir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestConfig(t *testing.T) {
	t.Parallel()
	c := config.Config{}
	x := c.List()
	assert.NotEmpty(t, x)
	s := c.Envs()
	assert.NotEmpty(t, s)
	s = c.Helps()
	assert.NotEmpty(t, s)
	s = c.Names()
	assert.NotEmpty(t, s)
	s = c.Values()
	assert.NotEmpty(t, s)
	cs := c.String()
	assert.Contains(t, cs, "configuration")
	cs, err := c.Addresses()
	require.Error(t, err)
	assert.Empty(t, cs)
}

func TestChecks(t *testing.T) {
	t.Parallel()
	c := config.Config{}
	err := c.Checks(nil)
	require.Error(t, err)
	err = c.LogStore()
	require.NoError(t, err)
	err = c.SetupLogDir(nil)
	require.Error(t, err)
}

func TestCheckDir(t *testing.T) {
	t.Parallel()
	err := config.CheckDir("", "")
	require.Error(t, err)
	err = config.CheckDir("xyz", "")
	require.Error(t, err)
}

func TestRecordCount(t *testing.T) {
	t.Parallel()
	i := config.RecordCount(t.Context(), nil)
	assert.Zero(t, i)
}

func TestSanityTmpDir(t *testing.T) {
	t.Parallel()
	var stderrBuf bytes.Buffer
	oldStdout := os.Stdout
	// defer to restore original stderr after the test
	defer func() { os.Stdout = oldStdout }()
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	config.SanityTmpDir()
	if err := w.Close(); err != nil {
		t.Error(err)
	}
	_, err = stderrBuf.ReadFrom(r)
	require.NoError(t, err)
	expectedMessage := "Temporary directory using"
	require.Contains(t, stderrBuf.String(), expectedMessage)
}

func TestValidate(t *testing.T) {
	t.Parallel()
	err := config.Validate(0)
	require.NoError(t, err)
	const tooLarge = 10000000
	err = config.Validate(tooLarge)
	require.Error(t, err)
}

func TestError(t *testing.T) {
	t.Parallel()
	i, s, err := config.StringErr(nil)
	assert.Zero(t, i)
	assert.Empty(t, s)
	require.NoError(t, err)
	i, s, err = config.StringErr(assert.AnError)
	assert.Equal(t, 500, i)
	assert.Contains(t, s, "internal server error")
	require.NoError(t, err)
}

func TestRepair(t *testing.T) {
	t.Parallel()
	c := config.Config{}
	err := c.Archives(t.Context(), nil)
	require.Error(t, err)

	r := config.Zip
	assert.Equal(t, "zip", r.String())

	err = c.Assets(t.Context(), nil)
	require.Error(t, err)
	err = c.MagicNumbers(t.Context(), nil, nil)
	require.Error(t, err)
	err = c.Previews(t.Context(), nil, nil)
	require.Error(t, err)

	err = c.ImageDirs(nil)
	require.NoError(t, err)
	err = config.DownloadDir(nil, "", "", "")
	require.Error(t, err)
	err = config.RenameDownload("", "")
	require.Error(t, err)
	err = config.RemoveDir("", "", "")
	require.Error(t, err)
	err = config.RemoveDownload("", "", "", "")
	require.Error(t, err)
	err = config.RemoveImage("", "", "")
	require.Error(t, err)
}

func TestReArchive(t *testing.T) {
	t.Parallel()
	r := config.Zip
	logger, _ := zap.NewProduction()
	_ = logger.Sync()
	ctx := context.WithValue(t.Context(), helper.LoggerKey, logger)
	err := r.ReArchive(ctx, "", "", "")
	require.Error(t, err)
}

func TestReArchiveImplode(t *testing.T) {
	r := config.Zip
	l, _ := zap.NewProduction()
	_ = l.Sync()
	logger := l.Sugar()
	ctx := context.WithValue(t.Context(), helper.LoggerKey, logger)

	// test the archive that uses the defunct implode method
	src, err := filepath.Abs(filepath.Join("testdata", "IMPLODE.ZIP"))
	require.NoError(t, err)
	readr, err := os.Open(src)
	require.NoError(t, err)
	defer func() {
		_ = readr.Close()
	}()
	sign := magicnumber.Find(readr)
	assert.Equal(t, magicnumber.PKWAREZipImplode, sign)

	err = r.ReArchive(ctx, src, "", "")
	require.Error(t, err)
	dst := dir.Directory(filepath.Dir(src))
	err = r.ReArchive(ctx, src, "", dst)
	require.Error(t, err)
	err = r.ReArchive(ctx, src, "newfile", dst)
	require.NoError(t, err)

	// test the new, re-created archive that uses the common deflate method
	name := dst.Join("newfile.zip")
	readr, err = os.Open(name)
	require.NoError(t, err)
	defer func() {
		err := readr.Close()
		require.NoError(t, err)
	}()
	sign = magicnumber.Find(readr)
	assert.Equal(t, magicnumber.PKWAREZip, sign)
	defer func() {
		err := os.Remove(name)
		require.NoError(t, err)
	}()
}
