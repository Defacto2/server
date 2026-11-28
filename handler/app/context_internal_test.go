package app

import (
	"database/sql"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/internal/fileslice"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

type okRenderer struct {
	name string
}

func (r *okRenderer) Render(_ *echo.Context, w io.Writer, name string, _ any) error {
	r.name = name
	_, err := io.WriteString(w, name)
	return err
}

func TestPageData(t *testing.T) {
	t.Parallel()

	p := emptyFiles(echoCtx(t))
	p.SetTitle("T")
	p.SetH1("H")
	p.SetSubheading("Sub")
	p.SetLead("L")
	p.SetLeadHTML(template.HTML("<b>L</b>"))
	p.SetLogo("Logo")
	p.SetDescription("Desc")
	p.SetCanonical("f/1")
	p.SetCarousel("#car")
	p.SetOGTitle("OG")
	p.SetNoIndex(true)
	p.SetDatabaseErr(true)
	p.SetPagination(Pagination{CurrPage: 2, SumPages: 5, BaseURL: "/files/bbs"})
	p.SetStats(map[string]string{"files": "2"})
	p.SetRecords([]string{"a"})

	be.Equal(t, p["title"], "T")
	be.Equal(t, p["h1"], "H")
	be.Equal(t, p["subheading"], "Sub")
	be.Equal(t, p["lead"], any(template.HTML("<b>L</b>")))
	be.Equal(t, p["logo"], "Logo")
	be.Equal(t, p["description"], "Desc")
	be.Equal(t, p[canonical], "f/1")
	be.Equal(t, p["carousel"], "#car")
	be.Equal(t, p["ogtitle"], "OG")
	be.Equal(t, p["noindex"], true)
	be.Equal(t, p["databaseErr"], true)
	be.Equal(t, p["bbs"], false)
	be.Equal(t, p["unknownYears"], true)
}

func TestArtifactsHelpers(t *testing.T) {
	t.Parallel()

	be.True(t, zeroOk(fileslice.ForApproval.String()))
	be.True(t, zeroOk(fileslice.Deletions.String()))
	be.True(t, zeroOk(fileslice.Unwanted.String()))
	be.True(t, !zeroOk("bbs"))

	be.Equal(t, pageLimit(fileslice.OneOffs.String()), 250)
	be.Equal(t, pageLimit(fileslice.OneOffsAZ.String()), 250)
	be.Equal(t, pageLimit("bbs"), limit)

	// artifactsDesc branches
	be.Equal(t, len(artifactsDesc("not-a-match", "", 0, PageData{})), 0)
	for _, uri := range []string{
		fileslice.NewUploads.String(),
		fileslice.NewUpdates.String(),
		fileslice.Deletions.String(),
		fileslice.Unwanted.String(),
		fileslice.ForApproval.String(),
		fileslice.Oldest.String(),
		fileslice.Newest.String(),
		fileslice.OneOffs.String(),
		fileslice.OneOffsAZ.String(),
		fileslice.Sensenstahl.String(),
		"bbs",
	} {
		got := artifactsDesc(uri, "1985 - 1995", 10, PageData{})
		be.True(t, got["title"] != "")
	}

	// buildStatRows
	st := fileslice.Statistics()
	st.Record.Count = 1001
	st.Record.Bytes = 2048
	st.Record.MinYear = 1985
	st.Record.MaxYear = 1995
	st.Ansi.Count = 1000
	st.Ansi.Bytes = 2000
	st.BBS.Count = 1
	st.BBS.Bytes = 1
	summary, rows := buildStatRows(st)
	be.Equal(t, summary.SpanYears, 11)
	be.Equal(t, summary.CountFmt, "1,001")
	be.True(t, len(rows) > 0)

	// fileWStats, stats, statsByURI, artifactsTable guards
	sl := logs.Discard()
	db := &sql.DB{}
	_, err := fileWStats(t.Context(), nil, db, nil, false)
	be.Err(t, err)
	got, err := fileWStats(t.Context(), sl, db, nil, false)
	be.Err(t, err, nil)
	be.True(t, got != nil)

	_, _, err = stats(t.Context(), nil, "bbs")
	be.Err(t, err)
	d, cnt, err := stats(t.Context(), db, "not-valid-uri")
	be.Err(t, err, nil)
	be.True(t, d == nil)
	be.Equal(t, cnt, 0)

	be.Err(t, statsByURI(t.Context(), nil, "bbs", nil))
	be.Err(t, artifactsTable(t.Context(), nil, nil, nil, "bbs", 1))
}

func TestEditorHelpers(t *testing.T) {
	t.Parallel()

	for _, method := range []string{
		http.MethodGet, http.MethodPost, http.MethodPatch,
		http.MethodPut, http.MethodDelete, http.MethodHead,
	} {
		be.True(t, routeBadge(method) != "")
	}

	sec := newDepSection("T", "C", []command.ToolInfo{
		{Found: true, Warning: false},
		{Found: true, Warning: true},
		{Found: false},
	})
	be.Equal(t, sec.OK, 1)
	be.Equal(t, sec.Total, 3)

	// UUIDCache cached hit
	var cache UUIDCache
	cache.used = true
	cache.vers = model.UUIDVers{V4: 5}
	got, err := cache.Get(t.Context(), nil)
	be.Err(t, err, nil)
	be.Equal(t, got.V4, 5)
}

func TestReleasersHelpers(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		alt  string
		d    map[string]string
		want string
	}{
		{"independent", "", map[string]string{"sum": "5", years: "1990"}, "unaffiliated"},
		{"none", "", map[string]string{"sum": "2"}, "not authored by the Scene"},
		{"Razor 1911", "RZR", map[string]string{"sum": "1", years: "1991"}, "single artifact"},
		{"Fairlight", "", map[string]string{"sum": "10"}, "10 artifacts"},
		{"Fairlight", "", map[string]string{}, "collection of artifacts"},
	} {
		got := releasersDesc(tc.name, tc.alt, tc.d, PageData{})
		be.True(t, strings.Contains(got["description"].(string), tc.want))
	}

	be.True(t, releaserLead("independent", PageData{})["lead"] != "")
	be.True(t, releaserLead("none", PageData{})["lead"] != "")
	be.Equal(t, len(releaserLead("other", PageData{})), 0)

	fsys := testutil.ProjectFS(t)
	be.Equal(t, renderTidbits(nil, "no-such-group-xyz", fsys), "")
	be.True(t, strings.Contains(renderTidbits(logs.Discard(), "fairlight", fsys), "tidbit-entry"))

	_, err := releaserSum(t.Context(), nil, "fairlight")
	be.Err(t, err)
	be.Err(t, bbsHandler(nil, nil, nil, model.Prolific))
	be.Err(t, magazines(nil, nil, nil, true))
	be.Err(t, releasers(nil, nil, nil, model.Prolific))
}

