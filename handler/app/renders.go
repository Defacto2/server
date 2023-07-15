package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Defacto2/server/pkg/helpers"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// initData is a map of default values for the app templates.
func initData() map[string]interface{} {
	return map[string]interface{}{
		"canonical":   "", // A canonical URL is the URL of the best representative page from a group of duplicate pages.
		"carousel":    "", // The ID of the carousel to display.
		"description": "", // A short description of the page that get inserted into the description meta element.
		"h1":          "", // The H1 heading of the page.
		"h1sub":       "", // The H1 sub-heading of the page.
		"lead":        "", // The enlarged, lead paragraph of the page.
		"logo":        "", // Text to insert into the monospaced, ASCII art logo.
		"title":       "", // The title of the page that get inserted into the title meta element.
	}
}

// Error renders a custom HTTP error page.
func Error(err error, c echo.Context) error {
	// Echo custom error handling: https://echo.labstack.com/guide/error-handling/
	start := helpers.Latency()
	code := http.StatusInternalServerError
	msg := "This is a server problem"
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = fmt.Sprint(he.Message)
	}
	return c.Render(code, "error", map[string]interface{}{
		"title":       fmt.Sprintf("%d error, there is a complication", code),
		"description": fmt.Sprintf("%s.", msg),
		"latency":     fmt.Sprintf("%s.", time.Since(*start)),
	})
}

// Index is the handler for the Home page.
func Index(s *zap.SugaredLogger, ctx echo.Context) error {
	data := initData()
	data["description"] = "demo"
	data["title"] = "demo"
	err := ctx.Render(http.StatusNotFound, "index", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Status is the handler for the HTTP status pages such as the 404 - not found.
func Status(s *zap.SugaredLogger, ctx echo.Context, code int, uri string) error {
	// todo: check valid status, or throw error

	data := initData()
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	title := fmt.Sprintf("%d error", code)
	alert := ""
	logo := "??!!"
	probl := "There is a complication"
	switch code {
	case http.StatusNotFound:
		title = "404 error, page not found"
		logo = "Page not found"
		alert = "The page cannot be found"
		probl = "The page you are looking for might have been removed, had its name changed, or is temporarily unavailable."
	case http.StatusForbidden:
		title = "403 error, forbidden"
		logo = "Forbidden"
		alert = "The page is locked"
		probl = "You don't have permission to access this resource."
	case http.StatusInternalServerError:
		title = "500 error, there is a complication"
		logo = "Server error"
		alert = "There is a complication"
		probl = "The server encountered an internal error or misconfiguration and was unable to complete your request."
	}
	data["title"] = title
	data["code"] = code
	data["logo"] = logo
	data["alert"] = alert
	data["probl"] = probl
	data["uri"] = uri
	err := ctx.Render(http.StatusNotFound, "status", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Interview is the handler for the People Interviews page.
func Interview(s *zap.SugaredLogger, ctx echo.Context) error {
	data := initData()
	data["description"] = "demo"
	data["title"] = "demo"
	err := ctx.Render(http.StatusOK, "interview", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// History is the handler for the History page.
func History(s *zap.SugaredLogger, ctx echo.Context) error {
	const lead = "Defacto founded in late February or early March of 1996, as an electronic magazine that wrote about The Scene subculture."
	const h1 = "Our history"
	data := initData()
	data["carousel"] = "#carouselDf2Artpacks"
	data["description"] = lead
	data["logo"] = "The history of Defacto"
	data["h1"] = h1
	data["lead"] = lead
	data["title"] = h1
	err := ctx.Render(http.StatusOK, "history", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Thanks is the handler for the Thanks page.
func Thanks(s *zap.SugaredLogger, ctx echo.Context) error {
	data := initData()
	data["description"] = "Defacto2 thankyous."
	data["h1"] = "Thank you!"
	data["lead"] = "Thanks to the hundreds of people who have contributed to Defacto2 over the decades with file submissions, hard drive donations, interviews, corrections, artwork and monetiary donations!"
	data["title"] = "Thanks!"
	err := ctx.Render(http.StatusOK, "thanks", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// TheScene is the handler for the The Scene page.
func TheScene(s *zap.SugaredLogger, ctx echo.Context) error {
	const h1 = "The Scene?"
	const lead = "Collectively referred to as The Scene, it is a subculture of different computer activities where participants actively share ideas and creations."
	data := initData()
	data["description"] = fmt.Sprint(h1, " ", lead)
	data["logo"] = "The underground"
	data["h1"] = h1
	data["lead"] = lead
	data["title"] = h1
	err := ctx.Render(http.StatusOK, "thescene", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}