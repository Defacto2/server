package pkzip_test

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Defacto2/server/internal/archive/pkzip"
	"github.com/Defacto2/server/internal/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func td(name string) string {
	_, file, _, usable := runtime.Caller(0)
	if !usable {
		panic("runtime.Caller failed")
	}
	d := filepath.Join(filepath.Dir(file), "../../..")
	x := filepath.Join(d, "assets", "testdata", name)
	return x
}

func TestPkzip(t *testing.T) {
	t.Parallel()

	comps, err := pkzip.Methods(td("PKZ204EX.TXT"))
	require.Error(t, err)
	assert.Nil(t, comps)

	comps, err = pkzip.Methods(td("PKZ204EX.ZIP"))
	require.NoError(t, err)
	assert.Equal(t, pkzip.Deflated, comps[1])
	assert.Equal(t, pkzip.Stored, comps[0])

	comps, err = pkzip.Methods(td("PKZ80A1.ZIP"))
	require.NoError(t, err)
	assert.Equal(t, pkzip.Shrunk, comps[1])
	assert.Equal(t, pkzip.Stored, comps[0])

	comps, err = pkzip.Methods(td("PKZ80A1.ZIP"))
	require.NoError(t, err)
	assert.Equal(t, pkzip.Shrunk.String(), comps[1].String())
	assert.Equal(t, pkzip.Stored.String(), comps[0].String())

	comps, err = pkzip.Methods(td("PKZ110EI.ZIP"))
	require.NoError(t, err)
	assert.Equal(t, "[Stored Imploded]", fmt.Sprint(comps))
	assert.False(t, comps[1].Zip())

	usable, err := pkzip.Zip(td("PKZ204EX.TXT"))
	require.Error(t, err)
	assert.False(t, usable)

	usable, err = pkzip.Zip(td("PKZ204EX.ZIP"))
	require.NoError(t, err)
	assert.True(t, usable)

	usable, err = pkzip.Zip(td("PKZ80A1.ZIP"))
	require.NoError(t, err)
	assert.False(t, usable)

	const invalid = 999
	comp := pkzip.Compression(invalid)
	assert.Equal(t, "Reserved", comp.String())
}

func TestExitStatus(t *testing.T) {
	t.Parallel()
	app, err := exec.LookPath(command.Unzip)
	require.NoError(t, err)

	err = exec.Command(app, "-T", "archive.zip").Run()
	require.Error(t, err)
	diag := pkzip.ExitStatus(err)
	assert.Equal(t, pkzip.ZipNotFound, diag)
	assert.Equal(t, "Zip file not found", diag.String())
}
