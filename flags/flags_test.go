package flags_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Defacto2/server/flags"
	"github.com/Defacto2/server/internal/config"
	"github.com/nalgeon/be"
)

func TestRun(t *testing.T) {
	t.Parallel()

	ec, err := flags.Run(nil, "", nil)
	be.Err(t, err)
	be.Equal(t, flags.UsageErr, ec)

	c := config.Config{}
	ec, err = flags.Run(nil, "", &c)
	be.Err(t, err)
	be.Equal(t, flags.GenericErr, ec)
}

func TestVersionRelease(t *testing.T) {
	t.Parallel()

	s := flags.VersionRelease("")
	be.Equal(t, "0.0.0 αlpha", s)

	s = flags.VersionRelease("1.2.3")
	be.Equal(t, "1.2.3", s)

	s = flags.VersionRelease("1.2.3-next")
	be.Equal(t, "1.2.3 βeta", s)

	var buf bytes.Buffer
	app := flags.App(&buf, "1.2.3", &config.Config{HTTPPort: 1323})
	be.Err(t, app.Run([]string{flags.Program, "--version"}), nil)
	out := buf.String()
	be.True(t, strings.HasPrefix(out, "defacto2-server 1.2.3 ("))
	be.True(t, !strings.Contains(out, "version"))
}

func TestAppHelpDescriptionAndWriter(t *testing.T) {
	t.Parallel()

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
}
