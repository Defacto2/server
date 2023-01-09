package main

// https://thedevelopercafe.com/articles/sql-in-go-with-sqlboiler-ac8efc4c5cb8

import (
	"context"
	"fmt"
	"log"
	"net/http"
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

	// Echo instance
	e := echo.New()
	e.Use(middleware.AddTrailingSlashWithConfig(middleware.TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	}))

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: Timeout,
	}))

	// Open handle to database like normal
	ctx := context.Background()
	db, err := ps.ConnectDB()
	if err != nil {
		log.Fatalln(err)
	}

	boil.SetDB(db) // SQLBoiler global variant

	count, err := models.Files().Count(ctx, db)
	if err != nil {
		log.Fatalln(err)
	}

	// Route => handler
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, fmt.Sprintf("Hello, World!\nThere are %d files\n",
			count))
	})

	// Routes
	e.GET("/users", router.GetAllUsers)
	e.POST("/users", router.CreateUser)
	e.GET("/users/:id", router.GetUser)
	e.PUT("/users/:id", router.UpdateUser)
	e.DELETE("/users/:id", router.DeleteUser)

	// Start server
	if ver, err := ps.Version(); err != nil {
		log.Info("could not obtain the postgres version", err)
	} else {
		fmt.Println(ver)
	}

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", cfg.DBPort)))
}
