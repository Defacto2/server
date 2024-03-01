package render_test

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
	"unicode/utf16"

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
	"golang.org/x/text/encoding/charmap"
)

const (
	txt = ".txt"
)

func TestEncoder(t *testing.T) {
	t.Parallel()
	ec := render.Encoder(nil)
	assert.Nil(t, ec)

	res := models.File{
		Platform: null.StringFrom("textamiga"),
	}
	ec = render.Encoder(&res)
	assert.Equal(t, ec, charmap.ISO8859_1)

	res = models.File{
		Platform: null.StringFrom(""),
	}
	res.Section = null.StringFrom("appleii")
	ec = render.Encoder(&res)
	assert.Equal(t, ec, charmap.ISO8859_1)

	res.Section = null.StringFrom("atarist")
	ec = render.Encoder(&res)
	assert.Equal(t, ec, charmap.ISO8859_1)

	res.Platform = null.StringFrom("textdos")
	res.Section = null.StringFrom("")
	b := []byte("Hello\nworld\nthis is some text.\n")
	ec = render.Encoder(&res, b...)
	assert.Equal(t, ec, charmap.ISO8859_1)

	b = []byte("Hello\nworld\nthis is some text. 👾\n")
	ec = render.Encoder(&res, b...)
	assert.Nil(t, ec)
}

func TestRead(t *testing.T) {
	t.Parallel()
	b, err := render.Read(nil, "")
	require.Error(t, err)
	assert.Equal(t, err, render.ErrFileModel)
	assert.Nil(t, b)

	res := models.File{
		Filename: null.StringFrom(""),
		UUID:     null.StringFrom(""),
	}
	b, err = render.Read(&res, "")
	require.Error(t, err)
	assert.Equal(t, err, render.ErrFilename)
	assert.Nil(t, b)

	res.Filename = null.StringFrom("../testdata/TEST.DOC")
	b, err = render.Read(&res, "")
	require.Error(t, err)
	assert.Equal(t, err, render.ErrUUID)
	assert.Nil(t, b)

	const uuid = "5b4c5f6e-8a1e-11e9-9f0e-000000000000"
	res.UUID = null.StringFrom(uuid)
	b, err = render.Read(&res, "")
	require.Error(t, err)
	assert.Nil(t, b)

	dir, err := os.MkdirTemp(os.TempDir(), uuid)
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	err = helper.Touch(filepath.Join(dir, uuid+txt))
	require.NoError(t, err)
	err = helper.Touch(filepath.Join(dir, uuid))
	require.NoError(t, err)

	b, err = render.Read(&res, dir)
	require.NoError(t, err)
	assert.Nil(t, b)
	assert.Empty(t, b)

	err = os.Remove(filepath.Join(dir, uuid+txt))
	require.NoError(t, err)

	s := []byte("This is a test file.\n")
	i, err := helper.TouchW(filepath.Join(dir, uuid+txt), s...)
	require.NoError(t, err)
	assert.Len(t, i, len(s))
	b, err = render.Read(&res, dir)
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

func TestIsUTF16(t *testing.T) {
	t.Parallel()
	b := []byte{0xff, 0xfe, 0x00, 0x00, 0x00, 0x00}
	assert.True(t, render.IsUTF16(b))

	b = []byte{0x00, 0x00, 0xfe, 0xff, 0x00, 0x00}
	assert.False(t, render.IsUTF16(b))

	b = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	assert.False(t, render.IsUTF16(b))

	s := "😀 some unicode text 😀"
	u := stringToUTF16(s)
	u = append([]uint16{0xFEFF}, u...)
	b = uint16ArrayToByteArray(u)
	assert.True(t, render.IsUTF16(b))
}

func TestViewer(t *testing.T) {
	t.Parallel()
	var res models.File
	assert.False(t, render.Viewer(&res))
	res.Platform = null.StringFrom("textamiga")
	assert.True(t, render.Viewer(&res))
}

func TestNoScreenshot(t *testing.T) {
	t.Parallel()
	var res models.File
	assert.True(t, render.NoScreenshot(nil, ""))
	res = models.File{}
	assert.True(t, render.NoScreenshot(&res, ""))
	res = models.File{}
	res.Platform = null.StringFrom("textamiga")
	assert.True(t, render.NoScreenshot(&res, ""))

	const uuid = "5b4c5f6e-8a1e-11e9-9f0e-000000000000"
	res.Platform = null.StringFrom("")
	res.UUID = null.StringFrom(uuid)
	name := filepath.Join(os.TempDir(), uuid) + ".webp"
	err := helper.Touch(name)
	require.NoError(t, err)
	defer os.Remove(name)
	assert.False(t, render.NoScreenshot(&res, os.TempDir()))
}
