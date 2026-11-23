package model_test

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/null/v8"
	"github.com/nalgeon/be"
)

const missingID = int64(999999999)

func TestUpdateBoolFrom(t *testing.T) {
	t.Parallel()

	got := model.ReadmeDisable.Update(t.Context(), nil, 0, false)
	be.Err(t, got)

	tx0 := testutil.Tx(t)
	got = model.ReadmeDisable.Update(t.Context(), tx0, 0, false)
	be.Err(t, got)

	got = model.ReadmeDisable.Update(t.Context(), tx0, missingID, false)
	be.True(t, errors.Is(got, sql.ErrNoRows))

	const safeID = 1
	before, err := model.OneFile(t.Context(), tx0, safeID)
	be.Err(t, err, nil)

	tests := []struct {
		name string
		col  model.BoolFrom
		val  bool
		want int16
	}{
		{"umb false", model.EmulateUMB, false, 1},
		{"umb true", model.EmulateUMB, true, 0},
		{"ems false", model.EmulateEMS, false, 1},
		{"ems true", model.EmulateEMS, true, 0},
		{"xms false", model.EmulateXMS, false, 1},
		{"xms true", model.EmulateXMS, true, 0},
		{"broken false", model.EmulateBroken, false, 1},
		{"broken true", model.EmulateBroken, true, 0},
		{"readme false", model.ReadmeDisable, false, 1},
		{"readme true", model.ReadmeDisable, true, 0},
	}
	for _, tc := range tests {
		err := tc.col.Update(t.Context(), tx0, safeID, tc.val)
		be.Err(t, err, nil)

		after, err := model.OneFile(t.Context(), tx0, safeID)
		be.Err(t, err, nil)
		be.Equal(t, after.UUID, before.UUID)
		be.Equal(t, after.Filename, before.Filename)
		be.True(t, after.Updatedat.Valid)

		switch tc.col {
		case model.EmulateUMB:
			be.Equal(t, after.DoseeNoUmb, null.Int16From(tc.want))
		case model.EmulateEMS:
			be.Equal(t, after.DoseeNoEms, null.Int16From(tc.want))
		case model.EmulateXMS:
			be.Equal(t, after.DoseeNoXMS, null.Int16From(tc.want))
		case model.EmulateBroken:
			be.Equal(t, after.DoseeIncompatible, null.Int16From(tc.want))
		case model.ReadmeDisable:
			be.Equal(t, after.RetrotxtNoReadme, null.Int16From(tc.want))
		}
	}
}

func TestUpdateInt64From(t *testing.T) {
	t.Parallel()

	got := model.Pouet.Update(t.Context(), nil, 0, "")
	be.Err(t, got)

	tx0 := testutil.Tx(t)
	got = model.Pouet.Update(t.Context(), nil, 0, "")
	be.Err(t, got)

	const safeID = 1
	before, err := model.OneFile(t.Context(), tx0, safeID)
	be.Err(t, err, nil)

	got = model.Pouet.Update(t.Context(), tx0, missingID, "1000")
	be.True(t, errors.Is(got, sql.ErrNoRows))

	got = model.Pouet.Update(t.Context(), tx0, safeID, "abc")
	be.Err(t, got)

	// Set and clear Pouet
	got = model.Pouet.Update(t.Context(), tx0, safeID, "1000")
	be.Err(t, got, nil)
	after, err := model.OneFile(t.Context(), tx0, safeID)
	be.Err(t, err, nil)
	be.Equal(t, after.WebIDPouet, null.Int64From(1000))
	be.Equal(t, after.UUID, before.UUID)

	got = model.Pouet.Update(t.Context(), tx0, safeID, "")
	be.Err(t, got, nil)
	after, err = model.OneFile(t.Context(), tx0, safeID)
	be.Err(t, err, nil)
	be.True(t, !after.WebIDPouet.Valid)

	// Set and clear Demozoo
	got = model.Demozoo.Update(t.Context(), tx0, safeID, "2000")
	be.Err(t, got, nil)
	after, err = model.OneFile(t.Context(), tx0, safeID)
	be.Err(t, err, nil)
	be.Equal(t, after.WebIDDemozoo, null.Int64From(2000))
	be.Equal(t, after.UUID, before.UUID)

	got = model.Demozoo.Update(t.Context(), tx0, safeID, "0")
	be.Err(t, got, nil)
	after, err = model.OneFile(t.Context(), tx0, safeID)
	be.Err(t, err, nil)
	be.True(t, !after.WebIDDemozoo.Valid)
}

