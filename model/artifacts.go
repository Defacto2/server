package model

// Package file artifacts.go contains the database queries for the listing of sorted files.

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"
	"uuid"

	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

// Artifacts statistics.
type Artifacts struct {
	Bytes     int       `boil:"size_total"`
	Count     int       `boil:"count_total"`
	MinYear   int       `boil:"min_year"`
	MaxYear   int       `boil:"max_year"`
	UpdatedAt time.Time `boil:"-"`
}

// Public sets the [Artifacts] statistics for file artifacts that are not marked as hidden.
func (obj *Artifacts) Public(ctx context.Context, exec boil.ContextExecutor) error {
	const format = "set public artifacts %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}
	if obj.useCache() {
		return nil
	}

	err := models.NewQuery(qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel), qm.From(From),
	).Bind(ctx, exec, obj)
	if err != nil {
		return fmt.Errorf(format, "new query", err)
	}
	obj.UpdatedAt = time.Now()

	return nil
}

type By struct {
	Clause string
	Offset int
	Limit  int
}

func (by By) all(ctx context.Context, exec boil.ContextExecutor, f *Artifacts) (models.FileSlice, error) {
	const format = "by artifacts: %w"

	if err := nils.Check(ctx, exec, f); err != nil {
		return models.FileSlice{}, fmt.Errorf(format, err)
	}

	if err := f.Public(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(format, err)
	}

	fs, err := models.Files(
		qm.Where(ClauseNoSoftDel),
		qm.OrderBy(by.Clause),
		qm.Offset(calc(by.Offset, by.Limit)),
		qm.Limit(by.Limit),
	).All(ctx, exec)
	if err != nil {
		return models.FileSlice{}, fmt.Errorf(format, err)
	}

	return fs, nil
}

// ByKey returns the public files reversed ordered, numeric key ID column.
func (obj *Artifacts) ByKey(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	by := By{
		Clause: "id DESC",
		Offset: offset, Limit: limit,
	}
	return by.all(ctx, exec, obj)
}

// ByOldest returns all of the file records sorted by the date issued.
func (obj *Artifacts) ByOldest(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	const clause = "date_issued_year ASC NULLS LAST, " +
		"date_issued_month ASC NULLS LAST, " +
		"date_issued_day ASC NULLS LAST"
	by := By{
		Clause: clause,
		Offset: offset, Limit: limit,
	}
	return by.all(ctx, exec, obj)
}

// ByNewest returns all of the file records sorted by the date issued.
func (obj *Artifacts) ByNewest(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	const clause = "date_issued_year DESC NULLS LAST, " +
		"date_issued_month DESC NULLS LAST, " +
		"date_issued_day DESC NULLS LAST"
	by := By{
		Clause: clause,
		Offset: offset, Limit: limit,
	}
	return by.all(ctx, exec, obj)
}

// ByUpdated returns all of the file records sorted by the date updated.
func (obj *Artifacts) ByUpdated(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	const clause = "updatedat DESC"
	by := By{
		Clause: clause,
		Offset: offset, Limit: limit,
	}
	return by.all(ctx, exec, obj)
}

// OnlyUnwanted returns all of the file records that are flagged by Google as unwanted.
func (obj *Artifacts) OnlyUnwanted(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	const format = "only unwanted: %w"

	if err := nils.Check(ctx, exec, obj); err != nil {
		return models.FileSlice{}, fmt.Errorf(format, err)
	}

	// fetch record mods
	empty := null.StringFrom("")
	const size = 6
	mods := make([]qm.QueryMod, 0, size)
	mods = append(mods,
		models.FileWhere.FileSecurityAlertURL.IsNotNull(),
		models.FileWhere.FileSecurityAlertURL.NEQ(empty),
		qm.WithDeleted(),
	)

	if err := obj.onlyUnwantedStats(ctx, exec, mods); err != nil {
		return models.FileSlice{}, fmt.Errorf(format, err)
	}
	obj.UpdatedAt = time.Now()

	// ordering mods
	const clause = "id DESC"
	mods = append(mods, qm.OrderBy(clause),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit))

	return models.Files(mods...).All(ctx, exec)
}

func (obj *Artifacts) onlyUnwantedStats(ctx context.Context, exec boil.ContextExecutor, mods []qm.QueryMod) error {
	if obj.useCache() {
		return nil
	}

	mods = append(mods,
		qm.Select(SummCols()...),
		qm.From(From))

	return models.NewQuery(mods...).Bind(ctx, exec, obj)
}

func (obj *Artifacts) useCache() bool {
	// return obj.Bytes > 0 && obj.Count > 0
	return !obj.UpdatedAt.IsZero() && time.Since(obj.UpdatedAt) < 10*time.Minute
}

// OnlyApproval returns all of the file records that are waiting to be marked for approval.
//
// This should not bind values to Artifacts struct as it can fail with a scan error due to unapproved files
// missing bytes and minyear/maxyear values.
func OnlyApproval(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	const format = "only approval: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(format, err)
	}

	const clause = "id DESC"

	return models.Files(
		qm.WithDeleted(),
		models.FileWhere.Deletedat.IsNotNull(),
		models.FileWhere.Deletedby.IsNull(),
		qm.OrderBy(clause),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
	).All(ctx, exec)
}

