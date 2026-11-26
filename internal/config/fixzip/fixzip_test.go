package fixzip_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/config/fixzip"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

var testLog = logs.Discard()

func requireHWZip(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath(command.HWZip); err != nil {
		t.Skipf("%s not in PATH: %v", command.HWZip, err)
	}
}

func mustPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	fn()
}

func TestLegacy(t *testing.T) {
	t.Parallel()

	t.Run("nil logger", func(t *testing.T) {
		t.Parallel()
		mustPanic(t, func() {
			fixzip.Legacy(nil, "file.zip")
		})
	})

	t.Run("legacy", func(t *testing.T) {
		t.Parallel()
		src := testutil.FileData("implodezip").Abs()
		be.True(t, fixzip.Legacy(testLog, src))
	})

	t.Run("deflate", func(t *testing.T) {
		t.Parallel()
		src := testutil.FileData("archivezip").Abs()
		be.True(t, !fixzip.Legacy(testLog, src))
	})

	t.Run("not zip", func(t *testing.T) {
		t.Parallel()

		txt := filepath.Join(t.TempDir(), "plain.txt")
		be.Err(t, os.WriteFile(txt, []byte("not a zip\n"), 0o600), nil)

		var logBuf bytes.Buffer
		sl := slog.New(slog.NewTextHandler(&logBuf, nil))

		be.True(t, !fixzip.Legacy(sl, txt))
		be.True(t, logBuf.Len() > 0)
	})

	t.Run("missing", func(t *testing.T) {
		t.Parallel()

		var logBuf bytes.Buffer
		sl := slog.New(slog.NewTextHandler(&logBuf, nil))

		missing := filepath.Join(t.TempDir(), "missing.zip")
		be.True(t, !fixzip.Legacy(sl, missing))
		be.True(t, logBuf.Len() > 0)
	})
}

func TestFiles(t *testing.T) {
	t.Parallel()

	t.Run("nil ctx", func(t *testing.T) {
		t.Parallel()
		_, err := fixzip.Files(nil, new(sql.DB)) //nolint:staticcheck // testing nil context
		be.True(t, errors.Is(err, nils.ErrArgument))
	})

	t.Run("nil exec", func(t *testing.T) {
		t.Parallel()
		_, err := fixzip.Files(t.Context(), nil)
		be.True(t, errors.Is(err, nils.ErrArgument))
	})

	t.Run("db", func(t *testing.T) {
		t.Parallel()
		db := testutil.DB(t)
		fs, err := fixzip.Files(t.Context(), db)
		be.Err(t, err, nil)
		be.True(t, len(fs) > 0)
	})
}

func TestInvalid(t *testing.T) {
	t.Parallel()

	t.Run("nil ctx", func(t *testing.T) {
		t.Parallel()
		mustPanic(t, func() {
			fixzip.Invalid(nil, testLog, "file.zip") //nolint:staticcheck // testing nil context
		})
	})

	t.Run("nil logger", func(t *testing.T) {
		t.Parallel()
		mustPanic(t, func() {
			fixzip.Invalid(t.Context(), nil, "file.zip")
		})
	})

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		requireHWZip(t)

		src := testutil.FileData("archivezip").Abs()
		be.True(t, !fixzip.Invalid(t.Context(), testLog, src))
	})

	t.Run("legacy", func(t *testing.T) {
		t.Parallel()
		requireHWZip(t)

		src := testutil.FileData("implodezip").Abs()
		be.True(t, !fixzip.Invalid(t.Context(), testLog, src))
	})

	t.Run("not archive", func(t *testing.T) {
		t.Parallel()
		requireHWZip(t)

		txt := filepath.Join(t.TempDir(), "plain.txt")
		be.Err(t, os.WriteFile(txt, []byte("not a zip\n"), 0o600), nil)

		be.True(t, fixzip.Invalid(t.Context(), testLog, txt))
	})

	t.Run("missing", func(t *testing.T) {
		t.Parallel()
		requireHWZip(t)

		var logBuf bytes.Buffer
		sl := slog.New(slog.NewTextHandler(&logBuf, nil))

		missing := filepath.Join(t.TempDir(), "missing.zip")
		be.True(t, fixzip.Invalid(t.Context(), sl, missing))
		be.True(t, logBuf.Len() > 0)
	})

	t.Run("canceled", func(t *testing.T) {
		t.Parallel()
		requireHWZip(t)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		missing := filepath.Join(t.TempDir(), "canceled.zip")
		be.True(t, fixzip.Invalid(ctx, testLog, missing))
	})
}
