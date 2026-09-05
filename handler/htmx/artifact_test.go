package htmx_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/htmx"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

const (
	unid = testutil.UID
)

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want error
	}{
		{
			name: "absolute path", path: "/absolute/path", want: htmx.ErrPath,
		},
		{
			name: "clean path", path: "relative/path", want: nil,
		},
		{
			name: "clean path", path: "relative/path/", want: nil,
		},
		{
			name: "unclean path 1", path: "relative/../path", want: htmx.ErrPath,
		},
		{
			name: "unclean path 2", path: "./relative/path", want: htmx.ErrPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := htmx.Validate(tt.path)
			be.Err(t, err, tt.want)
		})
	}
}

func TestPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		unid     string
		path     string
		wantUnid string
		wantName string
		wantErr  bool
	}{
		{
			name:     "valid unid and path",
			unid:     unid,
			path:     "relative/path",
			wantUnid: unid,
			wantName: "relative/path",
			wantErr:  false,
		},
		{
			name:     "invalid unid",
			unid:     "invalid-unid",
			path:     "relative/path",
			wantUnid: "",
			wantName: "",
			wantErr:  true,
		},
		{
			name:     "invalid path",
			unid:     unid,
			path:     "/absolute/path",
			wantUnid: "",
			wantName: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := testutil.NewContext(t, "")
			c.SetPathValues(echo.PathValues{
				{Name: "unid", Value: tt.unid},
				{Name: "path", Value: tt.path},
			})

			gotUnid, gotName, err := htmx.Paths(c)
			got := (err != nil)
			be.Equal(t, got, tt.wantErr)
			be.Equal(t, tt.wantUnid, gotUnid)
			be.Equal(t, tt.wantName, gotName)
		})
	}
}

func TestUUID(t *testing.T) {
	t.Parallel()

	c := testutil.NewInput(t, "", "unid", unid)
	_, err := htmx.UUID(c)
	be.Err(t, err)
}

// TestTxFilename also tests CommitSanitize and Key.
func TestTxFilename(t *testing.T) {
	t.Parallel()

	tx := testutil.Tx(t)

	c := testutil.NewInput(t, "", htmx.EditorKey, "1")
	got := htmx.TxFilename(c, tx)
	be.Err(t, got, nil)
}

// TestReadmeOff also tests CommitNotOn and KeyParam.
func TestReadmeOff(t *testing.T) {
	t.Parallel()

	tx := testutil.Tx(t)

	formInputs := testutil.Input{
		htmx.EditorKey: "1",
	}
	pathValues := echo.PathValues{
		{Name: "id", Value: "1"},
	}
	c := testutil.NewInputsPath(t, "", formInputs, pathValues)
	got := htmx.TxReadmeOff(c, tx)
	be.Err(t, got, nil)
}

// TestEmulateXMS also tests CommitOn.
func TestEmulateXMS(t *testing.T) {
	t.Parallel()

	tx := testutil.Tx(t)

	formInputs := testutil.Input{
		htmx.EditorKey:    "1",
		"emulate-ram-xms": "on",
	}
	pathValues := echo.PathValues{
		{Name: "id", Value: "1"},
	}
	c := testutil.NewInputsPath(t, "", formInputs, pathValues)
	got := htmx.TxEmulateXMS(c, tx)
	be.Err(t, got, nil)
}

// TestTitle also tests CommitStr.
func TestTitle(t *testing.T) {
	t.Parallel()

	tx := testutil.Tx(t)
	c := testutil.NewInput(t, "", htmx.EditorKey, "1")
	got := htmx.TxTitle(c, tx)
	be.Err(t, got, nil)
}

