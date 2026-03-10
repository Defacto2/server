package simple_test

import (
	"testing"

	"github.com/Defacto2/server/handler/app/internal/simple"
)

func TestIsRTF(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "RTF header detected",
			input:    `{\rtf1\ansi\deff0`,
			expected: true,
		},
		{
			name:     "Non-RTF content",
			input:    "This is plain text with {braces}",
			expected: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "Short string",
			input:    "short",
			expected: false,
		},
		{
			name:     "Actual RTF from file",
			input:    `{\rtf1\ansi\deff0\deftab720{\fonttbl{\f0\fnil MS Sans Serif;}}`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := simple.IsRTF([]byte(tt.input))
			if result != tt.expected {
				t.Errorf("IsRTF() = %v, want %v for input %q", result, tt.expected, tt.input)
			}
		})
	}
}

func TestStripRTF(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "RTF with formatting",
			input:    `{\rtf1\ansi\deff0{\fonttbl{\f0\fnil MS Sans Serif;}}{\colortbl\red0\green0\blue0;}\deflang1033\pard\plain\f0\fs20 This is formatted text.\par}`,
			expected: "This is formatted text.",
		},
		{
			name:     "Plain text unchanged",
			input:    "This is plain text with {braces} and \\ backslashes",
			expected: "This is plain text with {braces} and \\ backslashes",
		},
		{
			name:     "RTF with colors",
			input:    `{\rtf1\ansi\deff0{\colortbl\red255\green0\blue0;}\cf1 Red text\cf0 normal text}`,
			expected: "Red text normal text",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name: "Actual RTF from file",
			input: `{\rtf1\ansi\deff0\deftab720{\fonttbl{\f0\fnil MS Sans Serif;}{\f1\fnil\fcharset2 Symbol;}{\f2\fswiss\fprq2 System;}{\f3\fnil Times New Roman;}}
{\colortbl\red0\green0\blue0;\red255\green0\blue0;\red0\green0\blue128;}
\deflang1033\pard\plain\f3\fs20  \plain\f3\fs20\cf1 prestige blows giant monkey goats    
\par  
\par  \plain\f3\fs20\cf2 Razor rules!`,
			expected: "prestige blows giant monkey goats\nRazor rules!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := simple.StripRTF(tt.input)
			if result != tt.expected {
				t.Errorf("StripRTF() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestStripRTFBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "RTF bytes",
			input:    []byte(`{\rtf1\ansi\deff0{\fonttbl{\f0\fnil MS Sans Serif;}}{\colortbl\red0\green0\blue0;}\deflang1033\pard\plain\f0\fs20 This is formatted text.\par}`),
			expected: []byte("This is formatted text."),
		},
		{
			name:     "Plain text bytes unchanged",
			input:    []byte("This is plain text with {braces}"),
			expected: []byte("This is plain text with {braces}"),
		},
		{
			name:     "Empty bytes",
			input:    []byte(""),
			expected: []byte(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := simple.StripRTFBytes(tt.input)
			if string(result) != string(tt.expected) {
				t.Errorf("StripRTFBytes() = %q, want %q", string(result), string(tt.expected))
			}
		})
	}
}
