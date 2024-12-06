package zaplog_test

import (
	"testing"

	"github.com/Defacto2/server/internal/zaplog"
	"github.com/stretchr/testify/assert"
)

func TestCLI(t *testing.T) {
	logr := zaplog.Status()
	assert.NotNil(t, logr)
}

func TestLog(t *testing.T) {
	logr := zaplog.Debug()
	assert.NotNil(t, logr)
	logr = zaplog.Timestamp()
	assert.NotNil(t, logr)
}

func TestProduction(t *testing.T) {
	logr := zaplog.Store(zaplog.Json(), "")
	assert.NotNil(t, logr)
}
