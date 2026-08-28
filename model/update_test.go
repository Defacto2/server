package model_test

import (
	"testing"

	"github.com/Defacto2/server/model"
	"github.com/aarondl/null/v8"
	"github.com/nalgeon/be"
)

func TestUpdateBoolFrom(t *testing.T) {
	t.Parallel()

	got := model.ReadmeDisable.Update(t.Context(), nil, 0, false)
	be.Err(t, got)

	db := openDB(t)
	if db == nil {
		return
	}

	tx0, err := db.BeginTx(t.Context(), nil)
	be.Err(t, err, nil)

	got = model.ReadmeDisable.Update(t.Context(), tx0, 0, false)
	be.Err(t, got)

	const safeID = 1
	got = model.ReadmeDisable.Update(t.Context(), tx0, safeID, false)
	be.Err(t, got, nil)
	_ = tx0.Rollback()
}

func TestUpdateInt64From(t *testing.T) {
	t.Parallel()

	got := model.Pouet.Update(t.Context(), nil, 0, "")
	be.Err(t, got)

	db := openDB(t)
	if db == nil {
		return
	}

	tx0, err := db.BeginTx(t.Context(), nil)
	be.Err(t, err, nil)

	got = model.Pouet.Update(t.Context(), nil, 0, "")
	be.Err(t, got)

	const safeID = 1
	got = model.Pouet.Update(t.Context(), tx0, safeID, "abc")
	be.Err(t, got)

	got = model.Pouet.Update(t.Context(), tx0, safeID, "1000")
	be.Err(t, got, nil)
	_ = tx0.Rollback()
}

func TestUpdateStringFrom(t *testing.T) {
	t.Parallel()

	got := model.Colors16.Update(t.Context(), nil, -1, "")
	be.Err(t, got)

	db := openDB(t)
	if db == nil {
		return
	}

	tx0, err := db.BeginTx(t.Context(), nil)
	be.Err(t, err, nil)

	got = model.Colors16.Update(t.Context(), tx0, 0, "false")
	be.Err(t, got)

	const safeID = 1
	got = model.Colors16.Update(t.Context(), tx0, safeID, "1000")
	be.Err(t, got, nil)

	got = model.Colors16.Update(t.Context(), tx0, safeID, "")
	be.Err(t, got, nil)
	_ = tx0.Rollback()

	tx1, err := db.BeginTx(t.Context(), nil)
	be.Err(t, err, nil)
	got = model.Title.Update(t.Context(), tx1, safeID, "Hello")
	be.Err(t, got, nil)
	_ = tx1.Rollback()
}

func TestUpdateReleasers(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}

	tx0, err := db.BeginTx(t.Context(), nil)
	be.Err(t, err, nil)

	const safeUpdate = "defacto2net"
	const safeID = 1
	got := model.UpdateReleasers(t.Context(), tx0, 0, "", "")
	be.Err(t, got)

	got = model.UpdateReleasers(t.Context(), tx0, safeID, "", safeUpdate)
	be.Err(t, got)

	got = model.UpdateReleasers(t.Context(), tx0, safeID, safeUpdate, "")
	be.Err(t, got, nil)
	_ = tx0.Rollback()
}

func TestFileUploadUpdate(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}

	tx0, err := db.BeginTx(t.Context(), nil)
	be.Err(t, err, nil)

	const safeFname = "Defacto2_ISO_2007.7z"
	const safeSize = 499022368
	const safeID = 1

	fu := model.FileUpload{}
	got := fu.Update(t.Context(), db)
	be.Err(t, got)

	fu.Filename = safeFname
	fu.Filesize = safeSize
	fu.ID = safeID
	got = fu.Update(t.Context(), tx0)
	be.Err(t, got, nil)
	_ = tx0.Rollback()
}

func TestUpdateYMD(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}

	tx0, err := db.BeginTx(t.Context(), nil)
	be.Err(t, err, nil)

	const safeYear = 2007
	const safeMonth = 6
	const safeDay = 12
	const safeID = 1

	y := null.Int16From(safeYear)
	m := null.Int16From(safeMonth)
	d := null.Int16From(safeDay)
	x := null.Int16From(0)

	got := model.UpdateYMD(t.Context(), db, 0, y, m, d)
	be.Err(t, got)

	got = model.UpdateYMD(t.Context(), db, safeID, x, m, d)
	be.Err(t, got)

	got = model.UpdateYMD(t.Context(), db, safeID, y, m, d)
	be.Err(t, got, nil)
	_ = tx0.Rollback()
}

func TestUpdateYMDS(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}

	tx0, err := db.BeginTx(t.Context(), nil)
	be.Err(t, err, nil)

	const safeYear = "2007"
	const safeMonth = "6"
	const safeDay = "12"
	const safeID = 1

	y := safeYear
	m := safeMonth
	d := safeDay

	got := model.UpdateYMDS(t.Context(), db, 0, y, m, d)
	be.Err(t, got)

	got = model.UpdateYMDS(t.Context(), db, safeID, "", m, d)
	be.Err(t, got, nil) // NOTE: this behavour does not match UpdateYMD()

	got = model.UpdateYMDS(t.Context(), db, safeID, y, m, d)
	be.Err(t, got, nil)
	_ = tx0.Rollback()
}

func TestUpdateOnline(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}

	tx0, err := db.BeginTx(t.Context(), nil)
	be.Err(t, err, nil)

	const safeID = 1
	const safeStatus = true
	got := model.UpdateOnline(t.Context(), tx0, safeStatus, 0)
	be.Err(t, got)

	got = model.UpdateOnline(t.Context(), tx0, safeStatus, safeID)
	be.Err(t, got, nil)
	_ = tx0.Rollback()
}

func TestClassicationUpdate(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}

	tx0, err := db.BeginTx(t.Context(), nil)
	be.Err(t, err, nil)

	const safeID = 1
	const safePlatform = "dos"
	class := model.Classification{}
	got := class.Update(t.Context(), tx0)
	be.Err(t, got)

	class.ID = safeID
	class.Platform = safePlatform
	got = class.Update(t.Context(), tx0)
	be.Err(t, got)
	_ = tx0.Rollback()
}

func TestLinksUpdate(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}

	tx0, err := db.BeginTx(t.Context(), nil)
	be.Err(t, err, nil)

	const safeID = 1
	const safeGitHub = "defacto2"
	const safeYouTube = "62BuDfBIcMo"
	links := model.Links{}
	got := links.Update(t.Context(), tx0)
	be.Err(t, got)

	links.ID = safeID
	links.GitHub = safeGitHub
	links.YouTube = safeYouTube
	got = links.Update(t.Context(), tx0)
	be.Err(t, got, nil)

	_ = tx0.Rollback()
}

func TestCreatorsUpdate(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}

	tx0, err := db.BeginTx(t.Context(), nil)
	be.Err(t, err, nil)

	const safeID = 1
	const safeWriters = "ipggi"
	creators := model.Creators{}
	got := creators.Update(t.Context(), tx0)
	be.Err(t, got)

	creators.ID = safeID
	creators.Text = safeWriters
	got = creators.Update(t.Context(), tx0)
	be.Err(t, got, nil)

	_ = tx0.Rollback()
}
