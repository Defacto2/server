package model

// Package releaser.go contains the database queries the releasers and groups.

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/releaser"
	namer "github.com/Defacto2/releaser/name"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// OrderBy is the sorting order for ALL the releasers.
type OrderBy uint

const (
	Prolific     OrderBy = iota // Prolific orders by the total artifact count.
	Alphabetical                // Alphabetical orders by the releaser name.
	Oldest                      // Oldest orders by the year of the first artifact.
)

// ReleaserNames is a distinct data list of releasers.
type ReleaserNames []ReleaserName

// ReleaserName is a releaser name.
type ReleaserName struct {
	Name string `boil:"releaser"`
}

// Distinct gets the unique releaser names.
func (r *ReleaserNames) Distinct(ctx context.Context, exec boil.ContextExecutor) error {
	if invalidExec(exec) {
		return ErrDB
	}
	query := string(postgres.Releasers())
	return queries.Raw(query).Bind(ctx, exec, r)
}

// DistinctGroups gets the unique releaser names that are groups.
func (r *ReleaserNames) DistinctGroups(ctx context.Context, exec boil.ContextExecutor) error {
	if invalidExec(exec) {
		return ErrDB
	}
	query := string(postgres.ReleasersAlphabetical())
	return queries.Raw(query).Bind(ctx, exec, r)
}

// Releasers is a collection of releasers.
type Releasers []*struct {
	Unique Releaser `boil:",bind"` // Unique releaser.
}

