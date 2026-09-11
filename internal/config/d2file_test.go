package config_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

var testdata = testutil.FileData("implodezip")

func TestFileX(t *testing.T) {
	t.Parallel()

	const path = "this-is_notValid"
	got := config.File(path)

	be.Err(t, got.Check())

	s := got.Issue()
	be.True(t, strings.Contains(s, "not exist"))

	s = got.String()
	be.Equal(t, s, path)
}

func TestFileDirectory(t *testing.T) {
	t.Parallel()

	got := config.File(testdata.Dir())

	be.Err(t, got.Check())

	s := got.Issue()
	be.True(t, strings.Contains(s, "points to a directory"))

	s = got.String()
	be.Equal(t, s, testdata.Dir())
}

func TestConfigFile(t *testing.T) {
	t.Parallel()

	got := config.File(testdata.Abs())

	be.Err(t, got.Check(), nil)

	s := got.Issue()
	be.Equal(t, s, "")

	s = got.String()
	be.Equal(t, s, testdata.Abs())
}
