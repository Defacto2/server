package model

// Package file filter.go handles the database queries filtered by the artifact category tag or platform.

import (
	"context"
	"fmt"

	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model/querymod"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

type Key string

const (
	KeyPCBoard      Key = "pcboard"
	KeyPCBoardPPE   Key = "pcboard-ppe"
	KeyPCBoardText  Key = "pcboard-text"
	KeyTextAmiga    Key = "text-amiga"
	KeyTextApple2   Key = "text-apple2"
	KeyTextAtariST  Key = "text-atari-st"
	KeyPDF          Key = "pdf"
	KeyHTML         Key = "html"
	KeyNewsArticle  Key = "news-article"
	KeyStandards    Key = "standards"
	KeyAnnouncement Key = "announcement"
	KeyJobAdvert    Key = "job-advert"
	KeyTrialCrackme Key = "trial-crackme"
	KeyHack         Key = "hack"
	KeyTool         Key = "tool"
	KeyTakedown     Key = "takedown"
	KeyDrama        Key = "drama"
	KeyAdvert       Key = "advert"
	KeyRestrict     Key = "restrict"
	KeyHowTo        Key = "how-to"
	KeyNFOTool      Key = "nfo-tool"
	KeyImage        Key = "image"
	KeyMusic        Key = "music"
	KeyVideo        Key = "video"
	KeyMSDOS        Key = "msdos"
	KeyWindows      Key = "windows"
	KeyMacOS        Key = "macos"
	KeyLinux        Key = "linux"
	KeyJava         Key = "java"
	KeyScript       Key = "script"
	KeyDatabase     Key = "database"
	KeyMSDOSPack    Key = "msdos-pack"
	KeyWindowsPack  Key = "windows-pack"
	KeyImagePack    Key = "image-pack"
	KeyTextPack     Key = "text-pack"
	KeyText         Key = "text"
	KeyMagazine     Key = "magazine"
	KeyFTP          Key = "ftp"
	KeyBBSText      Key = "bbs-text"
	KeyBBSImage     Key = "bbs-image"
	KeyBBStro       Key = "bbstro"
	KeyBBS          Key = "bbs"
	KeyANSINFO      Key = "ansi-nfo"
	KeyANSIPack     Key = "ansi-pack"
	KeyANSIFTP      Key = "ansi-ftp"
	KeyANSIBBS      Key = "ansi-bbs"
	KeyANSIBrand    Key = "ansi-brand"
	KeyANSI         Key = "ansi"
	KeyProof        Key = "proof"
	KeyNFO          Key = "nfo"
	KeyDemoscene    Key = "demoscene"
	KeyInstaller    Key = "installer"
	KeyIntro        Key = "intro"
	KeyIntroMSDOS   Key = "intro-msdos"
	KeyIntroWindows Key = "intro-windows"
	KeyConsole      Key = "console"
)

// String implements the fmt.Stringer interface for Key.
func (k Key) String() string {
	return string(k)
}

// IsValid reports whether the Key is one of the defined key constants.
func (k Key) IsValid() bool {
	_, ok := matches[k]
	return ok
}

// columns holds the package-level slice reference initialized at package startup.
//
//nolint:gochecknoglobals
var columns = [...]string{
	postgres.SumSize,
	postgres.TotalCnt,
	postgres.MinYear,
	postgres.MaxYear,
}

// SummCols returns a read-only slice reference to the Summary columns.
func SummCols() []string {
	return columns[:]
}

const (
	statfmt = "stat argument: %w"
	listfmt = "list argument: %w"
)

// Advert is the model for the for sale.
type Advert struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (q *Advert) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Advert) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.AdvertExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Advert) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *Announcement) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Announcement) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.AnnouncementExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Announcement) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *Ansi) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Ansi) String() string {
	return "ANSI art and texts"
}

func (q *Ansi) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.AnsiExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Ansi) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *AnsiBrand) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *AnsiBrand) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.AnsiBrandExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *AnsiBrand) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *AnsiBBS) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *AnsiBBS) String() string {
	return "ANSI art"
}

