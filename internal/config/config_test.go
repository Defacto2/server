package config_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/zaplog"
	"github.com/nalgeon/be"
	"go.uber.org/zap"
)

func TestConfig(t *testing.T) {
	t.Parallel()
	c := config.Config{}
	s := c.Names()
	be.True(t, len(s) != 0)
	cs, err := c.Addresses()
	be.Err(t, err)
	be.True(t, cs == "")
}

func TestChecks(t *testing.T) {
	t.Parallel()
	c := config.Config{}
	err := c.Checks(nil)
	be.Err(t, err)
	err = c.LogStore()
	be.Err(t, err, nil)
	err = c.SetupLogDir(nil)
	be.Err(t, err)
}

func TestCheckDir(t *testing.T) {
	t.Parallel()
	err := config.CheckDir("", "")
	be.Err(t, err)
	err = config.CheckDir("xyz", "")
	be.Err(t, err)
}

func TestRecordCount(t *testing.T) {
	t.Parallel()
	i := config.RecordCount(t.Context(), nil)
	be.True(t, i == 0)
}

func TestSanityTmpDir(t *testing.T) {
	t.Parallel()
	var stderrBuf bytes.Buffer
	oldStdout := os.Stdout
	// defer to restore original stderr after the test
	defer func() { os.Stdout = oldStdout }()
	r, w, err := os.Pipe()
	be.Err(t, err, nil)
	os.Stdout = w
	config.SanityTmpDir()
	if err := w.Close(); err != nil {
		t.Error(err)
	}
	_, err = stderrBuf.ReadFrom(r)
	be.Err(t, err, nil)
	expectedMessage := "Temporary directory using"
	x := strings.Contains(stderrBuf.String(), expectedMessage)
	be.True(t, x)
}

func TestValidate(t *testing.T) {
	t.Parallel()
	c := config.Config{}
	err := c.HTTPPort.Check()
	be.Err(t, err, nil)
	const tooLarge = 10000000
	c.HTTPPort = tooLarge
	err = c.HTTPPort.Check()
	be.Err(t, err)
}

func TestError(t *testing.T) {
	t.Parallel()
	i, s, err := config.StringErr(nil)
	be.True(t, i == 0)
	be.Equal(t, s, "")
	be.Err(t, err, nil)
	anErr := errors.New("an error")
	i, s, err = config.StringErr(anErr)
	be.True(t, i == 500)
	x := strings.Contains(s, "internal server error")
	be.True(t, x)
	be.Err(t, err, nil)
}

func TestRepair(t *testing.T) {
	t.Parallel()
	c := config.Config{}
	err := c.Archives(t.Context(), nil)
	be.Err(t, err)
	r := config.Zip
	be.Equal(t, "zip", r.String())
	err = c.Assets(t.Context(), nil)
	be.Err(t, err)
	err = c.MagicNumbers(t.Context(), nil, nil)
	be.Err(t, err)
	err = c.Previews(t.Context(), nil, nil)
	be.Err(t, err)
	err = c.ImageDirs(nil)
	be.Err(t, err, nil)
	err = config.DownloadDir(nil, "", "", "")
	be.Err(t, err)
	err = config.RenameDownload("", "")
	be.Err(t, err)
	err = config.RemoveDir("", "", "")
	be.Err(t, err)
	err = config.RemoveDownload("", "", "", "")
	be.Err(t, err)
	err = config.RemoveImage("", "", "")
	be.Err(t, err)
}

func TestReArchive(t *testing.T) {
	t.Parallel()
	r := config.Zip
	logger, _ := zap.NewProduction()
	_ = logger.Sync()
	ctx := context.WithValue(t.Context(), zaplog.LoggerKey, logger)
	err := r.ReArchive(ctx, "", "", "")
	be.Err(t, err)
}

func TestReArchiveImplode(t *testing.T) {
	r := config.Zip
	l, _ := zap.NewProduction()
	_ = l.Sync()
	logger := l.Sugar()
	ctx := context.WithValue(t.Context(), zaplog.LoggerKey, logger)
	// test the archive that uses the defunct implode method
	src, err := filepath.Abs(filepath.Join("testdata", "IMPLODE.ZIP"))
	be.Err(t, err, nil)
	readr, err := os.Open(src)
	be.Err(t, err, nil)
	defer func() {
		_ = readr.Close()
	}()
	sign := magicnumber.Find(readr)
	be.Equal(t, magicnumber.PKWAREZipImplode, sign)
	err = r.ReArchive(ctx, src, "", "")
	be.Err(t, err)
	dst := dir.Directory(filepath.Dir(src))
	err = r.ReArchive(ctx, src, "", dst)
	be.Err(t, err)
	err = r.ReArchive(ctx, src, "newfile", dst)
	be.Err(t, err, nil)
	// test the new, re-created archive that uses the common deflate method
	name := dst.Join("newfile.zip")
	readr, err = os.Open(name)
	be.Err(t, err, nil)
	defer func() {
		err := readr.Close()
		be.Err(t, err, nil)
	}()
	sign = magicnumber.Find(readr)
	be.Equal(t, magicnumber.PKWAREZip, sign)
	defer func() {
		err := os.Remove(name)
		be.Err(t, err, nil)
	}()
}
