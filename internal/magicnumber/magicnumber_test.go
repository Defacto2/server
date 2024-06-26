package magicnumber_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Defacto2/server/internal/magicnumber"
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

func tduncompress(name string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
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
	assert.False(t, magicnumber.ANSI(b))
	b, err = os.ReadFile(tduncompress("TEST.ANS"))
	require.NoError(t, err)
	assert.True(t, magicnumber.ANSI(b))
}

func TestArcSeaMatcher(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile(td("PKZ204EX.TXT"))
	require.NoError(t, err)
	assert.False(t, magicnumber.ArcSea(b))

	match := []byte{0x1a, 0x10, 0x00, 0x00, 0x00, 0x00}
	require.NoError(t, err)
	assert.True(t, magicnumber.ArcSea(match))

	b, err = os.ReadFile(td("ARJ310.ARJ"))
	require.NoError(t, err)
	assert.True(t, magicnumber.ARJ(b))

	match = []byte{0xe9, 0xeb, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	assert.True(t, magicnumber.DOSCom(match))
}

func TestInterchangeMatcher(t *testing.T) {
	// Quick note, to create a test .IFF I used XnView MP.
	t.Parallel()
	b, err := os.ReadFile(td("PKZ204EX.TXT"))
	require.NoError(t, err)
	assert.False(t, magicnumber.InterchangeFF(b))

	b, err = os.ReadFile(tduncompress("TEST.IFF"))
	require.NoError(t, err)
	assert.True(t, magicnumber.InterchangeFF(b))
}

func TestPCXMatcher(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile(td("PKZ204EX.TXT"))
	require.NoError(t, err)
	assert.False(t, magicnumber.PCX(b))

	b, err = os.ReadFile(tduncompress("TEST.PCX"))
	require.NoError(t, err)
	assert.True(t, magicnumber.PCX(b))
}

func TestPNGMatcher(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile(td("PKZ204EX.TXT"))
	require.NoError(t, err)
	assert.False(t, magicnumber.PNG(b))

	b, err = os.ReadFile(tduncompress("TEST.PNG"))
	require.NoError(t, err)
	assert.True(t, magicnumber.PNG(b))
}
