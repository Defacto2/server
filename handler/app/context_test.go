package app_test

// Most of these tests are for nil values to ensure there are no panics.

import (
	"embed"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/nalgeon/be"
)

func newContext() echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func TestEmpty(t *testing.T) {
	t.Parallel()
	x := app.EmptyTester(newContext())
	be.Equal(t, x["title"], "")
}

func TestArtifacts(t *testing.T) {
	t.Parallel()
	x := app.Artifacts(newContext(), nil, "", "")
	be.Err(t, x)
	x = app.Artifacts(newContext(), nil, "for-approval", "1")
	be.Err(t, x)
}

func TestArtist(t *testing.T) {
	t.Parallel()
	x := app.Artist(newContext(), nil)
	be.Err(t, x)
}

func TestBBS(t *testing.T) {
	t.Parallel()
	x := app.BBS(newContext(), nil)
	be.Err(t, x)
}

func TestChecksum(t *testing.T) {
	t.Parallel()
	x := app.Checksum(newContext(), nil, "")
	be.Err(t, x)
}

func TestCoder(t *testing.T) {
	t.Parallel()
	x := app.Coder(newContext(), nil)
	be.Err(t, x)
}

func TestConfigurations(t *testing.T) {
	t.Parallel()
	x := app.Configurations(newContext(), nil, config.Config{})
	be.Err(t, x)
}

func TestDownloadJsDos(t *testing.T) {
	t.Parallel()
	x := app.DownloadJsDos(newContext(), nil, "", "")
	be.Err(t, x)
}

func TestDownload(t *testing.T) {
	t.Parallel()
	x := app.Download(newContext(), nil, nil, "")
	be.Err(t, x)
}

func TestFTP(t *testing.T) {
	t.Parallel()
	x := app.FTP(newContext(), nil)
	be.Err(t, x)
}

func TestCategories(t *testing.T) {
	t.Parallel()
	x := app.Categories(newContext(), nil, nil, false)
	be.Err(t, x)
}

func TestDeletions(t *testing.T) {
	t.Parallel()
	x := app.Deletions(newContext(), nil, "")
	be.Err(t, x)
}

func TestUnwanted(t *testing.T) {
	t.Parallel()
	x := app.Unwanted(newContext(), nil, "")
	be.Err(t, x)
}

func TestForApproval(t *testing.T) {
	t.Parallel()
	x := app.ForApproval(newContext(), nil, "")
	be.Err(t, x)
}

func TestGetDemozooParam(t *testing.T) {
	t.Parallel()
	x := app.GetDemozooParam(newContext(), nil, "")
	be.Err(t, x, nil)
}

func TestGetDemozoo(t *testing.T) {
	t.Parallel()
	x := app.GetDemozoo(newContext(), nil, -1, "", "")
	be.Err(t, x)
	x = app.GetPouet(newContext(), nil, -1, "", "")
	be.Err(t, x)
}

func TestGoogleCallback(t *testing.T) {
	t.Parallel()
	x := app.GoogleCallback(newContext(), "", -1, [48]byte{})
	be.Err(t, x)
}

func TestHistory(t *testing.T) {
	t.Parallel()
	x := app.History(newContext())
	be.Err(t, x)
}

func TestIndex(t *testing.T) {
	t.Parallel()
	x := app.Index(newContext())
	be.Err(t, x)
}

func TestInline(t *testing.T) {
	t.Parallel()
	x := app.Inline(newContext(), nil, nil, "")
	be.Err(t, x)
}

func TestInterview(t *testing.T) {
	t.Parallel()
	x := app.Interview(newContext())
	be.Err(t, x)
}

func TestMagazine(t *testing.T) {
	t.Parallel()
	x := app.Magazine(newContext(), nil)
	be.Err(t, x)
}

func TestMagazineAZ(t *testing.T) {
	t.Parallel()
	x := app.MagazineAZ(newContext(), nil)
	be.Err(t, x)
}

func TestMusician(t *testing.T) {
	t.Parallel()
	x := app.Musician(newContext(), nil)
	be.Err(t, x)
}

func TestNew(t *testing.T) {
	t.Parallel()
	x := app.New(newContext())
	be.Err(t, x)
}

func TestPage404(t *testing.T) {
	t.Parallel()
	x := app.Page404(newContext(), "", "")
	be.Err(t, x)
}

func TestPlatformEdit(t *testing.T) {
	t.Parallel()
	x := app.PlatformEdit(newContext(), nil)
	be.Err(t, x)
}

