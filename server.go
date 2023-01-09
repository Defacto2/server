package main

// https://thedevelopercafe.com/articles/sql-in-go-with-sqlboiler-ac8efc4c5cb8

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	"github.com/volatiletech/sqlboiler/boil"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/bengarrett/df2023/db/models"
	"github.com/bengarrett/df2023/router"
)

const (
	port    = "1323"
	timeout = 5 * time.Second
)

func main() {
	// Echo instance
	e := echo.New()
	e.Use(middleware.AddTrailingSlashWithConfig(middleware.TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	}))

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: timeout,
	}))

	// Open handle to database like normal
	ctx := context.Background()
	db := connectDB()

	boil.SetDB(db) // SQLBoiler global variant

	count, err := models.Files().Count(ctx, db)
	if err != nil {
		log.Fatalln(err)
	}

	// Route => handler
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, fmt.Sprintf("Hello, World!\nThere are %d files\n", count))
	})

	// Routes
	e.GET("/users", router.GetAllUsers)
	e.POST("/users", router.CreateUser)
	e.GET("/users/:id", router.GetUser)
	e.PUT("/users/:id", router.UpdateUser)
	e.DELETE("/users/:id", router.DeleteUser)

	// Start server
	e.Logger.Fatal(e.Start(":" + port))
}

func connectDB() *sql.DB {
	db, err := sql.Open("postgres", "postgres://root:example@localhost:5432/defacto2-ps?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	return db
}
