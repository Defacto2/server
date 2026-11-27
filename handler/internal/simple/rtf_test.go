package simple_test

import (
	"testing"

	"github.com/Defacto2/server/handler/internal/simple"
	"github.com/nalgeon/be"
)

func TestRTF(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []byte
		want  bool
	}{
		{"nil slice", nil, false},
		{"empty", []byte(""), false},
		{"short", []byte("short"), false},
		{"nine bytes", []byte(`{\rtf1\an`), false},
		{"ten bytes", []byte(`{\rtf1\ans`), true},
		{"non-digit version", []byte(`{\rtfa\ansi\deff0`), false},
		{"plain text", []byte("This is plain text with {braces}"), false},
		{"RTF header", []byte(`{\rtf1\ansi\deff0`), true},
		{"actual RTF header", []byte(`{\rtf1\ansi\deff0\deftab720{\fonttbl{\f0\fnil MS Sans Serif;}}`), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			be.Equal(t, simple.RTF(tt.input), tt.want)
		})
	}
}

func TestStripRTF(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{"nil slice", nil, ""},
		{"empty", []byte(""), ""},
		{
			"plain text unchanged",
			[]byte("This is plain text with {braces} and \\ backslashes"),
			"This is plain text with {braces} and \\ backslashes",
		},
		{
			"RTF with formatting",
			[]byte(`{\rtf1\ansi\deff0{\fonttbl{\f0\fnil MS Sans Serif;}}{\colortbl\red0\green0\blue0;}\deflang1033\pard\plain\f0\fs20 This is formatted text.\par}`),
			"This is formatted text.",
		},
		{
			"RTF with colors",
			[]byte(`{\rtf1\ansi\deff0{\colortbl\red255\green0\blue0;}\cf1 Red text\cf0 normal text}`),
			"Red text normal text",
		},
		{
			"negative control-word param",
			[]byte(`{\rtf1\ansi\li-200 Indented text.\par}`),
			"Indented text.",
		},
		{
			"preserves single-letter words and prose font names",
			[]byte(`{\rtf1\ansi\deff0{\fonttbl{\f0\fnil Custom Font;}{\f1\fswiss System;}}\pard\plain\f0\fs20 I am a System coder; A + B.\par}`),
			"I am a System coder; A + B.",
		},
		{
			"actual RTF from file",
			[]byte(`{\rtf1\ansi\deff0\deftab720{\fonttbl{\f0\fnil MS Sans Serif;}{\f1\fnil\fcharset2 Symbol;}{\f2\fswiss\fprq2 System;}{\f3\fnil Times New Roman;}}
{\colortbl\red0\green0\blue0;\red255\green0\blue0;\red0\green0\blue128;}
\deflang1033\pard\plain\f3\fs20  \plain\f3\fs20\cf1 prestige blows giant monkey goats    
\par  
\par  \plain\f3\fs20\cf2 Razor rules!`),
			"prestige blows giant monkey goats\nRazor rules!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			be.Equal(t, string(simple.StripRTF(tt.input)), tt.want)
		})
	}
}
