package htmx_test

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"html/template"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"uuid"

	"github.com/Defacto2/server/handler/fulltext"
	"github.com/Defacto2/server/handler/htmx"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/null/v8"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

func TestAreacodes(t *testing.T) {
	t.Parallel()

	be.Err(t, htmx.Areacodes(nil))

	cEmpty := testutil.NewInput(t, "/areacodes", "htmx-search", "   ")
	be.Err(t, htmx.Areacodes(cEmpty), nil)
	be.Equal(t, recorder(t, cEmpty).Code, http.StatusOK)
	be.Equal(t, recorder(t, cEmpty).Body.String(), "")

	cNone := testutil.NewInput(t, "/areacodes", "htmx-search", "<script>nope</script>")
	be.Err(t, htmx.Areacodes(cNone), nil)
	recNone := recorder(t, cNone)
	be.Equal(t, recNone.Code, http.StatusOK)
	be.True(t, strings.Contains(recNone.Body.String(), "&lt;script&gt;nope&lt;/script&gt;"))

	c := testutil.NewInput(t, "/areacodes", "htmx-search", "CA,201")
	be.Err(t, htmx.Areacodes(c), nil)
	rec := recorder(t, c)
	be.Equal(t, rec.Code, http.StatusOK)
	be.True(t, strings.Contains(rec.Body.String(), "201 - New Jersey (NJ)"))

	c2 := testutil.NewInput(t, "/areacodes", "htmx-search", "201")
	be.Err(t, htmx.Areacodes(c2), nil)
	rec2 := recorder(t, c2)
	be.True(t, !strings.Contains(rec2.Body.String(), "<br><br>"))
	be.True(t, strings.Contains(rec2.Body.String(), "201 - New Jersey (NJ)"))
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
	c.Set("zooauth", "")
	err := htmx.Demozoo.Lookup(c, db, false)
	be.Err(t, err, nil)

	c = testutil.NewInput(t,
		"/demozoo/production", "demozoo-submission", newrecord)
	err = htmx.Demozoo.Lookup(c, db, false)
	be.Err(t, err)

	c = testutil.NewInput(t,
		"/demozoo/production", "demozoo-submission", newrecord)
	c.Set("zooauth", "")
	err = htmx.Demozoo.Lookup(c, db, false)
	be.Err(t, err, nil)
}

func TestDemozooValid(t *testing.T) {
	t.Parallel()

	// Offline validation checks without dialing demozoo.org.
	_, err := htmx.ValidateDemozoo(nil, 1, false)
	be.Err(t, err)

	c := testutil.NewInput(t, "/demozoo/production", "demozoo-submission", "0")
	_, err = htmx.ValidateDemozoo(c, 0, false)
	be.Err(t, err, nil)
	be.Equal(t, recorder(t, c).Code, http.StatusNotAcceptable)

	c = testutil.NewInput(t, "/demozoo/production", "demozoo-submission", "1")
	_, err = htmx.ValidateDemozoo(c, 1, false)
	be.Err(t, err, config.ErrNoZooAuth)
}

func TestDBConnections(t *testing.T) {
	t.Parallel()

	be.Err(t, htmx.DBConnections(nil, nil))

	db := testutil.DB(t)
	c := testutil.NewContext(t, "/dbconns")
	err := htmx.DBConnections(c, db)
	be.Err(t, err, nil)
}

func TestDeleteForever(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	be.Err(t, htmx.DeleteForever(nil, nil, nil, "1"))

	cBad := testutil.NewContext(t, "/delete/forever/abc")
	be.Err(t, htmx.DeleteForever(sl, cBad, &sql.Tx{}, "abc"), nil)
	be.Equal(t, recorder(t, cBad).Code, http.StatusNotFound)

	tx := testutil.Tx(t)
	err := htmx.DeleteForever(sl, testutil.NewContext(t, "/delete/forever/0"), tx, "0")
	be.Err(t, err, nil)
	_ = tx.Rollback()
	// despite the rollback option, it is probably best not to test the delete forever func on an actual record
}

