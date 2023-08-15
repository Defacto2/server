package model

// Package file_mags.go contains the database queries for magazine files.

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/model/expr"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Magazine is a the model for the magazine files.
type Magazine struct {
	Bytes   int `boil:"size_sum"`
	Count   int `boil:"counter"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (m *Magazine) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.MagExpr(),
		qm.From(From)).Bind(ctx, db, m)
}

func (m *Magazine) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.MagExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}
