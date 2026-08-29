package model_test

import (
	"testing"

	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/null/v8"
	"github.com/nalgeon/be"
)

func TestUpdateBoolFrom(t *testing.T) {
	t.Parallel()

	got := model.ReadmeDisable.Update(t.Context(), nil, 0, false)
	be.Err(t, got)

	tx0 := testutil.Tx(t)
	got = model.ReadmeDisable.Update(t.Context(), tx0, 0, false)
	be.Err(t, got)

	const safeID = 1
	got = model.ReadmeDisable.Update(t.Context(), tx0, safeID, false)
	be.Err(t, got, nil)
}

func TestUpdateInt64From(t *testing.T) {
	t.Parallel()

	got := model.Pouet.Update(t.Context(), nil, 0, "")
	be.Err(t, got)

	tx0 := testutil.Tx(t)
	got = model.Pouet.Update(t.Context(), nil, 0, "")
	be.Err(t, got)

	const safeID = 1
	got = model.Pouet.Update(t.Context(), tx0, safeID, "abc")
	be.Err(t, got)

	got = model.Pouet.Update(t.Context(), tx0, safeID, "1000")
	be.Err(t, got, nil)
}

func TestUpdateStringFrom(t *testing.T) {
	t.Parallel()

	got := model.Colors16.Update(t.Context(), nil, -1, "")
	be.Err(t, got)

	tx0 := testutil.Tx(t)
	got = model.Colors16.Update(t.Context(), tx0, 0, "false")
	be.Err(t, got)

	const safeID = 1
	got = model.Colors16.Update(t.Context(), tx0, safeID, "1000")
	be.Err(t, got, nil)

	got = model.Colors16.Update(t.Context(), tx0, safeID, "")
	be.Err(t, got, nil)
}

func TestTitleUpdate(t *testing.T) {
	t.Parallel()

	const safeID = 1
	tx1 := testutil.Tx(t)
	got := model.Title.Update(t.Context(), tx1, safeID, "Hello")
	be.Err(t, got, nil)
}

func TestUpdateReleasers(t *testing.T) {
	t.Parallel()

	tx0 := testutil.Tx(t)
	const safeUpdate = "defacto2net"
	const safeID = 1
	got := model.UpdateReleasers(t.Context(), tx0, 0, "", "")
	be.Err(t, got)

	got = model.UpdateReleasers(t.Context(), tx0, safeID, "", safeUpdate)
	be.Err(t, got)

	got = model.UpdateReleasers(t.Context(), tx0, safeID, safeUpdate, "")
	be.Err(t, got, nil)
}

func TestFileUploadUpdate(t *testing.T) {
	t.Parallel()

	tx0 := testutil.Tx(t)
	const safeFname = "Defacto2_ISO_2007.7z"
	const safeSize = 499022368
	const safeID = 1

	fu := model.FileUpload{}
	got := fu.Update(t.Context(), tx0)
	be.Err(t, got)

	fu.Filename = safeFname
	fu.Filesize = safeSize
	fu.ID = safeID
	got = fu.Update(t.Context(), tx0)
	be.Err(t, got, nil)
}

func TestUpdateYMD(t *testing.T) {
	t.Parallel()

	tx0 := testutil.Tx(t)
	const safeYear = 2007
	const safeMonth = 6
	const safeDay = 12
	const safeID = 1

	y := null.Int16From(safeYear)
	m := null.Int16From(safeMonth)
	d := null.Int16From(safeDay)
	x := null.Int16From(0)

	got := model.UpdateYMD(t.Context(), tx0, 0, y, m, d)
	be.Err(t, got)

	got = model.UpdateYMD(t.Context(), tx0, safeID, x, m, d)
	be.Err(t, got)

	got = model.UpdateYMD(t.Context(), tx0, safeID, y, m, d)
	be.Err(t, got, nil)
}

func TestUpdateYMDS(t *testing.T) {
	t.Parallel()

	tx0 := testutil.Tx(t)
	const safeYear = "2007"
	const safeMonth = "6"
	const safeDay = "12"
	const safeID = 1

	y := safeYear
	m := safeMonth
	d := safeDay

	got := model.UpdateYMDS(t.Context(), tx0, 0, y, m, d)
	be.Err(t, got)

	got = model.UpdateYMDS(t.Context(), tx0, safeID, "", m, d)
	be.Err(t, got, nil) // NOTE: this behavour does not match UpdateYMD()

	got = model.UpdateYMDS(t.Context(), tx0, safeID, y, m, d)
	be.Err(t, got, nil)
}

