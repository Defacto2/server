package helper_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/helper"
	"github.com/stretchr/testify/assert"
)

func TestByteCount(t *testing.T) {
	s := helper.ByteCount(0)
	assert.Equal(t, "0B", s)
	s = helper.ByteCount(1023)
	assert.Equal(t, "1023B", s)
	s = helper.ByteCount(1024)
	assert.Equal(t, "1k", s)
	s = helper.ByteCount(-1026)
	assert.Equal(t, "-1026B", s)
	s = helper.ByteCount(1024*1024*1024 - 1)
	assert.Equal(t, "1024M", s)
}

func TestByteCountFloat(t *testing.T) {
	s := helper.ByteCountFloat(0)
	assert.Equal(t, "0 bytes", s)
	s = helper.ByteCountFloat(1023)
	assert.Equal(t, "1 kB", s)
	s = helper.ByteCountFloat(1024)
	assert.Equal(t, "1 kB", s)
	s = helper.ByteCountFloat(-1026)
	assert.Equal(t, "-1026 bytes", s)
	s = helper.ByteCountFloat(1024*1024*1024 - 1)
	assert.Equal(t, "1.1 GB", s)
	s = helper.ByteCountFloat(1024*1024*1024*1024 - 1)
	assert.Equal(t, "1.1 TB", s)
	s = helper.ByteCountFloat(1024*1024*1024*1024*1024 - 1)
	assert.Equal(t, "1.1 PB", s)
}

func TestCapitalize(t *testing.T) {
	s := helper.Capitalize("")
	assert.Equal(t, "", s)
	s = helper.Capitalize("hello")
	assert.Equal(t, "Hello", s)
	s = helper.Capitalize("hello world")
	assert.Equal(t, "Hello world", s)
	s = helper.Capitalize(strings.ToUpper("hello world!"))
	assert.Equal(t, "Hello WORLD!", s)
}

func TestDeleteDupe(t *testing.T) {
	s := helper.DeleteDupe(nil...)
	assert.EqualValues(t, []string{}, s)
	s = helper.DeleteDupe([]string{"a"}...)
	assert.EqualValues(t, []string{"a"}, s)
	s = helper.DeleteDupe([]string{"a", "b", "abcde"}...)
	assert.EqualValues(t, []string{"a", "abcde", "b"}, s) // sorted
	s = helper.DeleteDupe([]string{"a", "b", "a"}...)
	assert.EqualValues(t, []string{"a", "b"}, s)
}

func TestFmtSlice(t *testing.T) {
	s := helper.FmtSlice("")
	assert.Equal(t, "", s)
	s = helper.FmtSlice("a")
	assert.Equal(t, "A", s)
	s = helper.FmtSlice("a,b, abcde")
	assert.Equal(t, "A, B, Abcde", s)
	s = helper.FmtSlice("a , b , abcde")
	assert.Equal(t, "A, B, Abcde", s)
}

