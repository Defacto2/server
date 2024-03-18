package model_test

import (
	"testing"

	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
)

func TestPublishedFmt(t *testing.T) {
	t.Parallel()
	s := model.PublishedFmt(nil)
	assert.Equal(t, "error, no file model", s)

	ms := models.File{}
	s = model.PublishedFmt(&ms)
	assert.Equal(t, "unknown date", s)

	ms.DateIssuedYear = null.Int16From(1999)
	ms.DateIssuedMonth = null.Int16From(12)
	s = model.PublishedFmt(&ms)
	assert.Equal(t, "1999 December", s)
}

func TestDosPaths(t *testing.T) {
	t.Parallel()
	s := model.DosPaths("")
	assert.Empty(t, s)

	x := "filename.zip\nreadme.txt\nrunme.bat\napp.com\ndata.dat"
	s = model.DosPaths(x)
	assert.Equal(t, 5, len(s))

	x = "filename.zip\rreadme.txt\nrunme.bat\r\nAPP.COM\ndata.dat"
	s = model.DosPaths(x)
	assert.Equal(t, 5, len(s))
}

func TestDosBins(t *testing.T) {
	t.Parallel()
	bins := model.DosBins()
	assert.Empty(t, bins)

	x := "filename.zip\nreadme.xxx\nrunme.xxx\napp.xxx\ndata.dat"
	p := model.DosPaths(x)
	bins = model.DosBins(p...)
	assert.Empty(t, bins)

	x = "filename.zip\rreadme.txt\nrunme.bat\r\nAPP.COM\ndata.dat"
	p = model.DosPaths(x)
	bins = model.DosBins(p...)
	assert.Equal(t, 2, len(bins))
}

func TestDosMatch(t *testing.T) {
	t.Parallel()
	s := model.DosMatch("", "")
	assert.Empty(t, s)

	x := "filename.zip\nreadme.xxx\nrunme.xxx\napp.xxx\ndata.dat"
	p := model.DosPaths(x)
	s = model.DosMatch("filename.zip", p...)
	assert.Empty(t, s)

	x += "\nFILENAME.EXE\nfilename.xxx"
	p = model.DosPaths(x)
	s = model.DosMatch("filename.zip", p...)
	assert.Equal(t, "FILENAME.EXE", s)

	x = "FILENAME.COM\n" + x
	p = model.DosPaths(x)
	s = model.DosMatch("filename.zip", p...)
	assert.Equal(t, "FILENAME.EXE", s)
}

func TestDosBin(t *testing.T) {
	t.Parallel()
	s := model.DosBin()
	assert.Empty(t, s)

	x := "filename.zip\nreadme.xxx\nrunme.xxx\napp.xxx\ndata.dat"
	p := model.DosPaths(x)
	s = model.DosBin(p...)
	assert.Empty(t, s)

	x += "\nfilename.exe\nfilename.xxx"
	p = model.DosPaths(x)
	s = model.DosBin(p...)
	assert.Equal(t, "filename.exe", s)

	x = "FILENAME.COM\n" + x
	p = model.DosPaths(x)
	s = model.DosBin(p...)
	assert.Equal(t, "FILENAME.COM", s)

	x += "\nrunme.bat"
	p = model.DosPaths(x)
	s = model.DosBin(p...)
	assert.Equal(t, "runme.bat", s)
}

func TestDosBinary(t *testing.T) {

	example := "readme.txt\nRUN.BAT\napp.com\ndata.dat"

	t.Parallel()
	s := model.DosBinary("", "")
	assert.Equal(t, "", s)

	s = model.DosBinary("filename", "")
	assert.Equal(t, "filename", s)

	s = model.DosBinary("filename.xyz", "")
	assert.Equal(t, "filename.xyz", s)

	s = model.DosBinary("filename.zip", "zipcontent")
	assert.Equal(t, "", s)

	s = model.DosBinary("filename.zip", "readme.txt")
	assert.Equal(t, "", s)

	s = model.DosBinary("filename.zip", example)
	assert.Equal(t, "RUN.BAT", s)

	s = model.DosBinary("filename.zip", example+"\n"+"filename.exe")
	assert.Equal(t, "filename.exe", s)
}
