package model

// Package file file_nfo_files_for.go contains the database queries for NFO files, tools and release proofs.

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/model/modext"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type Nfo struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
	Year0 int `boil:"min_year"`
	YearX int `boil:"max_year"`
}

// Stat counts the total number and total byte size of releases of NFO files.
func (n *Nfo) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter, postgres.MinYear, postgres.MaxYear),
		modext.NfoExpr(),
		qm.From(From)).Bind(ctx, db, n)
}

// List returns a list of NFO files.
func (n *Nfo) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.NfoExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

type NfoTool struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of of NFO tool releases.
func (n *NfoTool) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.NfoToolExpr(),
		qm.From(From)).Bind(ctx, db, n)
}

// List returns a list of NFO tools.
func (n *NfoTool) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.NfoToolExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

type Proof struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of releases of file proofs.
func (p *Proof) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.ProofExpr(),
		qm.From(From)).Bind(ctx, db, p)
}

// List returns a list of file proofs.
func (p *Proof) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.ProofExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}
