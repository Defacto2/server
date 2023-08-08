package model

// Package file_database.go contains the database queries for the collections of databases.

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/model/expr"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Database is a the model for the database releases.
type Database struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

func (d *Database) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Stat()...),
		expr.DatabaseExpr(),
		qm.From(From)).Bind(ctx, db, d)
}

func (d *Database) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.DatabaseExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}
