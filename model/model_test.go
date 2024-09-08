// Package model_test requires an active database connection.
package model_test

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func TestValidDateIssue(t *testing.T) {
	t.Parallel()
	y, m, d := model.ValidDateIssue("", "", "")
	assert.False(t, y.Valid)
	assert.False(t, m.Valid)
	assert.False(t, d.Valid)
	y, _, _ = model.ValidDateIssue("1980", "", "")
	assert.True(t, y.Valid)
	assert.Equal(t, int16(1980), y.Int16)
	y, m, d = model.ValidDateIssue("9999", "999", "999")
	assert.False(t, y.Valid)
	assert.False(t, m.Valid)
	assert.False(t, d.Valid)
	y, m, d = model.ValidDateIssue("1980", "1", "2")
	assert.True(t, y.Valid)
	assert.Equal(t, int16(1980), y.Int16)
	assert.True(t, m.Valid)
	assert.Equal(t, int16(1), m.Int16)
	assert.True(t, d.Valid)
	assert.Equal(t, int16(2), d.Int16)
}

func TestValidFilename(t *testing.T) {
	t.Parallel()
	name := ""
	r := model.ValidFilename(name)
	assert.False(t, r.Valid)

	name = "somefile.txt"
	r = model.ValidFilename(name)
	assert.True(t, r.Valid)
	assert.Equal(t, name, r.String)

	name = strings.Repeat("a", model.LongFilename+100)
	r = model.ValidFilename(name)
	assert.True(t, r.Valid)
	assert.Len(t, r.String, model.LongFilename)
}

func TestValidFilesize(t *testing.T) {
	t.Parallel()
	size := ""
	actual0 := null.Int64From(0)
	actual100 := null.Int64From(100)
	actualN100 := null.Int64From(-100)
	i, err := model.ValidFilesize(size)
	require.NoError(t, err)
	assert.NotEqual(t, actual0, i)
	size = "100"
	i, err = model.ValidFilesize(size)
	require.NoError(t, err)
	assert.Equal(t, actual100, i)
	size = "-100"
	i, err = model.ValidFilesize(size)
	require.NoError(t, err)
	assert.Equal(t, actualN100, i)
}

func TestValidIntegrity(t *testing.T) {
	t.Parallel()
	integ := ""
	r := model.ValidIntegrity(integ)
	assert.False(t, r.Valid)
	assert.Empty(t, r.String)

	integ = "abcde"
	r = model.ValidIntegrity(integ)
	assert.False(t, r.Valid)
	assert.Empty(t, r.String)

	const valid = "8ac9e700d8d5467fb8f62c88628b1f30cbfa1d0696a81a78599af01bb913cc726a78f3817adfa557691db9ad1354df6b"
	r = model.ValidIntegrity(valid)
	assert.True(t, r.Valid)
	assert.Equal(t, valid, r.String)

	const invalid = "XXXXXX00d8d5467fb8f62c88628b1f30cbfa1d0696a81a78599af01bb913cc726a78f3817adfa557691db9ad1354df6b"
	r = model.ValidIntegrity(invalid)
	assert.False(t, r.Valid)
	assert.Empty(t, r.String)
}

func TestValidLastMod(t *testing.T) {
	t.Parallel()
	lastmod := ""
	r := model.ValidLastMod(lastmod)
	assert.False(t, r.Valid)

	lastmod = "100"
	r = model.ValidLastMod(lastmod)
	assert.False(t, r.Valid)

	oneHourAgo := time.Now().Add(-time.Hour).UnixMilli()
	lastmod = strconv.FormatInt(oneHourAgo, 10)
	r = model.ValidLastMod(lastmod)
	assert.True(t, r.Valid)
	assert.Greater(t, time.Now().UnixMilli(), r.Time.UnixMilli())

	oneHourFromNow := time.Now().Add(time.Hour).UnixMilli()
	lastmod = strconv.FormatInt(oneHourFromNow, 10)
	r = model.ValidLastMod(lastmod)
	assert.False(t, r.Valid)
}

