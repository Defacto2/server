package app

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// PostIntro handles the POST request for the intro upload form.
func PostIntro(z *zap.SugaredLogger, c echo.Context) error {
	const name = "post intro"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	x, err := c.FormParams()
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	c.JSONPretty(http.StatusOK, x, "  ")
	return nil
}

func (a AboutConf) EditorMeCP(z *zap.SugaredLogger, c echo.Context) error {
	const name = "editor copy readme"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}

	type Record struct {
		ID     int    `query:"id"`
		Target string `query:"readme"`
	}
	// in the handler for /users?id=<userID>
	var record Record
	err := c.Bind(&record)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request " + err.Error()})
	}

	f, err := model.Record(z, c, record.ID)
	if err != nil {
		return err
	}

	list := strings.Split(f.FileZipContent.String, "\n")
	target := ""
	for _, x := range list {
		s := strings.TrimSpace(x)
		if s == "" {
			continue
		}
		if strings.EqualFold(s, record.Target) {
			target = s
		}
	}
	if target == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request, target not found"})
	}

	fp := filepath.Join(a.DownloadDir, f.UUID.String)
	fmt.Println("fp", fp)
	st, err := os.Stat(fp)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request, " + err.Error()})
	}
	if st.Size() == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request, file is empty"})
	}

	_, err = exec.LookPath("unzip")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request, " + err.Error()})
	}

	out, err := exec.Command("unzip", "-l", fp).Output()
	if errors.Is(err, exec.ErrDot) {
		err = nil
	}
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request, " + err.Error()})
	}
	fmt.Println("out", string(out))

	tmp, err := os.MkdirTemp(os.TempDir(), "defacto2-")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request, " + err.Error()})
	}
	defer os.RemoveAll(tmp)
	fmt.Println("tmp", tmp)

	// unzip -j "myarchive.zip" "in/archive/file.txt" -d "/path/to/unzip/to"
	out, err = exec.Command("unzip", fp, target, "-d", tmp).Output()
	if errors.Is(err, exec.ErrDot) {
		err = nil
	}
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request, " + err.Error()})
	}
	fmt.Println("out", string(out))

	dst := filepath.Join(tmp, target)
	st, err = os.Stat(dst)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request, " + err.Error()})
	}
	if st.Size() == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request, file is empty"})
	}

	srcFile, err := os.Open(dst)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request, " + err.Error()})
	}
	defer srcFile.Close()

	txt := fmt.Sprintf("%s.txt", fp)
	dstFile, err := os.Create(txt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request, " + err.Error()})
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request, " + err.Error()})
	}

	err = dstFile.Sync()
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request, " + err.Error()})
	}

	return c.JSON(http.StatusOK, record)
}

// EditorMe handles the POST request for the editor readme forms.
func (a AboutConf) EditorMe(z *zap.SugaredLogger, c echo.Context) error {
	const name = "editor readme"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}

	type Record struct {
		ID     int  `query:"id"`
		Readme bool `query:"readme"`
	}
	// in the handler for /users?id=<userID>
	var record Record
	err := c.Bind(&record)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request " + err.Error()})
	}
	err = model.UpdateNoReadme(z, c, int64(record.ID), record.Readme)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request " + err.Error()})
	}
	return c.JSON(http.StatusOK, record)
}
