package html3_test

import (
	"context"
	"testing"
	"time"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model/html3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
)

func TestCreated(t *testing.T) {
	loc := time.Local
	tests := []struct {
		arg       time.Time
		expect    string
		assertion assert.ComparisonAssertionFunc
	}{
		{time.Time{}, "-- --- ----", assert.Equal},
		{time.Date(2022, time.December, 31, 0, 0, 0, 0, loc), "31-Dec-2022", assert.Equal},
		{time.Date(2022, time.January, 31, 0, 0, 0, 0, loc), "31-Jan-2022", assert.Equal},
	}
	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			f := models.File{
				Createdat: null.Time{
					Time:  tt.arg,
					Valid: true,
				},
			}
			tt.assertion(t, tt.expect, html3.Created(&f))
		})
	}
}

func TestIcon(t *testing.T) {
	s := html3.Icon(nil)
	assert.Equal(t, "error, no file model", s)
	f := models.File{}
	s = html3.Icon(&f)
	assert.Equal(t, "unknown", s)
	f.Filename = null.StringFrom("file.txt")
	s = html3.Icon(&f)
	assert.Equal(t, "doc", s)
}

func TestLeadStr(t *testing.T) {
	s := html3.LeadStr(0, "")
	assert.Equal(t, "", s)
	s = html3.LeadStr(10, "Hello")
	assert.Equal(t, "     ", s)
}

func TestPublished(t *testing.T) {
	const errS = "       ????"
	type args struct {
		y int
		m int
		d int
	}
	tests := []struct {
		name      string
		args      args
		expect    string
		assertion assert.ComparisonAssertionFunc
	}{
		{"-1s", args{-1, -1, -1}, errS, assert.Equal},
		{"0s", args{0, 0, 0}, errS, assert.Equal},
		{"1980", args{1980, 0, 0}, "       1980", assert.Equal},
		{"1280", args{1980, 12, 0}, "   Dec-1980", assert.Equal},
		{"1980", args{1980, 13, 0}, "       1980", assert.Equal},
		{"1980", args{1980, 13, 13}, "13-???-1980", assert.Equal},
		{"1980", args{1980, 1, 13}, "13-Jan-1980", assert.Equal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			y := null.Int16{Int16: int16(tt.args.y), Valid: true}
			m := null.Int16{Int16: int16(tt.args.m), Valid: true}
			d := null.Int16{Int16: int16(tt.args.d), Valid: true}
			f := models.File{
				DateIssuedYear:  y,
				DateIssuedMonth: m,
				DateIssuedDay:   d,
			}
			tt.assertion(t, tt.expect, html3.Published(&f))
		})
	}
}

func TestPublishedFW(t *testing.T) {
	s := html3.PublishedFW(0, nil)
	assert.Equal(t, "error, no file model", s)
	f := models.File{}
	s = html3.PublishedFW(0, &f)
	assert.Equal(t, "       ????", s)
	f.DateIssuedYear = null.Int16From(1980)
	s = html3.PublishedFW(0, &f)
	assert.Equal(t, "       1980", s)
}

func TestSelectHTML3(t *testing.T) {
	qm := html3.SelectHTML3()
	assert.NotEmpty(t, qm)
}

func TestArts_Stat(t *testing.T) {
	t.Parallel()
	a := html3.Arts{}
	ctx := context.TODO()
	err := a.Stat(ctx, nil)
	require.ErrorIs(t, err, html3.ErrDB)
	db, err := postgres.ConnectDB()
	require.NoError(t, err)
	defer db.Close()
	err = a.Stat(ctx, db)
	require.Error(t, err)
}

func TestDocuments_Stat(t *testing.T) {
	t.Parallel()
	a := html3.Documents{}
	ctx := context.TODO()
	err := a.Stat(ctx, nil)
	require.ErrorIs(t, err, html3.ErrDB)
	db, err := postgres.ConnectDB()
	require.NoError(t, err)
	defer db.Close()
	err = a.Stat(ctx, db)
	require.Error(t, err)
}
