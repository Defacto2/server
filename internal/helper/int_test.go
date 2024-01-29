package helper_test

import (
	"testing"

	"github.com/Defacto2/server/internal/helper"
	"github.com/stretchr/testify/assert"
)

func TestAdd1(t *testing.T) {
	tests := []struct {
		a         any
		expect    int64
		assertion assert.ComparisonAssertionFunc
	}{
		{0, 1, assert.Equal},
		{"xyz", 0, assert.Equal},
		{123, 124, assert.Equal},
		{1234567890, 1234567891, assert.Equal},
		{1234567890123456789, 1234567890123456790, assert.Equal},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.expect, helper.Add1(tt.a))
		})
	}
}
