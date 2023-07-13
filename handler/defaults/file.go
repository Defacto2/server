package defaults

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func File(s *zap.SugaredLogger, ctx echo.Context, stats bool) error {
	data := initData()
	data["title"] = "File categories"
	data["logo"] = "Categories"
	data["description"] = "Table of contents for the files."
	data["stats"] = stats
	err := ctx.Render(http.StatusOK, "file", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

func Files(s *zap.SugaredLogger, ctx echo.Context, id string) error {
	data := initData()
	data["title"] = "Files placeholder"
	data["logo"] = "Files placeholder"
	data["description"] = "Table of contents for the files."
	err := ctx.Render(http.StatusOK, "file", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}
