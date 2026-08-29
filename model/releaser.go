package model

// Package releaser.go contains the database queries the releasers and groups.

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/releaser"
	"github.com/Defacto2/server/handler/releaser/name"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

const PageSet = 10 // PageSet is the default number of results when limit/page arguments are offered.

// Set is a zero-byte memory map used to store and reference string values.
type Set map[string]struct{}

// Add the key to the Set map.
func (s Set) Add(key string) {
	s[key] = struct{}{}
}

// Has returns true when the key is found in the Set map.
func (s Set) Has(key string) bool {
	_, ok := s[key]
	return ok
}

// ReleaserNames is a distinct data list of releasers.
type ReleaserNames []ReleaserName

// Slugs returns a map of distinct releaser URLs for lookups
// that use little memory and are fast.
//
// Usage:
//
//	var names model.ReleaserNames
//	_ = names.DistinctGroups(ctx, db)
//	slugs := names.Slugs()
//	if slugs.Has("someone") {}
func (r *ReleaserNames) Slugs() Set {
	rs := make(Set, len(*r))
	for _, v := range *r {
		rs.Add(helper.Slug(v.String()))
	}
	return rs
}

// ReleaserName is a releaser name.
type ReleaserName struct {
	Name string `boil:"releaser"`
}

func (r *ReleaserName) String() string {
	return r.Name
}

// Distinct gets the unique releaser names.
func (r *ReleaserNames) Distinct(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf("distinct releaser names: %w", err)
	}
	query := string(postgres.Releasers())
	return queries.Raw(query).Bind(ctx, exec, r)
}

// DistinctGroups gets the unique releaser names that are groups.
func (r *ReleaserNames) DistinctGroups(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf("distinct group releaser names: %w", err)
	}
	query := string(postgres.ReleasersAlphabetical())
	return queries.Raw(query).Bind(ctx, exec, r)
}

// DistinctMagazines gets the unique releaser names that are magazines.
func (r *ReleaserNames) DistinctMagazines(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf("distinct magazine releaser names: %w", err)
	}
	query := string(postgres.MagazinesAlphabetical())
	return queries.Raw(query).Bind(ctx, exec, r)
}

// DistinctBBS gets the unique releaser names that are BBS sites.
func (r *ReleaserNames) DistinctBBS(ctx context.Context, exec boil.ContextExecutor) error {
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf("distinct bbs releaser names: %w", err)
	}
	query := string(postgres.BBSsAlphabetical())
	return queries.Raw(query).Bind(ctx, exec, r)
}

// Releasers is a collection of releasers.
type Releasers []*struct {
	Unique Releaser `boil:",bind"` // Unique releaser.
}

func (obj *Releasers) String() string {
	names := make([]string, 0, len(*obj))
	for _, name := range *obj {
		names = append(names, name.Unique.Name)
	}
	return strings.Join(names, ", ")
}

// Releaser is a collective, group or individual, that releases files.
type Releaser struct {
	Name  string `boil:"releaser"`   // Name of the releaser.
	URI   string ``                  // URI slug for the releaser, with no boiler bind.
	Bytes int    `boil:"size_total"` // Bytes are the total size of all the files under this releaser.
	Count int    `boil:"count_sum"`  // Count is the total number of files under this releaser.
	// Year is used for optional sorting and is the earliest year the releaser was active.
	Year null.Int `boil:"min_year"`
}

