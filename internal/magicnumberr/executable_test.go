package magicnumberr_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/server/internal/magicnumberr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func windows(name string) string {
	return td(filepath.Join("binaries", "windows", name))
}

func TestMSExe(t *testing.T) {
	t.Parallel()
	t.Log("TestMSExe")
	r, err := os.Open(windows("hellojs.com"))
	require.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.MSExe(r))
}

func TestFindBytesExecutableFreeDOS(t *testing.T) {
	t.Parallel()
	w, err := magicnumberr.FindExecutable(nil)
	require.Error(t, err)
	assert.Equal(t, magicnumberr.UnknownPE, w.PE)
	assert.Equal(t, magicnumberr.NoneNE, w.NE)

	freedos := []string{
		filepath.Join("exe", "EXE.EXE"),
		filepath.Join("exemenu", "exemenu.exe"),
		filepath.Join("press", "PRESS.EXE"),
		filepath.Join("rread", "rread.exe"),
	}
	for _, v := range freedos {
		p, err := os.Open(td(filepath.Join("binaries", "freedos", v)))
		require.NoError(t, err)
		defer p.Close()
		w, err = magicnumberr.FindExecutable(p)
		require.NoError(t, err)
		assert.Equal(t, magicnumberr.UnknownPE, w.PE)
		assert.Equal(t, magicnumberr.NoneNE, w.NE)
		sign, err := magicnumberr.Program(p)
		require.NoError(t, err)
		assert.Equal(t, magicnumberr.MicrosoftExecutable, sign)
	}
}

func TestFindBytesExecutableWinVista(t *testing.T) {
	vista := []string{
		"hello.com",
		"hellojs.com",
		"life.com",
	}
	for _, v := range vista {
		p, err := os.Open(td(filepath.Join("binaries", "windows", v)))
		require.NoError(t, err)
		defer p.Close()
		require.NoError(t, err)
		w, err := magicnumberr.FindExecutable(p)
		require.NoError(t, err)
		assert.Equal(t, magicnumberr.AMD64PE, w.PE)
		assert.Equal(t, 6, w.Major)
		assert.Equal(t, 0, w.Minor)
		assert.Equal(t, 2019, w.TimeDateStamp.Year())
		assert.Equal(t, "Windows Vista 64-bit", fmt.Sprint(w))
		assert.Equal(t, magicnumberr.NoneNE, w.NE)
		sign, err := magicnumberr.Program(p)
		require.NoError(t, err)
		assert.Equal(t, magicnumberr.MicrosoftExecutable, sign)
	}
}

func TestFindBytesExecutableWin3(t *testing.T) {
	winv3 := []string{
		filepath.Join("calmir10", "CALMIRA.EXE"),
		filepath.Join("calmir10", "TASKBAR.EXE"),
		filepath.Join("dskutl21", "DISKUTIL.EXE"),
	}
	for _, v := range winv3 {
		p, err := os.Open(td(filepath.Join("binaries", "windows3x", v)))
		require.NoError(t, err)
		defer p.Close()
		w, err := magicnumberr.FindExecutable(p)
		require.NoError(t, err)
		assert.Equal(t, magicnumberr.UnknownPE, w.PE)
		assert.Equal(t, magicnumberr.Windows286Exe, w.NE)
		assert.Equal(t, 3, w.Major)
		assert.Equal(t, 10, w.Minor)
		assert.Equal(t, "Windows v3.10 for 286", fmt.Sprint(w))
		sign, err := magicnumberr.Program(p)
		require.NoError(t, err)
		assert.Equal(t, magicnumberr.MicrosoftExecutable, sign)
	}

	p, err := os.Open(td(filepath.Join("binaries", "windowsXP", "CoreTempv13", "32bit", "Core Temp.exe")))
	require.NoError(t, err)
	defer p.Close()
	w, err := magicnumberr.FindExecutable(p)
	require.NoError(t, err)
	assert.Equal(t, magicnumberr.Intel386PE, w.PE)
	assert.Equal(t, magicnumberr.NoneNE, w.NE)
	assert.Equal(t, 5, w.Major)
	assert.Equal(t, 0, w.Minor)
	assert.Equal(t, "Windows 2000 32-bit", fmt.Sprint(w))
	sign, err := magicnumberr.Program(p)
	require.NoError(t, err)
	assert.Equal(t, magicnumberr.MicrosoftExecutable, sign)

	p, err = os.Open(td(filepath.Join("binaries", "windowsXP", "CoreTempv13", "64bit", "Core Temp.exe")))
	require.NoError(t, err)
	defer p.Close()
	require.NoError(t, err)
	w, err = magicnumberr.FindExecutable(p)
	require.NoError(t, err)
	assert.Equal(t, magicnumberr.AMD64PE, w.PE)
	assert.Equal(t, magicnumberr.NoneNE, w.NE)
	assert.Equal(t, 5, w.Major)
	assert.Equal(t, 2, w.Minor)
	assert.Equal(t, "Windows XP Professional x64 Edition 64-bit", fmt.Sprint(w))
	sign, err = magicnumberr.Program(p)
	require.NoError(t, err)
	assert.Equal(t, magicnumberr.MicrosoftExecutable, sign)
}

