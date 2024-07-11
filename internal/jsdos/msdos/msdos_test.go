package msdos_test

import (
	"testing"

	"github.com/Defacto2/server/internal/jsdos/msdos"
	"github.com/stretchr/testify/assert"
)

func TestRename(t *testing.T) {
	t.Parallel()
	s := msdos.Rename("")
	assert.Equal(t, "", s)
	s = msdos.Rename("filename.xyz")
	assert.Equal(t, "FILENAME.XYZ", s)
	s = msdos.Rename("résumé-01.zip")
	assert.Equal(t, "RESUME-01.ZIP", s)
	s = msdos.Rename("résumé 01.zip")
	assert.Equal(t, "RESUME_01.ZIP", s)
	s = msdos.Rename("A@cd#F$H!.D0C")
	assert.Equal(t, "A@CD#F$H!.D0C", s)
	s = msdos.Rename("Γεåd.më")
	assert.Equal(t, "XXAD.ME", s)
	s = msdos.Rename("Γεåd.më.")
	assert.Equal(t, "XXAD.MEX", s)
	s = msdos.Rename("Γεåd.më.7zip")
	assert.Equal(t, "XXADXMEX7ZIP", s)
}

func TestTruncate(t *testing.T) {
	t.Parallel()
	s := msdos.Truncate("")
	assert.Equal(t, "", s)

	s = msdos.Truncate("filename")
	assert.Equal(t, "filename", s)

	s = msdos.Truncate("filename.exe")
	assert.Equal(t, "filename.exe", s)

	s = msdos.Truncate("file_name.exe")
	assert.Equal(t, "file_n~1.exe", s)

	s = msdos.Truncate("my backup collection.7zip")
	assert.Equal(t, "my bac~1.7zi", s)

	s = msdos.Truncate("filename.zip.exe")
	assert.Equal(t, "filena~1.exe", s)
}
