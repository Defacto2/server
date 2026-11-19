package htmx_test

// These tests are mostly for nil checks to ensure the server doesn't panic.

// checked in Sep 26, test coverage was great at under 70%

import (
	"bytes"
	"embed"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"uuid"

	"github.com/Defacto2/server/handler/fulltext"
	"github.com/Defacto2/server/handler/htmx"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/null/v8"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

func TestAreacodes(t *testing.T) {
	t.Parallel()

	c := testutil.NewInput(t,
		"/areacodes", "htmx-search", "CA,201")
	err := htmx.Areacodes(c)
	be.Err(t, err, nil)

	form := url.Values{}
	form.Set("htmx-search", "201")
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/areacodes", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c2 := echo.New().NewContext(req, rec)
	err = htmx.Areacodes(c2)
	be.Err(t, err, nil)
	be.True(t, !strings.Contains(rec.Body.String(), "<br><br>"))
	be.True(t, strings.Contains(rec.Body.String(), "201 - New Jersey (NJ)"))
}

func TestDemozooLookup(t *testing.T) {
	t.Parallel()

	const (
		existing  = "1"
		newrecord = "99999"
	)

	db := testutil.DB(t)
	c := testutil.NewInput(t,
		"/demozoo/production", "demozoo-submission", existing)
	err := htmx.Demozoo.Lookup(c, db, false)
	be.Err(t, err)

	c.Set("zooauth", "")
	err = htmx.Demozoo.Lookup(c, db, false)
	be.Err(t, err, nil)

	c = testutil.NewInput(t,
		"/demozoo/production", "demozoo-submission", newrecord)
	c.Set("zooauth", "")
	err = htmx.Demozoo.Lookup(c, db, false)
	be.Err(t, err, nil)
}

func TestDemozooValid(t *testing.T) {
	t.Parallel()

	c := testutil.NewInput(t,
		"/demozoo/production", "demozoo-submission", "1")
	got, err := htmx.ValidateDemozoo(c, 1, false)
	be.Err(t, err)
	t.Log(got)

	c.Set("zooauth", "")
	got, err = htmx.ValidateDemozoo(c, 1, false)
	be.Err(t, err, nil)

	t.Log(got)
}

func TestDBConnections(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	c := testutil.NewContext(t, "/dbconns")
	err := htmx.DBConnections(c, db)
	be.Err(t, err, nil)
}

func TestDeleteForever(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	tx := testutil.Tx(t)

	err := htmx.DeleteForever(sl, testutil.NewContext(t, "/delete/forever/0"), tx, "0")
	be.Err(t, err, nil)
	_ = tx.Rollback()
	// despite the rollback option, it is probably best not to test the delete forever func on an actual record
}

func TestPings(t *testing.T) {
	t.Parallel()

	// ping errors get send and handled by the context
	err := htmx.Pings(testutil.NewContext(t, "/pings"), "http", 1323)
	be.Err(t, err, nil)
}

func TestPouetLookup(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	c := testutil.NewInput(t,
		"/pouet/production", "pouet-submission", "1")
	err := htmx.Pouet.Lookup(c, db, false)
	be.Err(t, err, nil)

	c = testutil.NewInput(t,
		"/pouet/production", "pouet-submission", "9999")
	err = htmx.Pouet.Lookup(c, db, false)
	be.Err(t, err, nil)
}

func TestPouetValid(t *testing.T) {
	t.Parallel()

	// WARN: do not run working tests over this, as the func pings the target website.
	_, err := htmx.ValidatePouet(nil, -1, true)
	be.Err(t, err)
}

func TestSearchByID(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	db := testutil.DB(t)
	c := testutil.NewInput(t, "/editor/search", "htmx-search", "1")
	err := htmx.SearchByID(sl, c, db)
	be.Err(t, err, nil)
}

func TestAlternatives(t *testing.T) {
	t.Parallel()

	s := htmx.Alternatives("")
	be.Equal(t, len(s), 0)

	s = htmx.Alternatives("trs")
	be.True(t, len(s) > 0)
}

func TestSearchReleaser(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	db := testutil.DB(t)
	c := testutil.NewInput(t, "/search/releaser", "htmx-search", "defacto2")

	ft := fulltext.Tidbits{}
	ft.New()
	got := htmx.SearchReleaser(sl, c, db, &ft)
	be.Err(t, got, nil)
}

func TestDataList(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	db := testutil.DB(t)

	c := testutil.NewContext(t, "/datalist/releasers")
	err := htmx.DataListReleasers(sl, c, db, "defacto2")
	be.Err(t, err, nil)

	c = testutil.NewContext(t, "/datalist/magazines")
	err = htmx.DataListMagazines(sl, c, db, "defacto2")
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

	got := htmx.Suggestion("", "", "")
	be.Equal(t, got, "suggestion type error: string")

	const count = 10
	got = htmx.Suggestion("a group", "grp", count)
	be.Equal(t, got, "a group, grp ("+strconv.Itoa(count)+" items)")
}

func TestAppendUniqueReleasers(t *testing.T) {
	t.Parallel()

	dst := model.Releasers{
		{Unique: model.Releaser{URI: "razor-1911", Name: "Razor 1911"}},
	}
	add := model.Releasers{
		{Unique: model.Releaser{URI: "razor-1911", Name: "Razor 1911"}},
		{Unique: model.Releaser{URI: "class", Name: "Class"}},
		{Unique: model.Releaser{URI: "paradigm", Name: "Paradigm"}},
	}
	got := htmx.AppendUniqueReleasers(dst, add, 2)
	be.Equal(t, len(got), 2)
	be.Equal(t, got[0].Unique.URI, "razor-1911")
	be.Equal(t, got[1].Unique.URI, "class")
}

func TestSearchReleasersTemplate(t *testing.T) {
	t.Parallel()

	tmpl, err := template.New("").Funcs(htmx.TemplateFuncMap()).ParseFiles(
		filepath.Join("..", "..", "view", "htmx", "layout.tmpl"),
		filepath.Join("..", "..", "view", "htmx", "searchreleasers.tmpl"),
	)
	be.Err(t, err, nil)

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{
		"name":    "razor",
		"input":   "razor",
		"maximum": 14,
		"result": model.Releasers{
			{Unique: model.Releaser{URI: "razor-1911", Name: "Razor 1911", Count: 5, Bytes: 1024}},
			{Unique: model.Releaser{URI: "class", Name: "Class", Count: 2, Bytes: 512}},
		},
		"fulltext": []fulltext.Result{},
	})
	be.Err(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, "row-cols-md-3"))
	be.True(t, !strings.Contains(out, "row-col-md-3"))
	be.True(t, !strings.Contains(out, "border border ")) //nolint:dupword
}

