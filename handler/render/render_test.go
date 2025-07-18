package render_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/render"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
	"golang.org/x/text/encoding/charmap"
)

const txt = ".txt"

func TestEncoder(t *testing.T) {
	t.Parallel()
	ec := render.Encoder(nil, nil)
	assert.Nil(t, ec)
}

func TestEncoderAmi(t *testing.T) {
	art := models.File{
		Platform: null.StringFrom("textamiga"),
	}
	ec := render.Encoder(&art, nil)
	assert.Equal(t, ec, charmap.ISO8859_1)
}

func TestEncoderAppleII(t *testing.T) {
	art := models.File{
		Platform: null.StringFrom(""),
	}
	art.Section = null.StringFrom("appleii")
	ec := render.Encoder(&art, nil)
	assert.Equal(t, ec, charmap.ISO8859_1)
}

func TestEncoderAtari(t *testing.T) {
	art := models.File{
		Platform: null.StringFrom(""),
	}
	art.Section = null.StringFrom("atarist")
	ec := render.Encoder(&art, nil)
	assert.Equal(t, ec, charmap.ISO8859_1)
}

func TestEncoderDOS(t *testing.T) {
	art := models.File{
		Platform: null.StringFrom(""),
	}
	art.Platform = null.StringFrom("textdos")
	art.Section = null.StringFrom("")
	sr := strings.NewReader("Hello\nworld\nthis is some text.\n")
	ec := render.Encoder(&art, sr)
	assert.Equal(t, ec, charmap.ISO8859_1)
}

func TestEncoderUTF8(t *testing.T) {
	art := models.File{
		Platform: null.StringFrom(""),
	}
	sr := strings.NewReader("Hello\nworld\nthis is some text. 👾\n")
	ec := render.Encoder(&art, sr)
	// Currently we cannot determine CP437 vs UTF8.
	// So the priority is to render legacy text.
	assert.Equal(t, ec, charmap.CodePage437)
}

func TestRead(t *testing.T) {
	t.Parallel()
	r, _, err := render.Read(nil, "", "")
	require.Error(t, err)
	assert.Equal(t, err, render.ErrFileModel)
	assert.Nil(t, r)

	art := models.File{
		Filename: null.StringFrom(""),
		UUID:     null.StringFrom(""),
	}
	r, _, err = render.Read(&art, "", "")
	require.Error(t, err)
	assert.Equal(t, err, render.ErrFilename)
	assert.Nil(t, r)

	art.Filename = null.StringFrom(filepath.Join("testdata", "TEST.DOC"))
	r, _, err = render.Read(&art, "", "")
	require.Error(t, err)
	assert.Equal(t, err, render.ErrUUID)
	assert.Nil(t, r)

	const unid = "5b4c5f6e-8a1e-11e9-9f0e-000000000000"
	art.UUID = null.StringFrom(unid)
	r, _, err = render.Read(&art, "", "")
	require.Error(t, err)
	assert.Nil(t, r)

	tmp := t.TempDir()
	err = helper.Touch(filepath.Join(tmp, unid+txt))
	require.NoError(t, err)
	err = helper.Touch(filepath.Join(tmp, unid))
	require.NoError(t, err)

	r, _, err = render.Read(&art, dir.Directory(tmp), dir.Directory(tmp))
	require.NoError(t, err)
	assert.Nil(t, r)
	assert.Empty(t, r)

	err = os.Remove(filepath.Join(tmp, unid+txt))
	require.NoError(t, err)

	s := []byte("This is a test file.\n")
	i, err := helper.TouchW(filepath.Join(tmp, unid+txt), s...)
	require.NoError(t, err)
	l := len(s)
	assert.Equal(t, i, l)
	b, _, err := render.Read(&art, dir.Directory(tmp), dir.Directory(tmp))
	require.NoError(t, err)
	assert.NotNil(t, b)
	assert.Equal(t, string(b), string(s))
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
	defer func() { _ = os.Remove(name) }()
	assert.False(t, render.NoScreenshot(&art, helper.TmpDir()))
}
