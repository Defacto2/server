package nils_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"log/slog"
	"mime/multipart"
	"testing"

	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/nalgeon/be"
)

func TestChecks(t *testing.T) {
	t.Parallel()
	// don't test for a valid DB value as it is requires too much boiler plate
	db1 := func(db *sql.DB) error {
		return nils.Check(db)
	}
	be.Err(t, db1(nil))

	sl := slog.Default()
	sl1 := func(sl *slog.Logger) error {
		return nils.Check(sl)
	}
	be.Err(t, sl1(nil))
	be.Err(t, sl1(sl), nil)

	multi1 := func(sl *slog.Logger, db *sql.DB) error {
		return nils.Check(sl, db)
	}
	be.Err(t, multi1(nil, nil))
	be.Err(t, multi1(sl, nil))

	ctx := t.Context()
	multi2 := func(ctx context.Context, sl *slog.Logger) error {
		return nils.Check(ctx, sl)
	}
	be.Err(t, multi2(nil, nil))
	be.Err(t, multi2(ctx, nil))
	be.Err(t, multi2(nil, sl))
	be.Err(t, multi2(ctx, sl), nil)

	gen1 := func(m *model.Summary) error {
		return nils.Check(m)
	}
	be.Err(t, gen1(nil))
	be.True(t, errors.Is(gen1(nil), nils.ErrArgument))
}

func TestSlog(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	sl := slog.Default()
	multi := func(
		ctx context.Context, sl *slog.Logger, fh *multipart.FileHeader,
	) (bool, error) {
		bad := nils.Slog("test of three args, with fh being nil", ctx, sl, fh)
		err := nils.Check(ctx, sl, fh)
		return bad, err
	}
	bad, got := multi(ctx, sl, nil)
	be.True(t, bad)
	be.Err(t, got)

	fh := &multipart.FileHeader{}
	bad, got = multi(ctx, sl, fh)
	be.True(t, !bad)
	be.Err(t, got, nil)
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
}

type fake struct {
	t *testing.T
}

func (f fake) Open(_ string) (driver.Conn, error) {
	f.t.Helper()
	return nil, driver.ErrSkip
}

func TestBoilExec(t *testing.T) {
	t.Parallel()

	sql.Register("boil_noop", fake{t: t})

	bad := nils.BoilExec(nil)
	be.True(t, bad)

	var exec boil.ContextExecutor
	bad = nils.BoilExec(exec)
	be.True(t, bad)

	var db *sql.DB
	bad = nils.BoilExec(db)
	be.True(t, bad)

	exec, _ = sql.Open("boil_noop", "")
	bad = nils.BoilExec(exec)
	be.True(t, !bad)
}
