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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.NotEmpty(t, x)
	assert.Empty(t, x["title"])
}

func TestArtifacts(t *testing.T) {
	t.Parallel()
	x := app.Artifacts(newContext(), nil, "", "")
	require.Error(t, x)
	x = app.Artifacts(newContext(), nil, "for-approval", "1")
	require.Error(t, x)
}

func TestArtist(t *testing.T) {
	t.Parallel()
	x := app.Artist(newContext(), nil)
	require.Error(t, x)
}

func TestBBS(t *testing.T) {
	t.Parallel()
	x := app.BBS(newContext(), nil)
	require.Error(t, x)
}

func TestChecksum(t *testing.T) {
	t.Parallel()
	x := app.Checksum(newContext(), nil, "")
	require.Error(t, x)
}

func TestCoder(t *testing.T) {
	t.Parallel()
	x := app.Coder(newContext(), nil)
	require.Error(t, x)
}

func TestConfigurations(t *testing.T) {
	t.Parallel()
	x := app.Configurations(newContext(), nil, config.Config{})
	require.Error(t, x)
}

func TestDownloadJsDos(t *testing.T) {
	t.Parallel()
	x := app.DownloadJsDos(newContext(), nil, "", "")
	require.Error(t, x)
}

func TestDownload(t *testing.T) {
	t.Parallel()
	x := app.Download(newContext(), nil, nil, "")
	require.Error(t, x)
}

func TestFTP(t *testing.T) {
	t.Parallel()
	x := app.FTP(newContext(), nil)
	require.Error(t, x)
}

func TestCategories(t *testing.T) {
	t.Parallel()
	x := app.Categories(newContext(), nil, nil, false)
	require.Error(t, x)
}

func TestDeletions(t *testing.T) {
	t.Parallel()
	x := app.Deletions(newContext(), nil, "")
	require.Error(t, x)
}

func TestUnwanted(t *testing.T) {
	t.Parallel()
	x := app.Unwanted(newContext(), nil, "")
	require.Error(t, x)
}

func TestForApproval(t *testing.T) {
	t.Parallel()
	x := app.ForApproval(newContext(), nil, "")
	require.Error(t, x)
}

func TestGetDemozooParam(t *testing.T) {
	t.Parallel()
	x := app.GetDemozooParam(newContext(), nil, "")
	require.NoError(t, x)
}

func TestGetDemozoo(t *testing.T) {
	t.Parallel()
	x := app.GetDemozoo(newContext(), nil, -1, "", "")
	require.Error(t, x)
	x = app.GetPouet(newContext(), nil, -1, "", "")
	require.Error(t, x)
}

func TestGoogleCallback(t *testing.T) {
	t.Parallel()
	x := app.GoogleCallback(newContext(), "", -1, [48]byte{})
	require.Error(t, x)
}

func TestHistory(t *testing.T) {
	t.Parallel()
	x := app.History(newContext())
	require.Error(t, x)
}

func TestIndex(t *testing.T) {
	t.Parallel()
	x := app.Index(newContext())
	require.Error(t, x)
}

func TestInline(t *testing.T) {
	t.Parallel()
	x := app.Inline(newContext(), nil, nil, "")
	require.Error(t, x)
}

func TestInterview(t *testing.T) {
	t.Parallel()
	x := app.Interview(newContext())
	require.Error(t, x)
}

func TestMagazine(t *testing.T) {
	t.Parallel()
	x := app.Magazine(newContext(), nil)
	require.Error(t, x)
}

func TestMagazineAZ(t *testing.T) {
	t.Parallel()
	x := app.MagazineAZ(newContext(), nil)
	require.Error(t, x)
}

func TestMusician(t *testing.T) {
	t.Parallel()
	x := app.Musician(newContext(), nil)
	require.Error(t, x)
}

func TestNew(t *testing.T) {
	t.Parallel()
	x := app.New(newContext())
	require.Error(t, x)
}

func TestPage404(t *testing.T) {
	t.Parallel()
	x := app.Page404(newContext(), "", "")
	require.Error(t, x)
}

func TestPlatformEdit(t *testing.T) {
	t.Parallel()
	x := app.PlatformEdit(newContext(), nil)
	require.Error(t, x)
}

func TestPlatformTagInfo(t *testing.T) {
	t.Parallel()
	x := app.PlatformTagInfo(newContext())
	require.NoError(t, x)
}

