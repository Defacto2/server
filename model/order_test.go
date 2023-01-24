package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrder_String(t *testing.T) {
	tests := []struct {
		o         Order
		expect    string
		assertion assert.ComparisonAssertionFunc
	}{
		{-1, "", assert.Equal},
		{NameAsc, "filename asc", assert.Equal},
		{DescDes, "record_title desc", assert.Equal},
	}
	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			tt.assertion(t, tt.expect, tt.o.String())
		})
	}
}
