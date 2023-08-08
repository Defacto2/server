package model

// Package file file_filepacks.go contains the database queries for file packages and collections.

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/model/expr"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// AnsiPack is a the model for the ANSI file packs.
type AnsiPack struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

func (a *AnsiPack) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Stat()...),
		expr.AnsiPackExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

func (a *AnsiPack) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.AnsiPackExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// ImagePack is a the model for the image file packs.
type ImagePack struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

func (i *ImagePack) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Stat()...),
		expr.ImagePackExpr(),
		qm.From(From)).Bind(ctx, db, i)
}

func (i *ImagePack) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.ImagePackExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// DosPack is a the model for the DOS file packs.
type DosPack struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

func (d *DosPack) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Stat()...),
		expr.DosPackExpr(),
		qm.From(From)).Bind(ctx, db, d)
}

func (d *DosPack) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.DosPackExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// TextPack is a the model for the text file packs.
type TextPack struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

func (t *TextPack) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Stat()...),
		expr.TextPackExpr(),
		qm.From(From)).Bind(ctx, db, t)
}

func (t *TextPack) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.TextPackExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// WindowsPack is a the model for the Windows file packs.
type WindowsPack struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

func (w *WindowsPack) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Stat()...),
		expr.WindowsPackExpr(),
		qm.From(From)).Bind(ctx, db, w)
}

func (w *WindowsPack) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.WindowsPackExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}