// TxEmulateCPU also tests CommitStrKey.
func TestEmulateCPU(t *testing.T) {
	t.Parallel()

	tx := testutil.Tx(t)

	formInputs := testutil.Input{
		htmx.EditorKey: "1",
		"emulate-cpu":  "8086",
	}
	pathValues := echo.PathValues{
		{Name: "id", Value: "1"},
	}
	c := testutil.NewInputsPath(t, "", formInputs, pathValues)
	got := htmx.TxEmulateCPU(c, tx)
	be.Err(t, got, nil)
}

func TestEmulateRunProg(t *testing.T) {
	t.Parallel()

	tx := testutil.Tx(t)

	formInputs := testutil.Input{
		"emulate-run-program": "RUN.EXE",
	}
	pathValues := echo.PathValues{
		{Name: "id", Value: "1"},
	}

	c := testutil.NewInputsPath(t, "", formInputs, pathValues)
	got := htmx.TxEmulateRunProg(c, tx)
	be.Err(t, got, nil)
}

func TestHTMLLinkTo(t *testing.T) {
	t.Parallel()

	const prefix = "artifact-editor-"
	formInputs := testutil.Input{
		htmx.EditorKey:     "1",
		prefix + "youtube": testutil.YT,
	}
	c := testutil.NewInputs(t, "", formInputs)
	got := htmx.HTMLLinkTo(c, nil)
	be.Err(t, got, nil)
}

func TestLinksUndo(t *testing.T) {
	t.Parallel()

	tx := testutil.Tx(t)

	const prefix = "artifact-editor-"
	formInputs := testutil.Input{
		htmx.EditorKey:        "1",
		prefix + "youtubeval": testutil.YT,
	}
	c := testutil.NewInputs(t, "", formInputs)
	got := htmx.TxLinksUndo(c, tx)
	be.Err(t, got, nil)
}

func TestCreditUndo(t *testing.T) {
	t.Parallel()

	tx := testutil.Tx(t)

	const prefix = "artifact-editor-credits"
	formInputs := testutil.Input{
		htmx.EditorKey:   "1",
		prefix + "text":  "abcde",
		prefix + "-undo": "qwerty;;;",
	}
	c := testutil.NewInputs(t, "", formInputs)
	got := htmx.TxCreditUndo(c, tx)
	be.Err(t, got, nil)
}

func TestTxTags(t *testing.T) {
	t.Parallel()

	tx := testutil.Tx(t)
	sl := logs.Default()

	const prefix = "artifact-editor-"
	formInputs0 := testutil.Input{
		htmx.EditorKey:             "1",
		prefix + "categories":      "nfo",
		prefix + "operatingsystem": "text",
	}
	c := testutil.NewInputs(t, "", formInputs0)
	got := htmx.TxTags(sl, c, tx)
	be.Err(t, got, nil)

	formInputs1 := testutil.Input{
		htmx.EditorKey:             "1",
		prefix + "categories":      "bbs",
		prefix + "operatingsystem": "dos",
	}
	c = testutil.NewInputs(t, "", formInputs1)
	got = htmx.TxTags(sl, c, tx)
	be.Err(t, got, nil)
}

func TestTxPublic(t *testing.T) {
	t.Parallel()

	tx := testutil.Tx(t)
	c := testutil.NewInput(t, "", htmx.EditorKey, "1")
	got := htmx.TxPublic(c, tx, true)
	be.Err(t, got, nil)
}

func TestTxPublicKey(t *testing.T) {
	t.Parallel()

	tx := testutil.Tx(t)
	c := testutil.NewContext(t, "")
	got := htmx.TxPublicByKey(c, tx, "1", true)
	be.Err(t, got, nil)
}

func TestYouTube(t *testing.T) {
	t.Parallel()

	tx := testutil.Tx(t)

	const prefix = "artifact-editor-youtube"
	formInputs := testutil.Input{
		htmx.EditorKey:  "1",
		prefix + "text": testutil.YT,
	}
	c := testutil.NewInputs(t, "", formInputs)
	got := htmx.TxYouTube(c, tx)
	be.Err(t, got, nil)
}

