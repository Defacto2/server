package model_test

import (
	"testing"
	"time"

	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
)

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
			tt.assertion(t, tt.expect, model.Published(&f))
		})
	}
}

func TestCreated(t *testing.T) {
	loc := time.Local
	tests := []struct {
		arg       time.Time
		expect    string
		assertion assert.ComparisonAssertionFunc
	}{
		{time.Time{}, "", assert.Equal},
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
			tt.assertion(t, tt.expect, model.Created(&f))
		})
	}
}
