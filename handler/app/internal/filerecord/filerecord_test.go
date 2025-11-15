package filerecord_test

import (
	"log/slog"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/server/handler/app/internal/filerecord"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/aarondl/null/v8"
	"github.com/nalgeon/be"
)

const r0 = "00000000-0000-0000-0000-000000000000"

func TestWebsites(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.Websites(&x)
	be.Equal(t, s, "")

	x.ListLinks = null.StringFrom("placeholder text")
	s = filerecord.Websites(&x)
	be.Equal(t, s, "")

	ex := "http://example.com"
	x.ListLinks = null.StringFrom(ex)
	s = filerecord.Websites(&x)
	be.Equal(t, s, "")

	x.ListLinks = null.StringFrom("Example page;http://example.com")
	s = filerecord.Websites(&x)
	find := strings.Contains(string(s), `href="http://example.com">Example page`)
	be.True(t, find)

	x.ListLinks = null.StringFrom("http://example.com;Example page")
	s = filerecord.Websites(&x)
	be.Equal(t, s, "")
}

func TestEmbedReadme(t *testing.T) {
	t.Parallel()
	x := models.File{}
	b := filerecord.EmbedReadme(&x)
	be.True(t, b)
	x.Filename = null.StringFrom("filename.txt")
	b = filerecord.EmbedReadme(&x)
	be.True(t, b)
	x.Filename = null.StringFrom("filename.rip")
	b = filerecord.EmbedReadme(&x)
	be.True(t, !b)
	x.Filename = null.StringFrom("filename.pdf")
	x.Platform = null.StringFrom("pdf")
	b = filerecord.EmbedReadme(&x)
	be.True(t, !b)
}

func TestRelations(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.Relations(&x)
	be.Equal(t, s, "")

	x.ListRelations = null.StringFrom("placeholder text")
	s = filerecord.Relations(&x)
	be.Equal(t, s, "")

	const id = "9b1c6"

	x.ListRelations = null.StringFrom(id)
	s = filerecord.Relations(&x)
	be.Equal(t, s, "")

	x.ListRelations = null.StringFrom("Info text;9b1c6")
	s = filerecord.Relations(&x)
	find := strings.Contains(string(s), `href="/f/9b1c6">Info text</a>`)
	be.True(t, find)

	x.ListRelations = null.StringFrom("9b1c6;Info text")
	s = filerecord.Relations(&x)
	be.Equal(t, s, "")
}

func TestRecordStatus(t *testing.T) {
	t.Parallel()

	x := models.File{}
	a := filerecord.RecordIsNew(&x)
	be.True(t, !a)
	b := filerecord.RecordOffline(&x)
	be.True(t, !b)
	c := filerecord.RecordOnline(&x)
	be.True(t, c)

	now := time.Now()
	x.Deletedat = null.TimeFrom(now)
	a = filerecord.RecordIsNew(&x)
	be.True(t, a)
	b = filerecord.RecordOffline(&x)
	be.True(t, !b)
	c = filerecord.RecordOnline(&x)
	be.True(t, !c)

	x.Deletedby = null.StringFrom("an operator")
	a = filerecord.RecordIsNew(&x)
	be.True(t, !a)
	b = filerecord.RecordOffline(&x)
	be.True(t, b)
	c = filerecord.RecordOnline(&x)
	be.True(t, !c)
}

func TestRecordProblems(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.RecordProblems(&x)
	errs := strings.Split(s, "+")
	be.True(t, len(errs) == 4)
}

func TestReadme(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.Readme(&x)
	be.Equal(t, s, "")

	files := "1.txt\n2.txt\nmy group.txt\n3.txt\n4.txt"
	x.Filename = null.StringFrom("filename.zip")
	x.GroupBrandBy = null.StringFrom("my group")
	x.FileZipContent = null.StringFrom(files)
	s = filerecord.Readme(&x)
	find := strings.Contains(s, "1.txt")
	be.True(t, find)
}

func TestLinkPreviewTip(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.LinkPreviewTip(&x)
	be.Equal(t, s, "")

	x.Filename = null.StringFrom("filename.txt")
	x.Platform = null.StringFrom("text")
	s = filerecord.LinkPreviewTip(&x)
	be.Equal(t, "Read this as text", s)
}

