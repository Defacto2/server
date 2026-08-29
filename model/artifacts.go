package model

// Package file artifacts.go contains the database queries for the listing of sorted files.

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"
)

// Artifacts statistics.
type Artifacts struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
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
	// TODO add cache time.Time value?
	return obj.Bytes > 0 && obj.Count > 0
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
// The search terms are matched against the record_title column.
// The results are ordered by the filename column in ascending order.
func OnlyDescriptions(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor, terms []string) (
	models.FileSlice, error,
) {
	const format = "only descriptions %s: %w"
	if err := nils.Check(ctx, sl, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(format, "check", err)
	}
	if len(terms) == 0 {
		return models.FileSlice{}, nil
	}

	mods := onlyDescriptions(terms)
	sl.Debug("only descriptions", slog.Any("terms", terms), slog.Any("mods", mods))

	fs, err := models.Files(mods...).All(ctx, exec)
	if err != nil {
		return models.FileSlice{}, fmt.Errorf(format, "all", err)
	}

	return fs, nil
}

func onlyDescriptions(terms []string) []qm.QueryMod {
	mods := []qm.QueryMod{}
	mods = append(mods, qm.Where(ClauseNoSoftDel))
	const title = "to_tsvector(record_title) @@ websearch_to_tsquery(?)"
	const comment = "to_tsvector(comment) @@ websearch_to_tsquery(?)"
	for i, term := range terms {
		term = fmt.Sprintf("'%s'", term) // the single quotes are required for terms containing spaces
		if i == 0 {
			mods = append(mods, qm.Where(title, term))
			mods = append(mods, qm.Or(comment, term))
			continue
		}
		mods = append(mods, qm.Or(title, term))
		mods = append(mods, qm.Or(comment, term))
	}
	mods = append(mods, qm.Limit(Maximum))
	return mods
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
	if len(terms) == 0 {
		return models.FileSlice{}, nil
	}

	mods := onlyFilenames(terms)

	fs, err := models.Files(mods...).All(ctx, exec)
	if err != nil {
		return models.FileSlice{}, fmt.Errorf(format, "all", err)
	}

	return fs, nil
}

func onlyFilenames(terms []string) []qm.QueryMod {
	mods := []qm.QueryMod{}
	mods = append(mods, qm.Where(ClauseNoSoftDel))
	const clause = "filename ~ ? OR filename ILIKE ? OR filename ILIKE ? OR filename ILIKE ?"
	for i, term := range terms {
		if i == 0 {
			mods = append(mods, qm.Where(clause, term, term+"%", "%"+term, "%"+term+"%"))
			continue
		}
		mods = append(mods, qm.Or(clause, term, term+"%", "%"+term, "%"+term+"%"))
	}
	mods = append(mods, qm.OrderBy("filename ASC"), qm.Limit(Maximum))
	return mods
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
	const format = "only unique ids %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(format, "check", err)
	}

	if len(ids) == 0 && len(uuids) == 0 {
		return models.FileSlice{}, nil
	}

	mods := onlyUniqueIDs(ids, uuids...)
	fs, err := models.Files(mods...).All(ctx, exec)
	if err != nil {
		return models.FileSlice{}, fmt.Errorf(format, "all", err)
	}

	return fs, nil
}

func onlyUniqueIDs(ids []int, uuids ...uuid.UUID) []qm.QueryMod {
	mods := []qm.QueryMod{}
	for id := range slices.Values(ids) {
		if id < 1 {
			continue
		}
		mods = append(mods, qm.Or("id = ?", id))
	}
	for uuid := range slices.Values(uuids) {
		mods = append(mods, qm.Or("uuid = ?", uuid.String()))
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
