package config

import (
	"crypto/sha512"
	"log/slog"
	"slices"
	"strings"
)

// Override the configuration settings fetched from the environment.
func (c *Config) Override() {
	// task 1, overwrite the stored google accounts
	c.GoogleAccounts = nil

	if rawIDs := c.GoogleIDs.String(); rawIDs != "" {
		for s := range slices.Values(strings.Split(rawIDs, ",")) {
			id := strings.TrimSpace(s)
			if id == "" {
				continue
			}
			sum := sha512.Sum384([]byte(id))
			c.GoogleAccounts = append(c.GoogleAccounts, sum)
		}
	}

	c.GoogleIDs = "zero out and overwrite placeholder"
	c.GoogleIDs = ""

	// task 2, setup the default http port whenever there is a misconfiguration
	if c.HTTPPort == 0 && c.TLSPort == 0 {
		c.HTTPPort = StdCustom
	}
}

// OAuth2s is a slice of Google OAuth2 accounts that are allowed to login.
// Each account is a 48 byte slice of bytes that represents the SHA-384 hash of the unique Google ID.
type OAuth2s [][48]byte

func (o OAuth2s) LogValue() slog.Value {
	return slog.Value{}
}

func (o OAuth2s) Values() [][48]byte {
	return o
}

func (o OAuth2s) String() string {
	switch len(o) {
	case 0:
		return ""
	case 1:
		return "single sign in account"
	default:
		return "multiple sign in accounts"
	}
}

func (o OAuth2s) Help() string {
	return Help(o)
}

// Help returns human readable help about the authorizations.
func Help(o OAuth2s) string {
	const none = "No accounts configured for the web administration"
	if o == nil {
		return none
	}

	switch len(o) {
	case 0:
		return none
	default: // intentionally keep the details vague
		return "Google account(s) in use for the web administration"
	}
}

type GoogleAuth string

func (g GoogleAuth) LogValue() slog.Value {
	if string(g) == "" {
		return slog.StringValue("Empty")
	}

	return slog.StringValue(mask)
}

func (g GoogleAuth) Help() string {
	if string(g) == "" {
		return "No accounts for web administration"
	}

	return ""
}

func (g GoogleAuth) String() string {
	return string(g)
}

type GoogleID string

func (g GoogleID) LogValue() slog.Value {
	if g == "" {
		return slog.StringValue("")
	}

	return slog.StringValue(mask)
}

func (g GoogleID) Help() string {
	const none = "No accounts configured for the web administration"
	if g == "" {
		return none
	}

	accounts := len(strings.Split(g.String(), ","))
	switch accounts {
	case 0:
		return none
	default:
		return "Google account(s) in use for sign in"
	}
}

func (g GoogleID) String() string {
	return string(g)
}
