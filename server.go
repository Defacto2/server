// The Defacto2 web application built in 2023 on Go.
// (c) 2023 Ben Garrett.
// https://defacto2.net
package main

//go:generate sqlboiler --config ".sqlboiler.toml" --wipe --no-hooks --add-soft-deletes psql

import (
	"bufio"
	"embed"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/Defacto2/server/cmd"
	"github.com/Defacto2/server/handler"
	"github.com/Defacto2/server/pkg/config"
	"github.com/Defacto2/server/pkg/logger"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/caarlos0/env"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

//go:embed public/texts/defacto2.txt
var brand []byte

//go:embed view/**/*.html
var views embed.FS

//go:embed public/images/*
var images embed.FS

var (
	version = ""
	date    = "" //nolint:gochecknoglobals
)

func main() {
	// Environment configuration
	configs := config.Config{
		// IsProduction: true,
	}
	if err := env.Parse(&configs); err != nil {
		log.Fatalln(err)
	}

	// Command-line arguments
	// By default the webserver runs when no arguments are provided
	b := cmd.Build{Version: version, Date: date}
	if code, err := b.Run(); err != nil {
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

	// Startup logo
	if logo := string(brand); len(logo) > 0 {
		w := bufio.NewWriter(os.Stdout)
		if _, err := fmt.Fprintf(w, "%s\n\n", logo); err != nil {
			log.Warnf("Could not print the brand logo: %s.", err)
		}
		w.Flush()
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

	// Start the HTTP server
	go c.StartHTTP(e)

	// Gracefully shutdown the HTTP server
	c.ShutdownHTTP(e)
}
