package htmx_test

// These tests are mostly for nil checks to ensure the server doesn't panic.

import (
	"embed"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/htmx"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	err := htmx.DemozooLookup(c, nil)
	require.NoError(t, err)
}

func TestDemozooValid(t *testing.T) {
	t.Parallel()
	c := newContext()
	x, err := htmx.DemozooValid(c, 0)
	require.NoError(t, err)
	assert.Empty(t, x)
}

func TestDemozooSubmit(t *testing.T) {
	t.Parallel()
	c := newContext()
	err := htmx.DemozooSubmit(c, nil, nil, "")
	require.NoError(t, err)
}

func TestDBConnections(t *testing.T) {
	t.Parallel()
	err := htmx.DBConnections(newContext(), nil)
	require.NoError(t, err)
}

func TestDeleteForever(t *testing.T) {
	t.Parallel()
	err := htmx.DeleteForever(newContext(), nil, nil, "")
	require.NoError(t, err)
	err = htmx.DeleteForever(newContext(), nil, nil, "1")
	require.NoError(t, err)
}

func TestPings(t *testing.T) {
	t.Parallel()
	err := htmx.Pings(newContext(), "", -1)
	require.NoError(t, err)
}

func TestPouetLookup(t *testing.T) {
	t.Parallel()
	c := newContext()
	err := htmx.PouetLookup(c, nil)
	require.NoError(t, err)
}

func TestPouetValid(t *testing.T) {
	t.Parallel()
	c := newContext()
	x, err := htmx.PouetValid(c, -1, true)
	require.NoError(t, err)
	assert.Empty(t, x)
}

func TestPouetSubmit(t *testing.T) {
	t.Parallel()
	c := newContext()
	err := htmx.PouetSubmit(c, nil, nil, "")
	require.NoError(t, err)
}

func TestSearchByID(t *testing.T) {
	t.Parallel()
	err := htmx.SearchByID(newContext(), nil, nil)
	require.NoError(t, err)
}

func TestSearchReleaser(t *testing.T) {
	t.Parallel()
	err := htmx.SearchReleaser(newContext(), nil, nil)
	require.NoError(t, err)
}

func TestDataList(t *testing.T) {
	t.Parallel()
	err := htmx.DataListReleasers(newContext(), nil, nil, "")
	require.NoError(t, err)
	err = htmx.DataListMagazines(newContext(), nil, nil, "")
	require.NoError(t, err)
}

func TestTemplates(t *testing.T) {
	t.Parallel()
	x := htmx.Templates(embed.FS{})
	assert.Len(t, x, 3)
}

func TestTemplateFuncMap(t *testing.T) {
	t.Parallel()
	x := htmx.TemplateFuncMap()
	assert.Empty(t, len(x))
}

func TestSuggestion(t *testing.T) {
	t.Parallel()
	s := htmx.Suggestion("", "", "")
	assert.Equal(t, "suggestion type error: string", s)
	s = htmx.Suggestion("a group", "grp", 10)
	assert.Equal(t, "a group, grp (10 items)", s)
}

func TestHumanizeCount(t *testing.T) {
	t.Parallel()
	err := htmx.HumanizeCount(newContext(), nil, nil, "")
	require.NoError(t, err)
}

func TestLookupSHA384(t *testing.T) {
	t.Parallel()
	err := htmx.LookupSHA384(newContext(), nil, nil)
	require.NoError(t, err)
}

func TestTransfer(t *testing.T) {
	t.Parallel()
	err := htmx.AdvancedSubmit(newContext(), nil, nil, "")
	require.NoError(t, err)
	dir, err := os.Getwd()
	require.NoError(t, err)
	err = htmx.AdvancedSubmit(newContext(), nil, nil, dir)
	require.NoError(t, err)
}

func TestProdSubmit(t *testing.T) {
	t.Parallel()
	prod := htmx.Demozoo
	err := prod.Submit(newContext(), nil, nil, "")
	require.NoError(t, err)
}

func TestUploadPreview(t *testing.T) {
	t.Parallel()
	err := htmx.UploadPreview(newContext(), "", "")
	require.NoError(t, err)
	dir, err := os.Getwd()
	require.NoError(t, err)
	err = htmx.UploadPreview(newContext(), dir, dir)
	require.NoError(t, err)
}

func TestUploadReplacement(t *testing.T) {
	t.Parallel()
	err := htmx.UploadReplacement(newContext(), nil, "", "")
	require.NoError(t, err)
	dir, err := os.Getwd()
	require.NoError(t, err)
	err = htmx.UploadReplacement(newContext(), nil, dir, "")
	require.NoError(t, err)
}
