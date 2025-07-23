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
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
	"github.com/nalgeon/be"
)

func TestValidDateIssue(t *testing.T) {
	t.Parallel()
	y, m, d := model.ValidDateIssue("", "", "")
	be.True(t, !y.Valid)
	be.True(t, !m.Valid)
	be.True(t, !d.Valid)
	y, _, _ = model.ValidDateIssue("1980", "", "")
	be.True(t, y.Valid)
	be.Equal(t, int16(1980), y.Int16)
	y, m, d = model.ValidDateIssue("9999", "999", "999")
	be.True(t, !y.Valid)
	be.True(t, !m.Valid)
	be.True(t, !d.Valid)
	y, m, d = model.ValidDateIssue("1980", "1", "2")
	be.True(t, y.Valid)
	be.Equal(t, int16(1980), y.Int16)
	be.True(t, m.Valid)
	be.Equal(t, int16(1), m.Int16)
	be.True(t, d.Valid)
	be.Equal(t, int16(2), d.Int16)
}

func TestValidFilename(t *testing.T) {
	t.Parallel()
	name := ""
	r := model.ValidFilename(name)
	be.True(t, !r.Valid)
	name = "somefile.txt"
	r = model.ValidFilename(name)
	be.True(t, r.Valid)
	be.Equal(t, name, r.String)
	name = strings.Repeat("a", model.LongFilename+100)
	r = model.ValidFilename(name)
	be.True(t, r.Valid)
	be.True(t, len(r.String) == model.LongFilename)
}

func TestValidFilesize(t *testing.T) {
	t.Parallel()
	size := ""
	actual0 := null.Int64From(0)
	actual100 := null.Int64From(100)
	actualN100 := null.Int64From(-100)
	i, err := model.ValidFilesize(size)
	be.Err(t, err, nil)
	be.True(t, actual0 != i)
	size = "100"
	i, err = model.ValidFilesize(size)
	be.Err(t, err, nil)
	be.Equal(t, actual100, i)
	size = "-100"
	i, err = model.ValidFilesize(size)
	be.Err(t, err, nil)
	be.Equal(t, actualN100, i)
}

func TestValidIntegrity(t *testing.T) {
	t.Parallel()
	integ := ""
	r := model.ValidIntegrity(integ)
	be.True(t, !r.Valid)
	be.Equal(t, r.String, "")
	integ = "abcde"
	r = model.ValidIntegrity(integ)
	be.True(t, !r.Valid)
	be.Equal(t, r.String, "")
	const valid = "8ac9e700d8d5467fb8f62c88628b1f30cbfa1d0696a81a78599af01bb913cc726a78f3817adfa557691db9ad1354df6b"
	r = model.ValidIntegrity(valid)
	be.True(t, r.Valid)
	be.Equal(t, valid, r.String)
	const invalid = "XXXXXX00d8d5467fb8f62c88628b1f30cbfa1d0696a81a78599af01bb913cc726a78f3817adfa557691db9ad1354df6b"
	r = model.ValidIntegrity(invalid)
	be.True(t, !r.Valid)
	be.Equal(t, r.String, "")
}

func TestValidLastMod(t *testing.T) {
	t.Parallel()
	lastmod := ""
	r := model.ValidLastMod(lastmod)
	be.True(t, !r.Valid)
	lastmod = "100"
	r = model.ValidLastMod(lastmod)
	be.True(t, !r.Valid)
	oneHourAgo := time.Now().Add(-time.Hour).UnixMilli()
	lastmod = strconv.FormatInt(oneHourAgo, 10)
	r = model.ValidLastMod(lastmod)
	be.True(t, r.Valid)
	be.True(t, time.Now().UnixMilli() > r.Time.UnixMilli())
	oneHourFromNow := time.Now().Add(time.Hour).UnixMilli()
	lastmod = strconv.FormatInt(oneHourFromNow, 10)
	r = model.ValidLastMod(lastmod)
	be.True(t, !r.Valid)
}

func TestValidMagic(t *testing.T) {
	t.Parallel()
	magic := ""
	r := model.ValidMagic(magic)
	be.True(t, !r.Valid)
	magic = "100"
	r = model.ValidMagic(magic)
	be.True(t, !r.Valid)
	magic = "defacto2"
	r = model.ValidMagic(magic)
	be.True(t, !r.Valid)
	magic = "Text/HTML"
	r = model.ValidMagic(magic)
	be.True(t, r.Valid)
	be.Equal(t, "text/html", r.String)
}

