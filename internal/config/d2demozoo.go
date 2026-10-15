package config

import "log/slog"

const ZooAuth = "zooauth" // ZooAuth is the name of the echo context key to hold the authorization

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
	const i = 8
	if len(s) <= i {
		return "Authorization key is too short"
	}
	if s[8] != '.' {
		return "Authorization key expects the 9th character to be a period (.)"
	}
	return ""
}

func (s DemozooAuth) String() string {
	return string(s)
}
