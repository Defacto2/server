// Package config for system environment variable configurations for the server.
package config

import "time"

type Config struct {
	DataPort     int  `env:"PORT" envDefault:"1323"`
	IsProduction bool `env:"PRODUCTION"`
	LogRequests  bool `env:"REQUESTS" envDefault:"false"`
	NoRobots     bool `env:"NOROBOTS" envDefault:"false"` // TODO
}

const (
	Timeout = 5 * time.Second
)
