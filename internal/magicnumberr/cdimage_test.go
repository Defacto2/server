package magicnumberr_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Defacto2/server/internal/magicnumberr"
	"github.com/stretchr/testify/assert"
)

const (
	DaaFile = "uncompress.daa"
	ISOFile = "uncompress.iso"
	MdfFile = "uncompress.bin"
)

func tduimages(name string) string {
	_, file, _, usable := runtime.Caller(0)
	if !usable {
		panic("runtime.Caller failed")
	}
	d := filepath.Join(filepath.Dir(file), "../..")
	x := filepath.Join(d, "assets", "testdata", "discimages", name)
	return x
}

func TestDaa(t *testing.T) {
	t.Parallel()
	t.Log("TestDaa")
	r, err := os.Open(tduimages(DaaFile))
	assert.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.Daa(r))
	sign := magicnumberr.Find(r)
	assert.Equal(t, magicnumberr.CDPowerISO, sign)
	assert.Equal(t, "CD, PowerISO", sign.String())
	assert.Equal(t, "CD PowerISO", sign.Title())
	b, sign, err := magicnumberr.MatchExt(DaaFile, r)
	assert.NoError(t, err)
	assert.True(t, b)
	assert.Equal(t, magicnumberr.CDPowerISO, sign)
}

func TestCDISO(t *testing.T) {
	t.Parallel()
	t.Log("TestCDISO")
	r, err := os.Open(tduimages(ISOFile))
	assert.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.ISO(r))
	sign := magicnumberr.Find(r)
	assert.Equal(t, magicnumberr.CDISO9660, sign)
	assert.Equal(t, "CD, ISO 9660", sign.String())
	assert.Equal(t, "CD ISO 9660", sign.Title())
	b, sign, err := magicnumberr.MatchExt(ISOFile, r)
	assert.NoError(t, err)
	assert.True(t, b)
	assert.Equal(t, magicnumberr.CDISO9660, sign)
}

func TestMdf(t *testing.T) {
	t.Parallel()
	t.Log("TestMdf")
	r, err := os.Open(tduimages(MdfFile))
	assert.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.Mdf(r))
	sign := magicnumberr.Find(r)
	assert.Equal(t, magicnumberr.CDAlcohol120, sign)
	b, sign, err := magicnumberr.MatchExt(DaaFile, r)
	assert.NoError(t, err)
	assert.False(t, b)
	assert.Equal(t, magicnumberr.CDAlcohol120, sign)
}
