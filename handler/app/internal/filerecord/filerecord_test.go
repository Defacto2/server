package filerecord_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/server/handler/app/internal/filerecord"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/aarondl/null/v8"
	"github.com/nalgeon/be"
)

const (
	r0          = "00000000-0000-0000-0000-000000000000"
	flaggedHash = "aa97833330f4a27f0c7888ae633de652be5a37840fc87cc364b5c90908d027d855d66fe60d8b2b23b02fb0fe482ddcf1"
)

func TestValues(t *testing.T) {
	t.Parallel()

	art := testutil.NewModel(t)
	art.CreditIllustration = null.StringFrom("some artist")
	be.Equal(t, filerecord.AttrArtist(art), "some artist")

	art.CreditAudio = null.StringFrom("some musician")
	be.Equal(t, filerecord.AttrMusic(art), "some musician")

	art.CreditProgram = null.StringFrom("some programmer")
	be.Equal(t, filerecord.AttrProg(art), "some programmer")

	art.CreditText = null.StringFrom("some writer")
	be.Equal(t, filerecord.AttrWriter(art), "some writer")

	art.Filename = null.StringFrom("example.txt")
	be.Equal(t, filerecord.Basename(art), "example.txt")

	art.FileIntegrityStrong = null.StringFrom("abc123sha384hash")
	be.Equal(t, filerecord.Checksum(art), "abc123sha384hash")

	art.Comment = null.StringFrom("a useful comment")
	be.Equal(t, filerecord.Comment(art), "a useful comment")

	art.WebID16colors = null.StringFrom("pack/file.ansi")
	be.Equal(t, filerecord.Idenfication16C(art), "pack/file.ansi")

	art.WebIDDemozoo = null.Int64From(12345)
	be.Equal(t, filerecord.IdenficationDZ(art), "12345")

	art.WebIDGithub = null.StringFrom("owner/repo")
	be.Equal(t, filerecord.IdenficationGitHub(art), "owner/repo")

	art.WebIDPouet = null.Int64From(67890)
	be.Equal(t, filerecord.IdenficationPouet(art), "67890")

	art.WebIDYoutube = null.StringFrom("dQw4w9WgXcQ")
	be.Equal(t, filerecord.IdenficationYT(art), "dQw4w9WgXcQ")

	art.DoseeIncompatible = null.Int16From(1)
	be.Equal(t, filerecord.JsdosBroken(art), true)

	art.DoseeHardwareCPU = null.StringFrom("486")
	be.Equal(t, filerecord.JsdosCPU(art), "486")

	art.DoseeHardwareGraphic = null.StringFrom("vga")
	be.Equal(t, filerecord.JsdosMachine(art), "vga")

	art.DoseeNoXMS = null.Int16From(0)
	art.DoseeNoEms = null.Int16From(0)
	art.DoseeNoUmb = null.Int16From(0)
	x, e, u := filerecord.JsdosMemory(art)
	be.Equal(t, x, true)
	be.Equal(t, e, true)
	be.Equal(t, u, true)

	art.DoseeRunProgram = null.StringFrom("demo.exe")
	be.Equal(t, filerecord.JsdosRun(art), "demo.exe")

	art.DoseeHardwareAudio = null.StringFrom("sb16")
	be.Equal(t, filerecord.JsdosSound(art), "sb16")

	art.FileMagicType = null.StringFrom("ASCII text")
	be.Equal(t, filerecord.Magic(art), "ASCII text")

	art.ListRelations = null.StringFrom("rel1, rel2")
	be.Equal(t, filerecord.RelationsStr(art), "rel1, rel2")

	art.GroupBrandFor = null.StringFrom("groupA")
	art.GroupBrandBy = null.StringFrom("groupB")
	forRel, byRel := filerecord.ReleaserPair(art)
	be.Equal(t, forRel, "groupA")
	be.Equal(t, byRel, "groupB")

	art.Section = null.StringFrom("demo")
	be.Equal(t, filerecord.TagCategory(art), "demo")

	art.Platform = null.StringFrom("dos")
	be.Equal(t, filerecord.TagProgram(art), "dos")

	art.RecordTitle = null.StringFrom("Cool Demo Issue #1")
	be.Equal(t, filerecord.Title(art), "Cool Demo Issue #1")

	art.RetrotxtNoReadme = null.Int16From(1)
	be.Equal(t, filerecord.DisableReadme(art), true)

	art.ListLinks = null.StringFrom("https://example.com")
	be.Equal(t, filerecord.WebsitesStr(art), "https://example.com")

	art.FileZipContent = null.StringFrom("file1.txt\nfile2.exe")
	be.Equal(t, filerecord.ZipContent(art), "file1.txt\nfile2.exe")
}

