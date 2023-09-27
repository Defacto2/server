package helper_test

import (
	"embed"
	"testing"

	"github.com/Defacto2/server/pkg/helper"
	"github.com/stretchr/testify/assert"
)

func TestIntegrity(t *testing.T) {
	t.Parallel()
	x, err := helper.Integrity("", embed.FS{})
	assert.Error(t, err)
	assert.Empty(t, x)
}

func TestIntegrityBytes(t *testing.T) {
	t.Parallel()
	x := helper.IntegrityBytes(nil)
	assert.Equal(t, "sha384-OLBgp1GsljhM2TJ+sbHjaiH9txEUvgdDTAzHv2P24donTt6/529l+9Ua0vFImLlb", x)
	x = helper.IntegrityBytes([]byte("hello world"))
	assert.Equal(t, "sha384-/b2OdaZ/KfcBpOBAOF4uI5hjA+oQI5IRr5B/y7g1eLPkF8txzmRu/QgZ3YwIjeG9", x)
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
	}
	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			tt.assertion(t, tt.expect, helper.Slug(tt.name))
		})
	}
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
