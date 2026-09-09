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
	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/labstack/echo/v5"
)

// DemozooLink is the response from the task of GetDemozooFile.
type DemozooLink struct {
	UUID        string   `json:"uuid"`          // UUID is the file production UUID.
	Github      string   `json:"github_repo"`   // GitHub is the GitHub repository URI.
	YouTube     string   `json:"youtube_video"` // YouTube is the YouTube watch video URI.
	Releaser1   string   `json:"releaser1"`     // Releaser1 is the first releaser of the file.
	Releaser2   string   `json:"releaser2"`     // Releaser2 is the second releaser of the file.
	Title       string   `json:"title"`         // Title is the file title.
	Filename    string   `json:"filename"`      // Filename is the file name of the download.
	Content     string   `json:"content"`       // Content is the file archive content.
	FileType    string   `json:"file_type"`     // Type is the file type.
	FileHash    string   `json:"file_hash"`     // Hash is the file integrity hash.
	Platform    string   `json:"platform"`      // Platform is the file platform.
	Section     string   `json:"section"`       // Section is the file section.
	Error       string   `json:"error"`         // Error is the error message if the download or record update failed.
	CreditText  []string `json:"credit_text"`   // credit_text, writer
	CreditCode  []string `json:"credit_code"`   // credit_program, programmer/coder
	CreditArt   []string `json:"credit_art"`    // credit_illustration, artist/graphics
	CreditAudio []string `json:"credit_audio"`  // credit_audio, musician/sound
	ID          int      `json:"id"`            // ID is the Demozoo production ID.
	Pouet       int      `json:"pouet_prod"`    // Pouet is the Pouet production ID.
	FileSize    int      `json:"file_size"`     // Size is the file size in bytes.
	IssuedYear  int16    `json:"issued_year"`   // Year is the year the file was issued.
	IssuedMonth int16    `json:"issued_month"`  // Month is the month the file was issued.
	IssuedDay   int16    `json:"issued_day"`    // Day is the day the file was issued.
	download    dir.Directory
	prod        demozoo.Production
	linkURL     string
	msg         string
	timeout     time.Duration
}

// Demozoo initializes a new [DemozooLink] for use with [Download].
//
// The prodID is the demozoo production id to probe.
// The uuid is the local artifact record unique id to update.
// The download directory is where the fetched remote file download will be saved.
//
// The timeout is optional and is for the remote file downloads.
// It can usually be set to 0 to use the default 15 second value for single links,
// and 8 seconds per-link for productions with multiple downloads.
func Demozoo(prodID int, unid string, download dir.Directory, timeout time.Duration) *DemozooLink {
	got := DemozooLink{}
	got.ID = prodID
	got.UUID = unid
	got.download = download
	got.timeout = timeout
	return &got
}

// Download fetches the download link from Demozoo and saves it to the download directory.
// It then runs Update to modify the database record with various metadata from the file and Demozoo record API data.
func (got *DemozooLink) Download(ctx context.Context, sl *slog.Logger, c *echo.Context, tx *sql.Tx) error {
	const format = "%s for id %d: %w"
	if err := nils.Check(ctx, sl, c, tx); err != nil {
		return fmt.Errorf(format, "check", 0, err)
	}

	got.msg = "demozoo link download"

	errStatus, err := got.prod.Get(ctx, got.ID)
	if err != nil {
		const s = " could not get record from demozoo api"
		got.log(sl, s, err)
		return fmt.Errorf(format, s, got.ID, err)
	}

	if errStatus > 0 {
		s := "status was not okay: " + strconv.Itoa(errStatus)
		got.log(sl, s, ErrNoRecord)
		return fmt.Errorf(format, s, got.ID, ErrNoRecord)
	}

	if err := got.remoteDo(ctx, sl, c, tx); err != nil {
		return err
	}

	if got.FileSize <= 0 {
		got.Error = "no usable download links found, they all returned a 404 error or were empty"
	}

	return c.JSON(http.StatusNotModified, got)
}

// Stat sets the file size, hash, type, and archive content of the file.
// The UUID is used to locate the file in the download directory.
func (got *DemozooLink) Stat(ctx context.Context, sl *slog.Logger, c *echo.Context, tx *sql.Tx) error {
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

	return got.ArchiveDo(ctx, sl, c, tx, name)
}

// ArchiveDo sets the archive content and readme text of the source file.
func (got *DemozooLink) ArchiveDo(
	ctx context.Context, sl *slog.Logger, c *echo.Context, tx *sql.Tx, src string,
) error {
	const format = "archive do: %w"
	if err := nils.Check(ctx, sl, c, tx); err != nil {
		return fmt.Errorf(format, err)
	}

	files, err := archive.Lists(ctx, src)
	if err != nil {
		got.log(sl, "archive lists issue", err)
		return nil
	}

	got.Content = strings.Join(files, "\n")
	if err := got.Update(ctx, c, tx); err != nil {
		got.log(sl, "archive lists update", err)
		return err
	}

	return c.HTML(http.StatusOK, `<p class="text-success">Successful Demozoo update</p>`)
}

// Update modifies the database record using data provided by the DemozooLink struct.
// A JSON response is returned with the success status of the update.
func (got *DemozooLink) Update(ctx context.Context, c *echo.Context, tx *sql.Tx) error {
	const format = "demozoo link update %s uuid %s: %w"
	if err := nils.Check(ctx, c, tx); err != nil {
		return fmt.Errorf(format, "check", "n/a", err)
	}

	uid := got.UUID
	f, err := model.OneByUUID(ctx, tx, true, uid)
	if err != nil {
		return fmt.Errorf(format, "one record by", uid, err)
	}

	got.setFile(f)
	got.setSceners(f)
	got.setMisc(f)

	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf(format, "infer", uid, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(format, "tx commit", uid, err)
	}

	return nil
}

