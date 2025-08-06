package config

import "log/slog"

type Absextra Directory

func (a Absextra) Help() string {
	if a == "" {
		return "No textfiles will be shown"
	}
	return ""
}

func (a Absextra) Issue() string {
	return Directory(a).Issue()
}

func (a Absextra) LogValue() slog.Value {
	return Directory(a).LogValue()
}

func (a Absextra) String() string {
	return Directory(a).String()
}
