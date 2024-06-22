package model

import (
	"context"
	"fmt"

	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// One retrieves a single file record from the database using the record key.
// This function can return records that have been marked as deleted.
func DeleteOne(ctx context.Context, exec boil.ContextExecutor, key int) error {
	if exec == nil {
		return ErrTx
	}
	if key < 1 {
		return fmt.Errorf("key value %d: %w", key, ErrKey)
	}
	mods := models.FileWhere.ID.EQ(int64(key))
	_, err := models.Files(mods, qm.WithDeleted()).DeleteAll(ctx, exec, true)
	if err != nil {
		return fmt.Errorf("one record %d: %w", key, err)
	}
	return nil
}
