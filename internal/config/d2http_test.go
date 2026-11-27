package config_test

import (
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/nalgeon/be"
)

func TestHTTP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		port    config.PortHTTP
		wantUse bool
		wantErr error
	}{
		{"disabled", 0, false, nil},
		{"http", config.PortHTTP(config.StdHTTP), true, config.ErrPortSys},
		{"port sys", config.PortSys, true, config.ErrPortSys},
		{"min valid", config.PortSys + 1, true, nil},
		{"max valid", config.PortMax, true, nil},
		{"too high", config.PortMax + 1, true, config.ErrPortMax},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := config.Config{HTTPPort: tt.port}
			be.Equal(t, c.UseHTTP(), tt.wantUse)
			be.Equal(t, c.HTTPPort.Value(), uint16(tt.port))
			be.Err(t, c.HTTPPort.Check(), tt.wantErr)
		})
	}
}
