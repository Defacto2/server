package model

// Package file file_ansi_art.go contains the database queries for ANSI art.

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/model/modext"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type Ansi struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
	Year0 int `boil:"min_year"`
	YearX int `boil:"max_year"`
}

// Stat counts the total number and total byte size of releases ANSI formatted text and art files.
func (a *Ansi) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter, postgres.MinYear, postgres.MaxYear),
		modext.AnsiExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

// List returns a list of ANSI formatted text and art files.
func (a *Ansi) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.AnsiExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

type AnsiBrand struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of brand logos created in ANSI text.
func (a *AnsiBrand) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.AnsiBrandExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

// List returns a list of brand logos created in ANSI text.
func (a *AnsiBrand) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.AnsiBrandExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

type AnsiBBS struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of BBS advertisements created in ANSI text.
func (a *AnsiBBS) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.AnsiBBSExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

// List returns a list of BBS advertisements created in ANSI text.
func (a *AnsiBBS) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.AnsiBBSExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

type AnsiFTP struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of FTP advertisements created in ANSI text.
func (a *AnsiFTP) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.AnsiFTPExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

// List returns a list of FTP advertisements created in ANSI text.
func (a *AnsiFTP) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.AnsiFTPExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

type AnsiNfo struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of NFO files created in ANSI text.
func (a *AnsiNfo) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.AnsiNfoExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

// List returns a list of NFO files created in ANSI text.
func (a *AnsiNfo) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.AnsiNfoExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}