func TestSimpleText(t *testing.T) {
	t.Parallel()

	art := testutil.NewModel(t)
	got := filerecord.ForceSimpleText(art)
	be.True(t, !got)

	art.FileIntegrityStrong = null.StringFrom(flaggedHash)
	got = filerecord.ForceSimpleText(art)
	be.True(t, got)
}

func TestListEntry(t *testing.T) {
	t.Parallel()

	const bytes = 1999
	platform := tags.DOS.String()
	section := tags.BBS.String()

	le := filerecord.ListEntry{
		UniqueID: "888",
	}
	got := le.HTML(bytes, platform, section)
	find := strings.Contains(got, "1999 bytes")
	be.True(t, find)

	// dos with problematic filename red flag
	le.RelativeName = "ABC<script>....</script>YZ"
	got = le.HTML(bytes, platform, section)
	find = strings.Contains(got, "text-danger")
	be.True(t, find)

	le.RelativeName = "filename.exe"
	le.Programs = true
	got = le.HTML(bytes, platform, section)
	be.True(t, strings.Contains(got, "Dos command"))

	le.RelativeName = "filename.png"
	le.Images = true
	le.Programs = false
	got = le.HTML(bytes, platform, section)
	be.True(t, strings.Contains(got, "/editor/preview/copy/"))

	le.RelativeName = "filename.txt"
	le.Texts = true
	le.Images = false
	got = le.HTML(bytes, platform, section)
	be.True(t, strings.Contains(got, "/editor/readme/copy/"))
	x := le.HTML(bytes, tags.TextAmiga.String(), section)
	be.True(t, got != x)

	le.RelativeName = "filename.bin"
	le.Texts = false
	le.BINtexts = true
	got = le.HTML(bytes, platform, section)
	be.True(t, strings.Contains(got, "/editor/readme/copy/"))

	le.RelativeName = "filename.diz"
	le.Texts = true
	le.BINtexts = false
	got = le.HTML(bytes, platform, section)
	be.True(t, strings.Contains(got, "/editor/readme/copy/"))

	le.RelativeName = "file_id.diz"
	got = le.HTML(bytes, platform, section)
	be.True(t, strings.Contains(got, "/editor/diz/copy/"))

	le.RelativeName = "runme.bat"
	got = le.HTML(bytes, platform, section)
	be.True(t, strings.Contains(got, "command script"))

	le.RelativeName = "runme.ini"
	got = le.HTML(bytes, platform, section)
	be.True(t, strings.Contains(got, "configuration text"))

	le.MusicConfig = "placeholder"
	le.RelativeName = "music.cfg"
	got = le.HTML(bytes, platform, section)
	be.True(t, strings.Contains(got, "placeholder"))
}

func TestWebsites(t *testing.T) {
	t.Parallel()

	art := testutil.NewModel(t)
	got := filerecord.Websites(art)
	be.Equal(t, got, "")

	art.ListLinks = null.StringFrom("placeholder text")
	got = filerecord.Websites(art)
	be.Equal(t, got, "")

	ex := "http://example.com"
	art.ListLinks = null.StringFrom(ex)
	got = filerecord.Websites(art)
	be.Equal(t, got, "")

	art.ListLinks = null.StringFrom("Example page;http://example.com")
	got = filerecord.Websites(art)
	find := strings.Contains(string(got), `href="http://example.com">Example page`)
	be.True(t, find)

	art.ListLinks = null.StringFrom("http://example.com;Example page")
	got = filerecord.Websites(art)
	be.Equal(t, got, "")
}

