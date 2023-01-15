// Package config for system environment variable configurations for the server.
package config

import "time"

// Config options for the Defacto2 server.
type Config struct {
	// HTTPPort is the port to be used by the HTTP server.
	HTTPPort int `env:"PORT" envDefault:"1323"`
	// IsProduction reduces the console feedback.
	IsProduction bool `env:"PRODUCTION"`
	// LogRequests uses the logger middleware to save HTTP requests to a file.
	LogRequests bool `env:"REQUESTS" envDefault:"false"`
	// NoRobots enables the X-Robots-Tag noindex and nofollow HTTP header for all server request.
	// This should never be enabled on production environments as search engines never crawl the website.
	NoRobots bool `env:"NOROBOTS" envDefault:"false"`
}

const (
	Timeout = 5 * time.Second
)
