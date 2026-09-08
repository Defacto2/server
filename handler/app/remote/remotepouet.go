//nolint:tagliatelle
package remote

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Defacto2/archive"
	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/app/internal/simple"
	"github.com/Defacto2/server/handler/pouet"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/labstack/echo/v5"
)

// PouetLink is the response from the task of GetDemozooFile.
type PouetLink struct {
	UUID        string `json:"uuid"`         // UUID is the file production UUID.
	Releaser1   string `json:"releaser1"`    // Releaser1 is the first releaser of the file.
	Releaser2   string `json:"releaser2"`    // Releaser2 is the second releaser of the file.
	Title       string `json:"title"`        // Title is the file title.
	Filename    string `json:"filename"`     // Filename is the file name of the download.
	Content     string `json:"content"`      // Content is the file archive content.
	FileType    string `json:"file_type"`    // Type is the file type.
	FileHash    string `json:"file_hash"`    // Hash is the file integrity hash.
	Platform    string `json:"platform"`     // Platform is the file platform.
	Section     string `json:"section"`      // Section is the file section.
	Error       string `json:"error"`        // Error is the error message if the download or record update failed.
	PouetID     int    `json:"id"`           // PouetID is the Pouet prod which ID.
	DemozooID   int    `json:"demozoo_prod"` // DemozooID is the production ID.
	FileSize    int    `json:"file_size"`    // Size is the file size in bytes.
	IssuedYear  int16  `json:"issued_year"`  // Year is the year the file was issued.
	IssuedMonth int16  `json:"issued_month"` // Month is the month the file was issued.
	IssuedDay   int16  `json:"issued_day"`   // Day is the day the file was issued.
}

// Download fetches the download link from Pouet and saves it to the download directory.
// It then runs Update to modify the database record with various metadata from the file and Pouet record API data.
func (got *PouetLink) Download(
	ctx context.Context, sl *slog.Logger, c *echo.Context, tx *sql.Tx, download dir.Directory,
) error {
	const msg = "pouet link download"
	const format = "%s for id %d: %w"
	id := got.PouetID
	if err := nils.Check(ctx, sl, c, tx); err != nil {
		return fmt.Errorf(format, "check", id, err)
	}
	var prod pouet.Production
	if _, err := prod.Get(ctx, id); err != nil {
		sl.Info(msg+" get production error", slog.Int("id", id), slog.Any("error", err))
		return fmt.Errorf(format, "but could not get record", id, err)
	}
	downloadURL := prod.Download
	if downloadURL == "" {
		sl.Info(msg+" offers no download url", slog.Int("id", id))
		return nil
	}
	resp, err := GetFile(ctx, sl, TimeoutLong, downloadURL)
	if err != nil {
		sl.Info(msg+" get file download error",
			slog.Int("id", id), slog.String("url", downloadURL), slog.Any("error", err))
		return fmt.Errorf(format, "but could not get the file download "+downloadURL, id, err)
	}
	base := filepath.Base(downloadURL)
	dst := filepath.Join(download.Path(), got.UUID)
	got.Filename = base
	if err := helper.RenameFileOW(resp.Path, dst); err != nil {
		sameFiles, err := helper.FileMatch(resp.Path, dst)
		if err != nil {
			sl.Info(msg+" got file but cannot rename error", slog.Int("id", id),
				slog.String("dst", dst), slog.Any("error", err))
			return fmt.Errorf(format, "but could not rename the file download to "+dst, id, err)
		}
		if !sameFiles {
			const s = "was successful but will not overwrite the existing file"
			sl.Info(msg+" "+s, slog.Int("id", id), slog.String("dst", dst))
			return fmt.Errorf(format, s+" "+dst, id, ErrExist)
		}
	}
	got.Filename = base
	got.Error = ""
	if i, err := strconv.Atoi(prod.Demozoo); err == nil && i > 0 {
		got.DemozooID = i
	}
	y, m, d := prod.Released()
	got.IssuedYear = y
	got.IssuedMonth = m
	got.IssuedDay = d
	r1, r2 := prod.Releasers()
	got.Releaser1 = r1
	got.Releaser2 = r2
	got.Title = prod.Title
	plat, sect := prod.PlatformType()
	got.Platform = plat.String()
	got.Section = sect.String()
	if err := got.Stat(ctx, sl, c, tx, download); err != nil {
		sl.Info(msg, slog.Int("id", id), slog.Any("error", err))
	}
	return nil
}