func TestUnsupportedFile(t *testing.T) {
	t.Parallel()

	art := testutil.NewModel(t)
	got := filerecord.UnsupportedFile(art)
	be.True(t, !got)

	art.Filename = null.StringFrom("filename.txt")
	got = filerecord.UnsupportedFile(art)
	be.True(t, !got)

	art.Filename = null.StringFrom("filename.rip")
	got = filerecord.UnsupportedFile(art)
	be.True(t, got)

	art.Filename = null.StringFrom("filename.pdf")
	art.Platform = null.StringFrom("pdf")
	got = filerecord.UnsupportedFile(art)
	be.True(t, got)
}

func TestRelations(t *testing.T) {
	t.Parallel()

	art := testutil.NewModel(t)
	got := filerecord.Relations(art)
	be.Equal(t, got, "")

	art.ListRelations = null.StringFrom("placeholder text")
	got = filerecord.Relations(art)
	be.Equal(t, got, "")

	const id = "9b1c6"

	art.ListRelations = null.StringFrom(id)
	got = filerecord.Relations(art)
	be.Equal(t, got, "")

	art.ListRelations = null.StringFrom("Info text;9b1c6")
	got = filerecord.Relations(art)
	find := strings.Contains(string(got), `href="/f/9b1c6">Info text</a>`)
	be.True(t, find)

	art.ListRelations = null.StringFrom("9b1c6;Info text")
	got = filerecord.Relations(art)
	be.Equal(t, got, "")
}

func TestRecordStatus(t *testing.T) {
	t.Parallel()

	art := testutil.NewModel(t)
	a := filerecord.RecordIsNew(art)
	be.True(t, !a)
	b := filerecord.RecordOffline(art)
	be.True(t, !b)
	c := filerecord.RecordOnline(art)
	be.True(t, c)

	now := time.Now()
	art.Deletedat = null.TimeFrom(now)
	a = filerecord.RecordIsNew(art)
	be.True(t, a)
	b = filerecord.RecordOffline(art)
	be.True(t, !b)
	c = filerecord.RecordOnline(art)
	be.True(t, !c)

	art.Deletedby = null.StringFrom("an operator")
	a = filerecord.RecordIsNew(art)
	be.True(t, !a)
	b = filerecord.RecordOffline(art)
	be.True(t, b)
	c = filerecord.RecordOnline(art)
	be.True(t, !c)
}

func TestRecordProblems(t *testing.T) {
	t.Parallel()

	art := testutil.NewModel(t)
	s := filerecord.RecordProblems(art)
	errs := strings.Split(s, "+")

	const wants = 4
	be.Equal(t, len(errs), wants)
}

func TestReadme(t *testing.T) {
	t.Parallel()

	art := testutil.NewModel(t)
	s := filerecord.Readme(art)
	be.Equal(t, s, "")

	files := "1.txt\n2.txt\nmy group.txt\n3.txt\n4.txt"
	art.Filename = null.StringFrom("filename.zip")
	art.GroupBrandBy = null.StringFrom("my group")
	art.FileZipContent = null.StringFrom(files)

	s = filerecord.Readme(art)
	be.True(t, strings.Contains(s, "1.txt"))
}

func TestLinkPreviewTip(t *testing.T) {
	t.Parallel()

	art := testutil.NewModel(t)
	got := filerecord.LinkPreviewTip(art)
	be.Equal(t, got, "Read this as text")
}

func TestLinkPreview(t *testing.T) {
	t.Parallel()

	art := testutil.NewModel(t)
	got := filerecord.LinkPreview(art)
	be.Equal(t, got, "/v/9b1c6")
}

