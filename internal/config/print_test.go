package config_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/nalgeon/be"
)

func TestPrint(t *testing.T) {
	t.Parallel()

	c := config.Config{
		DatabaseURL:   postgres.DefaultURL,
		HTTPPort:      config.StdCustom,
		SessionMaxAge: config.SessionHours,
		Compression:   true,
		ProdMode:      true,
		ReadOnly:      true,
	}

	var full bytes.Buffer
	c.Print(&full, true)
	out := full.String()

	be.True(t, !strings.Contains(out, "\033["))
	for env, desc := range c.Help() {
		be.True(t, strings.Contains(out, env))
		be.True(t, strings.Contains(out, desc))
	}

	var short bytes.Buffer
	c.Print(&short, false)
	shortOut := short.String()
	for env, desc := range c.Help() {
		be.True(t, strings.Contains(shortOut, env))
		be.True(t, !strings.Contains(shortOut, desc))
	}
}

func TestPrintMasking(t *testing.T) {
	t.Parallel()

	c := config.Config{
		AbsDownload:    config.Absdown("/does/not/exist/defacto2"),
		DatabaseURL:    config.Connection("postgres://root:supersecret@localhost:5432/defacto2_ps"),
		SessionKey:     config.Sessionkey("my-secret-session-key"),
		GoogleClientID: config.GoogleAuth("my-google-client-id"),
		GoogleIDs:      config.GoogleID("account-1,account-2"),
		DemozooAuth:    config.DemozooAuth("badkey"),
		TLSCert:        config.Abstlscrt(testdata.Dir()),
	}
	c.Override()

	var buf bytes.Buffer
	c.Print(&buf, false)
	out := buf.String()

	// Secrets must never leak in plaintext.
	be.True(t, !strings.Contains(out, "supersecret"))
	be.True(t, !strings.Contains(out, "my-secret-session-key"))
	be.True(t, !strings.Contains(out, "my-google-client-id"))
	be.True(t, !strings.Contains(out, "account-1"))
	be.True(t, !strings.Contains(out, "account-2"))
	be.True(t, !strings.Contains(out, "badkey"))

	// Masked values and issues must be displayed.
	be.True(t, strings.Contains(out, "xxxxx (multiple sign in accounts)"))
	be.True(t, strings.Contains(out, "ISSUE: Directory does not exist"))
	be.True(t, strings.Contains(out, "ISSUE: File path points to a directory"))
	be.True(t, strings.Contains(out, "ISSUE: Authorization key is too short"))
}

func TestPrintAddresses(t *testing.T) {
	t.Parallel()

	var empty bytes.Buffer
	be.Err(t, (config.Config{}).AddressGrouped(&empty), config.ErrNoPort)

	var buf bytes.Buffer
	c := config.Config{
		HTTPPort:  config.StdCustom,
		TLSPort:   config.PortTLS(config.StdHTTPS),
		MatchHost: config.Matchhost("127.0.0.1"),
	}
	be.Err(t, c.AddressGrouped(&buf), nil)
	out := buf.String()
	be.True(t, !strings.Contains(out, "\033["))
	be.True(t, strings.Contains(out, "Bound to host: 127.0.0.1"))
	be.True(t, strings.Contains(out, "HTTP (:1323)"))
	be.True(t, strings.Contains(out, "http://127.0.0.1:1323"))
	be.True(t, strings.Contains(out, "HTTPS (:443)"))
	be.True(t, strings.Contains(out, "https://127.0.0.1"))
}

func TestPurgeTemp(t *testing.T) {
	t.Setenv("TMPDIR", t.TempDir())

	var buf bytes.Buffer
	be.Err(t, (config.Config{}).PurgeTemp(&buf, true), nil)
	out := buf.String()
	be.True(t, strings.Contains(out, "Temporary Directory Purge (dry run)"))
	be.True(t, strings.Contains(out, "Found"))
	be.True(t, strings.Contains(out, "0 directories"))
}
