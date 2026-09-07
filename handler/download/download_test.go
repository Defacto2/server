package download_test

import (
	"testing"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/download"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

// checked in Sep 26, test coverage was fine at 45%+

func TestChecksum(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)

	c := testutil.NewContext(t, "")
	got := download.Checksum(t.Context(), c, db, "")
	be.Err(t, got)

	c = testutil.NewContext(t, "")
	obfsKey := helper.Obfuscate("1")
	got = download.Checksum(t.Context(), c, db, obfsKey)
	be.Err(t, got)
}

func TestHTTPSend(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	sl := logs.Discard()

	d := download.Download{}
	pathValues := echo.PathValues{
		{Name: "id", Value: "1"},
	}
	c := testutil.NewPath(t, "", pathValues)
	err := d.HTTPSend(sl, c, db)
	be.Err(t, err)
}

func TestExtraHTTPSend(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	ez := download.ExtraZip{}
	pathValues := echo.PathValues{
		{Name: "id", Value: "1"},
	}
	c := testutil.NewPath(t, "", pathValues)
	err := ez.HTTPSend(t.Context(), c, db)
	be.Err(t, err)
}
