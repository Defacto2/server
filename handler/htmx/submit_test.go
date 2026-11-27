package htmx_test

import (
	"bytes"
	"database/sql"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"uuid"

	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/handler/htmx"
	"github.com/Defacto2/server/handler/pouet"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

var errCacheMiss = errors.New("cache miss")

func newMultipartUpload(t *testing.T, target, fieldName, fileName string, content []byte, inputs testutil.Input) *echo.Context {
	t.Helper()

	if target == "" {
		target = "/"
	}
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	for k, v := range inputs {
		be.Err(t, w.WriteField(k, v), nil)
	}
	if fieldName != "" && fileName != "" {
		part, err := w.CreateFormFile(fieldName, fileName)
		be.Err(t, err, nil)
		_, err = part.Write(content)
		be.Err(t, err, nil)
	}
	be.Err(t, w.Close(), nil)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, target, &body)
	req.Header.Set(echo.HeaderContentType, w.FormDataContentType())
	rec := httptest.NewRecorder()
	return echo.New().NewContext(req, rec)
}

func TestUPCount(t *testing.T) {
	t.Parallel()

	err := htmx.UPCount(nil, nil, nil, "")
	be.Err(t, err)

	sl := logs.Discard()
	db := testutil.DB(t)
	c := testutil.NewContext(t, "/classifications")
	err = htmx.UPCount(sl, c, db, "placeholder")
	be.Err(t, err, nil)
}

func TestUPSHA384Offline(t *testing.T) {
	t.Parallel()

	be.Err(t, htmx.UPSHA384(nil, nil, nil))

	sl := logs.Discard()
	dummyDB := &sql.DB{}

	c := testutil.NewContext(t, "/uploader/sha384")
	be.Err(t, htmx.UPSHA384(sl, c, dummyDB), nil)
	rec := recorder(t, c)
	be.Equal(t, rec.Code, http.StatusBadRequest)
	be.Equal(t, rec.Body.String(), "empty hash error")

	c = testutil.NewPath(t, "/uploader/sha384", echo.PathValues{{Name: "hash", Value: "not-a-valid-sha384"}})
	be.Err(t, htmx.UPSHA384(sl, c, dummyDB), nil)
	rec = recorder(t, c)
	be.Equal(t, rec.Code, http.StatusBadRequest)
	be.True(t, strings.Contains(rec.Body.String(), "invalid hash error"))
}

func TestUPSHA384(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	db := testutil.DB(t)

	c := testutil.NewContext(t, "/classifications")
	err := htmx.UPSHA384(sl, c, db)
	be.Err(t, err, nil)

	const testhash = "97982a5b1414b9078103a1c008c4e3526c27b41cdbcf80790560a40f2a9bf2ed4427ab1428789915ed4b3dc07c454bd9"
	pathValues := echo.PathValues{
		{Name: "hash", Value: testhash},
	}
	c = testutil.NewPath(t, "", pathValues)
	err = htmx.UPSHA384(sl, c, db)
	be.Err(t, err, nil)
}

func TestSubmitOffline(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	dummyTx := &sql.Tx{}
	missingDir := dir.Directory(filepath.Join(t.TempDir(), "missing"))
	validDir := dir.Directory(t.TempDir())

	be.Err(t, htmx.UPAdvanced(nil, nil, nil, ""))
	be.Err(t, htmx.UPImage(nil, nil, nil, ""))
	be.Err(t, htmx.UPIntro(nil, nil, nil, ""))
	be.Err(t, htmx.UPMagazine(nil, nil, nil, ""))
	be.Err(t, htmx.UPText(nil, nil, nil, ""))
	be.Err(t, htmx.UPTrainer(nil, nil, nil, ""))

	// Wrappers set expected context metadata and fail cleanly on missing download dir.
	cImg := testutil.NewContext(t, "/uploader/image")
	be.Err(t, htmx.UPImage(sl, cImg, dummyTx, missingDir), nil)
	be.Equal(t, recorder(t, cImg).Code, http.StatusInternalServerError)
	be.Equal(t, cImg.Get("uploader-image-operating-system"), any(tags.Image.String()))

	cIntro := testutil.NewContext(t, "/uploader/intro")
	be.Err(t, htmx.UPIntro(sl, cIntro, dummyTx, missingDir), nil)
	be.Equal(t, recorder(t, cIntro).Code, http.StatusInternalServerError)
	be.Equal(t, cIntro.Get("uploader-intro-category"), any(tags.Intro.String()))

	cMag := testutil.NewContext(t, "/uploader/magazine")
	be.Err(t, htmx.UPMagazine(sl, cMag, dummyTx, missingDir), nil)
	be.Equal(t, recorder(t, cMag).Code, http.StatusInternalServerError)
	be.Equal(t, cMag.Get("uploader-magazine-category"), any(tags.Mag.String()))

	cTxt := testutil.NewContext(t, "/uploader/text")
	be.Err(t, htmx.UPText(sl, cTxt, dummyTx, validDir), nil)
	be.Equal(t, recorder(t, cTxt).Code, http.StatusBadRequest)

	cTrn := testutil.NewContext(t, "/uploader/trainer")
	be.Err(t, htmx.UPTrainer(sl, cTrn, dummyTx, validDir), nil)
	be.Equal(t, recorder(t, cTrn).Code, http.StatusBadRequest)

	cAdv := testutil.NewContext(t, "/uploader/advanced")
	be.Err(t, htmx.UPAdvanced(sl, cAdv, dummyTx, validDir), nil)
	be.Equal(t, recorder(t, cAdv).Code, http.StatusBadRequest)
}

