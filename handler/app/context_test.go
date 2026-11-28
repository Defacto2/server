package app_test

import (
	"database/sql"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

func newPostCtx(
	t *testing.T, target, contentType, body string, rend echo.Renderer,
) (*echo.Context, *httptest.ResponseRecorder) {
	t.Helper()

	e := echo.New()
	if rend != nil {
		e.Renderer = rend
	}
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, target, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set(echo.HeaderContentType, contentType)
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestEmpty(t *testing.T) {
	t.Parallel()

	x := app.EmptyTester(testutil.NewContext(t, ""))
	be.Equal(t, x["title"], "")
	be.Equal(t, x["uploader"], true)
	be.Equal(t, x["noindex"], false)
	be.Equal(t, x["readonlymode"], true)
	be.Equal(t, x["xxxxxxxx"], nil)
}

func TestStaticPages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tmpl string
		fn   func(*slog.Logger, *echo.Context) error
	}{
		{"APIInfo", "api-info", app.APIInfo},
		{"Apps", "apps", app.Apps},
		{"Areacodes", "areacodes", app.Areacodes},
		{"BrokenTexts", "brokentexts", app.BrokenTexts},
		{"Compression", "compression", app.Compression},
		{"Credits", "credits", app.Credits},
		{"Dependencies", "dependencies", app.Dependencies},
		{"Fixes", "fixes", app.Fixes},
		{"History", "history", app.History},
		{"Index", "index", app.Index},
		{"Interview", "interview", app.Interview},
		{"New", "new", app.New},
		{"Policy", "policy", app.Policy},
		{"Terms", "terms", app.Terms},
		{"Thanks", "thanks", app.Thanks},
		{"TheScene", "thescene", app.TheScene},
		{"Titles", "titles", app.Titles},
		{"SearchDesc", "searchpost", app.SearchDesc},
		{"SearchID", "searchhtmx", app.SearchID},
		{"SearchFile", "searchpost", app.SearchFile},
		{"SearchReleaser", "searchhtmx", app.SearchReleaser},
		{"SignOut", "signout", app.SignOut},
	}

	sl := logs.Discard()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, _ := newGetCtx(t, "/", nil)
			be.Err(t, tc.fn(nil, ctx))
			be.Err(t, tc.fn(sl, nil))
			be.Err(t, tc.fn(sl, ctx))

			rend := &stubRenderer{}
			ctx, rec := newGetCtx(t, "/", rend)
			be.Err(t, tc.fn(sl, ctx), nil)
			be.Equal(t, rec.Code, http.StatusOK)
			be.Equal(t, rend.name, tc.tmpl)
		})
	}
}

func TestWebsite(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	ctx, _ := newGetCtx(t, "/websites", nil)
	be.Err(t, app.Website(nil, ctx, ""))
	be.Err(t, app.Website(sl, nil, ""))

	for _, tc := range []struct {
		name string
		open string
		code int
		tmpl string
	}{
		{"all open", "", http.StatusOK, "websites"},
		{"single section", "youtube", http.StatusOK, "websites"},
		{"hide all", "hide", http.StatusOK, "websites"},
		{"unknown section", "does-not-exist", http.StatusNotFound, "status"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rend := &stubRenderer{}
			c, rec := newGetCtx(t, "/websites", rend)
			be.Err(t, app.Website(sl, c, tc.open), nil)
			be.Equal(t, rec.Code, tc.code)
			be.Equal(t, rend.name, tc.tmpl)
		})
	}
}

func TestRoutes(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	c, _ := newGetCtx(t, "/editor/routes", nil)

	be.Err(t, app.Routes(nil, c, nil))
	be.Err(t, app.Routes(sl, nil, nil))
	be.Err(t, app.Routes(sl, c, echo.Routes{{Method: http.MethodGet, Path: "/"}}))

	rend := &stubRenderer{}
	c, rec := newGetCtx(t, "/editor/routes", rend)
	be.Err(t, app.Routes(sl, c, echo.Routes{{Method: http.MethodGet, Path: "/"}}), nil)
	be.Equal(t, rec.Code, http.StatusOK)
	be.Equal(t, rend.name, "routes")
}

