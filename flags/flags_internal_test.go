package flags

import (
	"strings"
	"testing"

	"github.com/nalgeon/be"
)

func TestHelpAnsi(t *testing.T) {
	t.Parallel()

	be.Equal(t, ansi(false, ansiBold, "hi"), "hi")
	be.Equal(t, ansi(true, ansiBold, ""), "")
	be.Equal(t, ansi(true, ansiBold, "hi"), ansiBold+"hi"+ansiReset)

	be.True(t, strings.Contains(appHelpTemplate(true), ansiBold))
	be.True(t, !strings.Contains(appHelpTemplate(false), ansiBold))
	be.True(t, strings.Contains(commandHelpTemplate(true), ansiCyan))
	be.True(t, !strings.Contains(commandHelpTemplate(false), ansiCyan))
}

func TestSetupNilConfig(t *testing.T) {
	t.Parallel()

	ec, err := setup(nil, "", nil)
	be.Err(t, err, ErrNoConfig)
	be.Equal(t, ec, UsageErr)
}
