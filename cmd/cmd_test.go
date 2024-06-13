package cmd_test

import (
	"testing"

	"github.com/Defacto2/server/cmd"
	"github.com/Defacto2/server/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Parallel()
	ec, err := cmd.Run("", nil)
	require.Error(t, err)
	assert.Equal(t, cmd.UsageError, ec)
	c := config.Config{}
	ec, err = cmd.Run("", &c)
	require.Error(t, err)
	assert.Equal(t, cmd.GenericError, ec)
}
