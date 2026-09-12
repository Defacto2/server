//nolint:gochecknoglobals,mnd
package testutil

import (
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type (
	Testfile string                // Testfile is the key used to reference TestData.
	TestData = map[string]Testfile // TestData is a map of the filenames found in the "testdata/" subdirectory of testutil.
)

var testdata = TestData{
	"implodezip":  "IMPLODE.ZIP",
	"logotxt":     "LOGO.TXT",
	"screenpng":   "SCREEN.PNG",
	"testascii":   "TEST.ASCII",
	"testbmp":     "TEST.BMP",
	"testgif":     "TEST.GIF",
	"testjpg":     "TEST.JPG",
	"testpcx":     "TEST.PCX",
	"testpng":     "TEST.PNG",
	"testwebp":    "TEST.WEBP",
	"readmetxt":   "readme.txt",
	"archivezip":  "archive.zip",
	"defacto2com": "defacto2.com",
}

// Abs returns an absolute representation of the testfile path.
func (tf Testfile) Abs() string {
	path, _ := filepath.Abs(filepath.Join(Project(), "testutil", "testdata", string(tf)))
	return path
}

// Dir returns the directory path of the testfile.
func (tf Testfile) Dir() string {
	if tf == "" {
		return ""
	}
	return filepath.Dir(string(tf))
}

// CopyPNG copies a sample PNG image to the target destination.
func CopyPNG(tb testing.TB, dest string) {
	tb.Helper()

	cpEmbed(tb, dest, "SCREEN.PNG")
}

// CopyTXT copies a sample textfile to the target destination.
func CopyTXT(tb testing.TB, dest string) {
	tb.Helper()

	cpEmbed(tb, dest, "LOGO.TXT")
}

// Count returns the number of [TestData] entries.
func Count() int {
	return len(testdata)
}

// CountTest returns the number of [TestData] entries that have a filename prefix of "TEST." (test plus a dot).
func CountTest() int {
	cnt := 0
	for _, name := range testdata {
		if strings.HasPrefix(string(name), "TEST.") {
			cnt++
		}
	}
	return cnt
}

// EmbedFS returns the embed file system of the testdata subdirectory.
func EmbedFS(tb testing.TB) fs.FS {
	tb.Helper()

	return testdataFS
}

// FileData returns the filename of the key testdata.
//
// For example:
//
//	FileData("logotxt") = "LOGO.TXT"
func FileData(key string) Testfile {
	return testdata[key]
}

// Files returns a map of the [TestData].
func Files() TestData {
	data := make(TestData, len(testdata))
	maps.Copy(data, testdata)
	return data
}

// Project returns the absolute representation to the root directory for this project repository.
func Project() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	return filepath.Dir(filepath.Dir(filename))
}

// Testdata returns the absolute representation to the "testdata/" subdirectory found in this testutil package.
func Testdata() string {
	path, _ := filepath.Abs(filepath.Join(Project(), "testutil", "testdata"))
	return path
}

func cpEmbed(tb testing.TB, dest, name string) {
	tb.Helper()

	data, err := testdataFS.ReadFile("testdata/" + name)
	if err != nil {
		tb.Fatal(err)
	}
	err = os.WriteFile(dest, data, 0o600)
	if err != nil {
		tb.Fatal(err)
	}
}