func TestValidMagic(t *testing.T) {
	t.Parallel()
	magic := ""
	r := model.ValidMagic(magic)
	assert.False(t, r.Valid)

	magic = "100"
	r = model.ValidMagic(magic)
	assert.False(t, r.Valid)

	magic = "defacto2"
	r = model.ValidMagic(magic)
	assert.False(t, r.Valid)

	magic = "Text/HTML"
	r = model.ValidMagic(magic)
	assert.True(t, r.Valid)
	assert.Equal(t, "text/html", r.String)
}

func TestValidPlatform(t *testing.T) {
	t.Parallel()
	tag := ""
	r := model.ValidPlatform(tag)
	assert.False(t, r.Valid)

	tag = "100"
	r = model.ValidPlatform(tag)
	assert.False(t, r.Valid)

	tag = "bbs"
	r = model.ValidPlatform(tag)
	assert.False(t, r.Valid)

	tag = "Windows"
	r = model.ValidPlatform(tag)
	assert.True(t, r.Valid)
	assert.Equal(t, "windows", r.String)
}

func TestValidReleasers(t *testing.T) {
	t.Parallel()
	s1, s2 := "", ""
	r1, r2 := model.ValidReleasers(s1, s2)
	assert.False(t, r1.Valid)
	assert.False(t, r2.Valid)

	s1, s2 = "defacto2", "scene"
	r1, r2 = model.ValidReleasers(s1, s2)
	assert.True(t, r1.Valid)
	assert.True(t, r2.Valid)
	assert.Equal(t, "DEFACTO2", r1.String)
	assert.Equal(t, "SCENE", r2.String)

	// test the swapping of empty releasers
	r1, r2 = model.ValidReleasers("", "defacto2")
	assert.True(t, r1.Valid)
	assert.False(t, r2.Valid)
	assert.Equal(t, "DEFACTO2", r1.String)
	assert.Empty(t, r2.String)
}

func TestValidSceners(t *testing.T) {
	t.Parallel()
	sceners := ""
	r := model.ValidSceners(sceners)
	assert.False(t, r.Valid)

	sceners = "defacto"
	r = model.ValidSceners(sceners)
	assert.True(t, r.Valid)
	assert.Equal(t, "Defacto", r.String)

	sceners = "defacto, scener    , another person"
	r = model.ValidSceners(sceners)
	assert.True(t, r.Valid)
	assert.Equal(t, "Defacto,Scener,Another Person", r.String)

	sceners = "dëfå¢T0!"
	r = model.ValidSceners(sceners)
	assert.True(t, r.Valid)
	assert.Equal(t, "Dëfåt0", r.String)
}

func TestValidSection(t *testing.T) {
	t.Parallel()
	tag := ""
	r := model.ValidSection(tag)
	assert.False(t, r.Valid)

	tag = "100"
	r = model.ValidSection(tag)
	assert.False(t, r.Valid)

	tag = "windows"
	r = model.ValidSection(tag)
	assert.False(t, r.Valid)

	tag = "BBS"
	r = model.ValidSection(tag)
	assert.True(t, r.Valid)
	assert.Equal(t, "bbs", r.String)
}

func TestValidString(t *testing.T) {
	t.Parallel()
	s := "\n\r   \n"
	r := model.ValidString(s)
	assert.False(t, r.Valid)

	const nbsp = "\u00A0"
	r = model.ValidString(nbsp)
	assert.False(t, r.Valid)

	s = "hello world"
	r = model.ValidString(s)
	assert.True(t, r.Valid)
	assert.Equal(t, r.String, s)

	const emoji = "😃"
	r = model.ValidString(emoji)
	assert.True(t, r.Valid)
	assert.Equal(t, emoji, r.String)
}

