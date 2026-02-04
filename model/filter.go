package model

// Package file filter.go handles the database queries filtered by the artifact category tag or platform.

import (
	"context"
	"time"

	"github.com/Defacto2/server/internal/panics"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model/querymod"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

// columns is a cached reference to postgres.Columns() to avoid repeated function calls.
var columns []string

func getColumns() []string {
	if columns == nil {
		columns = postgres.Columns()
	}
	return columns
}

// Advert is the model for the for sale.
type Advert struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (a *Advert) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.AdvertExpr(),
		qm.From(From)).Bind(ctx, exec, a)
}

func (a *Advert) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.AdvertExpr(),
		qm.Offset(calc(offset, limit)),
		qm.OrderBy(ClauseOldDate),
		qm.Limit(limit),
	).All(ctx, exec)
}

// Announcement is the model for the public and community announcements.
type Announcement struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (a *Announcement) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.AnnouncementExpr(),
		qm.From(From)).Bind(ctx, exec, a)
}

func (a *Announcement) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.AnnouncementExpr(),
		qm.Offset(calc(offset, limit)),
		qm.OrderBy(ClauseOldDate),
		qm.Limit(limit),
	).All(ctx, exec)
}

// Ansi is the model for the ANSI formatted text and art files.
type Ansi struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (a *Ansi) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.AnsiExpr(),
		qm.From(From)).Bind(ctx, exec, a)
}

func (a *Ansi) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.AnsiExpr(),
		qm.Offset(calc(offset, limit)),
		qm.OrderBy(ClauseOldDate),
		qm.Limit(limit),
	).All(ctx, exec)
}

// AnsiBrand is the model for the brand logos created in ANSI text.
type AnsiBrand struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (a *AnsiBrand) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.AnsiBrandExpr(),
		qm.From(From)).Bind(ctx, exec, a)
}

func (a *AnsiBrand) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.AnsiBrandExpr(),
		qm.Offset(calc(offset, limit)),
		qm.OrderBy(ClauseOldDate),
		qm.Limit(limit),
	).All(ctx, exec)
}

// AnsiBBS is the model for the BBS advertisements created in ANSI text.
type AnsiBBS struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (a *AnsiBBS) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.AnsiBBSExpr(),
		qm.From(From)).Bind(ctx, exec, a)
}

func (a *AnsiBBS) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.AnsiBBSExpr(),
		qm.Offset(calc(offset, limit)),
		qm.OrderBy(ClauseOldDate),
		qm.Limit(limit),
	).All(ctx, exec)
}

// AnsiFTP is the model for the FTP advertisements created in ANSI text.
type AnsiFTP struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (a *AnsiFTP) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.AnsiFTPExpr(),
		qm.From(From)).Bind(ctx, exec, a)
}

func (a *AnsiFTP) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.AnsiFTPExpr(),
		qm.Offset(calc(offset, limit)),
		qm.OrderBy(ClauseOldDate),
		qm.Limit(limit),
	).All(ctx, exec)
}

// AnsiNfo is the model for the NFO files created in ANSI text.
type AnsiNfo struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (a *AnsiNfo) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.AnsiNfoExpr(),
		qm.From(From)).Bind(ctx, exec, a)
}

func (a *AnsiNfo) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.AnsiNfoExpr(),
		qm.Offset(calc(offset, limit)),
		qm.OrderBy(ClauseOldDate),
		qm.Limit(limit),
	).All(ctx, exec)
}

// AnsiPack is the model for the ANSI file packs.
type AnsiPack struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (a *AnsiPack) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.AnsiPackExpr(),
		qm.From(From)).Bind(ctx, exec, a)
}

func (a *AnsiPack) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.AnsiPackExpr(),
		qm.Offset(calc(offset, limit)),
		qm.OrderBy(ClauseOldDate),
		qm.Limit(limit),
	).All(ctx, exec)
}

// BBS is the model for the Bulletin Board System files.
type BBS struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (b *BBS) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.BBSExpr(),
		qm.From(From)).Bind(ctx, exec, b)
}

