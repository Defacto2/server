package fileslice_test

import (
	"slices"
	"testing"

	"github.com/Defacto2/server/handler/app/internal/fileslice"
	"github.com/nalgeon/be"
)

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
	be.Equal(t, fileslice.WindowsPack, fileslice.Match("windows-pack"))
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
	_, err := fileslice.Counter(nil)
	be.Err(t, err)
}