func TestTransferHelpers(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	dlDir := t.TempDir()
	tr := htmx.Transfer{Key: "uploader-test", Download: dir.Directory(dlDir)}

	// StatusOK nil check and HTML-escaped filename rendering.
	be.Err(t, tr.StatusOK(nil, "a.txt", 1))
	c := testutil.NewContext(t, "/")
	be.Err(t, tr.StatusOK(c, "a<script>.txt", 1), nil)
	rec := recorder(t, c)
	be.Equal(t, rec.Code, http.StatusOK)
	be.True(t, strings.Contains(rec.Body.String(), "a&lt;script&gt;.txt"))

	// CopyTemp error and happy paths.
	uid := uuid.New()
	tr.CopyTemp(sl, uuid.Nil(), filepath.Join(t.TempDir(), "ignored.zip"))
	tr.CopyTemp(nil, uid, filepath.Join(t.TempDir(), "missing.zip"))
	tr.CopyTemp(sl, uid, t.TempDir())

	src := filepath.Join(t.TempDir(), "upload.bin")
	be.Err(t, os.WriteFile(src, []byte("archive-bytes"), 0o600), nil)
	tr.CopyTemp(sl, uid, src)

	copied, err := os.ReadFile(filepath.Join(dlDir, uid.String()))
	be.Err(t, err, nil)
	be.Equal(t, string(copied), "archive-bytes")

	// SaveTemp nil guard and real multipart file save.
	_, err = htmx.SaveTemp(nil, nil, "k", nil)
	be.Err(t, err)

	cFile := testutil.NewFile(t, "/", "upfile", "note.txt")
	fh, err := cFile.FormFile("upfile")
	be.Err(t, err, nil)
	tmpSaved, err := htmx.SaveTemp(sl, cFile, "up", fh)
	be.Err(t, err, nil)
	be.True(t, tmpSaved != "")
	t.Cleanup(func() { _ = os.Remove(tmpSaved) })

	savedData, err := os.ReadFile(tmpSaved)
	be.Err(t, err, nil)
	be.Equal(t, string(savedData), "Hello world!")

	// Insert.Upload nil guard.
	_, _, err = (htmx.Insert{}).Upload(nil, nil, nil)
	be.Err(t, err)

	// Close on nil, open, and already-closed file.
	htmx.Close(sl, "test", nil)
	f, err := fh.Open()
	be.Err(t, err, nil)
	htmx.Close(sl, "test", f)
	htmx.Close(sl, "test", f)
}

func TestRecordFormData(t *testing.T) {
	t.Parallel()

	var rec htmx.Record
	be.True(t, strings.Contains(rec.FormData(nil), "broken"))

	c := testutil.NewInputs(t, "/", testutil.Input{
		"artifact-editor-unid": "bad/uuid",
	})
	be.Equal(t, rec.FormData(c), "The editor file upload unique identifier is invalid")

	c = testutil.NewInputs(t, "/", testutil.Input{
		"artifact-editor-unid": "not-a-uuid",
	})
	be.Equal(t, rec.FormData(c), "The editor file upload unique identifier is invalid")

	c = testutil.NewInputs(t, "/", testutil.Input{
		"artifact-editor-unid":       testutil.UID4,
		"artifact-editor-record-key": "abc",
	})
	be.Equal(t, rec.FormData(c), "The editor file upload record key is invalid")

	c = testutil.NewInputs(t, "/", testutil.Input{
		"artifact-editor-unid":              testutil.UID4,
		"artifact-editor-record-key":        "12",
		"artifact-editor-download-classify": "dos",
	})
	be.Equal(t, rec.FormData(c), "")
}

