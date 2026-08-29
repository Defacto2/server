package model_test

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/null/v8"
	"github.com/nalgeon/be"
)

type vstrings []struct {
	input string
	want  bool
	wantS string
}

func VStrings(t *testing.T, tests vstrings, fn func(string) null.String) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := fn(tt.input)
			be.Equal(t, got.Valid, tt.want)
			be.Equal(t, got.String, tt.wantS)
		})
	}
}

type vtimes []struct {
	input string
	want  bool
}

func VTimes(t *testing.T, tests vtimes, fn func(string) null.Time) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := fn(tt.input)
			be.Equal(t, got.Valid, tt.want)
			if got.Valid {
				be.True(t, time.Now().UnixMilli() > got.Time.UnixMilli())
			}
		})
	}
}

type vdates []struct {
	name   string
	year   string
	month  string
	day    string
	validY bool
	validM bool
	validD bool
	wantY  int16
	wantM  int16
	wantD  int16
}

func VDates(t *testing.T, tests vdates,
	fn func(string, string, string) (null.Int16, null.Int16, null.Int16),
) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			goty, gotm, gotd := fn(tt.year, tt.month, tt.day)
			be.Equal(t, goty.Valid, tt.validY)
			be.Equal(t, gotm.Valid, tt.validM)
			be.Equal(t, gotd.Valid, tt.validD)
			be.Equal(t, goty.Int16, tt.wantY)
			be.Equal(t, gotm.Int16, tt.wantM)
			be.Equal(t, gotd.Int16, tt.wantD)
		})
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	got := model.Validate(nil)
	be.Err(t, got)

	blank := &models.File{}
	got = model.Validate(blank)
	be.Err(t, got)

	blank.Platform = null.StringFrom("3D0")
	blank.Section = null.StringFrom("beepboop")
	got = model.Validate(blank)
	be.Err(t, got)

	blank.Section = null.StringFrom(tags.Mag.String())
	got = model.Validate(blank)
	be.Err(t, got)
}

func TestValidDateIssue(t *testing.T) {
	t.Parallel()

	tests := vdates{
		{
			"null",
			"", "", "", false, false, false, 0, 0, 0,
		},
		{
			"invalid",
			"9999999", "00000", "ABVSDF", false, false, false, 0, 0, 0,
		},
		{
			"1950",
			"1950", "", "", false, false, false, 0, 0, 0,
		},
		{
			"1980",
			"1980", "", "", true, false, false, 1980, 0, 0,
		},
		{
			"2nd Jan 1981",
			"1981", "1", "2", true, true, true, 1981, 1, 2,
		},
	}
	VDates(t, tests, model.ValidDateIssue)
}

func TestValidFilename(t *testing.T) {
	t.Parallel()

	name := ""
	r := model.ValidFilename(name)
	be.True(t, !r.Valid)

	name = "somefile.txt"
	r = model.ValidFilename(name)
	be.True(t, r.Valid)
	be.Equal(t, name, r.String)

	name = strings.Repeat("a", model.LongFilename+100)
	r = model.ValidFilename(name)
	be.True(t, r.Valid)
	be.True(t, len(r.String) == model.LongFilename)
}

func TestValidFilesize(t *testing.T) {
	t.Parallel()

	size := ""
	actual0 := null.Int64From(0)
	actual100 := null.Int64From(100)
	actualN100 := null.Int64From(-100)
	i, err := model.ValidFilesize(size)
	be.Err(t, err, nil)
	be.True(t, actual0 != i)

	size = "100"
	i, err = model.ValidFilesize(size)
	be.Err(t, err, nil)
	be.Equal(t, actual100, i)

	size = "-100"
	i, err = model.ValidFilesize(size)
	be.Err(t, err, nil)
	be.Equal(t, actualN100, i)
}