func TestUpdateStringFrom(t *testing.T) {
	t.Parallel()

	got := model.Colors16.Update(t.Context(), nil, -1, "")
	be.Err(t, got)

	tx0 := testutil.Tx(t)
	got = model.Colors16.Update(t.Context(), tx0, 0, "false")
	be.Err(t, got)

	got = model.Colors16.Update(t.Context(), tx0, missingID, "1000")
	be.True(t, errors.Is(got, sql.ErrNoRows))

	const safeID = 1
	before, err := model.OneFile(t.Context(), tx0, safeID)
	be.Err(t, err, nil)

	tests := []struct {
		sf  model.StringFrom
		val string
	}{
		{model.Colors16, "  pack/2000  "},
		{model.Comment, "  test comment  "},
		{model.CreatorAudio, "  musician  "},
		{model.CreatorIll, "  artist  "},
		{model.CreatorProg, "  coder  "},
		{model.CreatorText, "  writer  "},
		{model.Filename, "  test.zip  "},
		{model.GitHub, "  defacto2/server  "},
		{model.Integrity, "  abc123  "},
		{model.Magic, "  ZIP archive  "},
		{model.Platform, "  dos  "},
		{model.Relations, "  rel1;rel2  "},
		{model.Section, "  intro  "},
		{model.Sites, "  https://example.com  "},
		{model.Title, "  Test Title  "},
		{model.VirusTotal, "  https://vt.example  "},
		{model.YouTube, "  dQw4w9WgXcQ  "},
		{model.ZipContent, "  FILE_ID.DIZ  "},
	}

	for _, tc := range tests {
		err := tc.sf.Update(t.Context(), tx0, safeID, tc.val)
		be.Err(t, err, nil)

		after, err := model.OneFile(t.Context(), tx0, safeID)
		be.Err(t, err, nil)
		be.Equal(t, after.UUID, before.UUID)

		want := null.StringFrom(strings.TrimSpace(tc.val))
		switch tc.sf {
		case model.Colors16:
			be.Equal(t, after.WebID16colors, want)
		case model.Comment:
			be.Equal(t, after.Comment, want)
		case model.CreatorAudio:
			be.Equal(t, after.CreditAudio, want)
		case model.CreatorIll:
			be.Equal(t, after.CreditIllustration, want)
		case model.CreatorProg:
			be.Equal(t, after.CreditProgram, want)
		case model.CreatorText:
			be.Equal(t, after.CreditText, want)
		case model.Filename:
			be.Equal(t, after.Filename, want)
		case model.GitHub:
			be.Equal(t, after.WebIDGithub, want)
		case model.Integrity:
			be.True(t, after.FileIntegrityStrong.Valid)
			be.Equal(t, strings.TrimSpace(after.FileIntegrityStrong.String), want.String)
		case model.Magic:
			be.Equal(t, after.FileMagicType, want)
		case model.Platform:
			be.True(t, after.Platform.Valid)
			be.Equal(t, strings.TrimSpace(after.Platform.String), want.String)
		case model.Relations:
			be.Equal(t, after.ListRelations, want)
		case model.Section:
			be.True(t, after.Section.Valid)
			be.Equal(t, strings.TrimSpace(after.Section.String), want.String)
		case model.Sites:
			be.Equal(t, after.ListLinks, want)
		case model.Title:
			be.Equal(t, after.RecordTitle, want)
		case model.VirusTotal:
			be.Equal(t, after.FileSecurityAlertURL, want)
		case model.YouTube:
			be.Equal(t, after.WebIDYoutube, want)
		case model.ZipContent:
			be.Equal(t, after.FileZipContent, want)
		}
	}
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

	before, err := model.OneFile(t.Context(), tx0, safeID)
	be.Err(t, err, nil)

	got := model.UpdateMagic(t.Context(), tx0, 0, "")
	be.Err(t, got)

	got = model.UpdateMagic(t.Context(), tx0, missingID, safeMagic)
	be.True(t, errors.Is(got, sql.ErrNoRows))

	got = model.UpdateMagic(t.Context(), tx0, safeID, safeMagic)
	be.Err(t, got, nil)

	after, err := model.OneFile(t.Context(), tx0, safeID)
	be.Err(t, err, nil)
	be.Equal(t, after.FileMagicType, null.StringFrom(safeMagic))
	be.Equal(t, after.UUID, before.UUID)
	be.Equal(t, after.Filename, before.Filename)
}

