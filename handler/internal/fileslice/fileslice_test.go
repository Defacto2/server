package fileslice_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/internal/fileslice"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
	"github.com/nalgeon/be"
)

// checked in Sep 26, test coverage was great at under 90%

var errBoom = errors.New("boom")

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
	be.True(t, fileslice.Valid("animation"))
	be.True(t, fileslice.Valid("one-offs"))
	be.True(t, fileslice.Valid("one-offs-az"))
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
	sl := logs.Discard()
	got, err := fileslice.Counter(t.Context(), sl, db)
	be.Err(t, err, nil)
	be.True(t, got.Record.Bytes > 0)
	be.True(t, got.Record.Count > 0)
	be.True(t, got.Record.MinYear >= model.EpochYear)
	be.True(t, got.Record.MaxYear >= model.EpochYear)

	// Cross-check single-scan results against individual category/platform queries
	var ansi model.Ansi
	be.Err(t, ansi.Stat(t.Context(), db), nil)
	be.Equal(t, got.Ansi, ansi)

	var ansiBBS model.AnsiBBS
	be.Err(t, ansiBBS.Stat(t.Context(), db), nil)
	be.Equal(t, got.AnsiBBS, ansiBBS)

	var introD model.IntroMsDos
	be.Err(t, introD.Stat(t.Context(), db), nil)
	be.Equal(t, got.IntroD, introD)

	var nfo model.Nfo
	be.Err(t, nfo.Stat(t.Context(), db), nil)
	be.Equal(t, got.Nfo, nfo)

	// Verify yearly 1980-2010 breakdown
	be.Equal(t, len(got.Yearly), fileslice.YearlyMax-fileslice.YearlyMin+1)
	var sumTotal, sumDOS, sumWin, sumText int
	var sumBytes int64
	for i, pt := range got.Yearly {
		be.Equal(t, pt.Year, fileslice.YearlyMin+i)
		be.True(t, pt.DOS+pt.Windows+pt.Text <= pt.Total)
		sumTotal += pt.Total
		sumBytes += pt.Bytes
		sumDOS += pt.DOS
		sumWin += pt.Windows
		sumText += pt.Text
	}
	be.True(t, sumTotal > 0)
	be.True(t, sumBytes > 0)
	be.True(t, sumDOS > 0)
	be.True(t, sumWin > 0)
	be.True(t, sumText > 0)
}

type stubRows struct {
	cols []string
	data [][]driver.Value
	idx  int
}

func (r *stubRows) Columns() []string { return r.cols }
func (r *stubRows) Close() error      { return nil }
func (r *stubRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.data) {
		return io.EOF
	}
	copy(dest, r.data[r.idx])
	r.idx++
	return nil
}

type stubConn struct {
	queryFn func(query string) (driver.Rows, error)
}

