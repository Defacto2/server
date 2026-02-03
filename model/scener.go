package model

// Package scener.go contains the database queries for the sceners.

import (
	"context"
	"strings"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/panics"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

// Scener is a collective, group or individual, that releases files.
type Scener string

// Sceners is a collection of sceners.
type Sceners []*struct {
	Name Scener `boil:"scener"`
}

// Where gets the records of all files that have been credited to the named scener.
func (s *Scener) Where(ctx context.Context, exec boil.ContextExecutor, name string) (models.FileSlice, error) {
	panics.BoilExecCrash(exec)
	query, params := postgres.ScenerSQL(name)
	return models.Files(
		qm.Where(query, params...),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// Distinct gets a list of all, distinct sceners.
func (s *Sceners) Distinct(ctx context.Context, exec boil.ContextExecutor) error {
	if len(*s) > 0 {
		return nil
	}
	panics.BoilExecCrash(exec)
	query := string(postgres.Sceners())
	return queries.Raw(query).Bind(ctx, exec, s)
}

// Writer gets a list of sceners who have been credited for text.
func (s *Sceners) Writer(ctx context.Context, exec boil.ContextExecutor) error {
	if len(*s) > 0 {
		return nil
	}
	panics.BoilExecCrash(exec)
	query := string(postgres.Writers())
	return queries.Raw(query).Bind(ctx, exec, s)
}

// Artist gets a list of sceners who have been credited for graphics or art.
func (s *Sceners) Artist(ctx context.Context, exec boil.ContextExecutor) error {
	if len(*s) > 0 {
		return nil
	}
	panics.BoilExecCrash(exec)
	query := string(postgres.Artists())
	return queries.Raw(query).Bind(ctx, exec, s)
}

// Coder gets a list of sceners who have been credited for programming.
func (s *Sceners) Coder(ctx context.Context, exec boil.ContextExecutor) error {
	if len(*s) > 0 {
		return nil
	}
	panics.BoilExecCrash(exec)
	query := string(postgres.Coders())
	return queries.Raw(query).Bind(ctx, exec, s)
}

// Musician gets a list of sceners who have been credited for music or audio.
func (s *Sceners) Musician(ctx context.Context, exec boil.ContextExecutor) error {
	if len(*s) > 0 {
		return nil
	}
	panics.BoilExecCrash(exec)
	query := string(postgres.Musicians())
	return queries.Raw(query).Bind(ctx, exec, s)
}

// Sort gets a sorted slice of unique sceners.
func (s *Sceners) Sort() []string {
	sceners := make([]string, 0, len(*s))
	for _, scener := range *s {
		sceners = append(sceners, strings.Split(string(scener.Name), ",")...)
	}
	return helper.DeleteDupe(sceners...)
}