func TestValidTitle(t *testing.T) {
	t.Parallel()
	title := ""
	r := model.ValidTitle(title)
	assert.False(t, r.Valid)

	title = "hello world"
	r = model.ValidTitle(title)
	assert.True(t, r.Valid)
	assert.Equal(t, title, r.String)

	title = strings.Repeat("a", model.ShortLimit+100)
	r = model.ValidTitle(title)
	assert.True(t, r.Valid)
	assert.Len(t, r.String, model.ShortLimit)
}

func TestValidYouTube(t *testing.T) {
	t.Parallel()
	yt := ""
	r, err := model.ValidYouTube(yt)
	require.NoError(t, err)
	assert.False(t, r.Valid)

	yt = strings.Repeat("x", model.ShortLimit+10)
	r, err = model.ValidYouTube(yt)
	require.NoError(t, err)
	assert.False(t, r.Valid)

	const invalid = "$6BuDfBIcM!"
	r, err = model.ValidYouTube(invalid)
	require.NoError(t, err)
	assert.False(t, r.Valid)

	const valid = "62BuDfBIcMo"
	r, err = model.ValidYouTube(valid)
	require.NoError(t, err)
	assert.True(t, r.Valid)
}

func TestValidNewV7(t *testing.T) {
	t.Parallel()
	now1, unid, err := model.NewV7()
	require.NoError(t, err)

	now2 := time.Now()
	diff := now2.Sub(now1).Minutes()
	const oneMinute = 1.0
	assert.LessOrEqual(t, diff, oneMinute)

	err = uuid.Validate(unid.String())
	require.NoError(t, err)
}

func TestArtifacts(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	a := model.Artifacts{}
	err := a.Public(ctx, nil)
	require.Error(t, err)
	x, err := a.ByKey(ctx, nil, -1, -1)
	assert.Nil(t, x)
	require.Error(t, err)
	x, err = a.ByNewest(ctx, nil, -1, -1)
	assert.Nil(t, x)
	require.Error(t, err)
	x, err = a.ByUpdated(ctx, nil, -1, -1)
	assert.Nil(t, x)
	require.Error(t, err)
	x, err = a.ByHidden(ctx, nil, -1, -1)
	assert.Nil(t, x)
	require.Error(t, err)
	x, err = a.ByForApproval(ctx, nil, -1, -1)
	assert.Nil(t, x)
	require.Error(t, err)
	x, err = a.ByMagicErr(ctx, nil, true)
	assert.Nil(t, x)
	require.Error(t, err)
	x, err = a.ByTextPlatform(ctx, nil)
	assert.Nil(t, x)
	require.Error(t, err)
	x, err = a.ByUnwanted(ctx, nil, -1, -1)
	assert.Nil(t, x)
	require.Error(t, err)
	x, err = a.Description(ctx, nil, nil)
	assert.Nil(t, x)
	require.Error(t, err)
	x, err = a.Filename(ctx, nil, nil)
	assert.Nil(t, x)
	require.Error(t, err)
	x, err = a.ID(ctx, nil, nil)
	assert.Nil(t, x)
	require.Error(t, err)
}

func TestCount(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	_, _, _, err := model.Counts(ctx, nil)
	require.Error(t, err)
	_, err = model.CategoryCount(ctx, nil, "")
	require.Error(t, err)
	_, err = model.CategoryByteSum(ctx, nil, "")
	require.Error(t, err)
	_, err = model.ClassificationCount(ctx, nil, "", "")
	require.Error(t, err)
	_, err = model.PlatformCount(ctx, nil, "")
	require.Error(t, err)
	_, err = model.PlatformByteSum(ctx, nil, "")
	require.Error(t, err)
	_, err = model.ReleaserByteSum(ctx, nil, "")
	require.Error(t, err)
}

func TestDelete(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	err := model.DeleteOne(ctx, nil, -1)
	require.Error(t, err)
}

