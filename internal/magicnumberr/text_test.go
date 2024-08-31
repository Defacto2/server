package magicnumberr_test

import (
	"os"
	"testing"

	"github.com/Defacto2/server/internal/magicnumberr"
	"github.com/stretchr/testify/assert"
)

func TestASCII(t *testing.T) {
	t.Parallel()
	r, err := os.Open(uncompress(asciiFile))
	assert.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.ASCII(r))
	assert.Equal(t, magicnumberr.PlainText, magicnumberr.Find(r))

	r, err = os.Open(uncompress(txtFile))
	assert.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.ASCII(r))
	assert.Equal(t, magicnumberr.PlainText, magicnumberr.Find(r))

	r, err = os.Open(uncompress(gifFile))
	assert.NoError(t, err)
	defer r.Close()
	assert.False(t, magicnumberr.ASCII(r))

	r, err = os.Open(uncompress(badFile))
	assert.NoError(t, err)
	defer r.Close()
	assert.False(t, magicnumberr.ASCII(r))
	assert.Equal(t, magicnumberr.PlainText, magicnumberr.Find(r))

	r, err = os.Open(td(manualFile))
	assert.NoError(t, err)
	defer r.Close()

	assert.False(t, magicnumberr.ASCII(r))
	assert.Equal(t, magicnumberr.PlainText, magicnumberr.Find(r))
	sign, err := magicnumberr.Text(r)
	assert.NoError(t, err)
	assert.Equal(t, magicnumberr.PlainText, sign)

	sign, err = magicnumberr.Document(r)
	assert.NoError(t, err)
	assert.Equal(t, magicnumberr.PlainText, sign)

	sign, err = magicnumberr.Document(r)
	assert.NoError(t, err)
	assert.Equal(t, magicnumberr.PlainText, sign)
}

func TestANSI(t *testing.T) {
	t.Parallel()
	r, err := os.Open(uncompress(ansiFile))
	assert.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.Ansi(r))
	assert.Equal(t, magicnumberr.ANSIEscapeText, magicnumberr.Find(r))

	r, err = os.Open(uncompress(txtFile))
	assert.NoError(t, err)
	defer r.Close()
	assert.False(t, magicnumberr.Ansi(r))
	assert.Equal(t, magicnumberr.PlainText, magicnumberr.Find(r))

	r, err = os.Open(uncompress(gifFile))
	assert.NoError(t, err)
	defer r.Close()
	assert.False(t, magicnumberr.Ansi(r))

	r, err = os.Open(uncompress(badFile))
	assert.NoError(t, err)
	defer r.Close()
	assert.False(t, magicnumberr.Ansi(r))
	assert.Equal(t, magicnumberr.PlainText, magicnumberr.Find(r))
}

func TestRTF(t *testing.T) {
	t.Parallel()
	r, err := os.Open(uncompress(rtfFile))
	assert.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.Rtf(r))
	assert.Equal(t, magicnumberr.RichTextFormat, magicnumberr.Find(r))
}

func TestPDF(t *testing.T) {
	t.Parallel()
	r, err := os.Open(uncompress(pdfFile))
	assert.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.Pdf(r))
	assert.Equal(t, magicnumberr.PortableDocumentFormat, magicnumberr.Find(r))
}

func TestUTF16(t *testing.T) {
	t.Parallel()
	r, err := os.Open(uncompress(utf16File))
	assert.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.Utf16(r))
	assert.Equal(t, magicnumberr.UTF16Text, magicnumberr.Find(r))
}

func TestISO7(t *testing.T) {
	t.Parallel()
	r, err := os.Open(uncompress(iso7File))
	assert.NoError(t, err)
	defer r.Close()
	assert.False(t, magicnumberr.ASCII(r))
	assert.False(t, magicnumberr.Ansi(r))
	assert.True(t, magicnumberr.Txt(r))
	assert.True(t, magicnumberr.TxtLatin1(r))
	assert.True(t, magicnumberr.TxtWindows(r))
	assert.False(t, magicnumberr.Utf8(r))
	assert.False(t, magicnumberr.Utf16(r))
	assert.False(t, magicnumberr.Utf32(r))
}
