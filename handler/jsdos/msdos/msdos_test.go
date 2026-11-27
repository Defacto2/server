package msdos_test

import (
	"testing"

	"github.com/Defacto2/server/handler/jsdos/msdos"
	"github.com/nalgeon/be"
)

func TestDirName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"valid", "dirname.xyz", "DIRNAME.XYZ"},
		{"long accented", "résumés-99", "RESUMES-"},
		{"accented with ext", "résumé99.doc", "RESUME99.DOC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := msdos.DirName(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestRename(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"basic", "filename.xyz", "FILENAME.XYZ"},
		{"accented hyphen", "résumé-01.zip", "RESUME-01.ZIP"},
		{"accented space", "résumé 01.zip", "RESUME_01.ZIP"},
		{"allowed specials", "A@cd#F$H!.D0C", "A@CD#F$H!.D0C"},
		{"greek and accents", "Γεåd.më", "XXAD.ME"},
		{"trailing dot", "Γεåd.më.", "XXAD.MEX"},
		{"long ext", "Γεåd.më.7zip", "XXADXMEX7ZIP"},
		{"dotfile hidden", ".HIDDEN", "XHIDDEN"},
		{"dotfile ext", ".TXT", "XTXT"},
		{"multiple dots", "a.b.c", "AXB.C"},
		{"adjacent dots", "a..b", "AX.B"},
		{"non-ascii latin", "ØÆ².txt", "XXX.TXT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := msdos.Rename(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestTruncate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"exact 8", "filename", "filename"},
		{"9 chars", "filename1", "filena~1"},
		{"10 chars", "filename12", "filena~1"},
		{"11 chars", "filename123", "filena~1"},
		{"8.3", "filename.exe", "filename.exe"},
		{"long ext", "filename.binary", "filename.bin"},
		{"long base with ext", "file_name.exe", "file_n~1.exe"},
		{"long base and ext", "my backup collection.7zip", "my bac~1.7zi"},
		{"double ext", "filename.zip.exe", "filena~1.exe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := msdos.Truncate(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}
