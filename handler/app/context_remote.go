//nolint:exhaustruct_v5,wrapcheck
package app

// Package file context_remote.go contains handlers that proxy or import data
// from external scene APIs (Demozoo and Pouet).

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"uuid"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/app/remote"
	"github.com/Defacto2/server/handler/cache"
	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/handler/pouet"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/nils"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/labstack/echo/v5"
)

const sep = ";"

// GetDemozooParam fetches the multiple download_links values from the
// Demozoo production API and attempts to download and save one of the
// linked files. If multiple links are found, the first link is used as
// they should all point to the same asset.
//
// Both the Demozoo production ID param and the Defacto2 UUID query
// param values are required as params to fetch the production data and
// to save the file to the correct filename.
func GetDemozooParam(sl *slog.Logger, c *echo.Context, exec boil.ContextExecutor, download dir.Directory) error {
	const format = "get demozoo param context: %w"
	if err := nils.Check(sl, c, exec); err != nil {
		return fmt.Errorf(format, err)
	}

	got := remote.Demozoo(0, "", download, 0)
	id, err := echo.PathParam[int](c, "id")
	if err != nil {
		got.Error = "demozoo id must be a numeric value"
		return c.JSON(http.StatusBadRequest, got)
	}
	got.ID = id

	unid := c.QueryParam("unid")
	if _, err = uuid.Parse(unid); err != nil {
		got.Error = "uuid syntax did not validate, " + unid
		return c.JSON(http.StatusBadRequest, got)
	}
	got.UUID = unid

	ctx := c.Request().Context()
	return got.Download(ctx, sl, c, exec, true)
}

// GetDemozoo fetches the download link from Demozoo and saves it to the download directory.
// It then runs Update to modify the database record with various metadata from the file and Demozoo record API data.
//
// This function is a wrapper for the remote.DemozooLink.Download method.
func GetDemozoo(
	ctx context.Context, sl *slog.Logger, c *echo.Context, exec boil.ContextExecutor,
	prodID int, unid string, download dir.Directory,
) error {
	got := remote.Demozoo(prodID, unid, download, 0)
	return got.Download(ctx, sl, c, exec, false)
}

// GetPouet fetches the download link from Pouet and saves it to the download directory.
// It then runs Update to modify the database record with various metadata from the file and Pouet record API data.
//
// This function is a wrapper for the remote.PouetLink.Download method.
func GetPouet(
	ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor,
	prodID int, unid string, download dir.Directory,
) error {
	got := remote.Pouet(prodID, unid, download, 0)
	return got.Download(ctx, sl, exec)
}

// PouetCache parses the cached data for the Pouet production votes.
// If the cache is valid it is returned as JSON response.
// If the cache is invalid or corrupt an error will be returned
// and an API request should be made to Pouet.
func PouetCache(c *echo.Context, data string) error {
	const msg = "pouet cache context"
	if err := nils.Check(c); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	if data == "" {
		return fmt.Errorf("%s: empty data: %w", msg, ErrCorrupt)
	}

	pv := pouet.Votes{ID: 0, Stars: 0.0, VotesAvg: 0.0, VotesUp: 0, VotesMeh: 0, VotesDown: 0}
	votes := strings.Split(data, sep)
	const req = 4
	if len(votes) < req {
		return fmt.Errorf("%s: %d, want %d: %w", msg, len(votes), req, ErrCorrupt)
	}

	const format = msg + " %s: %w"
	stars, err := strconv.ParseFloat(votes[0], 64)
	if err != nil {
		return fmt.Errorf(format, votes[0], err)
	}

	down, err := strconv.Atoi(votes[1])
	if err != nil {
		return fmt.Errorf(format, votes[1], err)
	}

	up, err := strconv.Atoi(votes[2])
	if err != nil {
		return fmt.Errorf(format, votes[2], err)
	}

	meh, err := strconv.Atoi(votes[3])
	if err != nil {
		return fmt.Errorf(format, votes[3], err)
	}

	pv.Stars = stars
	pv.VotesDown = uint64(math.Abs(float64(down)))
	pv.VotesUp = uint64(math.Abs(float64(up)))
	pv.VotesMeh = uint64(math.Abs(float64(meh)))

	if err = c.JSON(http.StatusOK, pv); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return nil
}

