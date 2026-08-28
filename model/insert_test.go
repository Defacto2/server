package model_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/null/v8"
	"github.com/google/uuid"
	"github.com/nalgeon/be"
)

func TestInsertDemozoo(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}
	tx0, err := db.BeginTx(t.Context(), nil)
	be.Err(t, err, nil)

	const newProd = 1000
	key, unid, got := model.InsertDemozoo(t.Context(), tx0, newProd)
	be.Err(t, got, nil)
	be.True(t, key > 0)
	be.Err(t, uuid.Validate(unid), nil)

	err = tx0.Rollback()
	be.Err(t, err, nil)
}

func TestInsertPouet(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}
	tx0, err := db.BeginTx(t.Context(), nil)
	be.Err(t, err, nil)

	const newProd = 1000
	key, unid, got := model.InsertPouet(t.Context(), tx0, newProd)
	be.Err(t, got, nil)
	be.True(t, key > 0)
	be.Err(t, uuid.Validate(unid), nil)

	err = tx0.Rollback()
	be.Err(t, err, nil)
}

func TestInsertUpload(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}
	tx0, err := db.BeginTx(t.Context(), nil)
	be.Err(t, err, nil)

	_, _, got := model.InsertUpload(t.Context(), tx0, url.Values{}, "")
	be.Err(t, got)
	err = tx0.Rollback()
	be.Err(t, err, nil)

	tx1, err := db.BeginTx(t.Context(), nil)
	be.Err(t, err, nil)

	values := url.Values{}
	values.Add("test-filename", "testfile")
	values.Add("test-youtube", "")
	key, _, got := model.InsertUpload(t.Context(), tx1, values, "test")
	be.Err(t, got, nil)
	be.True(t, key > 0)

	err = tx1.Rollback()
	be.Err(t, err, nil)
}

func TestSiteAd(t *testing.T) {
	t.Parallel()

	empty := null.StringFrom("")

	got := model.SiteAd(empty, empty)
	be.Equal(t, got, empty)

	sexe := null.StringFrom(tags.Intro.String())
	rftp := null.StringFrom("some site fTp ") // test case and white space
	got = model.SiteAd(rftp, sexe)
	be.Equal(t, got, sexe)

	sftp := null.StringFrom(tags.Ftp.String())
	snfo := null.StringFrom(tags.Nfo.String())
	got = model.SiteAd(rftp, snfo)
	be.Equal(t, got, sftp)

	rbbs := null.StringFrom("some board bBS") // test casing
	sbbs := null.StringFrom(tags.BBS.String())
	got = model.SiteAd(rbbs, snfo)
	be.Equal(t, got, sbbs)

	rsomeone := null.StringFrom("someone")
	got = model.SiteAd(rsomeone, snfo)
	be.Equal(t, got, snfo)
}

func TestValidNewV7(t *testing.T) {
	t.Parallel()

	now1, unid, err := model.NewV7()
	be.Err(t, err, nil)

	now2 := time.Now()
	diff := now2.Sub(now1).Minutes()

	const oneMinute = 1.0
	be.True(t, diff <= oneMinute)
	err = uuid.Validate(unid.String())
	be.Err(t, err, nil)
}
