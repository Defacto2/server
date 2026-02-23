package msdos_test

import (
	"testing"

	"github.com/Defacto2/server/handler/jsdos/msdos"
	"github.com/nalgeon/be"
)

func TestDirName(t *testing.T) {
	t.Parallel()
	s := msdos.DirName("")
	be.Equal(t, s, "")
	s = msdos.DirName("dirname.xyz")
	be.Equal(t, "DIRNAME.XYZ", s)
	s = msdos.DirName("résumés-99")
	be.Equal(t, "RESUMES-", s)
	s = msdos.DirName("résumé99.doc")
	be.Equal(t, "RESUME99.DOC", s)
}

func TestRename(t *testing.T) {
	t.Parallel()
	s := msdos.Rename("")
	be.Equal(t, s, "")
	s = msdos.Rename("filename.xyz")
	be.Equal(t, "FILENAME.XYZ", s)
	s = msdos.Rename("résumé-01.zip")
	be.Equal(t, "RESUME-01.ZIP", s)
	s = msdos.Rename("résumé 01.zip")
	be.Equal(t, "RESUME_01.ZIP", s)
	s = msdos.Rename("A@cd#F$H!.D0C")
	be.Equal(t, "A@CD#F$H!.D0C", s)
	s = msdos.Rename("Γεåd.më")
	be.Equal(t, "XXAD.ME", s)
	s = msdos.Rename("Γεåd.më.")
	be.Equal(t, "XXAD.MEX", s)
	s = msdos.Rename("Γεåd.më.7zip")
	be.Equal(t, "XXADXMEX7ZIP", s)
	s = msdos.Rename(".HIDDEN")
	be.Equal(t, "XHIDDEN", s)
	s = msdos.Rename(".TXT")
	be.Equal(t, "XTXT", s)
}

func TestTruncate(t *testing.T) {
	t.Parallel()
	s := msdos.Truncate("")
	be.Equal(t, s, "")
	s = msdos.Truncate("filename")
	be.Equal(t, "filename", s)
	s = msdos.Truncate("filename1")
	be.Equal(t, "filena~1", s)
	s = msdos.Truncate("filename12")
	be.Equal(t, "filena~1", s)
	s = msdos.Truncate("filename123")
	be.Equal(t, "filena~1", s)
	s = msdos.Truncate("filename.exe")
	be.Equal(t, "filename.exe", s)
	s = msdos.Truncate("filename.binary")
	be.Equal(t, "filename.bin", s)
	s = msdos.Truncate("file_name.exe")
	be.Equal(t, "file_n~1.exe", s)
	s = msdos.Truncate("my backup collection.7zip")
	be.Equal(t, "my bac~1.7zi", s)
	s = msdos.Truncate("filename.zip.exe")
	be.Equal(t, "filena~1.exe", s)
}