func TestAuth(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()

	t.Run("GoogleCallback", func(t *testing.T) {
		t.Parallel()

		c, _ := newGetCtx(t, "/google/callback", nil)
		be.Err(t, app.GoogleCallback(nil, c, "cid", 10, [48]byte{}))
		be.Err(t, app.GoogleCallback(sl, nil, "cid", 10, [48]byte{}))

		// Missing CSRF cookie -> 400
		rend := &stubRenderer{}
		c, rec := newPostCtx(t, "/google/callback", echo.MIMEApplicationForm, "", rend)
		be.Err(t, app.GoogleCallback(sl, c, "cid", 10, [48]byte{}), nil)
		be.Equal(t, rec.Code, http.StatusBadRequest)
		be.Equal(t, rend.name, "status")

		// Mismatched CSRF cookie vs form token -> 400
		rend = &stubRenderer{}
		form := url.Values{"g_csrf_token": {"wrong"}}.Encode()
		c, rec = newPostCtx(t, "/google/callback", echo.MIMEApplicationForm, form, rend)
		c.Request().AddCookie(&http.Cookie{Name: "g_csrf_token", Value: "expected"}) //nolint:gosec
		be.Err(t, app.GoogleCallback(sl, c, "cid", 10, [48]byte{}), nil)
		be.Equal(t, rec.Code, http.StatusBadRequest)
	})

	t.Run("SignedOut", func(t *testing.T) {
		t.Parallel()

		c, _ := newGetCtx(t, "/signedout", nil)
		be.Err(t, app.SignedOut(nil, c))
		be.Err(t, app.SignedOut(sl, nil))

		rend := &stubRenderer{}
		c, rec := newGetCtx(t, "/signedout", rend)
		be.Err(t, app.SignedOut(sl, c), nil)
		be.Equal(t, rec.Code, http.StatusBadRequest)
	})

	t.Run("Signin", func(t *testing.T) {
		t.Parallel()

		c, _ := newGetCtx(t, "/signin", nil)
		be.Err(t, app.Signin(nil, c, "cid", nil))
		be.Err(t, app.Signin(sl, nil, "cid", nil))

		rend := &stubRenderer{}
		c, rec := newGetCtx(t, "/signin", rend)
		be.Err(t, app.Signin(sl, c, "client-id", []byte("nonce123")), nil)
		be.Equal(t, rec.Code, http.StatusOK)
		be.Equal(t, rend.name, "signin")
	})
}

