package model

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/model/modext"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Package file database.go contains the database queries for the collections of databases.

type Database struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of the database releases.
func (d *Database) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.DatabaseExpr(),
		qm.From(From)).Bind(ctx, db, d)
}

// List returns a list of database releases.
func (d *Database) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.DatabaseExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}
