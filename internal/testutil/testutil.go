//nolint:exhaustruct_v5,gochecknoglobals,mnd,wrapcheck
package testutil

import (
	"bytes"
	"embed"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/aarondl/null/v8"
)

// INFO: Check untested funcs, run:
// go test -coverprofile=coverage.out . && go tool cover -html=coverage.out
//
// INFO: To run benchmarks, run:
// go test -bench=Benchmark -benchmem
//
// INFO:Check test durability, run:
// go test . -count=1000 -race -cover

//go:embed testdata/*
var testdataFS embed.FS

// Helper values

const (
	YT        = "3V8rpJDpbKg"                          // YTID is an example YouTube video ID
	UID       = "123e4567-e89b-12d3-a456-426614174000" // UUID is a generic Universal Unique ID
	UID4      = "bb2310e1-93aa-475e-8b88-59eb1fb984a4" // UID4 is a UUID version 4
	SCREENPNG = 328_468                                // SCREENPNG is the byte file size of testdata/SCREEN.PNG
	// RTF is an example of Rich-Text-Format encoded text.
	RTF = `{\rtf1\ansi{\fonttbl\f0\fswiss Helvetica;}\f0\pard
 This is some {\b bold} text.\par
}`
)

// Model helpers

// NewModel returns a new [models.File] entry with the following boilerplate:
//
//   - ID:				1
//   - UUID:			bb2310e1-93aa-475e-8b88-59eb1fb984a4
//   - Filename:	filename.txt
//   - Filesize:	9876
func NewModel(tb testing.TB) *models.File {
	tb.Helper()

	art := models.File{}
	art.ID = 1
	art.UUID = null.StringFrom(UID4)
	art.Filename = null.StringFrom("filename.txt")
	art.Filesize = null.Int64From(9876)
	return &art
}

// File system helpers

// OpenFS opens the root directory of the testutil package and closes on cleanup.
func OpenFS(tb testing.TB) fs.FS {
	tb.Helper()

	return OpenRoot(tb).FS()
}

// OpenRoot opens the root directory of the testutil package and closes on cleanup.
func OpenRoot(tb testing.TB) *os.Root {
	tb.Helper()

	return openRoot(tb, "..")
}

// ProjectFS opens the root directory of this Go project and closes on cleanup.
func ProjectFS(tb testing.TB) fs.FS {
	tb.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		tb.Fatal("runtime caller failed")
	}
	name := filepath.Join(filepath.Dir(filename), "..", "..")

	return openRoot(tb, name).FS()
}

func openRoot(tb testing.TB, name string) *os.Root {
	tb.Helper()
	if name == "" {
		name = ".."
	}

	// open root of the repo relative to this file
	root, err := os.OpenRoot(name)
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

// ANSITests generates a set of testing data containing randomized data
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

// Buffer returns a custom debug logger that writes to [Logger].
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

// String returns the content of the logger.
func (l *Logger) String() string {
	return l.buf.String()
}

// Reset empties the content of the logger.
func (l *Logger) Reset() {
	l.buf.Reset()
}

// Contains returns true if the substr is found in the content of the logger.
func (l *Logger) Contains(substr string) bool {
	l.tb.Helper()

	return strings.Contains(l.buf.String(), substr)
}
