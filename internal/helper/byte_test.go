package helper_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/charmap"
)

func TestDetermineEncoding(t *testing.T) {
	e := helper.DetermineEncoding(nil)
	assert.Nil(t, e)

	sr := strings.NewReader("Hello world!")
	e = helper.DetermineEncoding(sr)
	assert.Equal(t, charmap.ISO8859_1, e)

	sr = strings.NewReader("Hello world! 👾")
	e = helper.DetermineEncoding(sr)
	assert.Nil(t, e)

	p := []byte("")
	p = append(p, 0x1b)
	p = append(p, []byte("[31mHelloWorld")...)
	br := bytes.NewReader(p)
	e = helper.DetermineEncoding(br)
	assert.Equal(t, charmap.ISO8859_1, e)

	sr = strings.NewReader("\nHello world!\n")
	e = helper.DetermineEncoding(sr)
	assert.Equal(t, charmap.ISO8859_1, e)

	p = []byte("")
	p = append(p, 0xb2)
	p = append(p, []byte(" Hello world! ")...)
	p = append(p, 0xb2)
	br = bytes.NewReader(p)
	e = helper.DetermineEncoding(br)
	assert.Equal(t, charmap.CodePage437, e)

	p = []byte("")
	p = append(p, 0x0D, 0x0E) // CP437 ♪ ♫
	p = append(p, []byte(" aah bah cah")...)
	br = bytes.NewReader(p)
	e = helper.DetermineEncoding(br)
	assert.Equal(t, charmap.CodePage437, e)

	const house = 0x7f
	p = []byte("")
	p = append(p, house)
	p = append(p, []byte(" a DOS house glyph ")...)
	br = bytes.NewReader(p)
	e = helper.DetermineEncoding(br)
	assert.Equal(t, charmap.CodePage437, e)

	const line = 0xc4
	p = []byte("")
	p = append(p, line)
	p = append(p, []byte(" a DOS line glyph ")...)
	br = bytes.NewReader(p)
	e = helper.DetermineEncoding(br)
	assert.Equal(t, charmap.CodePage437, e)
}

func TestCookieStore(t *testing.T) {
	t.Parallel()
	b, err := helper.CookieStore("")
	require.NoError(t, err)
	assert.Len(t, b, 32)

	const key = "my-secret-key"
	b, err = helper.CookieStore(key)
	require.NoError(t, err)
	assert.Len(t, b, len(key))
}