func TestSubmit(t *testing.T) {
	t.Parallel()

	err := htmx.UPAdvanced(nil, nil, nil, "")
	be.Err(t, err)

	sl := logs.Discard()
	tx := testutil.Tx(t)
	c := testutil.NewContext(t, "/uploader/advanced")

	download := dir.Directory(t.TempDir())
	err = htmx.UPAdvanced(sl, c, tx, download)
	be.Err(t, err, nil)

	const key = "uploader-advanced"
	c = testutil.NewFile(t, "", key+"file", "filename.txt")
	err = htmx.UPAdvanced(sl, c, tx, download)
	be.Err(t, err, nil)
}

func TestPouetSubmit(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	db := testutil.DB(t)
	tx := testutil.Tx(t)
	c := testutil.NewPath(t,
		"/pouet/production", echo.PathValues{{Name: "id", Value: "9999"}})
	err := htmx.Pouet.Submit(sl, c, db, tx, "")
	be.Err(t, err, nil)
	_ = tx.Rollback()
}

func TestSubmitImage(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	u := htmx.Submit{}
	be.Err(t, u.Image(sl, nil))
	be.Err(t, u.StatusOK(nil, "pic.png"))

	cStatus := testutil.NewContext(t, "/")
	be.Err(t, u.StatusOK(cStatus, "pic.png"), nil)
	be.Equal(t, recorder(t, cStatus).Code, http.StatusOK)
	be.True(t, strings.Contains(recorder(t, cStatus).Body.String(), "pic.png"))

	// Missing Preview / Thumbnail dirs return 500.
	missing := dir.Directory(filepath.Join(t.TempDir(), "missing"))
	u.Preview = missing
	c := testutil.NewContext(t, "/upload/preview")
	be.Err(t, u.Image(sl, c), nil)
	be.Equal(t, recorder(t, c).Code, http.StatusInternalServerError)

	tmp := t.TempDir()
	u.Preview = dir.Directory(tmp)
	u.Thumbnail = missing
	c = testutil.NewContext(t, "/upload/preview")
	be.Err(t, u.Image(sl, c), nil)
	be.Equal(t, recorder(t, c).Code, http.StatusInternalServerError)

	u.Thumbnail = dir.Directory(tmp)

	// Missing form data returns 400.
	c = testutil.NewContext(t, "/upload/preview")
	be.Err(t, u.Image(sl, c), nil)
	rec := recorder(t, c)
	be.Equal(t, rec.Code, http.StatusBadRequest)
	be.True(t, strings.Contains(rec.Body.String(), "unique identifier is invalid"))

	const uid = "5c735996-dca8-4a93-9737-43b576ffe366"
	formInputs := testutil.Input{
		"artifact-editor-unid":       uid,
		"artifact-editor-record-key": "1",
	}

	// Missing file input returns 400.
	c = testutil.NewInputs(t, "/upload/preview", formInputs)
	be.Err(t, u.Image(sl, c), nil)
	be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

	// Unsupported archive file returns 400.
	zipBytes, err := os.ReadFile(testutil.FileData("archivezip").Abs())
	be.Err(t, err, nil)
	c = newMultipartUpload(t, "/upload/preview", "artifact-editor-replace-preview",
		"archive.zip", zipBytes, formInputs)
	be.Err(t, u.Image(sl, c), nil)
	rec = recorder(t, c)
	be.Equal(t, rec.Code, http.StatusBadRequest)
	be.Equal(t, rec.Body.String(), "The chosen file is not a valid image or text file")

	// Valid PNG image succeeds with 200 OK.
	pngBytes, err := os.ReadFile(testutil.FileData("testpng").Abs())
	be.Err(t, err, nil)
	c = newMultipartUpload(t, "/upload/preview", "artifact-editor-replace-preview",
		"sample.png", pngBytes, formInputs)
	be.Err(t, u.Image(sl, c), nil)
	rec = recorder(t, c)
	be.Equal(t, rec.Code, http.StatusOK)
	be.True(t, strings.Contains(rec.Body.String(), "sample.png"))
}

