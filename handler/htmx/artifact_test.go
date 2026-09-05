package htmx_test

import (
	"testing"

	"github.com/Defacto2/server/handler/htmx"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

const (
	unid = "123e4567-e89b-12d3-a456-426614174000"
	unv4 = "bb2310e1-93aa-475e-8b88-59eb1fb984a4"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want error
	}{
		{
			name: "absolute path",
			path: "/absolute/path",
			want: htmx.ErrPath,
		},
		{
			name: "clean path",
			path: "relative/path",
			want: nil,
		},
		{
			name: "clean path",
			path: "relative/path/",
			want: nil,
		},
		{
			name: "unclean path 1",
			path: "relative/../path",
			want: htmx.ErrPath,
		},
		{
			name: "unclean path 2",
			path: "./relative/path",
			want: htmx.ErrPath,
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

			gotUnid, gotName, err := htmx.Path(c)
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
		prefix + "youtube": "abcdefghijk", // FIX: create a global in testutil
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
		prefix + "youtubeval": "abcdefghijk",
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
}
