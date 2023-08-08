package model

// Package file_text_files.go contains the database queries for text, markdown and document files.

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/model/expr"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// HTML is a the model for the HTML and markdown files.
type HTML struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

func (h *HTML) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Stat()...),
		expr.HTMLExpr(),
		qm.From(From)).Bind(ctx, db, h)
}

func (h *HTML) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.HTMLExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// Text is a the model for the text files.
type Text struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
	Year0 int `boil:"min_year"`
	YearX int `boil:"max_year"`
}

func (t *Text) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.TextExpr(),
		qm.From(From)).Bind(ctx, db, t)
}

func (t *Text) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.TextExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// TextAmiga is a the model for the text files for the Amiga operating system.
type TextAmiga struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

func (t *TextAmiga) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Stat()...),
		expr.TextAmigaExpr(),
		qm.From(From)).Bind(ctx, db, t)
}

func (t *TextAmiga) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.TextAmigaExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// TextAppleII is a the model for the text files for the Apple II operating system.
type TextAppleII struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

func (t *TextAppleII) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Stat()...),
		expr.AppleIIExpr(),
		qm.From(From)).Bind(ctx, db, t)
}

func (t *TextAppleII) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.AppleIIExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// TextAtariST is a the model for the text files for the Atari ST operating system.
type TextAtariST struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

func (t *TextAtariST) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Stat()...),
		expr.AtariSTExpr(),
		qm.From(From)).Bind(ctx, db, t)
}

func (t *TextAtariST) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.AtariSTExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// PDF is a the model for the documents in PDF format.
type PDF struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

func (p *PDF) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Stat()...),
		expr.PDFExpr(),
		qm.From(From)).Bind(ctx, db, p)
}

func (p *PDF) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.PDFExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}
