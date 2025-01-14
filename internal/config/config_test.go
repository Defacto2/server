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
	"github.com/Defacto2/server/internal/zaplog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	i := config.RecordCount(context.TODO(), nil)
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
	w.Close()
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
	err := c.Archives(context.TODO(), nil)
	require.Error(t, err)

	r := config.Zip
	assert.Equal(t, "zip", r.String())

	err = c.Assets(context.TODO(), nil)
	require.Error(t, err)
	err = c.MagicNumbers(context.TODO(), nil, nil)
	require.Error(t, err)
	err = c.Previews(context.TODO(), nil, nil)
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
	logger := zaplog.Timestamp().Sugar()
	ctx := context.WithValue(context.Background(), helper.LoggerKey, logger)
	err := r.ReArchive(ctx, "", "", "")
	require.Error(t, err)

	// test the archive that uses the defunct implode method
	src, err := filepath.Abs(filepath.Join("testdata", "IMPLODE.ZIP"))
	require.NoError(t, err)
	readr, err := os.Open(src)
	require.NoError(t, err)
	defer readr.Close()
	sign := magicnumber.Find(readr)
	assert.Equal(t, magicnumber.PKWAREZipImplode, sign)

	err = r.ReArchive(ctx, src, "", "")
	require.Error(t, err)
	dst := filepath.Dir(src)
	err = r.ReArchive(ctx, src, dst, "")
	require.Error(t, err)
	err = r.ReArchive(ctx, src, dst, "newfile")
	require.NoError(t, err)

	// test the new, re-created archive that uses the common deflate method
	name := filepath.Join(dst, "newfile.zip")
	readr, err = os.Open(name)
	require.NoError(t, err)
	defer readr.Close()
	sign = magicnumber.Find(readr)
	assert.Equal(t, magicnumber.PKWAREZip, sign)
	defer os.Remove(name)
}
