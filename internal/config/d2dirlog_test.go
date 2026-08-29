package config_test

import (
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/nalgeon/be"
)

func TestAbsLog(t *testing.T) {
	t.Parallel()

	c := config.Config{}
	got := c.LogStore()
	be.Err(t, got, nil)

	tmp := t.TempDir()
	c.AbsLog = config.Abslog(tmp)
	got = c.LogStore()
	be.Err(t, got, nil)
	be.Equal(t, c.AbsLog.String(), tmp)
}
