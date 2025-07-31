package config

import "log/slog"

const AbsPreview = "AbsPreview" // AbsPreview means the absolute preview assets directory.

type Absprev Directory

func (a Absprev) Help() string {
	if a == "" {
		return "No preview images will be shown"
	}
	return ""
}

func (a Absprev) Issue() string {
	return Directory(a).Issue()
}

func (a Absprev) LogValue() slog.Value {
	return Directory(a).LogValue()
}

func (a Absprev) String() string {
	return Directory(a).String()
}