func (b *BBS) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.BBSExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// BBStro is the model for the Bulletin Board System intro files.
type BBStro struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (b *BBStro) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.BBStroExpr(),
		qm.From(From)).Bind(ctx, exec, b)
}

func (b *BBStro) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.BBStroExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

func (b *BBStro) Sensenstahl(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	const clauseNewUpload = "id DESC"
	return models.Files(
		querymod.BBStroExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(clauseNewUpload),
	).All(ctx, exec)
}

// BBSImage is the model for the Bulletin Board System image files.
type BBSImage struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (b *BBSImage) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.BBSImageExpr(),
		qm.From(From)).Bind(ctx, exec, b)
}

func (b *BBSImage) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.BBSImageExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// BBSText is the model for the Bulletin Board System text files.
type BBSText struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (b *BBSText) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.BBSTextExpr(),
		qm.From(From)).Bind(ctx, exec, b)
}

func (b *BBSText) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.BBSTextExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// Database is the model for the database releases.
type Database struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (d *Database) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.DatabaseExpr(),
		qm.From(From)).Bind(ctx, exec, d)
}

func (d *Database) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.DatabaseExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// Demoscene is the model for the demoscene releases.
type Demoscene struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (d *Demoscene) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.DemoExpr(),
		qm.From(From)).Bind(ctx, exec, d)
}

func (d *Demoscene) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.DemoExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// Drama is the model for community drama.
type Drama struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (d *Drama) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.DramaExpr(),
		qm.From(From)).Bind(ctx, exec, d)
}

func (d *Drama) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.DramaExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// FTP is the model for the FTP files.
type FTP struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (f *FTP) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.FTPExpr(),
		qm.From(From)).Bind(ctx, exec, f)
}

func (f *FTP) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.FTPExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// Hack is the model for the game hacks.
type Hack struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (h *Hack) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.HackExpr(),
		qm.From(From)).Bind(ctx, exec, h)
}

func (h *Hack) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.HackExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// HowTo is the model for the guides and how-tos.
type HowTo struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (h *HowTo) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.HowToExpr(),
		qm.From(From)).Bind(ctx, exec, h)
}

func (h *HowTo) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.HowToExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// HTML is the model for the HTML and markdown files.
type HTML struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (h *HTML) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.HTMLExpr(),
		qm.From(From)).Bind(ctx, exec, h)
}

func (h *HTML) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.HTMLExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// Image is the model for the images.
type Image struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (i *Image) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.ImageExpr(),
		qm.From(From)).Bind(ctx, exec, i)
}

func (i *Image) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.ImageExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// ImagePack is the model for the image file packs.
type ImagePack struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (i *ImagePack) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.ImagePackExpr(),
		qm.From(From)).Bind(ctx, exec, i)
}

func (i *ImagePack) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.ImagePackExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// Intro contain statistics for releases that could be considered intros or cracktros.
type Intro struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (i *Intro) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.IntroExpr(),
		qm.From(From)).Bind(ctx, exec, i)
}

func (i *Intro) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.IntroExpr(),
		qm.OrderBy(ClauseOldDate),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
	).All(ctx, exec)
}

// IntroMsDos contain statistics for releases that could be considered DOS intros or cracktros.
type IntroMsDos struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (i *IntroMsDos) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.IntroDOSExpr(),
		qm.From(From)).Bind(ctx, exec, i)
}

func (i *IntroMsDos) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.IntroDOSExpr(),
		qm.OrderBy(ClauseOldDate),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
	).All(ctx, exec)
}

// IntroWindows contain statistics for releases that could be considered Windows intros or cracktros.
type IntroWindows struct {
	Cache   time.Time
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (i *IntroWindows) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.IntroWindowsExpr(),
		qm.From(From)).Bind(ctx, exec, i)
}

func (i *IntroWindows) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.IntroWindowsExpr(),
		qm.OrderBy(ClauseOldDate),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
	).All(ctx, exec)
}

