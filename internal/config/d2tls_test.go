package config_test

import (
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/nalgeon/be"
)

func TestTLS(t *testing.T) {
	t.Parallel()

	const HTTPS = 443
	c := config.Config{
		TLSPort: HTTPS,
	}

	be.True(t, c.UseLocal())

	// tls requires certs
	be.True(t, !c.UseTLS())

	be.Equal(t, c.TLSPort.Value(), HTTPS)

	be.Err(t, c.TLSPort.Check(), config.ErrPortSys)

	const valid = config.PortSys + 1

	c.TLSPort = valid
	be.Err(t, c.TLSPort.Check(), nil)
}
