// Package remote provides the remote download and update of artifact data from third-party sources such as API's.
//
//nolint:exhaustruct_v5
package remote

import (
	"errors"
	"log/slog"
	"net/url"
	"slices"
	"strings"

	"github.com/Defacto2/server/internal/logs"
)

var (
	ErrBodyNil  = errors.New("remote: body is nil")
	ErrExist    = errors.New("remote: file already exists")
	ErrNoRecord = errors.New("remote: could not the get record from demozoo api")
	ErrStatus   = errors.New("remote: wrong status code")
	ErrTimeout  = errors.New("remote: the time duration is out of range")
)

// ReplaceURL returns a valid URL if the provided rawURL is a known broken link to a scene.org file.
// Otherwise the rawURL is returned.
//
// For example, the following rawURL:
//
//	`http://files.scene.org/view/demos/groups/trsi/ms-dos/trsiscxt.zip`
//
// will return:
//
//	`https://files.scene.org/get/demos/groups/trsi/ms-dos/trsiscxt.zip`
func ReplaceURL(sl *slog.Logger, rawURL string) string {
	if sl == nil {
		sl = logs.Discard()
	}

	logRef := func(url string) {
		sl.Info("FixURL refactored the url", slog.String("old", rawURL), slog.String("new", url))
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		sl.Info("FixURL url would not be parsed", slog.String("url", rawURL), slog.Any("error", err))
		return rawURL
	}

	if u.Host == "scene.org" && u.Path == "/file.php" {
		return rawURL
	}

	p := u.Path
	if u.Host == "files.scene.org" {
		elem := strings.Split(p, "/")
		if len(elem) > 0 && elem[1] == "view" {
			url := sceneorg(elem).String()
			logRef(url)
			return url
		}
	}

	if u.Host == "ftp.scene.org" || u.Host == "ftp.pl.scene.org" || u.Host == "ftp.no.scene.org" {
		elem := strings.Split(p, "/")
		if len(elem) > 0 && elem[1] == "pub" {
			url := sceneorg(elem).String()
			logRef(url)
			return url
		}
	}

	if u.Host == "sceneorg.retropc.se" || u.Host == "mirror.netcologne.de" {
		elem := strings.Split(p, "/")
		if len(elem) > 0 && elem[1] == "scene.org" {
			elem = slices.Insert(elem, 1, "get")
			url := sceneorg(elem).String()
			logRef(url)
			return url
		}
	}

	if u.Host == "discmaster.textfiles.com" {
		elem := strings.Split(p, "/")
		if len(elem) > 0 && elem[1] == "view" {
			elem[1] = "file"
			url := refactor(u.Host, elem).String()
			logRef(url)
			return url
		}
	}

	return rawURL
}

func sceneorg(x []string) *url.URL {
	const minimum = 2
	if len(x) < minimum {
		return &url.URL{}
	}

	elem := append([]string{x[0], "get"}, x[2:]...)
	return &url.URL{
		Scheme: "https",
		Host:   "files.scene.org",
		Path:   strings.Join(elem, "/"),
	}
}

func refactor(host string, elem []string) *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   host,
		Path:   strings.Join(elem, "/"),
	}
}
