package config

import (
	"fmt"
	"log/slog"
	"strings"
)

// UseTLS returns true if the server is configured to use TLS.
func (c Config) UseTLS() bool {
	return c.TLSPort > 0 && c.TLSCert != "" || c.TLSKey != ""
}

// UseTLSLocal returns true if the server is configured to use the local-mode.
func (c Config) UseTLSLocal() bool {
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
	return Directory(a).Issue()
}

func (a Abstlskey) LogValue() slog.Value {
	return Directory(a).LogValue()
}

func (a Abstlskey) String() string {
	return Directory(a).String()
}

type Abstlscrt File

func (a Abstlscrt) Help() string {
	if a == "" {
		return "No TLS certificate is in use"
	}
	return ""
}

func (a Abstlscrt) Issue() string {
	return Directory(a).Issue()
}

func (a Abstlscrt) LogValue() slog.Value {
	return Directory(a).LogValue()
}

func (a Abstlscrt) String() string {
	return Directory(a).String()
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
	s := "The web server is using " + strings.ToUpper(proto) +
		", example: " + strings.ToLower(proto) + "://localhost"
	if p != stdport {
		s = fmt.Sprintf("The web server is using %s, example: %s://localhost:%d",
			strings.ToUpper(proto), strings.ToLower(proto), p)
	}
	return s
}
