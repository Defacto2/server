package model_test

import (
	"testing"

	"github.com/Defacto2/server/model"
	"github.com/stretchr/testify/assert"
)

func TestOrder_String(t *testing.T) {
	tests := []struct {
		o         model.Order
		expect    string
		assertion assert.ComparisonAssertionFunc
	}{
		{-1, "", assert.Equal},
		{model.NameAsc, "filename asc", assert.Equal},
		{model.DescDes, "record_title desc", assert.Equal},
	}
	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			tt.assertion(t, tt.expect, tt.o.String())
		})
	}
}