func TestRemote(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()

	t.Run("PouetCache", func(t *testing.T) {
		t.Parallel()

		be.Err(t, app.PouetCache(nil, "4.5;1;2;3"))

		c, _ := newGetCtx(t, "/", nil)
		be.Err(t, app.PouetCache(c, ""), app.ErrCorrupt)
		be.Err(t, app.PouetCache(c, "1;2;3"), app.ErrCorrupt)
		be.Err(t, app.PouetCache(c, "bad;1;2;3"))
		be.Err(t, app.PouetCache(c, "4.5;bad;2;3"))
		be.Err(t, app.PouetCache(c, "4.5;1;bad;3"))
		be.Err(t, app.PouetCache(c, "4.5;1;2;bad"))

		c, rec := newGetCtx(t, "/", nil)
		be.Err(t, app.PouetCache(c, "4.5;-2;10;1"), nil)
		be.Equal(t, rec.Code, http.StatusOK)
		be.True(t, strings.Contains(rec.Body.String(), `"stars":4.5`))
	})

	t.Run("ProdPouet", func(t *testing.T) {
		t.Parallel()

		be.Err(t, app.ProdPouet(nil, "1"))
		c, rec := newGetCtx(t, "/", nil)
		be.Err(t, app.ProdPouet(c, "not-an-int"), nil)
		be.Equal(t, rec.Code, http.StatusNotFound)
	})

	t.Run("ProdZoo", func(t *testing.T) {
		t.Parallel()

		be.Err(t, app.ProdZoo(nil, "1"))

		c, rec := newGetCtx(t, "/", nil)
		be.Err(t, app.ProdZoo(c, "not-an-int"), nil)
		be.Equal(t, rec.Code, http.StatusNotFound)

		c, _ = newGetCtx(t, "/", nil)
		be.Err(t, app.ProdZoo(c, "123"), config.ErrNoZooAuth)

		// Invalid Demozoo ID (0) returns code=0 from prod.Get; must return 502 instead of panicking.
		c, rec = newGetCtx(t, "/", nil)
		c.Set(config.ZooAuth, "token")
		be.Err(t, app.ProdZoo(c, "0"), nil)
		be.Equal(t, rec.Code, http.StatusBadGateway)
	})

	t.Run("TestZoo", func(t *testing.T) {
		t.Parallel()

		be.Err(t, app.TestZoo(nil))
		c, rec := newGetCtx(t, "/", nil)
		be.Err(t, app.TestZoo(c), nil)
		be.Equal(t, rec.Code, http.StatusInternalServerError)
	})

	t.Run("VotePouet", func(t *testing.T) {
		t.Parallel()

		c, _ := newGetCtx(t, "/", nil)
		be.Err(t, app.VotePouet(nil, c, "1"))
		be.Err(t, app.VotePouet(sl, nil, "1"))

		c, rec := newGetCtx(t, "/", nil)
		be.Err(t, app.VotePouet(sl, c, "not-an-int"), nil)
		be.Equal(t, rec.Code, http.StatusNotFound)
	})

	t.Run("GetDemozooParam", func(t *testing.T) {
		t.Parallel()

		db := &sql.DB{}
		c, _ := newGetCtx(t, "/", nil)
		be.Err(t, app.GetDemozooParam(nil, c, db, "x"))
		be.Err(t, app.GetDemozooParam(sl, nil, db, "x"))
		be.Err(t, app.GetDemozooParam(sl, c, nil, "x"))

		// Non-numeric :id param -> 400 JSON
		c, rec := newGetCtx(t, "/", nil)
		c.SetPathValues(echo.PathValues{{Name: "id", Value: "abc"}})
		be.Err(t, app.GetDemozooParam(sl, c, db, "x"), nil)
		be.Equal(t, rec.Code, http.StatusBadRequest)
		be.True(t, strings.Contains(rec.Body.String(), "numeric"))

		// Valid :id but invalid unid query param -> 400 JSON
		c, rec = newGetCtx(t, "/?unid=not-a-uuid", nil)
		c.SetPathValues(echo.PathValues{{Name: "id", Value: "123"}})
		be.Err(t, app.GetDemozooParam(sl, c, db, "x"), nil)
		be.Equal(t, rec.Code, http.StatusBadRequest)
		be.True(t, strings.Contains(rec.Body.String(), "uuid"))
	})

	t.Run("GetDemozooAndGetPouet", func(t *testing.T) {
		t.Parallel()

		c, _ := newGetCtx(t, "/", nil)
		be.Err(t, app.GetDemozoo(t.Context(), sl, c, nil, 1, testutil.UID, "x"))
		be.Err(t, app.GetPouet(t.Context(), sl, nil, 1, testutil.UID, "x"))
	})
}

func TestErrors(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()

	notFoundFns := []struct {
		name string
		fn   func(*slog.Logger, *echo.Context, string) error
	}{
		{"ArtifactErr", app.ArtifactErr},
		{"ArtifactsErr", app.ArtifactsErr},
		{"ScenerErr", app.ScenerErr},
		{"ReleaserErr", app.ReleaserErr},
	}
	for _, tc := range notFoundFns {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c, _ := newGetCtx(t, "/x", nil)
			be.Err(t, tc.fn(nil, c, "x"))
			be.Err(t, tc.fn(sl, nil, "x"))

			rend := &stubRenderer{}
			c, rec := newGetCtx(t, "/x", rend)
			be.Err(t, tc.fn(sl, c, "x"), nil)
			be.Equal(t, rec.Code, http.StatusNotFound)
			be.Equal(t, rend.name, "status")
		})
	}

	t.Run("PageErr", func(t *testing.T) {
		t.Parallel()

		c, _ := newGetCtx(t, "/x", nil)
		be.Err(t, app.PageErr(nil, c, "x", "1"))
		be.Err(t, app.PageErr(sl, nil, "x", "1"))

		rend := &stubRenderer{}
		c, rec := newGetCtx(t, "/x", rend)
		be.Err(t, app.PageErr(sl, c, "x", "1"), nil)
		be.Equal(t, rec.Code, http.StatusNotFound)
	})

	errFns := []struct {
		name string
		code int
		fn   func(*slog.Logger, *echo.Context, string, error) error
	}{
		{"BadRequestErr", http.StatusBadRequest, app.BadRequestErr},
		{"DatabaseErr", http.StatusServiceUnavailable, app.DatabaseErr},
		{"DownloadErr", http.StatusNotFound, app.DownloadErr},
		{"FileMissingErr", http.StatusServiceUnavailable, app.FileMissingErr},
		{"ForbiddenErr", http.StatusForbidden, app.ForbiddenErr},
		{"InternalErr", http.StatusInternalServerError, app.InternalErr},
	}
	for _, tc := range errFns {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c, _ := newGetCtx(t, "/x", nil)
			be.Err(t, tc.fn(nil, c, "x", app.ErrUser))
			be.Err(t, tc.fn(sl, nil, "x", app.ErrUser))

			rend := &stubRenderer{}
			c, rec := newGetCtx(t, "/x", rend)
			be.Err(t, tc.fn(sl, c, "x", app.ErrUser), nil)
			be.Equal(t, rec.Code, tc.code)
			be.Equal(t, rend.name, "status")
		})
	}
}

