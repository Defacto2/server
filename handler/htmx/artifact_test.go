package htmx_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"uuid"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/htmx"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

const (
	unid = testutil.UID
)

var (
	errCommitBoom = errors.New("commit callback failed")
	errLinkBoom   = errors.New("link update failed")
)

func recorder(t *testing.T, c *echo.Context) *httptest.ResponseRecorder {
	t.Helper()

	resp, err := echo.UnwrapResponse(c.Response())
	be.Err(t, err, nil)

	rec, ok := resp.Unwrap().(*httptest.ResponseRecorder)
	be.True(t, ok)

	return rec
}

func cleanupStale(t *testing.T, stale string) {
	t.Helper()

	t.Cleanup(func() {
		r, err := os.OpenRoot(os.TempDir())
		if err != nil {
			t.Log(err)
			return
		}
		defer func() { _ = r.Close() }()

		if err = r.RemoveAll(filepath.Base(stale)); err != nil {
			t.Log(err)
		}
	})
}

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
			name: "clean path trailing slash", path: "relative/path/", want: nil,
		},
		{
			name: "unclean path 1", path: "relative/../path", want: htmx.ErrPath,
		},
		{
			name: "unclean path 2", path: "./relative/path", want: htmx.ErrPath,
		},
		{
			name: "empty path", path: "", want: htmx.ErrPath,
		},
		{
			name: "current dir", path: ".", want: htmx.ErrPath,
		},
		{
			name: "parent dir", path: "..", want: htmx.ErrPath,
		},
		{
			name: "parent traversal", path: "../secret.txt", want: htmx.ErrPath,
		},
		{
			name: "nested parent traversal", path: "../../etc/passwd", want: htmx.ErrPath,
		},
		{
			name: "leading double dot filename", path: "..readme.txt", want: nil,
		},
		{
			name: "hidden filename", path: ".hidden", want: nil,
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

	_, _, err := htmx.Paths(nil)
	be.Err(t, err)

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
			name:     "invalid unid format",
			unid:     "invalid/unid",
			path:     "relative/path",
			wantUnid: "",
			wantName: "",
			wantErr:  true,
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
			name:     "bad query escape",
			unid:     unid,
			path:     "bad%zz",
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
			be.Equal(t, gotUnid, tt.wantUnid)
			be.Equal(t, gotName, tt.wantName)
		})
	}
}

func TestUUID(t *testing.T) {
	t.Parallel()

	_, err := htmx.UUID(nil)
	be.Err(t, err)

	c := testutil.NewPath(t, "", echo.PathValues{{Name: "unid", Value: unid}})
	got, err := htmx.UUID(c)
	be.Err(t, err, nil)
	be.Equal(t, got, unid)

	c = testutil.NewContext(t, "")
	_, err = htmx.UUID(c)
	be.Err(t, err)

	c = testutil.NewPath(t, "", echo.PathValues{{Name: "unid", Value: "bad/uuid"}})
	_, err = htmx.UUID(c)
	be.Err(t, err)

	c = testutil.NewPath(t, "", echo.PathValues{{Name: "unid", Value: "not-a-uuid"}})
	_, err = htmx.UUID(c)
	be.Err(t, err)
}

func TestKeyAndKeyParam(t *testing.T) {
	t.Parallel()

	_, err := htmx.Key(nil)
	be.Err(t, err)

	c := testutil.NewInput(t, "", htmx.EditorKey, "42")
	got, err := htmx.Key(c)
	be.Err(t, err, nil)
	be.Equal(t, got, int64(42))

	c = testutil.NewInput(t, "", htmx.EditorKey, "not-a-number")
	_, err = htmx.Key(c)
	be.Err(t, err)

	_, err = htmx.KeyParam(nil)
	be.Err(t, err)

	c = testutil.NewPath(t, "", echo.PathValues{{Name: "id", Value: "99"}})
	got, err = htmx.KeyParam(c)
	be.Err(t, err, nil)
	be.Equal(t, got, int64(99))

	c = testutil.NewPath(t, "", echo.PathValues{{Name: "id", Value: "abc"}})
	_, err = htmx.KeyParam(c)
	be.Err(t, err)
}

func TestStatusOK(t *testing.T) {
	t.Parallel()

	c := testutil.NewContext(t, "")
	be.Err(t, htmx.StatusOK(c, "artifact-editor-title-undo"), nil)
	be.Equal(t, recorder(t, c).Code, http.StatusNoContent)

	c = testutil.NewContext(t, "")
	be.Err(t, htmx.StatusOK(c, "artifact-editor-date-update"), nil)
	rec := recorder(t, c)
	be.Equal(t, rec.Code, http.StatusOK)
	be.Equal(t, rec.Body.String(), "Save &#x2713;")

	c = testutil.NewContext(t, "")
	be.Err(t, htmx.StatusOK(c, "artifact-editor-title"), nil)
	rec = recorder(t, c)
	be.Equal(t, rec.Code, http.StatusOK)
	be.True(t, strings.Contains(rec.Body.String(), "text-success"))
}

