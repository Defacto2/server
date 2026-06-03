package app_test

// Most of these tests are for nil values to ensure there are no panics.

import (
	"context"
	"embed"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/internal/config"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

func newContext() *echo.Context {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", strings.NewReader("{}"))
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
	x := app.Artifacts(context.TODO(), nil, newContext(), nil, "", "")
	be.Err(t, x)
	x = app.Artifacts(context.TODO(), nil, newContext(), nil, "for-approval", "1")
	be.Err(t, x)
}

func TestArtist(t *testing.T) {
	t.Parallel()
	x := app.Artist(context.TODO(), nil, newContext(), nil)
	be.Err(t, x)
}

func TestBBS(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	x := app.BBS(ctx, nil, newContext(), nil)
	be.Err(t, x)
}

func TestChecksum(t *testing.T) {
	t.Parallel()
	x := app.Checksum(nil, newContext(), nil, "")
	be.Err(t, x)
}

func TestCoder(t *testing.T) {
	t.Parallel()
	x := app.Coder(context.TODO(), nil, newContext(), nil)
	be.Err(t, x)
}

func TestConfigurations(t *testing.T) {
	t.Parallel()
	x := app.Configurations(context.TODO(), nil, newContext(), nil, config.Config{})
	be.Err(t, x)
}

func TestDownloadJsDos(t *testing.T) {
	t.Parallel()
	x := app.DownloadJsDos(nil, newContext(), nil, "", "")
	be.Err(t, x)
}

func TestDownload(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	x := app.Download(ctx, nil, newContext(), nil, "")
	be.Err(t, x)
}

func TestFTP(t *testing.T) {
	t.Parallel()
	x := app.FTP(context.TODO(), nil, newContext(), nil)
	be.Err(t, x)
}

func TestCategories(t *testing.T) {
	t.Parallel()
	x := app.Categories(nil, newContext(), nil, false)
	be.Err(t, x)
}

func TestDeletions(t *testing.T) {
	t.Parallel()
	x := app.Deletions(context.TODO(), nil, newContext(), nil, "")
	be.Err(t, x)
}

func TestUnwanted(t *testing.T) {
	t.Parallel()
	x := app.Unwanted(context.TODO(), nil, newContext(), nil, "")
	be.Err(t, x)
}

func TestForApproval(t *testing.T) {
	t.Parallel()
	x := app.ForApproval(context.TODO(), nil, newContext(), nil, "")
	be.Err(t, x)
}

func TestGetDemozooParam(t *testing.T) {
	t.Parallel()
	x := app.GetDemozooParam(newContext(), nil, "")
	be.Err(t, x)
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
	ctx := context.TODO()
	x := app.GoogleCallback(ctx, nil, newContext(), "", -1, [48]byte{})
	be.Err(t, x)
}

func TestHistory(t *testing.T) {
	t.Parallel()
	x := app.History(nil, newContext())
	be.Err(t, x)
}

func TestIndex(t *testing.T) {
	t.Parallel()
	x := app.Index(nil, newContext())
	be.Err(t, x)
}

func TestInline(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	x := app.Inline(ctx, nil, newContext(), nil, "")
	be.Err(t, x)
}

func TestInterview(t *testing.T) {
	t.Parallel()
	x := app.Interview(nil, newContext())
	be.Err(t, x)
}

func TestMagazine(t *testing.T) {
	t.Parallel()
	x := app.Magazine(context.TODO(), nil, newContext(), nil)
	be.Err(t, x)
}

func TestMagazineAZ(t *testing.T) {
	t.Parallel()
	x := app.MagazineAZ(context.TODO(), nil, newContext(), nil)
	be.Err(t, x)
}

func TestMusician(t *testing.T) {
	t.Parallel()
	x := app.Musician(context.TODO(), nil, newContext(), nil)
	be.Err(t, x)
}

func TestNew(t *testing.T) {
	t.Parallel()
	x := app.New(nil, newContext())
	be.Err(t, x)
}

func TestPage404(t *testing.T) {
	t.Parallel()
	x := app.Page404(nil, newContext(), "", "")
	be.Err(t, x)
}

func TestPlatformEdit(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	x := app.PlatformEdit(ctx, nil, newContext(), nil)
	be.Err(t, x)
}

func TestPlatformTagInfo(t *testing.T) {
	t.Parallel()
	x := app.PlatformTagInfo(newContext())
	be.Err(t, x, nil)
}

func TestPostDesc(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	x := app.PostDesc(ctx, nil, newContext(), nil, "")
	be.Err(t, x)
}

func TestPostFilename(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	x := app.PostFilename(ctx, nil, newContext(), nil)
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
	x := app.Releaser(context.TODO(), nil, newContext(), nil)
	be.Err(t, x)
}

func TestReleaserAZ(t *testing.T) {
	t.Parallel()
	x := app.ReleaserAZ(context.TODO(), nil, newContext(), nil)
	be.Err(t, x)
}

func TestReleaser404(t *testing.T) {
	t.Parallel()
	x := app.Releaser404(nil, newContext(), "")
	be.Err(t, x)
}

func TestReleasers(t *testing.T) {
	t.Parallel()
	x := app.Releasers(context.TODO(), nil, newContext(), nil, "", embed.FS{})
	be.Err(t, x)
}

func TestScener(t *testing.T) {
	t.Parallel()
	x := app.Scener(context.TODO(), nil, newContext(), nil)
	be.Err(t, x)
}

func TestScener404(t *testing.T) {
	t.Parallel()
	x := app.Scener404(nil, newContext(), "")
	be.Err(t, x)
}

func TestSceners(t *testing.T) {
	t.Parallel()
	x := app.Sceners(context.TODO(), nil, newContext(), nil, "")
	be.Err(t, x)
}

func TestSearchDesc(t *testing.T) {
	t.Parallel()
	x := app.SearchDesc(nil, newContext())
	be.Err(t, x)
}

func TestSearchID(t *testing.T) {
	t.Parallel()
	x := app.SearchID(nil, newContext())
	be.Err(t, x)
}

func TestSearchFile(t *testing.T) {
	t.Parallel()
	x := app.SearchFile(nil, newContext())
	be.Err(t, x)
}

func TestSearchReleaser(t *testing.T) {
	t.Parallel()
	x := app.SearchReleaser(nil, newContext())
	be.Err(t, x)
}

func TestSignedOut(t *testing.T) {
	t.Parallel()
	x := app.SignedOut(nil, newContext())
	be.Err(t, x)
}

func TestSignOut(t *testing.T) {
	t.Parallel()
	x := app.SignOut(nil, newContext())
	be.Err(t, x)
}

func TestSignin(t *testing.T) {
	t.Parallel()
	x := app.Signin(nil, newContext(), "", "")
	be.Err(t, x)
}

func TestTagEdit(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	x := app.TagEdit(ctx, nil, newContext(), nil)
	be.Err(t, x)
}

func TestTagInfo(t *testing.T) {
	t.Parallel()
	x := app.TagInfo(newContext())
	be.Err(t, x, nil)
}

func TestThanks(t *testing.T) {
	t.Parallel()
	x := app.Thanks(nil, newContext())
	be.Err(t, x)
}

func TestTheScene(t *testing.T) {
	t.Parallel()
	x := app.TheScene(nil, newContext())
	be.Err(t, x)
}

func TestVotePouet(t *testing.T) {
	t.Parallel()
	x := app.VotePouet(nil, newContext(), "")
	be.Err(t, x)
	const testNoCache = "1"
	x = app.VotePouet(nil, newContext(), testNoCache)
	be.Err(t, x)
	const testNewCache = "1"
	x = app.VotePouet(nil, newContext(), testNewCache)
	be.Err(t, x)
}

func TestWebsite(t *testing.T) {
	t.Parallel()
	x := app.Website(nil, newContext(), "")
	be.Err(t, x)
}

func TestWriter(t *testing.T) {
	t.Parallel()
	x := app.Writer(context.TODO(), nil, newContext(), nil)
	be.Err(t, x)
}