// Installer contain statistics for releases that could be considered installers.
type Installer struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (i *Installer) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.InstallExpr(),
		qm.From(From)).Bind(ctx, exec, i)
}

func (i *Installer) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.InstallExpr(),
		qm.OrderBy(ClauseOldDate),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
	).All(ctx, exec)
}

// Java is the model for the Java operating system.
type Java struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (j *Java) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.JavaExpr(),
		qm.From(From)).Bind(ctx, exec, j)
}

func (j *Java) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.JavaExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// JobAdvert is the model for group job advertisements.
type JobAdvert struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (j *JobAdvert) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.JobAdvertExpr(),
		qm.From(From)).Bind(ctx, exec, j)
}

func (j *JobAdvert) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.JobAdvertExpr(),
		qm.Offset(calc(offset, limit)),
		qm.OrderBy(ClauseOldDate),
		qm.Limit(limit),
	).All(ctx, exec)
}

// Linux is the model for the Linux operating system.
type Linux struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (l *Linux) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.LinuxExpr(),
		qm.From(From)).Bind(ctx, exec, l)
}

func (l *Linux) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.LinuxExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// Magazine is the model for the magazine files.
type Magazine struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (m *Magazine) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.MagExpr(),
		qm.From(From)).Bind(ctx, exec, m)
}

func (m *Magazine) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.MagExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// Macos is the model for the Macintosh operating system.
type Macos struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (m *Macos) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.MacExpr(),
		qm.From(From)).Bind(ctx, exec, m)
}

func (m *Macos) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.MacExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// MsDos is the model for the MS-DOS operating system.
type MsDos struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (d *MsDos) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.DOSExpr(),
		qm.From(From)).Bind(ctx, exec, d)
}

func (d *MsDos) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.DOSExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// MsDosPack is the model for the DOS file packs.
type MsDosPack struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (d *MsDosPack) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.DosPackExpr(),
		qm.From(From)).Bind(ctx, exec, d)
}

func (d *MsDosPack) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.DosPackExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// Music is the model for the music.
type Music struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (m *Music) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.MusicExpr(),
		qm.From(From)).Bind(ctx, exec, m)
}

func (m *Music) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.MusicExpr(),
		qm.Offset(calc(offset, limit)),
		qm.OrderBy(ClauseOldDate),
		qm.Limit(limit),
	).All(ctx, exec)
}

// NewsArticle is the model for mainstream news articles.
type NewsArticle struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (n *NewsArticle) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.NewsArticleExpr(),
		qm.From(From)).Bind(ctx, exec, n)
}

func (n *NewsArticle) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.NewsArticleExpr(),
		qm.Offset(calc(offset, limit)),
		qm.OrderBy(ClauseOldDate),
		qm.Limit(limit),
	).All(ctx, exec)
}

// Nfo is the model for the NFO files.
type Nfo struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (n *Nfo) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.NfoExpr(),
		qm.From(From)).Bind(ctx, exec, n)
}

func (n *Nfo) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.NfoExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// NfoTool is the model for the NFO tools.
type NfoTool struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (n *NfoTool) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.NfoToolExpr(),
		qm.From(From)).Bind(ctx, exec, n)
}

func (n *NfoTool) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.NfoToolExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// PDF is the model for the documents in PDF format.
type PDF struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (p *PDF) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.PDFExpr(),
		qm.From(From)).Bind(ctx, exec, p)
}

func (p *PDF) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.PDFExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// Proof is the model for the file proofs.
type Proof struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (p *Proof) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.ProofExpr(),
		qm.From(From)).Bind(ctx, exec, p)
}

func (p *Proof) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.ProofExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

type Restrict struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (r *Restrict) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.RestrictExpr(),
		qm.From(From)).Bind(ctx, exec, r)
}

func (r *Restrict) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.RestrictExpr(),
		qm.Offset(calc(offset, limit)),
		qm.OrderBy(ClauseOldDate),
		qm.Limit(limit),
	).All(ctx, exec)
}

