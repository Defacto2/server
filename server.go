// The Defacto2 web application built in 2023.
// (c) 2013 Ben Garrett.
// https://defacto2.net
package main

import (
	"context"
	_ "embed"
	"fmt"
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

	"github.com/bengarrett/df2023/internal/server"
	"github.com/bengarrett/df2023/logger"
	"github.com/bengarrett/df2023/postgres"
	"github.com/bengarrett/df2023/postgres/models"
	"github.com/bengarrett/df2023/router"
	"github.com/bengarrett/df2023/router/html3"
	"github.com/bengarrett/df2023/router/users"
)

//go:embed public/texts/defacto2.txt
var brand []byte

const (
	Timeout = 5 * time.Second
)

type config struct {
	DBPort       int  `env:"PORT" envDefault:"1323"`
	IsProduction bool `env:"PRODUCTION"`
	LogRequests  bool `env:"REQUESTS" envDefault:"false"`
	NoRobots     bool `env:"NOROBOTS" envDefault:"false"` // TODO
}

func main() {
	fmt.Println(string(brand))

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
		fmt.Print("Defacto2 web application")
	default:
		log = logger.Development().Sugar()
		defer log.Sync()
		fmt.Print("Defacto2 web application development")
	}

	// Database
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()
	// SQLBoiler global variant
	boil.SetDB(db)

	// Check the database connection
	if s, err := postgres.Version(); err != nil {
		log.Error("could not obtain the postgres version, is the database online? ", err)
	} else {
		fmt.Print(server.ParsePsVersion(s))
	}

	// Echo instance
	e := echo.New()
	e.HideBanner = true

	// Custom error pages
	e.HTTPErrorHandler = server.CustomErrorHandler

	// HTML templates
	e.Renderer = &router.TemplateRegistry{
		Templates: router.TmplHTML3(),
	}

	// Static images
	e.File("favicon.ico", "public/images/favicon.ico")
	e.Static("/images", "public/images")

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

	// Routes => html3
	html3.Routes("/html3", e)

	// Routes => users
	e.GET("/users", users.GetAllUsers)
	e.POST("/users", users.CreateUser)
	e.GET("/users/:id", users.GetUser)
	e.PUT("/users/:id", users.UpdateUser)
	e.DELETE("/users/:id", users.DeleteUser)

	// Router => downloads
	e.GET("/d/:id", router.Download)

	// Start server with graceful shutdown
	go func() {
		serverAddress := fmt.Sprintf(":%d", cfg.DBPort)
		if err := e.Start(serverAddress); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	const shutdown = 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdown)
	defer func() {
		fmt.Printf("\nDetected Ctrl-C, server will shutdown in %s\n", shutdown)
		select {
		case <-quit:
			cancel()
		case <-ctx.Done():
		}
		if err := e.Shutdown(ctx); err != nil {
			e.Logger.Fatal(err)
		}
		log.Infoln("graceful server shutdown complete")
		signal.Stop(quit)
		cancel()
	}()
}
