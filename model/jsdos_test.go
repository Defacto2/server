package model_test

import (
	"bytes"
	"testing"

	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/null/v8"
	"github.com/nalgeon/be"
)

func TestJsDosBinary(t *testing.T) {
	t.Parallel()

	got, err := model.JsDosBinary(nil)
	be.Err(t, err)
	be.Equal(t, "", got)

	got, err = model.JsDosBinary(&models.File{})
	be.Err(t, err, nil)
	be.Equal(t, "", got)

	got, err = model.JsDosBinary(&models.File{
		Filename: null.StringFrom("LONGFILENAME.EXE"),
	})
	be.Err(t, err, nil)
	be.Equal(t, "LONGFI~1.EXE", got)

	got, err = model.JsDosBinary(&models.File{
		Filename:       null.StringFrom("demo.zip"),
		FileZipContent: null.StringFrom("readme.txt\nsub/dir/LONGPROGRAM.EXE"),
	})
	be.Err(t, err, nil)
	be.Equal(t, `sub\dir\LONGPR~1.EXE`, got)

	got, err = model.JsDosBinary(&models.File{
		Filename:       null.StringFrom("demo.zip"),
		FileZipContent: null.StringFrom("readme.txt\n" + `SUB\LONGPROGRAM.EXE`),
	})
	be.Err(t, err, nil)
	be.Equal(t, `SUB\LONGPR~1.EXE`, got)
}

func TestJsDosCommand(t *testing.T) {
	t.Parallel()

	got, err := model.JsDosCommand(nil)
	be.Err(t, err)
	be.Equal(t, "", got)

	got, err = model.JsDosCommand(&models.File{
		Filename:        null.StringFrom("demo.zip"),
		FileZipContent:  null.StringFrom("APP.EXE"),
		DoseeRunProgram: null.StringFrom("CUSTOM.EXE"),
	})
	be.Err(t, err, nil)
	be.Equal(t, "CUSTOM.EXE", got)

	got, err = model.JsDosCommand(&models.File{
		Filename:       null.StringFrom("demo.zip"),
		FileZipContent: null.StringFrom("readme.txt\nAPP.COM"),
	})
	be.Err(t, err, nil)
	be.Equal(t, "APP.COM", got)
}

func TestJsDosConfig(t *testing.T) {
	t.Parallel()

	_, err := model.JsDosConfig(nil)
	be.Err(t, err)

	cfg, err := model.JsDosConfig(&models.File{
		DoseeHardwareCPU:     null.StringFrom("8086"),
		DoseeHardwareGraphic: null.StringFrom("vga"),
		DoseeHardwareAudio:   null.StringFrom("sb16"),
		DoseeNoEms:           null.Int16From(1),
		DoseeNoXMS:           null.Int16From(1),
		DoseeNoUmb:           null.Int16From(1),
	})
	be.Err(t, err, nil)
	be.True(t, bytes.Contains(cfg, []byte("cputype=386_slow")))
	be.True(t, bytes.Contains(cfg, []byte("machine=vga")))
	be.True(t, bytes.Contains(cfg, []byte("sbtype=sb16")))
	be.True(t, bytes.Contains(cfg, []byte("ems=false")))
	be.True(t, bytes.Contains(cfg, []byte("xms=false")))
	be.True(t, bytes.Contains(cfg, []byte("umb=false")))
}
