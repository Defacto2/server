//nolint:exhaustruct_v5,tagliatelle
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
	"time"

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
	download    dir.Directory
	prod        pouet.Production
	msg         string
	timeout     time.Duration
}

// Pouet initializes a new [PouetLink] for use with [Download].
//
// The prodID is the pouet production id to probe.
// The uuid is the local artifact record unique id to update.
// The download directory is where the fetched remote file download will be saved.
//
// The timeout is optional and is for the remote file download,
// but can usually be set to 0 to use the default 15 second value.
func Pouet(prodID int, unid string, download dir.Directory, timeout time.Duration) *PouetLink {
	got := PouetLink{}
	got.PouetID = prodID
	got.UUID = unid
	got.download = download
	got.timeout = timeout
	return &got
}

// Download fetches the download link from Pouet and saves it to the download directory.
// It then runs Update to modify the database record with various metadata from the file and Pouet record API data.
func (got *PouetLink) Download(ctx context.Context, sl *slog.Logger, c *echo.Context, tx *sql.Tx) error {
	const format = "%s for id %d: %w"
	if err := nils.Check(ctx, sl, c, tx); err != nil {
		return fmt.Errorf(format, "check", got.PouetID, err)
	}

	got.msg = "pouet link download"

	errStatus, err := got.prod.Get(ctx, got.PouetID)
	if err != nil {
		const s = " could not get record from pouet api"
		got.log(sl, s, err)
		return fmt.Errorf(format, s, got.PouetID, err)
	}

	if errStatus > 0 {
		s := "status was not okay: " + strconv.Itoa(errStatus)
		got.log(sl, s, ErrNoRecord)
		return fmt.Errorf(format, s, got.PouetID, ErrNoRecord)
	}

	const notUsable = "no usable download link found"

	if got.prod.Download == "" {
		got.log(sl, notUsable, nil)
		got.Error = notUsable
		return c.JSON(http.StatusNotModified, got)
	}

	if err := got.remoteDo(ctx, sl, c, tx); err != nil {
		return err
	}

	if got.FileSize <= 0 {
		got.Error = notUsable
	}

	return c.JSON(http.StatusNotModified, got)
}

// Stat sets the file size, hash, type, and archive content of the file.
// The UUID is used to locate the file in the download directory.
func (got *PouetLink) Stat(ctx context.Context, sl *slog.Logger, c *echo.Context, tx *sql.Tx) error {
	const format = "demozoo link stat file and integrity %s: %w"
	if err := nils.Check(ctx, sl, c, tx); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	name := filepath.Join(got.download.Path(), got.UUID)
	if got.FileSize == 0 {
		stat, err := os.Stat(name)
		if err != nil {
			return fmt.Errorf(format, "but could not stat file "+name, err)
		}

		got.FileSize = int(stat.Size())
	}

	strong, err := helper.StrongIntegrity(name)
	if err != nil {
		return fmt.Errorf(format, "but could not get the strong integrity hash "+name, err)
	}
	got.FileHash = strong

	if got.FileType == "" {
		got.FileType = simple.MagicAsTitle(sl, name)
	}

	return nil
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

func (got *PouetLink) log(sl *slog.Logger, s string, err error) {
	args := []any{slog.Int("id", got.PouetID)}
	if got.prod.Download != "" {
		args = append(args, slog.String("link_url", got.prod.Download))
	}
	if got.Filename != "" {
		args = append(args, slog.String("filename", got.Filename))
	}
	if err != nil {
		args = append(args, slog.Any("error", err))
	}
	sl.Info(got.msg+" "+s, args...)
}

func (got *PouetLink) remoteDo(ctx context.Context, sl *slog.Logger, c *echo.Context, tx *sql.Tx) error {
	got.Filename = filepath.Base(got.prod.Download)
	if id, err := strconv.Atoi(got.prod.Demozoo); err == nil && id > 0 {
		got.DemozooID = id
	}
	y, m, d := got.prod.Released()
	got.IssuedYear = y
	got.IssuedMonth = m
	got.IssuedDay = d
	r1, r2 := got.prod.Releasers()
	got.Releaser1 = r1
	got.Releaser2 = r2
	got.Title = got.prod.Title
	plat, sect := got.prod.PlatformType()
	got.Platform = plat.String()
	got.Section = sect.String()

	timeout := got.timeout
	if timeout == 0 {
		timeout = TimeoutLong
	}

	response, err := GetFile(ctx, sl, timeout, got.prod.Download)
	if err != nil {
		got.log(sl, "remote file issue: "+response.Path, err)
		if err1 := got.Update(ctx, c, tx); err1 != nil {
			got.log(sl, "download update", err1)
			return err1
		}
		return err
	} else if response == (Response{}) {
		got.log(sl, "download remote but it is empty", ErrBodyNil)
		return ErrBodyNil
	}

	dst := filepath.Join(got.download.Path(), got.UUID)
	if err := renameOW(response.Path, dst); err != nil {
		got.log(sl, "downloaded rename", err)
		return err
	}
	cl := response.ContentLength
	if size, err := strconv.Atoi(cl); err != nil {
		got.log(sl, "downloaded content length", err)
	} else {
		got.FileSize = size
	}
	if err := got.Stat(ctx, sl, c, tx); err != nil {
		got.log(sl, "download stat", err)
		return nil
	}

	files, err := archive.Lists(ctx, dst)
	if err != nil {
		got.log(sl, "archive lists issue", err)
		return nil
	}

	got.Content = strings.Join(files, "\n")
	if err := got.Update(ctx, c, tx); err != nil {
		got.log(sl, "archive lists update", err)
		return err
	}

	return c.HTML(http.StatusOK, `<p class="text-success">Successful Pouet update</p>`)
}
