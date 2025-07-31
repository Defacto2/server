package config

import "log/slog"

type Toggle bool // Toggle is a boolean value that returns a humanized string.

func (t Toggle) LogValue() slog.Value {
	if t {
		return slog.StringValue("TRUE")
	}
	return slog.StringValue("FALSE")
}

func (t Toggle) Bool() bool {
	return bool(t)
}
