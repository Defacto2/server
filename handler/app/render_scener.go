package app

import (
	"context"
	"net/http"

	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Package file render_scener.go contains the handler functions for the scener and people pages.

// Scener is the handler for the page to list all the sceners.
func Scener(z *zap.SugaredLogger, c echo.Context) error {
	data := empty()
	title := "Sceners, the people of The Scene"
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["description"] = demo
	return scener(z, c, postgres.Roles(), data)
}

// Artist is the handler for the Artist sceners page.
func Artist(z *zap.SugaredLogger, c echo.Context) error {
	data := empty()
	title := "Pixel artists and graphic designers"
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["description"] = demo
	return scener(z, c, postgres.Artist, data)
}

// Code is the handler for the Coder sceners page.
func Coder(z *zap.SugaredLogger, c echo.Context) error {
	data := empty()
	title := "Coder and programmers"
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["description"] = demo
	return scener(z, c, postgres.Writer, data)
}

// Musician is the handler for the Musiciansceners page.
func Musician(z *zap.SugaredLogger, c echo.Context) error {
	data := empty()
	title := "Musicians and composers"
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["description"] = demo
	return scener(z, c, postgres.Musician, data)
}

// Writer is the handler for the Writer page.
func Writer(z *zap.SugaredLogger, c echo.Context) error {
	data := empty()
	title := "Writers, editors and authors"
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["description"] = demo
	return scener(z, c, postgres.Writer, data)
}

func scener(z *zap.SugaredLogger, c echo.Context, r postgres.Role,
	data map[string]interface{}) error {
	const name = "scener"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	s := model.Sceners{}
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	switch r {
	case postgres.Writer:
		err = s.Writer(ctx, db)
	case postgres.Artist:
		err = s.Artist(ctx, db)
	case postgres.Musician:
		err = s.Musician(ctx, db)
	case postgres.Coder:
		err = s.Coder(ctx, db)
	case postgres.Roles():
		err = s.All(ctx, db)
	}
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	data["sceners"] = s.Sort()
	data["description"] = "Sceners and people who have been credited for their work in The Scene."
	data["lead"] = "This page shows the sceners and people who have been credited for their work in The Scene." +
		" The list is not complete or accurate, due to the amount of data and the lack of a standard format for crediting people." +
		" Sceners themselves would often use different names or spellings on their own work, " +
		" including character swaps, aliases, the use of initals and even single letter signatures."
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}
