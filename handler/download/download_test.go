package download_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/download"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/aarondl/null/v8"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

var testLog = logs.Discard()

func TestLastModified(t *testing.T) {
	t.Parallel()

	loc := time.FixedZone("UTC+2", 2*60*60)
	tests := []struct {
		name string
		art  *models.File
		want string
	}{
		{"nil", nil, ""},
		{"invalid", &models.File{}, ""},
		{"epoch year", &models.File{FileLastModified: null.TimeFrom(time.Date(1980, 12, 31, 0, 0, 0, 0, time.UTC))}, ""},
		{"pre-epoch", &models.File{FileLastModified: null.TimeFrom(time.Date(1979, 6, 1, 0, 0, 0, 0, time.UTC))}, ""},
		{
			"valid utc",
			&models.File{FileLastModified: null.TimeFrom(time.Date(2015, 10, 21, 7, 28, 0, 0, time.UTC))},
			"Wed, 21 Oct 2015 07:28:00 GMT",
		},
		{
			"converts timezone to utc",
			&models.File{FileLastModified: null.TimeFrom(time.Date(2015, 10, 21, 9, 28, 0, 0, loc))},
			"Wed, 21 Oct 2015 07:28:00 GMT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := download.LastModified(tt.art)
			be.Equal(t, got, tt.want)
			if got != "" {
				_, err := http.ParseTime(got)
				be.Err(t, err, nil)
			}
		})
	}
}

func TestChecksum(t *testing.T) {
	t.Parallel()

	t.Run("nil args", func(t *testing.T) {
		t.Parallel()

		c := testutil.NewContext(t, "")
		be.Err(t, download.Checksum(t.Context(), c, nil, "abc"))
		be.Err(t, download.Checksum(t.Context(), nil, new(sql.DB), "abc"))
	})

	t.Run("db error", func(t *testing.T) {
		t.Parallel()

		db, err := sql.Open("pgx", "")
		be.Err(t, err, nil)
		defer db.Close()

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		c := testutil.NewContext(t, "")
		be.Err(t, download.Checksum(ctx, c, db, helper.Obfuscate("1")))
	})

	t.Run("db", func(t *testing.T) {
		t.Parallel()

		db := testutil.DB(t)

		c := testutil.NewContext(t, "")
		got := download.Checksum(t.Context(), c, db, "")
		be.Err(t, got)

		c = testutil.NewContext(t, "")
		obfsKey := helper.Obfuscate("1")
		got = download.Checksum(t.Context(), c, db, obfsKey)
		be.Err(t, got)
	})
}

func TestHTTPSend(t *testing.T) {
	t.Parallel()

	t.Run("nil args", func(t *testing.T) {
		t.Parallel()

		d := download.Download{}
		c := testutil.NewContext(t, "")
		be.Err(t, d.HTTPSend(nil, c, new(sql.DB)))
		be.Err(t, d.HTTPSend(testLog, nil, new(sql.DB)))
		be.Err(t, d.HTTPSend(testLog, c, nil))
	})

	t.Run("db error", func(t *testing.T) {
		t.Parallel()

		db, err := sql.Open("pgx", "")
		be.Err(t, err, nil)
		defer db.Close()

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		e := echo.New()
		req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
		c := e.NewContext(req, httptest.NewRecorder())
		c.SetPathValues(echo.PathValues{{Name: "id", Value: helper.Obfuscate("1")}})

		d := download.Download{}
		be.Err(t, d.HTTPSend(testLog, c, db))
	})

	t.Run("db", func(t *testing.T) {
		t.Parallel()

		db := testutil.DB(t)

		d := download.Download{}
		pathValues := echo.PathValues{
			{Name: "id", Value: "1"},
		}
		c := testutil.NewPath(t, "", pathValues)
		err := d.HTTPSend(testLog, c, db)
		be.Err(t, err)
	})
}

func TestExtraHTTPSend(t *testing.T) {
	t.Parallel()

	t.Run("nil args", func(t *testing.T) {
		t.Parallel()

		ez := download.ExtraZip{}
		c := testutil.NewContext(t, "")
		be.Err(t, ez.HTTPSend(t.Context(), nil, new(sql.DB)))
		be.Err(t, ez.HTTPSend(t.Context(), c, nil))
	})

	t.Run("db error", func(t *testing.T) {
		t.Parallel()

		db, err := sql.Open("pgx", "")
		be.Err(t, err, nil)
		defer db.Close()

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		c := testutil.NewPath(t, "", echo.PathValues{{Name: "id", Value: helper.Obfuscate("1")}})
		ez := download.ExtraZip{}
		be.Err(t, ez.HTTPSend(ctx, c, db))
	})

	t.Run("db", func(t *testing.T) {
		t.Parallel()

		db := testutil.DB(t)
		ez := download.ExtraZip{}
		pathValues := echo.PathValues{
			{Name: "id", Value: "1"},
		}
		c := testutil.NewPath(t, "", pathValues)
		err := ez.HTTPSend(t.Context(), c, db)
		be.Err(t, err)

		pathValues = echo.PathValues{
			{Name: "id", Value: helper.Obfuscate("1")},
		}
		c = testutil.NewPath(t, "", pathValues)
		err = ez.HTTPSend(t.Context(), c, db)
		be.Err(t, err, download.ErrStat)
	})
}
