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
// preserving multi-word names like "Write-mode tools".
func parseCheckOutput(out string) map[string]checkRow {
	rows := make(map[string]checkRow)
	for line := range strings.SplitSeq(out, "\n") {
		rest, ok := strings.CutPrefix(line, "  ")
		if !ok || len(rest) < 6 {
			continue
		}
		status := strings.TrimSpace(rest[:4])
		body := strings.TrimSpace(rest[6:])
		name, detail, found := strings.Cut(body, "  ")
		if !found {
			name, detail, found = strings.Cut(body, " ")
		}
		if !found {
			continue
		}
		rows[strings.TrimSpace(name)] = checkRow{
			status: status,
			detail: strings.TrimSpace(detail),
		}
	}
	return rows
}

func TestChecks(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		c := config.Config{}
		got := c.Checks(t.Context(), &buf)
		be.Err(t, got, config.ErrCheckFailed)

		out := buf.String()
		be.True(t, strings.Contains(out, "Result: 5 failed"))
		rows := parseCheckOutput(out)
		for _, env := range []string{
			"D2_DIR_DOWNLOAD",
			"D2_DIR_PREVIEW",
			"D2_DIR_THUMBNAIL",
			"D2_DIR_EXTRA",
			"D2_DIR_ORPHANED",
		} {
			be.Equal(t, rows[env].status, "FAIL")
		}
	})

	t.Run("valid read-only", func(t *testing.T) {
		t.Parallel()

		tmp := t.TempDir()
		var buf bytes.Buffer
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
		got := valid.Checks(t.Context(), &buf)
		be.Err(t, got, nil)

		out := buf.String()
		rows := parseCheckOutput(out)
		be.True(t, !strings.Contains(out, "\033["))
		be.Equal(t, rows["D2_READ_ONLY"].status, "WARN")
		be.True(t, strings.Contains(out, "Result: all checks passed (1 warning)"))
	})
}

func TestChecksTLS(t *testing.T) {
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

func TestChecksReadOnly(t *testing.T) {
	t.Parallel()

	t.Run("skips tools", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		c := config.Config{ReadOnly: true}
		got := c.Checks(t.Context(), &buf)
		be.Err(t, got, config.ErrCheckFailed)

		out := buf.String()
		rows := parseCheckOutput(out)
		be.Equal(t, rows["Write-mode tools"].status, "SKIP")
		be.True(t, strings.Contains(rows["Write-mode tools"].detail, "skipped in read-only mode"))
		be.True(t, !strings.Contains(out, "missing from PATH"))
	})

	t.Run("masks bad db url", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		c := config.Config{
			DatabaseURL: config.Connection("postgres://user:supersecret@localhost:badport/db"),
			ReadOnly:    true,
		}
		got := c.Checks(t.Context(), &buf)
		be.Err(t, got, config.ErrCheckFailed)

		out := buf.String()
		rows := parseCheckOutput(out)
		be.Equal(t, rows["D2_DATABASE_URL"].status, "WARN")
		be.True(t, !strings.Contains(out, "supersecret"))
	})
}

func validReadOnlyConfig(tmp string) config.Config {
	return config.Config{
		AbsDownload:  config.Absdown(tmp),
		AbsPreview:   config.Absprev(tmp),
		AbsThumbnail: config.Absthumb(tmp),
		AbsOrphaned:  config.Absorphan(tmp),
		AbsExtra:     config.Absextra(tmp),
		DatabaseURL:  postgres.DefaultURL,
		HTTPPort:     config.StdCustom,
		ReadOnly:     true,
	}
}

