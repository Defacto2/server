package config_test

import (
	"log/slog"
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/nalgeon/be"
)

func TestChecks(t *testing.T) {
	t.Parallel()

	c := config.Config{}
	got := c.Checks(t.Context(), slog.Default())
	be.Err(t, got, nil)
}
