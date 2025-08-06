package config

import "log/slog"

const AbsThumbnail = "AbsThumbnail" // AbsThumbnail means the absolute thumbnail assets directory.

type Absthumb Directory

func (a Absthumb) Help() string {
	if a == "" {
		return "No thumbnails will be shown"
	}
	return ""
}

func (a Absthumb) Issue() string {
	return Directory(a).Issue()
}

func (a Absthumb) LogValue() slog.Value {
	return Directory(a).LogValue()
}

func (a Absthumb) String() string {
	return Directory(a).String()
}