// Stat sets the file size, hash, type, and archive content of the file.
// The UUID is used to locate the file in the download directory.
func (got *PouetLink) Stat(
	ctx context.Context, sl *slog.Logger, c *echo.Context, tx *sql.Tx, download dir.Directory,
) error {
	const format = "pouet link stat %s: %w"
	if err := nils.Check(ctx, sl, c, tx); err != nil {
		return fmt.Errorf(format, "check", err)
	}
	name := filepath.Join(download.Path(), got.UUID)
	if got.FileSize == 0 {
		stat, err := os.Stat(name)
		if err != nil {
			return fmt.Errorf(format, "file download "+name, err)
		}
		got.FileSize = int(stat.Size())
	}
	strong, err := helper.StrongIntegrity(name)
	if err != nil {
		return fmt.Errorf(format, "file download strong integrity hash "+name, err)
	}
	got.FileHash = strong
	if got.FileType == "" {
		got.FileType = simple.MagicAsTitle(sl, name)
	}
	return got.ArchiveContent(ctx, sl, c, tx, name)
}

// ArchiveContent sets the archive content and readme text of the source file.
func (got *PouetLink) ArchiveContent(
	ctx context.Context, sl *slog.Logger, c *echo.Context, tx *sql.Tx, src string,
) error {
	const msg = "pouet link archive content"
	const format = msg + " %s: %w"
	if err := nils.Check(ctx, sl, c, tx); err != nil {
		return fmt.Errorf(format, "check", err)
	}
	files, err := archive.Lists(ctx, src)
	if err != nil {
		sl.Info(msg+" list caused an error",
			slog.String("src", src), slog.String("filename", got.Filename), slog.Any("error", err))
		return c.JSON(http.StatusOK, got)
	}
	got.Content = strings.Join(files, "\n")
	if err := got.Update(ctx, c, tx); err != nil {
		sl.Info(msg + " update caused an error")
	}
	const html = `<p class="text-success">Successful Pouet update</p>`
	return c.HTML(http.StatusOK, html)
}

// Update modifies the database record using data provided by the DemozooLink struct.
// A JSON response is returned with the success status of the update.
func (got *PouetLink) Update(ctx context.Context, c *echo.Context, tx *sql.Tx) error {
	const format = "pouet link update %s uuid %s: %w"
	uid := got.UUID
	if err := nils.Check(ctx, c, tx); err != nil {
		return fmt.Errorf(format, "check", uid, err)
	}
	f, err := model.OneByUUID(ctx, tx, true, uid)
	if err != nil {
		return fmt.Errorf(format, "one record by", uid, err)
	}
	got.updateValues(f)
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf(format, "infer", uid, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(format, "tx commit", uid, err)
	}
	return nil
}

func (got *PouetLink) updateValues(f *models.File) {
	if f == nil {
		return
	}
	if i := got.DemozooID; i > 0 {
		f.WebIDDemozoo = null.Int64From(int64(i))
	}
	if s := strings.TrimSpace(got.Releaser1); s != "" {
		f.GroupBrandFor = null.StringFrom(s)
	}
	if s := strings.TrimSpace(got.Releaser2); s != "" {
		f.GroupBrandBy = null.StringFrom(s)
	}
	if s := strings.TrimSpace(got.Title); s != "" {
		f.RecordTitle = null.StringFrom(s)
	}
	if i := got.IssuedDay; i > 0 {
		f.DateIssuedDay = null.Int16From(i)
	}
	if i := got.IssuedMonth; i > 0 {
		f.DateIssuedMonth = null.Int16From(i)
	}
	if i := got.IssuedYear; i > 0 {
		f.DateIssuedYear = null.Int16From(i)
	}
	if s := strings.TrimSpace(got.Filename); s != "" {
		f.Filename = null.StringFrom(s)
	}
	if i := int64(got.FileSize); i > 0 {
		f.Filesize = null.Int64From(i)
	}
	if s := strings.TrimSpace(got.FileType); s != "" {
		f.FileMagicType = null.StringFrom(s)
	}
	if s := strings.TrimSpace(got.FileHash); s != "" {
		f.FileIntegrityStrong = null.StringFrom(s)
	}
	if s := strings.TrimSpace(got.Content); s != "" {
		f.FileZipContent = null.StringFrom(s)
	}
	if s := strings.TrimSpace(got.Platform); s != "" {
		f.Platform = null.StringFrom(s)
	}
	if s := strings.TrimSpace(got.Section); s != "" {
		f.Section = null.StringFrom(s)
	}
}