func TestValidPlatform(t *testing.T) {
	t.Parallel()
	tag := ""
	r := model.ValidPlatform(tag)
	be.True(t, !r.Valid)
	tag = "100"
	r = model.ValidPlatform(tag)
	be.True(t, !r.Valid)
	tag = "bbs"
	r = model.ValidPlatform(tag)
	be.True(t, !r.Valid)
	tag = "Windows"
	r = model.ValidPlatform(tag)
	be.True(t, r.Valid)
	be.Equal(t, "windows", r.String)
}

func TestValidReleasers(t *testing.T) {
	t.Parallel()
	s1, s2 := "", ""
	r1, r2 := model.ValidReleasers(s1, s2)
	be.True(t, !r1.Valid)
	be.True(t, !r2.Valid)
	s1, s2 = "defacto2", "scene"
	r1, r2 = model.ValidReleasers(s1, s2)
	be.True(t, r1.Valid)
	be.True(t, r2.Valid)
	be.Equal(t, "DEFACTO2", r1.String)
	be.Equal(t, "SCENE", r2.String)
	// test the swapping of empty releasers
	r1, r2 = model.ValidReleasers("", "defacto2")
	be.True(t, r1.Valid)
	be.True(t, !r2.Valid)
	be.Equal(t, "DEFACTO2", r1.String)
	be.Equal(t, r2.String, "")
}

func TestValidSceners(t *testing.T) {
	t.Parallel()
	sceners := ""
	r := model.ValidSceners(sceners)
	be.True(t, !r.Valid)
	sceners = "defacto"
	r = model.ValidSceners(sceners)
	be.True(t, r.Valid)
	be.Equal(t, "Defacto", r.String)
	sceners = "defacto, scener    , another person"
	r = model.ValidSceners(sceners)
	be.True(t, r.Valid)
	be.Equal(t, "Defacto,Scener,Another Person", r.String)
	sceners = "dëfå¢T0!"
	r = model.ValidSceners(sceners)
	be.True(t, r.Valid)
	be.Equal(t, "Dëfåt0", r.String)
}

func TestValidSection(t *testing.T) {
	t.Parallel()
	tag := ""
	r := model.ValidSection(tag)
	be.True(t, !r.Valid)
	tag = "100"
	r = model.ValidSection(tag)
	be.True(t, !r.Valid)
	tag = "windows"
	r = model.ValidSection(tag)
	be.True(t, !r.Valid)
	tag = "BBS"
	r = model.ValidSection(tag)
	be.True(t, r.Valid)
	be.Equal(t, "bbs", r.String)
}

func TestValidString(t *testing.T) {
	t.Parallel()
	s := "\n\r   \n"
	r := model.ValidString(s)
	be.True(t, !r.Valid)
	const nbsp = "\u00A0"
	r = model.ValidString(nbsp)
	be.True(t, !r.Valid)
	s = "hello world"
	r = model.ValidString(s)
	be.True(t, r.Valid)
	be.Equal(t, r.String, s)
	const emoji = "😃"
	r = model.ValidString(emoji)
	be.True(t, r.Valid)
	be.Equal(t, emoji, r.String)
}

func TestValidTitle(t *testing.T) {
	t.Parallel()
	title := ""
	r := model.ValidTitle(title)
	be.True(t, !r.Valid)
	title = "hello world"
	r = model.ValidTitle(title)
	be.True(t, r.Valid)
	be.Equal(t, title, r.String)
	title = strings.Repeat("a", model.ShortLimit+100)
	r = model.ValidTitle(title)
	be.True(t, r.Valid)
	be.True(t, len(r.String) == model.ShortLimit)
}

