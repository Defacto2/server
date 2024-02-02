package command_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/command"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func z() *zap.SugaredLogger {
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
	assert.NoError(t, err)
	td := os.TempDir()

	tmp, err := os.CreateTemp(td, "command_test")
	assert.NoError(t, err)
	defer os.Remove(tmp.Name())

	nname := tmp.Name() + ".jpg"
	err = os.Rename(tmp.Name(), nname)
	assert.NoError(t, err)

	name := strings.TrimSuffix(filepath.Base(tmp.Name()), filepath.Ext(tmp.Name()))

	err = command.RemoveImgs(name, td, "somerandomname")
	assert.NoError(t, err)
}

func TestRemoveMe(t *testing.T) {
	t.Parallel()
	err := command.RemoveMe("", "")
	assert.NoError(t, err)

	td := os.TempDir()
	tmp, err := os.CreateTemp(td, "command_test")
	assert.NoError(t, err)
	defer os.Remove(tmp.Name())

	nname := tmp.Name() + ".txt"
	err = os.Rename(tmp.Name(), nname)
	assert.NoError(t, err)

	name := strings.TrimSuffix(filepath.Base(tmp.Name()), filepath.Ext(tmp.Name()))
	err = command.RemoveMe(name, td)
	assert.NoError(t, err)
}

func TestCopyFile(t *testing.T) {
	t.Parallel()
	err := command.CopyFile(nil, "", "")
	assert.Equal(t, command.ErrZap, err)

	td := os.TempDir()
	tmp, err := os.CreateTemp(td, "command_test")
	assert.NoError(t, err)
	defer os.Remove(tmp.Name())

	z := zap.NewExample().Sugar()
	err = command.CopyFile(z, "", "")
	assert.Error(t, err)

	err = command.CopyFile(z, tmp.Name(), "")
	assert.Error(t, err)
	dst := tmp.Name() + ".txt"
	err = command.CopyFile(z, tmp.Name(), dst)
	assert.NoError(t, err)
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
	assert.Error(t, err)

	err = command.LookCmd("thiscommanddoesnotexist")
	assert.Error(t, err)

	err = command.LookCmd("go")
	assert.NoError(t, err)
}

func TestLookVersion(t *testing.T) {
	t.Parallel()
	err := command.LookVersion("", "", "")
	assert.Error(t, err)

	err = command.LookVersion("thiscommanddoesnotexist", "", "")
	assert.Error(t, err)

	err = command.LookVersion("go", "", "")
	assert.Error(t, err)

	// version arg output example:
	// go version go1.16.5 linux/amd64
	err = command.LookVersion("go", "version", "go version go1.")
	assert.NoError(t, err)
}

func TestRun(t *testing.T) {
	t.Parallel()
	err := command.Run(nil, "", "")
	assert.Equal(t, command.ErrZap, err)

	z := zap.NewExample().Sugar()
	err = command.Run(z, "", "")
	assert.Error(t, err)

	err = command.Run(z, "thiscommanddoesnotexist", "")
	assert.Error(t, err)

	const noArgs = ""
	err = command.Run(z, "go", noArgs)
	// go without args will return an unknown command error
	assert.Error(t, err)

	err = command.Run(z, "go", "version")
	assert.NoError(t, err)
}

func TestRunOut(t *testing.T) {
	t.Parallel()
	_, err := command.RunOut(nil, "", "")
	assert.Equal(t, command.ErrZap, err)

	z := zap.NewExample().Sugar()
	_, err = command.RunOut(z, "", "")
	assert.Error(t, err)

	_, err = command.RunOut(z, "thiscommanddoesnotexist", "")
	assert.Error(t, err)

	const noArgs = ""
	_, err = command.RunOut(z, "go", noArgs)
	// go without args will return an unknown command error
	assert.Error(t, err)

	out, err := command.RunOut(z, "go", "version")
	assert.NoError(t, err)
	assert.Contains(t, string(out), "go version go1.")
}

func TestRunQuiet(t *testing.T) {
	t.Parallel()
	err := command.RunQuiet(nil, "", "")
	assert.Equal(t, command.ErrZap, err)

	err = command.RunQuiet(z(), "", "")
	assert.Error(t, err)

	err = command.RunQuiet(z(), "thiscommanddoesnotexist", "")
	assert.Error(t, err)

	const noArgs = ""
	err = command.RunQuiet(z(), "go", noArgs)
	// go without args will return an unknown command error
	assert.Error(t, err)

	err = command.RunQuiet(z(), "go", "version")
	assert.NoError(t, err)
}

func TestRunWD(t *testing.T) {
	t.Parallel()
	const noWD = ""
	err := command.RunWD(z(), "go", noWD, "")
	// go without args will return an unknown command error
	assert.Error(t, err)

	wd, err := os.Getwd()
	assert.NoError(t, err)
	err = command.RunWD(z(), "go", wd, "version")
	assert.NoError(t, err)
}
