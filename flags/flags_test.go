//nolint:paralleltest
package flags_test

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/Defacto2/server/flags"
	"github.com/Defacto2/server/internal/config"
	"github.com/nalgeon/be"
)

// Note: urfave/cli/v2 mutates package-level HelpFlag and VersionFlag during
// app.Run, so tests that invoke Run cannot use t.Parallel().

var errRoutesFailed = errors.New("routes failed")

func TestRun(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	ec, err := flags.Run(nil, "", nil)
	be.Err(t, err, flags.ErrNoConfig)
	be.Equal(t, ec, flags.UsageErr)

	var buf bytes.Buffer
	c := config.Config{}

	os.Args = []string{flags.Program}
	ec, err = flags.Run(&buf, "", &c)
	be.Err(t, err, nil)
	be.Equal(t, ec, flags.Continue)

	os.Args = []string{flags.Program, "--version"}
	ec, err = flags.Run(&buf, "1.2.3", &c)
	be.Err(t, err, nil)
	be.Equal(t, ec, flags.ExitOK)
	be.True(t, strings.HasPrefix(buf.String(), "defacto2-server 1.2.3 ("))

	buf.Reset()
	os.Args = []string{flags.Program, "routes"}
	ec, err = flags.Run(&buf, "", &c)
	be.Err(t, err, flags.ErrPrinter)
	be.Equal(t, ec, flags.GenericErr)
}

func TestVersionRelease(t *testing.T) {
	be.Equal(t, flags.VersionRelease(""), "0.0.0 αlpha")
	be.Equal(t, flags.VersionRelease("1.2.3"), "1.2.3")
	be.Equal(t, flags.VersionRelease("1.2.3-next"), "1.2.3 βeta")

	be.Equal(t, flags.VersionCommit("1.2.3"), "1.2.3")
	be.True(t, flags.VersionCommit("") != "")
	be.True(t, flags.OS() != "")
	be.True(t, flags.Arch() != "")
	be.True(t, strings.HasPrefix(flags.Copyright(), "© 2023-"))
	be.True(t, strings.Contains(flags.Copyright(), flags.Author))

	ver := flags.Version("1.2.3")
	be.True(t, strings.HasPrefix(ver, "1.2.3 ("))
	be.True(t, strings.Contains(ver, flags.OS()))
	be.True(t, strings.Contains(ver, flags.Arch()))
	be.True(t, strings.Contains(ver, runtime.Version()))

	var buf bytes.Buffer
	app := flags.App(&buf, "1.2.3", &config.Config{HTTPPort: 1323})
	be.Err(t, app.Run([]string{flags.Program, "--version"}), nil)
	out := buf.String()
	be.True(t, strings.HasPrefix(out, "defacto2-server 1.2.3 ("))
	be.True(t, !strings.Contains(out, "version"))
}

func TestAppHelpDescriptionAndWriter(t *testing.T) {
	var buf bytes.Buffer
	app := flags.App(&buf, "1.0.0", &config.Config{TLSPort: 8443})
	be.Err(t, app.Run([]string{flags.Program, "--help"}), nil)
	out := buf.String()
	be.True(t, !strings.Contains(out, "\033["))
	be.True(t, strings.Contains(out, "HTTPS port 8443"))
	be.True(t, !strings.Contains(out, "port 0"))
	be.True(t, strings.Contains(out, "https://"+flags.Domain))
	be.True(t, strings.Contains(out, flags.Repository))
	be.True(t, strings.Contains(out, "Inspection"))
	be.True(t, strings.Contains(out, "Maintenance"))

	buf.Reset()
	dual := flags.App(&buf, "1.0.0", &config.Config{HTTPPort: 1323, TLSPort: 8443})
	be.Err(t, dual.Run([]string{flags.Program, "--help"}), nil)
	be.True(t, strings.Contains(buf.String(), "HTTP port 1323 and HTTPS port 8443"))

	buf.Reset()
	be.Err(t, dual.Run([]string{flags.Program, "fix", "--help"}), nil)
	subOut := buf.String()
	be.True(t, strings.Contains(subOut, "defacto2-server fix"))
	be.True(t, strings.Contains(subOut, "--dry-run"))

	buf.Reset()
	nilCfg := flags.App(&buf, "1.0.0", nil)
	be.Err(t, nilCfg.Run([]string{flags.Program, "--help"}), nil)
	be.True(t, !strings.Contains(buf.String(), "Launch the web server"))
}