func TestLastChr(t *testing.T) {
	tests := []struct {
		s    string
		want string
	}{
		{"", ""},
		{"abc", "c"},
		{"012", "2"},
		{"abc ", "c"},
		{"😃💁 People · 🐻🌻 Animals · 🎷", "🎷"},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			if got := helper.LastChr(tt.s); got != tt.want {
				t.Errorf("LastChr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMaxLineLength(t *testing.T) {
	i := helper.MaxLineLength("")
	assert.Equal(t, 0, i)
	i = helper.MaxLineLength("a")
	assert.Equal(t, 1, i)
	i = helper.MaxLineLength("a\nb")
	assert.Equal(t, 1, i)
	i = helper.MaxLineLength("a\nabcdefghijklmnopqrstuvwxyz\nabcde.")
	assert.Equal(t, 26, i)
}

func TestShortMonth(t *testing.T) {
	s := helper.ShortMonth(0)
	assert.Equal(t, "", s)
	s = helper.ShortMonth(1)
	assert.Equal(t, "Jan", s)
	s = helper.ShortMonth(12)
	assert.Equal(t, "Dec", s)
	s = helper.ShortMonth(13)
	assert.Equal(t, "", s)
}

func TestSplitAsSpace(t *testing.T) {
	s := helper.SplitAsSpaces("")
	assert.Equal(t, "", s)
	s = helper.SplitAsSpaces("a")
	assert.Equal(t, "a", s)
	s = helper.SplitAsSpaces("Hello world!")
	assert.Equal(t, "Hello world!", s)
	s = helper.SplitAsSpaces("HTTP Dir")
	assert.Equal(t, "HTTP Directory", s)
}

func TestTruncFilename(t *testing.T) {
	const fn = "one_two-three.file"
	type args struct {
		w    int
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{}, ""},
		{"zero", args{0, fn}, ""},
		{"ext", args{5, fn}, ".file"},
		{"too short", args{4, fn}, ".file"},
		{"short", args{14, fn}, "one_two-..file"},
		{"too short 2", args{6, "file"}, "file"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := helper.TruncFilename(tt.args.w, tt.args.name); got != tt.want {
				t.Errorf("TruncFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrimRoundBraket(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"empty", "", ""},
		{"hi", "Hello world", "Hello world"},
		{"okay", "Hello world (Hi!)", "Hello world"},
		{"search", "Razor 1911 (RZR, Razor)", "Razor 1911"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, helper.TrimRoundBraket(tt.s))
		})
	}
}

func TestTrimPunct(t *testing.T) {
	tests := []struct {
		s    string
		want string
	}{
		{"", ""},
		{"abc", "abc"},
		{"abc.", "abc"},
		{"abc?", "abc"},
		{"📙", "📙"},
		{"📙!?!", "📙"},
		{"📙 (a book)", "📙 (a book)"},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			if got := helper.TrimPunct(tt.s); got != tt.want {
				t.Errorf("TrimPunct() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYears(t *testing.T) {
	s := helper.Years(0, 0)
	assert.Equal(t, "the year 0", s)
	s = helper.Years(1990, 1991)
	assert.Equal(t, "the years 1990 and 1991", s)
	s = helper.Years(1990, 2000)
	assert.Equal(t, "the years 1990 - 2000", s)
}

// https://defacto2.net/f/ab27b2e

func TestDeobfuscateURL(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
		want   int
	}{
		{"record", "https://defacto2.net/f/ab27b2e", 13526},
		{"download", "https://defacto2.net/d/ab27b2e", 13526},
		{"query", "https://defacto2.net/f/ab27b2e?blahblahblah", 13526},
		{"typo", "https://defacto2.net/f/ab27b2", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, helper.DeobfuscateURL(tt.rawURL))
		})
	}
}

func TestSlug(t *testing.T) {
	tests := []struct {
		name      string
		expect    string
		assertion assert.ComparisonAssertionFunc
	}{
		{"the-group", "the_group", assert.Equal},
		{"group1, group2", "group1*group2", assert.Equal},
		{"group1 & group2", "group1-ampersand-group2", assert.Equal},
		{"group 1, group 2", "group-1*group-2", assert.Equal},
		{"GROUP 👾", "group", assert.Equal},
		{"Mooñpeople", "moonpeople", assert.Equal},
	}
	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			tt.assertion(t, tt.expect, helper.Slug(tt.name))
		})
	}
}

func TestPageCount(t *testing.T) {
	type args struct {
		sum   int
		limit int
	}
	tests := []struct {
		name string
		args args
		want uint
	}{
		{"-1", args{-1, -1}, 0},
		{"0", args{0, 500}, 0},
		{"1", args{1, 500}, 1},
		{"500", args{500, 750}, 1},
		{"750", args{750, 500}, 2},
		{"1k", args{1000, 500}, 2},
		{"1001", args{1001, 500}, 3},
		{"want 10", args{1000, 100}, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, helper.PageCount(tt.args.sum, tt.args.limit))
		})
	}
}

func TestObfuscates(t *testing.T) {
	keys := []int{1, 1000, 1236346, -123, 0}
	for _, key := range keys {
		s := helper.ObfuscateID(int64(key))
		assert.Equal(t, key, helper.DeobfuscateID(s))
	}
}

func TestSearchTerm(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty", "", []string{}},
		{"spaces", "   ", []string{""}},
		{"one", "one", []string{"one"}},
		{"two", "one two", []string{"one two"}},
		{"three", "one two three", []string{"one two three"}},
		{"quotes", `"one two" three`, []string{"\"one two\" three"}},
		{"two", "one,two", []string{"one", "two"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, helper.SearchTerm(tt.input))
		})
	}
}
