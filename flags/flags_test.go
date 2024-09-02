package flags_test

import (
	"testing"

	"github.com/Defacto2/server/flags"
	"github.com/Defacto2/server/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Parallel()
	ec, err := flags.Run("", nil)
	require.Error(t, err)
	assert.Equal(t, flags.UsageError, ec)
	c := config.Config{}
	ec, err = flags.Run("", &c)
	require.Error(t, err)
	assert.Equal(t, flags.GenericError, ec)
}
