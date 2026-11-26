package testutil_test

import (
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/Defacto2/server/internal/testutil"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

func TestDOSBIN(t *testing.T) {
	t.Parallel()

	const attr = byte(0x0F)

	t.Run("default", func(t *testing.T) {
		t.Parallel()
		buf := testutil.DOSBIN(t, "")
		be.Equal(t, buf[0], byte('h'))
		be.Equal(t, buf[1], attr)
		be.Equal(t, buf[len("hello world")*2], byte(' '))
		be.Equal(t, buf[len(buf)-2], byte(' '))
		be.Equal(t, buf[len(buf)-1], attr)
	})

	t.Run("limit", func(t *testing.T) {
		t.Parallel()
		s := strings.Repeat("A", 2000)
		buf := testutil.DOSBIN(t, s)
		be.Equal(t, buf[0], byte('A'))
		be.Equal(t, buf[len(buf)-2], byte('A'))
		be.Equal(t, buf[len(buf)-1], attr)
	})

	t.Run("overflow", func(t *testing.T) {
		t.Parallel()
		s := strings.Repeat("B", 2000) + "OVERFLOW"
		buf := testutil.DOSBIN(t, s)
		be.Equal(t, buf[0], byte('B'))
		be.Equal(t, buf[len(buf)-2], byte('B'))
		be.Equal(t, buf[len(buf)-1], attr)
	})
}

func TestPaths(t *testing.T) {
	t.Parallel()

	root := testutil.Project()
	_, err := os.Stat(filepath.Join(root, "go.mod"))
	be.Err(t, err, nil)

	td := testutil.Testdata()
	st, err := os.Stat(td)
	be.Err(t, err, nil)
	be.True(t, st.IsDir())

	logo := testutil.FileData("logotxt")
	be.Equal(t, logo.Dir(), td)
	_, err = os.Stat(logo.Abs())
	be.Err(t, err, nil)

	missing := testutil.FileData("nonexistent")
	be.Equal(t, missing.Abs(), "")
	be.Equal(t, missing.Dir(), "")

	for _, fsys := range []fs.FS{
		testutil.OpenFS(t),
		testutil.ProjectFS(t),
		testutil.OpenRoot(t).FS(),
	} {
		_, err := fs.Stat(fsys, "go.mod")
		be.Err(t, err, nil)
	}
}

func TestLoggerConcurrent(t *testing.T) {
	t.Parallel()

	l := testutil.Buffer(t)
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 50 {
				l.Log.Info("concurrent entry", slog.String("key", "val"))
				_ = l.Contains("concurrent")
				_ = l.String()
			}
		})
	}
	wg.Wait()

	be.True(t, l.Contains("concurrent entry"))
	l.Reset()
	be.Equal(t, l.String(), "")
}

func TestEchoHelpers(t *testing.T) {
	t.Parallel()

	t.Run("NewPath", func(t *testing.T) {
		t.Parallel()
		c := testutil.NewPath(t, "",
			echo.PathValues{{Name: "id", Value: "42"}})
		be.Equal(t, c.Request().Method, http.MethodGet)
		be.Equal(t, c.Request().URL.Path, "/")
		be.Equal(t, c.Request().Header.Get(echo.HeaderContentType), "")
		be.Equal(t, c.Param("id"), "42")
	})

	t.Run("NewInput", func(t *testing.T) {
		t.Parallel()
		c := testutil.NewInput(t, "/search", "q", "defacto2")
		be.Equal(t, c.Request().Method, http.MethodPost)
		be.Equal(t, c.Request().Header.Get(echo.HeaderContentType),
			echo.MIMEApplicationForm)
		be.Equal(t, c.FormValue("q"), "defacto2")
	})

	t.Run("multipart", func(t *testing.T) {
		t.Parallel()
		c := testutil.NewFileInputs(t, "/upload", "doc", "hello.txt", testutil.Input{"title": "demo"})
		be.Equal(t, c.Request().Method, http.MethodPost)
		be.Equal(t, c.FormValue("title"), "demo")

		fh, err := c.FormFile("doc")
		be.Err(t, err, nil)
		be.Equal(t, fh.Filename, "hello.txt")
		f, err := fh.Open()
		be.Err(t, err, nil)
		defer func() { _ = f.Close() }()
		b, err := io.ReadAll(f)
		be.Err(t, err, nil)
		be.Equal(t, string(b), "Hello world!")

		cp := testutil.NewInputsPath(t, "/edit",
			testutil.Input{"name": "nfo"},
			echo.PathValues{{Name: "id", Value: "7"}})
		be.Equal(t, cp.FormValue("name"), "nfo")
		be.Equal(t, cp.Param("id"), "7")
	})
}
