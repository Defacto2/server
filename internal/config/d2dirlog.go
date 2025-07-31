package config

import "log/slog"

type Abslog Directory

func (a Abslog) Help() string {
	if a == "" {
		return "No logs will be saved"
	}
	return ""
}

func (a Abslog) Issue() string {
	return Directory(a).Issue()
}

func (a Abslog) LogValue() slog.Value {
	return Directory(a).LogValue()
}

func (a Abslog) String() string {
	return Directory(a).String()
}
