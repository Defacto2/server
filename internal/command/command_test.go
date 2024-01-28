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
	t1 := command.Lookups()
	t2 := command.Infos()
	assert.Equal(t, len(t1), len(t2))
	assert.Contains(t, t2[0], command.Arc)
}

func TestRemoveImgs(t *testing.T) {
	err := command.RemoveImgs("", "", "")
	assert.Nil(t, err)
	td := os.TempDir()

	tmp, err := os.CreateTemp(td, "command_test")
	assert.Nil(t, err)
	defer os.Remove(tmp.Name())

	err = os.Rename(tmp.Name(), filepath.Join(tmp.Name()+".jpg"))
	assert.Nil(t, err)

	name := strings.TrimSuffix(filepath.Base(tmp.Name()), filepath.Ext(tmp.Name()))

	err = command.RemoveImgs(name, td, "somerandomname")
	assert.Nil(t, err)
}

func TestRemoveMe(t *testing.T) {
	err := command.RemoveMe("", "")
	assert.Nil(t, err)

	td := os.TempDir()
	tmp, err := os.CreateTemp(td, "command_test")
	assert.Nil(t, err)
	defer os.Remove(tmp.Name())

	err = os.Rename(tmp.Name(), filepath.Join(tmp.Name()+".txt"))
	assert.Nil(t, err)

	name := strings.TrimSuffix(filepath.Base(tmp.Name()), filepath.Ext(tmp.Name()))
	err = command.RemoveMe(name, td)
	assert.Nil(t, err)
}

func TestCopyFile(t *testing.T) {
	err := command.CopyFile(nil, "", "")
	assert.Equal(t, command.ErrZap, err)

	td := os.TempDir()
	tmp, err := os.CreateTemp(td, "command_test")
	assert.Nil(t, err)
	defer os.Remove(tmp.Name())

	z := zap.NewExample().Sugar()
	err = command.CopyFile(z, "", "")
	assert.NotNil(t, err)

	err = command.CopyFile(z, tmp.Name(), "")
	assert.NotNil(t, err)

	dst := filepath.Join(tmp.Name() + ".txt")
	err = command.CopyFile(z, tmp.Name(), dst)
	assert.Nil(t, err)
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
	err := command.LookCmd("")
	assert.NotNil(t, err)

	err = command.LookCmd("thiscommanddoesnotexist")
	assert.NotNil(t, err)

	err = command.LookCmd("go")
	assert.Nil(t, err)
}

func TestLookVersion(t *testing.T) {
	err := command.LookVersion("", "", "")
	assert.NotNil(t, err)

	err = command.LookVersion("thiscommanddoesnotexist", "", "")
	assert.NotNil(t, err)

	err = command.LookVersion("go", "", "")
	assert.NotNil(t, err)

	// version arg output example:
	// go version go1.16.5 linux/amd64
	err = command.LookVersion("go", "version", "go version go1.")
	assert.Nil(t, err)
}

func TestRun(t *testing.T) {
	err := command.Run(nil, "", "")
	assert.Equal(t, command.ErrZap, err)

	z := zap.NewExample().Sugar()
	err = command.Run(z, "", "")
	assert.NotNil(t, err)

	err = command.Run(z, "thiscommanddoesnotexist", "")
	assert.NotNil(t, err)

	const noArgs = ""
	err = command.Run(z, "go", noArgs)
	// go without args will return an unknown command error
	assert.NotNil(t, err)

	err = command.Run(z, "go", "version")
	assert.Nil(t, err)
}

func TestRunOut(t *testing.T) {
	_, err := command.RunOut(nil, "", "")
	assert.Equal(t, command.ErrZap, err)

	z := zap.NewExample().Sugar()
	_, err = command.RunOut(z, "", "")
	assert.NotNil(t, err)

	_, err = command.RunOut(z, "thiscommanddoesnotexist", "")
	assert.NotNil(t, err)

	const noArgs = ""
	_, err = command.RunOut(z, "go", noArgs)
	// go without args will return an unknown command error
	assert.NotNil(t, err)

	out, err := command.RunOut(z, "go", "version")
	assert.Nil(t, err)
	assert.Contains(t, string(out), "go version go1.")
}

func TestRunQuiet(t *testing.T) {
	err := command.RunQuiet(nil, "", "")
	assert.Equal(t, command.ErrZap, err)

	err = command.RunQuiet(z(), "", "")
	assert.NotNil(t, err)

	err = command.RunQuiet(z(), "thiscommanddoesnotexist", "")
	assert.NotNil(t, err)

	const noArgs = ""
	err = command.RunQuiet(z(), "go", noArgs)
	// go without args will return an unknown command error
	assert.NotNil(t, err)

	err = command.RunQuiet(z(), "go", "version")
	assert.Nil(t, err)
}

func TestRunWD(t *testing.T) {
	const noWD = ""
	err := command.RunWD(z(), "go", noWD, "")
	// go without args will return an unknown command error
	assert.NotNil(t, err)

	wd, err := os.Getwd()
	assert.Nil(t, err)
	err = command.RunWD(z(), "go", wd, "version")
	assert.Nil(t, err)
}
