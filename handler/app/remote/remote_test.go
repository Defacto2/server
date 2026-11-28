package remote_test

import (
	"archive/zip"
	"database/sql"
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/Defacto2/server/handler/app/remote"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

func TestDownload(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	tx := testutil.Tx(t)
	c := testutil.NewContext(t, "")

	dl := remote.Demozoo(0, "", "", 0)
	got := dl.Download(t.Context(), sl, c, tx, false)
	be.Err(t, got)
}

func TestStat(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	tx := testutil.Tx(t)
	c := testutil.NewContext(t, "")

	dl := remote.Demozoo(0, "", "", 0)
	got := dl.Stat(t.Context(), sl, c, tx)
	be.Err(t, got)
}

func TestArchiveContent(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	tx := testutil.Tx(t)
	c := testutil.NewContext(t, "")
	dl := remote.DemozooLink{}

	got := dl.ArchiveDo(t.Context(), sl, c, tx, "")
	be.Err(t, got, nil)
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	tx := testutil.Tx(t)
	c := testutil.NewContext(t, "")
	dl := remote.DemozooLink{}

	err := dl.Update(t.Context(), c, tx)
	be.Err(t, err)
}

func TestDemozooOffline(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	sl := logs.Discard()
	c := testutil.NewContext(t, "")
	dummyDB := &sql.DB{}

	dl := remote.Demozoo(42, testutil.UID4, dir.Directory(t.TempDir()), remote.TimeoutShort)
	be.Equal(t, dl.ID, 42)
	be.Equal(t, dl.UUID, testutil.UID4)

	// Nil-guard & missing ZooAuth checks.
	be.Err(t, dl.Download(ctx, nil, c, dummyDB, false))
	be.Err(t, dl.Download(ctx, sl, c, dummyDB, false), config.ErrNoZooAuth)
	be.Err(t, dl.Stat(ctx, nil, c, dummyDB))
	be.Err(t, dl.ArchiveDo(ctx, nil, c, dummyDB, ""))
	be.Err(t, dl.Update(ctx, nil, dummyDB))

	// Stat on missing file returns error.
	be.Err(t, dl.Stat(ctx, sl, c, dummyDB))

	// Stat and ArchiveDo on a real zip archive populate size, hash, type, and content.
	tmp := t.TempDir()
	zipUID := uuid.New().String()
	zipPath := filepath.Join(tmp, zipUID)
	f, err := os.Create(zipPath)
	be.Err(t, err, nil)
	zw := zip.NewWriter(f)
	w, err := zw.Create("FILE_ID.DIZ")
	be.Err(t, err, nil)
	_, err = w.Write([]byte("scene diz"))
	be.Err(t, err, nil)
	be.Err(t, zw.Close(), nil)
	be.Err(t, f.Close(), nil)

	dlZip := remote.Demozoo(42, zipUID, dir.Directory(tmp), 0)
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Join(os.TempDir(), dir.Pattern+"-"+zipUID)) })
	be.Err(t, dlZip.Stat(ctx, sl, c, dummyDB), nil)
	be.True(t, dlZip.FileSize > 0)
	be.True(t, len(dlZip.FileHash) == 96)
	be.True(t, dlZip.FileType != "")
	be.Equal(t, dlZip.Content, "FILE_ID.DIZ")

	// ArchiveDo on non-archive logs and returns nil.
	be.Err(t, dlZip.ArchiveDo(ctx, sl, c, dummyDB, filepath.Join(tmp, "missing.zip")), nil)
}

func TestPouetOffline(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	sl := logs.Discard()
	dummyDB := &sql.DB{}

	pl := remote.Pouet(99, testutil.UID4, dir.Directory(t.TempDir()), remote.TimeoutShort)
	be.Equal(t, pl.PouetID, 99)
	be.Equal(t, pl.UUID, testutil.UID4)

	// Nil-guard checks.
	be.Err(t, pl.Download(ctx, nil, dummyDB))
	be.Err(t, pl.Stat(ctx, nil, dummyDB))
	be.Err(t, pl.Update(ctx, nil))

	// Stat on missing file returns error.
	be.Err(t, pl.Stat(ctx, sl, dummyDB))

	// Stat on real file populates size, hash, and type.
	tmp := t.TempDir()
	p := filepath.Join(tmp, testutil.UID4)
	be.Err(t, os.WriteFile(p, []byte("hello pouet\n"), 0o600), nil)

	plFile := remote.Pouet(99, testutil.UID4, dir.Directory(tmp), 0)
	be.Err(t, plFile.Stat(ctx, sl, dummyDB), nil)
	be.Equal(t, plFile.FileSize, len("hello pouet\n"))
	be.True(t, len(plFile.FileHash) == 96)
	be.True(t, plFile.FileType != "")
}

