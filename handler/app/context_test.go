package app_test

// Most of these tests are for nil values to ensure there are no panics.

import (
	"context"
	"database/sql"
	"log/slog"
	"strconv"
	"testing"
	"time"

	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

func TestEmpty(t *testing.T) {
	t.Parallel()

	x := app.EmptyTester(testutil.NewContext(t, ""))
	be.Equal(t, x["title"], "")
	be.Equal(t, x["uploader"], true)
	be.Equal(t, x["xxxxxxxx"], nil)
}

type Fn1 func(c *echo.Context, s string) error

func TestFn1(t *testing.T) {
	t.Parallel()

	fns := []Fn1{
		app.PouetCache,
		app.ProdPouet,
		app.ProdZoo,
	}

	for n, fn := range fns {
		t.Run("fn1 #"+strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()

			c := testutil.NewContext(t, "")
			got := fn(c, "x")
			switch n {
			case 0, 3:
				be.Err(t, got)
			default:
				be.Err(t, got, nil)
			}
		})
	}
}

type Fn2 func(sl *slog.Logger, c *echo.Context, s string) error

func TestFn2(t *testing.T) {
	t.Parallel()

	fns := []Fn2{
		app.Website,
		app.VotePouet,
		app.ArtifactErr,
		app.ArtifactsErr,
		app.ScenerErr,
		app.ReleaserErr,
	}

	for n, fn := range fns {
		t.Run("fn2 #"+strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()

			sl := logs.Discard()
			c := testutil.NewContext(t, "")
			got := fn(sl, c, "x")
			switch n {
			case 1:
				be.Err(t, got, nil)
			default:
				be.Err(t, got)
			}
		})
	}
}

type Fn3 func(sl *slog.Logger, c *echo.Context) error

func TestFn3(t *testing.T) {
	t.Parallel()

	fns := []Fn3{
		app.APIInfo,
		app.Apps,
		app.Areacodes,
		app.BrokenTexts,
		app.Compression,
		app.Fixes,
		app.History,
		app.Index,
		app.Interview,
		app.Terms,
		app.New,
		app.SearchDesc,
		app.SearchID,
		app.SearchFile,
		app.SearchReleaser,
		app.SignedOut,
		app.SignOut,
		app.Titles,
		app.Thanks,
		app.TheScene,
	}

	for n, fn := range fns {
		t.Run("fn3 #"+strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()

			sl := logs.Discard()
			c := testutil.NewContext(t, "")
			got := fn(sl, c)
			be.Err(t, got)
		})
	}
}

type Fn4 func(sl *slog.Logger, c *echo.Context, db *sql.DB) error

func TestFn4(t *testing.T) {
	t.Parallel()

	fns := []Fn4{
		app.Artist,
		app.BBS,
		app.BBSAZ,
		app.BBSYear,
		app.Coder,
		app.FTP,
		app.Magazine,
		app.MagazineAZ,
		app.Musician,
		app.PlatformEdit,
		app.TagEdit,
		app.PostFilename,
		app.Releasers,
		app.ReleasersAZ,
		app.ReleasersYear,
		app.Scener,
		app.Fixers,
		app.FixNumericSuffix,
		app.Writer,
	}

	for n, fn := range fns {
		t.Run("fn4 #"+strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()

			sl := logs.Discard()
			c := testutil.NewContext(t, "")
			db := testutil.DB(t)
			got := fn(sl, c, db)

			switch n {
			case 11:
				be.Err(t, got, nil)
			default:
				be.Err(t, got)
			}
		})
	}
}

type Fn5 func(sl *slog.Logger, c *echo.Context, uri string, err error) error

func TestFn5(t *testing.T) {
	t.Parallel()

	fns := []Fn5{
		app.BadRequestErr,
		app.DatabaseErr,
		app.DownloadErr,
		app.FileMissingErr,
		app.ForbiddenErr,
		app.InternalErr,
	}

	for n, fn := range fns {
		t.Run("fn5 #"+strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()

			sl := logs.Discard()
			c := testutil.NewContext(t, "")
			got := fn(sl, c, "x", app.ErrUser)
			switch n {
			// case 1:
			// 	be.Err(t, got, nil)
			default:
				be.Err(t, got)
			}
		})
	}
}

