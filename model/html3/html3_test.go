package html3_test

import (
	"testing"
	"time"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model/html3"
	"github.com/aarondl/null/v8"
	"github.com/nalgeon/be"
)

func TestArts(t *testing.T) {
	t.Parallel()

	var arts html3.Arts
	be.Err(t, arts.Stat(t.Context(), nil))

	db := testutil.DB(t)
	be.Err(t, arts.Stat(t.Context(), db), nil)
	be.True(t, arts.Bytes > 0)
	be.True(t, arts.Count > 0)

	var docs html3.Documents
	be.Err(t, docs.Stat(t.Context(), db), nil)
	be.True(t, docs.Bytes > 0)
	be.True(t, docs.Count > 0)

	var sw html3.Softwares
	be.Err(t, sw.Stat(t.Context(), db), nil)
	be.True(t, sw.Bytes > 0)
	be.True(t, sw.Count > 0)
}

func TestArtsOrder(t *testing.T) {
	t.Parallel()

	_, err := html3.NameAsc.Art(t.Context(), nil, 0, 0)
	be.Err(t, err)

	db := testutil.DB(t)
	fs, err := html3.NameAsc.Art(t.Context(), db, 1, 1)
	be.Err(t, err, nil)
	be.Equal(t, len(fs), 1)
	id0 := fs[0].ID

	// Page 2 with limit 1 should return a different first record
	fs, err = html3.NameAsc.Art(t.Context(), db, 2, 1)
	be.Err(t, err, nil)
	be.Equal(t, len(fs), 1)
	be.True(t, fs[0].ID != id0)

	fs, err = html3.NameAsc.Art(t.Context(), db, 0, 0)
	be.Err(t, err, nil)
	be.True(t, len(fs) > 2)
}

func TestDocument(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	fs, err := html3.NameDes.Document(t.Context(), db, 0, 1)
	be.Err(t, err, nil)
	be.Equal(t, len(fs), 1)
}

func TestEverything(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	fs, err := html3.PublAsc.Everything(t.Context(), db, 0, 1)
	be.Err(t, err, nil)
	be.Equal(t, len(fs), 1)
}

func TestSoftware(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	fs, err := html3.PublDes.Software(t.Context(), db, 0, 1)
	be.Err(t, err, nil)
	be.Equal(t, len(fs), 1)
}

func TestByCategory(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	fs, err := html3.DescDes.ByCategory(t.Context(), db, 0, 1, "")
	be.Err(t, err, nil)
	be.Equal(t, len(fs), 0)

	fs, err = html3.DescDes.ByCategory(t.Context(), db, 0, 1, "hello world")
	be.Err(t, err, nil)
	be.Equal(t, len(fs), 0)

	fs, err = html3.DescDes.ByCategory(t.Context(), db, 0, 1, "bbs")
	be.Err(t, err, nil)
	be.Equal(t, len(fs), 1)
}

func TestByGroup(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	fs, err := html3.SizeDes.ByGroup(t.Context(), db, 0, 1, "")
	be.Err(t, err, nil)
	be.Equal(t, len(fs), 0)

	fs, err = html3.DescDes.ByGroup(t.Context(), db, 0, 1, "hello world")
	be.Err(t, err)
	be.Equal(t, len(fs), 0)

	fs, err = html3.DescDes.ByGroup(t.Context(), db, 0, 1, "defacto2")
	be.Err(t, err, nil)
	be.True(t, len(fs) >= 1)
}

func TestStat(t *testing.T) {
	t.Parallel()

	stats := html3.Stat()
	be.Equal(t, len(stats), 2)
	be.Equal(t, stats[0], postgres.SumSize)
	be.Equal(t, stats[1], postgres.TotalCnt)
}

func TestCreated(t *testing.T) {
	t.Parallel()

	be.Equal(t, html3.Created(nil), "html3: no file model")

	tests := []struct {
		arg  time.Time
		want string
	}{
		{time.Time{}, "-- --- ----"},
		{time.Date(2022, time.December, 31, 0, 0, 0, 0, time.UTC), "31-Dec-2022"},
		{time.Date(2022, time.January, 31, 0, 0, 0, 0, time.UTC), "31-Jan-2022"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			f := models.File{
				Createdat: null.Time{Time: tt.arg, Valid: true},
			}
			be.Equal(t, html3.Created(&f), tt.want)
		})
	}
}

func TestIcon(t *testing.T) {
	t.Parallel()

	be.Equal(t, html3.Icon(nil), "html3: no file model")

	f := models.File{}
	be.Equal(t, html3.Icon(&f), "unknown")

	f.Filename = null.StringFrom("file.txt")
	be.Equal(t, html3.Icon(&f), "doc")
}

func TestLeadStr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
		input string
		want  string
	}{
		{"zero", 0, "", ""},
		{"width3", 3, "x", "  "},
		{"width7", 7, "test", "   "},
		{"width10", 10, "Hello", "     "},
		{"overflow", 3, "Hello", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			be.Equal(t, html3.LeadStr(tt.width, tt.input), tt.want)
		})
	}
}

func TestPublished(t *testing.T) {
	t.Parallel()

	be.Equal(t, html3.Published(nil), "html3: no file model")

	const unknown = "       ????"
	tests := []struct {
		name    string
		y, m, d int16
		want    string
	}{
		{"-1s", -1, -1, -1, unknown},
		{"0s", 0, 0, 0, unknown},
		{"1980", 1980, 0, 0, "       1980"},
		{"1280", 1980, 12, 0, "   Dec-1980"},
		{"1980130", 1980, 13, 0, "       1980"},
		{"19801313", 1980, 13, 13, "13-???-1980"},
		{"1980113", 1980, 1, 13, "13-Jan-1980"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := models.File{
				DateIssuedYear:  null.Int16From(tt.y),
				DateIssuedMonth: null.Int16From(tt.m),
				DateIssuedDay:   null.Int16From(tt.d),
			}
			be.Equal(t, html3.Published(&f), tt.want)
		})
	}
}

func TestPublishedFW(t *testing.T) {
	t.Parallel()

	be.Equal(t, html3.PublishedFW(0, nil), "html3: no file model")

	f := models.File{}
	be.Equal(t, html3.PublishedFW(0, &f), "       ????")

	f.DateIssuedYear = null.Int16From(1980)
	be.Equal(t, html3.PublishedFW(0, &f), "       1980")
}

func TestSelectExpr(t *testing.T) {
	t.Parallel()

	be.True(t, html3.SelectExpr() != nil)
}

func TestOrderString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		o    html3.Order
		want string
	}{
		{"invalid", -1, ""},
		{"NameAsc", html3.NameAsc, "filename asc"},
		{"NameDes", html3.NameDes, "filename desc"},
		{"PublAsc", html3.PublAsc, "date_issued_year asc, date_issued_month asc, date_issued_day asc"},
		{"PublDes", html3.PublDes, "date_issued_year desc, date_issued_month desc, date_issued_day desc"},
		{"PostAsc", html3.PostAsc, "createdat asc"},
		{"PostDes", html3.PostDes, "createdat desc"},
		{"SizeAsc", html3.SizeAsc, "filesize asc"},
		{"SizeDes", html3.SizeDes, "filesize desc"},
		{"DescAsc", html3.DescAsc, "record_title asc"},
		{"DescDes", html3.DescDes, "record_title desc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			be.Equal(t, tt.o.String(), tt.want)
		})
	}
}

func TestStatsAccessors(t *testing.T) {
	t.Parallel()

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
