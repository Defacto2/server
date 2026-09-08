package remote_test

// This is a test file is to confirm there's no panics with nil values.

import (
	"context"
	"errors"
	"mime"
	"path"
	"testing"
	"time"

	"github.com/Defacto2/server/handler/app/remote"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

func TestDownload(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	tx := testutil.Tx(t)
	c := testutil.NewContext(t, "")

	dl := remote.Link("")
	got := dl.Download(t.Context(), sl, c, tx)
	be.Err(t, got)
}

func TestStat(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	tx := testutil.Tx(t)
	c := testutil.NewContext(t, "")

	dl := remote.Link("")
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

func TestFileURL(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	const fix1 = "http://files.scene.org/view/demos/groups/trsi/ms-dos/trsiscxt.zip"
	const wan1 = "https://files.scene.org/get/demos/groups/trsi/ms-dos/trsiscxt.zip"
	got := remote.ReplaceURL(sl, fix1)
	be.Equal(t, got, wan1)

	const fix2 = "https://discmaster.textfiles.com/view/4699/AmigaCD_One.bin/photos_1/screen.pic"
	const wan5 = "https://discmaster.textfiles.com/file/4699/AmigaCD_One.bin/photos_1/screen.pic"
	got = remote.ReplaceURL(sl, fix2)
	be.Equal(t, got, wan5)

	const ftp2 = "ftp://ftp.scene.org/pub/mirrors/ftp_klosz_art_pl/purgatory/Symphony2k3_Invitanimation_by_Brygada%251F_RR/RR-Symphony2k3.avi"
	const wan2 = "https://files.scene.org/get/mirrors/ftp_klosz_art_pl/purgatory/Symphony2k3_Invitanimation_by_Brygada%251F_RR/RR-Symphony2k3.avi"
	got = remote.ReplaceURL(sl, ftp2)
	be.Equal(t, got, wan2)

	const ftp3 = "ftp://ftp.pl.scene.org/pub/scene.org/parties/2003/assembly03/in64/zoom3_v1_02_final.zip"
	const wan3 = "https://files.scene.org/get/scene.org/parties/2003/assembly03/in64/zoom3_v1_02_final.zip"
	got = remote.ReplaceURL(sl, ftp3)
	be.Equal(t, got, wan3)

	const ftp4 = "ftp://sceneorg.retropc.se/scene.org/parties/2003/assembly03/in64/zoom3_v1_02_final.zip"
	const wan4 = "https://files.scene.org/get/scene.org/parties/2003/assembly03/in64/zoom3_v1_02_final.zip"
	got = remote.ReplaceURL(sl, ftp4)
	be.Equal(t, got, wan4)

	s := "this-is-an-invalid-url"
	got = remote.ReplaceURL(sl, s)
	be.Equal(t, got, s)
}

func TestGetFile_invalid(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	const timeout = remote.TimeoutShort

	r, err := remote.GetFile(t.Context(), sl, timeout, "://example.com")
	be.Err(t, err)
	be.Equal(t, r.Path, "")

	r, err = remote.GetFile(t.Context(), sl, timeout, "example.com")
	be.Err(t, err)
	be.Equal(t, r.Path, "")

	r, err = remote.GetFile(t.Context(), sl, timeout, "ftp://example.com")
	be.Err(t, err)
	be.Equal(t, r.Path, "")

	r, err = remote.GetFile(t.Context(), sl, timeout, "http://example")
	be.Err(t, err)
	be.Equal(t, r.Path, "")
}

func TestResponse(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()

	r, err := remote.GetFile(t.Context(), sl, remote.TimeoutShort, "http://example.com")
	be.True(t, (err == nil || errors.Is(err, context.DeadlineExceeded)))
	if err != nil {
		t.Log(err)
		return
	}

	/*
		ContentLength string // ContentLength is the size of the file in bytes.
		ContentType   string // ContentType is the MIME type of the file.
		LastModified  string // LastModified is the last modified date of the file.
		Path          string // Path is the path to the downloaded file.
	*/
	cl := r.ContentLength
	ct := r.ContentType
	lm := r.LastModified
	fp := r.Path

	be.True(t, len(cl) == 0) // is unused by example.com
	be.True(t, len(ct) > 0)
	be.True(t, len(lm) > 0)
	be.True(t, len(fp) > 0)

	mt, _, err := mime.ParseMediaType(ct)
	be.Err(t, err, nil)
	be.Equal(t, mt, "text/html")

	tt, err := time.Parse(time.RFC1123, lm)
	be.Err(t, err, nil)
	be.True(t, tt.Year() > 2000)

	dir, file := path.Split(fp)
	be.True(t, len(dir) > 0 && len(file) > 0)
}
