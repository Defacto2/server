package html3_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model/html3"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/nalgeon/be"
)

// checked in Aug 26, test coverage was good at around 75%+

func openDB(t *testing.T) *sql.DB {
	db, err := postgres.Open()
	if err != nil {
		t.Log("postgres open", err)
		return nil
	}

	if err := db.Ping(); err != nil {
		return nil
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Log("cleanup database", err)
		}
	})

	return db
}

func TestArts(t *testing.T) {
	t.Parallel()

	var arts html3.Arts
	got := arts.Stat(context.TODO(), nil)
	be.Err(t, got)

	db := openDB(t)
	if db == nil {
		return
	}
	got = arts.Stat(t.Context(), db)
	be.Err(t, got, nil)
	be.True(t, arts.Bytes > 0)
	be.True(t, arts.Count > 0)

	var docs html3.Documents
	got = docs.Stat(t.Context(), db)
	be.Err(t, got, nil)
	be.True(t, docs.Bytes > 0)
	be.True(t, docs.Count > 0)

	var sw html3.Softwares
	got = sw.Stat(t.Context(), db)
	be.Err(t, got, nil)
	be.True(t, sw.Bytes > 0)
	be.True(t, sw.Count > 0)
}

func TestArtsOrder(t *testing.T) {
	t.Parallel()

	_, got := html3.NameAsc.Art(context.TODO(), nil, 0, 0)
	be.Err(t, got)

	db := openDB(t)
	if db == nil {
		return
	}

	fs, got := html3.NameAsc.Art(t.Context(), db, 0, 1)
	be.Err(t, got, nil)
	be.True(t, len(fs) == 1)
	id0 := fs[0].ID

	// test offset and confirm id of the first record is different
	fs, got = html3.NameAsc.Art(t.Context(), db, 1, 2)
	be.Err(t, got, nil)
	be.True(t, len(fs) == 2)
	id1 := fs[1].ID
	be.True(t, id0 != id1)

	fs, got = html3.NameAsc.Art(t.Context(), db, 0, 0)
	be.Err(t, got, nil)
	be.True(t, len(fs) > 2)
}

func TestDocument(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}

	fs, got := html3.NameDes.Document(t.Context(), db, 0, 1)
	be.Err(t, got, nil)
	be.True(t, len(fs) == 1)
}

func TestEverything(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}

	fs, got := html3.PublAsc.Everything(t.Context(), db, 0, 1)
	be.Err(t, got, nil)
	be.True(t, len(fs) == 1)
}

func TestSoftware(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}

	fs, got := html3.PublDes.Software(t.Context(), db, 0, 1)
	be.Err(t, got, nil)
	be.True(t, len(fs) == 1)
}

func TestByCategory(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}

	const none = ""
	fs, got := html3.DescDes.ByCategory(t.Context(), db, 0, 1, none)
	be.Err(t, got, nil)
	be.True(t, len(fs) == 0)

	const invalid = "hello world"
	fs, got = html3.DescDes.ByCategory(t.Context(), db, 0, 1, invalid)
	be.Err(t, got, nil)
	be.True(t, len(fs) == 0)

	const valid = "bbs"
	fs, got = html3.DescDes.ByCategory(t.Context(), db, 0, 1, valid)
	be.Err(t, got, nil)
	be.True(t, len(fs) == 1)
}

func TestByGroup(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}

	const none = ""
	fs, got := html3.SizeDes.ByGroup(t.Context(), db, 0, 1, none)
	be.Err(t, got, nil)
	be.True(t, len(fs) == 0)

	const invalid = "hello world"
	fs, got = html3.DescDes.ByGroup(t.Context(), db, 0, 1, invalid)
	be.Err(t, got, nil)
	be.True(t, len(fs) == 0)

	const valid = "defacto2"
	fs, got = html3.DescDes.ByCategory(t.Context(), db, 0, 1, valid)
	be.Err(t, got, nil)
	be.True(t, len(fs) == 1)
}

// TestStat verifies the Stat function returns expected column selections.
func TestStat(t *testing.T) {
	t.Parallel()

	stats := html3.Stat()
	be.Equal(t, 2, len(stats))
	be.Equal(t, postgres.SumSize, stats[0])
	be.Equal(t, postgres.TotalCnt, stats[1])
}

func TestCreated(t *testing.T) {
	t.Parallel()

	loc := time.Local //nolint:gosmopolitan
	tests := []struct {
		arg    time.Time
		expect string
	}{
		{time.Time{}, "-- --- ----"},
		{time.Date(2022, time.December, 31, 0, 0, 0, 0, loc), "31-Dec-2022"},
		{time.Date(2022, time.January, 31, 0, 0, 0, 0, loc), "31-Jan-2022"},
	}
	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			t.Parallel()
			f := models.File{
				Createdat: null.Time{
					Time:  tt.arg,
					Valid: true,
				},
			}
			be.Equal(t, tt.expect, html3.Created(&f))
		})
	}
}

func TestIcon(t *testing.T) {
	t.Parallel()

	s := html3.Icon(nil)
	be.Equal(t, "error, no file model", s)

	f := models.File{}
	s = html3.Icon(&f)
	be.Equal(t, "unknown", s)

	f.Filename = null.StringFrom("file.txt")
	s = html3.Icon(&f)
	be.Equal(t, "doc", s)
}