func TestHTMXOfflineGuards(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	dummyDB := &sql.DB{}

	be.Err(t, htmx.DBIndexes(nil, nil))
	be.Err(t, htmx.DBAssets(nil, config.Config{}))

	// SearchByID offline branches.
	be.Err(t, htmx.SearchByID(nil, nil, nil))

	cEmpty := testutil.NewInput(t, "/editor/search/id", "htmx-search", "   ")
	be.Err(t, htmx.SearchByID(sl, cEmpty, dummyDB), nil)
	be.Equal(t, recorder(t, cEmpty).Code, http.StatusOK)
	be.Equal(t, recorder(t, cEmpty).Body.String(), `<!-- empty search query -->`)

	cNone := testutil.NewInput(t, "/editor/search/id", "htmx-search", "xyz!")
	be.Err(t, htmx.SearchByID(sl, cNone, dummyDB), nil)
	be.Equal(t, recorder(t, cNone).Code, http.StatusOK)
	be.True(t, strings.Contains(recorder(t, cNone).Body.String(), "No matching artifacts found."))

	// SearchReleaser offline branches.
	ft := fulltext.Tidbits{}
	be.Err(t, htmx.SearchReleaser(nil, nil, nil, nil))

	cRelEmpty := testutil.NewInput(t, "/search/releaser", "htmx-search", "()")
	be.Err(t, htmx.SearchReleaser(sl, cRelEmpty, dummyDB, &ft), nil)
	be.Equal(t, recorder(t, cRelEmpty).Code, http.StatusOK)
	be.Equal(t, recorder(t, cRelEmpty).Body.String(), `<!-- empty search query -->`)

	// DataList offline branches.
	be.Err(t, htmx.DataListReleasers(nil, nil, nil, "a"))
	be.Err(t, htmx.DataListMagazines(nil, nil, nil, "a"))

	cDL := testutil.NewContext(t, "/datalist/releasers")
	be.Err(t, htmx.DataListReleasers(sl, cDL, dummyDB, "()"), nil)
	be.Equal(t, recorder(t, cDL).Code, http.StatusOK)
	be.Equal(t, recorder(t, cDL).Body.String(), "")

	cDLM := testutil.NewContext(t, "/datalist/magazines")
	be.Err(t, htmx.DataListMagazines(sl, cDLM, dummyDB, "()"), nil)
	be.Equal(t, recorder(t, cDLM).Code, http.StatusOK)
	be.Equal(t, recorder(t, cDLM).Body.String(), "")
}

func TestPings(t *testing.T) {
	t.Parallel()

	be.Err(t, htmx.Pings(nil, "http", 1323))

	// Pre-cancelled context stops the loop cleanly and writes 0 URLs summary.
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/pings", nil)
	rec := httptest.NewRecorder()
	cCancel := echo.New().NewContext(req, rec)
	be.Err(t, htmx.Pings(cCancel, "http", 1323), nil)
	be.Equal(t, rec.Code, http.StatusOK)
	be.True(t, strings.HasPrefix(rec.Header().Get(echo.HeaderContentType), "multipart/mixed; boundary="))
	be.Equal(t, rec.Header().Get("Cache-Control"), "no-cache")
	be.Equal(t, rec.Header().Get("X-Accel-Buffering"), "no")
	be.True(t, strings.Contains(rec.Body.String(), "0 URLs were pinged"))

	// Local test server exercises full multipart streaming across 200 and 404 responses.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/this-is-an-invalid-url" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	_, portStr, err := net.SplitHostPort(srv.Listener.Addr().String())
	be.Err(t, err, nil)
	port, err := strconv.Atoi(portStr)
	be.Err(t, err, nil)

	c := testutil.NewContext(t, "/pings")
	be.Err(t, htmx.Pings(c, "http", port), nil)
	recSrv := recorder(t, c)
	be.Equal(t, recSrv.Code, http.StatusOK)
	body := recSrv.Body.String()
	be.True(t, strings.Contains(body, `text-success">200</span>`))
	be.True(t, strings.Contains(body, `text-danger-emphasis">404</span>`))
	be.True(t, strings.Contains(body, "URLs were pinged"))
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
	be.True(t, len(s) > 1)

	s = htmx.Alternatives("tristar")
	be.True(t, len(s) > 1)

	s = htmx.Alternatives("razor-1911")
	be.True(t, len(s) > 1)
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

	be.Equal(t, htmx.GlobTo("searchids.tmpl"), "view/htmx/searchids.tmpl")

	empty := htmx.Templates(embed.FS{})
	be.Equal(t, len(empty), 5)
	for _, tmpl := range empty {
		be.True(t, tmpl == nil)
	}

	nilMap := htmx.Templates(nil)
	be.Equal(t, len(nilMap), 5)

	loaded := htmx.Templates(testutil.ProjectFS(t))
	be.Equal(t, len(loaded), 5)
	for name, tmpl := range loaded {
		be.True(t, tmpl != nil)
		_ = name
	}
}