func TestReleasers(t *testing.T) {
	t.Parallel()

	tx := testutil.Tx(t)

	const prefix = "artifact-editor-releaser"
	formInputs := testutil.Input{
		htmx.EditorKey: "1",
		prefix + "1":   "defactoweb",
		prefix + "2":   "defacto2",
	}
	c := testutil.NewInputs(t, "", formInputs)
	got := htmx.TxReleasers(c, tx)
	be.Err(t, got, nil)
}

func TestTxYMD(t *testing.T) {
	t.Parallel()

	tx := testutil.Tx(t)

	const prefix = "artifact-editor-"
	formInputs := testutil.Input{
		htmx.EditorKey:   "1",
		prefix + "year":  "2026",
		prefix + "month": "9",
		prefix + "day":   "5",
	}
	c := testutil.NewInputs(t, "", formInputs)
	got := htmx.TxYMD(c, tx)
	be.Err(t, got, nil)
}

func TestTxYMDUndo(t *testing.T) {
	t.Parallel()

	tx := testutil.Tx(t)

	const prefix = "artifact-editor-"
	formInputs := testutil.Input{
		htmx.EditorKey:        "1",
		prefix + "year":       "2026",
		prefix + "month":      "9",
		prefix + "day":        "5",
		prefix + "date-undos": "2026-9-4",
	}
	c := testutil.NewInputs(t, "", formInputs)
	got := htmx.TxYMDUndo(c, tx)
	be.Err(t, got, nil)
}

func TestCopy(t *testing.T) {
	t.Parallel()

	fid := htmx.FileID
	txt := htmx.Text
	hlp := htmx.Helper

	be.Equal(t, fid.Ext(), ".diz")
	be.Equal(t, txt.Ext(), ".txt")
	be.Equal(t, hlp.Ext(), ".hlp")
	be.Equal(t, fid.String(), "DIZ")
}

func TestMkCopy(t *testing.T) {
	t.Parallel()
	// example params: htmx/copy/:unid/:path

	const (
		fname = "filename.txt"
	)
	tmp := t.TempDir()

	pathValues := echo.PathValues{
		{Name: "unid", Value: unid},
		{Name: "path", Value: fname},
	}

	cp := htmx.FileID
	dirs := command.Dirs{
		Extra: dir.Directory(tmp),
	}
	c := testutil.NewPath(t, "", pathValues)
	got := cp.MkCopy(c, dirs)
	be.Err(t, got, nil)

	// mkcopy will exit early without a valid file with content to stat
	stale, err := dir.MkdirStale(unid)
	be.Err(t, err, nil)

	r, err := os.OpenRoot(stale)
	be.Err(t, err, nil)

	b := []byte("hello world")
	n, err := helper.TouchWR(r, fname, b...)
	be.Err(t, err, nil)
	be.Equal(t, n, len(b))

	got = cp.MkCopy(c, dirs)
	be.Err(t, got, nil)

	t.Cleanup(func() {
		r, err := os.OpenRoot(os.TempDir())
		if err != nil {
			t.Log(err)
			return
		}
		if err = r.RemoveAll(filepath.Base(stale)); err != nil {
			t.Log(err)
		}
	})
}

func TestFSCopyReadme(t *testing.T) {
	t.Parallel()

	const (
		fname = "readme.txt"
	)
	tmp := t.TempDir()

	pathValues := echo.PathValues{
		{Name: "unid", Value: unid},
		{Name: "path", Value: fname},
	}

	sl := logs.Discard()
	dirs := command.Dirs{
		Extra:     dir.Directory(tmp),
		Preview:   dir.Directory(tmp),
		Thumbnail: dir.Directory(tmp),
	}
	c := testutil.NewPath(t, "", pathValues)

	stale, err := dir.MkdirStale(unid)
	be.Err(t, err, nil)

	r, err := os.OpenRoot(stale)
	be.Err(t, err, nil)

	b := []byte("hello world")
	n, err := helper.TouchWR(r, fname, b...)
	be.Err(t, err, nil)
	be.Equal(t, n, len(b))

	got := htmx.FSCopyReadme(sl, c, dirs)
	be.Err(t, got, nil)

	t.Cleanup(func() {
		r, err := os.OpenRoot(os.TempDir())
		if err != nil {
			t.Log(err)
			return
		}
		if err = r.RemoveAll(filepath.Base(stale)); err != nil {
			t.Log(err)
		}
	})
}

