package logger_test

import (
	"testing"

	"github.com/Defacto2/server/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestCLI(t *testing.T) {
	z := logger.CLI()
	assert.NotNil(t, z)
}

func TestLog(t *testing.T) {
	z := logger.Development()
	assert.NotNil(t, z)
}

func TestProduction(t *testing.T) {
	z := logger.Production("")
	assert.NotNil(t, z)
}