// Script is the model for the script and interpreted languages.
type Script struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (s *Script) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.ScriptExpr(),
		qm.From(From)).Bind(ctx, exec, s)
}

func (s *Script) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.ScriptExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// Standard is the model for community standards.
type Standard struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (s *Standard) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.StandardExpr(),
		qm.From(From)).Bind(ctx, exec, s)
}

func (s *Standard) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.StandardExpr(),
		qm.Offset(calc(offset, limit)),
		qm.OrderBy(ClauseOldDate),
		qm.Limit(limit),
	).All(ctx, exec)
}

// Takedown is the model for the bust and takedowns.
type Takedown struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (t *Takedown) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.TakedownExpr(),
		qm.From(From)).Bind(ctx, exec, t)
}

func (t *Takedown) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.TakedownExpr(),
		qm.Offset(calc(offset, limit)),
		qm.OrderBy(ClauseOldDate),
		qm.Limit(limit),
	).All(ctx, exec)
}

// Text is the model for the text files.
type Text struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (t *Text) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.TextExpr(),
		qm.From(From)).Bind(ctx, exec, t)
}

func (t *Text) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.TextExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// TextAmiga is the model for the text files for the Amiga operating system.
type TextAmiga struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (t *TextAmiga) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.TextAmigaExpr(),
		qm.From(From)).Bind(ctx, exec, t)
}

func (t *TextAmiga) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.TextAmigaExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// TextApple2 is the model for the text files for the Apple II operating system.
type TextApple2 struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (t *TextApple2) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.AppleIIExpr(),
		qm.From(From)).Bind(ctx, exec, t)
}

func (t *TextApple2) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.AppleIIExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// TextAtariST is the model for the text files for the Atari ST operating system.
type TextAtariST struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (t *TextAtariST) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.AtariSTExpr(),
		qm.From(From)).Bind(ctx, exec, t)
}

func (t *TextAtariST) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.AtariSTExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// TextPack is the model for the text file packs.
type TextPack struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (t *TextPack) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.TextPackExpr(),
		qm.From(From)).Bind(ctx, exec, t)
}

func (t *TextPack) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.TextPackExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// Tool is the model for the computer tools.
type Tool struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (t *Tool) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.ToolExpr(),
		qm.From(From)).Bind(ctx, exec, t)
}

func (t *Tool) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.ToolExpr(),
		qm.Offset(calc(offset, limit)),
		qm.OrderBy(ClauseOldDate),
		qm.Limit(limit),
	).All(ctx, exec)
}

// TrialCrackme is the model for group job trial crackme releases.
type TrialCrackme struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (t *TrialCrackme) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.TrialCrackmeExpr(),
		qm.From(From)).Bind(ctx, exec, t)
}

func (t *TrialCrackme) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.TrialCrackmeExpr(),
		qm.Offset(calc(offset, limit)),
		qm.OrderBy(ClauseOldDate),
		qm.Limit(limit),
	).All(ctx, exec)
}

// Video is the model for the videos.
type Video struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (v *Video) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.VideoExpr(),
		qm.From(From)).Bind(ctx, exec, v)
}

func (v *Video) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.VideoExpr(),
		qm.Offset(calc(offset, limit)),
		qm.OrderBy(ClauseOldDate),
		qm.Limit(limit),
	).All(ctx, exec)
}

// Windows is the model for the Windows operating system.
type Windows struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (w *Windows) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.WindowsExpr(),
		qm.From(From)).Bind(ctx, exec, w)
}

func (w *Windows) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		querymod.WindowsExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// WindowsPack is the model for the Windows file packs.
type WindowsPack struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (w *WindowsPack) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(getColumns()...),
		qm.Where(ClauseNoSoftDel),
		querymod.WindowsPackExpr(),
		qm.From(From)).Bind(ctx, exec, w)
}

func (w *WindowsPack) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	panics.BoilExecCrash(exec)
	return models.Files(
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		querymod.WindowsPackExpr(),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}
