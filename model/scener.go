package model

// Package scener.go contains the database queries for the sceners.

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/Defacto2/server/internal/nils"
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
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf("scener where: %w", err)
	}

	t := strings.TrimSpace(name)
	if t == "" {
		return models.FileSlice{}, nil
	}
	clause, args := postgres.ScenerSQL(t)

	return models.Files(qm.Where(clause, args...), qm.OrderBy(ClauseOldDate)).All(ctx, exec)
}

// Distinct gets a list of all, distinct sceners.
func (obj *Sceners) Distinct(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf("sceners distinct: %w", err)
	}
	if len(*obj) > 0 {
		return nil
	}

	query := string(postgres.Sceners())

	return queries.Raw(query).Bind(ctx, exec, obj)
}

// Writer gets a list of sceners who have been credited for text.
func (obj *Sceners) Writer(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf("sceners writers: %w", err)
	}
	if len(*obj) > 0 {
		return nil
	}
	query := string(postgres.Writers())
	return queries.Raw(query).Bind(ctx, exec, obj)
}

// Artist gets a list of sceners who have been credited for graphics or art.
func (obj *Sceners) Artist(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf("sceners artists: %w", err)
	}
	if len(*obj) > 0 {
		return nil
	}
	query := string(postgres.Artists())
	return queries.Raw(query).Bind(ctx, exec, obj)
}

// Coder gets a list of sceners who have been credited for programming.
func (obj *Sceners) Coder(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf("sceners coder: %w", err)
	}
	if len(*obj) > 0 {
		return nil
	}
	query := string(postgres.Coders())
	return queries.Raw(query).Bind(ctx, exec, obj)
}

// Musician gets a list of sceners who have been credited for music or audio.
func (obj *Sceners) Musician(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf("sceners musician: %w", err)
	}
	if len(*obj) > 0 {
		return nil
	}
	query := string(postgres.Musicians())
	return queries.Raw(query).Bind(ctx, exec, obj)
}

// Sort gets a sorted slice of unique sceners.
func (obj *Sceners) Sort() []string {
	if obj == nil || len(*obj) == 0 {
		return []string{}
	}

	sceners := make([]string, 0, len(*obj))
	for _, scener := range *obj {
		for name := range strings.SplitSeq(string(scener.Name), ",") {
			if t := strings.TrimSpace(name); t != "" {
				sceners = append(sceners, t)
			}
		}
	}

	slices.Sort(sceners)
	return slices.Compact(sceners)
}
