package render_test

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

const (
	txt = ".txt"
)

func TestEncoder(t *testing.T) {
	t.Parallel()
	ec := render.Encoder(nil, nil)
	assert.Nil(t, ec)

	art := models.File{
		Platform: null.StringFrom("textamiga"),
	}
	ec = render.Encoder(&art, nil)
	assert.Equal(t, ec, charmap.ISO8859_1)

	art = models.File{
		Platform: null.StringFrom(""),
	}
	art.Section = null.StringFrom("appleii")
	ec = render.Encoder(&art, nil)
	assert.Equal(t, ec, charmap.ISO8859_1)

	art.Section = null.StringFrom("atarist")
	ec = render.Encoder(&art, nil)
	assert.Equal(t, ec, charmap.ISO8859_1)

	art.Platform = null.StringFrom("textdos")
	art.Section = null.StringFrom("")
	sr := strings.NewReader("Hello\nworld\nthis is some text.\n")
	ec = render.Encoder(&art, sr)
	assert.Equal(t, ec, charmap.ISO8859_1)

	sr = strings.NewReader("Hello\nworld\nthis is some text. 👾\n")
	ec = render.Encoder(&art, sr)
	assert.Equal(t, ec, unicode.UTF8)
}

func TestRead(t *testing.T) {
	t.Parallel()
	r, err := render.Read(nil, "", "")
	require.Error(t, err)
	assert.Equal(t, err, render.ErrFileModel)
	assert.Nil(t, r)

	art := models.File{
		Filename: null.StringFrom(""),
		UUID:     null.StringFrom(""),
	}
	r, err = render.Read(&art, "", "")
	require.Error(t, err)
	assert.Equal(t, err, render.ErrFilename)
	assert.Nil(t, r)

	art.Filename = null.StringFrom("../testdata/TEST.DOC")
	r, err = render.Read(&art, "", "")
	require.Error(t, err)
	assert.Equal(t, err, render.ErrUUID)
	assert.Nil(t, r)

	const unid = "5b4c5f6e-8a1e-11e9-9f0e-000000000000"
	art.UUID = null.StringFrom(unid)
	r, err = render.Read(&art, "", "")
	require.Error(t, err)
	assert.Nil(t, r)

	dir, err := os.MkdirTemp(helper.TmpDir(), unid)
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	err = helper.Touch(filepath.Join(dir, unid+txt))
	require.NoError(t, err)
	err = helper.Touch(filepath.Join(dir, unid))
	require.NoError(t, err)

	r, err = render.Read(&art, dir, dir)
	require.NoError(t, err)
	assert.Nil(t, r)
	assert.Empty(t, r)

	err = os.Remove(filepath.Join(dir, unid+txt))
	require.NoError(t, err)

	s := []byte("This is a test file.\n")
	i, err := helper.TouchW(filepath.Join(dir, unid+txt), s...)
	require.NoError(t, err)
	l := len(s)
	assert.Equal(t, i, l)
	b, err := render.Read(&art, dir, dir)
	require.NoError(t, err)
	assert.NotNil(t, b)
	assert.Equal(t, string(b), string(s))
}

func stringToUTF16(s string) []uint16 {
	return utf16.Encode([]rune(s))
}

func uint16ArrayToByteArray(nums []uint16) []byte {
	bytes := make([]byte, len(nums)*2)
	for i, num := range nums {
		binary.LittleEndian.PutUint16(bytes[i*2:], num)
	}
	return bytes
}

func TestUTF16(t *testing.T) {
	t.Parallel()
	assert.False(t, render.UTF16(nil))

	r := bytes.NewReader(nil)
	assert.False(t, render.UTF16(r))

	b := []byte{0xff, 0xfe, 0x00, 0x00, 0x00, 0x00}
	r = bytes.NewReader(b)
	assert.True(t, render.UTF16(r))

	b = []byte{0x00, 0x00, 0xfe, 0xff, 0x00, 0x00}
	r = bytes.NewReader(b)
	assert.False(t, render.UTF16(r))

	b = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	r = bytes.NewReader(b)
	assert.False(t, render.UTF16(r))

	s := "😀 some unicode text 😀"
	u := stringToUTF16(s)
	u = append([]uint16{0xFEFF}, u...)
	b = uint16ArrayToByteArray(u)
	r = bytes.NewReader(b)
	assert.True(t, render.UTF16(r))
}

func TestViewer(t *testing.T) {
	t.Parallel()
	var art models.File
	assert.False(t, render.Viewer(&art))
	art.Platform = null.StringFrom("textamiga")
	assert.True(t, render.Viewer(&art))
}

func TestNoScreenshot(t *testing.T) {
	t.Parallel()
	var art models.File
	assert.True(t, render.NoScreenshot(nil, ""))
	art = models.File{}
	assert.True(t, render.NoScreenshot(&art, ""))
	art = models.File{}
	art.Platform = null.StringFrom("textamiga")
	assert.True(t, render.NoScreenshot(&art, ""))

	const unid = "5b4c5f6e-8a1e-11e9-9f0e-000000000000"
	art.Platform = null.StringFrom("")
	art.UUID = null.StringFrom(unid)
	name := filepath.Join(helper.TmpDir(), unid) + ".webp"
	err := helper.Touch(name)
	require.NoError(t, err)
	defer os.Remove(name)
	assert.False(t, render.NoScreenshot(&art, helper.TmpDir()))
}
