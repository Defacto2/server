package main

/*
Package main is the entry point for the Defacto2 server application.

Use the Task runner / build tool (https://taskfile.dev) to build or run the source code.
$ task --list

Repository: 	https://github.com/Defacto2/server
Website:		https://defacto2.net
License:

© Defacto2, 2024
*/

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io"

	//_ "net/http/pprof" // pprof is used for profiling and can be commented out.
	"os"
	"runtime"
	"slices"
	"strings"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/flags"
	"github.com/Defacto2/server/handler"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/zaplog"
	"github.com/caarlos0/env/v11"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

var (
	//go:embed public/text/defacto2.txt
	brand []byte
	//go:embed public/**/*
	public embed.FS
	//go:embed view/**/*
	view    embed.FS
	version string // version is generated by the GoReleaser ldflags.
)

var ErrLog = errors.New("cannot save logs")

// Main is the entry point for the application.
// By default the web server runs when no arguments are provided,
// otherwise, the command-line arguments are parsed and the application exits.
func main() {
	const exit = 0
	// initialize a temporary logger, get and print the environment variable configurations.
	logger, configs := environmentVars()
	if exitCode := parseFlags(logger, configs); exitCode >= exit {
		os.Exit(exitCode)
	}
	var w io.Writer = os.Stdout
	if configs.Quiet {
		w = io.Discard
	}
	fmt.Fprintf(w, "%s\n", configs)

	// connect to the database and perform some repairs and sanity checks.
	// if the database is cannot connect, the web server will continue.
	db, err := postgres.Open()
	if err != nil {
		logger.Errorf("main could not initialize the database data: %s", err)
	}
	defer db.Close()
	var database postgres.Version
	if err := database.Query(db); err != nil {
		logger.Errorf("postgres version query: %w", err)
	}
	config.SanityTmpDir()
	fmt.Fprintln(w)

	// start the web server and the sugared logger.
	ctx := context.Background()
	website := newInstance(ctx, db, configs)
	logger = serverLog(configs, website.RecordCount)
	router := website.Controller(db, logger)
	website.Info(logger, w)
	if err := website.Start(router, logger, configs); err != nil {
		logger.Fatalf("%s: please check the environment variables", err)
	}

	go func() {
		// get the owner and group of the current process and print them to the console.
		groups, usr, err := helper.Owner()
		if err != nil {
			logger.Errorf("owner in main: %s", err)
		}
		clean := slices.DeleteFunc(groups, func(e string) bool {
			return e == ""
		})
		fmt.Fprintf(w, "Running as %s for the groups, %s.\n",
			usr, strings.Join(clean, ","))
		// get the local IP addresses and print them to the console.
		localIPs, err := configs.Addresses()
		if err != nil {
			logger.Errorf("configs addresses in main: %s", err)
		}
		fmt.Fprintf(w, "%s\n", localIPs)
	}()

	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	// shutdown the web server after a signal is received.
	website.ShutdownHTTP(router, logger)
}

// environmentVars is used to parse the environment variables and set the Go runtime.
// Defaults are used if the environment variables are not set.
func environmentVars() (*zap.SugaredLogger, config.Config) {
	logger := zaplog.Status().Sugar()
	configs := config.Config{
		Compression:   true,
		DatabaseURL:   postgres.DefaultURL,
		HTTPPort:      config.HTTPPort,
		ProdMode:      true,
		ReadOnly:      true,
		SessionMaxAge: config.SessionHours,
	}
	if err := env.Parse(&configs); err != nil {
		logger.Fatalf("could not parse the environment variable, it probably contains an invalid value: %s", err)
	}
	configs.Override()
	if i := configs.MaxProcs; i > 0 {
		runtime.GOMAXPROCS(int(i))
	}
	return logger, configs
}

// newInstance is used to create the server controller instance.
func newInstance(ctx context.Context, db *sql.DB, configs config.Config) handler.Configuration {
	c := handler.Configuration{
		Brand:       brand,
		Environment: configs,
		Public:      public,
		Version:     version,
		View:        view,
	}
	if c.Version == "" {
		c.Version = flags.Commit("")
	}
	if ctx != nil && db != nil {
		c.RecordCount = config.RecordCount(ctx, db)
	}
	return c
}

// parseFlags is used to parse the commandline arguments.
// If an error is returned, the application will exit with the error code.
// Otherwise, a negative value is returned to indicate the application should continue.
func parseFlags(logger *zap.SugaredLogger, configs config.Config) int {
	if logger == nil {
		return -1
	}
	code, err := flags.Run(version, &configs)
	if err != nil {
		logger.Errorf("run command, parse flags: %s", err)
		return int(code)
	}
	useExitCode := code >= flags.ExitOK
	if useExitCode {
		return int(code)
	}
	return -1
}

// serverLog is used to setup the logger for the server and print the startup message.
func serverLog(configs config.Config, count int) *zap.SugaredLogger {
	logger := zaplog.Timestamp().Sugar()
	const welcome = "Welcome to the Defacto2 web application"
	switch {
	case count == 0:
		logger.Error(welcome + " with no database records")
	case config.MinimumFiles > count:
		logger.Warnf(welcome+" with only %d records, expecting at least %d+", count, config.MinimumFiles)
	default:
		logger.Infof(welcome+" containing %d records", count)
	}
	if configs.ProdMode {
		if err := configs.LogStore(); err != nil {
			logger.Fatalf("%w using server log: %s", ErrLog, err)
		}
		logger = zaplog.Store(configs.AbsLog).Sugar()
	}
	return logger
}
