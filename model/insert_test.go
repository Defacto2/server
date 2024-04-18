package model_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/volatiletech/null/v8"
)

func TestValidMagic(t *testing.T) {
	tests := []struct {
		name      string
		mediatype string
		expected  null.String
	}{
		{
			name:      "Empty mediatype",
			mediatype: "",
			expected:  null.String{String: "", Valid: false},
		},
		{
			name:      "Invalid mediatype",
			mediatype: "application/json",
			expected:  null.StringFrom("application/json"),
		},
		{
			name:      "Valid mediatype",
			mediatype: "image/jpeg",
			expected:  null.StringFrom("image/jpeg"),
		},
		{
			name:      "Valid mediatype",
			mediatype: "application/x-msdos-program",
			expected:  null.StringFrom("application/x-msdos-program"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.ValidMagic(tt.mediatype)
			if result != tt.expected {
				t.Errorf("Expected %v, but got %v", tt.expected, result)
			}
		})
	}
}

func TestDateIssue(t *testing.T) {
	tests := []struct {
		name     string
		y        string
		m        string
		d        string
		expected struct {
			year  null.Int16
			month null.Int16
			day   null.Int16
		}
	}{
		{
			name: "Valid date",
			y:    "2022",
			m:    "12",
			d:    "31",
			expected: struct {
				year  null.Int16
				month null.Int16
				day   null.Int16
			}{
				year:  null.Int16From(2022),
				month: null.Int16From(12),
				day:   null.Int16From(31),
			},
		},
		{
			name: "Invalid year",
			y:    "abcd",
			m:    "12",
			d:    "31",
			expected: struct {
				year  null.Int16
				month null.Int16
				day   null.Int16
			}{
				year:  null.Int16{},
				month: null.Int16From(12),
				day:   null.Int16From(31),
			},
		},
		{
			name: "Invalid month",
			y:    "2022",
			m:    "abcd",
			d:    "31",
			expected: struct {
				year  null.Int16
				month null.Int16
				day   null.Int16
			}{
				year:  null.Int16From(2022),
				month: null.Int16{},
				day:   null.Int16From(31),
			},
		},
		{
			name: "Invalid day",
			y:    "2022",
			m:    "12",
			d:    "abcd",
			expected: struct {
				year  null.Int16
				month null.Int16
				day   null.Int16
			}{
				year:  null.Int16From(2022),
				month: null.Int16From(12),
				day:   null.Int16{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			year, month, day := model.DateIssue(tt.y, tt.m, tt.d)
			if year != tt.expected.year || month != tt.expected.month || day != tt.expected.day {
				t.Errorf("Expected year=%v, month=%v, day=%v, but got year=%v, month=%v, day=%v",
					tt.expected.year, tt.expected.month, tt.expected.day, year, month, day)
			}
		})
	}
}

func TestValidD(t *testing.T) {
	tests := []struct {
		name     string
		d        int16
		expected null.Int16
	}{
		{
			name:     "Valid day",
			d:        15,
			expected: null.Int16From(15),
		},
		{
			name:     "Invalid day (less than 1)",
			d:        0,
			expected: null.Int16{Int16: 0, Valid: false},
		},
		{
			name:     "Invalid day (greater than 31)",
			d:        32,
			expected: null.Int16{Int16: 0, Valid: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.ValidD(tt.d)
			if result != tt.expected {
				t.Errorf("Expected %v, but got %v", tt.expected, result)
			}
		})
	}
}

func TestValidM(t *testing.T) {
	tests := []struct {
		name     string
		m        int16
		expected null.Int16
	}{
		{
			name:     "Valid month",
			m:        6,
			expected: null.Int16From(6),
		},
		{
			name:     "Invalid month (less than 1)",
			m:        0,
			expected: null.Int16{Int16: 0, Valid: false},
		},
		{
			name:     "Invalid month (greater than 12)",
			m:        13,
			expected: null.Int16{Int16: 0, Valid: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.ValidM(tt.m)
			if result != tt.expected {
				t.Errorf("Expected %v, but got %v", tt.expected, result)
			}
		})
	}
}

func TestValidY(t *testing.T) {
	tests := []struct {
		name     string
		y        int16
		expected null.Int16
	}{
		{
			name:     "Valid year",
			y:        2022,
			expected: null.Int16From(2022),
		},
		{
			name:     "Invalid year (less than EpochYear)",
			y:        1899,
			expected: null.Int16{Int16: 0, Valid: false},
		},
		{
			name:     "Invalid year (greater than current year)",
			y:        3000,
			expected: null.Int16{Int16: 0, Valid: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.ValidY(tt.y)
			if result != tt.expected {
				t.Errorf("Expected %v, but got %v", tt.expected, result)
			}
		})
	}
}

func TestTrimShort(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "String shorter than limit",
			input:    "Hello",
			expected: "Hello",
		},
		{
			name:     "String equal to limit",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "String longer than limit",
			input:    strings.Repeat("1234567890", 21),
			expected: strings.Repeat("1234567890", 10),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.TrimShort(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %v, but got %v", tt.expected, result)
			}
		})
	}
}

