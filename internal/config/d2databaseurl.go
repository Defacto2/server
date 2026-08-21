package config

import (
	"log/slog"
	"net/url"
)

type Connection string

func (c Connection) LogValue() slog.Value {
	rawURL := string(c)
	if rawURL == "" {
		return slog.StringValue("")
	}

	u, err := url.Parse(rawURL)
	if err != nil || u.User == nil {
		return slog.StringValue(rawURL)
	}

	_, ok := u.User.Password()
	if !ok {
		return slog.StringValue(rawURL)
	}

	u.User = url.UserPassword(u.User.Username(), mask)
	return slog.StringValue(u.String())
}