func TestFileURL(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	const fix1 = "http://files.scene.org/view/demos/groups/trsi/ms-dos/trsiscxt.zip"
	const wan1 = "https://files.scene.org/get/demos/groups/trsi/ms-dos/trsiscxt.zip"
	be.Equal(t, remote.ReplaceURL(nil, fix1), wan1)
	be.Equal(t, remote.ReplaceURL(sl, fix1), wan1)

	const fix2 = "https://discmaster.textfiles.com/view/4699/AmigaCD_One.bin/photos_1/screen.pic"
	const wan5 = "https://discmaster.textfiles.com/file/4699/AmigaCD_One.bin/photos_1/screen.pic"
	be.Equal(t, remote.ReplaceURL(sl, fix2), wan5)

	const ftp2 = "ftp://ftp.scene.org/pub/mirrors/ftp_klosz_art_pl/purgatory/Symphony2k3_Invitanimation_by_Brygada%251F_RR/RR-Symphony2k3.avi"
	const wan2 = "https://files.scene.org/get/mirrors/ftp_klosz_art_pl/purgatory/Symphony2k3_Invitanimation_by_Brygada%251F_RR/RR-Symphony2k3.avi"
	be.Equal(t, remote.ReplaceURL(sl, ftp2), wan2)

	const ftp3 = "ftp://ftp.pl.scene.org/pub/scene.org/parties/2003/assembly03/in64/zoom3_v1_02_final.zip"
	const wan3 = "https://files.scene.org/get/scene.org/parties/2003/assembly03/in64/zoom3_v1_02_final.zip"
	be.Equal(t, remote.ReplaceURL(sl, ftp3), wan3)

	const ftp4 = "ftp://sceneorg.retropc.se/scene.org/parties/2003/assembly03/in64/zoom3_v1_02_final.zip"
	const wan4 = "https://files.scene.org/get/scene.org/parties/2003/assembly03/in64/zoom3_v1_02_final.zip"
	be.Equal(t, remote.ReplaceURL(sl, ftp4), wan4)

	const php = "https://scene.org/file.php?id=123"
	be.Equal(t, remote.ReplaceURL(sl, php), php)

	const bad = "://bad-url"
	be.Equal(t, remote.ReplaceURL(sl, bad), bad)

	// Host-only URLs without a path must not panic.
	for _, bare := range []string{
		"https://files.scene.org",
		"ftp://ftp.scene.org",
		"ftp://sceneorg.retropc.se",
		"https://discmaster.textfiles.com",
	} {
		be.Equal(t, remote.ReplaceURL(sl, bare), bare)
	}

	s := "this-is-an-invalid-url"
	be.Equal(t, remote.ReplaceURL(sl, s), s)
}

func TestGetFile_invalid(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	const timeout = remote.TimeoutShort

	_, err := remote.GetFile(t.Context(), nil, timeout, "http://example.com")
	be.Err(t, err)

	r, err := remote.GetFile(t.Context(), sl, timeout, "://example.com")
	be.Err(t, err)
	be.Equal(t, r.Path, "")

	r, err = remote.GetFile(t.Context(), sl, timeout, "example.com")
	be.Err(t, err)
	be.Equal(t, r.Path, "")

	r, err = remote.GetFile(t.Context(), sl, timeout, "ftp://example.com")
	be.Err(t, err)
	be.Equal(t, r.Path, "")

	r, err = remote.GetFile(t.Context(), sl, -1*time.Second, "http://127.0.0.1:1")
	be.Err(t, err)
	be.Equal(t, r.Path, "")
}

func TestResponse(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	const payload = "<html><body>demo payload</body></html>"
	lastMod := time.Date(2024, time.May, 10, 12, 0, 0, 0, time.UTC).Format(http.TimeFormat)

	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		if r.URL.Path == "/missing" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Last-Modified", lastMod)
		_, _ = w.Write([]byte(payload))
	}))
	defer srv.Close()

	// 404 returns ErrStatus and leaves no temp file.
	r404, err := remote.GetFile(t.Context(), sl, remote.TimeoutShort, srv.URL+"/missing")
	be.Err(t, err, remote.ErrStatus)
	be.Equal(t, r404.Path, "")

	// 200 downloads payload to temp file and captures response headers.
	r, err := remote.GetFile(t.Context(), sl, remote.TimeoutShort, srv.URL+"/demo.html")
	be.Err(t, err, nil)
	t.Cleanup(func() { _ = os.Remove(r.Path) })

	be.Equal(t, gotUA, remote.UserAgent)
	be.True(t, len(r.ContentLength) > 0)
	be.True(t, len(r.ContentType) > 0)
	be.Equal(t, r.LastModified, lastMod)
	be.True(t, len(r.Path) > 0)

	mt, _, err := mime.ParseMediaType(r.ContentType)
	be.Err(t, err, nil)
	be.Equal(t, mt, "text/html")

	tt, err := time.Parse(time.RFC1123, r.LastModified)
	be.Err(t, err, nil)
	be.Equal(t, tt.Year(), 2024)

	d, file := path.Split(r.Path)
	be.True(t, len(d) > 0 && len(file) > 0)
	be.True(t, strings.HasPrefix(file, "df2app-get-remotefile-") || strings.Contains(file, "get-remotefile-"))

	body, err := os.ReadFile(r.Path)
	be.Err(t, err, nil)
	be.Equal(t, string(body), payload)

	// Timeouts above 60s clamp to 60s (not 60ns) and succeed cleanly.
	rLong, err := remote.GetFile(t.Context(), sl, 2*time.Minute, srv.URL+"/demo.html")
	be.Err(t, err, nil)
	t.Cleanup(func() { _ = os.Remove(rLong.Path) })
	be.True(t, rLong.Path != "")
}
