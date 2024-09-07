package config_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/Defacto2/server/internal/config"
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
	assert.Error(t, err)
	assert.Empty(t, cs)
}

func TestChecks(t *testing.T) {
	t.Parallel()
	c := config.Config{}
	err := c.Checks(nil)
	assert.Error(t, err)
	err = c.LogStore()
	assert.NoError(t, err)
	err = c.SetupLogDir(nil)
	assert.Error(t, err)
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
	assert.NoError(t, err)
	const tooLarge = 10000000
	err = config.Validate(tooLarge)
	assert.Error(t, err)
}

func TestError(t *testing.T) {
	t.Parallel()
	i, s, err := config.StringErr(nil)
	assert.Zero(t, i)
	assert.Empty(t, s)
	assert.NoError(t, err)
	i, s, err = config.StringErr(assert.AnError)
	assert.Equal(t, 500, i)
	assert.Contains(t, s, "internal server error")
	assert.NoError(t, err)
}

func TestRepair(t *testing.T) {
	t.Parallel()
	c := config.Config{}
	err := c.Archives(context.TODO(), nil)
	assert.Error(t, err)

	r := config.Zip
	assert.Equal(t, "zip", r.String())

	err = c.Assets(context.TODO(), nil)
	assert.Error(t, err)
	err = c.MagicNumbers(context.TODO(), nil, nil)
	assert.Error(t, err)
	err = c.Previews(context.TODO(), nil, nil)
	assert.Error(t, err)

	err = config.ImageDirs(nil, c)
	assert.NoError(t, err)
	err = config.DownloadDir(nil, "", "", "")
	assert.Error(t, err)
	err = config.RenameDownload("", "")
	assert.Error(t, err)
	err = config.RemoveDir("", "", "")
	assert.Error(t, err)
	err = config.RemoveDownload("", "", "", "")
	assert.Error(t, err)
	err = config.RemoveImage("", "", "")
	assert.Error(t, err)
}
