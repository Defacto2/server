package config_test

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/nalgeon/be"
)

func TestChecks(t *testing.T) {
	t.Parallel()

	c := config.Config{}
	got := c.Checks(t.Context(), slog.Default())
	be.Err(t, got, config.ErrCheckFailed)

	tmp := t.TempDir()
	valid := config.Config{
		AbsDownload:  config.Absdown(tmp),
		AbsPreview:   config.Absprev(tmp),
		AbsThumbnail: config.Absthumb(tmp),
		AbsOrphaned:  config.Absorphan(tmp),
		AbsExtra:     config.Absextra(tmp),
		HTTPPort:     config.StdCustom,
		ReadOnly:     true,
	}
	got = valid.Checks(t.Context(), slog.Default())
	be.Err(t, got, nil)
}

func TestChecksMisconfigTLSPort(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	sl := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	c := config.Config{
		HTTPPort: config.StdCustom,
		TLSPort:  config.PortTLS(config.StdHTTPS),
	}
	got := c.Checks(t.Context(), sl)
	be.Err(t, got, config.ErrCheckFailed)

	out := buf.String()
	be.True(t, strings.Contains(out, "tls_port=443"))
	be.True(t, !strings.Contains(out, "tls_port=1323"))
}

func TestChecksReadOnlySkipsTools(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	sl := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	c := config.Config{
		ReadOnly: true,
	}
	got := c.Checks(t.Context(), sl)
	be.Err(t, got, config.ErrCheckFailed)

	out := buf.String()
	be.True(t, !strings.Contains(out, "missing tool"))
	be.True(t, !strings.Contains(out, "WRITE MODE"))
}
