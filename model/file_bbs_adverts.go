package model

// Package file file_bbs_adverts.go contains the database queries for BBS and FTP files.

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/model/modext"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type BBS struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
	Year0 int `boil:"min_year"`
	YearX int `boil:"max_year"`
}

// Stat counts the total number and total byte size of releases BBS files.
func (b *BBS) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter, postgres.MinYear, postgres.MaxYear),
		modext.BBSExpr(),
		qm.From(From)).Bind(ctx, db, b)
}

// List returns a list of BBS files.
func (b *BBS) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.BBSExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

type BBStro struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of of BBStro releases.
func (b *BBStro) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.BBStroExpr(),
		qm.From(From)).Bind(ctx, db, b)
}

// List returns a list of BBStro files.
func (b *BBStro) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.BBStroExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

type BBSImage struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of of the BBS images.
func (b *BBSImage) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.BBSImageExpr(),
		qm.From(From)).Bind(ctx, db, b)
}

// List returns a list of BBS images.
func (b *BBSImage) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.BBSImageExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

type BBSText struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of of the BBS related text files.
func (b *BBSText) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.BBSTextExpr(),
		qm.From(From)).Bind(ctx, db, b)
}

// List returns a list of BBS text files.
func (b *BBSText) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.BBSTextExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

type FTP struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of of the FTP files.
func (f *FTP) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.FTPExpr(),
		qm.From(From)).Bind(ctx, db, f)
}

// List returns a list of FTP files.
func (f *FTP) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.FTPExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}
