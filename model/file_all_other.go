package model

// Package file_all_other.go contains the database queries the all other categories.

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/model/modext"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Advert is a the model for the for sale.
type Advert struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Announcement is a the model for the public and community announcements.
type Announcement struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Drama is the model for community drama.
type Drama struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Hack is a the model for the game hacks.
type Hack struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// HowTo is a the model for the guides and how-tos.
type HowTo struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Image is a the model for the images.
type Image struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// JobAdvert is a the model for group job advertisements.
type JobAdvert struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Music is a the model for the music.
type Music struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// NewsArticle is a the model for mainstream news articles.
type NewsArticle struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

type Restrict struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Standard is a the model for community standards.
type Standard struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Takedown is a the model for the bust and takedowns.
type Takedown struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Tool is a the model for the computer tools.
type Tool struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// TrialCrackme is a the model for group job trial crackme releases.
type TrialCrackme struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Video is a the model for the videos.
type Video struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

func (a *Advert) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.AdvertExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

func (a *Advert) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.AdvertExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

func (a *Announcement) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.AnnouncementExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

func (a *Announcement) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.AnnouncementExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

func (d *Drama) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.DramaExpr(),
		qm.From(From)).Bind(ctx, db, d)
}

func (d *Drama) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.DramaExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

func (h *Hack) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.HackExpr(),
		qm.From(From)).Bind(ctx, db, h)
}

func (h *Hack) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.HackExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

func (h *HowTo) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.HowToExpr(),
		qm.From(From)).Bind(ctx, db, h)
}

func (h *HowTo) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.HowToExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

func (i *Image) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.ImageExpr(),
		qm.From(From)).Bind(ctx, db, i)
}

func (i *Image) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.ImageExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

func (j *JobAdvert) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.JobAdvertExpr(),
		qm.From(From)).Bind(ctx, db, j)
}

func (j *JobAdvert) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.JobAdvertExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

func (m *Music) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.MusicExpr(),
		qm.From(From)).Bind(ctx, db, m)
}

func (m *Music) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.MusicExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

func (n *NewsArticle) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.NewsArticleExpr(),
		qm.From(From)).Bind(ctx, db, n)
}

func (n *NewsArticle) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.NewsArticleExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

func (r *Restrict) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.RestrictExpr(),
		qm.From(From)).Bind(ctx, db, r)
}

func (r *Restrict) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.RestrictExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

func (s *Standard) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.StandardExpr(),
		qm.From(From)).Bind(ctx, db, s)
}

func (s *Standard) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.StandardExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

func (t *Takedown) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.TakedownExpr(),
		qm.From(From)).Bind(ctx, db, t)
}

func (t *Takedown) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.TakedownExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

func (t *Tool) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.ToolExpr(),
		qm.From(From)).Bind(ctx, db, t)
}

func (t *Tool) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.ToolExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

func (t *TrialCrackme) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.TrialCrackmeExpr(),
		qm.From(From)).Bind(ctx, db, t)
}

func (t *TrialCrackme) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.TrialCrackmeExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

func (v *Video) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.VideoExpr(),
		qm.From(From)).Bind(ctx, db, v)
}

func (v *Video) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.VideoExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}
