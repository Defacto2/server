package model

// Package file text.go contains the database queries for text, markdown and document files.

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/model/modext"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type HTML struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of HTML files.
func (h *HTML) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.HTMLExpr(),
		qm.From(From)).Bind(ctx, db, h)
}

// List returns a list of HTML files.
func (h *HTML) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.HTMLExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

type Text struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
	Year0 int `boil:"min_year"`
	YearX int `boil:"max_year"`
}

// Stat counts the total number and total byte size of text files.
func (t *Text) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter, postgres.MinYear, postgres.MaxYear),
		modext.TextExpr(),
		qm.From(From)).Bind(ctx, db, t)
}

// List returns a list of text files.
func (t *Text) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.TextExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

type TextAmiga struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of text files for the Amiga operating system.
func (t *TextAmiga) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.TextAmigaExpr(),
		qm.From(From)).Bind(ctx, db, t)
}

// List returns a list of text files for the Amiga operating system.
func (t *TextAmiga) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.TextAmigaExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

type TextAppleII struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of text files for the Apple II operating system.
func (t *TextAppleII) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.AppleIIExpr(),
		qm.From(From)).Bind(ctx, db, t)
}

// List returns a list of text files for the Apple II operating system.
func (t *TextAppleII) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.AppleIIExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

type TextAtariST struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of text files for the Atari ST operating system.
func (t *TextAtariST) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.AtariSTExpr(),
		qm.From(From)).Bind(ctx, db, t)
}

// List returns a list of text files for the Atari ST operating system.
func (t *TextAtariST) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.AtariSTExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

type PDF struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of PDF files.
func (p *PDF) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.PDFExpr(),
		qm.From(From)).Bind(ctx, db, p)
}

// List returns a list of PDF files.
func (p *PDF) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.PDFExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}
