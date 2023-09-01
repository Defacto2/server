// The Defacto2 Go web application built in 2023.
// (c) 2023 Ben Garrett.
// https://github.com/Defacto2/server
// https://defacto2.net

package main

import (
	"context"
	"embed"
	"errors"
	"os"
	"runtime"

	"github.com/Defacto2/server/cmd"
	"github.com/Defacto2/server/handler"
	"github.com/Defacto2/server/model"
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

//go:embed view/**/*
var view embed.FS

// version is generated by the GoReleaser ldflags.
var version string

var ErrVersion = errors.New("could not obtain the database server version value")

func main() {
	// Logger
	// Use the development log until the environment vars are parsed
	zlog := logger.CLI().Sugar()

	// Environment configuration
	configs := config.Config{}
	if err := env.Parse(
		&configs, env.Options{Prefix: config.EnvPrefix}); err != nil {
		zlog.Fatalf("Environment variable probably contains an invalid value: %s.", err)
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
		zlog.Errorf("The command given did not work: %s.", err)
		os.Exit(code)
	} else if code >= exitProgram {
		os.Exit(code)
	}

	// Go runtime customizations
	if i := configs.MaxProcs; i > 0 {
		runtime.GOMAXPROCS(int(i))
	}

	// Configuration sanity checks
	configs.Checks(zlog)

	// Setup the logger
	switch configs.IsProduction {
	case true:
		if err := configs.LogStorage(); err != nil {
			zlog.Fatalf("The server cannot save any logs: %s.", err)
		}
		zlog = logger.Production(configs.LogDir).Sugar()
	default:
		zlog.Debug("The server is running in the DEVELOPMENT MODE.")
		zlog = logger.Development().Sugar()
	}

	// Cached global vars will go here, to avoid the garbage collection
	// They should be lockable

	// Echo router and controller instance
	server := handler.Configuration{
		Brand:   &brand,
		Import:  &configs,
		ZLog:    zlog,
		Public:  public,
		Version: version,
		View:    view,
	}

	// Database
	if err := repairDB(server); err != nil {
		if errors.Is(err, ErrVersion) {
			zlog.Errorf("The database server version could not be obtained, " +
				"is the database server down?")
		} else {
			zlog.Errorf("Could not initialize the database data: %s.", err)
		}
	}

	// Controllers and routes
	e := server.Controller()

	// Start the HTTP server
	go server.StartHTTP(e)

	// Gracefully shutdown the HTTP server
	server.ShutdownHTTP(e)
}

// repairDB, on startup check the database connection and make any data corrections.
func repairDB(server handler.Configuration) error {
	db, err := postgres.ConnectDB()
	if err != nil {
		return err
	}
	defer db.Close()
	var psql postgres.Version
	ctx := context.Background()
	if err := psql.Query(); err != nil {
		return ErrVersion
	}
	if err := model.RepairReleasers(ctx, db); err != nil {
		return err
	}
	return nil
}