func TestValidReleasers(t *testing.T) {
	tests := []struct {
		name     string
		s1       string
		s2       string
		expected struct {
			x1 null.String
			x2 null.String
		}
	}{
		{
			name: "Empty strings",
			s1:   "",
			s2:   "",
			expected: struct {
				x1 null.String
				x2 null.String
			}{
				x1: null.String{String: "", Valid: false},
				x2: null.String{String: "", Valid: false},
			},
		},
		{
			name: "Valid strings",
			s1:   "Releaser 1",
			s2:   "Releaser 2",
			expected: struct {
				x1 null.String
				x2 null.String
			}{
				x1: null.StringFrom("RELEASER 1"),
				x2: null.StringFrom("RELEASER 2"),
			},
		},
		{
			name: "Trimmed strings",
			s1:   "   Releaser 1   ",
			s2:   "   Releaser 2   ",
			expected: struct {
				x1 null.String
				x2 null.String
			}{
				x1: null.StringFrom("RELEASER 1"),
				x2: null.StringFrom("RELEASER 2"),
			},
		},
		{
			name: "Invalid strings",
			s1:   "   ",
			s2:   "   ",
			expected: struct {
				x1 null.String
				x2 null.String
			}{
				x1: null.String{String: "", Valid: false},
				x2: null.String{String: "", Valid: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x1, x2 := model.ValidReleasers(tt.s1, tt.s2)
			if x1 != tt.expected.x1 || x2 != tt.expected.x2 {
				t.Errorf("Expected x1=%v, x2=%v, but got x1=%v, x2=%v", tt.expected.x1, tt.expected.x2, x1, x2)
			}
		})
	}
}

func TestValidTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected null.String
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: null.String{String: "", Valid: false},
		},
		{
			name:     "String with spaces",
			input:    "   ",
			expected: null.String{String: "", Valid: false},
		},
		{
			name:     "String shorter than limit",
			input:    "Hello",
			expected: null.StringFrom("Hello"),
		},
		{
			name:     "String equal to limit",
			input:    "Hello World",
			expected: null.StringFrom("Hello World"),
		},
		{
			name:     "String longer than limit",
			input:    strings.Repeat("1234567890", 21),
			expected: null.StringFrom(strings.Repeat("1234567890", 10)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.ValidTitle(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %v, but got %v", tt.expected, result)
			}
		})
	}
}

func TestValidYouTube(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected null.String
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: null.String{String: "", Valid: false},
		},
		{
			name:     "Invalid length",
			input:    "1234567890",
			expected: null.String{String: "", Valid: false},
		},
		{
			name:     "Invalid characters",
			input:    "1234567890!",
			expected: null.String{String: "", Valid: false},
		},
		{
			name:     "Valid YouTube ID",
			input:    "abcdefghijk",
			expected: null.String{String: "abcdefghijk", Valid: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := model.ValidYouTube(tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %v, but got %v", tt.expected, result)
			}
		})
	}
}

