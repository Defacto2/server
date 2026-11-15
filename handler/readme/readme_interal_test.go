//nolint:testpackage
package readme

import (
	"strings"
	"testing"

	"github.com/nalgeon/be"
)

func Test_hasANSI(t *testing.T) {
	t.Parallel()

	got, err := hasANSI(nil)
	be.Err(t, err, nil)
	be.True(t, !got)

	r := strings.NewReader("a\x1b[1;cabc")
	got, err = hasANSI(r)
	be.Err(t, err, nil)
	be.True(t, got)

	r = strings.NewReader("a\x1b[Acabc")
	got, err = hasANSI(r)
	be.Err(t, err, nil)
	be.True(t, got)
}

func Test_removeControls(t *testing.T) {
	t.Parallel()

	b := []byte("a\x1b[1;cabc")
	got := removeControls(b)
	be.Equal(t, got, []byte("aabc"))
}

func Test_trimEOF(t *testing.T) {
	t.Parallel()

	wants := []byte("hello world")
	got := trimE(wants)
	be.Equal(t, got, wants)

	wants = []byte("requires._____")
	s := wants
	s = append(s, []byte("\x8a\x1a")...)
	got = trimE(s)
	be.Equal(t, got, wants)
}

func Test_viewer(t *testing.T) {
	t.Parallel()

	text := Text{}
	got := text.useViewer()
	be.True(t, !got)

	text.Platform = "textamiga"
	be.True(t, text.useViewer())
}

func Test_decodeUTF16(t *testing.T) {
	t.Parallel()

	le := []byte{0xff, 0xfe, 'H', 0x00, 'i', 0x00}
	got, err := decodeUTF16(le)
	be.Err(t, err, nil)
	be.Equal(t, got, []byte("Hi"))

	beBytes := []byte{0xfe, 0xff, 0x00, 'H', 0x00, 'i'}
	got, err = decodeUTF16(beBytes)
	be.Err(t, err, nil)
	be.Equal(t, got, []byte("Hi"))

	plain := []byte("Hi")
	got, err = decodeUTF16(plain)
	be.Err(t, err, nil)
	be.Equal(t, got, plain)
}