func TestUpdateOnline(t *testing.T) {
	t.Parallel()

	tx0 := testutil.Tx(t)
	const safeID = 1
	const safeStatus = true
	got := model.UpdateOnline(t.Context(), tx0, safeStatus, 0)
	be.Err(t, got)

	got = model.UpdateOnline(t.Context(), tx0, safeStatus, safeID)
	be.Err(t, got, nil)
}

func TestClassicationUpdate(t *testing.T) {
	t.Parallel()

	tx0 := testutil.Tx(t)
	const safeID = 1
	const safePlatform = "dos"
	class := model.Classification{}
	got := class.Update(t.Context(), tx0)
	be.Err(t, got)

	class.ID = safeID
	class.Platform = safePlatform
	got = class.Update(t.Context(), tx0)
	be.Err(t, got)
}

func TestLinksUpdate(t *testing.T) {
	t.Parallel()

	tx0 := testutil.Tx(t)
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
}

func TestCreatorsUpdate(t *testing.T) {
	t.Parallel()

	tx0 := testutil.Tx(t)
	const safeID = 1
	const safeWriters = "ipggi"
	creators := model.Creators{}
	got := creators.Update(t.Context(), tx0)
	be.Err(t, got)

	creators.ID = safeID
	creators.Text = safeWriters
	got = creators.Update(t.Context(), tx0)
	be.Err(t, got, nil)
}

func TestUpdateMagic(t *testing.T) {
	t.Parallel()

	tx0 := testutil.Tx(t)
	const safeID = 1
	const safeMagic = "7z compress archive"
	got := model.UpdateMagic(t.Context(), tx0, 0, "")
	be.Err(t, got)

	got = model.UpdateMagic(t.Context(), tx0, safeID, safeMagic)
	be.Err(t, got, nil)
}

func TestUpdateEmulateSfx(t *testing.T) {
	t.Parallel()

	tx0 := testutil.Tx(t)
	const safeID = 1
	const safeVal1 = "covox"
	const safeVal2 = "auto"
	got := model.UpdateEmulateSfx(t.Context(), tx0, 0, "")
	be.Err(t, got)

	got = model.UpdateEmulateSfx(t.Context(), tx0, safeID, "invalid value")
	be.Err(t, got)

	got = model.UpdateEmulateSfx(t.Context(), tx0, safeID, safeVal1)
	be.Err(t, got, nil)

	got = model.UpdateEmulateSfx(t.Context(), tx0, safeID, safeVal2)
	be.Err(t, got, nil)
}

func TestUpdateEmulateCPU(t *testing.T) {
	t.Parallel()

	tx0 := testutil.Tx(t)
	const safeID = 1
	const safeVal1 = "8086"
	const safeVal2 = "auto"
	got := model.UpdateEmulateSfx(t.Context(), tx0, 0, "")
	be.Err(t, got)

	got = model.UpdateEmulateCPU(t.Context(), tx0, safeID, "invalid value")
	be.Err(t, got)

	got = model.UpdateEmulateCPU(t.Context(), tx0, safeID, safeVal1)
	be.Err(t, got, nil)

	got = model.UpdateEmulateCPU(t.Context(), tx0, safeID, safeVal2)
	be.Err(t, got, nil)
}

func TestUpdateEmulateMachine(t *testing.T) {
	t.Parallel()

	tx0 := testutil.Tx(t)
	const safeID = 1
	const safeVal1 = "oldvbe"
	const safeVal2 = "auto"
	got := model.UpdateEmulateSfx(t.Context(), tx0, 0, "")
	be.Err(t, got)

	got = model.UpdateEmulateMachine(t.Context(), tx0, safeID, "invalid value")
	be.Err(t, got)

	got = model.UpdateEmulateMachine(t.Context(), tx0, safeID, safeVal1)
	be.Err(t, got, nil)

	got = model.UpdateEmulateMachine(t.Context(), tx0, safeID, safeVal2)
	be.Err(t, got, nil)
}

func TestUpdateEmulateRunProgram(t *testing.T) {
	t.Parallel()

	tx0 := testutil.Tx(t)
	const safeID = 1
	const safeVal = "defacto2.exe"

	got := model.UpdateEmulateRunProgram(t.Context(), tx0, safeID, safeVal)
	be.Err(t, got, nil)
}
