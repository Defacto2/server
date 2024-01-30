package html3_test

import (
	"context"
	"testing"

	"github.com/Defacto2/server/internal/postgres"
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

func TestOrder(t *testing.T) {
	var o html3.Order
	ctx := context.TODO()
	fs, err := o.Everything(ctx, nil, 0, 0)
	assert.ErrorIs(t, err, html3.ErrDB)
	assert.Empty(t, fs)
	db, err := postgres.ConnectDB()
	assert.NoError(t, err)
	defer db.Close()

	fs, err = o.Everything(ctx, db, 0, 0)
	assert.Error(t, err)
	assert.Empty(t, fs)

	fs, err = o.ByCategory(ctx, db, 0, 0, "")
	assert.Error(t, err)
	assert.Empty(t, fs)

	fs, err = o.ByPlatform(ctx, db, 0, 0, "")
	assert.Error(t, err)
	assert.Empty(t, fs)

	fs, err = o.ByGroup(ctx, db, "")
	assert.Error(t, err)
	assert.Empty(t, fs)

	fs, err = o.Art(ctx, db, 0, 0)
	assert.Error(t, err)
	assert.Empty(t, fs)

	fs, err = o.Document(ctx, db, 0, 0)
	assert.Error(t, err)
	assert.Empty(t, fs)

	fs, err = o.Software(ctx, db, 0, 0)
	assert.Error(t, err)
	assert.Empty(t, fs)
}
