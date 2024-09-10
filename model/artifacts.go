package model

// Package file artifacts.go contains the database queries for the listing of sorted files.

import (
	"context"
	"fmt"
	"io"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/google/uuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Artifacts contain statistics for every artifact.
type Artifacts struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

// Public returns the total number of artifacts and the summed filesize of all artifacts that are not hidden.
func (f *Artifacts) Public(ctx context.Context, exec boil.ContextExecutor) error {
	if invalidExec(exec) {
		return ErrDB
	}
	if f.Bytes > 0 && f.Count > 0 {
		return nil
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(ClauseNoSoftDel),
		qm.From(From)).Bind(ctx, exec, f)
}

// ByKey returns the public files reversed ordered by the ID, key column.
func (f *Artifacts) ByKey(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if invalidExec(exec) {
		return nil, ErrDB
	}
	if err := f.Public(ctx, exec); err != nil {
		return nil, fmt.Errorf("f.Public: %w", err)
	}
	const clause = "id DESC"
	return models.Files(
		qm.Where(ClauseNoSoftDel),
		qm.OrderBy(clause),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, exec)
}

// ByOldest returns all of the file records sorted by the date issued.
func (f *Artifacts) ByOldest(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if invalidExec(exec) {
		return nil, ErrDB
	}
	if err := f.Public(ctx, exec); err != nil {
		return nil, fmt.Errorf("f.Public: %w", err)
	}
	const clause = "date_issued_year ASC NULLS LAST, " +
		"date_issued_month ASC NULLS LAST, " +
		"date_issued_day ASC NULLS LAST"
	return models.Files(
		qm.Where(ClauseNoSoftDel),
		qm.OrderBy(clause),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, exec)
}

// ByNewest returns all of the file records sorted by the date issued.
func (f *Artifacts) ByNewest(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if invalidExec(exec) {
		return nil, ErrDB
	}
	if err := f.Public(ctx, exec); err != nil {
		return nil, fmt.Errorf("f.Public: %w", err)
	}
	const clause = "date_issued_year DESC NULLS LAST, " +
		"date_issued_month DESC NULLS LAST, " +
		"date_issued_day DESC NULLS LAST"
	return models.Files(
		qm.Where(ClauseNoSoftDel),
		qm.OrderBy(clause),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, exec)
}

// ByUpdated returns all of the file records sorted by the date updated.
func (f *Artifacts) ByUpdated(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if invalidExec(exec) {
		return nil, ErrDB
	}
	if err := f.Public(ctx, exec); err != nil {
		return nil, fmt.Errorf("f.Public: %w", err)
	}
	const clause = "updatedat DESC"
	return models.Files(
		qm.Where(ClauseNoSoftDel),
		qm.OrderBy(clause),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, exec)
}

// ByHidden returns all of the file records that are hidden ~ soft deleted.
func (f *Artifacts) ByHidden(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if invalidExec(exec) {
		return nil, ErrDB
	}
	if err := f.byHidden(ctx, exec); err != nil {
		return nil, fmt.Errorf("f.Stat: %w", err)
	}
	const clause = "deletedat DESC"
	return models.Files(
		models.FileWhere.Deletedat.IsNotNull(),
		models.FileWhere.Deletedby.IsNotNull(),
		qm.WithDeleted(),
		qm.OrderBy(clause),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, exec)
}

func (f *Artifacts) byHidden(ctx context.Context, exec boil.ContextExecutor) error {
	if invalidExec(exec) {
		return ErrDB
	}
	if f.Bytes > 0 && f.Count > 0 {
		return nil
	}
	return models.NewQuery(
		models.FileWhere.Deletedat.IsNotNull(),
		models.FileWhere.Deletedby.IsNotNull(),
		qm.WithDeleted(),
		qm.Select(postgres.Columns()...),
		qm.From(From)).Bind(ctx, exec, f)
}

// ByForApproval returns all of the file records that are waiting to be marked for approval.
func (f *Artifacts) ByForApproval(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if invalidExec(exec) {
		return nil, ErrDB
	}
	if err := f.byForApproval(ctx, exec); err != nil {
		fmt.Fprint(io.Discard, err)
		return nil, nil
	}
	const clause = "id DESC"
	return models.Files(
		models.FileWhere.Deletedat.IsNotNull(),
		models.FileWhere.Deletedby.IsNull(),
		qm.WithDeleted(),
		qm.OrderBy(clause),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, exec)
}

func (f *Artifacts) byForApproval(ctx context.Context, exec boil.ContextExecutor) error {
	if invalidExec(exec) {
		return ErrDB
	}
	if f.Bytes > 0 && f.Count > 0 {
		return nil
	}
	return models.NewQuery(
		models.FileWhere.Deletedat.IsNotNull(),
		models.FileWhere.Deletedby.IsNull(),
		qm.WithDeleted(),
		qm.Select(postgres.Columns()...),
		qm.From(From)).Bind(ctx, exec, f)
}

// ByMagicErr returns all of the file records that require new magic numbers.
func (f *Artifacts) ByMagicErr(ctx context.Context, exec boil.ContextExecutor, binaryData bool) (
	models.FileSlice, error,
) {
	if invalidExec(exec) {
		return nil, ErrDB
	}
	/*
		SELECT DISTINCT "file_magic_type"
		FROM "files"
		ORDER BY "file_magic_type"
		LIMIT 500;
	*/
	equals := []string{"data", "tar archive", "Microsoft ASF"}
	ilikes := []string{
		"application/%", "Zip archive data%", "ARC archive data%", "ARJ archive data%", "RAR archive data%",
		"7-zip archive data%", "gzip compressed data%", "ASCII text%", "HTML document%", "Pascal source%", "ISO-8859 text%",
		"JPEG image data%", "GIF image data%", "PNG image data%", "PDF document%", "RIFF (little-endian) data%",
		"ISO Media%", "Fasttracker II%", "Ogg data%", "Audio file with%", "MPEG ADTS%",
		"AIX core file%", "C source,%", "C++ source,%", "FORTRAN program%", "ISO-8859 text%",
		"Little-endian UTF-16%", "MIT scheme%", "MS Windows icon resource%", "Microsoft Cabinet archive data,%",
		"Non-ISO extended-ASCII text%", "PC bitmap, Windows 3.x format%", "PCX ver. 3.0 image data%",
		"PE32 executable (GUI) Intel 80386%", "PE32 executable (console)%", "Python script%", "Quake I or II world or extension%",
		"AmigaGuide file%", "COM executable for%", "DCL command file%", "LHa (%", "MS-DOS executable%", "RFC 822 mail%",
		"Rich Text Format data%", "SMTP mail%", "SysEx File%", "UTF-8 Unicode%", "core file (Xenix)%", "diff output,%",
		"news or mail,%", "news, ASCII text%", "saved news,%",
	}
	mods := []qm.QueryMod{
		qm.Select(models.FileColumns.UUID, models.FileColumns.ID, models.FileColumns.FileMagicType),
		models.FileWhere.FileMagicType.IsNull(),
	}
	for _, s := range equals {
		mods = append(mods, qm.Or2(models.FileWhere.FileMagicType.EQ(null.StringFrom(s))))
	}
	for _, s := range ilikes {
		mods = append(mods, qm.Or2(models.FileWhere.FileMagicType.ILIKE(null.StringFrom(s))))
	}
	if binaryData {
		mods = append(mods,
			qm.Or2(models.FileWhere.FileMagicType.EQ(null.StringFrom("Binary data"))))
	}
	mods = append(mods, qm.WithDeleted())
	return models.Files(mods...).All(ctx, exec)
}

// ByTextPlatform returns all of the file records that are text based, either text or textamiga.
func (f *Artifacts) ByTextPlatform(ctx context.Context, exec boil.ContextExecutor) (
	models.FileSlice, error,
) {
	if invalidExec(exec) {
		return nil, ErrDB
	}
	return models.Files(
		qm.Select(models.FileColumns.UUID, models.FileColumns.ID),
		models.FileWhere.Platform.EQ(null.StringFrom("text")),
		qm.Or2(models.FileWhere.Platform.EQ(null.StringFrom("textamiga"))),
		qm.WithDeleted()).All(ctx, exec)
}

// ByUnwanted returns all of the file records that are flagged by Google as unwanted.
func (f *Artifacts) ByUnwanted(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (
	models.FileSlice, error,
) {
	if invalidExec(exec) {
		return nil, ErrDB
	}
	if err := f.byUnwanted(ctx, exec); err != nil {
		return nil, fmt.Errorf("f.StatUnwanted: %w", err)
	}
	const clause = "id DESC"
	return models.Files(
		models.FileWhere.FileSecurityAlertURL.IsNotNull(),
		qm.WithDeleted(),
		qm.OrderBy(clause),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, exec)
}

func (f *Artifacts) byUnwanted(ctx context.Context, exec boil.ContextExecutor) error {
	if invalidExec(exec) {
		return ErrDB
	}
	if f.Bytes > 0 && f.Count > 0 {
		return nil
	}
	return models.NewQuery(
		models.FileWhere.FileSecurityAlertURL.IsNotNull(),
		qm.WithDeleted(),
		qm.Select(postgres.Columns()...),
		qm.From(From)).Bind(ctx, exec, f)
}

// Description returns a list of files that match the search terms.
// The search terms are matched against the record_title column.
// The results are ordered by the filename column in ascending order.
func (f *Artifacts) Description(ctx context.Context, exec boil.ContextExecutor, terms []string) (
	models.FileSlice, error,
) {
	if invalidExec(exec) {
		return nil, ErrDB
	}
	if terms == nil {
		return models.FileSlice{}, nil
	}
	mods := []qm.QueryMod{}
	mods = append(mods, qm.Where(ClauseNoSoftDel))
	const clauseT = "to_tsvector(record_title) @@ to_tsquery(?)"
	const clauseC = "to_tsvector(comment) @@ to_tsquery(?)"
	for i, term := range terms {
		term = fmt.Sprintf("'%s'", term) // the single quotes are required for terms containing spaces
		if i == 0 {
			mods = append(mods, qm.Where(clauseT, term))
			mods = append(mods, qm.Or(clauseC, term))
			continue
		}
		mods = append(mods, qm.Or(clauseT, term))
		mods = append(mods, qm.Or(clauseC, term))
	}
	mods = append(mods, qm.Limit(Maximum))
	fs, err := models.Files(mods...).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("models all files by description search: %w", err)
	}
	return fs, nil
}

// Filename returns a list of files that match the search terms.
// The search terms are matched against the filename column.
// The results are ordered by the filename column in ascending order.
func (f *Artifacts) Filename(ctx context.Context, exec boil.ContextExecutor, terms []string) (
	models.FileSlice, error,
) {
	if invalidExec(exec) {
		return nil, ErrDB
	}
	if terms == nil {
		return models.FileSlice{}, nil
	}
	mods := []qm.QueryMod{}
	mods = append(mods, qm.Where(ClauseNoSoftDel))
	for i, term := range terms {
		if i == 0 {
			mods = append(mods, qm.Where("filename ~ ? OR filename ILIKE ? OR filename ILIKE ? OR filename ILIKE ?",
				term, term+"%", "%"+term, "%"+term+"%"))
			continue
		}
		mods = append(mods, qm.Or("filename ~ ? OR filename ILIKE ? OR filename ILIKE ? OR filename ILIKE ?",
			term, term+"%", "%"+term, "%"+term+"%"))
	}
	mods = append(mods, qm.OrderBy("filename ASC"), qm.Limit(Maximum))
	fs, err := models.Files(mods...).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("models all files by filename search: %w", err)
	}
	return fs, nil
}

// ID returns a list of files that match the list of record ids or uuids.
func (f *Artifacts) ID(
	ctx context.Context, exec boil.ContextExecutor, ids []int, uuids ...uuid.UUID) (
	models.FileSlice, error,
) {
	if invalidExec(exec) {
		return nil, ErrDB
	}
	if (ids == nil && uuids == nil) || (len(ids) == 0 && len(uuids) == 0) {
		return models.FileSlice{}, nil
	}
	mods := []qm.QueryMod{}
	for _, id := range ids {
		if id < 1 {
			continue
		}
		mods = append(mods, qm.Or("id = ?", id))
	}
	for _, uuid := range uuids {
		mods = append(mods, qm.Or("uuid = ?", uuid.String()))
	}
	mods = append(mods, qm.Limit(Maximum), qm.WithDeleted())
	fs, err := models.Files(mods...).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("models all files by id search: %w", err)
	}
	return fs, nil
}
