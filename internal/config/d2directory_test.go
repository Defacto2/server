package config_test

import (
	"path/filepath"
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

	testdata := filepath.Join("testdata", "IMPLODE.ZIP")
	got := config.Directory(testdata)

	be.Err(t, got.Check())

	s := got.Issue()
	be.True(t, strings.Contains(s, "points to a file"))

	s = got.String()
	be.Equal(t, s, testdata)
}

func TestDirectory(t *testing.T) {
	t.Parallel()

	testdata := filepath.Join("testdata")
	got := config.Directory(testdata)

	be.Err(t, got.Check(), nil)

	s := got.Issue()
	be.Equal(t, s, "")

	s = got.String()
	be.Equal(t, s, testdata)
}
