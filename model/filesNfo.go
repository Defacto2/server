package model

// Package file fileNfo.go contains the database queries for NFO files, tools and release proofs.

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model/expr"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Nfo is a the model for the NFO files.
type Nfo struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (n *Nfo) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.NfoExpr(),
		qm.From(From)).Bind(ctx, db, n)
}

func (n *Nfo) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.NfoExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// NfoTool is a the model for the NFO tools.
type NfoTool struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (n *NfoTool) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.NfoToolExpr(),
		qm.From(From)).Bind(ctx, db, n)
}

func (n *NfoTool) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.NfoToolExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// Proof is a the model for the file proofs.
type Proof struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (p *Proof) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.ProofExpr(),
		qm.From(From)).Bind(ctx, db, p)
}

func (p *Proof) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.ProofExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}
