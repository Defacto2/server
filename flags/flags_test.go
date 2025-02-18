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
	ec, err := flags.Run(nil, "", nil)
	require.Error(t, err)
	assert.Equal(t, flags.UsageError, ec)
	c := config.Config{}
	ec, err = flags.Run(nil, "", &c)
	require.Error(t, err)
	assert.Equal(t, flags.GenericError, ec)
}

func TestVers(t *testing.T) {
	t.Parallel()
	s := flags.Vers("")
	assert.Equal(t, "defacto2-server version 0.0.0 αlpha", s)
	s = flags.Vers("1.2.3")
	assert.Equal(t, "defacto2-server version 1.2.3", s)
	s = flags.Vers("1.2.3-next")
	assert.Equal(t, "defacto2-server version 1.2.3 βeta", s)
}