func TestSubmitReplacementOffline(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	dummyTx := &sql.Tx{}
	u := htmx.Submit{Download: dir.Directory(filepath.Join(t.TempDir(), "missing"))}

	be.Err(t, u.Replacement(nil, nil, nil))

	c := testutil.NewContext(t, "/upload/file")
	be.Err(t, u.Replacement(sl, c, dummyTx), nil)
	be.Equal(t, recorder(t, c).Code, http.StatusInternalServerError)

	u.Download = dir.Directory(t.TempDir())
	c = testutil.NewContext(t, "/upload/file")
	be.Err(t, u.Replacement(sl, c, dummyTx), nil)
	be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)

	formInputs := testutil.Input{
		"artifact-editor-unid":       testutil.UID4,
		"artifact-editor-record-key": "1",
	}
	c = testutil.NewInputs(t, "/upload/file", formInputs)
	be.Err(t, u.Replacement(sl, c, dummyTx), nil)
	be.Equal(t, recorder(t, c).Code, http.StatusBadRequest)
}

func TestSubmitReplacement(t *testing.T) {
	t.Parallel()

	tx := testutil.Tx(t)
	sl := logs.Discard()
	u := htmx.Submit{}

	tmp := t.TempDir()
	u.Download = dir.Directory(tmp)
	u.Preview = dir.Directory(tmp)
	u.Thumbnail = dir.Directory(tmp)

	c := testutil.NewContext(t, "/upload/file")
	err := u.Replacement(sl, c, tx)
	be.Err(t, err, nil)

	const uid = "5c735996-dca8-4a93-9737-43b576ffe366"
	formInputs := testutil.Input{
		"artifact-editor-unid":              uid,
		"artifact-editor-record-key":        "1",
		"artifact-editor-download-classify": "text",
	}
	c = testutil.NewFileInputs(t,
		"", "artifact-editor-replace-file", "filename.txt", formInputs)
	err = u.Replacement(sl, c, tx)
	be.Err(t, err, nil)
}

