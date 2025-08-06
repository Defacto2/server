package command_test

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/nalgeon/be"
)

func logr() *slog.Logger {
	return slog.Default()
}

func testdata(name string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(file), "testdata", name)
}

func TestLookups(t *testing.T) {
	t.Parallel()
	t1 := command.Lookups()
	t2 := command.Infos()
	be.Equal(t, len(t1), len(t2))
	be.True(t, strings.Contains(t2[0], command.Arc))
}

func TestCopyFile(t *testing.T) {
	t.Parallel()
	err := command.CopyFile(nil, "", "")
	be.Err(t, err)
	td := t.TempDir()
	tmp, err := os.CreateTemp(td, "command_test")
	be.Err(t, err, nil)
	defer func() {
		err := os.Remove(tmp.Name())
		be.Err(t, err, nil)
	}()
	logr := logr()
	err = command.CopyFile(logr, "", "")
	be.Err(t, err)

	err = command.CopyFile(logr, tmp.Name(), "")
	be.Err(t, err)
	dst := tmp.Name() + ".txt"
	err = command.CopyFile(logr, tmp.Name(), dst)
	be.Err(t, err, nil)
	defer func() {
		err := os.Remove(dst)
		be.Err(t, err, nil)
	}()
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
			be.Equal(t, tt.expected, command.BaseName(tt.path))
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
			be.Equal(t, tt.expected, command.BaseNamePath(tt.path))
		})
	}
}

func TestLookCmd(t *testing.T) {
	t.Parallel()
	err := command.LookCmd("")
	be.Err(t, err)

	err = command.LookCmd("thiscommanddoesnotexist")
	be.Err(t, err)

	err = command.LookCmd("go")
	be.Err(t, err, nil)
}

func TestLookVersion(t *testing.T) {
	t.Parallel()
	err := command.LookVersion("", "", "")
	be.Err(t, err)

	err = command.LookVersion("thiscommanddoesnotexist", "", "")
	be.Err(t, err)

	err = command.LookVersion("go", "", "")
	be.Err(t, err)

	// version arg output example:
	// go version go1.16.5 linux/amd64
	err = command.LookVersion("go", "version", "go version go1.")
	be.Err(t, err, nil)
}

func TestRun(t *testing.T) {
	t.Parallel()
	err := command.Run(nil, "", "")
	be.Err(t, err)
	logr := logr()
	err = command.Run(logr, "", "")
	be.Err(t, err)

	err = command.Run(logr, "thiscommanddoesnotexist", "")
	be.Err(t, err)

	const noArgs = ""
	err = command.Run(logr, "go", noArgs)
	// go without args will return an unknown command error
	be.Err(t, err)

	err = command.Run(logr, "go", "version")
	be.Err(t, err, nil)
}

func TestRunQuiet(t *testing.T) {
	t.Parallel()
	err := command.RunQuiet("", "")
	be.Err(t, err)
	err = command.RunQuiet("thiscommanddoesnotexist", "")
	be.Err(t, err)
	const noArgs = ""
	err = command.RunQuiet("go", noArgs)
	// go without args will return an unknown command error
	be.Err(t, err)
	err = command.RunQuiet("go", "version")
	be.Err(t, err, nil)
}

func TestRunWD(t *testing.T) {
	t.Parallel()
	const noWD = ""
	err := command.RunWorkdir(logr(), "go", noWD, "")
	// go without args will return an unknown command error
	be.Err(t, err)

	wd, err := os.Getwd()
	be.Err(t, err, nil)
	err = command.RunWorkdir(logr(), "go", wd, "version")
	be.Err(t, err, nil)
}

func Test_PreviewPixels(t *testing.T) {
	t.Parallel()
	dir := command.Dirs{
		Download:  dir.Directory(t.TempDir()), // this prefixes to UUID
		Preview:   dir.Directory(t.TempDir()), // this is the output dest
		Thumbnail: dir.Directory(t.TempDir()), // this is the cropped output dest
	}
	imgs := []string{"TEST.BMP", "TEST.GIF", "TEST.JPG", "TEST.PCX", "TEST.PNG"}
	for _, name := range imgs {
		fp := testdata(name)
		err := dir.PreviewPixels(logr(), fp, "000000ABCDE")
		be.Err(t, err)
	}
	err := dir.PreviewPixels(logr(), "", "")
	be.Err(t, err)
}
