package jsdos_test

import (
	"bytes"
	"slices"
	"testing"

	"github.com/Defacto2/server/handler/jsdos"
	"github.com/nalgeon/be"
	"github.com/subpop/go-ini"
)

const mockZipContent = "filename.zip\nreadme.txt\nrunme.bat\napp.com\ndata.dat"

func TestIni(t *testing.T) {
	t.Parallel()
	cfg := jsdos.Jsdos{}
	cfg.CPU("8086")
	b, err := ini.Marshal(cfg)
	be.Err(t, err, nil)
	wants := []string{"core=simple", "cputype=386_slow", "core=simple", "cycles=fixed 330"}
	for v := range slices.Values(wants) {
		be.True(t, bytes.Contains(b, []byte(v)))
	}
	cfg.Machine("vga")
	b, err = ini.Marshal(cfg)
	be.Err(t, err, nil)
	be.True(t, bytes.Contains(b, []byte("machine=vga")))
	be.True(t, bytes.Contains(b, []byte("memsize=16")))
	cfg.Sound("pcspeaker")
	b, err = ini.Marshal(cfg)
	be.Err(t, err, nil)
	wants = []string{"pcspeaker=true", "pcrate=44100", "tandy=off", "disney=false"}
	for v := range slices.Values(wants) {
		be.True(t, bytes.Contains(b, []byte(v)))
	}
	cfg.Tandy()
	b, err = ini.Marshal(cfg)
	be.Err(t, err, nil)
	be.True(t, bytes.Contains(b, []byte("tandy=true")))
	be.True(t, bytes.Contains(b, []byte("tandyrate=44100")))
	cfg.NoEMS(true)
	b, err = ini.Marshal(cfg)
	be.Err(t, err, nil)
	be.True(t, bytes.Contains(b, []byte("ems=false")))
	cfg.NoXMS(true)
	b, err = ini.Marshal(cfg)
	be.Err(t, err, nil)
	be.True(t, bytes.Contains(b, []byte("xms=false")))
	cfg.NoUMB(true)
	b, err = ini.Marshal(cfg)
	be.Err(t, err, nil)
	be.True(t, bytes.Contains(b, []byte("umb=false")))
	cfg.NoMIDI()
	b, err = ini.Marshal(cfg)
	be.Err(t, err, nil)
	be.True(t, bytes.Contains(b, []byte("mididevice=none")))
	cfg.NoGUS()
	b, err = ini.Marshal(cfg)
	be.Err(t, err, nil)
	be.True(t, bytes.Contains(b, []byte("gus=false")))
	cfg.NoBlaster()
	b, err = ini.Marshal(cfg)
	be.Err(t, err, nil)
	be.True(t, bytes.Contains(b, []byte("sbtype=none")))
	cfg.NoBeeper()
	b, err = ini.Marshal(cfg)
	be.Err(t, err, nil)
	be.True(t, bytes.Contains(b, []byte("pcspeaker=false")))
}

func TestFM(t *testing.T) {
	t.Parallel()
	cfg := jsdos.Jsdos{}
	cfg.FM()
	b, err := ini.Marshal(cfg)
	be.Err(t, err, nil)
	be.True(t, bytes.Contains(b, []byte("oplmode=auto")))
}

func TestCovox(t *testing.T) {
	t.Parallel()
	cfg := jsdos.Jsdos{}
	cfg.Covox()
	b, err := ini.Marshal(cfg)
	be.Err(t, err, nil)
	be.True(t, bytes.Contains(b, []byte("disney=true")))
}

func TestSound(t *testing.T) {
	t.Parallel()
	cfg := jsdos.Jsdos{}
	cfg.Sound("sb16")
	b, err := ini.Marshal(cfg)
	be.Err(t, err, nil)
	be.True(t, bytes.Contains(b, []byte("sbtype=sb16")))
	cfg = jsdos.Jsdos{}
	cfg.Sound("sb1")
	b, err = ini.Marshal(cfg)
	be.Err(t, err, nil)
	be.True(t, bytes.Contains(b, []byte("sbtype=sb1")))
	cfg = jsdos.Jsdos{}
	cfg.Sound("none")
	b, err = ini.Marshal(cfg)
	be.Err(t, err, nil)
	be.True(t, bytes.Contains(b, []byte("mpu401=none")))
	cfg = jsdos.Jsdos{}
	cfg.Sound("gus")
	b, err = ini.Marshal(cfg)
	be.Err(t, err, nil)
	be.True(t, bytes.Contains(b, []byte("gus=true")))
	cfg = jsdos.Jsdos{}
	cfg.Sound("covox")
	b, err = ini.Marshal(cfg)
	be.Err(t, err, nil)
	be.True(t, bytes.Contains(b, []byte("disney=true")))
}

