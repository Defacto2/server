package zaplog_test

import (
	"testing"

	"github.com/Defacto2/server/internal/zaplog"
	"github.com/nalgeon/be"
)

func TestCLI(t *testing.T) {
	logr := zaplog.Status()
	be.True(t, logr != nil)
}

func TestLog(t *testing.T) {
	logr := zaplog.Debug()
	be.True(t, logr != nil)
	logr = zaplog.Timestamp()
	be.True(t, logr != nil)
}

func TestProduction(t *testing.T) {
	logr := zaplog.Store(zaplog.JSON(), "")
	be.True(t, logr != nil)
}