func TestLinkPreview(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.LinkPreview(&x)
	be.Equal(t, s, "")

	x.UUID = null.StringFrom(r0)
	s = filerecord.LinkPreview(&x)
	be.Equal(t, s, "")

	x.ID = 1
	x.Filename = null.StringFrom("filename.txt")
	x.Platform = null.StringFrom("dos")
	s = filerecord.LinkPreview(&x)
	be.Equal(t, "/v/9b1c6", s)
}

func TestLastModified(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.LastModification(&x)
	be.Equal(t, "no timestamp", s)

	t1970 := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	x.FileLastModified = null.TimeFrom(t1970)
	s = filerecord.LastModification(&x)
	be.Equal(t, "no timestamp", s)

	t1985 := time.Date(1985, time.January, 1, 0, 0, 0, 0, time.UTC)
	x.FileLastModified = null.TimeFrom(t1985)
	s = filerecord.LastModification(&x)
	be.Equal(t, "1985 Jan 1, 00:00", s)

	s = filerecord.LastModificationDate(&x)
	be.Equal(t, "1985-Jan-01", s)

	a, b, c := filerecord.LastModifications(&x)
	be.Equal(t, 1985, a)
	be.Equal(t, 1, b)
	be.Equal(t, 1, c)

	s = filerecord.LastModificationAgo(&x)
	find := strings.Contains(s, "years ago")
	be.True(t, find)
}

func TestJsdos(t *testing.T) {
	t.Parallel()
	pl, zip, exe := "dos", "FILENAME.ZIP", "FILENAME.EXE"
	x := models.File{
		Platform: null.StringFrom(pl),
		Filename: null.StringFrom(zip),
	}
	ok := filerecord.JsdosArchive(&x)
	be.True(t, ok)
	ok = filerecord.JsdosUse(&x)
	be.True(t, ok)

	x.Filename = null.StringFrom(exe)
	ok = filerecord.JsdosUse(&x)
	be.True(t, ok)

	ok = filerecord.JsdosUsage(zip, pl)
	be.True(t, ok)
	ok = filerecord.JsdosUsage(exe, pl)
	be.True(t, ok)
}

func TestFirstHeader(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.FirstHeader(&x)
	be.Equal(t, s, "")

	x.RecordTitle = null.StringFrom("Hello world")
	s = filerecord.FirstHeader(&x)
	be.Equal(t, "Hello world", s)

	x.RecordTitle = null.StringFrom("5")
	x.Section = null.StringFrom("magazine")
	s = filerecord.FirstHeader(&x)
	be.Equal(t, "Issue 5", s)
}

func TestFileEntry(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.FileEntry(&x)
	be.Equal(t, s, "")

	now := time.Now()
	yearAgo := now.AddDate(-1, 0, 0)

	x.Createdat = null.TimeFrom(yearAgo)
	s = filerecord.FileEntry(&x)
	find := strings.Contains(s, "Created about 1 year ago")
	be.True(t, find)

	x.Updatedat = null.TimeFrom(now)
	s = filerecord.FileEntry(&x)
	find = strings.Contains(s, "Updated just now")
	be.True(t, find)
}

func TestExtraZip(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.ExtraZip(&x, "")
	be.True(t, !s)
	x.UUID = null.StringFrom(r0)
	s = filerecord.ExtraZip(&x, "")
	be.True(t, !s)

	extra := dir.Directory(t.TempDir())
	err := command.CopyFile(nil,
		filepath.Join("testdata", "archive.zip"),
		filepath.Join(extra.Path(), r0+".zip"))
	be.Err(t, err)
}

func TestDownloadID(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.DownloadID(&x)
	be.Equal(t, "0", s)
	x.ID = 1
	s = filerecord.DownloadID(&x)
	be.Equal(t, "9b1c6", s)
}

func TestDescription(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.Description(&x)
	be.Equal(t, s, " released by .")

	x.Filename = null.StringFrom("myfile.txt")
	x.GroupBrandBy = null.StringFrom("my group")
	s = filerecord.Description(&x)
	be.Equal(t, s, "myfile.txt released by My Group.")
}

