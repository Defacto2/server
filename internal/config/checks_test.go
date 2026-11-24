package config_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/nalgeon/be"
)

func TestChecks(t *testing.T) {
	t.Parallel()

	var failBuf bytes.Buffer
	c := config.Config{}
	got := c.Checks(t.Context(), &failBuf)
	be.Err(t, got, config.ErrCheckFailed)
	be.True(t, strings.Contains(failBuf.String(), "Result: 5 failed"))

	tmp := t.TempDir()
	var okBuf bytes.Buffer
	valid := config.Config{
		AbsDownload:  config.Absdown(tmp),
		AbsPreview:   config.Absprev(tmp),
		AbsThumbnail: config.Absthumb(tmp),
		AbsOrphaned:  config.Absorphan(tmp),
		AbsExtra:     config.Absextra(tmp),
		DatabaseURL:  postgres.DefaultURL,
		HTTPPort:     config.StdCustom,
		ReadOnly:     true,
	}
	got = valid.Checks(t.Context(), &okBuf)
	be.Err(t, got, nil)
	out := okBuf.String()
	be.True(t, !strings.Contains(out, "\033["))
	be.True(t, strings.Contains(out, "Result: all checks passed"))
}

func TestChecksMisconfigTLSPortAndPair(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	tmp := t.TempDir()
	c := config.Config{
		AbsDownload:  config.Absdown(tmp),
		AbsPreview:   config.Absprev(tmp),
		AbsThumbnail: config.Absthumb(tmp),
		AbsOrphaned:  config.Absorphan(tmp),
		AbsExtra:     config.Absextra(tmp),
		DatabaseURL:  postgres.DefaultURL,
		HTTPPort:     config.StdCustom,
		TLSPort:      config.PortTLS(config.StdHTTPS),
		TLSCert:      config.Abstlscrt(testdata.Abs()),
		DemozooAuth:  config.DemozooAuth("badkey"),
		ReadOnly:     true,
	}
	got := c.Checks(t.Context(), &buf)
	be.Err(t, got, config.ErrCheckFailed)

	out := buf.String()
	be.True(t, strings.Contains(out, "FAIL  D2_TLS_PORT           443"))
	be.True(t, strings.Contains(out, "OK    D2_TLS_CERT"))
	be.True(t, strings.Contains(out, "FAIL  D2_TLS_KEY            not set — required when D2_TLS_CERT is set"))
	be.True(t, strings.Contains(out, "FAIL  D2_DEMOZOO_KEY        xxxxx — Authorization key is too short"))
	be.True(t, !strings.Contains(out, "badkey"))
}

func TestChecksReadOnlySkipsTools(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	c := config.Config{
		DatabaseURL: config.Connection("postgres://user:supersecret@localhost:badport/db"),
		ReadOnly:    true,
	}
	got := c.Checks(t.Context(), &buf)
	be.Err(t, got, config.ErrCheckFailed)

	out := buf.String()
	be.True(t, strings.Contains(out, "skipped in read-only mode"))
	be.True(t, !strings.Contains(out, "missing from PATH"))
	be.True(t, !strings.Contains(out, "supersecret"))
}
