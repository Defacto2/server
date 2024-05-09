package model

// Package file filterall.go contains sqlboiler models for the intros, installers and demoscene releases.

import (
	"context"
	"database/sql"
	"time"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model/expr"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Ansi is a the model for the ANSI formatted text and art files.
type Ansi struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (a *Ansi) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.AnsiExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

func (a *Ansi) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.AnsiExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// AnsiBrand is a the model for the brand logos created in ANSI text.
type AnsiBrand struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (a *AnsiBrand) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.AnsiBrandExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

func (a *AnsiBrand) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.AnsiBrandExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// AnsiBBS is a the model for the BBS advertisements created in ANSI text.
type AnsiBBS struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (a *AnsiBBS) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.AnsiBBSExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

func (a *AnsiBBS) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.AnsiBBSExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// AnsiFTP is a the model for the FTP advertisements created in ANSI text.
type AnsiFTP struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (a *AnsiFTP) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.AnsiFTPExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

func (a *AnsiFTP) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.AnsiFTPExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// AnsiNfo is a the model for the NFO files created in ANSI text.
type AnsiNfo struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (a *AnsiNfo) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.AnsiNfoExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

func (a *AnsiNfo) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.AnsiNfoExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// AnsiPack is a the model for the ANSI file packs.
type AnsiPack struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (a *AnsiPack) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.AnsiPackExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

func (a *AnsiPack) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.AnsiPackExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

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

// Database is a the model for the database releases.
type Database struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (d *Database) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.DatabaseExpr(),
		qm.From(From)).Bind(ctx, db, d)
}

func (d *Database) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.DatabaseExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// Demoscene is a the model for the demoscene releases.
type Demoscene struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (d *Demoscene) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.DemoExpr(),
		qm.From(From)).Bind(ctx, db, d)
}

func (d *Demoscene) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.DemoExpr(),
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

// HTML is a the model for the HTML and markdown files.
type HTML struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (h *HTML) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.HTMLExpr(),
		qm.From(From)).Bind(ctx, db, h)
}

func (h *HTML) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.HTMLExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// ImagePack is a the model for the image file packs.
type ImagePack struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (i *ImagePack) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.ImagePackExpr(),
		qm.From(From)).Bind(ctx, db, i)
}

func (i *ImagePack) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.ImagePackExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// Intro contain statistics for releases that could be considered intros or cracktros.
type Intro struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (i *Intro) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.IntroExpr(),
		qm.From(From)).Bind(ctx, db, i)
}

func (i *Intro) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.IntroExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// IntroMsDos contain statistics for releases that could be considered DOS intros or cracktros.
type IntroMsDos struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (i *IntroMsDos) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.IntroDOSExpr(),
		qm.From(From)).Bind(ctx, db, i)
}

func (i *IntroMsDos) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.IntroDOSExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// IntroWindows contain statistics for releases that could be considered Windows intros or cracktros.
type IntroWindows struct {
	Cache   time.Time
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (i *IntroWindows) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.IntroWindowsExpr(),
		qm.From(From)).Bind(ctx, db, i)
}

func (i *IntroWindows) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.IntroWindowsExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// Installer contain statistics for releases that could be considered installers.
type Installer struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (i *Installer) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.InstallExpr(),
		qm.From(From)).Bind(ctx, db, i)
}

func (i *Installer) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.InstallExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// Java is a the model for the Java operating system.
type Java struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (j *Java) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.JavaExpr(),
		qm.From(From)).Bind(ctx, db, j)
}

func (j *Java) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.JavaExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// Linux is a the model for the Linux operating system.
type Linux struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (l *Linux) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.LinuxExpr(),
		qm.From(From)).Bind(ctx, db, l)
}

func (l *Linux) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.LinuxExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// Magazine is a the model for the magazine files.
type Magazine struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
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

// Macos is a the model for the Macintosh operating system.
type Macos struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (m *Macos) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.MacExpr(),
		qm.From(From)).Bind(ctx, db, m)
}

func (m *Macos) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.MacExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// MsDos is a the model for the MS-DOS operating system.
type MsDos struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (d *MsDos) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.DOSExpr(),
		qm.From(From)).Bind(ctx, db, d)
}

func (d *MsDos) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.DOSExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// MsDosPack is a the model for the DOS file packs.
type MsDosPack struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (d *MsDosPack) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.DosPackExpr(),
		qm.From(From)).Bind(ctx, db, d)
}

func (d *MsDosPack) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.DosPackExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

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

// PDF is a the model for the documents in PDF format.
type PDF struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (p *PDF) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.PDFExpr(),
		qm.From(From)).Bind(ctx, db, p)
}

func (p *PDF) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.PDFExpr(),
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

// Script is a the model for the script and interpreted languages.
type Script struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (s *Script) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.ScriptExpr(),
		qm.From(From)).Bind(ctx, db, s)
}

func (s *Script) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.ScriptExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// Text is a the model for the text files.
type Text struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
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
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// TextAmiga is a the model for the text files for the Amiga operating system.
type TextAmiga struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (t *TextAmiga) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.TextAmigaExpr(),
		qm.From(From)).Bind(ctx, db, t)
}

func (t *TextAmiga) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.TextAmigaExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// TextApple2 is a the model for the text files for the Apple II operating system.
type TextApple2 struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (t *TextApple2) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.AppleIIExpr(),
		qm.From(From)).Bind(ctx, db, t)
}

func (t *TextApple2) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.AppleIIExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// TextAtariST is a the model for the text files for the Atari ST operating system.
type TextAtariST struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (t *TextAtariST) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.AtariSTExpr(),
		qm.From(From)).Bind(ctx, db, t)
}

func (t *TextAtariST) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.AtariSTExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// TextPack is a the model for the text file packs.
type TextPack struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (t *TextPack) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.TextPackExpr(),
		qm.From(From)).Bind(ctx, db, t)
}

func (t *TextPack) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.TextPackExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// Windows is a the model for the Windows operating system.
type Windows struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (w *Windows) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.WindowsExpr(),
		qm.From(From)).Bind(ctx, db, w)
}

func (w *Windows) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.WindowsExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}

// WindowsPack is a the model for the Windows file packs.
type WindowsPack struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (w *WindowsPack) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.WindowsPackExpr(),
		qm.From(From)).Bind(ctx, db, w)
}

func (w *WindowsPack) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.WindowsPackExpr(),
		qm.OrderBy(ClauseOldDate), qm.Offset(calc(offset, limit)), qm.Limit(limit),
	).All(ctx, db)
}
