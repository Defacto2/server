package config

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres"
)

// Checks runs a number of sanity checks for the environment variable configurations.
func (c *Config) Checks(ctx context.Context, sl *slog.Logger) error {
	const msg = "Config directory"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	c.checkHTTP(ctx, sl)

	c.checkHTTPS(ctx, sl)

	c.production(sl)

	const key = "check"
	// Check the download, preview and thumbnail directories.
	if err := CheckDir(dir.Directory(c.AbsDownload), "downloads"); err != nil {
		s := helper.Capitalize(err.Error())
		sl.Error(msg, slog.String(key, s))
	}
	if err := CheckDir(dir.Directory(c.AbsPreview), "previews"); err != nil {
		s := helper.Capitalize(err.Error())
		sl.Error(msg, slog.String(key, s))
	}
	if err := CheckDir(dir.Directory(c.AbsThumbnail), "thumbnails"); err != nil {
		s := helper.Capitalize(err.Error())
		sl.Error(msg, slog.String(key, s))
	}
	if err := CheckDir(dir.Directory(c.AbsOrphaned), "orphaned"); err != nil {
		s := helper.Capitalize(err.Error())
		sl.Error(msg, slog.String(key, s))
	}
	if err := CheckDir(dir.Directory(c.AbsExtra), "extra"); err != nil {
		s := helper.Capitalize(err.Error())
		sl.Error(msg, slog.String(key, s))
	}

	// configuration reminders
	inf := "Information"
	if c.NoCrawl {
		s := "Disallow search engine crawling is enabled"
		sl.Warn(inf, slog.String(key, s))
	}
	if c.ReadOnly {
		s := "The server is running in read-only mode, edits to the database are not allowed"
		sl.Warn(inf, slog.String(key, s))
	}

	if err := c.checkLogDir(sl); err != nil {
		s := helper.Capitalize(err.Error())
		sl.Error(msg, slog.String(key, s))
	}

	c.commands(ctx, sl)

	conn, err := postgres.New()
	if err != nil {
		sl.Error(msg,
			slog.String("issue", "sanity checks could not initialize the database data"),
			slog.Any("error", err))
	}
	if err := conn.Validate(sl); err != nil {
		panic(fmt.Errorf("%s conn validate: %w", msg, err))
	}

	return nil
}

// commands is used to confirm the required commands are available.
// These are skipped if readonly is true.
func (c *Config) commands(ctx context.Context, sl *slog.Logger) {
	const msg = "command checks"
	if err := nils.Check(sl); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}

	attrs := make([]slog.Attr, 0, len(command.Lookups))
	for i, file := range command.Lookups {
		if _, err := command.Lookup(file); err != nil {
			attrs = append(attrs, slog.String(file, command.Infos[i]))
		}
	}
	if len(attrs) > 0 {
		const msg = "The following tools are required for the server to run in WRITE MODE. " +
			"These need to be installed and accessible on the system path."
		sl.Warn(msg)
		for n, attr := range attrs {
			sl.Warn(fmt.Sprintf("missing tool #%2d", n),
				slog.String("command", attr.Key),
				slog.String("detail", attr.Value.String()))
		}
	}

	if err := command.LookupUnrar(ctx); err != nil {
		if errors.Is(err, command.ErrVersion) {
			const msg = "Found unrar but it is not authored by Alexander Roshal."
			sl.Warn(msg, slog.String("command", command.Lookups[10]),
				slog.String("detail", "is unrar-free mistakenly installed and using the command?"))
			return
		}
		sl.Warn("command unrar", slog.Any("error", err))
	}
}

func (c *Config) Warn(sl *slog.Logger, msg, name string, err error) {
	if sl == nil || err == nil {
		return
	}
	sl.Warn(msg, slog.String("name", name), slog.Any("error", err))
}