func TestTemplateFuncMap(t *testing.T) {
	t.Parallel()

	fm := htmx.TemplateFuncMap()
	be.True(t, fm != nil)

	markFn, ok := fm["mark"].(func(string, string) template.HTML)
	be.True(t, ok)
	be.True(t, strings.Contains(string(markFn("hi", "say hi")), "<mark"))

	safeFn, ok := fm["safeHTML"].(func(string) template.HTML)
	be.True(t, ok)
	be.Equal(t, safeFn("<b>ok</b>"), template.HTML("<b>ok</b>"))

	tidbitFn, ok := fm["tidbits"].(func(int) template.HTML)
	be.True(t, ok)
	_ = tidbitFn(0)
}

func TestSearchResult(t *testing.T) {
	t.Parallel()

	be.Equal(t, htmx.SearchResult("0"), template.HTML(""))
	be.Equal(t, htmx.SearchResult(-1), template.HTML(""))
	be.Equal(t, htmx.SearchResult(14), template.HTML(""))
	be.Equal(t, htmx.SearchResult(0), template.HTML("1"))
	be.Equal(t, htmx.SearchResult(8), template.HTML("9"))
	be.Equal(t, htmx.SearchResult(9), template.HTML("0"))
	be.Equal(t, htmx.SearchResult(10), template.HTML("-"))
	be.Equal(t, htmx.SearchResult(11), template.HTML("="))
	be.Equal(t, htmx.SearchResult(12), template.HTML("["))
	be.Equal(t, htmx.SearchResult(13), template.HTML("]"))
}

func TestSuggestion(t *testing.T) {
	t.Parallel()

	got := htmx.Suggestion("", "", "")
	be.Equal(t, got, "suggestion type error: string")
	be.Equal(t, htmx.Suggestion("", "", nil), "suggestion type error: <nil>")

	be.Equal(t, htmx.Suggestion("a group", "", 0), "a group (0 items)")
	be.Equal(t, htmx.Suggestion("a group", "", 1), "a group (1 item)")
	be.Equal(t, htmx.Suggestion("a group", "", uint(1)), "a group (1 item)")
	be.Equal(t, htmx.Suggestion("a group", "grp", uint64(1000)), "a group, grp (1,000 items)")

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
	be.Equal(t, len(htmx.AppendUniqueReleasers(dst, nil, 5)), 1)
	be.Equal(t, len(htmx.AppendUniqueReleasers(dst, add, 1)), 1)

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
	ids, uuids, prefixes := htmx.ParseSearchIDs(
		"1, 0, -5, 9b1c6, " + uid + ", 00000000-0000-0000-0000-000000000000, c8cd0b4c, c8cd0b4c-2f54, c8cd0b4cX2f54, c8cd0b4z, c8cd, invalid",
	)
	be.Equal(t, ids, []int{1, 1})
	be.Equal(t, uuids, []uuid.UUID{uuid.MustParse(uid)})
	be.Equal(t, prefixes, []string{"c8cd0b4c", "c8cd0b4c-2f54"})
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

func TestDBIndexesTemplate(t *testing.T) {
	t.Parallel()

	tmpl, err := template.New("").Funcs(htmx.TemplateFuncMap()).ParseFiles(
		filepath.Join("..", "..", "view", "htmx", "layout.tmpl"),
		filepath.Join("..", "..", "view", "htmx", "dbindexes.tmpl"),
		filepath.Join("..", "..", "view", "app", "dbindexes.tmpl"),
	)
	be.Err(t, err, nil)

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{
		"countArtifacts":     66112,
		"countPublic":        64500,
		"countNewUpload":     12,
		"countHidden":        1600,
		"dbConnections":      "2 of 100",
		"dbIndexesRefreshed": "12:34:56",
		"dbIndexes": postgres.TableIndexStats{
			TableName:   "files",
			TableSize:   "29 MB",
			IndexesSize: "26 MB",
			TotalSize:   "55 MB",
			SeqScan:     5,
			IdxScan:     95,
			LiveTuples:  66112,
			DeadTuples:  0,
			Updates:     10,
			HotUpdates:  8,
			Indexes: []postgres.IndexStat{
				{
					Name:       "idx_16386_primary",
					Method:     "btree",
					Primary:    true,
					Valid:      true,
					Size:       "1464 kB",
					Scans:      1200,
					TupRead:    1200,
					TupFetch:   1200,
					BlocksHit:  99,
					BlocksRead: 1,
					Definition: "CREATE UNIQUE INDEX idx_16386_primary ON public.files USING btree (id)",
				},
				{
					Name:       "idx_files_desc_fts",
					Method:     "gin",
					Valid:      false,
					Partial:    true,
					Size:       "3440 kB",
					Scans:      0,
					Definition: "CREATE INDEX idx_files_desc_fts ON public.files USING gin (to_tsvector('english', comment)) WHERE (deletedat IS NULL)",
				},
			},
		},
	})
	be.Err(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, "idx_16386_primary"))
	be.True(t, strings.Contains(out, "idx_files_desc_fts"))
	be.True(t, strings.Contains(out, "INVALID"))
	be.True(t, strings.Contains(out, "95.0%"))
	be.True(t, strings.Contains(out, "80.0%"))
	be.True(t, strings.Contains(out, "2 of 100"))
	be.True(t, strings.Contains(out, "Updated 12:34:56"))
}

