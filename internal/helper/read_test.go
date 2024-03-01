package helper_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Defacto2/server/internal/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func td(name string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("runtime.Caller failed")
	}
	d := filepath.Join(filepath.Dir(file), "../..")
	x := filepath.Join(d, "assets", "testdata", name)
	return x
}

func TestLines(t *testing.T) {
	i, err := helper.Lines("")
	require.Error(t, err)
	assert.Equal(t, 0, i)

	i, err = helper.Lines("nosuchfile")
	require.Error(t, err)
	assert.Equal(t, 0, i)

	i, err = helper.Lines(td(""))
	require.Error(t, err)
	assert.Equal(t, 0, i)

	i, err = helper.Lines(td("TEST.BMP"))
	require.Error(t, err)
	assert.Equal(t, 0, i)

	i, err = helper.Lines(td("PKZ80A1.TXT"))
	require.NoError(t, err)
	assert.Equal(t, 175, i)
}
