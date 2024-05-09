package model

// Package file filter.go handles the database queries filtered by the artifact category tag or platform.

import (
	"context"
	"database/sql"
	"time"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model/expr"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Advert is a the model for the for sale.
type Advert struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (a *Advert) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(ClauseNoSoftDel),
		expr.AdvertExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

func (a *Advert) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.AdvertExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// Announcement is a the model for the public and community announcements.
type Announcement struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (a *Announcement) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(ClauseNoSoftDel),
		expr.AnnouncementExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

func (a *Announcement) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.AnnouncementExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

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
		qm.OrderBy(ClauseOldDate),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
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
		qm.Where(ClauseNoSoftDel),
		expr.AnsiBrandExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

func (a *AnsiBrand) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.AnsiBrandExpr(),
		qm.OrderBy(ClauseOldDate),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
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
		qm.Where(ClauseNoSoftDel),
		expr.AnsiBBSExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

func (a *AnsiBBS) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.AnsiBBSExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.AnsiFTPExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

func (a *AnsiFTP) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.AnsiFTPExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.AnsiNfoExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

func (a *AnsiNfo) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.AnsiNfoExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.AnsiPackExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

func (a *AnsiPack) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.AnsiPackExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.BBSExpr(),
		qm.From(From)).Bind(ctx, db, b)
}

func (b *BBS) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.BBSExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.BBStroExpr(),
		qm.From(From)).Bind(ctx, db, b)
}

func (b *BBStro) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.BBStroExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.BBSImageExpr(),
		qm.From(From)).Bind(ctx, db, b)
}

func (b *BBSImage) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.BBSImageExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.BBSTextExpr(),
		qm.From(From)).Bind(ctx, db, b)
}

func (b *BBSText) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.BBSTextExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.DatabaseExpr(),
		qm.From(From)).Bind(ctx, db, d)
}

func (d *Database) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.DatabaseExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.DemoExpr(),
		qm.From(From)).Bind(ctx, db, d)
}

func (d *Demoscene) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.DemoExpr(),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, db)
}

// Drama is the model for community drama.
type Drama struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (d *Drama) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(ClauseNoSoftDel),
		expr.DramaExpr(),
		qm.From(From)).Bind(ctx, db, d)
}

func (d *Drama) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.DramaExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
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
		qm.Where(ClauseNoSoftDel),
		expr.FTPExpr(),
		qm.From(From)).Bind(ctx, db, f)
}

func (f *FTP) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.FTPExpr(),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, db)
}

// Hack is a the model for the game hacks.
type Hack struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (h *Hack) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(ClauseNoSoftDel),
		expr.HackExpr(),
		qm.From(From)).Bind(ctx, db, h)
}

func (h *Hack) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.HackExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// HowTo is a the model for the guides and how-tos.
type HowTo struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (h *HowTo) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(ClauseNoSoftDel),
		expr.HowToExpr(),
		qm.From(From)).Bind(ctx, db, h)
}

func (h *HowTo) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.HowToExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
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
		qm.Where(ClauseNoSoftDel),
		expr.HTMLExpr(),
		qm.From(From)).Bind(ctx, db, h)
}

func (h *HTML) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.HTMLExpr(),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, db)
}

// Image is a the model for the images.
type Image struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (i *Image) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(ClauseNoSoftDel),
		expr.ImageExpr(),
		qm.From(From)).Bind(ctx, db, i)
}

func (i *Image) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.ImageExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
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
		qm.Where(ClauseNoSoftDel),
		expr.ImagePackExpr(),
		qm.From(From)).Bind(ctx, db, i)
}

func (i *ImagePack) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.ImagePackExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.IntroExpr(),
		qm.From(From)).Bind(ctx, db, i)
}

func (i *Intro) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.IntroExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.IntroDOSExpr(),
		qm.From(From)).Bind(ctx, db, i)
}

func (i *IntroMsDos) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.IntroDOSExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.IntroWindowsExpr(),
		qm.From(From)).Bind(ctx, db, i)
}

func (i *IntroWindows) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.IntroWindowsExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.InstallExpr(),
		qm.From(From)).Bind(ctx, db, i)
}

func (i *Installer) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.InstallExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.JavaExpr(),
		qm.From(From)).Bind(ctx, db, j)
}

func (j *Java) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.JavaExpr(),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, db)
}

// JobAdvert is a the model for group job advertisements.
type JobAdvert struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (j *JobAdvert) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(ClauseNoSoftDel),
		expr.JobAdvertExpr(),
		qm.From(From)).Bind(ctx, db, j)
}

func (j *JobAdvert) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.JobAdvertExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
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
		qm.Where(ClauseNoSoftDel),
		expr.LinuxExpr(),
		qm.From(From)).Bind(ctx, db, l)
}

