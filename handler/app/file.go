package app

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// File is the handler for the file categories page.
func File(s *zap.SugaredLogger, ctx echo.Context, stats bool) error {
	data := initData()
	data["title"] = "File categories"
	data["description"] = "Table of contents for the files."
	data["logo"] = "File categories"
	data["h1"] = "File categories"
	data["stats"] = stats
	if stats {
		data["h1sub"] = "with statistics"
	}
	err := ctx.Render(http.StatusOK, "file", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Files is the handler for the files page.
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
