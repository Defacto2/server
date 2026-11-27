package config_test

import (
	"bytes"
	"crypto/sha512"
	"log/slog"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/nalgeon/be"
)

func TestOverride(t *testing.T) {
	t.Parallel()

	t.Run("hashes and defaults port", func(t *testing.T) {
		t.Parallel()

		c := config.Config{GoogleIDs: config.GoogleID(" abcde , , zxcvb ")}
		c.Override()

		if len(c.GoogleAccounts) != 2 {
			t.Fatalf("got %d accounts, want 2", len(c.GoogleAccounts))
		}
		be.Equal(t, c.GoogleAccounts[0], sha512.Sum384([]byte("abcde")))
		be.Equal(t, c.GoogleAccounts[1], sha512.Sum384([]byte("zxcvb")))
		be.Equal(t, c.GoogleAccounts.String(), "multiple sign in accounts")
		be.Equal(t, c.GoogleIDs.String(), "")
		be.Equal(t, c.HTTPPort, config.StdCustom)
	})

	t.Run("clears previous accounts", func(t *testing.T) {
		t.Parallel()

		c := config.Config{GoogleIDs: config.GoogleID("one")}
		c.Override()
		be.Equal(t, c.GoogleAccounts.String(), "single sign in account")

		c.Override()
		be.Equal(t, len(c.GoogleAccounts), 0)
		be.Equal(t, c.GoogleAccounts.String(), "")
	})

	t.Run("keeps tls-only port", func(t *testing.T) {
		t.Parallel()

		c := config.Config{TLSPort: config.PortTLS(8443)}
		c.Override()
		be.Equal(t, c.HTTPPort, config.PortHTTP(0))
	})

	t.Run("keeps custom http port", func(t *testing.T) {
		t.Parallel()

		c := config.Config{HTTPPort: config.PortHTTP(9090)}
		c.Override()
		be.Equal(t, c.HTTPPort, config.PortHTTP(9090))
	})
}

func TestGoogleLogging(t *testing.T) {
	t.Parallel()

	be.Equal(t, config.GoogleAuth("").LogValue().String(), "")
	be.Equal(t, config.GoogleAuth("client-id").LogValue().String(), "xxxxx")
	be.Equal(t, config.GoogleID("").LogValue().String(), "")
	be.Equal(t, config.GoogleID("account-id").LogValue().String(), "xxxxx")

	var buf bytes.Buffer
	sl := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	c := config.Config{}
	c.Configurations(sl)
	out := buf.String()
	be.True(t, strings.Contains(out, `msg="Google account(s):"`))
	be.True(t, strings.Contains(out, `D2_GOOGLE_IDS=""`))
	be.True(t, strings.Contains(out, `D2_GOOGLE_CLIENT_ID=""`))

	buf.Reset()
	c.GoogleIDs = config.GoogleID("account-one")
	c.GoogleClientID = config.GoogleAuth("secret-client")
	c.Override()
	c.Configurations(sl)
	out = buf.String()
	be.True(t, strings.Contains(out, `msg="Google account(s), single sign in account:"`))
	be.True(t, strings.Contains(out, `D2_GOOGLE_IDS=xxxxx`))
	be.True(t, strings.Contains(out, `D2_GOOGLE_CLIENT_ID=xxxxx`))
	be.True(t, !strings.Contains(out, "account-one"))
	be.True(t, !strings.Contains(out, "secret-client"))
}
