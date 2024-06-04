// Package main is the entry point for the Defacto2 server application.
//
// Use the Task runner / build tool (https://taskfile.dev) to build or run the source code.
// $ task --list
//
// Repository: 	https://github.com/Defacto2/server
// Website:		https://defacto2.net
// License:
//
// © Defacto2, 2024
package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/Defacto2/server/cmd"
	"github.com/Defacto2/server/handler"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/zaplog"
	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/model/fix"
	"github.com/caarlos0/env/v10"
	_ "github.com/lib/pq"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/zap"
)

//go:embed public/text/defacto2.txt
var brand []byte

//go:embed public/**/*
var public embed.FS

//go:embed view/**/*
var view embed.FS

// version is generated by the GoReleaser ldflags.
var version string

var (
	ErrCmd = errors.New("the command given did not work")
	ErrDB  = errors.New("could not initialize the database data")
	ErrEnv = errors.New("environment variable probably contains an invalid value")
	ErrFS  = errors.New("the directories repair broke")
	ErrLog = errors.New("the server cannot save any logs")
	ErrVer = errors.New("postgresql version request failed")
)

// main is the entry point for the application.
// By default the web server runs when no arguments are provided.
// Otherwise, the command-line arguments are parsed and the application exits.
func main() {
	logger, configs := environmentVars()
	if code := parseFlags(logger, configs); code >= 0 {
		os.Exit(code)
	}

	defer sanityChecks(logger, configs)
	defer repairChecks(logger, configs)

	logger = serverLog(configs)
	website := newInstance(configs)
	router := website.Controller(logger)
	website.Info(logger)
	if err := website.Start(router, logger, configs); err != nil {
		logger.Fatalf("%s: please check the environment variables.", err)
	}

	w := os.Stdout
	go func() {
		localIPs, err := configs.Addresses()
		if err != nil {
			logger.Errorf("%s: %s", ErrEnv, err)
		}
		fmt.Fprintf(w, "%s\n", localIPs)
	}()

	website.ShutdownHTTP(router, logger)
}

// environmentVars is used to parse the environment variables and set the Go runtime.
func environmentVars() (*zap.SugaredLogger, config.Config) {
	logger := zaplog.Development().Sugar()
	configs := config.Config{}
	if err := env.Parse(&configs); err != nil {
		logger.Fatalf("%w: %s", ErrEnv, err)
	}
	configs.Override()

	if i := configs.MaxProcs; i > 0 {
		runtime.GOMAXPROCS(int(i))
	}
	return logger, configs
}

// newInstance is used to create the server controller instance.
func newInstance(configs config.Config) handler.Configuration {
	c := handler.Configuration{
		Brand:       brand,
		Environment: configs,
		Public:      public,
		Version:     version,
		View:        view,
	}
	if c.Version == "" {
		c.Version = cmd.Commit("")
	}
	c.RecordCount = recordCount()
	return c
}

// parseFlags is used to parse the commandline arguments.
// If an error is returned, the application will exit with the error code.
// Otherwise, a negative value is returned to indicate the application should continue.
func parseFlags(logger *zap.SugaredLogger, configs config.Config) int {
	if logger == nil {
		return -1
	}
	code, err := cmd.Run(version, &configs)
	if err != nil {
		logger.Errorf("%s: %s", ErrCmd, err)
		return int(code)
	}
	useExitCode := code >= cmd.ExitOK
	if useExitCode {
		return int(code)
	}
	return -1
}

// sanityChecks is used to perform a number of sanity checks on the file assets and database.
// These are skipped if the FastStart environment variable is set.
func sanityChecks(logger *zap.SugaredLogger, configs config.Config) {
	if configs.FastStart || logger == nil {
		return
	}
	if err := configs.Checks(logger); err != nil {
		logger.Errorf("%s: %s", ErrEnv, err)
	}
	checks(logger, configs.ReadMode)
	conn, err := postgres.New()
	if err != nil {
		logger.Errorf("%s: %s", ErrDB, err)
		return
	}
	_ = conn.Check(logger)
}