func TestPostDesc(t *testing.T) {
	t.Parallel()
	x := app.PostDesc(newContext(), nil, "")
	require.Error(t, x)
}

func TestPostFilename(t *testing.T) {
	t.Parallel()
	x := app.PostFilename(newContext(), nil)
	require.Error(t, x)
}

func TestPouetCache(t *testing.T) {
	t.Parallel()
	x := app.PouetCache(newContext(), "")
	require.NoError(t, x)
	x = app.PouetCache(newContext(), "abc")
	require.Error(t, x)
	x = app.PouetCache(newContext(), "3;1;1;1")
	require.NoError(t, x)
}

func TestProdPouet(t *testing.T) {
	t.Parallel()
	x := app.ProdPouet(newContext(), "")
	require.NoError(t, x)
	x = app.ProdPouet(newContext(), "abc")
	require.NoError(t, x)
}

func TestProdZoo(t *testing.T) {
	t.Parallel()
	x := app.ProdZoo(newContext(), "")
	require.NoError(t, x)
	x = app.ProdZoo(newContext(), "abc")
	require.NoError(t, x)
}

func TestReleaser(t *testing.T) {
	t.Parallel()
	x := app.Releaser(newContext(), nil)
	require.Error(t, x)
}

func TestReleaserAZ(t *testing.T) {
	t.Parallel()
	x := app.ReleaserAZ(newContext(), nil)
	require.Error(t, x)
}

func TestReleaser404(t *testing.T) {
	t.Parallel()
	x := app.Releaser404(newContext(), "")
	require.Error(t, x)
}

func TestReleasers(t *testing.T) {
	t.Parallel()
	x := app.Releasers(newContext(), nil, nil, "", embed.FS{})
	require.Error(t, x)
}

func TestScener(t *testing.T) {
	t.Parallel()
	x := app.Scener(newContext(), nil)
	require.Error(t, x)
}

func TestScener404(t *testing.T) {
	t.Parallel()
	x := app.Scener404(newContext(), "")
	require.Error(t, x)
}

func TestSceners(t *testing.T) {
	t.Parallel()
	x := app.Sceners(newContext(), nil, "")
	require.Error(t, x)
}

func TestSearchDesc(t *testing.T) {
	t.Parallel()
	x := app.SearchDesc(newContext())
	require.Error(t, x)
}

func TestSearchID(t *testing.T) {
	t.Parallel()
	x := app.SearchID(newContext())
	require.Error(t, x)
}

func TestSearchFile(t *testing.T) {
	t.Parallel()
	x := app.SearchFile(newContext())
	require.Error(t, x)
}

func TestSearchReleaser(t *testing.T) {
	t.Parallel()
	x := app.SearchReleaser(newContext())
	require.Error(t, x)
}

func TestSignedOut(t *testing.T) {
	t.Parallel()
	x := app.SignedOut(newContext())
	require.Error(t, x)
}

func TestSignOut(t *testing.T) {
	t.Parallel()
	x := app.SignOut(newContext())
	require.Error(t, x)
}

func TestSignin(t *testing.T) {
	t.Parallel()
	x := app.Signin(newContext(), "", "")
	require.Error(t, x)
}

func TestTagEdit(t *testing.T) {
	t.Parallel()
	x := app.TagEdit(newContext(), nil)
	require.Error(t, x)
}

func TestTagInfo(t *testing.T) {
	t.Parallel()
	x := app.TagInfo(newContext())
	require.NoError(t, x)
}

func TestThanks(t *testing.T) {
	t.Parallel()
	x := app.Thanks(newContext())
	require.Error(t, x)
}

func TestTheScene(t *testing.T) {
	t.Parallel()
	x := app.TheScene(newContext())
	require.Error(t, x)
}

func TestVotePouet(t *testing.T) {
	t.Parallel()
	x := app.VotePouet(newContext(), nil, "")
	require.NoError(t, x)
	const testNoCache = "1"
	x = app.VotePouet(newContext(), nil, testNoCache)
	require.NoError(t, x)
	const testNewCache = "1"
	x = app.VotePouet(newContext(), nil, testNewCache)
	require.NoError(t, x)
}

func TestWebsite(t *testing.T) {
	t.Parallel()
	x := app.Website(newContext(), "")
	require.Error(t, x)
}

func TestWriter(t *testing.T) {
	t.Parallel()
	x := app.Writer(newContext(), nil)
	require.Error(t, x)
}