func TestScenersHelpers(t *testing.T) {
	t.Parallel()

	be.Err(t, scenerPage(t.Context(), nil, nil, nil, postgres.Artist, PageData{}))
	_, err := scenerSum(t.Context(), nil, "scener")
	be.Err(t, err)

	e := echo.New()
	rend := &okRenderer{}
	e.Renderer = rend
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	be.Err(t, scenerPage(t.Context(), logs.Discard(), c, &sql.DB{}, postgres.Role("unknown"), PageData{}), nil)
	be.Equal(t, rec.Code, http.StatusServiceUnavailable)
}

func TestSearchHelpers(t *testing.T) {
	t.Parallel()

	be.Equal(t, len(cleanSearchTerms("")), 0)
	be.Equal(t, len(cleanSearchTerms(" , , ")), 0)
	got := cleanSearchTerms(" doom , quake , ")
	be.Equal(t, len(got), 2)
	be.Equal(t, got[0], "doom")
	be.Equal(t, got[1], "quake")

	sl := logs.Discard()
	be.Equal(t, Filenames.postStats(t.Context(), sl, nil, nil)[files], "no files found")
	be.Equal(t, Filenames.postStats(t.Context(), sl, nil, []string{"doom"})[files], "no files found")
	be.Equal(t, FileSearch(99).postStats(t.Context(), sl, &sql.DB{}, []string{"doom"})[files], "no files found")
}

func TestAuthHelpers(t *testing.T) {
	t.Parallel()

	be.Err(t, sessionHandler(nil, 1, nil))
	be.Err(t, sessionHandler(echoCtx(t), 1, map[string]any{"sub": "1"}))
	be.Err(t, expireCookie(nil, nil, signin, nil))

	e := echo.New()
	rend := &okRenderer{}
	e.Renderer = rend
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/signin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	be.Err(t, expireCookie(logs.Discard(), c, signin, map[string]any{}), nil)
	be.Equal(t, rec.Code, http.StatusOK)
	be.Equal(t, rend.name, signin)

	// Exercise archiveMeta without mutating global Caching under -race.
	_ = archiveMeta()
}
