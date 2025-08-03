package flags_test

import (
	"testing"

	"github.com/Defacto2/server/flags"
	"github.com/Defacto2/server/internal/config"
	"github.com/nalgeon/be"
)

func TestRun(t *testing.T) {
	t.Parallel()
	ec, err := flags.Run(nil, "", nil)
	be.Err(t, err)
	be.Equal(t, flags.UsageErr, ec)
	c := config.Config{}
	ec, err = flags.Run(nil, "", &c)
	be.Err(t, err)
	be.Equal(t, flags.GenericErr, ec)
}

func TestVers(t *testing.T) {
	t.Parallel()
	s := flags.Vers("")
	be.Equal(t, "defacto2-server version 0.0.0 αlpha", s)
	s = flags.Vers("1.2.3")
	be.Equal(t, "defacto2-server version 1.2.3", s)
	s = flags.Vers("1.2.3-next")
	be.Equal(t, "defacto2-server version 1.2.3 βeta", s)
}