func (q *AnsiBBS) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.AnsiBBSExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *AnsiBBS) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *AnsiFTP) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *AnsiFTP) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.AnsiFTPExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *AnsiFTP) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *AnsiNfo) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *AnsiNfo) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.AnsiNfoExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *AnsiNfo) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *AnsiPack) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *AnsiPack) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.AnsiPackExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *AnsiPack) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *BBS) String() string {
	return "BBS adverts"
}

func (q *BBS) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *BBS) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.BBSExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *BBS) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *BBStro) String() string {
	return "BBStros"
}
func (q *BBStro) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *BBStro) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.BBStroExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *BBStro) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
	return models.Files(
		querymod.BBStroExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

func (q *BBStro) Sensenstahl(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *BBSImage) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *BBSImage) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.BBSImageExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *BBSImage) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *BBSText) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *BBSText) String() string {
	return "Text files"
}

func (q *BBSText) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.BBSTextExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *BBSText) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
	return models.Files(
		querymod.BBSTextExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// Console is the model for console releases.
type Console struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (q *Console) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Console) String() string {
	return "Console and video game files"
}

func (q *Console) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.ConsoleExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Console) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
	return models.Files(
		querymod.ConsoleExpr(),
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

func (q *Database) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Database) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.DatabaseExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Database) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (d *Demoscene) String() string {
	return "Demoscene productions"
}

func (m *Demoscene) Values() (int, int, int, int) { return m.Count, m.Bytes, m.MinYear, m.MaxYear }

func (d *Demoscene) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.DemoExpr(),
		qm.From(From),
	).Bind(ctx, exec, d)
}

func (d *Demoscene) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *Drama) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Drama) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.DramaExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Drama) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *FTP) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *FTP) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.FTPExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *FTP) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *Hack) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Hack) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.HackExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Hack) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
	return models.Files(
		querymod.HackExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// HowTo is the model for the guides and how-to texts.
type HowTo struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (q *HowTo) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *HowTo) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.HowToExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *HowTo) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *HTML) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *HTML) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.HTMLExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *HTML) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *Image) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Image) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.ImageExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Image) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *ImagePack) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *ImagePack) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.ImagePackExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *ImagePack) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *Intro) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Intro) String() string {
	return "Cracktros and intros"
}

func (q *Intro) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.IntroExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Intro) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *IntroMsDos) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *IntroMsDos) String() string {
	return "Cracktros and intros on MS Dos"
}

func (q *IntroMsDos) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.IntroDOSExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *IntroMsDos) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
	return models.Files(
		querymod.IntroDOSExpr(),
		qm.OrderBy(ClauseOldDate),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
	).All(ctx, exec)
}

// IntroWindows contain statistics for releases that could be considered Windows intros or cracktros.
type IntroWindows struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (q *IntroWindows) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *IntroWindows) String() string {
	return "Cracktros and intros on Windows"
}

func (q *IntroWindows) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.IntroWindowsExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *IntroWindows) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *Installer) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Installer) String() string {
	return "Installers"
}

func (q *Installer) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.InstallExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Installer) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *Java) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Java) String() string {
	return "Java files"
}

func (q *Java) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.JavaExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Java) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *JobAdvert) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *JobAdvert) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.JobAdvertExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *JobAdvert) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *Linux) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Linux) String() string {
	return "Linux files"
}

func (q *Linux) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.LinuxExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Linux) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *Magazine) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Magazine) String() string {
	return "Magazines"
}

func (q *Magazine) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.MagExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Magazine) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *Macos) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Macos) String() string {
	return "MacOS software"
}

func (q *Macos) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.MacExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Macos) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *MsDos) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *MsDos) String() string {
	return "MS Dos files"
}

