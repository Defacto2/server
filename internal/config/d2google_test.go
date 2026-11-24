package config_test

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/nalgeon/be"
)

func TestOverride(t *testing.T) {
	t.Parallel()

	c := config.Config{}
	c.GoogleIDs = config.GoogleID("abcde,zxcvb")
	be.True(t, len(c.GoogleAccounts) == 0)

	c.Override() // hash and replace configs

	const hashed = 48
	be.True(t, len(c.GoogleAccounts) == 2)
	be.True(t, len(c.GoogleAccounts[0]) == hashed)
	be.True(t, len(c.GoogleAccounts[1]) == hashed)

	be.True(t, len(c.GoogleIDs) == 0)
	be.True(t, c.HTTPPort == config.StdCustom)
}

func TestGoogleLogging(t *testing.T) {
	t.Parallel()

	be.Equal(t, "", config.GoogleAuth("").LogValue().String())
	be.Equal(t, "xxxxx", config.GoogleAuth("client-id").LogValue().String())

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
	c.Override()
	c.Configurations(sl)
	out = buf.String()
	be.True(t, strings.Contains(out, `msg="Google account(s), single sign in account:"`))
	be.True(t, strings.Contains(out, `D2_GOOGLE_IDS=xxxxx`))
}