func (l *Linux) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.LinuxExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.MagExpr(),
		qm.From(From)).Bind(ctx, db, m)
}

func (m *Magazine) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.MagExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.MacExpr(),
		qm.From(From)).Bind(ctx, db, m)
}

func (m *Macos) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.MacExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.DOSExpr(),
		qm.From(From)).Bind(ctx, db, d)
}

func (d *MsDos) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.DOSExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.DosPackExpr(),
		qm.From(From)).Bind(ctx, db, d)
}

func (d *MsDosPack) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.DosPackExpr(),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, db)
}

// Music is a the model for the music.
type Music struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (m *Music) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(ClauseNoSoftDel),
		expr.MusicExpr(),
		qm.From(From)).Bind(ctx, db, m)
}

func (m *Music) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.MusicExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// NewsArticle is a the model for mainstream news articles.
type NewsArticle struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (n *NewsArticle) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(ClauseNoSoftDel),
		expr.NewsArticleExpr(),
		qm.From(From)).Bind(ctx, db, n)
}

func (n *NewsArticle) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.NewsArticleExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
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
		qm.Where(ClauseNoSoftDel),
		expr.NfoExpr(),
		qm.From(From)).Bind(ctx, db, n)
}

func (n *Nfo) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.NfoExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.NfoToolExpr(),
		qm.From(From)).Bind(ctx, db, n)
}

func (n *NfoTool) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.NfoToolExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.PDFExpr(),
		qm.From(From)).Bind(ctx, db, p)
}

func (p *PDF) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.PDFExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.ProofExpr(),
		qm.From(From)).Bind(ctx, db, p)
}

func (p *Proof) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.ProofExpr(),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, db)
}

type Restrict struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (r *Restrict) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(ClauseNoSoftDel),
		expr.RestrictExpr(),
		qm.From(From)).Bind(ctx, db, r)
}

func (r *Restrict) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.RestrictExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
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
		qm.Where(ClauseNoSoftDel),
		expr.ScriptExpr(),
		qm.From(From)).Bind(ctx, db, s)
}

func (s *Script) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.ScriptExpr(),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, db)
}

// Standard is a the model for community standards.
type Standard struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (s *Standard) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(ClauseNoSoftDel),
		expr.StandardExpr(),
		qm.From(From)).Bind(ctx, db, s)
}

func (s *Standard) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.StandardExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// Takedown is a the model for the bust and takedowns.
type Takedown struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (t *Takedown) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(ClauseNoSoftDel),
		expr.TakedownExpr(),
		qm.From(From)).Bind(ctx, db, t)
}

func (t *Takedown) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(expr.TakedownExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
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
		qm.Where(ClauseNoSoftDel),
		expr.TextExpr(),
		qm.From(From)).Bind(ctx, db, t)
}

func (t *Text) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.TextExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.TextAmigaExpr(),
		qm.From(From)).Bind(ctx, db, t)
}

func (t *TextAmiga) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.TextAmigaExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.AppleIIExpr(),
		qm.From(From)).Bind(ctx, db, t)
}

func (t *TextApple2) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.AppleIIExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.AtariSTExpr(),
		qm.From(From)).Bind(ctx, db, t)
}

func (t *TextAtariST) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.AtariSTExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.TextPackExpr(),
		qm.From(From)).Bind(ctx, db, t)
}

func (t *TextPack) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.TextPackExpr(),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, db)
}

// Tool is a the model for the computer tools.
type Tool struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (t *Tool) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(ClauseNoSoftDel),
		expr.ToolExpr(),
		qm.From(From)).Bind(ctx, db, t)
}

func (t *Tool) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.ToolExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// TrialCrackme is a the model for group job trial crackme releases.
type TrialCrackme struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (t *TrialCrackme) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(ClauseNoSoftDel),
		expr.TrialCrackmeExpr(),
		qm.From(From)).Bind(ctx, db, t)
}

func (t *TrialCrackme) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.TrialCrackmeExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// Video is a the model for the videos.
type Video struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (v *Video) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(ClauseNoSoftDel),
		expr.VideoExpr(),
		qm.From(From)).Bind(ctx, db, v)
}

func (v *Video) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.VideoExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
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
		qm.Where(ClauseNoSoftDel),
		expr.WindowsExpr(),
		qm.From(From)).Bind(ctx, db, w)
}

func (w *Windows) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.WindowsExpr(),
		qm.OrderBy(ClauseOldDate),
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
		qm.Where(ClauseNoSoftDel),
		expr.WindowsPackExpr(),
		qm.From(From)).Bind(ctx, db, w)
}

func (w *WindowsPack) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.WindowsPackExpr(),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, db)
}