// ProdPouet is the handler for the Pouet prod JSON page.
func ProdPouet(c *echo.Context, id string) error {
	const format = "prod pouet context: %w"
	if err := nils.Check(c); err != nil {
		return fmt.Errorf(format, err)
	}

	p := pouet.Production{}
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	ctx := c.Request().Context()
	if _, err = p.Get(ctx, i); err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	if err = c.JSON(http.StatusOK, p); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

// ProdZoo is the handler for the Demozoo production JSON page.
//
// It requires a zooAuth value to be set using echo context.
func ProdZoo(c *echo.Context, id string) error {
	const format = "prod zoo context: %w"
	if err := nils.Check(c); err != nil {
		return fmt.Errorf(format, err)
	}

	prod := demozoo.Production{}
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	ctx := c.Request().Context()
	auth, ok := c.Get(config.ZooAuth).(string)
	if !ok {
		return fmt.Errorf(format, config.ErrNoZooAuth)
	}
	if code, err := prod.Get(ctx, auth, i); err != nil {
		const maxStatusCode = 599
		if code < http.StatusContinue || code > maxStatusCode {
			code = http.StatusBadGateway
		}
		return c.String(code, err.Error())
	}

	if err = c.JSON(http.StatusOK, prod); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

// TestZoo is the handler for the Demozoo API authorization test.
// Introduced in October 2026, it returns either 200 OK or 403 Forbidden.
func TestZoo(c *echo.Context) error {
	const format = "test zoo context: %w"
	if err := nils.Check(c); err != nil {
		return fmt.Errorf(format, err)
	}

	const url = "https://demozoo.org/api/v1/auth_test/"
	const timeout = 15 * time.Second
	const maxBody = 64 << 10

	auth, ok := c.Get(config.ZooAuth).(string)
	if !ok {
		return c.String(http.StatusInternalServerError, config.ErrNoZooAuth.Error())
	}
	req, err := http.NewRequestWithContext(
		c.Request().Context(), http.MethodGet, url, nil)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	req.Header.Set("User-Agent", helper.UserAgent)
	req.Header.Set("Authorization", "Api-Key "+auth)
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if resp == nil {
		return c.String(http.StatusInternalServerError, "demozoo response body is empty")
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(resp.StatusCode, resp.Status+": "+string(b))
}

// VotePouet is the handler for the Pouet production votes JSON page.
func VotePouet(sl *slog.Logger, c *echo.Context, id string) error {
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf("vote pouet context: %w", err)
	}

	pv := pouet.Votes{ID: 0, Stars: 0.0, VotesAvg: 0.0, VotesUp: 0, VotesMeh: 0, VotesDown: 0}
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	cp := cache.PouetVote
	if s, err := cp.Read(id); err == nil {
		if err := PouetCache(c, s); err == nil {
			sl.Debug("vote.pouet", slog.String("cache.hit.id", id))
			return nil
		}
	}
	sl.Debug("vote.pouet", slog.String("cache.miss.for.pouet.id", id))

	ctx := c.Request().Context()
	if err = pv.Votes(ctx, i); err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	if err = c.JSON(http.StatusOK, pv); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	value := fmt.Sprintf("%.1f%s%d%s%d%s%d", pv.Stars, sep, pv.VotesDown, sep, pv.VotesUp, sep, pv.VotesMeh)
	if err := cp.Write(id, value, cache.ExpiredAt); err != nil {
		logID(sl, "vote pouet cache failure", id, 0, err)
	}
	return nil
}
