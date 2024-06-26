package magicnumber_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Defacto2/server/internal/magicnumber"
	"github.com/Defacto2/server/internal/magicnumber/pkzip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func tduncompress(name string) string {
	_, file, _, usable := runtime.Caller(0)
	if !usable {
		panic("runtime.Caller failed")
	}
	d := filepath.Join(filepath.Dir(file), "../..")
	x := filepath.Join(d, "assets", "testdata", "uncompress", name)
	return x
}

func TestANSIMatch(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile(td("PKZ204EX.TXT"))
	require.NoError(t, err)
	assert.False(t, magicnumber.ANSIB(b))
	b, err = os.ReadFile(tduncompress("TEST.ANS"))
	require.NoError(t, err)
	assert.True(t, magicnumber.ANSIB(b))
}

func TestArcSeaBMatcher(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile(td("PKZ204EX.TXT"))
	require.NoError(t, err)
	assert.False(t, magicnumber.ArcSeaB(b))

	match := []byte{0x1a, 0x10, 0x00, 0x00, 0x00, 0x00}
	require.NoError(t, err)
	assert.True(t, magicnumber.ArcSeaB(match))

	b, err = os.ReadFile(td("ARJ310.ARJ"))
	require.NoError(t, err)
	assert.True(t, magicnumber.ARJB(b))

	match = []byte{0xe9, 0xeb, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	assert.True(t, magicnumber.DOSComB(match))
}

func TestInterchangeMatcher(t *testing.T) {
	// Quick note, to create a test .IFF I used XnView MP.
	t.Parallel()
	b, err := os.ReadFile(td("PKZ204EX.TXT"))
	require.NoError(t, err)
	assert.False(t, magicnumber.InterchangeFFB(b))

	b, err = os.ReadFile(tduncompress("TEST.IFF"))
	require.NoError(t, err)
	assert.True(t, magicnumber.InterchangeFFB(b))
}

func TestPCXMatcher(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile(td("PKZ204EX.TXT"))
	require.NoError(t, err)
	assert.False(t, magicnumber.PCXB(b))

	b, err = os.ReadFile(tduncompress("TEST.PCX"))
	require.NoError(t, err)
	assert.True(t, magicnumber.PCXB(b))

	f, err := os.Open(tduncompress("TEST.PCX"))
	require.NoError(t, err)
	defer f.Close()
	assert.True(t, magicnumber.PCX(f))
}

func TestPNGMatcher(t *testing.T) {
	t.Parallel()
	f, err := os.Open(td("PKZ204EX.TXT"))
	require.NoError(t, err)
	defer f.Close()
	assert.False(t, magicnumber.PNG(f))

	f, err = os.Open(tduncompress("TEST.PNG"))
	require.NoError(t, err)
	defer f.Close()
	assert.True(t, magicnumber.PNG(f))
}

func TestPkzip(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile(td("PKZ204EX.TXT"))
	require.NoError(t, err)
	assert.False(t, magicnumber.PkzipB(b))

	b, err = os.ReadFile(td("PKZ204EX.ZIP"))
	require.NoError(t, err)
	assert.True(t, magicnumber.PkzipB(b))

	comps, err := magicnumber.PkzipComp(td("PKZ204EX.TXT"))
	require.Error(t, err)
	assert.Nil(t, comps)

	comps, err = magicnumber.PkzipComp(td("PKZ204EX.ZIP"))
	require.NoError(t, err)
	assert.Equal(t, pkzip.Deflated, comps[0])
	assert.Equal(t, pkzip.Stored, comps[1])

	comps, err = magicnumber.PkzipComp(td("PKZ80A1.ZIP"))
	require.NoError(t, err)
	assert.Equal(t, pkzip.Shrunk, comps[0])
	assert.Equal(t, pkzip.Stored, comps[1])

	usable, err := magicnumber.Zip(td("PKZ204EX.TXT"))
	require.Error(t, err)
	assert.False(t, usable)

	usable, err = magicnumber.Zip(td("PKZ204EX.ZIP"))
	require.NoError(t, err)
	assert.True(t, usable)

	usable, err = magicnumber.Zip(td("PKZ80A1.ZIP"))
	require.NoError(t, err)
	assert.False(t, usable)
}