func TestCommands(t *testing.T) {
	origLog := slog.Default()
	t.Cleanup(func() { slog.SetDefault(origLog) })
	t.Setenv("TMPDIR", t.TempDir())
	t.Setenv("D2_DATABASE_URL", "invalid://")

	c := &config.Config{HTTPPort: 1323}

	t.Run("config", func(t *testing.T) {
		var buf bytes.Buffer
		app := flags.App(&buf, "1.0.0", c)
		be.Err(t, app.Run([]string{flags.Program, "config"}), nil)
		be.True(t, buf.Len() > 0)

		buf.Reset()
		be.Err(t, app.Run([]string{flags.Program, "config", "--detail"}), nil)
		be.True(t, buf.Len() > 0)
	})

	t.Run("address", func(t *testing.T) {
		var buf bytes.Buffer
		app := flags.App(&buf, "1.0.0", c)
		be.Err(t, app.Run([]string{flags.Program, "address"}), nil)
		be.True(t, buf.Len() > 0)
	})

	t.Run("routes", func(t *testing.T) {
		var buf bytes.Buffer
		noPrinter := flags.App(&buf, "1.0.0", c)
		be.Err(t, noPrinter.Run([]string{flags.Program, "routes"}), flags.ErrPrinter)

		var gotMethod string
		printer := func(w io.Writer, method string) error {
			gotMethod = method
			_, _ = io.WriteString(w, "GET /health\n")
			return nil
		}
		app := flags.App(&buf, "1.0.0", c, printer)
		be.Err(t, app.Run([]string{flags.Program, "routes", "--method", "GET"}), nil)
		be.Equal(t, gotMethod, "GET")
		be.Equal(t, buf.String(), "GET /health\n")

		failApp := flags.App(&buf, "1.0.0", c, func(io.Writer, string) error {
			return errRoutesFailed
		})
		be.Err(t, failApp.Run([]string{flags.Program, "routes"}), errRoutesFailed)
	})

	t.Run("temp dry-run", func(t *testing.T) {
		var buf bytes.Buffer
		app := flags.App(&buf, "1.0.0", c)
		be.Err(t, app.Run([]string{flags.Program, "temp", "--dry-run"}), nil)
		be.True(t, buf.Len() > 0)
	})

	t.Run("check", func(t *testing.T) {
		var buf bytes.Buffer
		app := flags.App(&buf, "1.0.0", c)
		_ = app.Run([]string{flags.Program, "check"})
		be.True(t, strings.Contains(buf.String(), "Web Server Configuration Checks"))
	})

	t.Run("fix dry-run error", func(t *testing.T) {
		var buf bytes.Buffer
		app := flags.App(&buf, "1.0.0", c)
		err := app.Run([]string{flags.Program, "fix", "--dry-run"})
		be.Err(t, err)
		be.True(t, strings.Contains(buf.String(), "Database and Asset Fixes (dry run)"))
	})

	t.Run("index dry-run error", func(t *testing.T) {
		var buf bytes.Buffer
		app := flags.App(&buf, "1.0.0", c)
		err := app.Run([]string{flags.Program, "index", "--dry-run"})
		be.Err(t, err)
		be.True(t, strings.Contains(buf.String(), "Database Index Rebuild (dry run)"))
	})

	t.Run("nil config", func(t *testing.T) {
		var buf bytes.Buffer
		app := flags.App(&buf, "1.0.0", nil)
		for _, cmd := range []string{"config", "check", "address", "temp", "fix", "index"} {
			be.Err(t, app.Run([]string{flags.Program, cmd}), flags.ErrNoConfig)
		}
	})
}