func TestProdOffline(t *testing.T) {
	t.Parallel()

	be.Equal(t, htmx.Demozoo.String(), "demozoo")
	be.Equal(t, htmx.Pouet.String(), "pouet")
	be.Equal(t, htmx.Prod(-1).String(), "")
	be.Equal(t, htmx.Prod(99).String(), "")

	be.Err(t, htmx.Demozoo.Submit(nil, nil, nil, nil, ""))

	// Prod.ID validation across Demozoo and Pouet.
	_, err := htmx.Demozoo.ID(nil)
	be.Err(t, err)

	c := testutil.NewPath(t, "/", echo.PathValues{{Name: "id", Value: "abc"}})
	id, err := htmx.Demozoo.ID(c)
	be.Err(t, err, nil)
	be.Equal(t, id, 0)
	be.Equal(t, recorder(t, c).Code, http.StatusNotAcceptable)

	c = testutil.NewPath(t, "/", echo.PathValues{{Name: "id", Value: "0"}})
	id, err = htmx.Demozoo.ID(c)
	be.Err(t, err, nil)
	be.Equal(t, id, 0)
	be.Equal(t, recorder(t, c).Code, http.StatusNotAcceptable)

	c = testutil.NewPath(t, "/", echo.PathValues{{Name: "id", Value: "99999999"}})
	be.True(t, 99999999 > demozoo.Sanity && 99999999 > pouet.Sanity)
	_, err = htmx.Pouet.ID(c)
	be.Err(t, err, nil)
	be.Equal(t, recorder(t, c).Code, http.StatusNotAcceptable)

	c = testutil.NewPath(t, "/", echo.PathValues{{Name: "id", Value: "101"}})
	id, err = htmx.Demozoo.ID(c)
	be.Err(t, err, nil)
	be.Equal(t, id, 101)

	c = testutil.NewPath(t, "/", echo.PathValues{{Name: "id", Value: "202"}})
	id, err = htmx.Pouet.ID(c)
	be.Err(t, err, nil)
	be.Equal(t, id, 202)

	// Prod.Lookup nil and non-numeric guards.
	dummyDB := &sql.DB{}
	be.Err(t, htmx.Prod(99).Lookup(testutil.NewContext(t, "/"), dummyDB, false), nil)
	be.Err(t, htmx.Demozoo.Lookup(nil, nil, false))
	be.Err(t, htmx.Pouet.Lookup(nil, nil, false))

	c = testutil.NewInput(t, "/", "demozoo-submission", "abc")
	be.Err(t, htmx.Demozoo.Lookup(c, dummyDB, false), nil)
	be.Equal(t, recorder(t, c).Code, http.StatusNotAcceptable)

	c = testutil.NewInput(t, "/", "pouet-submission", "abc")
	be.Err(t, htmx.Pouet.Lookup(c, dummyDB, false), nil)
	be.Equal(t, recorder(t, c).Code, http.StatusNotAcceptable)

	// Prod.ButtonOK renders HTML forms for Demozoo and Pouet.
	c = testutil.NewContext(t, "/")
	be.Err(t, htmx.Demozoo.ButtonOK(c, 42, "Demo Title"), nil)
	rec := recorder(t, c)
	be.Equal(t, rec.Code, http.StatusOK)
	be.True(t, strings.Contains(rec.Body.String(), `hx-put="/demozoo/production/42"`))

	c = testutil.NewContext(t, "/")
	be.Err(t, htmx.Pouet.ButtonOK(c, 43, "Pouet Title"), nil)
	rec = recorder(t, c)
	be.Equal(t, rec.Code, http.StatusOK)
	be.True(t, strings.Contains(rec.Body.String(), `hx-put="/pouet/production/43"`))

	be.Err(t, htmx.Prod(99).ButtonOK(testutil.NewContext(t, "/"), 44), nil)

	// Prod.ValidFn nil, invalid id, cache hit, cache miss, and cache disabled.
	_, err = htmx.Demozoo.ValidFn(nil, 1, false, nil)
	be.Err(t, err)

	c = testutil.NewContext(t, "/")
	exit, err := htmx.Demozoo.ValidFn(c, 0, false, nil)
	be.Err(t, err, nil)
	be.True(t, exit)
	be.Equal(t, recorder(t, c).Code, http.StatusNotAcceptable)

	c = testutil.NewContext(t, "/")
	exit, err = htmx.Demozoo.ValidFn(c, 55, true, func(string) (string, error) {
		return "Amiga - Demo", nil
	})
	be.Err(t, err, nil)
	be.True(t, exit)
	be.True(t, strings.Contains(recorder(t, c).Body.String(), "Amiga - Demo"))

	c = testutil.NewContext(t, "/")
	exit, err = htmx.Demozoo.ValidFn(c, 55, true, func(string) (string, error) {
		return "", errCacheMiss
	})
	be.Err(t, err, nil)
	be.True(t, !exit)

	// ValidateDemozoo and ValidatePouet offline checks.
	_, err = htmx.ValidateDemozoo(nil, 1, false)
	be.Err(t, err)

	c = testutil.NewContext(t, "/")
	_, err = htmx.ValidateDemozoo(c, 0, false)
	be.Err(t, err, nil)
	be.Equal(t, recorder(t, c).Code, http.StatusNotAcceptable)

	c = testutil.NewContext(t, "/")
	_, err = htmx.ValidateDemozoo(c, 1, false)
	be.Err(t, err, config.ErrNoZooAuth)

	_, err = htmx.ValidatePouet(nil, 1, false)
	be.Err(t, err)

	c = testutil.NewContext(t, "/")
	_, err = htmx.ValidatePouet(c, 0, false)
	be.Err(t, err, nil)
	be.Equal(t, recorder(t, c).Code, http.StatusNotAcceptable)
}

func TestProdSubmit(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	db := testutil.DB(t)
	tx := testutil.Tx(t)
	defer func() { _ = tx.Rollback() }()

	const target = "/pouet/production"
	c := testutil.NewContext(t, target)

	download := dir.Directory(t.TempDir())
	prod := htmx.Demozoo
	err := prod.Submit(sl, c, db, tx, download)
	be.Err(t, err, nil)

	pathValues := echo.PathValues{
		{Name: "id", Value: "10101"},
	}
	c = testutil.NewPath(t, target, pathValues)
	err = prod.Submit(sl, c, db, tx, download)
	be.Err(t, err, nil)

	prod = htmx.Pouet
	c = testutil.NewContext(t, target)
	err = prod.Submit(sl, c, db, tx, download)
	be.Err(t, err, nil)

	c = testutil.NewPath(t, target, pathValues)
	err = prod.Submit(sl, c, db, tx, download)
	be.Err(t, err, nil)
}
