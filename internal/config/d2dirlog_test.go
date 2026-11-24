package config_test

import (
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/nalgeon/be"
)

func TestAbsLog(t *testing.T) {
	t.Parallel()

	var empty config.Abslog
	be.Equal(t, "No logs will be saved", empty.Help())
	be.Equal(t, "", empty.Issue())

	tmp := t.TempDir()
	set := config.Abslog(tmp)
	be.Equal(t, "", set.Help())
	be.Equal(t, "", set.Issue())
	be.Equal(t, tmp, set.String())
}
