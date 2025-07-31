package config

import "log/slog"

type Sessionkey string

func (s Sessionkey) LogValue() slog.Value {
	if s == "" {
		return slog.StringValue("")
	}
	return slog.StringValue(hide)
}

func (s Sessionkey) Help() string {
	if s == "" {
		return "A random key will be generated during the server start."
	}
	return ""
}

func (s Sessionkey) String() string {
	return string(s)
}