func TestTagInfo(t *testing.T) {
	t.Parallel()

	be.Err(t, app.TagInfo(nil))
	be.Err(t, app.PlatformTagInfo(nil))

	// Invalid JSON -> 400
	c, rec := newPostCtx(t, "/", echo.MIMEApplicationJSON, "{bad", nil)
	be.Err(t, app.TagInfo(c), nil)
	be.Equal(t, rec.Code, http.StatusBadRequest)

	c, rec = newPostCtx(t, "/", echo.MIMEApplicationJSON, "{bad", nil)
	be.Err(t, app.PlatformTagInfo(c), nil)
	be.Equal(t, rec.Code, http.StatusBadRequest)

	// Unknown tag -> 400
	c, rec = newPostCtx(t, "/", echo.MIMEApplicationJSON, `{"tag":"nonexistent"}`, nil)
	be.Err(t, app.TagInfo(c), nil)
	be.Equal(t, rec.Code, http.StatusBadRequest)

	c, rec = newPostCtx(t, "/", echo.MIMEApplicationJSON, `{"platform":"bad","tag":"bad"}`, nil)
	be.Err(t, app.PlatformTagInfo(c), nil)
	be.Equal(t, rec.Code, http.StatusBadRequest)

	// Valid tag & platform -> 200
	body := `{"platform":"` + tags.DOS.String() + `","tag":"` + tags.Interview.String() + `"}`
	c, rec = newPostCtx(t, "/", echo.MIMEApplicationJSON, body, nil)
	be.Err(t, app.TagInfo(c), nil)
	be.Equal(t, rec.Code, http.StatusOK)
	be.True(t, rec.Body.Len() > 0)

	c, rec = newPostCtx(t, "/", echo.MIMEApplicationJSON, body, nil)
	be.Err(t, app.PlatformTagInfo(c), nil)
	be.Equal(t, rec.Code, http.StatusOK)
	be.True(t, rec.Body.Len() > 0)
}

