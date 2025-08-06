package config

import "log/slog"

type Absorphan Directory

func (a Absorphan) Help() string {
	if a == "" {
		return "Artifact backups are not possible"
	}
	return ""
}

func (a Absorphan) Issue() string {
	return Directory(a).Issue()
}

func (a Absorphan) LogValue() slog.Value {
	return Directory(a).LogValue()
}

func (a Absorphan) String() string {
	return Directory(a).String()
}
