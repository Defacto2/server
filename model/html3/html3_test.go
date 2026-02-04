package html3_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/Defacto2/server/internal/panics"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model/html3"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/nalgeon/be"
)

func TestCreated(t *testing.T) {
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
	s := html3.LeadStr(0, "")
	be.Equal(t, s, "")
	s = html3.LeadStr(10, "Hello")
	be.Equal(t, "     ", s)
}

func TestPublished(t *testing.T) {
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
	qm := html3.SelectHTML3()
	be.True(t, qm != nil)
}

func TestOrder_String(t *testing.T) {
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
			be.Equal(t, tt.o.String(), tt.expect)
		})
	}
}

func TestInvalidExec(t *testing.T) {
	be.True(t, panics.BoilExec(nil))
	var x boil.ContextExecutor
	be.True(t, panics.BoilExec(x))
	db := sql.DB{}
	be.True(t, !panics.BoilExec(&db))
}

func TestOrderStringConsistent(t *testing.T) {
	// Test that String() returns consistent results
	s1 := html3.NameAsc.String()
	s2 := html3.NameAsc.String()
	be.Equal(t, s1, s2)
	be.True(t, len(s1) > 0)
}

func TestOrderStringAllValues(t *testing.T) {
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
// Test that common widths use cached padding
s1 := html3.LeadStr(3, "x")
s2 := html3.LeadStr(3, "y")
be.Equal(t, s1, s2)
be.Equal(t, len(s1), 2)
}

func TestLeadStrWidth7(t *testing.T) {
// Test width 7 cache
s := html3.LeadStr(7, "test")
be.Equal(t, len(s), 3)
}

func TestPublishedStateFlags(t *testing.T) {
// Test that Published works with new state flag approach
f := models.File{}
s := html3.Published(&f)
be.True(t, len(s) > 0)
}

func TestStatsDRYRefactoring(t *testing.T) {
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
