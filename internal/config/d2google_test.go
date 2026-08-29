package config_test

import (
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
