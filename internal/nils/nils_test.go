package nils_test

import (
	"bytes"
	"context"
	"database/sql"
	"log/slog"
	"mime/multipart"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/fulltext"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

// checked in Sep 26, test coverage was great at over 70%+

func TestChecks(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	sl1 := func(sl *slog.Logger) error {
		return nils.Check(sl)
	}
	be.Err(t, sl1(nil), nils.ErrSlogLogger)
	be.Err(t, sl1(sl), nil)

	multi1 := func(sl *slog.Logger, db *sql.DB) error {
		return nils.Check(sl, db)
	}
	be.Err(t, multi1(nil, nil), nils.ErrSlogLogger)
	be.Err(t, multi1(sl, nil), nils.ErrBoil)

	ctx := t.Context()
	multi2 := func(ctx context.Context, sl *slog.Logger) error {
		return nils.Check(ctx, sl)
	}
	be.Err(t, multi2(nil, nil), nils.ErrArgument)
	be.Err(t, multi2(ctx, nil), nils.ErrSlogLogger)
	be.Err(t, multi2(nil, sl), nils.ErrArgument)
	be.Err(t, multi2(ctx, sl), nil)

	db1 := func(db *sql.DB) error {
		return nils.Check(db)
	}
	be.Err(t, db1(nil), nils.ErrBoil)
	be.Err(t, db1(new(sql.DB)), nil)

	tests := []struct {
		name string
		arg  any
		want error
	}{
		{"echo handler", echo.HandlerFunc(nil), nils.ErrEchoHandler},
		{"echo routes", echo.Routes(nil), nils.ErrEchoRoutes},
		{"echo context", (*echo.Context)(nil), nils.ErrEchoContext},
		{"echo", (*echo.Echo)(nil), nils.ErrEchoEcho},
		{"echo group", (*echo.Group)(nil), nils.ErrEchoGroup},
		{"models file", (*models.File)(nil), nils.ErrModels},
		{"bytes buffer", (*bytes.Buffer)(nil), nils.ErrBytesBuf},
		{"file header", (*multipart.FileHeader)(nil), nils.ErrFileHeader},
		{"tidbits", (*fulltext.Tidbits)(nil), nils.ErrPointer},
		{"fallback", (*model.Summary)(nil), nils.ErrArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			be.Err(t, nils.Check(tt.arg), tt.want)
		})
	}
}

func TestSlog(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	var buf bytes.Buffer
	sl := slog.New(slog.NewTextHandler(&buf, nil))
	const msg = "test of three args, with fh being nil"

	multi := func(
		ctx context.Context, sl *slog.Logger, fh *multipart.FileHeader,
	) (bool, error) {
		bad := nils.Slog(msg, ctx, sl, fh)
		err := nils.Check(ctx, sl, fh)
		return bad, err
	}
	bad, got := multi(ctx, sl, nil)
	be.True(t, bad)
	be.Err(t, got, nils.ErrFileHeader)
	be.True(t, strings.Contains(buf.String(), msg))
	be.True(t, strings.Contains(buf.String(), nils.ErrFileHeader.Error()))

	buf.Reset()
	fh := &multipart.FileHeader{}
	bad, got = multi(ctx, sl, fh)
	be.True(t, !bad)
	be.Err(t, got, nil)
	be.Equal(t, buf.String(), "")
}

func TestIsNil(t *testing.T) {
	t.Parallel()

	got := nils.IsNil(nil)
	be.True(t, got)

	a := ""
	got = nils.IsNil(a)
	be.True(t, !got)

	one := 1
	got = nils.IsNil(one)
	be.True(t, !got)

	fh := &multipart.FileHeader{}
	got = nils.IsNil(fh)
	be.True(t, !got)

	var v multipart.FileHeader
	got = nils.IsNil(v)
	be.True(t, !got)

	chk0 := func(x *slog.Logger) bool {
		return nils.IsNil(x)
	}
	got = chk0(nil)
	be.True(t, got)

	chk1 := func(x any) bool { //nolint:gocritic
		return nils.IsNil(x)
	}
	got = chk1(nil)
	be.True(t, got)
	got = chk1("nil")
	be.True(t, !got)

	fh = nil
	got = chk1(fh)
	be.True(t, got)

	var (
		nilSlice []int
		nilMap   map[string]int
		nilFunc  func()
		nilChan  chan int
	)
	be.True(t, nils.IsNil(nilSlice))
	be.True(t, !nils.IsNil([]int{}))
	be.True(t, nils.IsNil(nilMap))
	be.True(t, !nils.IsNil(map[string]int{}))
	be.True(t, nils.IsNil(nilFunc))
	be.True(t, !nils.IsNil(func() {}))
	be.True(t, nils.IsNil(nilChan))
	be.True(t, !nils.IsNil(make(chan int)))
}

func TestBoilExec(t *testing.T) {
	t.Parallel()

	bad := nils.BoilExec(nil)
	be.True(t, bad)

	var exec boil.ContextExecutor
	bad = nils.BoilExec(exec)
	be.True(t, bad)

	var db *sql.DB
	bad = nils.BoilExec(db)
	be.True(t, bad)

	bad = nils.BoilExec(new(sql.DB))
	be.True(t, !bad)
}

func TestBoilExecCrash(t *testing.T) {
	t.Parallel()

	nils.BoilExecCrash(new(sql.DB))

	for _, tc := range []struct {
		name string
		exec boil.ContextExecutor
	}{
		{"nil", nil},
		{"typed nil", (*sql.DB)(nil)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				r := recover()
				err, ok := r.(error)
				be.True(t, ok)
				be.Err(t, err, nils.ErrBoil)
			}()
			nils.BoilExecCrash(tc.exec)
			t.Error("expected panic")
		})
	}
}