// checks is used to confirm the required commands are available.
// These are skipped if readonly is true.
func checks(logger *zap.SugaredLogger, readonly bool) {
	if logger == nil || readonly {
		return
	}
	var buf strings.Builder
	for i, name := range command.Lookups() {
		if err := command.LookCmd(name); err != nil {
			buf.WriteString("\n\t\t\tmissing: " + name)
			buf.WriteString("\t" + command.Infos()[i])
		}
	}
	if buf.Len() > 0 {
		logger.Warnln("The following commands are required for the server to run in WRITE MODE",
			"\n\t\t\tThese need to be installed and accessible on the system path:"+
				"\t\t\t"+buf.String())
	}
	if err := command.LookupUnrar(); err != nil {
		if errors.Is(err, command.ErrVers) {
			logger.Warnf("Found unrar but " +
				"could not find unrar by Alexander Roshal, " +
				"is unrar-free mistakenly installed?")
			return
		}
		logger.Warnf("%s: %s", ErrCmd, err)
	}
}

// repairChecks is used to fix any known issues with the file assets and the database entries.
// These are skipped if the FastStart environment variable is set.
func repairChecks(logger *zap.SugaredLogger, configs config.Config) {
	if configs.FastStart || logger == nil {
		return
	}
	if err := configs.RepairFS(logger); err != nil {
		logger.Errorf("%s: %s", ErrFS, err)
	}
	if err := repairDB(logger); err != nil {
		repairdb(logger, err)
	}
}

// serverLog is used to setup the logger for the server and print the startup message.
func serverLog(configs config.Config) *zap.SugaredLogger {
	logger := zaplog.Development().Sugar()
	const welcome = "Welcome to the local Defacto2 web application."
	logger.Info(welcome)
	mode := "read-only mode"
	if !configs.ReadMode {
		mode = "write mode"
	}
	switch configs.ProductionMode {
	case true:
		if err := configs.LogStore(); err != nil {
			logger.Fatalf("%w: %s", ErrLog, err)
		}
		logger = zaplog.Production(configs.LogDir).Sugar()
		s := "The server is running in a "
		s += strings.ToUpper("production, "+mode) + "."
		logger.Info(s)
	default:
		s := "The server is running in a "
		s += strings.ToUpper("development, "+mode) + "."
		logger.Warn(s)
	}
	return logger
}

// repairDB on startup checks the database connection and make any data corrections.
func repairDB(logger *zap.SugaredLogger) error {
	if logger == nil {
		return fmt.Errorf("%w: %s", ErrLog, "no logger")
	}
	db, err := postgres.ConnectDB()
	if err != nil {
		return fmt.Errorf("postgres.ConnectDB: %w", err)
	}
	defer db.Close()
	var ver postgres.Version
	if err := ver.Query(); err != nil {
		return ErrVer
	}
	ctx := context.Background()
	err = fix.All.Run(ctx, logger, db)
	if err != nil {
		return fmt.Errorf("fix.All.Run: %w", err)
	}
	return nil
}

// repairdb is used to log the database repair error.
func repairdb(logger *zap.SugaredLogger, err error) {
	if logger == nil || err == nil {
		return
	}
	if errors.Is(err, ErrVer) {
		logger.Warnf("A %s, is the database server down?", ErrVer)
	} else {
		logger.Errorf("%s: %s", ErrDB, err)
	}
}

// recordCount returns the number of records in the database.
func recordCount() int {
	db, err := postgres.ConnectDB()
	if err != nil {
		return 0
	}
	defer db.Close()
	ctx := context.Background()
	fs, err := models.Files(qm.Where(model.ClauseNoSoftDel)).Count(ctx, db)
	if err != nil {
		return 0
	}
	return int(fs)
}