func TestChecksProdAuth(t *testing.T) {
	t.Parallel()

	t.Run("missing", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		c := validReadOnlyConfig(t.TempDir())
		c.ProdMode = true
		c.ReadOnly = false

		be.Err(t, c.Checks(t.Context(), &buf), nil)
		rows := parseCheckOutput(buf.String())
		be.Equal(t, rows["D2_GOOGLE_CLIENT_ID"].status, "WARN")
		be.Equal(t, rows["D2_GOOGLE_IDS"].status, "WARN")
		be.Equal(t, rows["D2_SESSION_MAX_AGE"].status, "WARN")
	})

	t.Run("configured", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		c := validReadOnlyConfig(t.TempDir())
		c.ProdMode = true
		c.ReadOnly = false
		c.GoogleClientID = config.GoogleAuth("client-secret-id")
		c.GoogleIDs = config.GoogleID("admin-secret-id")
		c.Override()
		c.SessionMaxAge = config.Hours(24)

		be.Err(t, c.Checks(t.Context(), &buf), nil)
		out := buf.String()
		rows := parseCheckOutput(out)
		be.Equal(t, rows["D2_GOOGLE_CLIENT_ID"].status, "OK")
		be.Equal(t, rows["D2_GOOGLE_IDS"].status, "OK")
		be.True(t, strings.Contains(rows["D2_GOOGLE_IDS"].detail, "single sign in account"))
		be.Equal(t, rows["D2_SESSION_MAX_AGE"].status, "OK")
		be.True(t, !strings.Contains(out, "client-secret-id"))
		be.True(t, !strings.Contains(out, "admin-secret-id"))
	})
}

func TestChecksLogDir(t *testing.T) {
	t.Parallel()

	t.Run("writable", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		tmp := t.TempDir()
		c := validReadOnlyConfig(tmp)
		c.AbsLog = config.Abslog(tmp)

		be.Err(t, c.Checks(t.Context(), &buf), nil)
		rows := parseCheckOutput(buf.String())
		be.Equal(t, rows["D2_DIR_LOG"].status, "OK")
		be.True(t, strings.Contains(rows["D2_DIR_LOG"].detail, "(writable)"))
	})

	t.Run("missing", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		tmp := t.TempDir()
		c := validReadOnlyConfig(tmp)
		c.AbsLog = config.Abslog(filepath.Join(tmp, "no-such-dir"))

		be.Err(t, c.Checks(t.Context(), &buf), config.ErrCheckFailed)
		rows := parseCheckOutput(buf.String())
		be.Equal(t, rows["D2_DIR_LOG"].status, "FAIL")
	})

	t.Run("file", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		tmp := t.TempDir()
		f := filepath.Join(tmp, "not-a-dir")
		be.Err(t, os.WriteFile(f, []byte("x"), 0o600), nil)

		c := validReadOnlyConfig(tmp)
		c.AbsLog = config.Abslog(f)

		be.Err(t, c.Checks(t.Context(), &buf), config.ErrCheckFailed)
		rows := parseCheckOutput(buf.String())
		be.Equal(t, rows["D2_DIR_LOG"].status, "FAIL")
	})

	t.Run("read-only", func(t *testing.T) {
		t.Parallel()
		if os.Geteuid() == 0 {
			t.Skip("root bypasses directory write permissions")
		}

		var buf bytes.Buffer
		tmp := t.TempDir()
		roDir := filepath.Join(tmp, "ro")
		be.Err(t, os.Mkdir(roDir, 0o500), nil)
		t.Cleanup(func() { _ = os.Chmod(roDir, 0o700) })

		c := validReadOnlyConfig(tmp)
		c.AbsLog = config.Abslog(roDir)

		be.Err(t, c.Checks(t.Context(), &buf), config.ErrCheckFailed)
		rows := parseCheckOutput(buf.String())
		be.Equal(t, rows["D2_DIR_LOG"].status, "FAIL")
	})
}

func TestChecksSecrets(t *testing.T) {
	t.Parallel()

	t.Run("unexpected db scheme", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		c := validReadOnlyConfig(t.TempDir())
		c.DatabaseURL = config.Connection("mysql://user:supersecret@localhost/db")

		be.Err(t, c.Checks(t.Context(), &buf), nil)
		out := buf.String()
		rows := parseCheckOutput(out)
		be.Equal(t, rows["D2_DATABASE_URL"].status, "WARN")
		be.True(t, strings.Contains(rows["D2_DATABASE_URL"].detail, "unexpected scheme"))
		be.True(t, !strings.Contains(out, "supersecret"))
	})

	t.Run("valid demozoo key", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		c := validReadOnlyConfig(t.TempDir())
		c.DemozooAuth = config.DemozooAuth("12345678.topsecret")

		be.Err(t, c.Checks(t.Context(), &buf), nil)
		out := buf.String()
		rows := parseCheckOutput(out)
		be.Equal(t, rows["D2_DEMOZOO_KEY"].status, "OK")
		be.True(t, !strings.Contains(out, "topsecret"))
	})
}