func TestDBAssetsTemplate(t *testing.T) {
	t.Parallel()

	tmpl, err := template.New("").Funcs(htmx.TemplateFuncMap()).ParseFiles(
		filepath.Join("..", "..", "view", "htmx", "layout.tmpl"),
		filepath.Join("..", "..", "view", "htmx", "dbassets.tmpl"),
		filepath.Join("..", "..", "view", "app", "dbassets.tmpl"),
	)
	be.Err(t, err, nil)

	// Cold loading placeholder triggers HTMX fetch on load.
	var loadingBuf bytes.Buffer
	err = tmpl.ExecuteTemplate(&loadingBuf, "content", map[string]any{
		"assetsLoading": true,
		"wdFree":        "120 GB",
		"wdPercent":     "60%",
		"wdTotal":       "200 GB",
		"osFree":        "40 GB",
		"osPercent":     "40%",
		"osTotal":       "100 GB",
		"memFree":       "8192 MB",
		"memPercent":    "50%",
		"memTotal":      "16384 MB",
		"metric0":       uint64(42),
		"metric1":       uint64(96),
		"tmpFree":       "10 GB",
		"tmpPercent":    "90%",
		"tmpTotal":      "11 GB",
		"swapFree":      "2048 MB",
		"swapPercent":   "100%",
		"swapTotal":     "2048 MB",
	})
	be.Err(t, err, nil)
	loadingOut := loadingBuf.String()
	be.True(t, strings.Contains(loadingOut, `hx-get="/editor/configurations/assets"`))
	be.True(t, strings.Contains(loadingOut, `hx-trigger="load"`))
	be.True(t, strings.Contains(loadingOut, "120 GB"))

	// Populated response renders directory rows and refresh timestamp.
	var loadedBuf bytes.Buffer
	err = tmpl.ExecuteTemplate(&loadedBuf, "content", map[string]any{
		"assetsRefreshed": "12:34:56",
		"usageDownloads":  "48.2 GB",
		"countDownloads":  "64,500",
		"usagePreviews":   "1.2 GB",
		"countPreviews":   "12,000",
		"usageThumbnails": "320 MB",
		"countThumbnails": "12,000",
		"usageExtras":     "50 MB",
		"countExtras":     "120",
		"usageOrphaned":   "0 B",
		"countOrphaned":   "0",
	})
	be.Err(t, err, nil)
	loadedOut := loadedBuf.String()
	be.True(t, strings.Contains(loadedOut, "48.2 GB"))
	be.True(t, strings.Contains(loadedOut, "Updated 12:34:56"))
}
