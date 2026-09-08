package fileslice_test

import (
	"slices"
	"testing"

	"github.com/Defacto2/server/handler/app/internal/fileslice"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
	"github.com/nalgeon/be"
)

// checked in Sep 26, test coverage was great at under 90%

func TestRecordsSubs(t *testing.T) {
	t.Parallel()

	s := fileslice.RecordsSub("")
	be.Equal(t, "unknown uri", s)
	s = fileslice.RecordsSub("hack")
	be.Equal(t, "game trainers or hacks", s)
}

func TestValid(t *testing.T) {
	t.Parallel()

	be.True(t, !fileslice.Valid("not-a-valid-uri"))
	be.True(t, !fileslice.Valid("/files/newest"))
	be.True(t, fileslice.Valid("newest"))
	be.True(t, fileslice.Valid("windows-pack"))
	be.True(t, fileslice.Valid("advert"))
}

func TestMatch(t *testing.T) {
	t.Parallel()

	be.Equal(t, fileslice.URI(-1), fileslice.Match("not-a-valid-uri"))
	be.Equal(t, fileslice.Newest, fileslice.Match("newest"))
	be.Equal(t, fileslice.LastURI, fileslice.Match("windows-pack"))
	be.Equal(t, fileslice.URI(1), fileslice.Match("advert"))
}

func TestRecordsSub(t *testing.T) {
	t.Parallel()

	s := fileslice.RecordsSub("")
	be.Equal(t, "unknown uri", s)
	for i := range 57 {
		be.True(t, fileslice.URI(i).String() != "unknown uri")
	}
}

func Slices() []fileslice.URI {
	return []fileslice.URI{
		fileslice.NewUploads,
		fileslice.NewUpdates,
		fileslice.ForApproval,
		fileslice.Deletions,
		fileslice.Unwanted,
		fileslice.Oldest,
		fileslice.Newest,
		fileslice.Sensenstahl,
	}
}

func TestFileInfo(t *testing.T) {
	t.Parallel()

	a, b, c := fileslice.FileInfo("")
	be.Equal(t, "unknown uri", a)
	be.Equal(t, "unknown uri", b)
	be.Equal(t, c, "")
	for uri := range slices.Values(Slices()) {
		a, b, c = fileslice.FileInfo(uri.String())
		be.True(t, a != "")
		be.True(t, b != "")
		be.True(t, c != "")
	}
}

func TestCounter(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	got, err := fileslice.Counter(t.Context(), db)
	be.Err(t, err, nil)
	be.True(t, got.Record.Bytes > 0)
	be.True(t, got.Record.Count > 0)
	be.True(t, got.Record.MinYear >= model.EpochYear)
	be.True(t, got.Record.MaxYear >= model.EpochYear) // TODO: apply model const site wide
}

func TestSorts(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	st, err := fileslice.Counter(t.Context(), db)
	be.Err(t, err, nil)

	got := st.SortName()
	be.Equal(t, got[0].Name, "ANSI art")
	be.Equal(t, got[20].Name, "Windows files")

	got = st.SortCount()
	be.True(t, got[0].Count > got[1].Count)
	be.True(t, got[1].Count > got[2].Count)
	be.True(t, got[2].Count > got[3].Count)

	got = st.SortByte()
	be.True(t, got[0].Bytes > got[1].Bytes)
	be.True(t, got[1].Bytes > got[2].Bytes)
	be.True(t, got[2].Bytes > got[3].Bytes)

	era := st.SortYear()
	be.True(t, era[3].MinYear >= era[2].MinYear)
	be.True(t, era[2].MinYear >= era[1].MinYear)
	be.True(t, era[1].MinYear >= era[0].MinYear)
}

func TestRecords(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	_, err := fileslice.Records(t.Context(), db, "", 1, 1)
	be.Err(t, err, fileslice.ErrCategory)

	_, err = fileslice.Records(t.Context(), db, "nfo", 0, 0)
	be.Err(t, err, fileslice.ErrPagi)

	got, err := fileslice.Records(t.Context(), db, "nfo", 1, 1)
	be.Err(t, err, nil)
	be.True(t, len(got) == 1)

	got, err = fileslice.Records(t.Context(), db, "nfo", 1, 2)
	be.Err(t, err, nil)
	be.True(t, len(got) > 1)

	got, err = fileslice.Records(t.Context(), db, "oldest", 1, 2)
	be.Err(t, err, nil)
	be.True(t, len(got) > 1)
}