func TestValidFilesize(t *testing.T) {
	tests := []struct {
		name     string
		size     string
		expected int64
		errNil   bool
	}{
		{
			name:     "Empty size",
			size:     "",
			expected: 0,
			errNil:   true,
		},
		{
			name:     "Valid size",
			size:     "1024",
			expected: 1024,
			errNil:   true,
		},
		{
			name:     "Invalid size",
			size:     "abc",
			expected: 0,
			errNil:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := model.ValidFilesize(tt.size)
			if (err == nil) != tt.errNil {
				t.Errorf("Expected error, but got %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %v, but got %v", tt.expected, result)
			}
		})
	}
}

func TestValidIntegrity(t *testing.T) {
	tests := []struct {
		name      string
		integrity string
		expected  null.String
	}{
		{
			name:      "Empty integrity",
			integrity: "",
			expected:  null.String{String: "", Valid: false},
		},
		{
			name:      "Invalid integrity (incorrect length)",
			integrity: "1234567890",
			expected:  null.String{String: "", Valid: false},
		},
		{
			name:      "Invalid integrity (non-hex characters)",
			integrity: "g1h2i3j4k5l6m7n8o9p0q1r2s3t4u5v6w7x8y9z0",
			expected:  null.String{String: "", Valid: false},
		},
		{
			name:      "Valid integrity",
			integrity: "2d28edf3bd78230486ad52ae31f13a031a97e3b377d64826965d68174a5815d36f18c5a14394eeeb3cce491d356c8689",
			expected:  null.StringFrom("2d28edf3bd78230486ad52ae31f13a031a97e3b377d64826965d68174a5815d36f18c5a14394eeeb3cce491d356c8689"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.ValidIntegrity(tt.integrity)
			if result != tt.expected {
				t.Errorf("Expected %v, but got %v", tt.expected, result)
			}
		})
	}
}

func TestValidLastMod(t *testing.T) {
	tests := []struct {
		name     string
		lastmod  string
		expected null.Time
	}{
		{
			name:     "Empty lastmod",
			lastmod:  "",
			expected: null.Time{Time: time.Time{}, Valid: false},
		},
		{
			name:     "Invalid lastmod (not a number)",
			lastmod:  "abc",
			expected: null.Time{Time: time.Time{}, Valid: false},
		},
		{
			name:     "Invalid lastmod (future date)",
			lastmod:  "2000000000000",
			expected: null.Time{Time: time.Time{}, Valid: false},
		},
		{
			name:     "Invalid lastmod (before EpochYear)",
			lastmod:  fmt.Sprintf("%d", time.Date(1979, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano()),
			expected: null.Time{Time: time.Time{}, Valid: false},
		},
		{
			name:     "Valid lastmod",
			lastmod:  "1640995200000",
			expected: null.TimeFrom(time.Unix(1640995200, 0)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.ValidLastMod(tt.lastmod)
			if result != tt.expected {
				t.Errorf("Expected %v, but got %v", tt.expected, result)
			}
		})
	}
}

func TestValidPlatform(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		expected null.String
	}{
		{
			name:     "Empty platform",
			platform: "",
			expected: null.String{String: "", Valid: false},
		},
		{
			name:     "Valid platform",
			platform: tags.Windows.String(),
			expected: null.StringFrom(tags.Windows.String()),
		},
		{
			name:     "Invalid platform",
			platform: tags.Intro.String(),
			expected: null.String{String: "", Valid: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.ValidPlatform(tt.platform)
			if result != tt.expected {
				t.Errorf("Expected %v, but got %v", tt.expected, result)
			}
		})
	}
}

func TestValidSection(t *testing.T) {
	tests := []struct {
		name     string
		section  string
		expected null.String
	}{
		{
			name:     "Empty section",
			section:  "",
			expected: null.String{String: "", Valid: false},
		},
		{
			name:     "Valid category",
			section:  tags.Intro.String(),
			expected: null.StringFrom(tags.Intro.String()),
		},
		{
			name:     "Invalid section",
			section:  "Invalid",
			expected: null.String{String: "", Valid: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.ValidSection(tt.section)
			if result != tt.expected {
				t.Errorf("Expected %v, but got %v", tt.expected, result)
			}
		})
	}
}
