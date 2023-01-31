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

	// ConfigDir will overwrite the default directory that will store the server logs.
	// It is recommended this is left blank to use the home .config path.
	ConfigDir string `env:"DEFACTO2_CONFIG"`

	// DownloadDir provides the directory that holds UUID named files to offer as release downloads.
	DownloadDir string `env:"DEFACTO2_DOWNLOAD"`

	// MaxProcs overrides and limits the number of operating system threads this application can use.
	MaxProcs uint `env:"MAXPROCS" envDefault:"0"`
}
