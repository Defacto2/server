package filerecord_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/server/handler/app/internal/filerecord"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
)

const r0 = "00000000-0000-0000-0000-000000000000"

func TestWebsites(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.Websites(&x)
	assert.Empty(t, s)

	x.ListLinks = null.StringFrom("placeholder text")
	s = filerecord.Websites(&x)
	assert.Empty(t, s)

	ex := "http://example.com"
	x.ListLinks = null.StringFrom(ex)
	s = filerecord.Websites(&x)
	assert.Empty(t, s)

	x.ListLinks = null.StringFrom("Example page;http://example.com")
	s = filerecord.Websites(&x)
	assert.Contains(t, s, `href="http://example.com">Example page`)

	x.ListLinks = null.StringFrom("http://example.com;Example page")
	s = filerecord.Websites(&x)
	assert.Empty(t, s)
}

func TestEmbedReadme(t *testing.T) {
	t.Parallel()
	x := models.File{}
	b := filerecord.EmbedReadme(&x)
	assert.True(t, b)
	x.Filename = null.StringFrom("filename.txt")
	b = filerecord.EmbedReadme(&x)
	assert.True(t, b)
	x.Filename = null.StringFrom("filename.rip")
	b = filerecord.EmbedReadme(&x)
	assert.False(t, b)
	x.Filename = null.StringFrom("filename.pdf")
	x.Platform = null.StringFrom("pdf")
	b = filerecord.EmbedReadme(&x)
	assert.False(t, b)
}

func TestRelations(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.Relations(&x)
	assert.Empty(t, s)

	x.ListRelations = null.StringFrom("placeholder text")
	s = filerecord.Relations(&x)
	assert.Empty(t, s)

	const id = "9b1c6"

	x.ListRelations = null.StringFrom(id)
	s = filerecord.Relations(&x)
	assert.Empty(t, s)

	x.ListRelations = null.StringFrom("Info text;9b1c6")
	s = filerecord.Relations(&x)
	assert.Contains(t, s, `href="/f/9b1c6">Info text</a>`)

	x.ListRelations = null.StringFrom("9b1c6;Info text")
	s = filerecord.Relations(&x)
	assert.Empty(t, s)
}

func TestRecordStatus(t *testing.T) {
	t.Parallel()

	x := models.File{}
	a := filerecord.RecordIsNew(&x)
	assert.False(t, a)
	b := filerecord.RecordOffline(&x)
	assert.False(t, b)
	c := filerecord.RecordOnline(&x)
	assert.True(t, c)

	now := time.Now()
	x.Deletedat = null.TimeFrom(now)
	a = filerecord.RecordIsNew(&x)
	assert.True(t, a)
	b = filerecord.RecordOffline(&x)
	assert.False(t, b)
	c = filerecord.RecordOnline(&x)
	assert.False(t, c)

	x.Deletedby = null.StringFrom("an operator")
	a = filerecord.RecordIsNew(&x)
	assert.False(t, a)
	b = filerecord.RecordOffline(&x)
	assert.True(t, b)
	c = filerecord.RecordOnline(&x)
	assert.False(t, c)
}

func TestRecordProblems(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.RecordProblems(&x)
	errs := strings.Split(s, "+")
	assert.Len(t, errs, 4)
}

func TestReadme(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.Readme(&x)
	assert.Empty(t, s)

	files := "1.txt\n2.txt\nmy group.txt\n3.txt\n4.txt"
	x.Filename = null.StringFrom("filename.zip")
	x.GroupBrandBy = null.StringFrom("my group")
	x.FileZipContent = null.StringFrom(files)
	s = filerecord.Readme(&x)
	assert.Contains(t, s, "1.txt")
}

func TestLinkPreviewTip(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.LinkPreviewTip(&x)
	assert.Empty(t, s)

	x.Filename = null.StringFrom("filename.txt")
	x.Platform = null.StringFrom("text")
	s = filerecord.LinkPreviewTip(&x)
	assert.Equal(t, "Read this as text", s)
}