func TestFindExecutableWinNT(t *testing.T) {
	win9x := []string{
		filepath.Join("rlowe-encrypt", "DEMOCD.EXE"),
		filepath.Join("rlowe-encrypt", "DISKDVR.EXE"),
		filepath.Join("rlowe-cdrools", "DEMOCD.EXE"),
		filepath.Join("7za920", "7za.exe"),
		filepath.Join("7z1604-extra", "7za.exe"),
	}
	for _, v := range win9x {
		p, err := os.Open(td(filepath.Join("binaries", "windows9x", v)))
		require.NoError(t, err)
		defer p.Close()
		w, err := magicnumberr.FindExecutable(p)
		require.NoError(t, err)
		assert.Equal(t, magicnumberr.Intel386PE, w.PE)
		assert.Equal(t, 4, w.Major)
		assert.Equal(t, 0, w.Minor)
		assert.Greater(t, w.TimeDateStamp.Year(), 2000)
		assert.Equal(t, "Windows NT v4.0", fmt.Sprint(w))
		assert.Equal(t, magicnumberr.NoneNE, w.NE)
	}
}

func TestFindExecutableWin9x(t *testing.T) {
	unknown := []string{
		filepath.Join("rlowe-rformat", "RFORMATD.EXE"),
		filepath.Join("rlowe-encrypt", "DFMINST.COM"),
		filepath.Join("rlowe-encrypt", "UNINST.COM"),
	}
	for _, v := range unknown {
		p, err := os.Open(td(filepath.Join("binaries", "windows9x", v)))
		require.NoError(t, err)
		defer p.Close()
		w, _ := magicnumberr.FindExecutable(p)
		assert.Equal(t, magicnumberr.UnknownPE, w.PE)
		assert.Equal(t, 0, w.Major)
		assert.Equal(t, 0, w.Minor)
		assert.Equal(t, 1, w.TimeDateStamp.Year())
		assert.Equal(t, "Unknown PE executable", fmt.Sprint(w))
		assert.Equal(t, magicnumberr.NoneNE, w.NE)
	}

	p, err := os.Open(td(filepath.Join("binaries", "windows9x", "7z1604-extra", "x64", "7za.exe")))
	require.NoError(t, err)
	defer p.Close()
	w, err := magicnumberr.FindExecutable(p)
	require.NoError(t, err)
	assert.Equal(t, magicnumberr.AMD64PE, w.PE)
	assert.Equal(t, 4, w.Major)
	assert.Equal(t, 0, w.Minor)
	assert.Equal(t, 2016, w.TimeDateStamp.Year())
	assert.Equal(t, "Windows NT v4.0 64-bit", fmt.Sprint(w))
	assert.Equal(t, magicnumberr.NoneNE, w.NE)
}
