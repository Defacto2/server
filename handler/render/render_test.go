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
	"github.com/aarondl/null/v8"
	"github.com/nalgeon/be"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

const txt = ".txt"

var (
	latin1 encoding.Encoding = charmap.ISO8859_1   //nolint:gochecknoglobals
	cp437  encoding.Encoding = charmap.CodePage437 //nolint:gochecknoglobals
)

func TestEncoder(t *testing.T) {
	t.Parallel()
	ec := render.Encoder(nil, nil)
	be.True(t, ec == nil)
}

func TestEncoderAmi(t *testing.T) {
	art := models.File{
		Platform: null.StringFrom("textamiga"),
	}
	ec := render.Encoder(&art, nil)
	be.Equal(t, ec, latin1)
}

func TestEncoderAppleII(t *testing.T) {
	art := models.File{
		Platform: null.StringFrom(""),
	}
	art.Section = null.StringFrom("appleii")
	ec := render.Encoder(&art, nil)
	be.Equal(t, ec, latin1)
}

func TestEncoderAtari(t *testing.T) {
	art := models.File{
		Platform: null.StringFrom(""),
	}
	art.Section = null.StringFrom("atarist")
	ec := render.Encoder(&art, nil)
	be.Equal(t, ec, latin1)
}

func TestEncoderDOS(t *testing.T) {
	art := models.File{
		Platform: null.StringFrom(""),
	}
	art.Platform = null.StringFrom("textdos")
	art.Section = null.StringFrom("")
	sr := strings.NewReader("Hello\nworld\nthis is some text.\n")
	ec := render.Encoder(&art, sr)
	be.Equal(t, ec, latin1)
}

func TestEncoderUTF8(t *testing.T) {
	art := models.File{
		Platform: null.StringFrom(""),
	}
	sr := strings.NewReader("Hello\nworld\nthis is some text. 👾\n")
	ec := render.Encoder(&art, sr)
	// Currently we cannot determine CP437 vs UTF8.
	// So the priority is to render legacy text.
	be.Equal(t, ec, cp437)
}

func TestRead(t *testing.T) {
	t.Parallel()
	r, _, err := render.Read(nil, "", "")
	be.Err(t, err)
	be.Equal(t, err, render.ErrFileModel)
	be.Equal(t, len(r), 0)

	art := models.File{
		Filename: null.StringFrom(""),
		UUID:     null.StringFrom(""),
	}
	r, _, err = render.Read(&art, "", "")
	be.Err(t, err)
	be.Equal(t, err, render.ErrFilename)
	be.Equal(t, len(r), 0)

	art.Filename = null.StringFrom(filepath.Join("testdata", "TEST.DOC"))
	r, _, err = render.Read(&art, "", "")
	be.Err(t, err)
	be.Equal(t, err, render.ErrUUID)
	be.Equal(t, len(r), 0)

	const unid = "5b4c5f6e-8a1e-11e9-9f0e-000000000000"
	art.UUID = null.StringFrom(unid)
	r, _, err = render.Read(&art, "", "")
	be.Err(t, err)
	be.Equal(t, len(r), 0)

	tmp := t.TempDir()
	err = helper.Touch(filepath.Join(tmp, unid+txt))
	be.Err(t, err, nil)
	err = helper.Touch(filepath.Join(tmp, unid))
	be.Err(t, err, nil)

	r, _, err = render.Read(&art, dir.Directory(tmp), dir.Directory(tmp))
	be.Err(t, err, nil)
	be.Equal(t, len(r), 0)

	err = os.Remove(filepath.Join(tmp, unid+txt))
	be.Err(t, err, nil)

	s := []byte("This is a test file.\n")
	i, err := helper.TouchW(filepath.Join(tmp, unid+txt), s...)
	be.Err(t, err, nil)
	l := len(s)
	be.Equal(t, i, l)
	b, _, err := render.Read(&art, dir.Directory(tmp), dir.Directory(tmp))
	be.Err(t, err, nil)
	be.True(t, len(b) > 0)
	be.Equal(t, string(b), string(s))
}

func TestViewer(t *testing.T) {
	t.Parallel()
	var art models.File
	be.True(t, !render.Viewer(&art))
	art.Platform = null.StringFrom("textamiga")
	be.True(t, render.Viewer(&art))
}

func TestNoScreenshot(t *testing.T) {
	t.Parallel()
	var art models.File
	be.True(t, render.NoScreenshot(nil, ""))
	art = models.File{}
	be.True(t, render.NoScreenshot(&art, ""))
	art = models.File{}
	art.Platform = null.StringFrom("textamiga")
	be.True(t, render.NoScreenshot(&art, ""))

	const unid = "5b4c5f6e-8a1e-11e9-9f0e-000000000000"
	art.Platform = null.StringFrom("")
	art.UUID = null.StringFrom(unid)
	name := filepath.Join(helper.TmpDir(), unid) + ".webp"
	err := helper.Touch(name)
	be.Err(t, err, nil)
	defer func() { _ = os.Remove(name) }()
	be.True(t, !render.NoScreenshot(&art, helper.TmpDir()))
}
