package fixarc_test

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
	"github.com/Defacto2/server/internal/config/fixarc"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

var testLog = logs.Discard()

func requireArc(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath(command.Arc); err != nil {
		t.Skipf("%s not in PATH: %v", command.Arc, err)
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

func TestFiles(t *testing.T) {
	t.Parallel()

	t.Run("nil ctx", func(t *testing.T) {
		t.Parallel()
		_, err := fixarc.Files(nil, new(sql.DB)) //nolint:staticcheck // testing nil context
		be.True(t, errors.Is(err, nils.ErrArgument))
	})

	t.Run("nil exec", func(t *testing.T) {
		t.Parallel()
		_, err := fixarc.Files(t.Context(), nil)
		be.True(t, errors.Is(err, nils.ErrArgument))
	})

	t.Run("db", func(t *testing.T) {
		t.Parallel()
		db := testutil.DB(t)
		fs, err := fixarc.Files(t.Context(), db)
		be.Err(t, err, nil)
		be.True(t, len(fs) > 0)
	})
}

func TestInvalid(t *testing.T) {
	t.Parallel()

	t.Run("nil ctx", func(t *testing.T) {
		t.Parallel()
		mustPanic(t, func() {
			fixarc.Invalid(nil, testLog, "file.arc") //nolint:staticcheck // testing nil context
		})
	})

	t.Run("nil logger", func(t *testing.T) {
		t.Parallel()
		mustPanic(t, func() {
			fixarc.Invalid(t.Context(), nil, "file.arc")
		})
	})

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		requireArc(t)

		tmpDir := t.TempDir()
		src := filepath.Join(tmpDir, "hello.txt")
		be.Err(t, os.WriteFile(src, []byte("hello arc\n"), 0o600), nil)

		arcPath := filepath.Join(tmpDir, "valid.arc")
		out, err := exec.CommandContext(t.Context(), command.Arc, "a", arcPath, src).CombinedOutput() //nolint:gosec // test fixture paths
		if err != nil {
			t.Fatalf("create arc fixture: %v: %s", err, out)
		}

		be.True(t, !fixarc.Invalid(t.Context(), testLog, arcPath))
	})

	t.Run("not archive", func(t *testing.T) {
		t.Parallel()
		requireArc(t)

		txt := filepath.Join(t.TempDir(), "plain.txt")
		be.Err(t, os.WriteFile(txt, []byte("not an archive\n"), 0o600), nil)

		var logBuf bytes.Buffer
		sl := slog.New(slog.NewTextHandler(&logBuf, nil))

		be.True(t, fixarc.Invalid(t.Context(), sl, txt))
		be.True(t, logBuf.Len() > 0)
	})

	t.Run("missing", func(t *testing.T) {
		t.Parallel()
		requireArc(t)

		missing := filepath.Join(t.TempDir(), "missing.arc")
		be.True(t, fixarc.Invalid(t.Context(), testLog, missing))
	})

	t.Run("canceled", func(t *testing.T) {
		t.Parallel()
		requireArc(t)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		missing := filepath.Join(t.TempDir(), "canceled.arc")
		be.True(t, fixarc.Invalid(ctx, testLog, missing))
	})
}