func TestPlatform(t *testing.T) {
	t.Parallel()
	be.Equal(t, "hercules", jsdos.Hercules)
	be.Equal(t, "cga", jsdos.CGA)
	be.Equal(t, "tandy", jsdos.Tandy)
	be.Equal(t, "pcjr", jsdos.PCjr)
	be.Equal(t, "ega", jsdos.EGA)
	be.Equal(t, "vgaonly", jsdos.VGAOnly)
	be.Equal(t, "svga_s3", jsdos.SuperVgaS3)
	be.Equal(t, "svga_et3000", jsdos.SuperVgaET3000)
	be.Equal(t, "svga_et4000", jsdos.SuperVgaET4000)
	be.Equal(t, "svga_paradise", jsdos.SuperVgaParadise)
	be.Equal(t, "vesa_nolfb", jsdos.VesaNoFrameBuff)
	be.Equal(t, "vesa_oldvbe", jsdos.VesaV1)
}

func TestCore(t *testing.T) {
	t.Parallel()
	be.Equal(t, "auto", jsdos.AutoCore)
	be.Equal(t, "normal", jsdos.Normal)
	be.Equal(t, "simple", jsdos.Simple)
}

func TestCPUType(t *testing.T) {
	t.Parallel()
	be.Equal(t, "auto", jsdos.IAuto)
	be.Equal(t, "386", jsdos.I386)
	be.Equal(t, "386_prefetch", jsdos.I386Pre)
	be.Equal(t, "386_slow", jsdos.I386Slow)
	be.Equal(t, "486_slow", jsdos.I486Slow)
	be.Equal(t, "pentium_slow", jsdos.I586Slow)
}

func TestCycles(t *testing.T) {
	t.Parallel()
	be.Equal(t, "auto", jsdos.AutoCycles)
	be.Equal(t, "max", jsdos.Max)
	be.Equal(t, "fixed 330", jsdos.Fix5Mhz)
}

func TestRAM(t *testing.T) {
	t.Parallel()
	be.Equal(t, "1", jsdos.Mem286)
	be.Equal(t, "4", jsdos.Mem386)
	be.Equal(t, "16", jsdos.Mem486)
}

func TestDosPaths(t *testing.T) {
	t.Parallel()
	s := jsdos.Paths("")
	be.True(t, len(s) == 0)
	s = jsdos.Paths(mockZipContent)
	be.True(t, len(s) == 5)

	x := "filename.zip\rreadme.txt\nrunme.bat\r\nAPP.COM\ndata.dat"
	s = jsdos.Paths(x)
	be.True(t, len(s) == 5)
}

func TestDosBins(t *testing.T) {
	t.Parallel()
	bins := jsdos.Binaries()
	be.True(t, len(bins) == 0)

	// p := jsdos.Paths(mockZipContent)
	// bins = jsdos.Binaries(p...)
	// assert.Empty(t, bins)

	x := "filename.zip\rreadme.txt\nrunme.bat\r\nAPP.COM\ndata.dat"
	p := jsdos.Paths(x)
	bins = jsdos.Binaries(p...)
	be.True(t, len(bins) == 2)
}

func TestFinds(t *testing.T) {
	t.Parallel()
	s := jsdos.Finds("", "")
	be.True(t, len(s) == 0)

	p := jsdos.Paths(mockZipContent)
	s = jsdos.Finds("filename.zip", p...)
	be.True(t, len(s) == 0)

	x := mockZipContent
	x += "\nFILENAME.EXE\nfilename.xxx"
	p = jsdos.Paths(x)
	s = jsdos.Finds("filename.zip", p...)
	be.Equal(t, "FILENAME.EXE", s)

	x = "FILENAME.COM\n" + x
	p = jsdos.Paths(x)
	s = jsdos.Finds("filename.zip", p...)
	be.Equal(t, "FILENAME.EXE", s)
}

func TestFindBinary(t *testing.T) {
	example := "readme.txt\nRUN.BAT\napp.com\ndata.dat"

	t.Parallel()
	s := jsdos.FindBinary("", "")
	be.True(t, len(s) == 0)

	s = jsdos.FindBinary("filename", "")
	be.Equal(t, "filename", s)

	s = jsdos.FindBinary("filename.xyz", "")
	be.Equal(t, "filename.xyz", s)

	s = jsdos.FindBinary("filename.zip", "zipcontent")
	be.True(t, len(s) == 0)

	s = jsdos.FindBinary("filename.zip", "readme.txt")
	be.True(t, len(s) == 0)

	s = jsdos.FindBinary("filename.zip", example)
	be.Equal(t, "RUN.BAT", s)

	s = jsdos.FindBinary("filename.zip", example+"\n"+"filename.exe")
	be.Equal(t, "filename.exe", s)
}

func TestValid(t *testing.T) {
	t.Parallel()
	be.True(t, jsdos.Valid(""))
	be.True(t, jsdos.Valid("filename"))
	be.True(t, jsdos.Valid("filename.zip"))
	be.True(t, jsdos.Valid("dir mygame"))
	be.True(t, jsdos.Valid("dir mygame && cd mygame && runme.bat"))
	be.True(t, !jsdos.Valid("dir mygamë && cd mygamë && runme.bat"))
	be.True(t, !jsdos.Valid("dir mygame && cd mygame && mysuperlongcommand"))
	be.True(t, !jsdos.Valid(".TXT"))
	be.True(t, !jsdos.Valid(".HIDDEN"))
}
