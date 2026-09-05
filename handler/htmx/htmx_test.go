package htmx_test

// These tests are mostly for nil checks to ensure the server doesn't panic.

// checked in Sep 26, test coverage was poor at under 10%

import (
	"embed"
	"strconv"
	"testing"

	"github.com/Defacto2/server/handler/fulltext"
	"github.com/Defacto2/server/handler/htmx"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

func TestAreacodes(t *testing.T) {
	t.Parallel()

	c := testutil.NewInput(t,
		"/areacodes", "htmx-search", "CA,201")
	err := htmx.Areacodes(c)
	be.Err(t, err, nil)
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
	be.Err(t, err, nil)

	c = testutil.NewInput(t,
		"/demozoo/production", "demozoo-submission", newrecord)
	err = htmx.Demozoo.Lookup(c, db, false)
	be.Err(t, err, nil)
}

func TestDemozooValid(t *testing.T) {
	t.Parallel()

	c := testutil.NewInput(t,
		"/demozoo/production", "demozoo-submission", "1")
	got, err := htmx.ValidateDemozoo(c, 1, false)
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
	// despite the rollback option, it is probably best not to test the delete forever on an actual record
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