func TestValidYouTube(t *testing.T) {
	t.Parallel()
	yt := ""
	r, err := model.ValidYouTube(yt)
	be.Err(t, err, nil)
	be.True(t, !r.Valid)
	yt = strings.Repeat("x", model.ShortLimit+10)
	r, err = model.ValidYouTube(yt)
	be.Err(t, err, nil)
	be.True(t, !r.Valid)
	const invalid = "$6BuDfBIcM!"
	r, err = model.ValidYouTube(invalid)
	be.Err(t, err, nil)
	be.True(t, !r.Valid)
	const valid = "62BuDfBIcMo"
	r, err = model.ValidYouTube(valid)
	be.Err(t, err, nil)
	be.True(t, r.Valid)
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

func TestArtifacts(t *testing.T) {
	t.Parallel()
	a := model.Artifacts{}
	err := a.Public(t.Context(), nil)
	be.Err(t, err)
	x, err := a.ByKey(t.Context(), nil, -1, -1)
	be.True(t, x == nil)
	be.Err(t, err)
	x, err = a.ByNewest(t.Context(), nil, -1, -1)
	be.True(t, x == nil)
	be.Err(t, err)
	x, err = a.ByUpdated(t.Context(), nil, -1, -1)
	be.True(t, x == nil)
	be.Err(t, err)
	x, err = a.ByHidden(t.Context(), nil, -1, -1)
	be.True(t, x == nil)
	be.Err(t, err)
	x, err = model.ByForApproval(t.Context(), nil, -1, -1)
	be.True(t, x == nil)
	be.Err(t, err)
	x, err = a.ByMagicErr(t.Context(), nil, true)
	be.True(t, x == nil)
	be.Err(t, err)
	x, err = a.ByTextPlatform(t.Context(), nil)
	be.True(t, x == nil)
	be.Err(t, err)
	x, err = a.ByUnwanted(t.Context(), nil, -1, -1)
	be.True(t, x == nil)
	be.Err(t, err)
	x, err = a.Description(t.Context(), nil, nil)
	be.True(t, x == nil)
	be.Err(t, err)
	x, err = a.Filename(t.Context(), nil, nil)
	be.True(t, x == nil)
	be.Err(t, err)
	x, err = a.ID(t.Context(), nil, nil)
	be.True(t, x == nil)
	be.Err(t, err)
}

func TestCount(t *testing.T) {
	t.Parallel()
	_, _, _, err := model.Counts(t.Context(), nil)
	be.Err(t, err)
	_, err = model.CategoryCount(t.Context(), nil, "")
	be.Err(t, err)
	_, err = model.CategoryByteSum(t.Context(), nil, "")
	be.Err(t, err)
	_, err = model.ClassificationCount(t.Context(), nil, "", "")
	be.Err(t, err)
	_, err = model.PlatformCount(t.Context(), nil, "")
	be.Err(t, err)
	_, err = model.PlatformByteSum(t.Context(), nil, "")
	be.Err(t, err)
	_, err = model.ReleaserByteSum(t.Context(), nil, "")
	be.Err(t, err)
}

func TestDelete(t *testing.T) {
	t.Parallel()
	err := model.DeleteOne(t.Context(), nil, -1)
	be.Err(t, err)
}

func TestExists(t *testing.T) {
	t.Parallel()
	_, err := model.DemozooExists(t.Context(), nil, -1)
	be.Err(t, err)
	_, err = model.PouetExists(t.Context(), nil, -1)
	be.Err(t, err)
	_, err = model.SHA384Exists(t.Context(), nil, nil)
	be.Err(t, err)
	_, err = model.HashExists(t.Context(), nil, "")
	be.Err(t, err)
}

func TestFilter(t *testing.T) {
	t.Parallel()

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
		err := m.Stat(t.Context(), nil)
		be.Err(t, err)
		_, err = m.List(t.Context(), nil, -1, -1)
		be.Err(t, err)
	}
}

func TestInsert(t *testing.T) {
	t.Parallel()
	_, _, err := model.InsertDemozoo(t.Context(), nil, -1)
	be.Err(t, err)
	_, _, err = model.InsertPouet(t.Context(), nil, -1)
	be.Err(t, err)
	_, _, err = model.InsertUpload(t.Context(), nil, nil, "")
	be.Err(t, err)
}

func TestModel(t *testing.T) {
	t.Parallel()
	_, err := model.JsDosBinary(nil)
	be.Err(t, err)
	_, err = model.JsDosConfig(nil)
	be.Err(t, err)
	_, err = model.JsDosCommand(nil)
	be.Err(t, err)
}

func TestOne(t *testing.T) {
	t.Parallel()
	_, err := model.One(t.Context(), nil, false, -1)
	be.Err(t, err)
	_, err = model.OneEditByKey(t.Context(), nil, "")
	be.Err(t, err)
	_, err = model.OneByUUID(t.Context(), nil, false, "")
	be.Err(t, err)
	_, err = model.OneFile(t.Context(), nil, -1)
	be.Err(t, err)
	_, err = model.OneFileByKey(t.Context(), nil, "")
	be.Err(t, err)
	_, _, err = model.OneDemozoo(t.Context(), nil, -1)
	be.Err(t, err)
	_, _, err = model.OnePouet(t.Context(), nil, -1)
	be.Err(t, err)
}