func (got *DemozooLink) log(sl *slog.Logger, s string, err error) {
	args := []any{slog.Int("id", got.ID)}
	if got.linkURL != "" {
		args = append(args, slog.String("link_url", got.linkURL))
	}
	if got.Filename != "" {
		args = append(args, slog.String("filename", got.Filename))
	}
	if err != nil {
		args = append(args, slog.Any("error", err))
	}
	sl.Info(got.msg+" "+s, args...)
}

func (got *DemozooLink) remoteDo(ctx context.Context, sl *slog.Logger, c *echo.Context, tx *sql.Tx) error {
	// Originally we would return an error and abort the database record update if the download link
	// could not be fetched. However, this happens too frequently with some of the more popular Scene
	// websites. Sites such as scene.org frequently time out requests. So, as of 20-Jul-25, the
	// behavior now updates the record even when the download fails.

	for i, link := range got.prod.DownloadLinks {
		got.linkURL = link.URL
		if link.URL == "" {
			got.log(sl, "link url is empty", nil)
			continue
		}
		// append demozoo record metadata
		got.Filename = filepath.Base(link.URL)
		got.Github = got.prod.GithubRepo()
		got.Pouet = got.prod.PouetProd()
		got.YouTube = got.prod.YouTubeVideo()
		y, m, d := got.prod.Released()
		got.IssuedYear = y
		got.IssuedMonth = m
		got.IssuedDay = d
		r1, r2 := got.prod.Groups()
		got.Releaser1 = r1
		got.Releaser2 = r2
		got.Title = got.prod.Title
		ctext, ccode, cart, caudio := got.prod.Releasers()
		got.CreditText = ctext
		got.CreditCode = ccode
		got.CreditArt = cart
		got.CreditAudio = caudio
		plat, sect := got.prod.SuperType()
		got.Platform = plat.String()
		got.Section = sect.String()

		// attempt to download the remote file
		// if this task fails, update the record with the demozoo metadata
		response, err := got.remote(ctx, sl, i)
		if err != nil {
			got.log(sl, "download remote", err)
			if err1 := got.Update(ctx, c, tx); err1 != nil {
				got.log(sl, "download update", err1)
				return err1
			}
			return err
		} else if response == (Response{}) {
			got.log(sl, "download remote but it is empty", nil)
			continue
		}
		// assuming the download link was successful in being fetched,
		// we now obtain the file's metadata and incorporate those into the database record.
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

		got.Error = ""
		if err := got.Stat(ctx, sl, c, tx); err != nil {
			got.log(sl, "download stat", err)
		}

		return nil
	}

	return nil
}

// remote fetches the download link from Demozoo and saves it to the download directory.
// If the DownloadResponse is empty due to a production without a download link or a timeout,
// then it should be handled as a continue in the calling function.
func (got *DemozooLink) remote(ctx context.Context, sl *slog.Logger, index int) (Response, error) {
	const format = "cannot get the remote file from %s: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return Response{}, fmt.Errorf("get remove file check: %w", err)
	}

	timeout := got.timeout
	count := len(got.prod.DownloadLinks)
	switch {
	case timeout == 0 && count == 1:
		timeout = TimeoutLong
	case timeout == 0 && count > 1:
		// keep the timeout short, and move onto the next possible link
		timeout = TimeoutShort
	}

	resp, err := GetFile(ctx, sl, timeout, got.linkURL)
	if skip := err != nil || resp.Path == ""; skip {
		got.log(sl, "remote file issue: "+resp.Path, err)

		// If the last link failed then return the error, otherwise this will fail silently.
		if lastLink := index+1 >= len(got.prod.DownloadLinks); lastLink {
			return Response{}, fmt.Errorf(format, "any linked url "+got.linkURL, err)
		}

		return Response{}, nil
	}

	return resp, nil
}

func (got *DemozooLink) setFile(f *models.File) {
	if f == nil {
		return
	}

	if s := strings.TrimSpace(got.Platform); s != "" {
		f.Platform = null.StringFrom(s)
	}
	if s := strings.TrimSpace(got.Section); s != "" {
		f.Section = null.StringFrom(s)
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
}

func (got *DemozooLink) setSceners(f *models.File) {
	if f == nil {
		return
	}

	if s := strings.Join(got.CreditAudio, ","); s != "" {
		f.CreditAudio = null.StringFrom(s)
	}
	if s := strings.Join(got.CreditArt, ","); s != "" {
		f.CreditIllustration = null.StringFrom(s)
	}
	if s := strings.Join(got.CreditCode, ","); s != "" {
		f.CreditProgram = null.StringFrom(s)
	}
	if s := strings.Join(got.CreditText, ","); s != "" {
		f.CreditText = null.StringFrom(s)
	}
}

func (got *DemozooLink) setMisc(f *models.File) {
	if f == nil {
		return
	}

	if s := strings.TrimSpace(got.Github); s != "" {
		f.WebIDGithub = null.StringFrom(s)
	}
	if s := strings.TrimSpace(got.YouTube); s != "" {
		f.WebIDYoutube = null.StringFrom(s)
	}
	if i := int64(got.Pouet); i > 0 {
		f.WebIDPouet = null.Int64From(i)
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
}
