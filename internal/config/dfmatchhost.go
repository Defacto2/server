package config

import "log/slog"

type Matchhost string

func (m Matchhost) LogValue() slog.Value {
	if m == "" {
		return slog.StringValue("")
	}
	return slog.StringValue(string(m))
}

func (m Matchhost) Help() string {
	if m == "" {
		return "No host address restrictions."
	}
	return ""
}

func (m Matchhost) String() string {
	return string(m)
}
