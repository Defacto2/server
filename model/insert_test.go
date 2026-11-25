package model_test

import (
	"net/url"
	"testing"
	"uuid"

	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/null/v8"
	"github.com/nalgeon/be"
)

func TestInsertDemozoo(t *testing.T) {
	t.Parallel()

	_, _, err := model.InsertDemozoo(t.Context(), nil, 1000)
	be.Err(t, err)

	tx := testutil.Tx(t)
	const newProd = 1000
	key, unid, err := model.InsertDemozoo(t.Context(), tx, newProd)
	be.Err(t, err, nil)
	be.True(t, key > 0)
	_, err = uuid.Parse(unid)
	be.Err(t, err, nil)
}

func TestInsertPouet(t *testing.T) {
	t.Parallel()

	_, _, err := model.InsertPouet(t.Context(), nil, 1000)
	be.Err(t, err)

	tx := testutil.Tx(t)
	const newProd = 1000
	key, unid, err := model.InsertPouet(t.Context(), tx, newProd)
	be.Err(t, err, nil)
	be.True(t, key > 0)
	_, err = uuid.Parse(unid)
	be.Err(t, err, nil)
}

func TestInsertUpload(t *testing.T) {
	t.Parallel()

	_, _, err := model.InsertUpload(t.Context(), nil, url.Values{}, "")
	be.Err(t, err)

	tx := testutil.Tx(t)
	_, _, err = model.InsertUpload(t.Context(), tx, url.Values{}, "")
	be.Err(t, err)

	values := url.Values{}
	values.Add("test-filename", "testfile")
	values.Add("test-youtube", "")
	key, unid, err := model.InsertUpload(t.Context(), tx, values, "test")
	be.Err(t, err, nil)
	be.True(t, key > 0)
	be.True(t, unid != uuid.Nil())
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
