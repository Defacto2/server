package config

import (
	"log/slog"
	"testing"

	"github.com/nalgeon/be"
)

func TestChecks(t *testing.T) {
	t.Parallel()

	c := Config{}
	err := c.checkLogDir(slog.Default())
	be.Err(t, err)
}
