package app

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func PostIntro(z *zap.SugaredLogger, c echo.Context) error {
	const name = "post intro"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	x, err := c.FormParams()
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	c.JSONPretty(200, x, "  ")
	return nil
}
