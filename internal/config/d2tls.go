package config

import (
	"fmt"
	"log/slog"
	"strings"
)

// UseTLS returns true if the server is configured to use TLS.
func (c Config) UseTLS() bool {
	return c.TLSPort > 0 && (c.TLSCert != "" || c.TLSKey != "")
}

// UseLocal returns true if the server is configured to use the local TLS-mode.
func (c Config) UseLocal() bool {
	return c.TLSPort > 0 && c.TLSCert == "" && c.TLSKey == ""
}

type Abstlskey File

func (a Abstlskey) Help() string {
	if a == "" {
		return "No TLS key is in use"
	}

	return ""
}

func (a Abstlskey) Issue() string {
	return File(a).Issue()
}

func (a Abstlskey) LogValue() slog.Value {
	return File(a).LogValue()
}

func (a Abstlskey) String() string {
	return File(a).String()
}

type Abstlscrt File

func (a Abstlscrt) Help() string {
	if a == "" {
		return "No TLS certificate is in use"
	}

	return ""
}

func (a Abstlscrt) Issue() string {
	return File(a).Issue()
}

func (a Abstlscrt) LogValue() slog.Value {
	return File(a).LogValue()
}

func (a Abstlscrt) String() string {
	return File(a).String()
}

type PortTLS Port

func (p PortTLS) LogValue() slog.Value {
	return Port(p).LogValue()
}

func (p PortTLS) Help() string {
	return protoPort(Port(p), StdHTTPS, "https")
}

func (p PortTLS) Value() uint16 {
	return Port(p).Value()
}

func (p PortTLS) Check() error {
	return Port(p).Check()
}

func protoPort(p, stdport Port, proto string) string {
	if p == 0 {
		return "The web server is not using " + strings.ToUpper(proto)
	}

	if p != stdport {
		return fmt.Sprintf("The web server is using %s, example: %s://localhost:%d",
			strings.ToUpper(proto), strings.ToLower(proto), p)
	}

	return "The web server is using " + strings.ToUpper(proto) +
		", example: " + strings.ToLower(proto) + "://localhost"
}
