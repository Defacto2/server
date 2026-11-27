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
	wants := []string{"core=simple", "cputype=386_slow", "cycles=fixed 330"}
	for v := range slices.Values(wants) {
		be.True(t, bytes.Contains(b, []byte(v)))
	}
	cfg.Machine("vga")
	b, err = ini.Marshal(cfg)
	be.Err(t, err, nil)
	be.True(t, bytes.Contains(b, []byte("machine=vgaonly")))
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

func TestDosPaths(t *testing.T) {
	t.Parallel()

	be.Equal(t, jsdos.Paths(""), []string{})
	be.Equal(t, jsdos.Paths(mockZipContent), []string{"filename.zip", "readme.txt", "runme.bat", "app.com", "data.dat"})

	x := "filename.zip\rreadme.txt\nrunme.bat\r\nAPP.COM\ndata.dat"
	be.Equal(t, jsdos.Paths(x), []string{"filename.zip", "readme.txt", "runme.bat", "APP.COM", "data.dat"})
}

func TestDosBins(t *testing.T) {
	t.Parallel()

	be.Equal(t, jsdos.Binaries(), []string{})

	x := "filename.zip\rreadme.txt\nrunme.bat\r\nAPP.COM\ndata.dat"
	p := jsdos.Paths(x)
	be.Equal(t, jsdos.Binaries(p...), []string{"runme.bat", "APP.COM"})
}

func TestFinds(t *testing.T) {
	t.Parallel()

	be.Equal(t, jsdos.Finds("", ""), "")

	p := jsdos.Paths(mockZipContent)
	be.Equal(t, jsdos.Finds("filename.zip", p...), "")

	x := mockZipContent + "\nFILENAME.EXE\nfilename.xxx"
	p = jsdos.Paths(x)
	be.Equal(t, jsdos.Finds("filename.zip", p...), "FILENAME.EXE")

	x = "FILENAME.COM\n" + x
	p = jsdos.Paths(x)
	be.Equal(t, jsdos.Finds("filename.zip", p...), "FILENAME.EXE")

	p = jsdos.Paths("sub/filename.exe\napp.exe")
	be.Equal(t, jsdos.Finds("filename.zip", p...), "")

	p = jsdos.Paths(`sub\filename.exe` + "\nFILENAME.BAT")
	be.Equal(t, jsdos.Finds("filename.zip", p...), "FILENAME.BAT")

	p = jsdos.Paths("sub/filename.exe\ndeep/dir/app.com")
	be.Equal(t, jsdos.Finds("filename.zip", p...), "")
}

func TestBinary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		paths []string
		want  string
	}{
		{"empty", nil, ""},
		{"no executables", []string{"readme.txt", "data.dat"}, ""},
		{"bat over exe and com", []string{"app.com", "run.exe", "start.bat"}, "start.bat"},
		{"exe over com", []string{"app.com", "run.exe"}, "run.exe"},
		{"com only", []string{"readme.txt", "app.com"}, "app.com"},
		{"root over subdir", []string{"sub/start.bat", "root.bat"}, "root.bat"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := jsdos.Binary(tt.paths...)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestFindBinary(t *testing.T) {
	t.Parallel()
	const example = "readme.txt\nRUN.BAT\napp.com\ndata.dat"

	be.Equal(t, jsdos.FindBinary("", ""), "")
	be.Equal(t, jsdos.FindBinary("filename", ""), "filename")
	be.Equal(t, jsdos.FindBinary("filename.xyz", ""), "filename.xyz")
	be.Equal(t, jsdos.FindBinary("filename.zip", "zipcontent"), "")
	be.Equal(t, jsdos.FindBinary("filename.zip", "readme.txt"), "")
	be.Equal(t, jsdos.FindBinary("filename.zip", example), "RUN.BAT")
	be.Equal(t, jsdos.FindBinary("filename.zip", example+"\n"+"filename.exe"), "filename.exe")
	be.Equal(t, jsdos.FindBinary("filename.zip", "sub/dir/setup.bat\nrun.bat"), "run.bat")
	be.Equal(t, jsdos.FindBinary("filename.zip", "sub/dir/setup.exe\napp.exe"), "app.exe")
	be.Equal(t, jsdos.FindBinary("filename.zip", `sub\dir\setup.com`+"\napp.com"), "app.com")
	be.Equal(t, jsdos.FindBinary("app.exe", "readme.txt\nother.exe"), "app.exe")
	be.Equal(t, jsdos.FindBinary("filename.zip", "readme.txt\nonly.exe"), "only.exe")
}

func TestCPU(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		value      string
		wantModel  jsdos.CPU
		wantCore   jsdos.Core
		wantCycles jsdos.Cycles
	}{
		{"8086", "8086", jsdos.I386Slow, jsdos.Simple, jsdos.Fix5Mhz},
		{"386", "386", jsdos.I386Slow, jsdos.Normal, ""},
		{"486", "486", jsdos.IAuto, jsdos.Normal, jsdos.Max},
		{"auto", "AUTO", jsdos.IAuto, jsdos.AutoCore, jsdos.AutoCycles},
		{"unknown", "pentium", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := jsdos.Jsdos{}
			cfg.CPU(tt.value)
			be.Equal(t, cfg.Model, tt.wantModel)
			be.Equal(t, cfg.Core, tt.wantCore)
			be.Equal(t, cfg.Cycles, tt.wantCycles)
		})
	}
}

