package config

import (
	"log/slog"
	"os"
	"testing"

	"github.com/nalgeon/be"
)

func TestChecks(t *testing.T) {
	t.Parallel()

	c := Config{}
	err := c.checkLogDir(slog.Default())
	be.Err(t, err, nil)

	st, err := os.Stat(c.AbsLog.String())
	be.Err(t, err, nil)
	be.True(t, st.IsDir())
}
