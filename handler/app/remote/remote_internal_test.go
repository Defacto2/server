package remote

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/nalgeon/be"
)

func TestRemoteInternalHelpers(t *testing.T) {
	t.Parallel()

	// sceneorg with < 2 elements returns empty URL.
	be.Equal(t, sceneorg([]string{"only"}).String(), "")

	// renameOW: success, matching files, and mismatched overwrite failure.
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.bin")
	dst := filepath.Join(tmp, "dst.bin")
	be.Err(t, os.WriteFile(src, []byte("same-bytes"), 0o600), nil)
	be.Err(t, renameOW(src, dst), nil)
	data, err := os.ReadFile(dst)
	be.Err(t, err, nil)
	be.Equal(t, string(data), "same-bytes")

	// Missing source fails FileMatch.
	be.Err(t, renameOW(filepath.Join(tmp, "missing"), dst))

	// cleanup with nil response does not panic.
	var g got
	g.msg = "test"
	g.cleanup(logs.Discard())
}

func TestDemozooSettersAndRemote(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	dl := &DemozooLink{
		ID:          10,
		Platform:    "dos",
		Section:     "demo",
		Filename:    "demo.zip",
		FileSize:    1234,
		FileType:    "PKZIP archive",
		FileHash:    "abc123",
		Content:     "FILE_ID.DIZ",
		CreditAudio: []string{"Musician"},
		CreditArt:   []string{"Artist"},
		CreditCode:  []string{"Coder"},
		CreditText:  []string{"Writer"},
		Github:      "Defacto2/server",
		YouTube:     "3V8rpJDpbKg",
		Pouet:       777,
		Releaser1:   "GroupA",
		Releaser2:   "GroupB",
		Title:       "Cool Demo",
		IssuedYear:  1995,
		IssuedMonth: 6,
		IssuedDay:   15,
		linkURL:     "https://example.invalid/demo.zip",
		msg:         "test",
	}

	// Nil model guards do not panic.
	dl.setFile(nil)
	dl.setSceners(nil)
	dl.setMisc(nil)

	var f models.File
	dl.setFile(&f)
	dl.setSceners(&f)
	dl.setMisc(&f)
	be.Equal(t, f.Platform.String, "dos")
	be.Equal(t, f.Section.String, "demo")
	be.Equal(t, f.Filename.String, "demo.zip")
	be.Equal(t, f.Filesize.Int64, int64(1234))
	be.Equal(t, f.FileMagicType.String, "PKZIP archive")
	be.Equal(t, f.FileIntegrityStrong.String, "abc123")
	be.Equal(t, f.FileZipContent.String, "FILE_ID.DIZ")
	be.Equal(t, f.CreditAudio.String, "Musician")
	be.Equal(t, f.CreditIllustration.String, "Artist")
	be.Equal(t, f.CreditProgram.String, "Coder")
	be.Equal(t, f.CreditText.String, "Writer")
	be.Equal(t, f.WebIDGithub.String, "Defacto2/server")
	be.Equal(t, f.WebIDYoutube.String, "3V8rpJDpbKg")
	be.Equal(t, f.WebIDPouet.Int64, int64(777))
	be.Equal(t, f.GroupBrandFor.String, "GroupA")
	be.Equal(t, f.GroupBrandBy.String, "GroupB")
	be.Equal(t, f.RecordTitle.String, "Cool Demo")
	be.Equal(t, f.DateIssuedYear.Int16, int16(1995))
	be.Equal(t, f.DateIssuedMonth.Int16, int16(6))
	be.Equal(t, f.DateIssuedDay.Int16, int16(15))

	dl.log(sl, "info", nil)

	// remote nil check.
	_, err := dl.remote(t.Context(), nil, 0)
	be.Err(t, err)

	// remote against local httptest server: non-last 404 returns empty Response + nil; last 404 returns error; 200 succeeds.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/bad" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte("zip-data"))
	}))
	defer srv.Close()

	dl.prod = demozoo.Production{
		DownloadLinks: []struct {
			LinkClass string `json:"link_class"`
			URL       string `json:"url"`
		}{
			{URL: srv.URL + "/bad"},
			{URL: srv.URL + "/good"},
		},
	}
	dl.linkURL = srv.URL + "/bad"
	resp, err := dl.remote(t.Context(), sl, 0)
	be.Err(t, err, nil)
	be.Equal(t, resp, Response{})

	_, err = dl.remote(t.Context(), sl, 1)
	be.Err(t, err)

	dl.linkURL = srv.URL + "/good"
	resp, err = dl.remote(t.Context(), sl, 1)
	be.Err(t, err, nil)
	t.Cleanup(func() { _ = os.Remove(resp.Path) })
	be.True(t, resp.Path != "")
}

func TestPouetUpdateValuesAndLog(t *testing.T) {
	t.Parallel()

	pl := &PouetLink{
		PouetID:     55,
		DemozooID:   123,
		Releaser1:   "Group1",
		Releaser2:   "Group2",
		Title:       "Intro",
		IssuedYear:  1994,
		IssuedMonth: 3,
		IssuedDay:   20,
		Filename:    "intro.zip",
		FileSize:    512,
		FileType:    "PKZIP",
		FileHash:    "hash99",
		Content:     "INTRO.COM",
		Platform:    "dos",
		Section:     "intro",
		msg:         "test",
	}
	pl.prod.Download = "https://files.scene.org/intro.zip"

	pl.updateValues(nil)

	var f models.File
	pl.updateValues(&f)
	be.Equal(t, f.WebIDDemozoo.Int64, int64(123))
	be.Equal(t, f.GroupBrandFor.String, "Group1")
	be.Equal(t, f.GroupBrandBy.String, "Group2")
	be.Equal(t, f.RecordTitle.String, "Intro")
	be.Equal(t, f.DateIssuedYear.Int16, int16(1994))
	be.Equal(t, f.DateIssuedMonth.Int16, int16(3))
	be.Equal(t, f.DateIssuedDay.Int16, int16(20))
	be.Equal(t, f.Filename.String, "intro.zip")
	be.Equal(t, f.Filesize.Int64, int64(512))
	be.Equal(t, f.FileMagicType.String, "PKZIP")
	be.Equal(t, f.FileIntegrityStrong.String, "hash99")
	be.Equal(t, f.FileZipContent.String, "INTRO.COM")
	be.Equal(t, f.Platform.String, "dos")
	be.Equal(t, f.Section.String, "intro")

	pl.log(logs.Discard(), "done", nil)
}