func TestLeadStr(t *testing.T) {
	t.Parallel()

	s := html3.LeadStr(0, "")
	be.Equal(t, s, "")

	s = html3.LeadStr(10, "Hello")
	be.Equal(t, "     ", s)
}

func TestPublished(t *testing.T) {
	t.Parallel()
	const errS = "       ????"
	type args struct {
		y int
		m int
		d int
	}
	tests := []struct {
		name   string
		args   args
		expect string
	}{
		{"-1s", args{-1, -1, -1}, errS},
		{"0s", args{0, 0, 0}, errS},
		{"1980", args{1980, 0, 0}, "       1980"},
		{"1280", args{1980, 12, 0}, "   Dec-1980"},
		{"1980130", args{1980, 13, 0}, "       1980"},
		{"19801313", args{1980, 13, 13}, "13-???-1980"},
		{"1980113", args{1980, 1, 13}, "13-Jan-1980"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			y := null.Int16{Int16: int16(tt.args.y), Valid: true} //nolint:gosec
			m := null.Int16{Int16: int16(tt.args.m), Valid: true} //nolint:gosec
			d := null.Int16{Int16: int16(tt.args.d), Valid: true} //nolint:gosec
			f := models.File{
				DateIssuedYear:  y,
				DateIssuedMonth: m,
				DateIssuedDay:   d,
			}
			be.Equal(t, html3.Published(&f), tt.expect)
		})
	}
}

func TestPublishedFW(t *testing.T) {
	t.Parallel()

	s := html3.PublishedFW(0, nil)
	be.Equal(t, "error, no file model", s)

	f := models.File{}
	s = html3.PublishedFW(0, &f)
	be.Equal(t, "       ????", s)

	f.DateIssuedYear = null.Int16From(1980)
	s = html3.PublishedFW(0, &f)
	be.Equal(t, "       1980", s)
}

func TestSelectHTML3(t *testing.T) {
	t.Parallel()

	qm := html3.SelectExpr()
	be.True(t, qm != nil)
}

func TestOrder_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o      html3.Order
		expect string
	}{
		{-1, ""},
		{html3.NameAsc, "filename asc"},
		{html3.DescDes, "record_title desc"},
	}
	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			t.Parallel()
			be.Equal(t, tt.o.String(), tt.expect)
		})
	}
}

func TestInvalidExec(t *testing.T) {
	t.Parallel()

	be.True(t, nils.BoilExec(nil))

	var x boil.ContextExecutor
	be.True(t, nils.BoilExec(x))

	db := sql.DB{}
	be.True(t, !nils.BoilExec(&db))
}

func TestOrderStringConsistent(t *testing.T) {
	t.Parallel()

	// Test that String() returns consistent results
	s1 := html3.NameAsc.String()
	s2 := html3.NameAsc.String()
	be.Equal(t, s1, s2)
	be.True(t, len(s1) > 0)
}

func TestOrderStringAllValues(t *testing.T) {
	t.Parallel()

	// Test all Order values return non-empty strings
	orders := []html3.Order{
		html3.NameAsc, html3.NameDes,
		html3.PublAsc, html3.PublDes,
		html3.PostAsc, html3.PostDes,
		html3.SizeAsc, html3.SizeDes,
		html3.DescAsc, html3.DescDes,
	}
	for _, o := range orders {
		s := o.String()
		be.True(t, len(s) > 0)
	}
}

func TestOrderStringValues(t *testing.T) {
	t.Parallel()

	// Test specific order string values
	tests := []struct {
		o      html3.Order
		expect string
	}{
		{html3.NameAsc, "filename asc"},
		{html3.NameDes, "filename desc"},
		{html3.SizeAsc, "filesize asc"},
		{html3.SizeDes, "filesize desc"},
	}
	for _, tt := range tests {
		be.Equal(t, tt.o.String(), tt.expect)
	}
}

func TestLeadStrCaching(t *testing.T) {
	t.Parallel()

	// Test that common widths use cached padding
	const width = 3
	s1 := html3.LeadStr(width, "x")
	s2 := html3.LeadStr(width, "y")
	be.Equal(t, s1, s2)
	be.Equal(t, len(s1), 2)
}

func TestLeadStrWidth7(t *testing.T) {
	t.Parallel()

	const width = 7
	s := html3.LeadStr(width, "test")
	be.Equal(t, len(s), 3)
}

func TestPublishedFlag(t *testing.T) {
	t.Parallel()

	// Test that Published works with new state flag approach
	f := models.File{}
	s := html3.Published(&f)
	be.True(t, len(s) > 0)
}

func TestStats(t *testing.T) {
	t.Parallel()

	// Test that Arts, Documents, Softwares work correctly
	a := &html3.Arts{}
	be.Equal(t, a.GetBytes(), 0)
	a.SetBytes(100)
	be.Equal(t, a.GetBytes(), 100)

	d := &html3.Documents{}
	be.Equal(t, d.GetCount(), 0)
	d.SetCount(50)
	be.Equal(t, d.GetCount(), 50)

	s := &html3.Softwares{}
	be.Equal(t, s.GetBytes(), 0)
	s.SetBytes(200)
	be.Equal(t, s.GetBytes(), 200)
}
