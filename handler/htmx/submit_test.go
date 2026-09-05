package htmx_test

import (
	"os"
	"testing"

	"github.com/Defacto2/server/handler/htmx"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

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

func TestSubmit(t *testing.T) {
	t.Parallel()

	err := htmx.UPAdvanced(nil, nil, nil, "")
	be.Err(t, err)

	sl := logs.Discard()
	tx := testutil.Tx(t)
	c := testutil.NewContext(t, "/uploader/advanced")

	wd, err := os.Getwd()
	download := dir.Directory(wd)
	be.Err(t, err, nil)

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
	tx := testutil.Tx(t)
	c := testutil.NewInput(t,
		"/pouet/production", "pouet-submission", "9999")
	err := htmx.Pouet.Submit(sl, c, tx, "")
	be.Err(t, err, nil)
	_ = tx.Rollback()
}

func TestSubmitImage(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	u := htmx.Submit{}
	err := u.Image(sl, nil)
	be.Err(t, err)

	c := testutil.NewContext(t, "/upload/preview")
	wd, err := os.Getwd()
	be.Err(t, err, nil)

	u.Preview = dir.Directory(wd)
	u.Thumbnail = dir.Directory(wd)
	err = u.Image(sl, c)
	be.Err(t, err, nil)

	const unid = "5c735996-dca8-4a93-9737-43b576ffe366"
	formInputs := testutil.Input{
		"artifact-editor-unid":       unid,
		"artifact-editor-record-key": "1",
	}
	c = testutil.NewFileInputs(t,
		"", "artifact-editor-replace-preview", "filename.txt", formInputs)
	err = u.Image(sl, c)
	be.Err(t, err, nil)
}

func TestSubmitReplacement(t *testing.T) {
	t.Parallel()

	tx := testutil.Tx(t)
	sl := logs.Discard()
	u := htmx.Submit{}

	wd, err := os.Getwd()
	be.Err(t, err, nil)
	u.Download = dir.Directory(wd)
	u.Preview = dir.Directory(wd)
	u.Thumbnail = dir.Directory(wd)

	c := testutil.NewContext(t, "/upload/file")
	err = u.Image(sl, c)
	be.Err(t, err, nil)

	const unid = "5c735996-dca8-4a93-9737-43b576ffe366"
	formInputs := testutil.Input{
		"artifact-editor-unid":              unid,
		"artifact-editor-record-key":        "1",
		"artifact-editor-download-classify": "text",
	}
	c = testutil.NewFileInputs(t,
		"", "artifact-editor-replace-file", "filename.txt", formInputs)
	err = u.Replacement(sl, c, tx)
	be.Err(t, err, nil)
}

func TestProdSubmit(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	tx := testutil.Tx(t)

	const target = "/pouet/production"
	c := testutil.NewContext(t, target)

	wd, err := os.Getwd()
	be.Err(t, err, nil)

	download := dir.Directory(wd)
	prod := htmx.Demozoo
	err = prod.Submit(sl, c, tx, download)
	be.Err(t, err, nil)

	pathValues := echo.PathValues{
		{Name: "id", Value: "10101"},
	}
	c = testutil.NewPath(t, target, pathValues)
	err = prod.Submit(sl, c, tx, download)
	be.Err(t, err, nil)

	prod = htmx.Pouet
	err = prod.Submit(sl, c, tx, download)
	be.Err(t, err, nil)

	pathValues = echo.PathValues{
		{Name: "id", Value: "10101"},
	}
	c = testutil.NewPath(t, target, pathValues)
	err = prod.Submit(sl, c, tx, download)
	be.Err(t, err, nil)
}
