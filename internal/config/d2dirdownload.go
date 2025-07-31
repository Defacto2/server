package config

import "log/slog"

type Absdown Directory

func (a Absdown) Help() string {
	if a == "" {
		return "No downloads will be served"
	}
	return ""
}

func (a Absdown) Issue() string {
	return Directory(a).Issue()
}

func (a Absdown) LogValue() slog.Value {
	return Directory(a).LogValue()
}

func (a Absdown) String() string {
	return Directory(a).String()
}
