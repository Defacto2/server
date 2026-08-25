package command_test

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/nalgeon/be"
)

func TestLookups(t *testing.T) {
	t.Parallel()

	be.Equal(t, len(command.Lookups), len(command.Infos))
	be.True(t, strings.Contains(command.Infos[0], command.Arc))
}

func TestCopyFile(t *testing.T) {
	t.Parallel()
	sl := slog.Default()

	got := command.CopyFile(nil, "", "")
	be.Err(t, got)

	tmp, got := dir.CreateTemp("command_test")
	be.Err(t, got, nil)

	got = command.CopyFile(sl, "", "")
	be.Err(t, got)

	got = command.CopyFile(sl, tmp.Name(), "")
	be.Err(t, got)

	dst := tmp.Name() + ".txt"
	got = command.CopyFile(sl, tmp.Name(), dst)
	be.Err(t, got, nil)

	t.Cleanup(func() {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
	})
}

func TestLookup(t *testing.T) {
	t.Parallel()

	s, got := command.Lookup("")
	be.Err(t, got, nil)
	be.Equal(t, s, "")

	_, got = command.Lookup("thiscommanddoesnotexist")
	be.Err(t, got)

	s, got = command.Lookup("go")
	be.Err(t, got, nil)
	be.True(t, strings.Contains(s, "go"))
}

func TestLookupS(t *testing.T) {
	t.Parallel()

	got := command.LookupS(t.Context(), "", "", "")
	be.Err(t, got)

	got = command.LookupS(t.Context(), "thiscommanddoesnotexist", "", "")
	be.Err(t, got)

	got = command.LookupS(t.Context(), "go", "", "")
	be.Err(t, got)

	// version arg output example:
	// go version go1.16.5 linux/amd64
	got = command.LookupS(t.Context(), "go", "version", "go version go1.")
	be.Err(t, got, nil)
}

func TestRun(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	r := command.Runner{}
	_, got := r.Run(ctx, "", "")
	be.Err(t, got)

	const none = ""
	r.Log = slog.Default()
	_, got = r.Run(ctx, "", none)
	be.Err(t, got)

	_, got = r.Run(ctx, "thiscommanddoesnotexist", none)
	be.Err(t, got)

	_, got = r.Run(ctx, "go", none)
	// go without args will return an unknown command error
	be.Err(t, got)

	b, got := r.Run(ctx, "go", "version")
	be.Err(t, got, nil)
	be.True(t, bytes.HasPrefix(b, []byte("go version go1.")))

	// test timeout
}
