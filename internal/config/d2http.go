package config

import (
	"fmt"
	"log/slog"
)

var (
	ErrPortMax = fmt.Errorf("http port value must be between 1-%d", PortMax)
	ErrPortSys = fmt.Errorf("http port values between 1-%d require system access", PortSys)
)

// UseHTTP returns true if the server is configured to use HTTP.
func (c Config) UseHTTP() bool {
	return c.HTTPPort > 0
}

type Port uint // Port is a network port number.

func (p Port) LogValue() slog.Value {
	return slog.IntValue(int(p))
}

func (p Port) Value() uint {
	return uint(p)
}

func (p Port) Check() error {
	return Validate(uint(p))
}

type PortHttp Port

func (p PortHttp) LogValue() slog.Value {
	return Port(p).LogValue()
}

func (p PortHttp) Help() string {
	return protoPort(Port(p), StdHTTP, "http")
}

func (p PortHttp) Value() uint {
	return Port(p).Value()
}

func (p PortHttp) Check() error {
	return Port(p).Check()
}

// Validate returns an error if the HTTP or TLS port is invalid.
func Validate(port uint) error {
	const disabled = 0
	if port == disabled {
		return nil
	}
	if port > PortMax {
		return ErrPortMax
	}
	if port <= PortSys {
		return ErrPortSys
	}
	return nil
}
