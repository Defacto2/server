// Package config for system environment variable configurations for the server.
package config

// Config options for the Defacto2 server.
type Config struct {

	// HTTPPort is the port to be used by the HTTP server.
	HTTPPort int `env:"DEFACTO2_PORT" envDefault:"1323"`

	// IsProduction reduces the console feedback.
	IsProduction bool `env:"DEFACTO2_PRODUCTION"`

	// LogRequests uses the logger middleware to save HTTP requests to a file.
	LogRequests bool `env:"DEFACTO2_REQUESTS" envDefault:"false"`

	// NoRobots enables the X-Robots-Tag noindex and nofollow HTTP header for all server request.
	// This should never be enabled on production environments as search engines never crawl the website.
	NoRobots bool `env:"DEFACTO2_NOROBOTS" envDefault:"false"`

	// Timeout in seconds for the HTTP server.
	Timeout uint `env:"DEFACTO2_TIMEOUT" envDefault:"5"`
}
