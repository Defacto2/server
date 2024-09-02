package command_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func logr() *zap.SugaredLogger {
	return zap.NewExample().Sugar()
}

func tduncompress(name string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("runtime.Caller failed")
	}
	d := filepath.Join(filepath.Dir(file), "../..")
	x := filepath.Join(d, "assets", "testdata", "uncompress", name)
	return x
}

func TestLookups(t *testing.T) {
	t.Parallel()
	t1 := command.Lookups()
	t2 := command.Infos()
	assert.Equal(t, len(t1), len(t2))
	assert.Contains(t, t2[0], command.Arc)
}

func TestCopyFile(t *testing.T) {
	t.Parallel()
	err := command.CopyFile(nil, "", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "no such file or directory")

	td := helper.TmpDir()
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
	require.Error(t, err)
	require.ErrorContains(t, err, "executable file not found in $PATH")

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

func TestRunQuiet(t *testing.T) {
	t.Parallel()
	err := command.RunQuiet("", "")
	require.Error(t, err)
	err = command.RunQuiet("thiscommanddoesnotexist", "")
	require.Error(t, err)
	const noArgs = ""
	err = command.RunQuiet("go", noArgs)
	// go without args will return an unknown command error
	require.Error(t, err)
	err = command.RunQuiet("go", "version")
	require.NoError(t, err)
}

func TestRunWD(t *testing.T) {
	t.Parallel()
	const noWD = ""
	err := command.RunWorkdir(logr(), "go", noWD, "")
	// go without args will return an unknown command error
	require.Error(t, err)

	wd, err := os.Getwd()
	require.NoError(t, err)
	err = command.RunWorkdir(logr(), "go", wd, "version")
	require.NoError(t, err)
}

func Test_PreviewPixels(t *testing.T) {
	t.Parallel()
	prev, err := os.MkdirTemp(helper.TmpDir(), "preview")
	require.NoError(t, err)
	thumb, err := os.MkdirTemp(helper.TmpDir(), "thumb")
	require.NoError(t, err)
	dl, err := os.MkdirTemp(helper.TmpDir(), "download")
	require.NoError(t, err)
	dir := command.Dirs{
		Download:  dl,    // this prefixes to UUID
		Preview:   prev,  // this is the output dest
		Thumbnail: thumb, // this is the cropped output dest
	}
	imgs := []string{"TEST.BMP", "TEST.GIF", "TEST.JPG", "TEST.PCX", "TEST.PNG"}
	for _, name := range imgs {
		fp := tduncompress(name)
		err = dir.PreviewPixels(logr(), fp, "000000ABCDE")
		require.NoError(t, err)
	}

	err = dir.PreviewPixels(logr(), "", "")
	require.Error(t, err)
}
