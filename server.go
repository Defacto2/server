// The Defacto2 web application built in 2023 on Go.
// (c) 2023 Ben Garrett.
// https://defacto2.net
package main

import (
	"embed"
	"log"
	"os"
	"runtime"

	"github.com/Defacto2/server/cmd"
	"github.com/Defacto2/server/handler"
	"github.com/Defacto2/server/pkg/config"
	"github.com/Defacto2/server/pkg/logger"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/caarlos0/env/v7"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

//go:embed public/texts/defacto2.txt
var brand []byte

//go:embed view/**/*.html
var views embed.FS

//go:embed public/images/*
var images embed.FS

// version generated by GoReleaser ldflags
var version string

func main() {
	// Environment configuration
	configs := config.Config{
		// IsProduction: true,
	}
	if err := env.Parse(&configs, env.Options{
		Prefix: config.EnvPrefix,
	}); err != nil {
		log.Fatalln(err)
	}

	// Command-line arguments
	// By default the webserver runs when no arguments are provided
	if code, err := cmd.Run(version, &configs); err != nil {
		log.Printf("The command given did not work: %s.", err)
		os.Exit(code)
	} else if code >= 0 {
		os.Exit(code)
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
		Import:  &configs,
		Log:     log,
		Brand:   &brand,
		Version: version,
		Images:  images,
		Views:   views,
	}
	e := c.Controller()

	// Start the HTTP server
	go c.StartHTTP(e)

	// Gracefully shutdown the HTTP server
	c.ShutdownHTTP(e)
}
