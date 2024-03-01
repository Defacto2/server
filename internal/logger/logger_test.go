package logger_test

import (
	"testing"

	"github.com/Defacto2/server/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestCLI(t *testing.T) {
	logr := logger.CLI()
	assert.NotNil(t, logr)
}

func TestLog(t *testing.T) {
	logr := logger.Development()
	assert.NotNil(t, logr)
}

func TestProduction(t *testing.T) {
	logr := logger.Production("")
	assert.NotNil(t, logr)
}