func TestCommitHelpersOffline(t *testing.T) {
	t.Parallel()

	dummyTx := &sql.Tx{}
	okStr := func(gotKey *int64, gotVal *string) func(context.Context, boil.ContextExecutor, int64, string) error {
		return func(_ context.Context, _ boil.ContextExecutor, k int64, v string) error {
			*gotKey = k
			*gotVal = v
			return nil
		}
	}
	failStr := func(context.Context, boil.ContextExecutor, int64, string) error {
		return errCommitBoom
	}
	okBool := func(gotKey *int64, gotVal *bool) func(context.Context, boil.ContextExecutor, int64, bool) error {
		return func(_ context.Context, _ boil.ContextExecutor, k int64, v bool) error {
			*gotKey = k
			*gotVal = v
			return nil
		}
	}
	failBool := func(context.Context, boil.ContextExecutor, int64, bool) error {
		return errCommitBoom
	}

	t.Run("CommitSanitize", func(t *testing.T) {
		t.Parallel()

		be.Err(t, htmx.CommitSanitize(nil, dummyTx, "f", strings.TrimSpace, failStr))

		c := testutil.NewInput(t, "", htmx.EditorKey, "5")
		be.Err(t, htmx.CommitSanitize(c, nil, "f", strings.TrimSpace, failStr), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		c = testutil.NewInput(t, "", htmx.EditorKey, "bad")
		be.Err(t, htmx.CommitSanitize(c, dummyTx, "f", strings.TrimSpace, failStr), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		c = testutil.NewInputs(t, "", testutil.Input{htmx.EditorKey: "5", "f": " hi "})
		be.Err(t, htmx.CommitSanitize(c, dummyTx, "f", strings.TrimSpace, failStr), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		var k int64
		var v string
		c = testutil.NewInputs(t, "", testutil.Input{htmx.EditorKey: "5", "f": " hi "})
		be.Err(t, htmx.CommitSanitize(c, dummyTx, "f", strings.TrimSpace, okStr(&k, &v)), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusOK)
		be.Equal(t, k, int64(5))
		be.Equal(t, v, "hi")

		// Nil sanitize passes value through without panicking.
		c = testutil.NewInputs(t, "", testutil.Input{htmx.EditorKey: "6", "f": " raw "})
		be.Err(t, htmx.CommitSanitize(c, dummyTx, "f", nil, okStr(&k, &v)), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusOK)
		be.Equal(t, k, int64(6))
		be.Equal(t, v, " raw ")
	})

	t.Run("CommitNotOn", func(t *testing.T) {
		t.Parallel()

		c := testutil.NewPath(t, "", echo.PathValues{{Name: "id", Value: "7"}})
		be.Err(t, htmx.CommitNotOn(c, nil, "flag", failBool), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		c = testutil.NewPath(t, "", echo.PathValues{{Name: "id", Value: "bad"}})
		be.Err(t, htmx.CommitNotOn(c, dummyTx, "flag", failBool), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		c = testutil.NewPath(t, "", echo.PathValues{{Name: "id", Value: "7"}})
		be.Err(t, htmx.CommitNotOn(c, dummyTx, "flag", failBool), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		var k int64
		var v bool
		c = testutil.NewInputsPath(t, "", testutil.Input{"flag": "on"}, echo.PathValues{{Name: "id", Value: "7"}})
		be.Err(t, htmx.CommitNotOn(c, dummyTx, "flag", okBool(&k, &v)), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusOK)
		be.Equal(t, k, int64(7))
		be.Equal(t, v, false)
	})

	t.Run("CommitOn", func(t *testing.T) {
		t.Parallel()

		c := testutil.NewPath(t, "", echo.PathValues{{Name: "id", Value: "8"}})
		be.Err(t, htmx.CommitOn(c, nil, "flag", failBool), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		c = testutil.NewPath(t, "", echo.PathValues{{Name: "id", Value: "bad"}})
		be.Err(t, htmx.CommitOn(c, dummyTx, "flag", failBool), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		c = testutil.NewPath(t, "", echo.PathValues{{Name: "id", Value: "8"}})
		be.Err(t, htmx.CommitOn(c, dummyTx, "flag", failBool), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		var k int64
		var v bool
		c = testutil.NewInputsPath(t, "", testutil.Input{"flag": "on"}, echo.PathValues{{Name: "id", Value: "8"}})
		be.Err(t, htmx.CommitOn(c, dummyTx, "flag", okBool(&k, &v)), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusOK)
		be.Equal(t, k, int64(8))
		be.Equal(t, v, true)
	})

	t.Run("CommitStr", func(t *testing.T) {
		t.Parallel()

		c := testutil.NewInput(t, "", htmx.EditorKey, "9")
		be.Err(t, htmx.CommitStr(c, nil, "f", failStr), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		c = testutil.NewInput(t, "", htmx.EditorKey, "bad")
		be.Err(t, htmx.CommitStr(c, dummyTx, "f", failStr), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		c = testutil.NewInputs(t, "", testutil.Input{htmx.EditorKey: "9", "f": "val"})
		be.Err(t, htmx.CommitStr(c, dummyTx, "f", failStr), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		var k int64
		var v string
		c = testutil.NewInputs(t, "", testutil.Input{htmx.EditorKey: "9", "f": "val"})
		be.Err(t, htmx.CommitStr(c, dummyTx, "f", okStr(&k, &v)), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusOK)
		be.Equal(t, k, int64(9))
		be.Equal(t, v, "val")
	})

	t.Run("CommitStrKey", func(t *testing.T) {
		t.Parallel()

		c := testutil.NewPath(t, "", echo.PathValues{{Name: "id", Value: "10"}})
		be.Err(t, htmx.CommitStrKey(c, nil, "f", failStr), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		c = testutil.NewPath(t, "", echo.PathValues{{Name: "id", Value: "bad"}})
		be.Err(t, htmx.CommitStrKey(c, dummyTx, "f", failStr), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		c = testutil.NewPath(t, "", echo.PathValues{{Name: "id", Value: "10"}})
		be.Err(t, htmx.CommitStrKey(c, dummyTx, "f", failStr), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		var k int64
		var v string
		c = testutil.NewInputsPath(t, "", testutil.Input{"f": "vga"}, echo.PathValues{{Name: "id", Value: "10"}})
		be.Err(t, htmx.CommitStrKey(c, dummyTx, "f", okStr(&k, &v)), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusOK)
		be.Equal(t, k, int64(10))
		be.Equal(t, v, "vga")
	})
}

func TestTxOfflinePaths(t *testing.T) {
	t.Parallel()

	dummyTx := &sql.Tx{}
	sl := logs.Discard()

	t.Run("TxEmulateRunProg", func(t *testing.T) {
		t.Parallel()

		c := testutil.NewContext(t, "")
		be.Err(t, htmx.TxEmulateRunProg(c, nil))

		c = testutil.NewPath(t, "", echo.PathValues{{Name: "id", Value: "bad"}})
		be.Err(t, htmx.TxEmulateRunProg(c, dummyTx), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		c = testutil.NewInputsPath(t, "",
			testutil.Input{"emulate-run-program": "bad;cmd"},
			echo.PathValues{{Name: "id", Value: "1"}},
		)
		be.Err(t, htmx.TxEmulateRunProg(c, dummyTx), nil)
		rec := recorder(t, c)
		be.Equal(t, rec.Code, http.StatusOK)
		be.True(t, strings.Contains(rec.Body.String(), "invalid-feedback"))
	})

	t.Run("TxLinksUndo", func(t *testing.T) {
		t.Parallel()

		be.Err(t, htmx.TxLinksUndo(testutil.NewContext(t, ""), nil))

		c := testutil.NewInput(t, "", htmx.EditorKey, "bad")
		be.Err(t, htmx.TxLinksUndo(c, dummyTx), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		c = testutil.NewInputs(t, "", testutil.Input{
			htmx.EditorKey:                "1",
			"artifact-editor-youtubeundo": "not-valid-yt!",
		})
		be.Err(t, htmx.TxLinksUndo(c, dummyTx), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		c = testutil.NewInputs(t, "", testutil.Input{
			htmx.EditorKey:                "1",
			"artifact-editor-demozooundo": "abc",
		})
		be.Err(t, htmx.TxLinksUndo(c, dummyTx), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		c = testutil.NewInputs(t, "", testutil.Input{
			htmx.EditorKey:                "1",
			"artifact-editor-demozooundo": "999999999",
		})
		be.Err(t, htmx.TxLinksUndo(c, dummyTx), nil)
		rec := recorder(t, c)
		be.Equal(t, rec.Code, http.StatusBadRequest)
		be.True(t, !strings.Contains(rec.Body.String(), "%!w"))

		c = testutil.NewInputs(t, "", testutil.Input{
			htmx.EditorKey:                "1",
			"artifact-editor-demozooundo": "-1",
		})
		be.Err(t, htmx.TxLinksUndo(c, dummyTx), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		c = testutil.NewInputs(t, "", testutil.Input{
			htmx.EditorKey:              "1",
			"artifact-editor-pouetundo": "abc",
		})
		be.Err(t, htmx.TxLinksUndo(c, dummyTx), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		c = testutil.NewInputs(t, "", testutil.Input{
			htmx.EditorKey:              "1",
			"artifact-editor-pouetundo": "999999999",
		})
		be.Err(t, htmx.TxLinksUndo(c, dummyTx), nil)
		rec = recorder(t, c)
		be.Equal(t, rec.Code, http.StatusBadRequest)
		be.True(t, !strings.Contains(rec.Body.String(), "%!w"))

		c = testutil.NewInputs(t, "", testutil.Input{
			htmx.EditorKey:              "1",
			"artifact-editor-pouetundo": "-1",
		})
		be.Err(t, htmx.TxLinksUndo(c, dummyTx), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)
	})

	t.Run("TxCreditUndo", func(t *testing.T) {
		t.Parallel()

		be.Err(t, htmx.TxCreditUndo(testutil.NewContext(t, ""), nil))

		c := testutil.NewInput(t, "", htmx.EditorKey, "bad")
		be.Err(t, htmx.TxCreditUndo(c, dummyTx), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		c = testutil.NewInputs(t, "", testutil.Input{
			htmx.EditorKey:                 "1",
			"artifact-editor-credits-undo": "too;few",
		})
		be.Err(t, htmx.TxCreditUndo(c, dummyTx), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		c = testutil.NewInputs(t, "", testutil.Input{
			htmx.EditorKey:                 "1",
			"artifact-editor-credits-undo": "a;b;c;d",
			"artifact-editor-credittext":   "a",
			"artifact-editor-creditill":    "b",
			"artifact-editor-creditprog":   "c",
			"artifact-editor-creditaudio":  "d",
		})
		be.Err(t, htmx.TxCreditUndo(c, dummyTx), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusNoContent)
	})

	t.Run("TxTags Public YouTube", func(t *testing.T) {
		t.Parallel()

		be.Err(t, htmx.TxTags(nil, testutil.NewContext(t, ""), dummyTx))
		c := testutil.NewInputs(t, "", testutil.Input{
			htmx.EditorKey:                    "bad",
			"artifact-editor-categories":      "bbs",
			"artifact-editor-operatingsystem": "dos",
		})
		be.Err(t, htmx.TxTags(sl, c, dummyTx), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		be.Err(t, htmx.TxPublic(testutil.NewContext(t, ""), nil, true))
		c = testutil.NewInput(t, "", htmx.EditorKey, "bad")
		be.Err(t, htmx.TxPublic(c, dummyTx, true), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		be.Err(t, htmx.TxPublicByKey(testutil.NewContext(t, ""), nil, "1", true))
		c = testutil.NewContext(t, "")
		be.Err(t, htmx.TxPublicByKey(c, dummyTx, "bad", true), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		be.Err(t, htmx.TxYouTube(testutil.NewContext(t, ""), nil))
		c = testutil.NewInput(t, "", htmx.EditorKey, "bad")
		be.Err(t, htmx.TxYouTube(c, dummyTx), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		c = testutil.NewInputs(t, "", testutil.Input{
			htmx.EditorKey:            "1",
			"artifact-editor-youtube": "short",
		})
		be.Err(t, htmx.TxYouTube(c, dummyTx), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusNoContent)
	})

	t.Run("TxReleasers and TxYMD", func(t *testing.T) {
		t.Parallel()

		be.Err(t, htmx.TxReleasers(testutil.NewContext(t, ""), nil))
		be.Err(t, htmx.TxReleasersUndo(testutil.NewContext(t, ""), nil))
		be.Err(t, htmx.TxReleasers(testutil.NewInput(t, "", htmx.EditorKey, "bad"), dummyTx))

		c := testutil.NewInputs(t, "", testutil.Input{
			htmx.EditorKey:              "1",
			"releaser1":                 "alpha",
			"releaser2":                 "beta",
			"artifact-editor-releaser1": "alpha",
			"artifact-editor-releaser2": "beta",
		})
		be.Err(t, htmx.TxReleasers(c, dummyTx), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusNoContent)

		be.Err(t, htmx.TxYMD(testutil.NewContext(t, ""), nil))
		be.Err(t, htmx.TxYMDUndo(testutil.NewContext(t, ""), nil))
		be.Err(t, htmx.TxLastMod(testutil.NewContext(t, ""), nil))
		be.Err(t, htmx.TxYMD(testutil.NewInput(t, "", htmx.EditorKey, "bad"), dummyTx))

		// Unchanged DateForm returns 204.
		c = testutil.NewInputs(t, "", testutil.Input{
			htmx.EditorKey:                "1",
			"artifact-editor-year":        "1995",
			"artifact-editor-month":       "5",
			"artifact-editor-day":         "10",
			"artifact-editor-year-store":  "1995",
			"artifact-editor-month-store": "5",
			"artifact-editor-day-store":   "10",
		})
		be.Err(t, htmx.TxYMD(c, dummyTx), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusNoContent)

		// Invalid date in DateForm returns 400.
		c = testutil.NewInputs(t, "", testutil.Input{
			htmx.EditorKey:          "1",
			"artifact-editor-year":  "9999",
			"artifact-editor-month": "99",
			"artifact-editor-day":   "99",
		})
		be.Err(t, htmx.TxYMD(c, dummyTx), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		// Bad YYYY-MM-DD in DateUndo / DateLast returns 400.
		c = testutil.NewInputs(t, "", testutil.Input{
			htmx.EditorKey:               "1",
			"artifact-editor-date-undos": "bad-date",
		})
		be.Err(t, htmx.TxYMDUndo(c, dummyTx), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

		// Unchanged DateLast returns 204.
		c = testutil.NewInputs(t, "", testutil.Input{
			htmx.EditorKey:                  "1",
			"artifact-editor-year":          "1995",
			"artifact-editor-month":         "5",
			"artifact-editor-day":           "10",
			"artifact-editor-date-lastmods": "1995-5-10",
		})
		be.Err(t, htmx.TxLastMod(c, dummyTx), nil)
		be.Equal(t, recorder(t, c).Code, http.StatusNoContent)
	})
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

// TestEmulateCPU also tests CommitStrKey.
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
	rec := recorder(t, c)
	be.Equal(t, rec.Code, http.StatusOK)
	be.True(t, strings.Contains(rec.Body.String(), "Link to"))

	cErr := testutil.NewContext(t, "")
	be.Err(t, htmx.HTMLLinkTo(cErr, errLinkBoom), nil)
	be.Equal(t, recorder(t, cErr).Code, http.StatusBadRequest)

	be.Err(t, htmx.HTMLLinkTo(nil, nil))
}

func TestLinksUndo(t *testing.T) {
	t.Parallel()

	tx := testutil.Tx(t)

	const prefix = "artifact-editor-"
	formInputs := testutil.Input{
		htmx.EditorKey:         "1",
		prefix + "youtubeundo": testutil.YT,
	}
	c := testutil.NewInputs(t, "", formInputs)
	got := htmx.TxLinksUndo(c, tx)
	be.Err(t, got, nil)
}

func TestCreditUndo(t *testing.T) {
	t.Parallel()

	tx := testutil.Tx(t)

	const prefix = "artifact-editor-credit"
	formInputs := testutil.Input{
		htmx.EditorKey:    "1",
		prefix + "text":   "abcde",
		prefix + "s-undo": "qwerty;;;",
	}
	c := testutil.NewInputs(t, "", formInputs)
	got := htmx.TxCreditUndo(c, tx)
	be.Err(t, got, nil)
}

func TestTxTags(t *testing.T) {
	t.Parallel()

	tx := testutil.Tx(t)
	sl := logs.Discard()

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

	formInputs := testutil.Input{
		htmx.EditorKey:            "1",
		"artifact-editor-youtube": testutil.YT,
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

	be.Equal(t, htmx.FileID.Ext(), ".diz")
	be.Equal(t, htmx.Text.Ext(), ".txt")
	be.Equal(t, htmx.Helper.Ext(), ".hlp")
	be.Equal(t, htmx.Copy(99).Ext(), "")

	be.Equal(t, htmx.FileID.String(), "DIZ")
	be.Equal(t, htmx.Text.String(), "Text")
	be.Equal(t, htmx.Helper.String(), "Helper text")
	be.Equal(t, htmx.Copy(99).String(), "")
}

func TestMkCopy(t *testing.T) {
	t.Parallel()

	uid := uuid.New().String()
	const fname = "filename.txt"
	tmp := t.TempDir()

	pathValues := echo.PathValues{
		{Name: "unid", Value: uid},
		{Name: "path", Value: fname},
	}

	dirs := command.Dirs{Extra: dir.Directory(tmp)}

	// Nil context returns error.
	be.Err(t, htmx.FileID.MkCopy(nil, dirs))

	// Invalid path returns 400.
	cBad := testutil.NewPath(t, "", echo.PathValues{{Name: "unid", Value: "bad"}, {Name: "path", Value: fname}})
	be.Err(t, htmx.FileID.MkCopy(cBad, dirs), nil)
	be.Equal(t, recorder(t, cBad).Code, http.StatusBadRequest)

	// Missing source file returns 400.
	cMissing := testutil.NewPath(t, "", pathValues)
	be.Err(t, htmx.FileID.MkCopy(cMissing, dirs), nil)
	be.Equal(t, recorder(t, cMissing).Code, http.StatusBadRequest)

	stale, err := dir.MkdirStale(uid)
	be.Err(t, err, nil)
	cleanupStale(t, stale)

	r, err := os.OpenRoot(stale)
	be.Err(t, err, nil)
	defer func() { _ = r.Close() }()

	// Empty file returns 200 without copying.
	_, err = helper.TouchWR(r, "empty.txt")
	be.Err(t, err, nil)
	cEmpty := testutil.NewPath(t, "", echo.PathValues{
		{Name: "unid", Value: uid},
		{Name: "path", Value: "empty.txt"},
	})
	be.Err(t, htmx.FileID.MkCopy(cEmpty, dirs), nil)
	recEmpty := recorder(t, cEmpty)
	be.Equal(t, recEmpty.Code, http.StatusOK)
	be.Equal(t, recEmpty.Body.String(), "The file is empty and was not copied.")

	// Valid file is copied by FSCopyDIZ and FSCopyHelp.
	b := []byte("hello world")
	n, err := helper.TouchWR(r, fname, b...)
	be.Err(t, err, nil)
	be.Equal(t, n, len(b))

	cDiz := testutil.NewPath(t, "", pathValues)
	be.Err(t, htmx.FSCopyDIZ(cDiz, dirs), nil)
	recDiz := recorder(t, cDiz)
	be.Equal(t, recDiz.Code, http.StatusOK)
	be.Equal(t, recDiz.Header().Get("Hx-Refresh"), "true")
	be.True(t, strings.Contains(recDiz.Body.String(), "DIZ copied"))
	copied, err := os.ReadFile(filepath.Join(tmp, uid+".diz"))
	be.Err(t, err, nil)
	be.Equal(t, string(copied), "hello world")

	cHlp := testutil.NewPath(t, "", pathValues)
	be.Err(t, htmx.FSCopyHelp(cHlp, dirs), nil)
	recHlp := recorder(t, cHlp)
	be.True(t, strings.Contains(recHlp.Body.String(), "Helper text copied"))
	copiedHlp, err := os.ReadFile(filepath.Join(tmp, uid+".hlp"))
	be.Err(t, err, nil)
	be.Equal(t, string(copiedHlp), "hello world")
}

func TestFSCopyReadme(t *testing.T) {
	t.Parallel()

	uid := uuid.New().String()
	const fname = "readme.txt"
	tmp := t.TempDir()

	pathValues := echo.PathValues{
		{Name: "unid", Value: uid},
		{Name: "path", Value: fname},
	}

	sl := logs.Discard()
	dirs := command.Dirs{
		Extra:     dir.Directory(tmp),
		Preview:   dir.Directory(tmp),
		Thumbnail: dir.Directory(tmp),
	}

	be.Err(t, htmx.FSCopyReadme(nil, testutil.NewContext(t, ""), dirs))

	cBad := testutil.NewPath(t, "", echo.PathValues{{Name: "unid", Value: "bad"}, {Name: "path", Value: fname}})
	be.Err(t, htmx.FSCopyReadme(sl, cBad, dirs), nil)
	be.Equal(t, recorder(t, cBad).Code, http.StatusBadRequest)

	stale, err := dir.MkdirStale(uid)
	be.Err(t, err, nil)
	cleanupStale(t, stale)

	r, err := os.OpenRoot(stale)
	be.Err(t, err, nil)
	defer func() { _ = r.Close() }()

	// Empty file returns 200 without copying.
	_, err = helper.TouchWR(r, "empty.txt")
	be.Err(t, err, nil)
	cEmpty := testutil.NewPath(t, "", echo.PathValues{
		{Name: "unid", Value: uid},
		{Name: "path", Value: "empty.txt"},
	})
	be.Err(t, htmx.FSCopyReadme(sl, cEmpty, dirs), nil)
	be.Equal(t, recorder(t, cEmpty).Body.String(), "The file is empty and was not copied.")

	// Pre-create thumbnail so offline test verifies copy without invoking external ansilove binary.
	be.Err(t, os.WriteFile(filepath.Join(tmp, uid+".png"), []byte("png"), 0o600), nil)

	b := []byte("hello world")
	n, err := helper.TouchWR(r, fname, b...)
	be.Err(t, err, nil)
	be.Equal(t, n, len(b))

	c := testutil.NewPath(t, "", pathValues)
	be.Err(t, htmx.FSCopyReadme(sl, c, dirs), nil)
	rec := recorder(t, c)
	be.Equal(t, rec.Header().Get("Hx-Refresh"), "true")
	be.True(t, strings.Contains(rec.Body.String(), "Images copied"))

	saved, err := os.ReadFile(filepath.Join(tmp, uid+".txt"))
	be.Err(t, err, nil)
	be.Equal(t, string(saved), "hello world")
}

func TestFSRemoveExtras(t *testing.T) {
	t.Parallel()

	be.Equal(t, htmx.Diz.String(), ".diz")
	be.Equal(t, htmx.Hlp.String(), ".hlp")
	be.Equal(t, htmx.Txt.String(), ".txt")
	be.Equal(t, htmx.Zip.String(), ".zip")
	be.Equal(t, htmx.Ext(-1).String(), "")
	be.Equal(t, htmx.Ext(99).String(), "")

	tmp := t.TempDir()
	extra := dir.Directory(tmp)
	pathValues := echo.PathValues{{Name: "unid", Value: unid}}

	// Nil context and invalid UUID checks.
	be.Err(t, htmx.FSRemoveDIZ(nil, extra))
	cBad := testutil.NewPath(t, "", echo.PathValues{{Name: "unid", Value: "bad"}})
	be.Err(t, htmx.FSRemoveDIZ(cBad, extra), nil)
	be.Equal(t, recorder(t, cBad).Code, http.StatusBadRequest)

	// Missing file returns 200 OK.
	cMissing := testutil.NewPath(t, "", pathValues)
	be.Err(t, htmx.FSRemoveDIZ(cMissing, extra), nil)
	be.Equal(t, recorder(t, cMissing).Code, http.StatusOK)

	// Directory at target path returns 400.
	dirTarget := filepath.Join(tmp, unid+".diz")
	be.Err(t, os.Mkdir(dirTarget, 0o755), nil)
	cDir := testutil.NewPath(t, "", pathValues)
	be.Err(t, htmx.FSRemoveDIZ(cDir, extra), nil)
	be.Equal(t, recorder(t, cDir).Code, http.StatusBadRequest)
	be.Err(t, os.Remove(dirTarget), nil)

	// Existing files for all 4 extra types are deleted with Hx-Refresh header.
	cases := []struct {
		ext string
		fn  func(*echo.Context, dir.Directory) error
	}{
		{".diz", htmx.FSRemoveDIZ},
		{".hlp", htmx.FSRemoveHelp},
		{".txt", htmx.FSRemoveReadme},
		{".zip", htmx.FSRemoveZipRepack},
	}
	for _, tc := range cases {
		p := filepath.Join(tmp, unid+tc.ext)
		be.Err(t, os.WriteFile(p, []byte("data"), 0o600), nil)

		c := testutil.NewPath(t, "", pathValues)
		be.Err(t, tc.fn(c, extra), nil)
		be.Equal(t, recorder(t, c).Header().Get("Hx-Refresh"), "true")

		_, err := os.Stat(p)
		be.True(t, os.IsNotExist(err))
	}
}

func TestFSRemoveCache(t *testing.T) {
	t.Parallel()

	cBad := testutil.NewPath(t, "", echo.PathValues{{Name: "unid", Value: "not-a-uuid"}})
	be.Err(t, htmx.FSRemoveCache(cBad), nil)
	be.Equal(t, recorder(t, cBad).Code, http.StatusBadRequest)

	cBadFmt := testutil.NewPath(t, "", echo.PathValues{{Name: "unid", Value: "bad/uuid"}})
	be.Err(t, htmx.FSRemoveCache(cBadFmt), nil)
	be.Equal(t, recorder(t, cBadFmt).Code, http.StatusBadRequest)

	uid := uuid.New().String()
	stale, err := dir.MkdirStale(uid)
	be.Err(t, err, nil)
	cleanupStale(t, stale)

	c := testutil.NewPath(t, "", echo.PathValues{{Name: "unid", Value: uid}})
	be.Err(t, htmx.FSRemoveCache(c), nil)
	rec := recorder(t, c)
	be.Equal(t, rec.Code, http.StatusOK)
	be.Equal(t, rec.Header().Get("Hx-Refresh"), "true")
	be.Equal(t, rec.Body.String(), "cache removed")

	be.True(t, htmx.PageReload(nil) == nil)

	_, err = os.Stat(stale)
	be.True(t, os.IsNotExist(err))
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

	// Nil-guard & invalid-UUID paths.
	be.Err(t, htmx.FSCrop(nil, testutil.NewContext(t, ""), command.OneTwo, dirs))
	be.Err(t, htmx.FSAlign(nil, testutil.NewContext(t, ""), command.Left, dirs))
	be.Err(t, htmx.FSThumb(nil, testutil.NewContext(t, ""), command.Pixel, dirs))
	be.Err(t, htmx.FSPixelate(nil, testutil.NewContext(t, ""), dirs.Preview))
	be.Err(t, htmx.FSRemoveImages(nil, dirs.Preview))

	cBad := testutil.NewPath(t, "", echo.PathValues{{Name: "unid", Value: "bad"}})
	be.Err(t, htmx.FSRemoveImages(cBad, dirs.Preview), nil)
	be.Equal(t, recorder(t, cBad).Code, http.StatusBadRequest)

	// No images yet -> ErrNoImages returns 200 with error text.
	cNone := testutil.NewPath(t, "", pathValues)
	be.Err(t, htmx.FSRemoveImages(cNone, dirs.Preview), nil)
	be.Equal(t, recorder(t, cNone).Code, http.StatusOK)

	dest := filepath.Join(tmp, unid+".png")
	testutil.CopyPNG(t, dest)
	st, err := os.Stat(dest)
	be.Err(t, err, nil)
	be.Equal(t, st.Size(), int64(testutil.SCREENPNG))

	cCrop := testutil.NewPath(t, "", pathValues)
	be.Err(t, htmx.FSCrop(sl, cCrop, command.OneTwo, dirs), nil)

	cAlign := testutil.NewPath(t, "", pathValues)
	be.Err(t, htmx.FSAlign(sl, cAlign, command.Left, dirs), nil)

	cThumb := testutil.NewPath(t, "", pathValues)
	be.Err(t, htmx.FSThumb(sl, cThumb, command.Pixel, dirs), nil)

	cPix := testutil.NewPath(t, "", pathValues)
	be.Err(t, htmx.FSPixelate(sl, cPix, dirs.Preview), nil)

	cRm := testutil.NewPath(t, "", pathValues)
	be.Err(t, htmx.FSRemoveImages(cRm, dirs.Preview), nil)
	recRm := recorder(t, cRm)
	be.Equal(t, recRm.Code, http.StatusOK)
	be.True(t, strings.Contains(recRm.Body.String(), "Images are erased."))

	_, err = os.Stat(dest)
	be.True(t, os.IsNotExist(err))
}

func TestFSUseReadme(t *testing.T) {
	t.Parallel()

	uid := uuid.New().String()
	pathValues := echo.PathValues{
		{Name: "unid", Value: uid},
		{Name: "path", Value: "LOGO.TXT"},
	}

	tmp := t.TempDir()
	sl := logs.Discard()
	dirs := command.Dirs{
		Extra:     dir.Directory(tmp),
		Preview:   dir.Directory(tmp),
		Thumbnail: dir.Directory(tmp),
	}

	be.Err(t, htmx.FSUseReadme(nil, testutil.NewContext(t, ""), false, dirs))

	cBad := testutil.NewPath(t, "", echo.PathValues{{Name: "unid", Value: "bad"}, {Name: "path", Value: "LOGO.TXT"}})
	be.Err(t, htmx.FSUseReadme(sl, cBad, false, dirs), nil)
	be.Equal(t, recorder(t, cBad).Code, http.StatusBadRequest)

	stale, err := dir.MkdirStale(uid)
	be.Err(t, err, nil)
	cleanupStale(t, stale)

	// Missing file returns 400.
	cMissing := testutil.NewPath(t, "", echo.PathValues{{Name: "unid", Value: uid}, {Name: "path", Value: "MISSING.TXT"}})
	be.Err(t, htmx.FSUseReadme(sl, cMissing, false, dirs), nil)
	be.Equal(t, recorder(t, cMissing).Code, http.StatusBadRequest)

	// Empty file returns 200 without invoking external imager.
	be.Err(t, os.WriteFile(filepath.Join(stale, "EMPTY.TXT"), nil, 0o600), nil)
	cEmpty := testutil.NewPath(t, "", echo.PathValues{{Name: "unid", Value: uid}, {Name: "path", Value: "EMPTY.TXT"}})
	be.Err(t, htmx.FSUseReadme(sl, cEmpty, false, dirs), nil)
	be.Equal(t, recorder(t, cEmpty).Body.String(), "The file is empty and was not used.")

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

	uid := uuid.New().String()
	pathValues := echo.PathValues{
		{Name: "unid", Value: uid},
		{Name: "path", Value: "SCREEN.PNG"},
	}

	tmp := t.TempDir()
	sl := logs.Discard()
	dirs := command.Dirs{
		Extra:     dir.Directory(tmp),
		Preview:   dir.Directory(tmp),
		Thumbnail: dir.Directory(tmp),
	}

	be.Err(t, htmx.FSUseImage(nil, testutil.NewContext(t, ""), dirs))
	be.Err(t, htmx.FSUseCGA(nil, testutil.NewContext(t, ""), dirs))
	be.Err(t, htmx.FSUseBinText(nil, testutil.NewContext(t, ""), dirs))

	stale, err := dir.MkdirStale(uid)
	be.Err(t, err, nil)
	cleanupStale(t, stale)

	// Missing and empty file checks across FSUseImage / FSUseCGA / FSUseBinText.
	cMissing := testutil.NewPath(t, "", echo.PathValues{{Name: "unid", Value: uid}, {Name: "path", Value: "NOPE.BIN"}})
	be.Err(t, htmx.FSUseCGA(sl, cMissing, dirs), nil)
	be.Equal(t, recorder(t, cMissing).Code, http.StatusBadRequest)

	be.Err(t, os.WriteFile(filepath.Join(stale, "EMPTY.BIN"), nil, 0o600), nil)
	emptyPath := echo.PathValues{{Name: "unid", Value: uid}, {Name: "path", Value: "EMPTY.BIN"}}

	cCGA := testutil.NewPath(t, "", emptyPath)
	be.Err(t, htmx.FSUseCGA(sl, cCGA, dirs), nil)
	be.Equal(t, recorder(t, cCGA).Body.String(), "The file is empty and was not used.")

	cBin := testutil.NewPath(t, "", emptyPath)
	be.Err(t, htmx.FSUseBinText(sl, cBin, dirs), nil)
	be.Equal(t, recorder(t, cBin).Body.String(), "The file is empty and was not used.")

	dest := filepath.Join(stale, "SCREEN.PNG")
	testutil.CopyPNG(t, dest)

	st, err := os.Stat(dest)
	be.Err(t, err, nil)
	be.True(t, st.Size() > 0)

	c := testutil.NewPath(t, "", pathValues)
	got := htmx.FSUseImage(sl, c, dirs)
	be.Err(t, got, nil)
	be.True(t, strings.Contains(recorder(t, c).Body.String(), "Images copied"))
}
