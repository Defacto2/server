package jsdos_test

import (
	"testing"

	"github.com/Defacto2/server/internal/jsdos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/subpop/go-ini"
)

const mockZipContent = "filename.zip\nreadme.txt\nrunme.bat\napp.com\ndata.dat"

func TestIni(t *testing.T) {
	t.Parallel()
	cfg := jsdos.Jsdos{}
	cfg.CPU("8086")
	b, err := ini.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(b), "cputype=386_slow")
	assert.Contains(t, string(b), "core=simple")
	assert.Contains(t, string(b), "cycles=fixed 330")
	cfg.Machine("vga")
	b, err = ini.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(b), "machine=vga")
	assert.Contains(t, string(b), "memsize=16")
	cfg.Sound("pcspeaker")
	b, err = ini.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(b), "pcspeaker=true")
	assert.Contains(t, string(b), "pcrate=44100")
	assert.Contains(t, string(b), "tandy=off")
	assert.Contains(t, string(b), "disney=false")
	cfg.Tandy()
	b, err = ini.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(b), "tandy=true")
	assert.Contains(t, string(b), "tandyrate=44100")
	cfg.NoEMS(true)
	b, err = ini.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(b), "ems=false")
	cfg.NoXMS(true)
	b, err = ini.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(b), "xms=false")
	cfg.NoUMB(true)
	b, err = ini.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(b), "umb=false")
	cfg.NoMIDI()
	b, err = ini.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(b), "mididevice=none")
	cfg.NoGUS()
	b, err = ini.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(b), "gus=false")
	cfg.NoBlaster()
	b, err = ini.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(b), "sbtype=none")
	cfg.NoBeeper()
	b, err = ini.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(b), "pcspeaker=false")
}

func TestFM(t *testing.T) {
	t.Parallel()
	cfg := jsdos.Jsdos{}
	cfg.FM()
	b, err := ini.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(b), "oplmode=auto")
}

func TestCovox(t *testing.T) {
	t.Parallel()
	cfg := jsdos.Jsdos{}
	cfg.Covox()
	b, err := ini.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(b), "disney=true")
}

func TestSound(t *testing.T) {
	t.Parallel()
	cfg := jsdos.Jsdos{}
	cfg.Sound("sb16")
	b, err := ini.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(b), "sbtype=sb16")
	cfg = jsdos.Jsdos{}
	cfg.Sound("sb1")
	b, err = ini.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(b), "sbtype=sb1")
	cfg = jsdos.Jsdos{}
	cfg.Sound("none")
	b, err = ini.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(b), "mpu401=none")
	cfg = jsdos.Jsdos{}
	cfg.Sound("gus")
	b, err = ini.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(b), "gus=true")
	cfg = jsdos.Jsdos{}
	cfg.Sound("covox")
	b, err = ini.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(b), "disney=true")
}

func TestPlatform(t *testing.T) {
	t.Parallel()
	assert.EqualValues(t, "hercules", jsdos.Hercules)
	assert.EqualValues(t, "cga", jsdos.CGA)
	assert.EqualValues(t, "tandy", jsdos.Tandy)
	assert.EqualValues(t, "pcjr", jsdos.PCjr)
	assert.EqualValues(t, "ega", jsdos.EGA)
	assert.EqualValues(t, "vgaonly", jsdos.VGAOnly)
	assert.EqualValues(t, "svga_s3", jsdos.SuperVgaS3)
	assert.EqualValues(t, "svga_et3000", jsdos.SuperVgaET3000)
	assert.EqualValues(t, "svga_et4000", jsdos.SuperVgaET4000)
	assert.EqualValues(t, "svga_paradise", jsdos.SuperVgaParadise)
	assert.EqualValues(t, "vesa_nolfb", jsdos.VesaNoFrameBuff)
	assert.EqualValues(t, "vesa_oldvbe", jsdos.VesaV1)
}

func TestCore(t *testing.T) {
	t.Parallel()
	assert.EqualValues(t, "auto", jsdos.AutoCore)
	assert.EqualValues(t, "dynamic", jsdos.Dynamic)
	assert.EqualValues(t, "normal", jsdos.Normal)
	assert.EqualValues(t, "simple", jsdos.Simple)
}

func TestCPUType(t *testing.T) {
	t.Parallel()
	assert.EqualValues(t, "auto", jsdos.IAuto)
	assert.EqualValues(t, "386", jsdos.I386)
	assert.EqualValues(t, "386_prefetch", jsdos.I386Pre)
	assert.EqualValues(t, "386_slow", jsdos.I386Slow)
	assert.EqualValues(t, "486_slow", jsdos.I486Slow)
	assert.EqualValues(t, "pentium_slow", jsdos.I586Slow)
}