func TestExists(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	_, err := model.DemozooExists(ctx, nil, -1)
	require.Error(t, err)
	_, err = model.PouetExists(ctx, nil, -1)
	require.Error(t, err)
	_, err = model.SHA384Exists(ctx, nil, nil)
	require.Error(t, err)
	_, err = model.HashExists(ctx, nil, "")
	require.Error(t, err)
}

func TestFilter(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()

	// Define a slice of structs that implement Stat and List methods
	models := []interface {
		Stat(ctx context.Context, exec boil.ContextExecutor) error
		List(ctx context.Context, exec boil.ContextExecutor, x int, y int) (models.FileSlice, error)
	}{
		&model.Advert{},
		&model.Announcement{},
		&model.Ansi{},
		&model.AnsiBrand{},
		&model.AnsiBBS{},
		&model.AnsiFTP{},
		&model.AnsiNfo{},
		&model.AnsiPack{},
		&model.BBS{},
		&model.BBStro{},
		&model.BBSImage{},
		&model.BBSText{},
		&model.Database{},
		&model.Demoscene{},
		&model.Drama{},
		&model.FTP{},
		&model.Hack{},
		&model.HowTo{},
		&model.HTML{},
		&model.Image{},
		&model.ImagePack{},
		&model.Intro{},
		&model.IntroMsDos{},
		&model.IntroWindows{},
		&model.Installer{},
		&model.Java{},
		&model.JobAdvert{},
		&model.Linux{},
		&model.Magazine{},
		&model.Macos{},
		&model.MsDosPack{},
		&model.Music{},
		&model.NewsArticle{},
		&model.Nfo{},
		&model.NfoTool{},
		&model.PDF{},
		&model.Proof{},
		&model.Restrict{},
		&model.Script{},
		&model.Standard{},
		&model.Takedown{},
		&model.Text{},
		&model.TextAmiga{},
		&model.TextApple2{},
		&model.TextAtariST{},
		&model.TextPack{},
		&model.Tool{},
		&model.TrialCrackme{},
		&model.Video{},
		&model.Windows{},
		&model.WindowsPack{},
	}
	for _, m := range models {
		err := m.Stat(ctx, nil)
		require.Error(t, err)
		_, err = m.List(ctx, nil, -1, -1)
		require.Error(t, err)
	}
}

func TestInsert(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	_, _, err := model.InsertDemozoo(ctx, nil, -1)
	require.Error(t, err)
	_, _, err = model.InsertPouet(ctx, nil, -1)
	require.Error(t, err)
	_, _, err = model.InsertUpload(ctx, nil, nil, "")
	require.Error(t, err)
}

func TestModel(t *testing.T) {
	t.Parallel()
	_, err := model.JsDosBinary(nil)
	require.Error(t, err)
	_, err = model.JsDosConfig(nil)
	require.Error(t, err)
	_, err = model.JsDosCommand(nil)
	require.Error(t, err)
}

func TestOne(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	_, err := model.One(ctx, nil, false, -1)
	require.Error(t, err)
	_, err = model.OneEditByKey(ctx, nil, "")
	require.Error(t, err)
	_, err = model.OneByUUID(ctx, nil, false, "")
	require.Error(t, err)
	_, err = model.OneFile(ctx, nil, -1)
	require.Error(t, err)
	_, err = model.OneFileByKey(ctx, nil, "")
	require.Error(t, err)
	_, _, err = model.OneDemozoo(ctx, nil, -1)
	require.Error(t, err)
	_, _, err = model.OnePouet(ctx, nil, -1)
	require.Error(t, err)
}