func TestFSImageManipulation(t *testing.T) {
	t.Parallel()

	pathValues := echo.PathValues{
		{Name: "unid", Value: unid},
	}

	tmp := t.TempDir()
	sl := logs.Discard()
	dirs := command.Dirs{
		Extra:     dir.Directory(tmp),
		Preview:   dir.Directory(tmp),
		Thumbnail: dir.Directory(tmp),
	}

	dest := filepath.Join(tmp, unid+".png")
	testutil.CopyPNG(t, dest)
	st, err := os.Stat(dest)
	be.Err(t, err, nil)
	be.Equal(t, st.Size(), testutil.SCREENPNG)

	c := testutil.NewPath(t, "", pathValues)
	crop := command.OneTwo
	got := htmx.FSCrop(sl, c, crop, dirs)
	be.Err(t, got, nil)

	c = testutil.NewPath(t, "", pathValues)
	align := command.Left
	got = htmx.FSAlign(sl, c, align, dirs)
	be.Err(t, got, nil)

	c = testutil.NewPath(t, "", pathValues)
	thumb := command.Pixel
	got = htmx.FSThumb(sl, c, thumb, dirs)
	be.Err(t, got, nil)

	got = htmx.FSPixelate(sl, c, dirs.Preview)
	be.Err(t, got, nil)

	got = htmx.FSRemoveImages(c, dirs.Preview)
	be.Err(t, got, nil)
}

func TestFSUseReadme(t *testing.T) {
	t.Parallel()

	unid := uuid.New().String() // need an actual unique value

	pathValues := echo.PathValues{
		{Name: "unid", Value: unid},
		{Name: "path", Value: "LOGO.TXT"},
	}

	tmp := t.TempDir()
	sl := logs.Discard()
	dirs := command.Dirs{
		Extra:     dir.Directory(tmp),
		Preview:   dir.Directory(tmp),
		Thumbnail: dir.Directory(tmp),
	}

	stale, err := dir.MkdirStale(unid)
	be.Err(t, err, nil)

	dest := filepath.Join(stale, "LOGO.TXT")
	testutil.CopyTXT(t, dest)

	st, err := os.Stat(dest)
	be.Err(t, err, nil)
	be.True(t, st.Size() > 0)

	c := testutil.NewPath(t, "", pathValues)
	got := htmx.FSUseReadme(sl, c, false, dirs)
	be.Err(t, got, nil)
}

func TestFSUseImage(t *testing.T) {
	t.Parallel()

	unid := uuid.New().String() // need an actual unique valid

	pathValues := echo.PathValues{
		{Name: "unid", Value: unid},
		{Name: "path", Value: "SCREEN.PNG"},
	}

	tmp := t.TempDir()
	sl := logs.Discard()
	dirs := command.Dirs{
		Extra:     dir.Directory(tmp),
		Preview:   dir.Directory(tmp),
		Thumbnail: dir.Directory(tmp),
	}

	stale, err := dir.MkdirStale(unid)
	be.Err(t, err, nil)

	dest := filepath.Join(stale, "SCREEN.PNG")
	testutil.CopyPNG(t, dest)

	st, err := os.Stat(dest)
	be.Err(t, err, nil)
	be.True(t, st.Size() > 0)

	c := testutil.NewPath(t, "", pathValues)
	got := htmx.FSUseImage(sl, c, dirs)
	be.Err(t, got, nil)
}
