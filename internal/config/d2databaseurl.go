package config

import (
	"log/slog"
	"net/url"
)

type Connection string

func (c Connection) LogValue() slog.Value {
	return slog.StringValue(c.String())
}

func (c Connection) String() string {
	rawURL := string(c)
	if rawURL == "" {
		return ""
	}

	u, err := url.Parse(rawURL)
	if err != nil || u.User == nil {
		return rawURL
	}

	_, ok := u.User.Password()
	if !ok {
		return rawURL
	}

	u.User = url.UserPassword(u.User.Username(), mask)
	return u.String()
}
