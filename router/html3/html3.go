package html3

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/bengarrett/df2023/db/models"
	"github.com/bengarrett/df2023/str"
	"github.com/labstack/echo/v4"

	ps "github.com/bengarrett/df2023/db"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func Index(c echo.Context) error {
	start := time.Now()
	r := new(big.Int)
	r.Binomial(1000, 10)

	const pad = 5
	ctx := context.Background()
	db, err := ps.ConnectDB()
	if err != nil {
		return err
	}
	art, err := models.Files(Where("platform = ?", "image"), Where("section != ?", "bbs")).Count(ctx, db)
	if err != nil {
		return err
	}
	doc, err := models.Files(
		Where("platform = ?", "ansi"),
		Or("platform = ?", "text"),
		Or("platform = ?", "textamiga"),
		Or("platform = ?", "pdf")).Count(ctx, db)
	if err != nil {
		return err
	}
	sw, err := models.Files(
		Where("platform = ?", "java"),
		Or("platform = ?", "linux"),
		Or("platform = ?", "dos"),
		Or("platform = ?", "php"),
		Or("platform = ?", "windows")).Count(ctx, db)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "layout", map[string]interface{}{
		"title":   "Index of /html3",
		"art":     str.FWInt(int(art), pad),
		"doc":     str.FWInt(int(doc), pad),
		"sw":      str.FWInt(int(sw), pad),
		"latency": fmt.Sprintf("%s.", time.Since(start)),
	})
}

func Categories(c echo.Context) error {
	return nil
}
