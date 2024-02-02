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

	p = []byte("Hello world!")
	e = helper.DetermineEncoding(p)
	assert.Equal(t, charmap.ISO8859_1, e)

	p = []byte("Hello world! 👾")
	e = helper.DetermineEncoding(p)
	assert.Nil(t, e)

	p = []byte("")
	p = append(p, 0x1b)
	p = append(p, []byte("[31mHelloWorld")...)
	e = helper.DetermineEncoding(p)
	assert.Equal(t, charmap.ISO8859_1, e)

	p = []byte("\nHello world!\n")
	e = helper.DetermineEncoding(p)
	assert.Equal(t, charmap.ISO8859_1, e)

	p = []byte("")
	p = append(p, 0xb2)
	p = append(p, []byte(" Hello world! ")...)
	p = append(p, 0xb2)
	e = helper.DetermineEncoding(p)
	assert.Equal(t, charmap.CodePage437, e)

	p = []byte("")
	p = append(p, 0x0D, 0x0E) // CP437 ♪ ♫
	p = append(p, []byte(" aah bah cah")...)
	e = helper.DetermineEncoding(p)
	assert.Equal(t, charmap.CodePage437, e)

	const house = 0x7f
	p = []byte("")
	p = append(p, house)
	p = append(p, []byte(" a DOS house glyph ")...)
	e = helper.DetermineEncoding(p)
	assert.Equal(t, charmap.CodePage437, e)

	const line = 0xc4
	p = []byte("")
	p = append(p, line)
	p = append(p, []byte(" a DOS line glyph ")...)
	e = helper.DetermineEncoding(p)
	assert.Equal(t, charmap.CodePage437, e)
}