// Where gets the records that match the named releaser.
// If the provided name is invalid, no results but no errors are returned.
func ReleasersWhere(ctx context.Context, exec boil.ContextExecutor, releasers string) (models.FileSlice, error) {
	const format = "releasers where: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(format, err)
	}

	s, _ := name.Humanize(name.Path(releasers))
	if s == "" {
		return models.FileSlice{}, nil
	}

	val := strings.ToUpper(s)
	arg := null.StringFrom(val)
	return models.Files(
		qm.Where("upper(group_brand_for) = ? OR upper(group_brand_by) = ?", arg, arg),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// OrderBy is the sorting order for ALL the releasers.
type OrderBy uint

const (
	Prolific     OrderBy = iota // Prolific orders by the total artifact count.
	Alphabetical                // Alphabetical orders by the releaser name.
	Oldest                      // Oldest orders by the year of the first artifact.
)

// BBS gets the unique BBS site names and their total file count and file sizes.
func (order OrderBy) BBS(ctx context.Context, exec boil.ContextExecutor, obj *Releasers) error {
	const format = "unique bbs names: %w"
	if err := nils.Check(ctx, exec, obj); err != nil {
		return fmt.Errorf(format, err)
	}
	if len(*obj) > 0 {
		return nil
	}

	var query string
	switch order {
	case Prolific:
		query = string(postgres.BBSsProlific())
	case Alphabetical:
		query = string(postgres.BBSsAlphabetical())
	case Oldest:
		query = string(postgres.BBSsOldest())
	default:
		return fmt.Errorf(format, ErrMethod)
	}

	if err := queries.Raw(query).Bind(ctx, exec, obj); err != nil {
		return fmt.Errorf(format, err)
	}

	obj.Slugs()
	return nil
}

// Limit gets the unique releaser names and their total file count and file sizes.
// When reorder is true the results are ordered by the total file counts.
func (order OrderBy) Limit(ctx context.Context, exec boil.ContextExecutor, obj *Releasers, limit, page int) error {
	const format = "releasers limit: %w"
	if err := nils.Check(ctx, exec, obj); err != nil {
		return fmt.Errorf(format, err)
	}
	if len(*obj) > 0 {
		return nil
	}

	var query string
	switch order {
	case Prolific:
		query = string(postgres.ReleasersProlific())
	case Alphabetical:
		query = string(postgres.ReleasersAlphabetical())
	case Oldest:
		query = string(postgres.ReleasersOldest())
	default:
		return fmt.Errorf(format, ErrMethod)
	}
	// strconv.Itoa
	if limit > 0 {
		if page < 1 {
			page = 1
		}
		limit, offset := limits(page, limit)
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	}
	if err := queries.Raw(query).Bind(ctx, exec, obj); err != nil {
		return fmt.Errorf(format, err)
	}

	obj.Slugs()
	return nil
}

// Similar finds the unique releaser names that are similar to the named strings.
// The results are ordered by the total file counts.
// The required limit is the maximum number of results to return or defaults to 10.
func (obj *Releasers) Similar(ctx context.Context, exec boil.ContextExecutor, limit int, names ...string) error {
	return Only.releasers(ctx, exec, obj, limit, names...)
}

// Initialism finds the unique releaser names that match the named strings.
// The results are ordered by the total file counts.
// The required limit is the maximum number of results to return or defaults to 10.
func (obj *Releasers) Initialism(ctx context.Context, exec boil.ContextExecutor, limit int, names ...string) error {
	return OnlyExact.releasers(ctx, exec, obj, limit, names...)
}

// SimilarMagazine finds the unique releaser names that are similar to the named strings.
// The results are ordered by the total file counts.
// The required limit is the maximum number of results to return or defaults to 10.
func (obj *Releasers) SimilarMagazine(ctx context.Context, exec boil.ContextExecutor, limit int, names ...string) error {
	return OnlyMagazine.releasers(ctx, exec, obj, limit, names...)
}

func limits(pageNumber, pageSize int) (limit, offset int) {
	if pageNumber < 1 {
		pageNumber = 1
	}
	if pageSize < 1 {
		pageSize = PageSet
	}

	limit = pageSize
	offset = (pageNumber - 1) * pageSize
	return limit, offset
}

// FTP gets the unique FTP site names and their total file count and file sizes.
func (obj *Releasers) FTP(ctx context.Context, exec boil.ContextExecutor) error {
	const format = "unique ftp names: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, err)
	}
	if len(*obj) > 0 {
		return nil
	}

	if err := queries.Raw(string(postgres.FTPsAlphabetical())).Bind(ctx, exec, obj); err != nil {
		return fmt.Errorf(format, err)
	}

	obj.Slugs()
	return nil
}

// MagazineAZ gets the unique magazine titles and their total issue count and file sizes.
func (obj *Releasers) MagazineAZ(ctx context.Context, exec boil.ContextExecutor) error {
	const format = "unique magazine names a-z: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, err)
	}
	if len(*obj) > 0 {
		return nil
	}

	if err := queries.Raw(string(postgres.MagazinesAlphabetical())).Bind(ctx, exec, obj); err != nil {
		return fmt.Errorf(format, err)
	}

	obj.Slugs()
	return nil
}

// Magazine gets the unique magazine titles and their total issue count and file sizes.
func (obj *Releasers) Magazine(ctx context.Context, exec boil.ContextExecutor) error {
	const format = "unique magazine names: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, err)
	}
	if len(*obj) > 0 {
		return nil
	}

	if err := queries.Raw(string(postgres.MagazinesOldest())).Bind(ctx, exec, obj); err != nil {
		return fmt.Errorf(format, err)
	}
	obj.Slugs()

	return nil
}

// Slugs sets URL friendly strings to the Group names.
func (obj *Releasers) Slugs() {
	for _, releaser := range *obj {
		releaser.Unique.URI = helper.Slug(releaser.Unique.Name)
	}
}

type Lookup int

const (
	Only Lookup = iota
	OnlyExact
	OnlyMagazine
)

func (match Lookup) releasers(
	ctx context.Context, exec boil.ContextExecutor, obj *Releasers, limit int, names ...string,
) error {
	boil.DebugMode = false // enable to see the raw SQL queries.
	const format = "similar releasers: %w"
	if err := nils.Check(ctx, exec, obj); err != nil {
		return fmt.Errorf(format, err)
	}
	if len(names) == 0 || len(*obj) > 0 {
		return nil
	}

	const escapedChar = "''"
	likes := make([]string, 0, len(names)*3)
	for _, name := range names {
		s := strings.ReplaceAll(name, "'", escapedChar)
		likes = append(likes, strings.ToUpper(s))
		likes = append(likes, strings.ToUpper(releaser.Title(s)))
		likes = append(likes, strings.ToUpper(releaser.Cell(s)))
	}
	slices.Sort(likes)
	likes = slices.Compact(likes)

	var query string
	var params []any

	switch match {
	case OnlyExact:
		sqlStr, p := postgres.SimilarToExact(likes...)
		query, params = string(sqlStr), p
	case OnlyMagazine:
		sqlStr, p := postgres.SimilarToMagazine(likes...)
		query, params = string(sqlStr), p
	case Only:
		sqlStr, p := postgres.SimilarToReleaser(likes...)
		query, params = string(sqlStr), p
	default:
		return fmt.Errorf(format, ErrMethod)
	}

	const pageNumber = 1
	pageSize := min(limit, PageSet)
	if pageSize <= 0 {
		pageSize = PageSet
	}
	l, o := limits(pageNumber, pageSize)
	query += " LIMIT " + strconv.Itoa(l) + " OFFSET " + strconv.Itoa(o)

	if err := queries.Raw(query, params...).Bind(ctx, exec, obj); err != nil {
		return fmt.Errorf(format, err)
	}

	obj.Slugs()
	return nil
}
