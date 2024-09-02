package command_test

import (
	"testing"

	"github.com/Defacto2/server/command"
	"github.com/Defacto2/server/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Parallel()
	ec, err := command.Run("", nil)
	require.Error(t, err)
	assert.Equal(t, command.UsageError, ec)
	c := config.Config{}
	ec, err = command.Run("", &c)
	require.Error(t, err)
	assert.Equal(t, command.GenericError, ec)
}