func TestValidIntegrity(t *testing.T) {
	t.Parallel()

	const invalid = "XXXXXX00d8d5467fb8f62c88628b1f30cbfa1d0696a81a78599af01bb913cc726a78f3817adfa557691db9ad1354df6b"
	const valid = "8ac9e700d8d5467fb8f62c88628b1f30cbfa1d0696a81a78599af01bb913cc726a78f3817adfa557691db9ad1354df6b"
	tests := vstrings{
		{"", false, ""},
		{"100", false, ""},
		{"abcde", false, ""},
		{invalid, false, ""},
		{valid, true, valid},
	}
	VStrings(t, tests, model.ValidIntegrity)
}

func TestValidLastMod(t *testing.T) {
	t.Parallel()

	oneHourAgo := time.Now().Add(-time.Hour).UnixMilli()
	lastmodAgo := strconv.FormatInt(oneHourAgo, 10)
	oneHourFromNow := time.Now().Add(time.Hour).UnixMilli()
	lastmodFromNow := strconv.FormatInt(oneHourFromNow, 10)
	tests := vtimes{
		{"", false},
		{"100", false},
		{lastmodAgo, true},
		{lastmodFromNow, false},
	}
	VTimes(t, tests, model.ValidLastMod)
}

func TestValidMagic(t *testing.T) {
	t.Parallel()

	tests := vstrings{
		{"", false, ""},
		{"100", false, ""},
		{"defacto2", false, ""},
		{"Text/HTML", true, "text/html"},
	}
	VStrings(t, tests, model.ValidMagic)
}

func TestValidPlatform(t *testing.T) {
	t.Parallel()

	tests := vstrings{
		{"", false, ""},
		{"100", false, ""},
		{"bbs", false, ""},
		{"Windows", true, "windows"},
	}
	VStrings(t, tests, model.ValidPlatform)
}

func TestValidReleasers(t *testing.T) {
	t.Parallel()

	s1, s2 := "", ""
	r1, r2 := model.ValidReleasers(s1, s2)
	be.True(t, !r1.Valid)
	be.True(t, !r2.Valid)

	s1, s2 = "defacto2", "scene"
	r1, r2 = model.ValidReleasers(s1, s2)
	be.True(t, r1.Valid)
	be.True(t, r2.Valid)
	be.Equal(t, "DEFACTO2", r1.String)
	be.Equal(t, "SCENE", r2.String)

	// test the swapping of empty releasers
	r1, r2 = model.ValidReleasers("", "defacto2")
	be.True(t, r1.Valid)
	be.True(t, !r2.Valid)
	be.Equal(t, "DEFACTO2", r1.String)
	be.Equal(t, r2.String, "")
}

func TestValidSection(t *testing.T) {
	t.Parallel()

	tests := vstrings{
		{"", false, ""},
		{"100", false, ""},
		{"BBS", true, "bbs"},
		{"Windows", false, ""},
	}
	VStrings(t, tests, model.ValidSection)
}

func TestValidString(t *testing.T) {
	t.Parallel()

	const emoji = "😃"
	tests := vstrings{
		{"\n\r   \n", false, ""},
		{"\u00A0", false, ""},
		{"hello world", true, "hello world"},
		{emoji, true, emoji},
	}
	VStrings(t, tests, model.ValidString)
}

func TestValidTitle(t *testing.T) {
	t.Parallel()

	title := strings.Repeat("a", model.ShortLimit+100)
	wantS := strings.Repeat("a", model.ShortLimit)
	tests := vstrings{
		{"", false, ""},
		{"100", true, "100"},
		{"hello world", true, "hello world"},
		{title, true, wantS},
	}
	VStrings(t, tests, model.ValidTitle)
}

func TestValidYouTube(t *testing.T) {
	t.Parallel()

	const invalid = "$6BuDfBIcM!"
	const valid = "62BuDfBIcMo"
	yt := strings.Repeat("x", model.ShortLimit+10)
	tests := vstrings{
		{"", false, ""},
		{yt, false, ""},
		{invalid, false, ""},
		{valid, true, valid},
	}
	VStrings(t, tests, model.ValidYouTube)
}
