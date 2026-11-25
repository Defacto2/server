package config_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/nalgeon/be"
)

type checkRow struct {
	status string
	detail string
}

// parseCheckOutput extracts indented "STATUS  NAME  detail..." rows into a map keyed by NAME,
// so tests do not depend on column padding or separator formatting.
func parseCheckOutput(out string) map[string]checkRow {
	rows := make(map[string]checkRow)
	for line := range strings.SplitSeq(out, "\n") {
		if !strings.HasPrefix(line, "  ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		rows[fields[1]] = checkRow{
			status: fields[0],
			detail: strings.Join(fields[2:], " "),
		}
	}
	return rows
}

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
	certFile := filepath.Join(tmp, "cert.pem")
	be.Err(t, os.WriteFile(certFile, []byte("dummy"), 0o600), nil)

	c := config.Config{
		AbsDownload:  config.Absdown(tmp),
		AbsPreview:   config.Absprev(tmp),
		AbsThumbnail: config.Absthumb(tmp),
		AbsOrphaned:  config.Absorphan(tmp),
		AbsExtra:     config.Absextra(tmp),
		DatabaseURL:  postgres.DefaultURL,
		HTTPPort:     config.StdCustom,
		TLSPort:      config.PortTLS(config.StdHTTPS),
		TLSCert:      config.Abstlscrt(certFile),
		DemozooAuth:  config.DemozooAuth("badkey"),
		ReadOnly:     true,
	}
	got := c.Checks(t.Context(), &buf)
	be.Err(t, got, config.ErrCheckFailed)

	out := buf.String()
	rows := parseCheckOutput(out)

	be.Equal(t, rows["D2_TLS_PORT"].status, "FAIL")
	be.True(t, strings.Contains(rows["D2_TLS_PORT"].detail, "443"))

	be.Equal(t, rows["D2_TLS_CERT"].status, "OK")

	be.Equal(t, rows["D2_TLS_KEY"].status, "FAIL")
	be.True(t, strings.Contains(rows["D2_TLS_KEY"].detail, "D2_TLS_CERT"))

	be.Equal(t, rows["D2_DEMOZOO_KEY"].status, "FAIL")
	be.True(t, strings.Contains(rows["D2_DEMOZOO_KEY"].detail, "too short"))
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