func TestParseSearchIDs(t *testing.T) {
	t.Parallel()

	uid := "c8cd0b4c-2f54-11e0-8827-cc1607e15609"
	ids, uuids := htmx.ParseSearchIDs("1, 0, -5, " + uid + ", 00000000-0000-0000-0000-000000000000, invalid")
	be.Equal(t, ids, []int{1})
	be.Equal(t, uuids, []uuid.UUID{uuid.MustParse(uid)})
}

func TestSearchIDsTemplate(t *testing.T) {
	t.Parallel()

	tmpl, err := template.New("").Funcs(htmx.TemplateFuncMap()).ParseFiles(
		filepath.Join("..", "..", "view", "htmx", "layout.tmpl"),
		filepath.Join("..", "..", "view", "htmx", "searchids.tmpl"),
	)
	be.Err(t, err, nil)

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{
		"result": models.FileSlice{
			{ID: 1, UUID: null.StringFrom("c8cd0b4c-2f54-11e0-8827-cc1607e15609"), Filename: null.StringFrom("a.zip")},
			{ID: 2, UUID: null.StringFrom("c8cd0f9e-2f54-11e0-8827-cc1607e15609"), Filename: null.StringFrom("b.zip")},
		},
	})
	be.Err(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, `class="table-responsive"`))
	be.Equal(t, strings.Count(out, "<tbody>"), 1)
	be.Equal(t, strings.Count(out, "</tbody>"), 1)
}