func (c *stubConn) Prepare(string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (c *stubConn) Close() error                        { return nil }
func (c *stubConn) Begin() (driver.Tx, error)           { return nil, driver.ErrSkip }
func (c *stubConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	return c.queryFn(query)
}

type stubConnector struct {
	queryFn func(query string) (driver.Rows, error)
}

func (c stubConnector) Connect(context.Context) (driver.Conn, error) {
	return &stubConn{queryFn: c.queryFn}, nil
}
func (c stubConnector) Driver() driver.Driver { return nil }

func TestStatsGet(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	var s fileslice.Stats
	be.Err(t, s.Get(t.Context(), nil, nil))
	be.Err(t, s.Get(t.Context(), sl, nil))

	uris := tags.URIs()
	catRows := [][]driver.Value{
		{uris[tags.BBS], uris[tags.ANSI], int64(2), int64(200), int64(1992), int64(1995)},
		{uris[tags.BBS], uris[tags.Text], int64(3), int64(300), int64(1989), int64(1996)},
		{uris[tags.BBS], uris[tags.DOS], int64(1), int64(100), int64(0), int64(0)},
		{uris[tags.Demo], uris[tags.Console], int64(4), int64(400), int64(1994), int64(1999)},
		{uris[tags.Install], uris[tags.Java], int64(5), int64(500), int64(2000), int64(2002)},
		{uris[tags.Intro], uris[tags.DOS], int64(6), int64(600), int64(1991), int64(1993)},
		{uris[tags.Intro], uris[tags.Windows], int64(7), int64(700), int64(1998), int64(2004)},
		{uris[tags.Mag], uris[tags.Linux], int64(8), int64(800), int64(1995), int64(2001)},
		{uris[tags.Nfo], uris[tags.Mac], int64(9), int64(900), int64(1990), int64(1997)},
		{uris[tags.NfoTool], uris[tags.PHP], int64(10), int64(1000), int64(2003), int64(2005)},
		{uris[tags.Proof], uris[tags.Text], int64(11), int64(1100), int64(1988), int64(2006)},
	}
	yearlyRows := [][]driver.Value{
		{int64(1979), uris[tags.DOS], int64(1), int64(50)},
		{int64(1980), uris[tags.DOS], int64(2), int64(100)},
		{int64(1995), uris[tags.DOS], int64(10), int64(1000)},
		{int64(1995), uris[tags.ANSI], int64(4), int64(400)},
		{int64(1995), uris[tags.Text], int64(3), int64(300)},
		{int64(1995), uris[tags.Windows], int64(5), int64(500)},
		{int64(1995), uris[tags.Linux], int64(1), int64(250)},
		{int64(2010), uris[tags.Windows], int64(6), int64(600)},
		{int64(2011), uris[tags.Windows], int64(9), int64(900)},
	}

	db := sql.OpenDB(stubConnector{
		queryFn: func(query string) (driver.Rows, error) {
			if strings.Contains(query, "BETWEEN 1980 AND 2010") {
				return &stubRows{
					cols: []string{"date_issued_year", "platform", "count", "sum"},
					data: yearlyRows,
				}, nil
			}
			return &stubRows{
				cols: []string{"section", "platform", "count", "sum", "min", "max"},
				data: catRows,
			}, nil
		},
	})
	defer func() { _ = db.Close() }()

	be.Err(t, s.Get(t.Context(), sl, db), nil)
	be.Equal(t, s.Record.Count, 66)
	be.Equal(t, s.Record.Bytes, 6600)
	be.Equal(t, s.Record.MinYear, 1988)
	be.Equal(t, s.Record.MaxYear, 2006)
	be.True(t, !s.Record.UpdatedAt.IsZero())

	be.Equal(t, s.BBS.Count, 6)
	be.Equal(t, s.BBS.MinYear, 1989)
	be.Equal(t, s.BBS.MaxYear, 1996)
	be.Equal(t, s.AnsiBBS.Count, 2)
	be.Equal(t, s.BBSText.Count, 3)
	be.Equal(t, s.BBStro.Count, 1)
	be.Equal(t, s.Demoscene.Count, 4)
	be.Equal(t, s.Installer.Count, 5)
	be.Equal(t, s.Intro.Count, 13)
	be.Equal(t, s.IntroD.Count, 6)
	be.Equal(t, s.IntroW.Count, 7)
	be.Equal(t, s.Magazine.Count, 8)
	be.Equal(t, s.Nfo.Count, 9)
	be.Equal(t, s.NfoTool.Count, 10)
	be.Equal(t, s.Proof.Count, 11)

	be.Equal(t, s.Ansi.Count, 2)
	be.Equal(t, s.Console.Count, 4)
	be.Equal(t, s.MsDos.Count, 7)
	be.Equal(t, s.Java.Count, 5)
	be.Equal(t, s.Linux.Count, 8)
	be.Equal(t, s.Macos.Count, 9)
	be.Equal(t, s.Script.Count, 10)
	be.Equal(t, s.Text.Count, 14)
	be.Equal(t, s.Windows.Count, 7)

	be.Equal(t, len(s.Yearly), fileslice.YearlyMax-fileslice.YearlyMin+1)
	be.Equal(t, s.Yearly[0], fileslice.YearlyPoint{Year: 1980, Total: 2, Bytes: 100, DOS: 2})
	be.Equal(t, s.Yearly[1995-fileslice.YearlyMin], fileslice.YearlyPoint{
		Year: 1995, Total: 23, Bytes: 2450, DOS: 10, Windows: 5, Text: 7,
	})
	be.Equal(t, s.Yearly[len(s.Yearly)-1], fileslice.YearlyPoint{Year: 2010, Total: 6, Bytes: 600, Windows: 6})

	// Query error paths (category query and yearly query)
	errDB := sql.OpenDB(stubConnector{
		queryFn: func(string) (driver.Rows, error) { return nil, errBoom },
	})
	defer func() { _ = errDB.Close() }()
	var sErr fileslice.Stats
	be.Err(t, sErr.Get(t.Context(), sl, errDB))

	yearlyErrDB := sql.OpenDB(stubConnector{
		queryFn: func(query string) (driver.Rows, error) {
			if strings.Contains(query, "BETWEEN 1980 AND 2010") {
				return nil, errBoom
			}
			return &stubRows{
				cols: []string{"section", "platform", "count", "sum", "min", "max"},
			}, nil
		},
	})
	defer func() { _ = yearlyErrDB.Close() }()
	be.Err(t, sErr.Get(t.Context(), sl, yearlyErrDB))
}

func TestSorts(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	sl := logs.Discard()
	st, err := fileslice.Counter(t.Context(), sl, db)
	be.Err(t, err, nil)

	got := st.SortName()
	be.Equal(t, got[0].Name, "ANSI art")
	be.Equal(t, got[len(got)-1].Name, "Windows files")

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