func TestLinkPreview(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.LinkPreview(&x)
	assert.Empty(t, s)

	x.UUID = null.StringFrom(r0)
	s = filerecord.LinkPreview(&x)
	assert.Empty(t, s)

	x.ID = 1
	x.Filename = null.StringFrom("filename.txt")
	x.Platform = null.StringFrom("dos")
	s = filerecord.LinkPreview(&x)
	assert.Equal(t, "/v/9b1c6", s)
}

func TestLastModified(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.LastModification(&x)
	assert.Equal(t, "no timestamp", s)

	t1970 := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	x.FileLastModified = null.TimeFrom(t1970)
	s = filerecord.LastModification(&x)
	assert.Equal(t, "no timestamp", s)

	t1985 := time.Date(1985, time.January, 1, 0, 0, 0, 0, time.UTC)
	x.FileLastModified = null.TimeFrom(t1985)
	s = filerecord.LastModification(&x)
	assert.Equal(t, "1985 Jan 1, 00:00", s)

	s = filerecord.LastModificationDate(&x)
	assert.Equal(t, "1985-Jan-01", s)

	a, b, c := filerecord.LastModifications(&x)
	assert.Equal(t, 1985, a)
	assert.Equal(t, 1, b)
	assert.Equal(t, 1, c)

	s = filerecord.LastModificationAgo(&x)
	assert.Contains(t, s, "years ago")
}

func TestJsdos(t *testing.T) {
	t.Parallel()
	pl, zip, exe := "dos", "FILENAME.ZIP", "FILENAME.EXE"
	x := models.File{
		Platform: null.StringFrom(pl),
		Filename: null.StringFrom(zip),
	}
	ok := filerecord.JsdosArchive(&x)
	assert.True(t, ok)
	ok = filerecord.JsdosUse(&x)
	assert.True(t, ok)

	x.Filename = null.StringFrom(exe)
	ok = filerecord.JsdosUse(&x)
	assert.True(t, ok)

	ok = filerecord.JsdosUsage(zip, pl)
	assert.True(t, ok)
	ok = filerecord.JsdosUsage(exe, pl)
	assert.True(t, ok)
}

func TestFirstHeader(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.FirstHeader(&x)
	assert.Empty(t, s)

	x.RecordTitle = null.StringFrom("Hello world")
	s = filerecord.FirstHeader(&x)
	assert.Equal(t, "Hello world", s)

	x.RecordTitle = null.StringFrom("5")
	x.Section = null.StringFrom("magazine")
	s = filerecord.FirstHeader(&x)
	assert.Equal(t, "Issue 5", s)
}

func TestFileEntry(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.FileEntry(&x)
	assert.Empty(t, s)

	now := time.Now()
	yearAgo := now.AddDate(-1, 0, 0)

	x.Createdat = null.TimeFrom(yearAgo)
	s = filerecord.FileEntry(&x)
	assert.Contains(t, s, "Created about 1 year ago")

	x.Updatedat = null.TimeFrom(now)
	s = filerecord.FileEntry(&x)
	assert.Contains(t, s, "Updated just now")
}

func TestExtraZip(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.ExtraZip(&x, "")
	assert.Empty(t, s)
	x.UUID = null.StringFrom(r0)
	s = filerecord.ExtraZip(&x, "")
	assert.Empty(t, s)

	extra := dir.Directory(t.TempDir())
	err := command.CopyFile(nil,
		filepath.Join("testdata", "archive.zip"),
		filepath.Join(extra.Path(), r0+".zip"))
	require.NoError(t, err)

	ok := filerecord.ExtraZip(&x, extra)
	assert.True(t, ok)
}

func TestDownloadID(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.DownloadID(&x)
	assert.Equal(t, "0", s)
	x.ID = 1
	s = filerecord.DownloadID(&x)
	assert.Equal(t, "9b1c6", s)
}

func TestDescription(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.Description(&x)
	assert.Equal(t, " released by ", s)

	x.Filename = null.StringFrom("myfile.txt")
	x.GroupBrandBy = null.StringFrom("my group")
	s = filerecord.Description(&x)
	assert.Equal(t, "myfile.txt released by My Group", s)
}

