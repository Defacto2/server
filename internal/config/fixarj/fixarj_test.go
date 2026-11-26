package fixarj_test

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
	"github.com/Defacto2/server/internal/config/fixarj"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

var testLog = logs.Discard()

func requireTool(t *testing.T, name string) {
	t.Helper()
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("%s not in PATH: %v", name, err)
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
		_, err := fixarj.Files(nil, new(sql.DB)) //nolint:staticcheck // testing nil context
		be.True(t, errors.Is(err, nils.ErrArgument))
	})

	t.Run("nil exec", func(t *testing.T) {
		t.Parallel()
		_, err := fixarj.Files(t.Context(), nil)
		be.True(t, errors.Is(err, nils.ErrArgument))
	})

	t.Run("db", func(t *testing.T) {
		t.Parallel()
		db := testutil.DB(t)
		fs, err := fixarj.Files(t.Context(), db)
		be.Err(t, err, nil)
		be.True(t, len(fs) > 0)
	})
}

func TestInvalid(t *testing.T) {
	t.Parallel()

	t.Run("nil ctx", func(t *testing.T) {
		t.Parallel()
		mustPanic(t, func() {
			fixarj.Invalid(nil, testLog, "file.arj") //nolint:staticcheck // testing nil context
		})
	})

	t.Run("nil logger", func(t *testing.T) {
		t.Parallel()
		mustPanic(t, func() {
			fixarj.Invalid(t.Context(), nil, "file.arj")
		})
	})

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		requireTool(t, command.Zip7)
		requireTool(t, "arj")

		tmpDir := t.TempDir()
		src := filepath.Join(tmpDir, "hello.txt")
		be.Err(t, os.WriteFile(src, []byte("hello arj\n"), 0o600), nil)

		arjPath := filepath.Join(tmpDir, "valid.arj")
		out, err := exec.CommandContext(t.Context(), "arj", "a", arjPath, src).CombinedOutput()
		if err != nil {
			t.Fatalf("create arj fixture: %v: %s", err, out)
		}

		be.True(t, !fixarj.Invalid(t.Context(), testLog, arjPath))
	})

	t.Run("not archive", func(t *testing.T) {
		t.Parallel()
		requireTool(t, command.Zip7)

		txt := filepath.Join(t.TempDir(), "plain.txt")
		be.Err(t, os.WriteFile(txt, []byte("not an archive\n"), 0o600), nil)

		var logBuf bytes.Buffer
		sl := slog.New(slog.NewTextHandler(&logBuf, nil))

		be.True(t, fixarj.Invalid(t.Context(), sl, txt))
		be.True(t, logBuf.Len() > 0)
	})

	t.Run("missing", func(t *testing.T) {
		t.Parallel()
		requireTool(t, command.Zip7)

		missing := filepath.Join(t.TempDir(), "missing.arj")
		be.True(t, fixarj.Invalid(t.Context(), testLog, missing))
	})

	t.Run("canceled", func(t *testing.T) {
		t.Parallel()
		requireTool(t, command.Zip7)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		missing := filepath.Join(t.TempDir(), "canceled.arj")
		be.True(t, fixarj.Invalid(ctx, testLog, missing))
	})
}
