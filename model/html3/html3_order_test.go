package html3_test

import (
	"testing"

	"github.com/Defacto2/server/model/html3"
	"github.com/stretchr/testify/assert"
)

func TestOrder_String(t *testing.T) {
	tests := []struct {
		o         html3.Order
		expect    string
		assertion assert.ComparisonAssertionFunc
	}{
		{-1, "", assert.Equal},
		{html3.NameAsc, "filename asc", assert.Equal},
		{html3.DescDes, "record_title desc", assert.Equal},
	}
	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			tt.assertion(t, tt.expect, tt.o.String())
		})
	}
}
