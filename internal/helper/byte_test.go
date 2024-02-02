package helper_test

import (
	"testing"

	"github.com/Defacto2/server/internal/helper"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/encoding/charmap"
)

func TestDetermineEncoding(t *testing.T) {
	p := []byte{}

	e := helper.DetermineEncoding(p)
	assert.Equal(t, charmap.ISO8859_1, e)

	p = nil
	p = []byte("Hello world!")
	e = helper.DetermineEncoding(p)
	assert.Equal(t, charmap.ISO8859_1, e)

	p = nil
	p = []byte("Hello world! 👾")
	e = helper.DetermineEncoding(p)
	assert.Nil(t, e)

	p = nil
	p = append(p, 0x1b)
	p = append(p, []byte("[31mHelloWorld")...)
	e = helper.DetermineEncoding(p)
	assert.Equal(t, charmap.ISO8859_1, e)

	p = nil
	p = []byte("\nHello world!\n")
	e = helper.DetermineEncoding(p)
	assert.Equal(t, charmap.ISO8859_1, e)

	p = nil
	p = append(p, 0xb2)
	p = append(p, []byte(" Hello world! ")...)
	p = append(p, 0xb2)
	e = helper.DetermineEncoding(p)
	assert.Equal(t, charmap.CodePage437, e)

	p = nil
	p = append(p, 0x0D, 0x0E) // CP437 ♪ ♫
	p = append(p, []byte(" lah lah lah")...)
	e = helper.DetermineEncoding(p)
	assert.Equal(t, charmap.CodePage437, e)

	p = nil
	const house = 0x7f
	p = append(p, house)
	p = append(p, []byte(" a DOS house glyph ")...)
	e = helper.DetermineEncoding(p)
	assert.Equal(t, charmap.CodePage437, e)

	p = nil
	const line = 0xc4
	p = append(p, line)
	p = append(p, []byte(" a DOS line glyph ")...)
	e = helper.DetermineEncoding(p)
	assert.Equal(t, charmap.CodePage437, e)
}