func TestReleaser(t *testing.T) {
	t.Parallel()
	r := model.ReleaserNames{}
	err := r.Distinct(t.Context(), nil)
	be.Err(t, err)
	err = r.DistinctGroups(t.Context(), nil)
	be.Err(t, err)
	rls := model.Releasers{}
	_, err = rls.Where(t.Context(), nil, "")
	be.Err(t, err)
	err = rls.Limit(t.Context(), nil, 0, -1, -1)
	be.Err(t, err)
	err = rls.Similar(t.Context(), nil, 0)
	be.Err(t, err)
	err = rls.SimilarMagazine(t.Context(), nil, 0)
	be.Err(t, err)
	err = rls.FTP(t.Context(), nil)
	be.Err(t, err)
	err = rls.MagazineAZ(t.Context(), nil)
	be.Err(t, err)
	err = rls.Magazine(t.Context(), nil)
	be.Err(t, err)
	rls.Slugs()
}

func TestScener(t *testing.T) {
	t.Parallel()
	s := model.Sceners{}
	err := s.Distinct(t.Context(), nil)
	be.Err(t, err)
	err = s.Writer(t.Context(), nil)
	be.Err(t, err)
	err = s.Artist(t.Context(), nil)
	be.Err(t, err)
	err = s.Coder(t.Context(), nil)
	be.Err(t, err)
	err = s.Musician(t.Context(), nil)
	be.Err(t, err)
	x := s.Sort()
	be.True(t, len(x) == 0)
	var o model.Scener
	_, err = o.Where(t.Context(), nil, "")
	be.Err(t, err)
}

func TestSummary(t *testing.T) {
	t.Parallel()
	s := model.Summary{}
	err := s.ByDescription(t.Context(), nil, nil)
	be.Err(t, err)
	err = s.ByFilename(t.Context(), nil, nil)
	be.Err(t, err)
	err = s.ByForApproval(t.Context(), nil)
	be.Err(t, err)
	err = s.ByHidden(t.Context(), nil)
	be.Err(t, err)
	err = s.ByPublic(t.Context(), nil)
	be.Err(t, err)
	err = s.ByReleaser(t.Context(), nil, "")
	be.Err(t, err)
	err = s.ByUnwanted(t.Context(), nil)
	be.Err(t, err)

	err = s.ByMatch(t.Context(), nil, "")
	be.Err(t, err)
	for uri := range s.Matches() {
		err = s.ByMatch(t.Context(), nil, uri)
		be.Err(t, err)
	}
}

func TestUpdateBoolFrom(t *testing.T) {
	t.Parallel()
	err := model.UpdateBoolFrom(nil, -1, -1, false)
	be.Err(t, err)
	err = model.UpdateEmulateRunProgram(nil, -1, "")
	be.Err(t, err)
	err = model.UpdateEmulateMachine(nil, -1, "")
	be.Err(t, err)
	err = model.UpdateEmulateCPU(nil, -1, "")
	be.Err(t, err)
	err = model.UpdateEmulateSfx(nil, -1, "")
	be.Err(t, err)
	err = model.UpdateInt64From(nil, -1, -1, "")
	be.Err(t, err)
	err = model.UpdateStringFrom(nil, -1, -1, "")
	be.Err(t, err)
	err = model.UpdateCreators(nil, -1, "", "", "", "")
	be.Err(t, err)
	err = model.UpdateLinks(nil, -1, "", "", "", "", "", -1, -1)
	be.Err(t, err)
	err = model.UpdateClassification(nil, -1, "", "")
	be.Err(t, err)
	err = model.UpdateDateIssued(nil, -1, "", "", "")
	be.Err(t, err)
	err = model.UpdateOffline(nil, -1)
	be.Err(t, err)
	err = model.UpdatePlatform(nil, -1, "")
	be.Err(t, err)
	err = model.UpdateOnline(nil, -1)
	be.Err(t, err)
	err = model.UpdateReleasers(nil, -1, "")
	be.Err(t, err)
	x := null.Int16From(-1)
	err = model.UpdateYMD(t.Context(), nil, -1, x, x, x)
	be.Err(t, err)
	err = model.UpdateMagic(t.Context(), nil, -1, "")
	be.Err(t, err)
	fu := model.FileUpload{}
	err = fu.Update(t.Context(), nil, 1)
	be.Err(t, err)
}

func TestValidate(t *testing.T) {
	t.Parallel()
	err := model.Validate(nil)
	be.Err(t, err)
}
