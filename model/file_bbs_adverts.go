package model

// Package file file_bbs_adverts.go contains the database queries for BBS and FTP files.

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model/expr"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// BBS is a the model for the Bulletin Board System files.
type BBS struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (b *BBS) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.BBSExpr(),
		qm.From(From)).Bind(ctx, db, b)
}

func (b *BBS) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.BBSExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// BBStro is a the model for the Bulletin Board System intro files.
type BBStro struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (b *BBStro) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.BBStroExpr(),
		qm.From(From)).Bind(ctx, db, b)
}

func (b *BBStro) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.BBStroExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// BBSImage is a the model for the Bulletin Board System image files.
type BBSImage struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (b *BBSImage) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.BBSImageExpr(),
		qm.From(From)).Bind(ctx, db, b)
}

func (b *BBSImage) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.BBSImageExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// BBSText is a the model for the Bulletin Board System text files.
type BBSText struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (b *BBSText) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.BBSTextExpr(),
		qm.From(From)).Bind(ctx, db, b)
}

func (b *BBSText) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.BBSTextExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// FTP is a the model for the FTP files.
type FTP struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (f *FTP) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.FTPExpr(),
		qm.From(From)).Bind(ctx, db, f)
}

func (f *FTP) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.FTPExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}
