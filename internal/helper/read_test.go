package helper_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Defacto2/server/internal/helper"
	"github.com/stretchr/testify/assert"
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
	assert.Error(t, err)
	assert.Equal(t, 0, i)

	i, err = helper.Lines("nosuchfile")
	assert.Error(t, err)
	assert.Equal(t, 0, i)

	i, err = helper.Lines(td(""))
	assert.Error(t, err)
	assert.Equal(t, 0, i)

	i, err = helper.Lines(td("TEST.BMP"))
	assert.Error(t, err)
	assert.Equal(t, 0, i)

	i, err = helper.Lines(td("PKZ80A1.TXT"))
	assert.NoError(t, err)
	assert.Equal(t, 175, i)
}