func TestPlatformTagInfo(t *testing.T) {
	t.Parallel()
	x := app.PlatformTagInfo(newContext())
	be.Err(t, x, nil)
}

func TestPostDesc(t *testing.T) {
	t.Parallel()
	x := app.PostDesc(newContext(), nil, "")
	be.Err(t, x)
}

func TestPostFilename(t *testing.T) {
	t.Parallel()
	x := app.PostFilename(newContext(), nil)
	be.Err(t, x)
}

func TestPouetCache(t *testing.T) {
	t.Parallel()
	x := app.PouetCache(newContext(), "")
	be.Err(t, x, nil)
	x = app.PouetCache(newContext(), "abc")
	be.Err(t, x)
	x = app.PouetCache(newContext(), "3;1;1;1")
	be.Err(t, x, nil)
}

func TestProdPouet(t *testing.T) {
	t.Parallel()
	x := app.ProdPouet(newContext(), "")
	be.Err(t, x, nil)
	x = app.ProdPouet(newContext(), "abc")
	be.Err(t, x, nil)
}

func TestProdZoo(t *testing.T) {
	t.Parallel()
	x := app.ProdZoo(newContext(), "")
	be.Err(t, x, nil)
	x = app.ProdZoo(newContext(), "abc")
	be.Err(t, x, nil)
}

func TestReleaser(t *testing.T) {
	t.Parallel()
	x := app.Releaser(newContext(), nil)
	be.Err(t, x)
}

func TestReleaserAZ(t *testing.T) {
	t.Parallel()
	x := app.ReleaserAZ(newContext(), nil)
	be.Err(t, x)
}

func TestReleaser404(t *testing.T) {
	t.Parallel()
	x := app.Releaser404(newContext(), "")
	be.Err(t, x)
}

func TestReleasers(t *testing.T) {
	t.Parallel()
	x := app.Releasers(newContext(), nil, nil, "", embed.FS{})
	be.Err(t, x)
}

func TestScener(t *testing.T) {
	t.Parallel()
	x := app.Scener(newContext(), nil)
	be.Err(t, x)
}

func TestScener404(t *testing.T) {
	t.Parallel()
	x := app.Scener404(newContext(), "")
	be.Err(t, x)
}

func TestSceners(t *testing.T) {
	t.Parallel()
	x := app.Sceners(newContext(), nil, "")
	be.Err(t, x)
}

func TestSearchDesc(t *testing.T) {
	t.Parallel()
	x := app.SearchDesc(newContext())
	be.Err(t, x)
}

func TestSearchID(t *testing.T) {
	t.Parallel()
	x := app.SearchID(newContext())
	be.Err(t, x)
}

func TestSearchFile(t *testing.T) {
	t.Parallel()
	x := app.SearchFile(newContext())
	be.Err(t, x)
}

func TestSearchReleaser(t *testing.T) {
	t.Parallel()
	x := app.SearchReleaser(newContext())
	be.Err(t, x)
}

func TestSignedOut(t *testing.T) {
	t.Parallel()
	x := app.SignedOut(newContext())
	be.Err(t, x)
}

func TestSignOut(t *testing.T) {
	t.Parallel()
	x := app.SignOut(newContext())
	be.Err(t, x)
}

func TestSignin(t *testing.T) {
	t.Parallel()
	x := app.Signin(newContext(), "", "")
	be.Err(t, x)
}

func TestTagEdit(t *testing.T) {
	t.Parallel()
	x := app.TagEdit(newContext(), nil)
	be.Err(t, x)
}

func TestTagInfo(t *testing.T) {
	t.Parallel()
	x := app.TagInfo(newContext())
	be.Err(t, x, nil)
}

func TestThanks(t *testing.T) {
	t.Parallel()
	x := app.Thanks(newContext())
	be.Err(t, x)
}

func TestTheScene(t *testing.T) {
	t.Parallel()
	x := app.TheScene(newContext())
	be.Err(t, x)
}

func TestVotePouet(t *testing.T) {
	t.Parallel()
	x := app.VotePouet(newContext(), nil, "")
	be.Err(t, x, nil)
	const testNoCache = "1"
	x = app.VotePouet(newContext(), nil, testNoCache)
	be.Err(t, x, nil)
	const testNewCache = "1"
	x = app.VotePouet(newContext(), nil, testNewCache)
	be.Err(t, x, nil)
}

func TestWebsite(t *testing.T) {
	t.Parallel()
	x := app.Website(newContext(), "")
	be.Err(t, x)
}

func TestWriter(t *testing.T) {
	t.Parallel()
	x := app.Writer(newContext(), nil)
	be.Err(t, x)
}