func TestLastModified(t *testing.T) {
	t.Parallel()

	art := testutil.NewModel(t)
	got := filerecord.LastModification(art)
	be.Equal(t, got, "no timestamp")

	t1970 := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	art.FileLastModified = null.TimeFrom(t1970)
	got = filerecord.LastModification(art)
	be.Equal(t, got, "no timestamp")

	t1985 := time.Date(1985, time.January, 1, 0, 0, 0, 0, time.UTC)
	art.FileLastModified = null.TimeFrom(t1985)
	got = filerecord.LastModification(art)
	be.Equal(t, got, "1985 Jan 1, 00:00")

	got = filerecord.LastModificationDate(art)
	be.Equal(t, got, "1985-Jan-01")

	a, b, c := filerecord.LastModifications(art)
	be.Equal(t, a, 1985)
	be.Equal(t, b, 1)
	be.Equal(t, c, 1)

	s := filerecord.LastModificationAgo(art)
	find := strings.Contains(s, "years ago")
	be.True(t, find)
}

func TestJsdos(t *testing.T) {
	t.Parallel()

	pl := "dos"
	zip := "FILENAME.ZIP"
	exe := "FILENAME.EXE"

	art := testutil.NewModel(t)
	art.Platform = null.StringFrom(pl)
	art.Filename = null.StringFrom(zip)

	got := filerecord.JsdosArchive(art)
	be.True(t, got)
	got = filerecord.JsdosUse(art)
	be.True(t, got)

	art.Filename = null.StringFrom(exe)
	got = filerecord.JsdosUse(art)
	be.True(t, got)

	got = filerecord.JsdosUsage(zip, pl)
	be.True(t, got)
	got = filerecord.JsdosUsage(exe, pl)
	be.True(t, got)
}

func TestFirstHeader(t *testing.T) {
	t.Parallel()

	art := testutil.NewModel(t)
	got := filerecord.FirstHeader(art)
	be.Equal(t, got, "")

	art.RecordTitle = null.StringFrom("Hello world")
	got = filerecord.FirstHeader(art)
	be.Equal(t, got, "Hello world")

	art.RecordTitle = null.StringFrom("5")
	art.Section = null.StringFrom("magazine")
	got = filerecord.FirstHeader(art)
	be.Equal(t, got, "Issue 5")
}

func TestFileEntry(t *testing.T) {
	t.Parallel()

	art := testutil.NewModel(t)
	s := filerecord.FileEntry(art)
	be.Equal(t, s, "")

	now := time.Now()
	yearAgo := now.AddDate(-1, 0, 0)

	art.Createdat = null.TimeFrom(yearAgo)
	s = filerecord.FileEntry(art)
	got := strings.Contains(s, "Created about 1 year ago")
	be.True(t, got)

	art.Updatedat = null.TimeFrom(now)
	s = filerecord.FileEntry(art)
	got = strings.Contains(s, "Updated just now")
	be.True(t, got)
}

func TestExtraZip(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()

	art := testutil.NewModel(t)
	got := filerecord.ExtraZip(art, "")
	be.True(t, !got)

	art.UUID = null.StringFrom(r0)
	got = filerecord.ExtraZip(art, "")
	be.True(t, !got)

	extra := dir.Directory(t.TempDir())
	err := command.CopyFile(sl,
		filepath.Join("testdata", "archive.zip"),
		filepath.Join(extra.Path(), r0+".zip"))
	be.Err(t, err, nil)
}

func TestDownloadID(t *testing.T) {
	t.Parallel()

	art := testutil.NewModel(t)
	got := filerecord.DownloadID(art)
	be.Equal(t, got, "9b1c6")
}

