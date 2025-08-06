package config

import (
	"crypto/sha512"
	"log/slog"
	"slices"
	"strings"
)

// Override the configuration settings fetched from the environment.
func (c *Config) Override() {
	// hash and delete any supplied google ids
	ids := strings.Split(c.GoogleIDs.String(), ",")
	for id := range slices.Values(ids) {
		sum := sha512.Sum384([]byte(id))
		c.GoogleAccounts = append(c.GoogleAccounts, sum)
	}
	c.GoogleIDs = "overwrite placeholder"
	c.GoogleIDs = "" // empty the string

	// set the default HTTP port if both ports are configured to zero
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
	cnt := len(o)
	switch cnt {
	case 0:
		return ""
	case 1:
		return "one sign-in account"
	default:
		return "multiple sign-in accounts"
	}
}

func (o OAuth2s) Help() string {
	return Googles(o)
}

// Googles returns human readable help about the ids.
func Googles(ids [][48]byte) string {
	const none = "No accounts configured for the web administration"
	if ids == nil {
		return none
	}
	cnt := len(ids)
	switch cnt {
	case 0:
		return none
	default:
		return "Google account(s) in use for the web administration"
	}
}

type Googleauth string

func (g Googleauth) LogValue() slog.Value {
	if string(g) == "" {
		return slog.StringValue("Empty")
	}
	return slog.StringValue(hide)
}

func (g Googleauth) Help() string {
	if string(g) == "" {
		return "No accounts for web administration"
	}
	return ""
}

func (g Googleauth) String() string {
	return string(g)
}

type Googleids string

func (g Googleids) LogValue() slog.Value {
	if g == "" {
		return slog.StringValue("")
	}
	return slog.StringValue(hide)
}

func (g Googleids) Help() string {
	const none = "No accounts configured for the web administration"
	if g == "" {
		return none
	}
	cnt := len(strings.Split(g.String(), ","))
	switch cnt {
	case 0:
		return none
	default:
		return "Google account(s) in use for sign-ins"
	}
}

func (g Googleids) String() string {
	return string(g)
}
