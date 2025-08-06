package config

import (
	"log/slog"
	"net/url"
)

type Connection string

func (c Connection) LogValue() slog.Value {
	rawURL := string(c)
	u, err := url.Parse(rawURL)
	if err != nil {
		return slog.StringValue(rawURL)
	}
	_, exists := u.User.Password()
	if !exists {
		return slog.StringValue(rawURL)
	}
	u.User = url.UserPassword(u.User.Username(), hide)
	return slog.StringValue(u.String())
}
