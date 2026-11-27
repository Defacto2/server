package config_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/nalgeon/be"
)

func TestTLS(t *testing.T) {
	t.Parallel()

	modes := []struct {
		name      string
		cfg       config.Config
		wantTLS   bool
		wantLocal bool
	}{
		{"disabled", config.Config{}, false, false},
		{"local", config.Config{TLSPort: config.PortTLS(config.StdHTTPS)}, false, true},
		{"key no port", config.Config{TLSKey: config.Abstlskey("/tmp/key.pem")}, false, false},
		{
			"tls",
			config.Config{
				TLSPort: config.PortTLS(8443),
				TLSCert: config.Abstlscrt("/tmp/cert.pem"),
				TLSKey:  config.Abstlskey("/tmp/key.pem"),
			},
			true,
			false,
		},
	}

	for _, tt := range modes {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			be.Equal(t, tt.cfg.UseTLS(), tt.wantTLS)
			be.Equal(t, tt.cfg.UseLocal(), tt.wantLocal)
		})
	}

	t.Run("port", func(t *testing.T) {
		t.Parallel()

		p := config.PortTLS(config.StdHTTPS)
		be.Equal(t, p.Value(), uint16(config.StdHTTPS))
		be.Err(t, p.Check(), config.ErrPortSys)
		be.True(t, strings.HasSuffix(p.Help(), "https://localhost"))

		be.True(t, strings.Contains(config.PortTLS(0).Help(), "not using HTTPS"))

		valid := config.PortTLS(config.PortSys + 1)
		be.Err(t, valid.Check(), nil)
		be.True(t, strings.HasSuffix(valid.Help(), "https://localhost:1025"))
	})
}

func TestTLSFiles(t *testing.T) {
	t.Parallel()

	be.Equal(t, config.Abstlscrt("").Issue(), "")
	be.Equal(t, config.Abstlskey("").Issue(), "")
	be.Equal(t, config.Abstlscrt("").Help(), "No TLS certificate is in use")
	be.Equal(t, config.Abstlskey("").Help(), "No TLS key is in use")

	be.Equal(t, config.Abstlscrt(testdata.Abs()).Issue(), "")
	be.Equal(t, config.Abstlskey(testdata.Abs()).Issue(), "")
	be.Equal(t, config.Abstlscrt(testdata.Abs()).Help(), "")
	be.Equal(t, config.Abstlskey(testdata.Abs()).Help(), "")

	be.True(t, strings.Contains(config.Abstlscrt(testdata.Dir()).Issue(), "points to a directory"))
	be.True(t, strings.Contains(config.Abstlskey(testdata.Dir()).Issue(), "points to a directory"))

	be.True(t, strings.Contains(config.Abstlscrt("missing-cert.pem").Issue(), "not exist"))
	be.True(t, strings.Contains(config.Abstlskey("missing-key.pem").Issue(), "not exist"))
}