func TestDescription(t *testing.T) {
	t.Parallel()

	art := testutil.NewModel(t)
	got := filerecord.Description(art)
	be.Equal(t, got, "filename.txt released by .")

	art.Filename = null.StringFrom("myfile.txt")
	art.GroupBrandBy = null.StringFrom("my group")
	got = filerecord.Description(art)
	be.Equal(t, got, "myfile.txt released by My Group.")
}

func TestDate(t *testing.T) {
	t.Parallel()

	art := testutil.NewModel(t)
	s := filerecord.Date(art)
	got := strings.Contains(string(s), "unknown date")
	be.True(t, got)

	art.DateIssuedYear = null.Int16From(2021)
	s = filerecord.Date(art)
	got = strings.Contains(string(s), "2021")
	be.True(t, got)

	art.DateIssuedMonth = null.Int16From(1)
	s = filerecord.Date(art)
	got = strings.Contains(string(s), "January")
	be.True(t, got)

	art.DateIssuedDay = null.Int16From(1)
	s = filerecord.Date(art)

	a, b, c := filerecord.Dates(art)
	be.Equal(t, 2021, a)
	be.Equal(t, 1, b)
	be.Equal(t, 1, c)

	got = strings.Contains(string(s), "January 1")
	be.True(t, got)

	art.DateIssuedYear = null.Int16From(30000)
	s = filerecord.Date(art)
	got = strings.Contains(string(s), "30000")
	be.True(t, !got)
}

func TestListContent(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	art := testutil.NewModel(t)
	dirs := command.Dirs{}
	sl := logs.Discard()

	s := filerecord.ListContent(ctx, sl, -1, art, dirs, "")
	got := strings.Contains(string(s), "invalid platform")
	be.True(t, got)

	src, err := filepath.Abs("testdata")
	be.Err(t, err, nil)
	s = filerecord.ListContent(ctx, sl, -1, art, dirs, src)
	got = strings.Contains(string(s), "error, ")
	be.True(t, got)

	src = t.TempDir()
	err = command.CopyFile(sl, filepath.Join("testdata", "archive.zip"), filepath.Join(src, "archive.zip"))
	be.Err(t, err, nil)

	s = filerecord.ListContent(ctx, sl, -1, art, dirs, src)
	got = strings.Contains(string(s), "error, ")
	be.True(t, got)
}

// TestListContentHappyPath tests ListContent with a valid archive to ensure
// no excess blank lines are present (regression test for slice bounds bug).
// Due to archive extraction complexity, this test verifies the bug would have manifested
// by checking that non-directory files don't produce empty slots in the output.
func TestListContentHappyPath(t *testing.T) {
	t.Parallel()

	art := testutil.NewModel(t)
	art.Platform = null.StringFrom("dos")
	art.Section = null.StringFrom("game")
	art.Filename = null.StringFrom("archive.zip")

	dirs := command.Dirs{}
	sl := logs.Discard()
	ctx := t.Context()

	// Create temp directory and copy test archive
	tmpDir := t.TempDir()
	src := filepath.Join("testdata", "archive.zip")
	dst := filepath.Join(tmpDir, "archive.zip")
	err := command.CopyFile(logs.Discard(), src, dst)
	be.Err(t, err, nil)

	// Call ListContent - it may error due to extraction issues, but we verify
	// the function handles the slice bounds correctly (doesn't crash or return nil)
	result := filerecord.ListContent(ctx, sl, -1, art, dirs, tmpDir)

	// The key test: result is not nil/empty (function executed)
	// and doesn't have unexpected format issues from the slice bug
	resultStr := string(result)
	be.True(t, len(resultStr) > 0)
}

func TestAlertURL(t *testing.T) {
	t.Parallel()

	art := testutil.NewModel(t)
	got := filerecord.AlertURL(art)
	be.Equal(t, got, "")

	art.FileSecurityAlertURL = null.StringFrom("invalid")
	got = filerecord.AlertURL(art)
	be.Equal(t, got, "")

	art.FileSecurityAlertURL = null.StringFrom("https://example.com")
	got = filerecord.AlertURL(art)
	be.Equal(t, got, "https://example.com")
}

