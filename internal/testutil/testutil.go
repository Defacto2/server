//nolint:gochecknoglobals
package testutil

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/labstack/echo/v5"
)

// PostgreSQL database helpers

var share = sync.OnceValue(func() *sql.DB {
	db, err := postgres.Open()
	if err != nil {
		return nil
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil
	}
	return db
})

// DB returns a shared Postgres database connection for use with tests.
// If no connection can be made, the test is skipped.
func DB(tb testing.TB) *sql.DB {
	tb.Helper()

	db := share()
	if db == nil {
		tb.SkipNow()
	}
	return db
}

// DBConns logs the active and maximum Postgres database connections.
func DBConns(tb testing.TB) {
	tb.Helper()

	db := share()
	if db == nil {
		return
	}
	a, m, err := postgres.Connections(db)
	if err != nil {
		tb.Fatalf("connections err: %v", err)
	}
	tb.Log("active database connections:", a, "maximum:", m)
}

// Tx returns a shared Postgres database connection for use with tests.
// The transaction with any database edits get rolled back during the test cleanup.
// If no connection can be made, the test is skipped.
//
// Note: Only one Tx should be used per-test otherwise the test may never resolve.
func Tx(tb testing.TB) *sql.Tx {
	tb.Helper()

	db := DB(tb)

	// leave the context to background, however it might be useful to create
	// a TxWithContext() in the future.
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		tb.Skip("tx failed")
	}

	tb.Cleanup(func() {
		// returning errors for rollback is important due to the sensitivity
		// it has with contexts, timeouts, etc.
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			tb.Errorf("tx rollback: %v", err)
		}
	})

	return tx
}

// Echo package helpers

// EchoContextS returns an echo response using http request using the target url and http status.
//
// The echo.Echo can be provided using [echo.New].
func EchoContextS(tb testing.TB, e *echo.Echo, target string, status int) *echo.Context {
	tb.Helper()

	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.Response().WriteHeader(status)

	return c
}

func EchoContext(tb testing.TB, e *echo.Echo, target string) *echo.Context {
	tb.Helper()

	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()

	return e.NewContext(req, rec)
}

// File system helpers

// OpenFS opens the root directory of this Go project and closes on cleanup.
func OpenFS(tb testing.TB) fs.FS {
	return OpenRoot(tb).FS()
}

// OpenRoot opens the root directory of this Go project and closes on cleanup.
func OpenRoot(tb testing.TB) *os.Root {
	tb.Helper()

	// open root of the repo relative to this file
	root, err := os.OpenRoot(filepath.Join(".."))
	if err != nil {
		tb.Fatal("cannot open the root path", err)
	}

	tb.Cleanup(func() {
		if err := root.Close(); err != nil {
			tb.Log("root close error", err)
		}
	})

	return root
}

// Logger helpers

type Logger struct {
	tb  testing.TB
	buf *bytes.Buffer
	Log *slog.Logger
}

func Buffer(tb testing.TB) *Logger {
	tb.Helper()

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	var w bytes.Buffer
	return &Logger{
		tb:  tb,
		buf: &w,
		Log: slog.New(slog.NewTextHandler(&w, opts)),
	}
}

func (l *Logger) String() string {
	return l.buf.String()
}

func (l *Logger) Reset() {
	l.buf.Reset()
}

func (l *Logger) Contains(substr string) bool {
	l.tb.Helper()

	return strings.Contains(l.buf.String(), substr)
}