func (r *Releasers) String() string {
	names := make([]string, 0, len(*r))
	for _, name := range *r {
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
func (r *Releasers) Where(ctx context.Context, exec boil.ContextExecutor, name string) (models.FileSlice, error) {
	if invalidExec(exec) {
		return nil, ErrDB
	}
	s, _ := namer.Humanize(namer.Path(name))
	if s == "" {
		return nil, nil
	}
	n := strings.ToUpper(s)
	x := null.StringFrom(n)
	return models.Files(
		qm.Where("upper(group_brand_for) = ? OR upper(group_brand_by) = ?", x, x),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, exec)
}

// Limit gets the unique releaser names and their total file count and file sizes.
// When reorder is true the results are ordered by the total file counts.
func (r *Releasers) Limit(ctx context.Context, exec boil.ContextExecutor, order OrderBy, limit, page int) error {
	if r != nil && len(*r) > 0 {
		return nil
	}
	if invalidExec(exec) {
		return ErrDB
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
		return ErrOrderBy
	}
	if limit > 0 {
		if page < 1 {
			page = 1
		}
		limit, offset := calculateLimitAndOffset(page, limit)
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	}
	if err := queries.Raw(query).Bind(ctx, exec, r); err != nil {
		return fmt.Errorf("queries.Raw: %w", err)
	}
	r.Slugs()
	return nil
}

// Similar finds the unique releaser names that are similar to the named strings.
// The results are ordered by the total file counts.
// The required limit is the maximum number of results to return or defaults to 10.
func (r *Releasers) Similar(ctx context.Context, exec boil.ContextExecutor, limit int, names ...string) error {
	if invalidExec(exec) {
		return ErrDB
	}
	return r.similar(ctx, exec, limit, "releaser", names...)
}

// Initialism finds the unique releaser names that match the named strings.
// The results are ordered by the total file counts.
// The required limit is the maximum number of results to return or defaults to 10.
func (r *Releasers) Initialism(ctx context.Context, exec boil.ContextExecutor, limit int, names ...string) error {
	if invalidExec(exec) {
		return ErrDB
	}
	return r.similar(ctx, exec, limit, "initialism", names...)
}

// SimilarMagazine finds the unique releaser names that are similar to the named strings.
// The results are ordered by the total file counts.
// The required limit is the maximum number of results to return or defaults to 10.
func (r *Releasers) SimilarMagazine(ctx context.Context, exec boil.ContextExecutor, limit int, names ...string) error {
	if invalidExec(exec) {
		return ErrDB
	}
	return r.similar(ctx, exec, limit, "magazine", names...)
}

func (r *Releasers) similar(
	ctx context.Context, exec boil.ContextExecutor, limit int, lookup string, names ...string,
) error {
	boil.DebugMode = false // Enable to see the raw SQL queries.
	if len(names) == 0 {
		return nil
	}
	if r != nil && len(*r) > 0 {
		return nil
	}
	if invalidExec(exec) {
		return ErrDB
	}

	likes := names
	for name := range slices.Values(names) {
		likes = append(likes, releaser.Title(name))
		likes = append(likes, releaser.Cell(name))
	}
	for i := range likes {
		likes[i] = strings.ToUpper(likes[i])
	}
	slices.Sort(likes)
	likes = removeDuplicates(likes)
	likes = slices.Compact(likes)
	var query string
	switch lookup {
	case "initialism":
		query = string(postgres.SimilarInitialism(likes...))
	case "magazine":
		query = string(postgres.SimilarToMagazine(likes...))
	default:
		query = string(postgres.SimilarToReleaser(likes...))
	}
	{
		const page, maxPages = 1, 10
		size := limit | maxPages
		val, offset := calculateLimitAndOffset(page, size)
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", val, offset)
	}
	if err := queries.Raw(query).Bind(ctx, exec, r); err != nil {
		return fmt.Errorf("similar magazine queries raw: %w", err)
	}
	r.Slugs()
	return nil
}

func removeDuplicates(strings []string) []string {
	unique := make(map[string]struct{})
	var result []string
	for s := range slices.Values(strings) {
		if _, exists := unique[s]; !exists {
			unique[s] = struct{}{}
			result = append(result, s)
		}
	}
	return result
}

func calculateLimitAndOffset(pageNumber int, pageSize int) (int, int) {
	limit := pageSize
	offset := (pageNumber - 1) * pageSize
	return limit, offset
}

// BBS gets the unique BBS site names and their total file count and file sizes.
func (r *Releasers) BBS(ctx context.Context, exec boil.ContextExecutor, order OrderBy) error {
	if len(*r) > 0 {
		return nil
	}
	if invalidExec(exec) {
		return ErrDB
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
		return ErrOrderBy
	}
	if err := queries.Raw(query).Bind(ctx, exec, r); err != nil {
		return fmt.Errorf("queries.Raw: %w", err)
	}
	r.Slugs()
	return nil
}

// FTP gets the unique FTP site names and their total file count and file sizes.
func (r *Releasers) FTP(ctx context.Context, exec boil.ContextExecutor) error {
	if len(*r) > 0 {
		return nil
	}
	if invalidExec(exec) {
		return ErrDB
	}
	if err := queries.Raw(string(postgres.FTPsAlphabetical())).Bind(ctx, exec, r); err != nil {
		return fmt.Errorf("queries.Raw: %w", err)
	}
	r.Slugs()
	return nil
}

// Magazine gets the unique magazine titles and their total issue count and file sizes.
func (r *Releasers) MagazineAZ(ctx context.Context, exec boil.ContextExecutor) error {
	if len(*r) > 0 {
		return nil
	}
	if invalidExec(exec) {
		return ErrDB
	}
	if err := queries.Raw(string(postgres.MagazinesAlphabetical())).Bind(ctx, exec, r); err != nil {
		return fmt.Errorf("queries.Raw: %w", err)
	}
	r.Slugs()
	return nil
}

// Magazine gets the unique magazine titles and their total issue count and file sizes.
func (r *Releasers) Magazine(ctx context.Context, exec boil.ContextExecutor) error {
	if len(*r) > 0 {
		return nil
	}
	if invalidExec(exec) {
		return ErrDB
	}
	if err := queries.Raw(string(postgres.MagazinesOldest())).Bind(ctx, exec, r); err != nil {
		return fmt.Errorf("queries.Raw: %w", err)
	}
	r.Slugs()
	return nil
}

// Slugs sets URL friendly strings to the Group names.
func (r *Releasers) Slugs() {
	for _, releaser := range *r {
		releaser.Unique.URI = helper.Slug(releaser.Unique.Name)
	}
}