func TestDate(t *testing.T) {
	t.Parallel()
	x := models.File{}
	s := filerecord.Date(&x)
	assert.Contains(t, s, "unknown date")
	x.DateIssuedYear = null.Int16From(2021)
	s = filerecord.Date(&x)
	assert.Contains(t, s, "2021")
	x.DateIssuedMonth = null.Int16From(1)
	s = filerecord.Date(&x)
	assert.Contains(t, s, "January")
	x.DateIssuedDay = null.Int16From(1)
	s = filerecord.Date(&x)

	a, b, c := filerecord.Dates(&x)
	assert.EqualValues(t, 2021, a)
	assert.EqualValues(t, 1, b)
	assert.EqualValues(t, 1, c)

	assert.Contains(t, s, "January 1")
	x.DateIssuedYear = null.Int16From(30000)
	s = filerecord.Date(&x)
	assert.NotContains(t, s, "30000")
}

func TestListContent(t *testing.T) {
	x := models.File{}
	dirs := command.Dirs{}
	s := filerecord.ListContent(&x, dirs, "")
	assert.Contains(t, s, "no UUID")

	x.UUID = null.StringFrom(r0)
	s = filerecord.ListContent(&x, dirs, "")
	assert.Contains(t, s, "invalid platform")

	x.Platform = null.StringFrom("dos")
	s = filerecord.ListContent(&x, dirs, "")
	assert.Contains(t, s, "cannot stat file")

	src, err := filepath.Abs("testdata")
	require.NoError(t, err)
	s = filerecord.ListContent(&x, dirs, src)
	assert.Contains(t, s, "error, ")

	tmpDir := t.TempDir()
	err = command.CopyFile(nil, filepath.Join("testdata", "archive.zip"), filepath.Join(tmpDir, "archive.zip"))
	require.NoError(t, err)
	src = filepath.Join(tmpDir, "archive.zip")
	s = filerecord.ListContent(&x, dirs, tmpDir)
	assert.Contains(t, s, "error, ")

	s = filerecord.ListContent(&x, dirs, src)
	assert.Contains(t, s, "FILE_ID.DIZ")
	assert.Contains(t, s, "/editor/readme/preview/"+r0+"/FILE_ID.DIZ")
	assert.Contains(t, s, "/editor/diz/copy/"+r0+"/FILE_ID.DIZ")

	diz := r0 + ".diz"
	txt := r0 + ".txt"
	webp := r0 + ".webp"

	assert.FileExists(t, diz)
	assert.FileExists(t, txt)
	assert.FileExists(t, webp)

	defer func() {
		// Cleanup assets that are generated by ListContent.
		_ = os.Remove(diz)
		_ = os.Remove(txt)
		_ = os.Remove(webp)
	}()
}

func TestAlertURL(t *testing.T) {
	x := models.File{}
	s := filerecord.AlertURL(&x)
	assert.Empty(t, s)
	x.FileSecurityAlertURL = null.StringFrom("invalid")
	s = filerecord.AlertURL(&x)
	assert.Empty(t, s)
	x.FileSecurityAlertURL = null.StringFrom("https://example.com")
	s = filerecord.AlertURL(&x)
	assert.Equal(t, "https://example.com", s)
}

func TestListEntry(t *testing.T) {
	t.Parallel()
	le := filerecord.ListEntry{}
	assert.Empty(t, le)
	s := le.HTML(1, "x", "y")
	assert.Contains(t, s, "1 bytes")
}

func TestLinkPreviewHref(t *testing.T) {
	t.Parallel()
	s := filerecord.LinkPreviewHref(nil, "", "")
	assert.Empty(t, s)
	s = filerecord.LinkPreviewHref(1, "filename.xxx", "invalid")
	assert.Empty(t, s)
	s = filerecord.LinkPreviewHref(1, "filename.txt", "text")
	assert.Equal(t, "/v/9b1c6", s)
}

func TestLegacyString(t *testing.T) {
	t.Parallel()
	s := filerecord.LegacyString("")
	assert.Empty(t, s)
	s = filerecord.LegacyString("Hello world 123.")
	assert.Equal(t, "Hello world 123.", s)
	s = filerecord.LegacyString("£100")
	assert.Equal(t, "£100", s)
	s = filerecord.LegacyString("\xa3100")
	assert.Equal(t, "£100", s)
	s = filerecord.LegacyString("€100")
	assert.Equal(t, "€100", s)
	s = filerecord.LegacyString("\x80100")
	assert.Equal(t, "€100", s)
}
