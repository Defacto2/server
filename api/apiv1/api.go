package apiv1

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Routes acts as a JSON, API placeholder.
func Routes(e *echo.Echo, log *zap.SugaredLogger) *echo.Group {
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		log.DPanic(err)
	}
	g := e.Group("api/v1")
	g.GET("/files", func(c echo.Context) error {
		const limit = 1000
		all, err := model.PostAsc.AllFiles(ctx, db, 0, limit)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusCreated, all)
	})
	g.GET("/file/:id", func(c echo.Context) error {
		key, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusNotAcceptable, "")
		}
		one, err := model.One(ctx, db, key)
		if err != nil {
			return c.JSON(http.StatusNotFound, "")
		}
		return c.JSON(http.StatusCreated, one)
	})
	return g
}
