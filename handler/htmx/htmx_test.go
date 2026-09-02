package htmx_test

// These tests are mostly for nil checks to ensure the server doesn't panic.

// checked in Sep 26, test coverage was poor at under 10%

import (
	"embed"
	"os"
	"strconv"
	"testing"

	"github.com/Defacto2/server/handler/fulltext"
	"github.com/Defacto2/server/handler/htmx"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

func TestAreacodes(t *testing.T) {
	t.Parallel()

	c := testutil.NewForm(t,
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
	c := testutil.NewForm(t,
		"/demozoo/production", "demozoo-submission", existing)
	err := htmx.DemozooLookup(c, db, false)
	be.Err(t, err, nil)

	c = testutil.NewForm(t,
		"/demozoo/production", "demozoo-submission", newrecord)
	err = htmx.DemozooLookup(c, db, false)
	be.Err(t, err, nil)
}

func TestDemozooValid(t *testing.T) {
	t.Parallel()

	c := testutil.NewForm(t,
		"/demozoo/production", "demozoo-submission", "1")
	got, err := htmx.DemozooValid(c, false, 1)
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
	c := testutil.NewForm(t,
		"/pouet/production", "pouet-submission", "1")
	err := htmx.PouetLookup(c, db)
	be.Err(t, err, nil)

	c = testutil.NewForm(t,
		"/pouet/production", "pouet-submission", "9999")
	err = htmx.PouetLookup(c, db)
	be.Err(t, err, nil)
}

func TestPouetValid(t *testing.T) {
	t.Parallel()

	// WARN: do not run working tests over this, as the func pings the target website.
	_, err := htmx.PouetValid(nil, -1, true)
	be.Err(t, err)
}

func TestPouetSubmit(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	tx := testutil.Tx(t)
	c := testutil.NewForm(t,
		"/pouet/production", "pouet-submission", "9999")
	err := htmx.PouetSubmit(sl, c, tx, "")
	be.Err(t, err, nil)
	_ = tx.Rollback()
}

func TestSearchByID(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	db := testutil.DB(t)
	c := testutil.NewForm(t, "/editor/search", "htmx-search", "1")
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

	// TODO: this is causing race cond
	// it might be linked to fulltext.go
	return

	sl := logs.Discard()
	db := testutil.DB(t)
	c := testutil.NewForm(t, "/search/releaser", "htmx-search", "defacto2")

	ft := fulltext.Tidbits{}
	ft.New()
	err := htmx.SearchReleaser(sl, c, db, &ft) // this panics with FT
	be.Err(t, err)
}

func TestDataList(t *testing.T) {
	t.Parallel()
	return // TODO: this is causing a race condition

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

func TestHumanizeCount(t *testing.T) {
	t.Parallel()

	err := htmx.HumanizeCount(nil, nil, nil, "")
	be.Err(t, err)

	sl := logs.Discard()
	db := testutil.DB(t)
	c := testutil.NewContext(t, "/classifications")
	err = htmx.HumanizeCount(sl, c, db, "placeholder")
	be.Err(t, err, nil)
}

func TestLookupSHA384(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	db := testutil.DB(t)
	c := testutil.NewContext(t, "/classifications")
	err := htmx.LookupSHA384(sl, c, db)
	be.Err(t, err, nil)
}

func TestTransfer(t *testing.T) {
	t.Parallel()

	err := htmx.AdvancedSubmit(nil, nil, nil, "")
	be.Err(t, err)

	sl := logs.Discard()
	db := testutil.DB(t)
	c := testutil.NewContext(t, "/uploader/advanced")

	wd, err := os.Getwd()
	be.Err(t, err, nil)
	err = htmx.AdvancedSubmit(sl, c, db, dir.Directory(wd))
	be.Err(t, err, nil)
}

func TestProdSubmit(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	tx := testutil.Tx(t)
	c := testutil.NewContext(t, "/pouet/production")

	wd, err := os.Getwd()
	be.Err(t, err, nil)

	prod := htmx.Demozoo
	err = prod.Submit(sl, c, tx, dir.Directory(wd))
	be.Err(t, err, nil)
}

func TestUploadPreview(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	u := htmx.Upload{}
	err := u.ImagePreview(sl, nil)
	be.Err(t, err)

	c := testutil.NewContext(t, "/upload/preview")
	wd, err := os.Getwd()
	be.Err(t, err, nil)

	u.Preview = dir.Directory(wd)
	u.Thumbnail = dir.Directory(wd)
	err = u.ImagePreview(logs.Discard(), c)
	be.Err(t, err, nil)
}

func TestUploadReplacement(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()

	u := htmx.Upload{}
	err := u.Replacement(sl, nil, nil)
	be.Err(t, err)

	db := testutil.DB(t)
	wd, err := os.Getwd()
	be.Err(t, err, nil)
	u.Download = dir.Directory(wd)
	c := testutil.NewContext(t, "/upload/file")
	err = u.Replacement(sl, c, db)
	be.Err(t, err, nil)
}
