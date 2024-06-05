package config_test

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/helper"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	unid = "00000000-0000-0000-0000-000000000000" // common universal unique identifier example
	cfid = "00000000-0000-0000-0000000000000000"  // coldfusion uuid example
)

func TestDownloadFS(t *testing.T) {
	t.Parallel()
	unid := uuid.New()
	dir, err := os.MkdirTemp(os.TempDir(), "testdownloadfs")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// create and test empty, mock image files
	const noExt = ""
	exts := []string{
		".txt",      // valid
		".webp",     // invalid
		".png",      // invalid
		".chiptune", // valid
		".zip",      // valid
		".tiff",     // invalid
		".svg",      // invalid
		noExt,       // valid
	}
	const invalid = "invalid-base-name"
	for _, ext := range exts {
		name := filepath.Join(dir, unid.String()+ext)
		_ = helper.Touch(name)
		badName := filepath.Join(dir, invalid+ext)
		_ = helper.Touch(badName)
		cfName := filepath.Join(dir, cfid+ext)
		_ = helper.Touch(cfName)
	}

	const expectedCount = 24
	const expectedResult = 12

	i, err := helper.Count(dir)
	require.NoError(t, err)
	assert.Equal(t, expectedCount, i)

	// test the images function with invalid parameters
	err = config.DownloadFS(nil, "")
	require.NoError(t, err)

	i, err = helper.Count(dir)
	require.NoError(t, err)
	assert.Equal(t, expectedCount, i)

	// test the images function with valid parameters
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		err = config.DownloadFS(nil, path)
		fmt.Fprintln(io.Discard, path)
		require.NoError(t, err)
		return nil
	})
	require.NoError(t, err)

	i, err = helper.Count(dir)
	require.NoError(t, err)

	assert.Equal(t, expectedResult, i)
}

func TestRemoveDownload(t *testing.T) {
	t.Parallel()
	unid := uuid.New()
	dir, err := os.MkdirTemp(os.TempDir(), "testdownload")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// create and test empty, mock image files
	const noExt = ""
	exts := []string{
		".txt",      // valid
		".webp",     // invalid
		".png",      // invalid
		".chiptune", // valid
		".zip",      // valid
		".tiff",     // invalid
		".svg",      // invalid
		noExt,       // valid
	}
	const invalid = "invalid-base-name"
	for _, ext := range exts {
		name := filepath.Join(dir, unid.String()+ext)
		_ = helper.Touch(name)
		badName := filepath.Join(dir, invalid+ext)
		_ = helper.Touch(badName)
		cfName := filepath.Join(dir, cfid+ext)
		_ = helper.Touch(cfName)
	}

	const expectedCount = 24
	const expectedResult = 12

	i, err := helper.Count(dir)
	require.NoError(t, err)
	assert.Equal(t, expectedCount, i)

	// test the images function with invalid parameters
	err = config.RemoveDownload("", dir)
	require.Error(t, err)

	i, err = helper.Count(dir)
	require.NoError(t, err)
	assert.Equal(t, expectedCount, i)

	// test the images function with valid parameters
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := filepath.Base(path)
		fmt.Fprintln(io.Discard, name, path)
		err = config.RemoveDownload(name, path)
		require.NoError(t, err)
		return nil
	})
	require.NoError(t, err)

	i, err = helper.Count(dir)
	require.NoError(t, err)

	assert.Equal(t, expectedResult, i)
}

