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
// TODO: create a struct with pagination data.
func (s *sugared) List(tt RecordsBy, c echo.Context) error {
	start := latency()
	id := c.Param("id")
	offset := strings.TrimPrefix(c.Param("offset"), "/")
	fmt.Println(c.ParamNames(), c.ParamValues(), "-->", offset)
	page, _ := strconv.Atoi(offset) // TODO: if err, return 404
	if page < 1 {
		return echo.NewHTTPError(http.StatusNotFound,
			fmt.Sprintf("Page %d of %s doesn't exist", page, tt))
	}
	limit := 0
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
		limit = 1000
		records, err = order.ArtFiles(page, limit, ctx, db)
	case AsDocuments:
		limit = 1000
		records, err = order.DocumentFiles(page, limit, ctx, db)
	case AsSoftware:
		limit = 1000
		records, err = order.SoftwareFiles(page, limit, ctx, db)
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
		count, _ = models.ArtCount(ctx, db)
	case AsDocuments:
		byteSum, err = models.DocumentByteCount(ctx, db)
		count, _ = models.DocumentCount(ctx, db)
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

	maxPage := helpers.PageCount(count, limit)
	if page > int(maxPage) {
		return echo.NewHTTPError(http.StatusNotFound,
			fmt.Sprintf("Page %d of %d for %s doesn't exist", page, maxPage, tt))
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
	n1, n2, n3 := pagi(page, maxPage)
	err = c.Render(http.StatusOK, tt.String(), map[string]interface{}{
		"title":       fmt.Sprintf("%s%s%s", title, fmt.Sprintf("/%s/", tt), id),
		"home":        "",
		"description": desc,
		"parent":      tt.Parent(),
		"stats":       stat,
		"sort":        sorter,
		"records":     records,
		"latency":     fmt.Sprintf("%s.", time.Since(*start)),
		// work-in-progress

		"current":  tt.String(),
		"limit":    limit,
		"pages":    helpers.Pages(count, limit, page),
		"pageZ":    maxPage,
		"page":     page,
		"previous": previous(page),
		"next":     next(page, maxPage),
		"link1":    n1,
		"link2":    n2,
		"link3":    n3,
	})
	if err != nil {
		s.log.Errorf("%s: %s %d", errTmpl, err, tt)
		return echo.NewHTTPError(http.StatusInternalServerError, errTmpl)
	}
	return nil
}

func previous(page int) int {
	if page == 1 {
		return 1
	}
	return page - 1
}

func next(page int, maxPage uint) int {
	max := int(maxPage)
	if page >= max {
		return max
	}
	return page + 1
}

func pagi(page int, maxPage uint) (int, int, int) {
	max := int(maxPage)
	if max < 3 {
		return 0, 0, 0
	}
	a := page + -1
	b := page + 0
	c := page + 1
	if c > max {
		diff := c - max
		c = max - diff
		b = max - diff - 1
		a = max - diff - 2
		return a, b, c
	}
	if c == max {
		diff := c - max + 1
		c = max - diff
		b = max - diff - 1
		a = max - diff - 2
		return a, b, c
	}
	if a <= 1 {
		a = 2
		b = 3
		c = 4
	}
	return a, b, c
}
