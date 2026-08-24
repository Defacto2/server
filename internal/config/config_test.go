package config_test

import (
	"bytes"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/nalgeon/be"
)

var ErrTest = errors.New("an error")

func TestConfig(t *testing.T) {
	t.Parallel()

	c := config.Config{}
	s := c.Names()
	be.True(t, len(s) != 0)

	got := c.Addresses(nil)
	be.Err(t, got)

	got = c.Addresses(slog.Default())
	be.Err(t, got, nil)
}

func TestChecks(t *testing.T) {
	t.Parallel()

	c := config.Config{}
	got := c.Checks(t.Context(), slog.Default())
	be.Err(t, got)

	got = c.LogStore()
	be.Err(t, got, nil)
}

func TestCheckDir(t *testing.T) {
	t.Parallel()

	got := config.CheckDir("", "")
	be.Err(t, got)

	got = config.CheckDir("xyz", "")
	be.Err(t, got)
}

func TestSanityTmpDir(t *testing.T) {
	t.Parallel()

	var stderrBuf bytes.Buffer
	oldStdout := os.Stdout
	// defer to restore original stderr after the test
	defer func() { os.Stdout = oldStdout }()

	r, w, got := os.Pipe()
	be.Err(t, got, nil)
	os.Stdout = w
	config.TmpInfo(logs.Discard())
	if err := w.Close(); err != nil {
		t.Error(err)
	}
	_, got = stderrBuf.ReadFrom(r)
	be.Err(t, got, nil)
}

func TestValidate(t *testing.T) {
	t.Parallel()

	c := config.Config{}
	got := c.HTTPPort.Check()
	be.Err(t, got, nil)

	const tooLarge = 10000
	c.HTTPPort = tooLarge
	got = c.HTTPPort.Check()
	be.Err(t, got, nil)
}

func TestRepair(t *testing.T) {
	t.Parallel()

	c := config.Config{}
	disc := logs.Discard()
	got := c.RepairArchive(t.Context(), nil, nil)
	be.Err(t, got)

	r := config.Zip
	be.Equal(t, "zip", r.String())
	got = c.Assets(t.Context(), nil, nil)
	be.Err(t, got)

	got = c.MagicNumbers(t.Context(), nil, nil)
	be.Err(t, got)

	got = c.Previews(t.Context(), nil, nil)
	be.Err(t, got)

	sl := logs.Discard()
	got = c.ImageDirs(sl)
	be.Err(t, got, nil)

	got = config.DownloadDir(nil, "", "", "")
	be.Err(t, got)

	got = config.RenameDownload(disc, "", "")
	be.Err(t, got)

	got = config.RemoveDir(disc, "", "", "")
	be.Err(t, got)

	got = config.RemoveDownload(disc, "", "", "", "")
	be.Err(t, got)

	got = config.RemoveImage(disc, "", "", "")
	be.Err(t, got)
}

func TestReArchive(t *testing.T) {
	t.Parallel()

	r := config.Zip
	got := r.RePack(t.Context(), nil, config.Rearchiving{})
	be.Err(t, got)
}

func TestReArchiveImplode(t *testing.T) {
	r := config.Zip
	ctx := t.Context()

	// test an archive that uses the defunct implode zip method
	src, err := filepath.Abs(filepath.Join("testdata", "IMPLODE.ZIP"))
	if err != nil {
		t.Error(err)
		return
	}

	f1, err := os.Open(src)
	defer f1.Close()
	if err != nil {
		t.Error(err)
		return
	}

	sign := magicnumber.Find(f1)
	be.Equal(t, magicnumber.PKWAREZipImplode, sign)

	ra0 := config.Rearchiving{}
	got := r.RePack(ctx, nil, ra0)
	be.Err(t, got)

	tmp := t.TempDir()
	dst := dir.Directory(filepath.Dir(tmp))

	ra1 := config.Rearchiving{Source: src, Destination: dst}
	got = r.RePack(ctx, nil, ra1)
	be.Err(t, got)

	const newfile = "newfile"
	ra1.UID = newfile
	sl := logs.Discard()
	got = r.RePack(ctx, sl, ra1)
	be.Err(t, got, nil)

	name := dst.Join(newfile + ".zip")
	f2, got := os.Open(name)
	defer f2.Close()
	be.Err(t, got, nil)

	sign = magicnumber.Find(f2)
	be.Equal(t, magicnumber.PKWAREZip, sign)
}
