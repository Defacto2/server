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

// checked in Sep 26, test coverage was good at over 65+%

func TestLookups(t *testing.T) {
	t.Parallel()

	be.Equal(t, len(command.Lookups), len(command.Infos))
	be.True(t, strings.Contains(command.Infos[0], "HWZip"))
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

func TestFormatOutput(t *testing.T) {
	t.Parallel()

	gifOut := []byte("WebP Encoder version: 1.6.0\nWebP Mux version: 1.6.0\nlibsharpyuv: 0.4.2\n")
	be.Equal(t, command.FormatOutput(command.Gif2webp, gifOut), "WebP Encoder version: 1.6.0")

	hwOut := []byte("\nHWZIP 2.4 -- A simple ZIP program from https://www.hanshq.net/zip.html\n\nUsage:\n")
	be.Equal(t, command.FormatOutput(command.HWZip, hwOut), "HWZIP 2.4 -- A simple ZIP program from https://www.hanshq.net/zip.html")

	magickOut := []byte("Version: ImageMagick 7.1.2-36 Q16-HDRI x86_64 25140 https://imagemagick.org\nCopyright: ...\nDelegates (built-in): bzlib cairo\n")
	be.Equal(t, command.FormatOutput(command.Magick, magickOut), "ImageMagick 7.1.2-36 Q16-HDRI x86_64 25140")

	be.Equal(t, command.FormatOutput(command.Lsar, []byte("v1.10.7\n")), "v1.10.7")
}

func TestDependencyInfos(t *testing.T) {
	t.Parallel()

	arc, img := command.DependencyInfos(t.Context())
	be.Equal(t, len(arc), 11)
	be.Equal(t, len(img), 8)
	be.Equal(t, arc[0].Prog, command.HWZip)
	be.True(t, arc[0].Expected != "")

	for _, info := range append(arc, img...) {
		be.True(t, info.Prog != "")
		be.True(t, info.Expected != "")
		if info.Found {
			be.True(t, info.Abs != "")
		} else {
			be.Equal(t, info.Abs, "")
			be.True(t, info.Err != "")
		}
	}
}