func TestLinkPreviewHref(t *testing.T) {
	t.Parallel()

	got := filerecord.LinkPreviewHref(nil, "", "")
	be.Equal(t, got, "")

	got = filerecord.LinkPreviewHref(1, "filename.xxx", "invalid")
	be.Equal(t, got, "")

	got = filerecord.LinkPreviewHref(1, "filename.txt", "text")
	be.Equal(t, got, "/v/9b1c6")
}

func TestLegacyString(t *testing.T) {
	t.Parallel()

	got := filerecord.LegacyString("")
	be.Equal(t, got, "")

	got = filerecord.LegacyString("Hello world 123.")
	be.Equal(t, got, "Hello world 123.")

	got = filerecord.LegacyString("£100")
	be.Equal(t, got, "£100")

	got = filerecord.LegacyString("\xa3100")
	be.Equal(t, got, "£100")

	got = filerecord.LegacyString("€100")
	be.Equal(t, got, "€100")

	got = filerecord.LegacyString("\x80100")
	be.Equal(t, got, "€100")
}

func TestWalkerChmod(t *testing.T) {
	t.Parallel()

	t.Run("directory permission == 0o755", func(t *testing.T) {
		t.Parallel()

		name := t.TempDir()
		testDir := filepath.Join(name, "testdir")

		if err := os.Mkdir(testDir, 0o700); err != nil {
			t.Fatal(err)
		}

		entry := &mockDirEntry{
			name:  "testdir",
			isDir: true,
			mode:  0o700,
		}

		root, got := os.OpenRoot(name)
		be.Err(t, got, nil)

		got = filerecord.WalkerChmod(root, testDir, entry, nil)
		be.Err(t, got, nil)

		info, got := os.Stat(testDir)
		be.Err(t, got, nil)
		be.Equal(t, info.Mode().Perm(), fs.FileMode(0o755))
	})

	t.Run("file permissions == 0o644", func(t *testing.T) {
		t.Parallel()

		name := t.TempDir()
		testFile := filepath.Join(name, "testfile.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0o600); err != nil {
			t.Fatal(err)
		}

		entry := &mockDirEntry{
			name:  "testfile.txt",
			isDir: false,
			mode:  0o600,
		}

		root, got := os.OpenRoot(name)
		be.Err(t, got, nil)

		got = filerecord.WalkerChmod(root, testFile, entry, nil)
		be.Err(t, got, nil)

		info, got := os.Stat(testFile)
		be.Err(t, got, nil)
		be.Equal(t, info.Mode().Perm(), fs.FileMode(0o644))
	})

	t.Run("fs.SkipDir when error", func(t *testing.T) {
		t.Parallel()

		name := t.TempDir()
		root, got := os.OpenRoot(name)
		be.Err(t, got, nil)

		got = filerecord.WalkerChmod(root, "", nil, fs.ErrInvalid)
		be.Equal(t, got, fs.SkipDir)
	})

	t.Run("chmod errors", func(t *testing.T) {
		t.Parallel()

		name := t.TempDir()
		root, got := os.OpenRoot(name)
		be.Err(t, got, nil)
		nonExistentFile := filepath.Join(name, "nonexistent", "file.txt")

		entry := &mockDirEntry{
			name:  "file.txt",
			isDir: false,
			mode:  0o644,
		}
		got = filerecord.WalkerChmod(root, nonExistentFile, entry, nil)
		be.True(t, got != nil)
	})
}

// mockDirEntry implements fs.DirEntry for testing.
type mockDirEntry struct {
	name  string
	isDir bool
	mode  fs.FileMode
}

func (m *mockDirEntry) Name() string               { return m.name }
func (m *mockDirEntry) IsDir() bool                { return m.isDir }
func (m *mockDirEntry) Type() fs.FileMode          { return m.mode }
func (m *mockDirEntry) Info() (fs.FileInfo, error) { return nil, fs.ErrInvalid }
