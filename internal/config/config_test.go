package config_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
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

func logr() *zap.SugaredLogger {
	return zap.NewExample().Sugar()
}

func TestConfig_String(t *testing.T) {
	t.Parallel()
	c := config.Config{}
	s := c.String()
	assert.Contains(t, s, "active configuration options")
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

func TestLocalSkip(t *testing.T) {
	t.Parallel()
	skip := config.LocalSkip("")
	assert.False(t, skip)
	skip = config.LocalSkip("readmode")
	assert.False(t, skip)
	skip = config.LocalSkip("ReadMode")
	assert.True(t, skip)
}

func TestAccountSkip(t *testing.T) {
	t.Parallel()
	skip := config.AccountSkip("")
	assert.False(t, skip)
	skip = config.AccountSkip("googleids")
	assert.False(t, skip)
	skip = config.AccountSkip("GoogleIDs")
	assert.True(t, skip)
}

func TestConfig_Checks(t *testing.T) {
	t.Parallel()
	c := config.Config{}
	err := c.Checks(nil)
	require.Error(t, err)
	err = c.Checks(logr())
	require.NoError(t, err)

	c.HTTPPort = 8080
	c.TLSPort = 8443
	err = c.Checks(logr())
	require.NoError(t, err)

	c.ReadMode = false
	c.ProductionMode = true
	require.NoError(t, err)
	err = c.Checks(logr())
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