func (q *MsDos) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.DOSExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *MsDos) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *MsDosPack) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *MsDosPack) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.DosPackExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *MsDosPack) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *Music) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Music) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.MusicExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Music) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *NewsArticle) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *NewsArticle) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.NewsArticleExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *NewsArticle) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *Nfo) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Nfo) String() string {
	return "Nfo texts"
}

func (q *Nfo) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.NfoExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Nfo) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *NfoTool) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *NfoTool) String() string {
	return "Nfo text editors"
}

func (q *NfoTool) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.NfoToolExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *NfoTool) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
	return models.Files(
		querymod.NfoToolExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// PCBoard is the model for PCBoard platform which can include text files and applications.
type PCBoard struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (q *PCBoard) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *PCBoard) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.PCBoardExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *PCBoard) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
	return models.Files(
		querymod.PCBoardExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// PCBoardPPE is the model for PCBoard applications.
type PCBoardPPE struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (q *PCBoardPPE) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *PCBoardPPE) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.PCBoardPPEExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *PCBoardPPE) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
	return models.Files(
		querymod.PCBoardPPEExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// PCBoardText is the model for PCBoard applications.
type PCBoardText struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (q *PCBoardText) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *PCBoardText) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.PCBoardTextExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *PCBoardText) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
	return models.Files(
		querymod.PCBoardTextExpr(),
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

func (q *PDF) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *PDF) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.PDFExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *PDF) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *Proof) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Proof) String() string {
	return "Proofs"
}

func (q *Proof) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.ProofExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Proof) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *Restrict) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Restrict) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.RestrictExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Restrict) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *Script) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Script) String() string {
	return "Scripting software"
}

func (q *Script) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.ScriptExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Script) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *Standard) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Standard) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.StandardExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Standard) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
	return models.Files(
		querymod.StandardExpr(),
		qm.Offset(calc(offset, limit)),
		qm.OrderBy(ClauseOldDate),
		qm.Limit(limit),
	).All(ctx, exec)
}

// Takedown is the model for the bust and take downs.
type Takedown struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (q *Takedown) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Takedown) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.TakedownExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Takedown) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *Text) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Text) String() string {
	return "Texts"
}

func (q *Text) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.TextExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Text) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *TextAmiga) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *TextAmiga) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.TextAmigaExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *TextAmiga) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
	return models.Files(
		querymod.TextAmigaExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// TextApple2 is the model for the text files about the Apple II microcomputer.
type TextApple2 struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (q *TextApple2) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *TextApple2) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.AppleIIExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *TextApple2) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
	return models.Files(
		querymod.AppleIIExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// TextAtariST is the model for the text files about the Atari ST microcomputer.
type TextAtariST struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (q *TextAtariST) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *TextAtariST) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.AtariSTExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *TextAtariST) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *TextPack) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *TextPack) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.TextPackExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *TextPack) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *Tool) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Tool) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.ToolExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Tool) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
	return models.Files(
		querymod.ToolExpr(),
		qm.Offset(calc(offset, limit)),
		qm.OrderBy(ClauseOldDate),
		qm.Limit(limit),
	).All(ctx, exec)
}

// TrialCrackme is the model for group job trial "crackme" releases.
type TrialCrackme struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (q *TrialCrackme) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *TrialCrackme) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.TrialCrackmeExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *TrialCrackme) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *Video) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Video) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.VideoExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Video) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *Windows) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *Windows) String() string {
	return "Windows files"
}

func (q *Windows) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.WindowsExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *Windows) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
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

func (q *WindowsPack) Values() (int, int, int, int) { return q.Count, q.Bytes, q.MinYear, q.MaxYear }

func (q *WindowsPack) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(statfmt, err)
	}
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		querymod.WindowsPackExpr(),
		qm.From(From),
	).Bind(ctx, exec, q)
}

func (q *WindowsPack) List(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(listfmt, err)
	}
	return models.Files(
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
		querymod.WindowsPackExpr(),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}
