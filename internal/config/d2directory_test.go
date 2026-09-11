package config_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/nalgeon/be"
)

func TestDirectoryX(t *testing.T) {
	t.Parallel()

	const path = "this-is_notValid"
	got := config.Directory(path)

	be.Err(t, got.Check())

	s := got.Issue()
	be.True(t, strings.Contains(s, "not exist"))

	s = got.String()
	be.Equal(t, s, path)
}

func TestDirectoryFile(t *testing.T) {
	t.Parallel()

	got := config.Directory(testdata.Abs())

	be.Err(t, got.Check())

	s := got.Issue()
	be.True(t, strings.Contains(s, "points to a file"))

	s = got.String()
	be.Equal(t, s, testdata.Abs())
}

func TestConfigDirectory(t *testing.T) {
	t.Parallel()

	got := config.Directory(testdata.Dir())

	be.Err(t, got.Check(), nil)

	s := got.Issue()
	be.Equal(t, s, "")

	s = got.String()
	be.Equal(t, s, ".")
}