func TestRemoveImage(t *testing.T) {
	t.Parallel()
	unid := uuid.New()
	dir, err := os.MkdirTemp(os.TempDir(), "testimage")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// create and test empty, mock image files
	exts := []string{
		".jpg",
		".webp",
		".png",
		".gif",
		".bmp",
		".tiff",
		".svg",
	}
	const invalid = "invalid-base-name"
	for _, ext := range exts {
		name := filepath.Join(dir, unid.String()+ext)
		_ = helper.Touch(name)
		badName := filepath.Join(dir, invalid+ext)
		_ = helper.Touch(badName)
		cfName := filepath.Join(dir, cfid+ext)
		_ = helper.Touch(cfName)
	}

	const expectedCount = 21
	i, err := helper.Count(dir)
	require.NoError(t, err)
	assert.Equal(t, expectedCount, i)

	// test the images function with invalid parameters
	err = config.RemoveImage("", dir)
	require.Error(t, err)

	i, err = helper.Count(dir)
	require.NoError(t, err)
	assert.Equal(t, expectedCount, i)

	// test the images function with valid parameters
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := filepath.Base(path)
		err = config.RemoveImage(name, path)
		require.NoError(t, err)
		return nil
	})
	require.NoError(t, err)

	i, err = helper.Count(dir)
	require.NoError(t, err)

	const expectedResult = 4
	assert.Equal(t, expectedResult, i)
}

func TestOverride(t *testing.T) {
	t.Parallel()
	c := config.Config{}
	assert.Empty(t, c)
	c.GoogleIDs = "googleids,googleids2,googleids3"
	c.Override()
	// confirm override
	assert.Empty(t, c.GoogleIDs)
	// confirm, required default port if not set
	assert.Equal(t, uint(config.HTTPPort), c.HTTPPort)
	// defaults
	assert.False(t, c.ReadOnly)
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

func logger() *zap.SugaredLogger {
	return zap.NewExample().Sugar()
}

func TestConfig_String(t *testing.T) {
	t.Parallel()
	c := config.Config{}
	s := c.String()
	assert.Contains(t, s, "Defacto2 server configuration")
}

func TestConfig_Addresses(t *testing.T) {
	t.Parallel()
	c := config.Config{}
	_, err := c.AddressesCLI()
	require.Error(t, err)
	c.HTTPPort = 8080
	s, err := c.AddressesCLI()
	require.NoError(t, err)
	assert.Contains(t, s, "http://localhost:8080")
}

func TestConfig_Startup(t *testing.T) {
	t.Parallel()
	c := config.Config{}
	_, err := c.Addresses()
	require.Error(t, err)
	c.HTTPPort = 8080
	s, err := c.Addresses()
	require.NoError(t, err)
	assert.Contains(t, s, "http://localhost:8080")
}

func TestConfig_Checks(t *testing.T) {
	t.Parallel()
	c := config.Config{}
	err := c.Checks(nil)
	require.Error(t, err)
	err = c.Checks(logger())
	require.NoError(t, err)

	c.HTTPPort = 8080
	c.TLSPort = 8443
	err = c.Checks(logger())
	require.NoError(t, err)

	c.ReadOnly = false
	c.Production = true
	require.NoError(t, err)
	err = c.Checks(logger())
	require.NoError(t, err)
}

func TestCheckDir(t *testing.T) {
	t.Parallel()
	err := config.CheckDir("", "")
	require.Error(t, err)
	err = config.CheckDir("nosuchdir", "")
	require.Error(t, err)
	dir, err := filepath.Abs(td(""))
	require.NoError(t, err)
	err = config.CheckDir(dir, "")
	require.NoError(t, err)
}

func TestStringErr(t *testing.T) {
	t.Parallel()
	_, _, err := config.StringErr(nil)
	require.NoError(t, err)
	c, s, err := config.StringErr(os.ErrNotExist)
	require.NoError(t, err)
	assert.Equal(t, 500, c)
	assert.Equal(t, "500 - internal server error", s)
}

func TestIsHTML3(t *testing.T) {
	t.Parallel()
	ok := config.IsHTML3("")
	assert.False(t, ok)
	ok = config.IsHTML3("html3")
	assert.True(t, ok)
	ok = config.IsHTML3("/html3")
	assert.True(t, ok)
	ok = config.IsHTML3("/html3/")
	assert.True(t, ok)
	ok = config.IsHTML3("/html3/404.html")
	assert.True(t, ok)
	ok = config.IsHTML3("/files/html3/404.html")
	assert.False(t, ok)
}