// OnlyDescriptions returns a list of files that match the search terms.
// The search terms are matched against the record_title and comment columns.
// The results are ordered by the filename column in ascending order.
func OnlyDescriptions(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor, terms []string) (
	models.FileSlice, error,
) {
	const format = "only descriptions %s: %w"
	if err := nils.Check(ctx, sl, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(format, "check", err)
	}
	mods := onlyDescriptions(terms)
	if len(mods) == 0 {
		return models.FileSlice{}, nil
	}

	sl.Debug("only descriptions", slog.Any("terms", terms), slog.Any("mods", mods))

	fs, err := models.Files(mods...).All(ctx, exec)
	if err != nil {
		return models.FileSlice{}, fmt.Errorf(format, "all", err)
	}

	return fs, nil
}

func onlyDescriptions(terms []string) []qm.QueryMod {
	orMods := make([]qm.QueryMod, 0, len(terms))
	const clause = "to_tsvector('english', concat_ws(' ', record_title, comment)) @@ plainto_tsquery('english', ?)"
	for _, term := range terms {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}
		if len(orMods) == 0 {
			orMods = append(orMods, qm.Where(clause, term))
			continue
		}
		orMods = append(orMods, qm.Or(clause, term))
	}
	if len(orMods) == 0 {
		return nil
	}
	return []qm.QueryMod{
		qm.Where(ClauseNoSoftDel),
		qm.Expr(orMods...),
		qm.OrderBy("filename ASC"),
		qm.Limit(Maximum),
	}
}

// OnlyFilenames returns a list of files that match the search terms.
// The search terms are matched against the filename column.
// The results are ordered by the filename column in ascending order.
func OnlyFilenames(ctx context.Context, exec boil.ContextExecutor, terms []string) (
	models.FileSlice, error,
) {
	const format = "only filenames %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(format, "check", err)
	}
	mods := onlyFilenames(terms)
	if len(mods) == 0 {
		return models.FileSlice{}, nil
	}

	fs, err := models.Files(mods...).All(ctx, exec)
	if err != nil {
		return models.FileSlice{}, fmt.Errorf(format, "all", err)
	}

	return fs, nil
}

// escapeILIKE escapes PostgreSQL LIKE/ILIKE wildcard characters using '\' as the escape character.
func escapeILIKE(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func onlyFilenames(terms []string) []qm.QueryMod {
	orMods := make([]qm.QueryMod, 0, len(terms))
	const clause = `filename ILIKE ? ESCAPE '\'`
	for _, term := range terms {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}
		pattern := "%" + escapeILIKE(term) + "%"
		if len(orMods) == 0 {
			orMods = append(orMods, qm.Where(clause, pattern))
			continue
		}
		orMods = append(orMods, qm.Or(clause, pattern))
	}
	if len(orMods) == 0 {
		return nil
	}
	return []qm.QueryMod{
		qm.Where(ClauseNoSoftDel),
		qm.Expr(orMods...),
		qm.OrderBy("filename ASC"),
		qm.Limit(Maximum),
	}
}

// OnlyHidden returns all of the file records that are marked as hidden using soft delete.
func OnlyHidden(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	const format = "by hidden artifacts: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(format, err)
	}

	const clause = "deletedat DESC"

	return models.Files(
		models.FileWhere.Deletedat.IsNotNull(),
		models.FileWhere.Deletedby.IsNotNull(),
		qm.WithDeleted(),
		qm.OrderBy(clause),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit),
	).All(ctx, exec)
}

// OnlyTexts returns all of the file records that are text based, both text or textamiga.
func OnlyTexts(ctx context.Context, exec boil.ContextExecutor) (
	models.FileSlice, error,
) {
	const format = "only texts %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(format, "check", err)
	}

	return models.Files(
		qm.Select(models.FileColumns.UUID, models.FileColumns.ID),
		models.FileWhere.Platform.EQ(null.StringFrom("text")),
		qm.Or2(models.FileWhere.Platform.EQ(null.StringFrom("textamiga"))),
		qm.WithDeleted(),
	).All(ctx, exec)
}

// OnlyUniqueIDs returns a list of files that match the lists of records with id or uuid.
func OnlyUniqueIDs(
	ctx context.Context, exec boil.ContextExecutor, ids []int, uuids ...uuid.UUID) (
	models.FileSlice, error,
) {
	return OnlyUniqueIDsWithPrefixes(ctx, exec, ids, uuids, nil)
}

