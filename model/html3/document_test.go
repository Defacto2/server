package html3_test

import (
	"context"
	"testing"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model/html3"
	"github.com/stretchr/testify/assert"
)

func TestDocuments_Stat(t *testing.T) {
	t.Parallel()
	a := html3.Documents{}
	ctx := context.TODO()
	err := a.Stat(ctx, nil)
	assert.ErrorIs(t, err, html3.ErrDB)
	db, err := postgres.ConnectDB()
	assert.NoError(t, err)
	defer db.Close()
	err = a.Stat(ctx, db)
	assert.Error(t, err)
}
