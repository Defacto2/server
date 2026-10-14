package config

import "log/slog"

type DemozooAuth string

func (s DemozooAuth) LogValue() slog.Value {
	if s == "" {
		return slog.StringValue("")
	}
	return slog.StringValue(mask)
}

func (s DemozooAuth) Help() string {
	if s == "" {
		return "No authorization key, one can be generated at: https://demozoo.org/account/api_keys"
	}
	return ""
}

func (s DemozooAuth) String() string {
	return string(s)
}