func TestOfflineHandlers(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	db := &sql.DB{}

	t.Run("ArtifactsValidation", func(t *testing.T) {
		t.Parallel()

		c, _ := newGetCtx(t, "/files/bad", nil)
		be.Err(t, app.Artifacts(sl, c, nil, "bbs", "1"))
		be.Err(t, app.Deletions(sl, c, nil, "1"))
		be.Err(t, app.Unwanted(sl, c, nil, "1"))
		be.Err(t, app.ForApproval(sl, c, nil, "1"))

		// Invalid URI -> 404
		rend := &stubRenderer{}
		c, rec := newGetCtx(t, "/files/unknown", rend)
		be.Err(t, app.Artifacts(sl, c, db, "not-a-valid-uri", ""), nil)
		be.Equal(t, rec.Code, http.StatusNotFound)

		// Invalid page number -> 404
		rend = &stubRenderer{}
		c, rec = newGetCtx(t, "/files/for-approval/0", rend)
		be.Err(t, app.Artifacts(sl, c, db, "for-approval", "0"), nil)
		be.Equal(t, rec.Code, http.StatusNotFound)

		rend = &stubRenderer{}
		c, rec = newGetCtx(t, "/files/for-approval/abc", rend)
		be.Err(t, app.Artifacts(sl, c, db, "for-approval", "abc"), nil)
		be.Equal(t, rec.Code, http.StatusNotFound)
	})

	t.Run("SearchEmptyRedirects", func(t *testing.T) {
		t.Parallel()

		c, _ := newGetCtx(t, "/search/desc", nil)
		be.Err(t, app.PostDesc(sl, c, nil, "test"))
		be.Err(t, app.PostFilename(sl, c, nil))

		c, rec := newGetCtx(t, "/search/desc", nil)
		be.Err(t, app.PostDesc(sl, c, db, " ,  "), nil)
		be.Equal(t, rec.Code, http.StatusFound)
		be.Equal(t, rec.Header().Get("Location"), "/search/desc")

		c, rec = newGetCtx(t, "/search/file", nil)
		be.Err(t, app.PostFilename(sl, c, db), nil)
		be.Equal(t, rec.Code, http.StatusFound)
		be.Equal(t, rec.Header().Get("Location"), "/search/file")
	})

	t.Run("EditorValidation", func(t *testing.T) {
		t.Parallel()

		c, _ := newPostCtx(t, "/", echo.MIMEApplicationJSON, "{bad", nil)
		be.Err(t, app.EditFn(sl, c, db, nil))
		be.Err(t, app.PlatformEdit(sl, c, nil))
		be.Err(t, app.TagEdit(sl, c, nil))
		be.Err(t, app.MergeDuplicate(sl, c, nil))
		be.Err(t, app.FixNumericSuffix(sl, c, nil))

		// Invalid JSON bind -> 400
		c, rec := newPostCtx(t, "/", echo.MIMEApplicationJSON, "{bad", nil)
		be.Err(t, app.PlatformEdit(sl, c, db), nil)
		be.Equal(t, rec.Code, http.StatusBadRequest)

		// MergeDuplicate missing/equal params -> 400
		c, rec = newPostCtx(t, "/", echo.MIMEApplicationForm, "from=A&to=A", nil)
		be.Err(t, app.MergeDuplicate(sl, c, db), nil)
		be.Equal(t, rec.Code, http.StatusBadRequest)

		// FixNumericSuffix missing / invalid obfuscated ID
		c, _ = newPostCtx(t, "/", "", "", nil)
		be.Err(t, app.FixNumericSuffix(sl, c, db), app.ErrMissingObfuscatedID)

		c, _ = newPostCtx(t, "/", "", "", nil)
		c.SetPathValues(echo.PathValues{{Name: "id", Value: "zzzz"}})
		be.Err(t, app.FixNumericSuffix(sl, c, db), app.ErrInvalidObfuscatedID)
	})

	t.Run("NilDBGuards", func(t *testing.T) {
		t.Parallel()

		c, _ := newGetCtx(t, "/", nil)
		for _, fn := range []func(*slog.Logger, *echo.Context, *sql.DB) error{
			app.Artist, app.Coder, app.Musician, app.Writer, app.Scener,
			app.BBS, app.BBSAZ, app.BBSYear, app.FTP,
			app.Magazine, app.MagazineAZ,
			app.Releasers, app.ReleasersAZ, app.ReleasersYear,
			app.Fixers, app.Duplicates,
		} {
			be.Err(t, fn(sl, c, nil))
			be.Err(t, fn(sl, nil, db))
		}
		be.Err(t, app.Sceners(sl, c, nil, "abc"))
		be.Err(t, app.Releaser(sl, c, nil, "abc", testutil.ProjectFS(t)))
		be.Err(t, app.Configurations(sl, c, nil, config.Config{}))
		be.Err(t, app.Categories(sl, c, nil, false))
		be.Err(t, app.Checksum(sl, c, nil, "abc"))
		be.Err(t, app.Download(sl, c, nil, dir.Directory("x")))
		be.Err(t, app.Inline(sl, c, nil, dir.Directory("x")))
		be.Err(t, app.DownloadJsDos(sl, c, nil, "x", "y"))
	})
}

func TestDBHandlers(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	sl := logs.Discard()
	c, _ := newGetCtx(t, "/", nil)

	be.Err(t, app.Artifacts(sl, c, db, "for-approval", "1"))
	be.Err(t, app.Configurations(sl, c, db, config.Config{}))
	be.Err(t, app.Categories(sl, c, db, false))
	be.Err(t, app.DownloadJsDos(sl, c, db, "abc", "xyz"))
	be.Err(t, app.Download(sl, c, db, "x"))
	be.Err(t, app.Inline(sl, c, db, "x"))
	be.Err(t, app.Sceners(sl, c, db, "/abc"))
	be.Err(t, app.Releaser(sl, c, db, "x", testutil.ProjectFS(t)))
}
