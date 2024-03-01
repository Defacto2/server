package command_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func logr() *zap.SugaredLogger {
	return zap.NewExample().Sugar()
}

func TestLookups(t *testing.T) {
	t.Parallel()
	t1 := command.Lookups()
	t2 := command.Infos()
	assert.Equal(t, len(t1), len(t2))
	assert.Contains(t, t2[0], command.Arc)
}

func TestRemoveImgs(t *testing.T) {
	t.Parallel()
	err := command.RemoveImgs("", "", "")
	require.NoError(t, err)
	td := os.TempDir()

	tmp, err := os.CreateTemp(td, "command_test")
	require.NoError(t, err)
	defer os.Remove(tmp.Name())

	nname := tmp.Name() + ".jpg"
	err = os.Rename(tmp.Name(), nname)
	require.NoError(t, err)

	name := strings.TrimSuffix(filepath.Base(tmp.Name()), filepath.Ext(tmp.Name()))

	err = command.RemoveImgs(name, td, "somerandomname")
	require.NoError(t, err)
}

func TestRemoveMe(t *testing.T) {
	t.Parallel()
	err := command.RemoveMe("", "")
	require.NoError(t, err)

	td := os.TempDir()
	tmp, err := os.CreateTemp(td, "command_test")
	require.NoError(t, err)
	defer os.Remove(tmp.Name())

	nname := tmp.Name() + ".txt"
	err = os.Rename(tmp.Name(), nname)
	require.NoError(t, err)

	name := strings.TrimSuffix(filepath.Base(tmp.Name()), filepath.Ext(tmp.Name()))
	err = command.RemoveMe(name, td)
	require.NoError(t, err)
}

func TestCopyFile(t *testing.T) {
	t.Parallel()
	err := command.CopyFile(nil, "", "")
	assert.Equal(t, command.ErrZap, err)

	td := os.TempDir()
	tmp, err := os.CreateTemp(td, "command_test")
	require.NoError(t, err)
	defer os.Remove(tmp.Name())

	logr := zap.NewExample().Sugar()
	err = command.CopyFile(logr, "", "")
	require.Error(t, err)

	err = command.CopyFile(logr, tmp.Name(), "")
	require.Error(t, err)
	dst := tmp.Name() + ".txt"
	err = command.CopyFile(logr, tmp.Name(), dst)
	require.NoError(t, err)
	defer os.Remove(dst)
}

func TestBaseName(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"Empty path", "", ""},
		{"No extension", "/path/to/file", "file"},
		{"With extension", "/path/to/file.txt", "file"},
		{"Multiple extensions", "/path/to/file.tar.gz", "file.tar"},
		{"Hidden file", "/path/to/.hidden", ""},
		{"Hidden file with extension", "/path/to/.hidden.txt", ".hidden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, command.BaseName(tt.path))
		})
	}
}

func TestBaseNamePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"Empty path", "", ""},
		{"No extension", "/path/to/file", "/path/to/file"},
		{"With extension", "/path/to/file.txt", "/path/to/file"},
		{"Multiple extensions", "/path/to/file.tar.gz", "/path/to/file.tar"},
		{"Hidden file", "/path/to/.hidden", "/path/to"},
		{"Hidden file with extension", "/path/to/.hidden.txt", "/path/to/.hidden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, command.BaseNamePath(tt.path))
		})
	}
}

func TestLookCmd(t *testing.T) {
	t.Parallel()
	err := command.LookCmd("")
	require.Error(t, err)

	err = command.LookCmd("thiscommanddoesnotexist")
	require.Error(t, err)

	err = command.LookCmd("go")
	require.NoError(t, err)
}

func TestLookVersion(t *testing.T) {
	t.Parallel()
	err := command.LookVersion("", "", "")
	require.Error(t, err)

	err = command.LookVersion("thiscommanddoesnotexist", "", "")
	require.Error(t, err)

	err = command.LookVersion("go", "", "")
	require.Error(t, err)

	// version arg output example:
	// go version go1.16.5 linux/amd64
	err = command.LookVersion("go", "version", "go version go1.")
	require.NoError(t, err)
}

func TestRun(t *testing.T) {
	t.Parallel()
	err := command.Run(nil, "", "")
	assert.Equal(t, command.ErrZap, err)

	logr := zap.NewExample().Sugar()
	err = command.Run(logr, "", "")
	require.Error(t, err)

	err = command.Run(logr, "thiscommanddoesnotexist", "")
	require.Error(t, err)

	const noArgs = ""
	err = command.Run(logr, "go", noArgs)
	// go without args will return an unknown command error
	require.Error(t, err)

	err = command.Run(logr, "go", "version")
	require.NoError(t, err)
}

func TestRunOut(t *testing.T) {
	t.Parallel()
	_, err := command.RunOut(nil, "", "")
	assert.Equal(t, command.ErrZap, err)

	logr := zap.NewExample().Sugar()
	_, err = command.RunOut(logr, "", "")
	require.Error(t, err)

	_, err = command.RunOut(logr, "thiscommanddoesnotexist", "")
	require.Error(t, err)

	const noArgs = ""
	_, err = command.RunOut(logr, "go", noArgs)
	// go without args will return an unknown command error
	require.Error(t, err)

	out, err := command.RunOut(logr, "go", "version")
	require.NoError(t, err)
	assert.Contains(t, string(out), "go version go1.")
}

func TestRunQuiet(t *testing.T) {
	t.Parallel()
	err := command.RunQuiet(nil, "", "")
	assert.Equal(t, command.ErrZap, err)

	err = command.RunQuiet(logr(), "", "")
	require.Error(t, err)

	err = command.RunQuiet(logr(), "thiscommanddoesnotexist", "")
	require.Error(t, err)

	const noArgs = ""
	err = command.RunQuiet(logr(), "go", noArgs)
	// go without args will return an unknown command error
	require.Error(t, err)

	err = command.RunQuiet(logr(), "go", "version")
	require.NoError(t, err)
}

func TestRunWD(t *testing.T) {
	t.Parallel()
	const noWD = ""
	err := command.RunWD(logr(), "go", noWD, "")
	// go without args will return an unknown command error
	require.Error(t, err)

	wd, err := os.Getwd()
	require.NoError(t, err)
	err = command.RunWD(logr(), "go", wd, "version")
	require.NoError(t, err)
}