// OnlyUniqueIDsWithPrefixes returns a list of files matching exact IDs, exact UUIDs,
// or case-insensitive UUID prefixes (e.g. "c8cd0b4c").
func OnlyUniqueIDsWithPrefixes(
	ctx context.Context, exec boil.ContextExecutor, ids []int, uuids []uuid.UUID, prefixes []string,
) (models.FileSlice, error) {
	const format = "only unique ids %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(format, "check", err)
	}

	if len(ids) == 0 && len(uuids) == 0 && len(prefixes) == 0 {
		return models.FileSlice{}, nil
	}

	mods := onlyUniqueIDs(ids, uuids, prefixes)
	if len(mods) == 0 {
		return models.FileSlice{}, nil
	}
	fs, err := models.Files(mods...).All(ctx, exec)
	if err != nil {
		return models.FileSlice{}, fmt.Errorf(format, "all", err)
	}

	return fs, nil
}

func onlyUniqueIDs(ids []int, uuids []uuid.UUID, prefixes []string) []qm.QueryMod {
	validIDs := make([]int64, 0, len(ids))
	for id := range slices.Values(ids) {
		if id >= 1 {
			validIDs = append(validIDs, int64(id))
		}
	}
	validUUIDs := make([]string, 0, len(uuids))
	for u := range slices.Values(uuids) {
		if u != uuid.Nil() {
			validUUIDs = append(validUUIDs, u.String())
		}
	}
	validPrefixes := make([]string, 0, len(prefixes))
	for p := range slices.Values(prefixes) {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			validPrefixes = append(validPrefixes, trimmed)
		}
	}

	var mods []qm.QueryMod
	if len(validIDs) > 0 {
		mods = append(mods, models.FileWhere.ID.IN(validIDs))
	}
	if len(validUUIDs) > 0 {
		if len(mods) == 0 {
			mods = append(mods, models.FileWhere.UUID.IN(validUUIDs))
		} else {
			mods = append(mods, qm.Or2(models.FileWhere.UUID.IN(validUUIDs)))
		}
	}
	const prefixClause = `uuid ILIKE ? ESCAPE '\'`
	for _, p := range validPrefixes {
		pattern := escapeILIKE(p) + "%"
		if len(mods) == 0 {
			mods = append(mods, qm.Where(prefixClause, pattern))
		} else {
			mods = append(mods, qm.Or(prefixClause, pattern))
		}
	}
	if len(mods) == 0 {
		return nil
	}
	mods = append(mods, qm.Limit(Maximum), qm.WithDeleted())
	return mods
}

// OnlyMagicErrs returns all of the file records using legacy magic numbers that require replacements.
// The binary data bool will also replace magic strings that are set to "Binary data".
func OnlyMagicErrs(ctx context.Context, exec boil.ContextExecutor, binaryData bool) (
	models.FileSlice, error,
) {
	const format = "only magic errs %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(format, "check", err)
	}
	exacts := [...]string{"data", "tar archive", "Microsoft ASF"}
	ilikes := [...]string{
		"application/%", "Zip archive data%", "ARC archive data%", "ARJ archive data%",
		"RAR archive data%", "7-zip archive data%", "gzip compressed data%", "ASCII text%",
		"HTML document%", "Pascal source%", "ISO-8859 text%", "JPEG image data%", "GIF image data%",
		"PNG image data%", "PDF document%", "RIFF (little-endian) data%", "ISO Media%",
		"Fasttracker II%", "Ogg data%", "Audio file with%", "MPEG ADTS%", "AIX core file%",
		"C source,%", "C++ source,%", "FORTRAN program%", "ISO-8859 text%",
		"Little-endian UTF-16%", "MIT scheme%", "MS Windows icon resource%",
		"Microsoft Cabinet archive data,%", "Non-ISO extended-ASCII text%",
		"PC bitmap, Windows 3.x format%", "PCX ver. 3.0 image data%",
		"PE32 executable (GUI) Intel 80386%", "PE32 executable (console)%", "Python script%",
		"Quake I or II world or extension%", "AmigaGuide file%", "COM executable for%",
		"DCL command file%", "LHa (%", "MS-DOS executable%", "RFC 822 mail%",
		"Rich Text Format data%", "SMTP mail%", "SysEx File%", "UTF-8 Unicode%",
		"core file (Xenix)%", "diff output,%", "news or mail,%", "news, ASCII text%",
		"saved news,%", "ID tags data%", "VISX image file%",
	}

	mods := []qm.QueryMod{
		qm.Select(
			models.FileColumns.UUID,
			models.FileColumns.ID,
			models.FileColumns.FileMagicType),
		models.FileWhere.FileMagicType.IsNull(),
	}

	for _, exact := range exacts {
		mods = append(mods,
			qm.Or2(models.FileWhere.FileMagicType.EQ(null.StringFrom(exact))))
	}

	for _, ilike := range ilikes {
		mods = append(mods,
			qm.Or2(models.FileWhere.FileMagicType.ILIKE(null.StringFrom(ilike))))
	}

	if binaryData {
		mods = append(mods,
			qm.Or2(models.FileWhere.FileMagicType.EQ(null.StringFrom("Binary data"))))
	}
	mods = append(mods, qm.WithDeleted())

	return models.Files(mods...).All(ctx, exec)
}