func TestDate(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.Date(&x)
	find := strings.Contains(string(s), "unknown date")
	be.True(t, find)
	x.DateIssuedYear = null.Int16From(2021)
	s = filerecord.Date(&x)
	find = strings.Contains(string(s), "2021")
	be.True(t, find)
	x.DateIssuedMonth = null.Int16From(1)
	s = filerecord.Date(&x)
	find = strings.Contains(string(s), "January")
	be.True(t, find)
	x.DateIssuedDay = null.Int16From(1)
	s = filerecord.Date(&x)

	a, b, c := filerecord.Dates(&x)
	be.Equal(t, 2021, a)
	be.Equal(t, 1, b)
	be.Equal(t, 1, c)

	find = strings.Contains(string(s), "January 1")
	be.True(t, find)
	x.DateIssuedYear = null.Int16From(30000)
	s = filerecord.Date(&x)
	find = strings.Contains(string(s), "30000")
	be.True(t, !find)
}

func TestListContent(t *testing.T) {
	x := models.File{}
	dirs := command.Dirs{}
	sl := slog.Default()
	s := filerecord.ListContent(sl, -1, &x, dirs, "")
	find := strings.Contains(string(s), "no UUID")
	be.True(t, find)

	x.UUID = null.StringFrom(r0)
	s = filerecord.ListContent(sl, -1, &x, dirs, "")
	find = strings.Contains(string(s), "invalid platform")
	be.True(t, find)

	x.Platform = null.StringFrom("dos")
	s = filerecord.ListContent(sl, -1, &x, dirs, "")
	find = strings.Contains(string(s), "cannot stat file")
	be.True(t, find)

	src, err := filepath.Abs("testdata")
	be.Err(t, err, nil)
	s = filerecord.ListContent(sl, -1, &x, dirs, src)
	find = strings.Contains(string(s), "error, ")
	be.True(t, find)

	tmpDir := t.TempDir()
	err = command.CopyFile(logs.Discard(), filepath.Join("testdata", "archive.zip"), filepath.Join(tmpDir, "archive.zip"))
	be.Err(t, err, nil)
	s = filerecord.ListContent(sl, -1, &x, dirs, tmpDir)
	find = strings.Contains(string(s), "error, ")
	be.True(t, find)
}

func TestAlertURL(t *testing.T) {
	x := models.File{}
	s := filerecord.AlertURL(&x)
	be.Equal(t, s, "")
	x.FileSecurityAlertURL = null.StringFrom("invalid")
	s = filerecord.AlertURL(&x)
	be.Equal(t, s, "")
	x.FileSecurityAlertURL = null.StringFrom("https://example.com")
	s = filerecord.AlertURL(&x)
	be.Equal(t, "https://example.com", s)
}

func TestListEntry(t *testing.T) {
	t.Parallel()
	le := filerecord.ListEntry{}
	s := le.HTML(1, "x", "y")
	find := strings.Contains(s, "1 bytes")
	be.True(t, find)
}

func TestLinkPreviewHref(t *testing.T) {
	t.Parallel()
	s := filerecord.LinkPreviewHref(nil, "", "")
	be.Equal(t, s, "")
	s = filerecord.LinkPreviewHref(1, "filename.xxx", "invalid")
	be.Equal(t, s, "")
	s = filerecord.LinkPreviewHref(1, "filename.txt", "text")
	be.Equal(t, "/v/9b1c6", s)
}

func TestLegacyString(t *testing.T) {
	t.Parallel()
	s := filerecord.LegacyString("")
	be.Equal(t, s, "")
	s = filerecord.LegacyString("Hello world 123.")
	be.Equal(t, "Hello world 123.", s)
	s = filerecord.LegacyString("£100")
	be.Equal(t, "£100", s)
	s = filerecord.LegacyString("\xa3100")
	be.Equal(t, "£100", s)
	s = filerecord.LegacyString("€100")
	be.Equal(t, "€100", s)
	s = filerecord.LegacyString("\x80100")
	be.Equal(t, "€100", s)
}