func TestReleaser(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	r := model.ReleaserNames{}
	err := r.Distinct(ctx, nil)
	require.Error(t, err)
	err = r.DistinctGroups(ctx, nil)
	require.Error(t, err)
	rls := model.Releasers{}
	_, err = rls.Where(ctx, nil, "")
	require.Error(t, err)
	err = rls.Limit(ctx, nil, 0, -1, -1)
	require.Error(t, err)
	err = rls.Similar(ctx, nil, 0)
	require.Error(t, err)
	err = rls.SimilarMagazine(ctx, nil, 0)
	require.Error(t, err)
	err = rls.FTP(ctx, nil)
	require.Error(t, err)
	err = rls.MagazineAZ(ctx, nil)
	require.Error(t, err)
	err = rls.Magazine(ctx, nil)
	require.Error(t, err)
	rls.Slugs()
}

func TestScener(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	s := model.Sceners{}
	err := s.Distinct(ctx, nil)
	require.Error(t, err)
	err = s.Writer(ctx, nil)
	require.Error(t, err)
	err = s.Artist(ctx, nil)
	require.Error(t, err)
	err = s.Coder(ctx, nil)
	require.Error(t, err)
	err = s.Musician(ctx, nil)
	require.Error(t, err)
	x := s.Sort()
	assert.Empty(t, x)
	var o model.Scener
	_, err = o.Where(ctx, nil, "")
	require.Error(t, err)
}

func TestSummary(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	s := model.Summary{}
	err := s.ByDescription(ctx, nil, nil)
	require.Error(t, err)
	err = s.ByFilename(ctx, nil, nil)
	require.Error(t, err)
	err = s.ByForApproval(ctx, nil)
	require.Error(t, err)
	err = s.ByHidden(ctx, nil)
	require.Error(t, err)
	err = s.ByPublic(ctx, nil)
	require.Error(t, err)
	err = s.ByReleaser(ctx, nil, "")
	require.Error(t, err)
	err = s.ByUnwanted(ctx, nil)
	require.Error(t, err)

	err = s.ByMatch(ctx, nil, "")
	require.Error(t, err)
	for uri := range s.Matches() {
		err = s.ByMatch(ctx, nil, uri)
		require.Error(t, err)
	}
}

func TestUpdateBoolFrom(t *testing.T) {
	t.Parallel()
	err := model.UpdateBoolFrom(nil, -1, -1, false)
	require.Error(t, err)
	err = model.UpdateEmulateRunProgram(nil, -1, "")
	require.Error(t, err)
	err = model.UpdateEmulateMachine(nil, -1, "")
	require.Error(t, err)
	err = model.UpdateEmulateCPU(nil, -1, "")
	require.Error(t, err)
	err = model.UpdateEmulateSfx(nil, -1, "")
	require.Error(t, err)
	err = model.UpdateInt64From(nil, -1, -1, "")
	require.Error(t, err)
	err = model.UpdateStringFrom(nil, -1, -1, "")
	require.Error(t, err)
	err = model.UpdateCreators(nil, -1, "", "", "", "")
	require.Error(t, err)
	err = model.UpdateLinks(nil, -1, "", "", "", "", "", -1, -1)
	require.Error(t, err)
	err = model.UpdateClassification(nil, -1, "", "")
	require.Error(t, err)
	err = model.UpdateDateIssued(nil, -1, "", "", "")
	require.Error(t, err)
	err = model.UpdateOffline(nil, -1)
	require.Error(t, err)
	err = model.UpdatePlatform(nil, -1, "")
	require.Error(t, err)
	err = model.UpdateOnline(nil, -1)
	require.Error(t, err)
	err = model.UpdateReleasers(nil, -1, "")
	require.Error(t, err)
	x := null.Int16From(-1)
	ctx := context.TODO()
	err = model.UpdateYMD(ctx, nil, -1, x, x, x)
	require.Error(t, err)
	err = model.UpdateMagic(ctx, nil, -1, "")
	require.Error(t, err)
	fu := model.FileUpload{}
	err = fu.Update(ctx, nil, 1)
	require.Error(t, err)
}

func TestValidate(t *testing.T) {
	t.Parallel()
	err := model.Validate(nil)
	require.Error(t, err)
}
