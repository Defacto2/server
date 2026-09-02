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
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/labstack/echo/v5"
)

// INFO: Check untested funcs, run:
// go test -coverprofile=cover.out . && go tool cover -html=cover.out
//
// INFO: To run benchmarks, run:
// go test -bench=Benchmark -benchmem
//
// INFO:Check test durability, run:
// go test . -count=1000 -race -cover

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

func NewContext(tb testing.TB, target string) *echo.Context {
	tb.Helper()

	e := echo.New()
	return EchoContext(tb, e, target)
}

func NewForm(tb testing.TB, target, key, value string) *echo.Context {
	tb.Helper()

	form := url.Values{}
	form.Set(key, value)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	rec := httptest.NewRecorder()

	e := echo.New()
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

// Binary text helpers

// DOSBIN converts the string into a single line, MS-DOS BIN [binary text].
// If s is left empty, "hello world" will be used.
// Long strings will be cropped to fit the 80 column per line limit.
//
// [binary text]: http://fileformats.archiveteam.org/wiki/BIN_(Binary_Text)
func DOSBIN(tb testing.TB, s string) [161]byte {
	tb.Helper()

	const (
		text = "hello world"
		cols = 80
		attr = byte(0x0F) // bright white on black background
		eof  = byte(0x1A) // MS-DOS end-of-file marker
	)

	if s == "" {
		s = text
	}
	if len(s) > 80 {
		s = s[:80]
	}

	var buf [161]byte
	for i := range cols {
		char := byte(' ')
		if i < len(s) {
			char = s[i]
		}
		buf[i*2] = char
		buf[i*2+1] = attr
	}
	buf[len(buf)-1] = eof

	return buf
}

// MkDOSBIN converts the string into a single line, MS-DOS BIN file
// saved to the named file and placed in a temporary test directory.
// The file location is returned.
//
// See [DOSBIN].
func MkDOSBIN(tb testing.TB, name, s string) (string, error) {
	tb.Helper()

	tmp := tb.TempDir()
	dst := filepath.Join(tmp, name)

	f, err := os.Create(dst)
	if err != nil {
		return "", err
	}

	b := DOSBIN(tb, s)
	_, err = f.Write(b[:])
	return f.Name(), errors.Join(err, f.Close())
}

// Text files and named file helpers

type ANSITest struct {
	Name string
	Data []byte
}

// ANSITexts generates a set of testing data containing randomized data
// embedded with ANSI escape codes. Being intended for use in tests
// requiring detection and removal.
func ANSITests(tb testing.TB) []ANSITest {
	tb.Helper()

	makeData := func(size int, ansi string, pos int) []byte {
		buf := bytes.Repeat([]byte("Lorem ipsum dolor sit amet, consectetur. "), (size/40)+1)[:size]
		if pos >= 0 && pos < size {
			copy(buf[pos:], ansi)
		}
		return buf
	}
	const kb, mb = 1024, 1024 * 1024
	tests := []ANSITest{
		{"ImmediateMatch", []byte("\x1b[31mRed Text")},
		{"Match1KB      ", makeData(1*kb, "\x1b[1A", 500)},
		{"Match500KB    ", makeData(500*kb, "\x1b[10;20H", 450*kb)},
		{"NoMatch32KB   ", makeData(32*kb, "", -1)},
		{"NoMatch1MB    ", makeData(1*mb, "", -1)},
		{"BoundarySplit ", func() []byte {
			d := makeData(64*kb, "", -1)
			d[32*kb-1], d[32*kb] = '\x1b', '['
			return d
		}()},
	}
	return tests
}

var Content1 = [...]string{
	"file1.nfo",
	"file1.txt",
	"file1.unp",
	"file1.doc",
}

var Content2 = [...]string{
	"file.diz",
	"file.asc",
	"file.1st",
	"group2.dox",
}

var Content3 = [...]string{
	"file3.nfo",
	"file.txt",
	"file30.unp",
	"file3x.doc",
	"filex3.diz",
	"file3.asc",
	"file3.1st",
	"file3.dox",
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
