package htmx_test

// These tests are mostly for nil checks to ensure the server doesn't panic.

import (
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
	"github.com/labstack/echo/v4"
	"github.com/nalgeon/be"
)

func newContext() echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func TestDemozooLookup(t *testing.T) {
	t.Parallel()
	c := newContext()
	err := htmx.DemozooLookup(c, false, nil)
	be.Err(t, err, nil)
}

func TestDemozooValid(t *testing.T) {
	t.Parallel()
	c := newContext()
	x, err := htmx.DemozooValid(c, false, 0)
	be.Err(t, err, nil)
	be.Equal(t, x, demozoo.Production{})
}

func TestDemozooSubmit(t *testing.T) {
	t.Parallel()
	c := newContext()
	err := htmx.DemozooSubmit(c, nil, nil, "")
	be.Err(t, err, nil)
}

func TestDBConnections(t *testing.T) {
	t.Parallel()
	err := htmx.DBConnections(newContext(), nil)
	be.Err(t, err, nil)
}

func TestDeleteForever(t *testing.T) {
	t.Parallel()
	err := htmx.DeleteForever(newContext(), nil, nil, "")
	be.Err(t, err, nil)
	err = htmx.DeleteForever(newContext(), nil, nil, "1")
	be.Err(t, err, nil)
}

func TestPings(t *testing.T) {
	t.Parallel()
	err := htmx.Pings(newContext(), "", -1)
	be.Err(t, err, nil)
}

func TestPouetLookup(t *testing.T) {
	t.Parallel()
	c := newContext()
	err := htmx.PouetLookup(c, nil)
	be.Err(t, err, nil)
}

func TestPouetValid(t *testing.T) {
	t.Parallel()
	c := newContext()
	x, err := htmx.PouetValid(c, -1, true)
	be.Err(t, err, nil)
	be.Equal(t, x, pouet.Response{})
}

func TestPouetSubmit(t *testing.T) {
	t.Parallel()
	c := newContext()
	err := htmx.PouetSubmit(c, nil, nil, "")
	be.Err(t, err, nil)
}

func TestSearchByID(t *testing.T) {
	t.Parallel()
	err := htmx.SearchByID(newContext(), nil, nil)
	be.Err(t, err, nil)
}

func TestSearchReleaser(t *testing.T) {
	t.Parallel()
	err := htmx.SearchReleaser(newContext(), nil, nil)
	be.Err(t, err, nil)
}

func TestDataList(t *testing.T) {
	t.Parallel()
	err := htmx.DataListReleasers(newContext(), nil, nil, "")
	be.Err(t, err, nil)
	err = htmx.DataListMagazines(newContext(), nil, nil, "")
	be.Err(t, err, nil)
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
	err := htmx.HumanizeCount(newContext(), nil, nil, "")
	be.Err(t, err, nil)
}

func TestLookupSHA384(t *testing.T) {
	t.Parallel()
	err := htmx.LookupSHA384(newContext(), nil, nil)
	be.Err(t, err, nil)
}

func TestTransfer(t *testing.T) {
	t.Parallel()
	err := htmx.AdvancedSubmit(newContext(), nil, nil, "")
	be.Err(t, err, nil)
	wd, err := os.Getwd()
	be.Err(t, err, nil)
	err = htmx.AdvancedSubmit(newContext(), nil, nil, dir.Directory(wd))
	be.Err(t, err, nil)
}

func TestProdSubmit(t *testing.T) {
	t.Parallel()
	prod := htmx.Demozoo
	err := prod.Submit(newContext(), nil, nil, "")
	be.Err(t, err, nil)
}

func TestUploadPreview(t *testing.T) {
	t.Parallel()
	err := htmx.UploadPreview(newContext(), "", "")
	be.Err(t, err, nil)
	wd, err := os.Getwd()
	be.Err(t, err, nil)
	err = htmx.UploadPreview(newContext(), dir.Directory(wd), dir.Directory(wd))
	be.Err(t, err, nil)
}

func TestUploadReplacement(t *testing.T) {
	t.Parallel()
	err := htmx.UploadReplacement(newContext(), nil, "", "")
	be.Err(t, err, nil)
	wd, err := os.Getwd()
	be.Err(t, err, nil)
	err = htmx.UploadReplacement(newContext(), nil, dir.Directory(wd), "")
	be.Err(t, err, nil)
}
