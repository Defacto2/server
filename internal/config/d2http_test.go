package config_test

import (
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/nalgeon/be"
)

func TestHTTP(t *testing.T) {
	t.Parallel()

	const HTTP = 80
	c := config.Config{HTTPPort: HTTP}

	be.True(t, c.UseHTTP())

	be.Equal(t, c.HTTPPort.Value(), HTTP)

	be.Err(t, c.HTTPPort.Check(), config.ErrPortSys)

	const valid = config.PortSys + 1

	c.HTTPPort = valid
	be.Err(t, c.HTTPPort.Check(), nil)
}