func TestMachine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		value       string
		wantMachine jsdos.Platform
		wantMem     jsdos.RAM
		wantTandy   string
	}{
		{"vga", "vga", jsdos.VGAOnly, jsdos.Mem486, ""},
		{"tandy", "tandy", jsdos.Tandy, jsdos.Mem286, "true"},
		{"svga", "svga", jsdos.SuperVgaS3, jsdos.Mem486, ""},
		{"paradise", "paradise", jsdos.SuperVgaParadise, jsdos.Mem486, ""},
		{"oldvbe", "oldvbe", jsdos.VesaV1, jsdos.Mem486, ""},
		{"nolfb", "nolfb", jsdos.VesaNoFrameBuff, jsdos.Mem486, ""},
		{"et4000", "et4000", jsdos.SuperVgaET4000, jsdos.Mem486, ""},
		{"et3000", "et3000", jsdos.SuperVgaET3000, jsdos.Mem486, ""},
		{"ega", "ega", jsdos.EGA, jsdos.Mem386, ""},
		{"cga", "CGA", jsdos.CGA, jsdos.Mem286, ""},
		{"unknown", "other", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := jsdos.Jsdos{}
			cfg.Machine(tt.value)
			be.Equal(t, cfg.Dosbox.Machine, tt.wantMachine)
			be.Equal(t, cfg.MemSize, tt.wantMem)
			be.Equal(t, cfg.Speaker.Tandy, tt.wantTandy)
		})
	}
}

func TestNoMemory(t *testing.T) {
	t.Parallel()

	cfg := jsdos.Jsdos{}
	cfg.NoEMS(false)
	cfg.NoXMS(false)
	cfg.NoUMB(false)
	be.Equal(t, cfg.EMS, "")
	be.Equal(t, cfg.XMS, "")
	be.Equal(t, cfg.UMB, "")

	cfg.NoEMS(true)
	cfg.NoXMS(true)
	cfg.NoUMB(true)
	be.Equal(t, cfg.EMS, "false")
	be.Equal(t, cfg.XMS, "false")
	be.Equal(t, cfg.UMB, "false")
}

func TestValid(t *testing.T) {
	t.Parallel()
	be.True(t, jsdos.Valid(""))
	be.True(t, jsdos.Valid("filename"))
	be.True(t, jsdos.Valid("filename.zip"))
	be.True(t, jsdos.Valid("dir mygame"))
	be.True(t, jsdos.Valid("dir mygame && cd mygame && runme.bat"))
	be.True(t, jsdos.Valid(`MYGAME\RUNME.BAT`))
	be.True(t, jsdos.Valid(`CD SUB\MYGAME && RUNME.BAT`))
	be.True(t, !jsdos.Valid("dir mygamë && cd mygamë && runme.bat"))
	be.True(t, !jsdos.Valid("dir mygame && cd mygame && mysuperlongcommand"))
	be.True(t, !jsdos.Valid(".TXT"))
	be.True(t, !jsdos.Valid(".HIDDEN"))
	be.True(t, !jsdos.Valid(`\MYGAME\RUNME.BAT`))
	be.True(t, !jsdos.Valid(`MYGAME\\RUNME.BAT`))
	be.True(t, !jsdos.Valid(`MYSUPERLONGDIR\RUNME.BAT`))
	be.True(t, jsdos.Valid("dir /p && app.exe -v"))
}
