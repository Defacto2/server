package htmx_test

// These tests are mostly for nil checks to ensure the server doesn't panic.

import (
	"context"
	"database/sql"
	"embed"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/handler/htmx"
	"github.com/Defacto2/server/handler/pouet"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

func newContext() *echo.Context {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", strings.NewReader("{}"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func TestDemozooLookup(t *testing.T) {
	t.Parallel()
	c := newContext()
	var db sql.DB
	err := htmx.DemozooLookup(context.TODO(), c, &db, false)
	be.Err(t, err, nil)
}

func TestDemozooValid(t *testing.T) {
	t.Parallel()
	c := newContext()
	x, err := htmx.DemozooValid(context.TODO(), c, false, 0)
	be.Err(t, err, nil)
	be.Equal(t, x, demozoo.Production{})
}

func TestDemozooSubmit(t *testing.T) {
	t.Parallel()
	c := newContext()
	err := htmx.DemozooSubmit(context.TODO(), nil, c, nil, "")
	be.Err(t, err)
}

func TestDBConnections(t *testing.T) {
	t.Parallel()
	err := htmx.DBConnections(newContext(), nil)
	be.Err(t, err)
}

func TestDeleteForever(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	err := htmx.DeleteForever(ctx, nil, newContext(), nil, "")
	be.Err(t, err)
	err = htmx.DeleteForever(ctx, nil, newContext(), nil, "1")
	be.Err(t, err)
}

func TestPings(t *testing.T) {
	t.Parallel()
	err := htmx.Pings(newContext(), "", -1)
	be.Err(t, err, nil)
}

func TestPouetLookup(t *testing.T) {
	t.Parallel()
	c := newContext()
	err := htmx.PouetLookup(context.TODO(), c, nil)
	be.Err(t, err)
}

func TestPouetValid(t *testing.T) {
	t.Parallel()
	c := newContext()
	x, err := htmx.PouetValid(context.TODO(), c, -1, true)
	be.Err(t, err, nil)
	be.Equal(t, x, pouet.Response{})
}

func TestPouetSubmit(t *testing.T) {
	t.Parallel()
	c := newContext()
	err := htmx.PouetSubmit(context.TODO(), nil, c, nil, "")
	be.Err(t, err)
}

func TestSearchByID(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	err := htmx.SearchByID(ctx, nil, newContext(), nil)
	be.Err(t, err)
}

func TestSearchReleaser(t *testing.T) {
	t.Parallel()
	err := htmx.SearchReleaser(context.TODO(), nil, newContext(), nil, nil)
	be.Err(t, err)
}

func TestDataList(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	err := htmx.DataListReleasers(ctx, nil, newContext(), nil, "")
	be.Err(t, err)
	err = htmx.DataListMagazines(ctx, nil, newContext(), nil, "")
	be.Err(t, err)
}

func TestTemplates(t *testing.T) {
	t.Parallel()
	x := htmx.Templates(embed.FS{})
	be.True(t, len(x) == 3)
}

func TestTemplateFuncMap(t *testing.T) {
	t.Parallel()
	x := htmx.TemplateFuncMap()
	be.True(t, x != nil)
}

func TestSuggestion(t *testing.T) {
	t.Parallel()
	s := htmx.Suggestion("", "", "")
	be.Equal(t, "suggestion type error: string", s)
	s = htmx.Suggestion("a group", "grp", 10)
	be.Equal(t, "a group, grp (10 items)", s)
}

func TestHumanizeCount(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	err := htmx.HumanizeCount(ctx, nil, newContext(), nil, "")
	be.Err(t, err)
}

func TestLookupSHA384(t *testing.T) {
	t.Parallel()
	err := htmx.LookupSHA384(context.TODO(), nil, newContext(), nil)
	be.Err(t, err)
}

func TestTransfer(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	err := htmx.AdvancedSubmit(ctx, nil, newContext(), nil, "")
	be.Err(t, err)
	wd, err := os.Getwd()
	be.Err(t, err, nil)
	err = htmx.AdvancedSubmit(ctx, nil, newContext(), nil, dir.Directory(wd))
	be.Err(t, err)
}

func TestProdSubmit(t *testing.T) {
	t.Parallel()
	prod := htmx.Demozoo
	err := prod.Submit(context.TODO(), nil, newContext(), nil, "")
	be.Err(t, err)
}

func TestUploadPreview(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	err := htmx.UploadPreview(ctx, logs.Discard(), newContext(), "", "")
	be.Err(t, err, nil)
	wd, err := os.Getwd()
	be.Err(t, err, nil)
	err = htmx.UploadPreview(ctx, logs.Discard(), newContext(), dir.Directory(wd), dir.Directory(wd))
	be.Err(t, err, nil)
}

func TestUploadReplacement(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	d := logs.Discard()
	err := htmx.UploadReplacement(ctx, d, newContext(), nil, "", "")
	be.Err(t, err)
	wd, err := os.Getwd()
	be.Err(t, err, nil)
	err = htmx.UploadReplacement(ctx, d, newContext(), nil, dir.Directory(wd), "")
	be.Err(t, err)
}
