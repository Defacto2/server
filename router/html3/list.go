package html3

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/sceners"
	"github.com/Defacto2/server/models"
	"github.com/Defacto2/server/pkg/helpers"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/tags"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	pgm "github.com/Defacto2/server/pkg/postgres/models"
)

type sugared struct {
	log *zap.SugaredLogger
}

// Category lists the file records associated with the category tag that is provided by the ID param in the URL.
func (s *sugared) Category(c echo.Context) error {
	return s.List(BySection, c)
}

// Platform lists the file records associated with the platform tag that is provided by the ID param in the URL.
func (s *sugared) Platform(c echo.Context) error {
	return s.List(ByPlatform, c)
}

// Group lists the file records associated with the group that is provided by the ID param in the URL.
func (s *sugared) Group(c echo.Context) error {
	return s.List(ByGroup, c)
}

func (s *sugared) Art(c echo.Context) error {
	return s.List(AsArt, c)
}

func (s *sugared) Documents(c echo.Context) error {
	return s.List(AsDocuments, c)
}

func (s *sugared) Software(c echo.Context) error {
	return s.List(AsSoftware, c)
}

// List all the records associated with the RecordsBy grouping.
func (s *sugared) List(tt RecordsBy, c echo.Context) error {
	start := latency()
	id := c.Param("id")
	offset := strings.TrimPrefix(c.Param("offset"), "/")
	fmt.Println(c.ParamNames(), c.ParamValues(), "-->", offset)
	page, _ := strconv.Atoi(offset) // TODO: if err, return 404
	name := sceners.CleanURL(id)
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		s.log.Warnf("%s: %s", errConn, err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errConn)
	}
	defer db.Close()
	var records pgm.FileSlice
	order := Clauses(c.QueryString())
	switch tt {
	case BySection:
		records, err = order.FilesByCategory(id, ctx, db)
	case ByPlatform:
		records, err = order.FilesByPlatform(id, ctx, db)
	case ByGroup:
		records, err = order.FilesByGroup(name, ctx, db)
	case AsArt:
		records, err = order.ArtFiles(ctx, db)
	case AsDocuments:
		records, err = order.DocumentFiles(ctx, db)
	case AsSoftware:
		records, err = order.SoftwareFiles(page, 1000, ctx, db)
	default:
		s.log.Warnf("%s: %s", errTag, tt)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errTag)
	}
	if err != nil {
		s.log.Warnf("%s: %s", errConn, err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errConn)
	}
	count := len(records)
	if count == 0 {
		// TODO: update the error when page is invalid
		return echo.NewHTTPError(http.StatusNotFound,
			fmt.Sprintf("The %s %q doesn't exist", tt, id))
	}
	var byteSum int64
	switch tt {
	case BySection:
		byteSum, err = models.ByteCountByCategory(id, ctx, db)
	case ByPlatform:
		byteSum, err = models.ByteCountByPlatform(id, ctx, db)
	case ByGroup:
		byteSum, err = models.ByteCountByGroup(name, ctx, db)
	case AsArt:
		byteSum, err = models.ArtByteCount(ctx, db)
	case AsDocuments:
		byteSum, err = models.DocumentByteCount(ctx, db)
	case AsSoftware:
		byteSum, err = models.SoftwareByteCount(ctx, db)
		count, _ = models.SoftwareCount(ctx, db)
	default:
		s.log.Warnf("%s: %s", errTag, tt)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errTag)
	}
	if err != nil {
		s.log.Warnf("%s %s", errConn, err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errConn)
	}
	desc := ""
	switch tt {
	case BySection, ByPlatform:
		key := tags.TagByURI(id)
		info := tags.Infos[key]
		name := tags.Names[key]
		desc = fmt.Sprintf("%s - %s.", name, info)
	case AsArt:
		desc = fmt.Sprintf("%s, %s.", "Digital + pixel art", textArt)
	case AsDocuments:
		desc = fmt.Sprintf("%s, %s.", "Document + text art", textDoc)
	case AsSoftware:
		desc = fmt.Sprintf("%s, %s.", "Software", textSof)
	}
	stat := fmt.Sprintf("%d files, %s", count, helpers.ByteCountFloat(byteSum))
	sorter := sorter(c.QueryString())

	err = c.Render(http.StatusOK, tt.String(), map[string]interface{}{
		"title":       fmt.Sprintf("%s%s%s", title, fmt.Sprintf("/%s/", tt), id),
		"home":        "",
		"description": desc,
		"parent":      tt.Parent(),
		"stats":       stat,
		"sort":        sorter,
		"records":     records,
		"latency":     fmt.Sprintf("%s.", time.Since(*start)),
	})
	if err != nil {
		s.log.Errorf("%s: %s %d", errTmpl, err, tt)
		return echo.NewHTTPError(http.StatusInternalServerError, errTmpl)
	}
	return nil
}
