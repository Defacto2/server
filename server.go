// The Defacto2 Go web application built in 2023.
// (c) 2023 Ben Garrett.
// https://github.com/Defacto2/server
// https://defacto2.net

package main

import (
	"embed"
	"os"
	"runtime"

	"github.com/Defacto2/server/cmd"
	"github.com/Defacto2/server/handler"
	"github.com/Defacto2/server/pkg/config"
	"github.com/Defacto2/server/pkg/logger"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/caarlos0/env/v7"
	_ "github.com/lib/pq"
)

//go:embed public/text/defacto2.txt
var brand []byte

//go:embed public/**/*
var public embed.FS

//go:embed view/**/*.html
var views embed.FS

// version is generated by the GoReleaser ldflags.
var version string

func main() {
	// Logger
	// Use the development log until the environment vars are parsed
	log := logger.Development().Sugar()

	// Environment configuration
	configs := config.Config{}
	if err := env.Parse(
		&configs, env.Options{Prefix: config.EnvPrefix}); err != nil {
		log.Fatalf("Environment variable probably contains an invalid value: %s.", err)
	}
	// Any hardcoded overrides can be placed in here,
	// but they must be commented out in PRODUCTION
	// configs.IsProduction = true // This will enable the production logger
	// configs.HTTPSRedirect = true // This requires HTTPS certificates to be installed and configured
	// configs.NoRobots = true // This will disable search engine crawling
	// configs.LogRequests = true // This will log all HTTP requests to the server or stdout

	// Command-line arguments
	// By default the web server runs when no arguments are provided
	const exitProgram = 0
	if code, err := cmd.Run(version, &configs); err != nil {
		log.Errorf("The command given did not work: %s.", err)
		os.Exit(code)
	} else if code >= exitProgram {
		os.Exit(code)
	}

	// Go runtime customizations
	if i := configs.MaxProcs; i > 0 {
		runtime.GOMAXPROCS(int(i))
	}

	// Setup the logger
	switch configs.IsProduction {
	case true:
		if err := configs.LogStorage(); err != nil {
			log.Fatalf("The server cannot save any logs: %s.", err)
		}
		log = logger.Production(configs.LogDir).Sugar()
	default:
		log.Debug("The server is running in the DEVELOPMENT MODE.")
	}

	// Configuration sanity checks
	configs.Checks(log)

	// Cached global vars will go here, to avoid the garbage collection
	// They should be lockable

	// Echo router and controller instance
	server := handler.Configuration{
		Brand:   &brand,
		Import:  &configs,
		ZLog:    log,
		Public:  public,
		Version: version,
		Views:   views,
	}

	// Database
	db, err := postgres.ConnectDB()
	if err != nil {
		server.DatbaseErr = true
		log.Errorf("Could not connect to the database: %s.", err)
	}
	defer db.Close()

	e := server.Controller()

	// Start the HTTP server
	go server.StartHTTP(e)

	// Gracefully shutdown the HTTP server
	server.ShutdownHTTP(e)
}