// checkHTTP logs a fatal error if the HTTP port is invalid.
func (c *Config) checkHTTP(ctx context.Context, sl *slog.Logger) {
	const msg = "check http port"
	if err := nils.Check(ctx, sl); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}

	if c.HTTPPort == 0 {
		return
	}

	if err := c.HTTPPort.Check(); err != nil {
		c.misconfigPort(ctx, sl, msg, "http_port", err)
	}
}

// The production mode checks when not in read-only mode.
// It expects the server to be configured with OAuth2 and Google IDs.
// The server should be running over HTTPS and not unencrypted HTTP.
func (c *Config) production(sl *slog.Logger) {
	const msg = "production mode"
	if err := nils.Check(sl); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	if !bool(c.ProdMode) || bool(c.ReadOnly) {
		return
	}
	if c.GoogleClientID == "" {
		c.Warn(sl, msg, "GoogleClientID", ErrNoOAuth2)
	}
	if c.GoogleIDs == "" && len(c.GoogleAccounts) == 0 {
		c.Warn(sl, msg, "GoogleIDs, GoogleAccounts", ErrNoAccounts)
	}
	if c.SessionMaxAge == 0 {
		c.Warn(sl, msg, "SessionMaxAge", ErrSession)
	}
}

// checkHTTPS logs a fatal error if the HTTPS port is invalid.
func (c *Config) checkHTTPS(ctx context.Context, sl *slog.Logger) {
	const msg = "check https port"
	if err := nils.Check(ctx, sl); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}

	if c.TLSPort == 0 {
		return
	}

	if err := c.TLSPort.Check(); err != nil {
		c.misconfigPort(ctx, sl, msg, "tls_port", err)
	}
}

func (c *Config) misconfigPort(ctx context.Context, sl *slog.Logger, msg, key string, err error) {
	if err := nils.Check(ctx, sl); err != nil {
		panic(fmt.Errorf("config fatal port: %w", err))
	}

	var port uint16
	port = uint16(c.HTTPPort)

	inf := "HTTP"
	if msg == "tls_port" {
		inf = "HTTPS"
		port = uint16(c.TLSPort)
	}

	value := "The server cannot use the system port"
	if errors.Is(err, ErrPortMax) {
		value = "The server cannot use the " + inf + " port"
	}

	if errors.Is(err, ErrPortMax) || errors.Is(err, ErrPortSys) {
		logs.Fatal(ctx, sl, msg, slog.String("issue", value), slog.Int(key, int(port)),
			slog.String("error", err.Error()))
	}
}

// checkLogDir runs checks against the configured log directory.
// If no log directory is configured, a default directory is used.
func (c *Config) checkLogDir(sl *slog.Logger) error {
	const msg = "setup log directory"
	const format = msg + "%s: %w"
	if err := nils.Check(sl); err != nil {
		return fmt.Errorf(format, "", err)
	}

	if c.AbsLog == "" {
		if err := c.LogStore(); err != nil {
			return fmt.Errorf(format, "abslog was empty, attempted log store", err)
		}
	}

	path := string(c.AbsLog)
	st, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf(format, path, err)
	}
	if !st.IsDir() {
		return fmt.Errorf(format, st.Name(), ErrNotDir)
	}

	// test write permissions by creating and removing a temp file directly
	name := filepath.Join(path, ".defacto2_touch_test")
	if _, sErr := os.Stat(name); sErr == nil {
		if rErr := os.Remove(name); rErr != nil {
			c.Warn(sl, msg+" cannot remove test touch file", name, rErr)
			return nil
		}
	}

	f, err := os.Create(name)
	if err != nil {
		err = errors.Join(err, fmt.Errorf(format, name, ErrTouch))
		return err
	}
	if err := f.Close(); err != nil {
		c.Warn(sl, msg+" cannot close test touch file", name, err)
	}
	if err := os.Remove(name); err != nil {
		c.Warn(sl, msg+" cannot remove test touch file", name, err)
	}
	return nil
}