func TestCycles(t *testing.T) {
	t.Parallel()
	assert.EqualValues(t, "auto", jsdos.AutoCycles)
	assert.EqualValues(t, "max", jsdos.Max)
	assert.EqualValues(t, "fixed 330", jsdos.Fix5Mhz)
}

func TestRAM(t *testing.T) {
	t.Parallel()
	assert.EqualValues(t, "1", jsdos.Mem286)
	assert.EqualValues(t, "4", jsdos.Mem386)
	assert.EqualValues(t, "16", jsdos.Mem486)
}

func TestDosPaths(t *testing.T) {
	t.Parallel()
	s := jsdos.Paths("")
	assert.Empty(t, s)

	s = jsdos.Paths(mockZipContent)
	assert.Len(t, s, 5)

	x := "filename.zip\rreadme.txt\nrunme.bat\r\nAPP.COM\ndata.dat"
	s = jsdos.Paths(x)
	assert.Len(t, s, 5)
}

func TestDosBins(t *testing.T) {
	t.Parallel()
	bins := jsdos.Binaries()
	assert.Empty(t, bins)

	// p := jsdos.Paths(mockZipContent)
	// bins = jsdos.Binaries(p...)
	// assert.Empty(t, bins)

	x := "filename.zip\rreadme.txt\nrunme.bat\r\nAPP.COM\ndata.dat"
	p := jsdos.Paths(x)
	bins = jsdos.Binaries(p...)
	assert.Len(t, bins, 2)
}

func TestFinds(t *testing.T) {
	t.Parallel()
	s := jsdos.Finds("", "")
	assert.Empty(t, s)

	p := jsdos.Paths(mockZipContent)
	s = jsdos.Finds("filename.zip", p...)
	assert.Empty(t, s)

	x := mockZipContent
	x += "\nFILENAME.EXE\nfilename.xxx"
	p = jsdos.Paths(x)
	s = jsdos.Finds("filename.zip", p...)
	assert.Equal(t, "FILENAME.EXE", s)

	x = "FILENAME.COM\n" + x
	p = jsdos.Paths(x)
	s = jsdos.Finds("filename.zip", p...)
	assert.Equal(t, "FILENAME.EXE", s)
}

func TestDosBin(t *testing.T) {
	// t.Parallel()
	// s := jsdos.Binary()
	// assert.Empty(t, s)

	// x := mockZipContent
	// p := jsdos.Paths(x)
	// s = jsdos.Binary(p...)
	// assert.Empty(t, s)

	// x += "\nfilename.exe\nfilename.xxx"
	// p = jsdos.Paths(x)
	// s = jsdos.Binary(p...)
	// assert.Equal(t, "filename.exe", s)

	// x = "FILENAME.COM\n" + x
	// p = jsdos.Paths(x)
	// s = jsdos.Binary(p...)
	// assert.Equal(t, "FILENAME.COM", s)

	// x += "\nrunme.bat"
	// p = jsdos.Paths(x)
	// s = jsdos.Binary(p...)
	// assert.Equal(t, "runme.bat", s)
}

func TestFindBinary(t *testing.T) {
	example := "readme.txt\nRUN.BAT\napp.com\ndata.dat"

	t.Parallel()
	s := jsdos.FindBinary("", "")
	assert.Equal(t, "", s)

	s = jsdos.FindBinary("filename", "")
	assert.Equal(t, "filename", s)

	s = jsdos.FindBinary("filename.xyz", "")
	assert.Equal(t, "filename.xyz", s)

	s = jsdos.FindBinary("filename.zip", "zipcontent")
	assert.Equal(t, "", s)

	s = jsdos.FindBinary("filename.zip", "readme.txt")
	assert.Equal(t, "", s)

	s = jsdos.FindBinary("filename.zip", example)
	assert.Equal(t, "RUN.BAT", s)

	s = jsdos.FindBinary("filename.zip", example+"\n"+"filename.exe")
	assert.Equal(t, "filename.exe", s)
}

func TestFmt8dot3(t *testing.T) {
	t.Parallel()
	s := jsdos.Fmt8dot3("")
	assert.Equal(t, "", s)

	s = jsdos.Fmt8dot3("filename")
	assert.Equal(t, "filename", s)

	s = jsdos.Fmt8dot3("filename.exe")
	assert.Equal(t, "filename.exe", s)

	s = jsdos.Fmt8dot3("my backup collection.7zip")
	assert.Equal(t, "my bac~1.7zi", s)

	s = jsdos.Fmt8dot3("filename.zip.exe")
	assert.Equal(t, "filena~1.exe", s)
}
