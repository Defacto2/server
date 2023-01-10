package main

// https://thedevelopercafe.com/articles/sql-in-go-with-sqlboiler-ac8efc4c5cb8

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/caarlos0/env"
	_ "github.com/lib/pq"
	"github.com/volatiletech/sqlboiler/boil"
	"go.uber.org/zap"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	ps "github.com/bengarrett/df2023/db"
	"github.com/bengarrett/df2023/db/models"
	"github.com/bengarrett/df2023/logger"
	"github.com/bengarrett/df2023/router"
)

const (
	Timeout = 5 * time.Second
)

type config struct {
	DBPort       int  `env:"PORT" envDefault:"1323"`
	IsProduction bool `env:"PRODUCTION"`
	LogRequests  bool `env:"REQUESTS" envDefault:"false"`
}

// TemplateRenderer is a custom html/template renderer for Echo framework
type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {

	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	// Enviroment configuration
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalln(err)
	}

	// Logger
	var log *zap.SugaredLogger
	switch cfg.IsProduction {
	case true:
		log = logger.Production().Sugar()
		defer log.Sync()
		// todo: make a meaningful startup message for file logging
		log.Info("Defacto2 web application")
	default:
		log = logger.Development().Sugar()
		defer log.Sync()
		log.Info("Defacto2 web application development")
	}

	// Database
	ctx := context.Background()
	db, err := ps.ConnectDB()
	if err != nil {
		log.Fatalln(err)
	}
	// SQLBoiler global variant
	boil.SetDB(db)

	// Check the database connection
	if ver, err := ps.Version(); err != nil {
		log.Error("could not obtain the postgres version", err)
	} else {
		fmt.Println(ver)
	}

	// Echo instance
	e := echo.New()
	e.HideBanner = true

	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("*.html")),
	}
	e.Renderer = renderer

	// Middleware
	e.Use(middleware.RemoveTrailingSlashWithConfig(middleware.TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	}))
	if cfg.LogRequests {
		e.Use(middleware.Logger())
	}
	e.Use(middleware.Recover())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: Timeout,
	}))

	// Route => handler
	e.GET("/", func(c echo.Context) error {
		count, err := models.Files().Count(ctx, db)
		if err != nil {
			log.Fatalln(err)
		}
		return c.String(http.StatusOK, fmt.Sprintf("Hello, World!\nThere are %d files\n",
			count))
	})

	e.GET("/oooo", func(c echo.Context) error {
		return c.Render(http.StatusOK, "template.html", map[string]interface{}{
			"name": "OoooOoooOoooO",
		})
	}).Name = "foobar"

	// Routes
	e.GET("/users", router.GetAllUsers)
	e.POST("/users", router.CreateUser)
	e.GET("/users/:id", router.GetUser)
	e.PUT("/users/:id", router.UpdateUser)
	e.DELETE("/users/:id", router.DeleteUser)

	// Start server with graceful shutdown
	go func() {
		addr := fmt.Sprintf(":%d", cfg.DBPort)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	// TODO: confirm shutdown with a prompt requiring an all caps YES.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	log.Infoln("graceful server shutdown complete")
}