type Fn6 func(sl *slog.Logger, c *echo.Context, db *sql.DB, uri string) error

func TestExec(t *testing.T) {
	t.Parallel()

	fns := []Fn6{
		app.Deletions,
		app.ForApproval,
		app.Unwanted,
		app.PostDesc,
		app.Checksum,
	}

	for n, fn := range fns {
		t.Run("fn6 #"+strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()

			sl := logs.Discard()
			c := testutil.NewContext(t, "")

			got := fn(sl, c, nil, "x")
			switch n {
			// case 1:
			// 	be.Err(t, got, nil)
			default:
				be.Err(t, got)
			}
		})
	}
}

// exec boil.ContextExecutor
type Fn7 func(sl *slog.Logger, c *echo.Context, db *sql.DB, path dir.Directory) error

func TestDirs(t *testing.T) {
	t.Parallel()

	fns := []Fn7{
		app.Inline,
		app.Download,
	}

	for n, fn := range fns {
		t.Run("fn7 #"+strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()

			db := testutil.DB(t)
			sl := logs.Discard()
			c := testutil.NewContext(t, "")
			got := fn(sl, c, db, "x")
			switch n {
			// case 1:
			// 	be.Err(t, got, nil)
			default:
				be.Err(t, got)
			}
		})
	}
}

type Fn8 func(c *echo.Context) error

func TestCtx(t *testing.T) {
	t.Parallel()

	fns := []Fn8{
		app.TagInfo,
		app.PlatformTagInfo,
	}

	for n, fn := range fns {
		t.Run("fn8 #"+strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()

			c := testutil.NewContext(t, "")
			got := fn(c)
			switch n {
			case 0, 1:
				be.Err(t, got, nil)
			default:
				be.Err(t, got)
			}
		})
	}
}

func TestArtifacts(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	c := testutil.NewContext(t, "")
	db := testutil.DB(t)

	x := app.Artifacts(sl, c, db, "", "")
	be.Err(t, x)

	x = app.Artifacts(sl, c, db, "for-approval", "1")
	be.Err(t, x)
}

func TestConfigurations(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	c := testutil.NewContext(t, "")
	db := testutil.DB(t)

	x := app.Configurations(sl, c, db, config.Config{})
	be.Err(t, x)
}

func TestDownloadJsDos(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	c := testutil.NewContext(t, "")
	db := testutil.DB(t)

	x := app.DownloadJsDos(sl, c, db, "abc", "xyz")
	be.Err(t, x)
}

func TestCategories(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	c := testutil.NewContext(t, "")
	db := testutil.DB(t)

	x := app.Categories(sl, c, db, false)
	be.Err(t, x)
}

func TestGetDemozooParam(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	tx := testutil.Tx(t)
	c := testutil.NewContext(t, "")

	x := app.GetDemozooParam(sl, c, tx, "abc")
	be.Err(t, x, nil)
}

func TestGetDemozoo(t *testing.T) {
	t.Parallel()

	tx := testutil.Tx(t)
	sl := logs.Discard()
	c := testutil.NewContext(t, "")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	x := app.GetDemozoo(ctx, sl, c, tx, 1, testutil.UID, "abc")
	be.Err(t, x)

	x = app.GetPouet(ctx, sl, c, tx, 1, testutil.UID, "abc")
	be.Err(t, x)
}

func TestGoogleCallback(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	c := testutil.NewContext(t, "")

	x := app.GoogleCallback(sl, c, "abc", 100, [48]byte{})
	be.Err(t, x)
}

func TestPage404(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	c := testutil.NewContext(t, "")

	x := app.PageErr(sl, c, "x", "1")
	be.Err(t, x)
}

func TestSignin(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	c := testutil.NewContext(t, "")

	x := app.Signin(sl, c, "", nil)
	be.Err(t, x)
}

func TestSceners(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	sl := logs.Discard()
	c := testutil.NewContext(t, "")

	x := app.Sceners(sl, c, db, "/abc")
	be.Err(t, x)
}

func TestReleasers(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	sl := logs.Discard()
	c := testutil.NewContext(t, "")

	x := app.Releaser(sl, c, db, "x", testutil.OpenFS(t))
	be.Err(t, x)
}