func TestUpdateEmulateSfx(t *testing.T) {
	t.Parallel()

	tx0 := testutil.Tx(t)
	const safeID = 1
	const safeVal1 = "covox"
	const safeVal2 = "auto"

	before, err := model.OneFile(t.Context(), tx0, safeID)
	be.Err(t, err, nil)

	got := model.UpdateEmulateSfx(t.Context(), tx0, 0, safeVal1)
	be.Err(t, got)

	got = model.UpdateEmulateSfx(t.Context(), tx0, missingID, safeVal1)
	be.True(t, errors.Is(got, sql.ErrNoRows))

	got = model.UpdateEmulateSfx(t.Context(), tx0, safeID, "invalid value")
	be.Err(t, got)

	got = model.UpdateEmulateSfx(t.Context(), tx0, safeID, safeVal1)
	be.Err(t, got, nil)
	after, err := model.OneFile(t.Context(), tx0, safeID)
	be.Err(t, err, nil)
	be.Equal(t, after.DoseeHardwareAudio, null.StringFrom(safeVal1))
	be.Equal(t, after.UUID, before.UUID)

	got = model.UpdateEmulateSfx(t.Context(), tx0, safeID, safeVal2)
	be.Err(t, got, nil)
	after, err = model.OneFile(t.Context(), tx0, safeID)
	be.Err(t, err, nil)
	be.Equal(t, after.DoseeHardwareAudio, null.StringFrom(""))
	be.Equal(t, after.UUID, before.UUID)
}

func TestUpdateEmulateCPU(t *testing.T) {
	t.Parallel()

	tx0 := testutil.Tx(t)
	const safeID = 1
	const safeVal1 = "8086"
	const safeVal2 = "auto"

	before, err := model.OneFile(t.Context(), tx0, safeID)
	be.Err(t, err, nil)

	got := model.UpdateEmulateCPU(t.Context(), tx0, 0, safeVal1)
	be.Err(t, got)

	got = model.UpdateEmulateCPU(t.Context(), tx0, missingID, safeVal1)
	be.True(t, errors.Is(got, sql.ErrNoRows))

	got = model.UpdateEmulateCPU(t.Context(), tx0, safeID, "invalid value")
	be.Err(t, got)

	got = model.UpdateEmulateCPU(t.Context(), tx0, safeID, safeVal1)
	be.Err(t, got, nil)
	after, err := model.OneFile(t.Context(), tx0, safeID)
	be.Err(t, err, nil)
	be.Equal(t, after.DoseeHardwareCPU, null.StringFrom(safeVal1))
	be.Equal(t, after.UUID, before.UUID)

	got = model.UpdateEmulateCPU(t.Context(), tx0, safeID, safeVal2)
	be.Err(t, got, nil)
	after, err = model.OneFile(t.Context(), tx0, safeID)
	be.Err(t, err, nil)
	be.Equal(t, after.DoseeHardwareCPU, null.StringFrom(""))
	be.Equal(t, after.UUID, before.UUID)
}

func TestUpdateEmulateMachine(t *testing.T) {
	t.Parallel()

	tx0 := testutil.Tx(t)
	const safeID = 1
	const safeVal1 = "oldvbe"
	const safeVal2 = "auto"

	before, err := model.OneFile(t.Context(), tx0, safeID)
	be.Err(t, err, nil)

	got := model.UpdateEmulateMachine(t.Context(), tx0, 0, safeVal1)
	be.Err(t, got)

	got = model.UpdateEmulateMachine(t.Context(), tx0, missingID, safeVal1)
	be.True(t, errors.Is(got, sql.ErrNoRows))

	got = model.UpdateEmulateMachine(t.Context(), tx0, safeID, "invalid value")
	be.Err(t, got)

	got = model.UpdateEmulateMachine(t.Context(), tx0, safeID, safeVal1)
	be.Err(t, got, nil)
	after, err := model.OneFile(t.Context(), tx0, safeID)
	be.Err(t, err, nil)
	be.Equal(t, after.DoseeHardwareGraphic, null.StringFrom(safeVal1))
	be.Equal(t, after.UUID, before.UUID)

	got = model.UpdateEmulateMachine(t.Context(), tx0, safeID, safeVal2)
	be.Err(t, got, nil)
	after, err = model.OneFile(t.Context(), tx0, safeID)
	be.Err(t, err, nil)
	be.Equal(t, after.DoseeHardwareGraphic, null.StringFrom(""))
	be.Equal(t, after.UUID, before.UUID)
}

func TestUpdateEmulateRunProgram(t *testing.T) {
	t.Parallel()

	tx0 := testutil.Tx(t)
	const safeID = 1
	const safeVal = "  defacto2.exe  "

	before, err := model.OneFile(t.Context(), tx0, safeID)
	be.Err(t, err, nil)

	got := model.UpdateEmulateRunProgram(t.Context(), tx0, 0, safeVal)
	be.Err(t, got)

	got = model.UpdateEmulateRunProgram(t.Context(), tx0, missingID, safeVal)
	be.True(t, errors.Is(got, sql.ErrNoRows))

	got = model.UpdateEmulateRunProgram(t.Context(), tx0, safeID, safeVal)
	be.Err(t, got, nil)

	after, err := model.OneFile(t.Context(), tx0, safeID)
	be.Err(t, err, nil)
	be.Equal(t, after.DoseeRunProgram, null.StringFrom("DEFACTO2.EXE"))
	be.Equal(t, after.UUID, before.UUID)
	be.Equal(t, after.Filename, before.Filename)
}
