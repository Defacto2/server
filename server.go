// The Defacto2 web application built in 2023.
// (c) 2023 Ben Garrett.
// https://defacto2.net
package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/caarlos0/env"
	_ "github.com/lib/pq"
	"github.com/urfave/cli/v2"
	"github.com/volatiletech/sqlboiler/boil"
	"go.uber.org/zap"

	"github.com/Defacto2/server/internal/server"
	"github.com/Defacto2/server/pkg/config"
	"github.com/Defacto2/server/pkg/logger"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/router"
)

//go:embed public/texts/defacto2.txt
var brand []byte

//go:embed public/views/html3/*.html
var views embed.FS

//go:embed public/images/*
var images embed.FS

func main() {
	// Enviroment configuration
	configs := config.Config{}
	//configs.IsProduction = true
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
	if args := len(os.Args[1:]); args > 0 {
		greet()
	}

	// Startup logo
	if logo := string(brand); len(logo) > 0 {
		if _, err := fmt.Printf("%s\n\n", logo); err != nil {
			log.Warnf("Could not print the brand logo: %s.", err)
		}
	}

	// Database
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		log.Errorf("Could not connect to the database: %s.", err)
	}
	defer db.Close()

	// Cached global vars will go here to avoid the garbage collection.
	// They should be lockable.

	// SQLBoiler global variant
	boil.SetDB(db)

	// Echo router/controller instance
	e := router.Router{
		Configs: configs,
		Log:     log,
		Images:  images,
		Views:   views,
	}.Controller()

	// Start server with graceful shutdown
	go func() {
		const mark = `⇨ `

		// Check the database connection
		if s, err := postgres.Version(); err != nil {
			log.Warnln("Could not obtain the PostgreSQL server version. Is the database online?")
		} else {
			fmt.Printf("%sDefacto2 web application %s.\n", mark, server.ParsePsVersion(s))
		}

		fmt.Printf("%s%d active routines sharing %d usable threads on %d CPU cores.\n", mark,
			runtime.NumGoroutine(), runtime.GOMAXPROCS(-1), runtime.NumCPU())

		fmt.Printf("%sCompiled with Go %s.\n", mark, runtime.Version()[2:])
		if configs.IsProduction {
			fmt.Printf("%sserver logs are found in: %s\n", mark, configs.ConfigDir)
		}

		// x, _ := models.GroupsTotalCount(ctx, db)
		// y, _ := models.Tester(ctx, db)
		// fmt.Printf("Group compare: %d total vs %d tester.\n", x, y)

		serverAddress := fmt.Sprintf(":%d", configs.HTTPPort)
		err := e.Start(serverAddress)
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server could not start: %s.", err)
		}
		// nothing should be placed here
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	const shutdown = 5
	ctx, cancel := context.WithTimeout(context.Background(), shutdown*time.Second)
	defer func() {
		const alert = "Detected Ctrl-C, server will shutdown in "
		if err := log.Sync(); err != nil {
			log.Warnf("Could not sync the log before shutdown: %s.\n", err)
		}
		fmt.Printf("\n%s%s", alert, shutdown*time.Second)
		count := shutdown
		for range time.Tick(1 * time.Second) {
			count--
			fmt.Printf("\r%s%ds", alert, count)
			if count <= 0 {
				fmt.Printf("\r%s%ds\n", alert, count)
				break
			}
		}
		select {
		case <-quit:
			cancel()
		case <-ctx.Done():
		}
		if err := e.Shutdown(ctx); err != nil {
			log.Fatalf("Server shutdown caused an error: %w.", err)
		}
		log.Infoln("Server shutdown complete.")
		if err := log.Sync(); err != nil {
			log.Warnf("Could not sync the log before shutdown: %s.\n", err)
		}
		signal.Stop(quit)
		cancel()
	}()
}

// Command-line arguments handler placeholder.
// TODO: https://cli.urfave.org/v2/examples/full-api-example/
func greet() {
	app := &cli.App{
		Name:     "Defacto2 webserver",
		Version:  Commit,
		Compiled: time.Now(),
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Ben Garrett",
				Email: "contact@defacto2.net",
			},
		},
		Copyright: "(c) 2023 Defacto2 & Ben Garrett",
		HelpName:  "server",
		Usage:     "Serve the Defacto2 website",
		UsageText: "server [options]",
		ArgsUsage: "[args and such]",
	}
	app.EnableBashCompletion = true
	app.HideHelp = false
	app.HideVersion = false
	app.Suggest = true
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
	fmt.Println(debug.ReadBuildInfo())
	os.Exit(0)
}

/*
https://pkg.go.dev/runtime/debug
https://shibumi.dev/posts/go-18-feature/
build	-compiler=gc
build	CGO_ENABLED=1
build	CGO_CFLAGS=
build	CGO_CPPFLAGS=
build	CGO_CXXFLAGS=
build	CGO_LDFLAGS=
build	GOARCH=amd64
build	GOOS=linux
build	GOAMD64=v1
build	vcs=git
build	vcs.revision=7e22e19e829d84170072d2459e5870876df495ed
build	vcs.time=2022-04-03T16:59:50Z
build	vcs.modified=false
*/
var Commit = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}
	return "n/a"
}()
