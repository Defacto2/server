package cmd_test

import (
	"testing"

	"github.com/Defacto2/server/cmd"
	"github.com/Defacto2/server/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	t.Parallel()
	ec, err := cmd.Run("", nil)
	assert.Error(t, err)
	assert.Equal(t, ec, cmd.UsageError)
	c := config.Config{}
	ec, err = cmd.Run("", &c)
	assert.Error(t, err)
	assert.Equal(t, ec, cmd.GenericError)
}

func TestVersion(t *testing.T) {
	s := cmd.Version("")
	assert.Contains(t, s, "not a build")
	s = cmd.Version("1.2.3")
	assert.Contains(t, s, "1.2.3")

	s = cmd.Vers("")
	assert.Contains(t, s, "v0.0.0")
}
