package magicnumber_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Defacto2/server/internal/magicnumber"
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

func TestFinds(t *testing.T) {
	t.Parallel()
	f, err := os.Open(td("PKZ204EX.TXT"))
	require.NoError(t, err)
	defer f.Close()
	sign, err := magicnumber.Find(f)
	require.NoError(t, err)
	assert.Equal(t, magicnumber.PlainText, sign)

	f, err = os.Open(tduncompress("TEST.PNG"))
	require.NoError(t, err)
	defer f.Close()
	sign, err = magicnumber.Find(f)
	require.NoError(t, err)
	assert.Equal(t, magicnumber.PortableNetworkGraphics, sign)

	f, err = os.Open(tduncompress("TEST.GIF"))
	require.NoError(t, err)
	defer f.Close()
	sign, err = magicnumber.Find(f)
	require.NoError(t, err)
	assert.Equal(t, magicnumber.GraphicsInterchangeFormat, sign)

	f, err = os.Open(tduncompress("TEST.PCX"))
	require.NoError(t, err)
	defer f.Close()
	sign, err = magicnumber.Find(f)
	require.NoError(t, err)
	assert.Equal(t, magicnumber.PersonalComputereXchange, sign)

	fmt.Println(td("TAR135.TAR"))
	f, err = os.Open(td("TAR135.TAR"))
	require.NoError(t, err)
	defer f.Close()
	sign, err = magicnumber.Find(f)
	require.NoError(t, err)
	assert.Equal(t, magicnumber.TapeARchive, sign)
}

func TestANSIMatch(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile(td("PKZ204EX.TXT"))
	require.NoError(t, err)
	assert.False(t, magicnumber.Ansi(b))
	b, err = os.ReadFile(tduncompress("TEST.ANS"))
	require.NoError(t, err)
	assert.True(t, magicnumber.Ansi(b))
}

func TestAscii(t *testing.T) {
	t.Parallel()
	p := []byte("Hello, World!")
	assert.True(t, magicnumber.Ascii(p))
	p = []byte("Hello, World!\x00")
	assert.True(t, magicnumber.Ascii(p))
	p = []byte("Hello, World!\x01")
	assert.False(t, magicnumber.Ascii(p))
	const esc = "\x1b"
	p = []byte("Hello, World!" + esc)
	assert.True(t, magicnumber.Ascii(p))

	b, err := os.ReadFile(td("PKZ204EX.TXT"))
	require.NoError(t, err)
	assert.True(t, magicnumber.Txt(b))

	b, err = os.ReadFile(td("PKZ204EX.TXT"))
	require.NoError(t, err)
	assert.True(t, magicnumber.TxtWindows(b))

	b, err = os.ReadFile(td("PKZ204EX.ZIP"))
	require.NoError(t, err)
	assert.False(t, magicnumber.Txt(b))
}

func TestTextLatin1(t *testing.T) {
	t.Parallel()
	p := []byte("Hello, World!")
	assert.True(t, magicnumber.TxtLatin1(p))
	p = []byte("Hello, World! \x92")
	assert.False(t, magicnumber.TxtLatin1(p))
	p = []byte("Hello, World! \x03")
	assert.False(t, magicnumber.TxtLatin1(p))
	b, err := os.ReadFile(td("PKZ204EX.TXT"))
	require.NoError(t, err)
	assert.True(t, magicnumber.TxtLatin1(b))
}

func TestTxtWindows(t *testing.T) {
	t.Parallel()
	p := []byte("Hello, World!")
	assert.True(t, magicnumber.TxtWindows(p))

	p = []byte("Hello, World! \x92")
	assert.True(t, magicnumber.TxtWindows(p))

	p = []byte("Hello, World! \x8f")
	assert.False(t, magicnumber.TxtWindows(p))

	p = []byte("Hello, World! \x03")
	assert.False(t, magicnumber.TxtWindows(p))

	b, err := os.ReadFile(td("PKZ204EX.TXT"))
	require.NoError(t, err)
	assert.True(t, magicnumber.TxtWindows(b))
}
