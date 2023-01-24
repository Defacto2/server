// The Defacto2 web application built in 2023 on Go.
// (c) 2023 Ben Garrett.
// https://defacto2.net
package main

//go:generate sqlboiler --config ".sqlboiler.toml" --wipe --no-hooks --add-soft-deletes psql

import (
	"embed"
	"fmt"
	"log"
	"runtime"

	"github.com/caarlos0/env"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/Defacto2/server/cmd"
	"github.com/Defacto2/server/handler"
	"github.com/Defacto2/server/pkg/config"
	"github.com/Defacto2/server/pkg/logger"
	"github.com/Defacto2/server/pkg/postgres"
)

//go:embed public/texts/defacto2.txt
var brand []byte

//go:embed view/html3/*.html
var views embed.FS

//go:embed public/images/*
var images embed.FS

func main() {
	// Enviroment configuration
	configs := config.Config{
		IsProduction: true,
	}
	if err := env.Parse(&configs); err != nil {
		log.Fatalln(err)
	}

	// Go runtime customizations
	if i := configs.MaxProcs; i > 0 {
		runtime.GOMAXPROCS(int(i))
	}

	// Logger
	var log *zap.SugaredLogger
	switch configs.IsProduction {
	case true:
		if err := configs.LogStorage(); err != nil {
			log.Errorf("The server cannot save any logs: %s.", err)
		}
		log = logger.Production(configs.ConfigDir).Sugar()
	default:
		log = logger.Development().Sugar()
		log.Debug("The server is running in the development mode.")
	}

	// Command-line arguments
	// By default the webserver runs when no arguments are provided
	if err := cmd.Run(); err != nil {
		log.Fatalln(err) // TODO:
	}

	// Startup logo
	if logo := string(brand); len(logo) > 0 {
		if _, err := fmt.Printf("%s\n\n", logo); err != nil {
			log.Warnf("Could not print the brand logo: %s.", err)
		}
	}

	// Database
	db, err := postgres.ConnectDB()
	if err != nil {
		log.Errorf("Could not connect to the database: %s.", err)
	}
	defer db.Close()

	// Cached global vars will go here to avoid the garbage collection.
	// They should be lockable.

	// Echo router/controller instance
	c := handler.Configuration{
		Import: &configs,
		Log:    log,
		Images: images,
		Views:  views,
	}
	e := c.Controller()

	// Placeholder API example
	//e.GET("/api/get-filename", api.GetFilename)

	// Start the HTTP server
	go c.StartHTTP(e)

	// Gracefully shutdown the HTTP server
	c.ShutdownHTTP(e)
}
