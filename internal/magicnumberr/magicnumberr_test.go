package magicnumberr_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/magicnumberr"
	"github.com/stretchr/testify/assert"
)

const (
	emptyFile = "EMPTY.ZIP"
)

func td(name string) string {
	_, file, _, usable := runtime.Caller(0)
	if !usable {
		panic("runtime.Caller failed")
	}
	d := filepath.Join(filepath.Dir(file), "../..")
	x := filepath.Join(d, "assets", "testdata", name)
	return x
}

func TestUnknowns(t *testing.T) {
	t.Parallel()
	t.Log("TestUnknowns")
	data := "some binary data"
	nr := strings.NewReader(data)
	sign, err := magicnumberr.Archive(nr)
	assert.NoError(t, err)
	assert.True(t, sign == magicnumberr.Unknown)
	assert.Equal(t, magicnumberr.Unknown, sign)
	assert.Equal(t, "binary data", sign.String())
	assert.Equal(t, "Binary data", sign.Title())

	b, sign, err := magicnumberr.MatchExt(emptyFile, nr)
	assert.NoError(t, err)
	assert.False(t, b)
	assert.True(t, sign == magicnumberr.Unknown)

	r, err := os.Open(td(emptyFile))
	assert.NoError(t, err)
	defer r.Close()
	sign = magicnumberr.Find(r)
	assert.True(t, sign == magicnumberr.ZeroByte)
	b, sign, err = magicnumberr.MatchExt(emptyFile, r)
	assert.NoError(t, err)
	assert.False(t, b)
	assert.True(t, sign == magicnumberr.ZeroByte)
}
