//nolint:wrapcheck
package app

// Package file context_sceners.go contains handlers for scener listings, role filters,
// and individual scener profile artifact pages.

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/releaser"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/labstack/echo/v5"
)

const scener = "scener"

// Artist is the handler for the Artist sceners page.
func Artist(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	const format = "artist context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Artist and design"
	data := empty(c)
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["noindex"] = true

	ctx := c.Request().Context()
	return scenerPage(ctx, sl, c, db, postgres.Artist, data)
}

// Coder is the handler for the Coder sceners page.
func Coder(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	const format = "coder context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Programmer, code, and hacking"
	data := empty(c)
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["noindex"] = true

	ctx := c.Request().Context()
	return scenerPage(ctx, sl, c, db, postgres.Coder, data)
}

// Musician is the handler for the Musician sceners page.
func Musician(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	const format = "musician context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Composer, music, or audio presenter"
	data := empty(c)
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["noindex"] = true

	ctx := c.Request().Context()
	return scenerPage(ctx, sl, c, db, postgres.Musician, data)
}

// Writer is the handler for the Writer sceners page.
func Writer(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf("writer context: %w", err)
	}

	const title = "Writer or magazine, author, interviewee, or mention"
	data := empty(c)
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["noindex"] = true

	ctx := c.Request().Context()
	return scenerPage(ctx, sl, c, db, postgres.Writer, data)
}

// Scener is the handler for the page to list all the sceners.
func Scener(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	const format = "scener context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Sceners, people who were a part of the Scene"
	data := empty(c)
	data["title"] = title
	data["logo"] = title
	data["h1"] = title

	ctx := c.Request().Context()
	return scenerPage(ctx, sl, c, db, postgres.Roles(), data)
}

// scenerPage renders a scener directory page for the given role.
func scenerPage(ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB, r postgres.Role,
	data map[string]any,
) error {
	if err := nils.Check(ctx, sl, c, db); err != nil {
		return fmt.Errorf("scener context: %w", err)
	}

	const descr = "This is a massive but incomplete list of aliases and pseudonyms used " +
		"in the Scene and offers links to individual profiles."

	role := ""
	s := model.Sceners{}
	var err error
	switch r {
	case postgres.Writer:
		role = "writer"
		err = s.Writer(ctx, db)
	case postgres.Artist:
		role = "artist"
		err = s.Artist(ctx, db)
	case postgres.Musician:
		role = "musician"
		err = s.Musician(ctx, db)
	case postgres.Coder:
		role = "coder"
		err = s.Coder(ctx, db)
	case postgres.Roles():
		role = "all"
		err = s.Distinct(ctx, db)
	default:
		// Handle unknown roles gracefully
		err = fmt.Errorf("unknown role %s: %w", r, model.ErrModel)
	}
	if err != nil {
		return DatabaseErr(sl, c, scener, err)
	}

	sorted := s.Sort()
	groups, letters := BuildScenerIndex(sorted)
	data["sceners"] = sorted
	data["scenerRole"] = role
	data["scenerGroups"] = groups
	data["scenerLetters"] = letters
	data["description"] = descr
	data["lead"] = `<small class="fw-lighter">` +
		"Here are Sceners and people credited in the artifacts we hold. " +
		"However, these lists will never be accurate due to the sheer numbers, " +
		"and the lack of a consistency by Sceners. " +
		"Who often used different names and spellings in their work." +
		"</small>"

	if err = c.Render(http.StatusOK, scener, data); err != nil {
		return InternalErr(sl, c, scener, err)
	}
	return nil
}

// Sceners is the handler for the list and preview of files credited to a scener.
func Sceners(sl *slog.Logger, c *echo.Context, db *sql.DB, uri string) error {
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf("sceners context: %w", err)
	}

	errs := "sceners page for, " + uri
	s := releaser.Link(uri)
	var ms model.Scener
	ctx := c.Request().Context()
	fs, err := ms.Where(ctx, db, uri)
	if err != nil {
		return DatabaseErr(sl, c, errs, err)
	}
	if len(fs) == 0 {
		return ScenerErr(sl, c, uri)
	}

	data := emptyFiles(c)
	data[canonical] = strings.Join([]string{"p", uri}, "/")
	data["title"] = s + attr
	data["h1"] = s
	data["lead"] = "Artifacts attributed to " + s + "."
	data["logo"] = s
	data["description"] = "These are the documented artifacts attributed to the person known as " + s + "."
	data["scener"] = s
	data[records] = fs

	d, err := scenerSum(ctx, db, uri)
	if err != nil {
		return DatabaseErr(sl, c, errs, err)
	}
	data["stats"] = d

	if err = c.Render(http.StatusOK, artifacts, data); err != nil {
		return InternalErr(sl, c, errs, err)
	}
	return nil
}

// scenerSum is a helper function for Sceners that returns the statistics for the files page.
func scenerSum(ctx context.Context, exec boil.ContextExecutor, uri string) (map[string]string, error) {
	const format = "scener sum %q: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return nil, fmt.Errorf(format, uri, err)
	}

	m := model.Summary{
		SumBytes: sql.NullInt64{Int64: 0, Valid: false},
		SumCount: sql.NullInt64{Int64: 0, Valid: false},
		MinYear:  sql.NullInt16{Int16: 0, Valid: false},
		MaxYear:  sql.NullInt16{Int16: 0, Valid: false},
	}
	if err := m.ByScener(ctx, exec, uri); err != nil {
		return nil, fmt.Errorf(format, uri, err)
	}

	d := map[string]string{
		files: string(ByteFileS(ffile, m.SumCount.Int64, m.SumBytes.Int64)),
		years: helper.Years(m.MinYear.Int16, m.MaxYear.Int16),
	}
	return d, nil
}
