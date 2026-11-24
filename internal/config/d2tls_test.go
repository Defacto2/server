package config_test

import (
	"strings"
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

	// tls requires a non-zero port even when a key is set
	noPort := config.Config{
		TLSKey: config.Abstlskey("/tmp/key.pem"),
	}
	be.True(t, !noPort.UseTLS())

	be.Equal(t, c.TLSPort.Value(), HTTPS)

	be.Err(t, c.TLSPort.Check(), config.ErrPortSys)

	const valid = config.PortSys + 1

	c.TLSPort = valid
	be.Err(t, c.TLSPort.Check(), nil)
}

func TestTLSFiles(t *testing.T) {
	t.Parallel()

	be.Equal(t, "", config.Abstlscrt(testdata.Abs()).Issue())
	be.Equal(t, "", config.Abstlskey(testdata.Abs()).Issue())

	be.True(t, strings.Contains(config.Abstlscrt(testdata.Dir()).Issue(), "points to a directory"))
	be.True(t, strings.Contains(config.Abstlskey(testdata.Dir()).Issue(), "points to a directory"))

	be.True(t, strings.Contains(config.Abstlscrt("missing-cert.pem").Issue(), "not exist"))
	be.True(t, strings.Contains(config.Abstlskey("missing-key.pem").Issue(), "not exist"))
}
