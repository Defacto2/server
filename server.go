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

	"github.com/Defacto2/server/config"
	"github.com/Defacto2/server/internal/server"
	"github.com/Defacto2/server/logger"
	"github.com/Defacto2/server/postgres"
	"github.com/Defacto2/server/router"
)

//go:embed public/texts/defacto2.txt
var brand []byte

func main() {
	// Startup logo
	if logo := string(brand); len(logo) > 0 {
		fmt.Println(logo)
	}

	// Enviroment configuration
	cfg := config.Config{}
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

	// Echo router/controller instance
	e := router.Route(cfg)

	// Start server with graceful shutdown
	go func() {
		serverAddress := fmt.Sprintf(":%d", cfg.DataPort)
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
