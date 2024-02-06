// Package zoo handles the retrieval of [production records] from the
// [Demozoo] API and the extraction of relevant data for the Defacto2 website.
//
// [production records]: https://demozoo.org/api/v1/productions/
// [Demozoo]: https://demozoo.org
package zoo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Defacto2/server/internal/helper"
)

const (
	// https://demozoo.org/api/v1/productions/183715/?format=json
	ProdURL = "https://demozoo.org/api/v1/productions/"
	Timeout = 5 * time.Second
)

// Demozoo is a production record from the Demozoo API.
// Only the fields required for the Defacto2 website are included,
// with everything else being ignored.
type Demozoo struct {
	// Title is the production title.
	Title string `json:"title"`
	// ReleaseDate is the production release date.
	ReleaseDate string `json:"release_date"`
	// Supertype is the production type.
	Supertype string `json:"supertype"`
	// Authors
	Authors []struct {
		Releaser struct {
			Name    string `json:"name"`
			IsGroup bool   `json:"is_group"`
		} `json:"releaser"`
	} `json:"author_nicks"`
	// Platforms is the production platform.
	Platforms []struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
	} `json:"platforms"`
	// Types is the production type.
	Types []struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
	} `json:"types"`
	// Download links to the remotely hosted files.
	DownloadLinks []struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	} `json:"download_links"`
	// ExternalLinks links to the remotely hosted files.
	ExternalLinks []struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	} `json:"external_links"`
	// ID is the production ID.
	ID int `json:"id"`
}

var (
	ErrID      = errors.New("demozoo production id is invalid")
	ErrSuccess = errors.New("demozoo production not found")
	ErrStatus  = errors.New("demozoo production status is not ok")
)

// Get requests data for a production record from the [Demozoo API].
// It returns an error if the production ID is invalid, when the request
// reaches a [Timeout] or fails.
//
// [Demozoo API]: https://demozoo.org/api/v1/productions/
func (d *Demozoo) Get(id int) error {
	if id < 1 {
		return fmt.Errorf("%w: %d", ErrID, id)
	}
	client := http.Client{
		Timeout: Timeout,
	}
	url := ProdURL + strconv.Itoa(id)
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", helper.UserAgent)
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %d - %s", ErrStatus, res.StatusCode, res.Status)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &d)
	if err != nil {
		return err
	}
	if d.ID != id {
		return fmt.Errorf("%w: %d", ErrSuccess, id)
	}
	return nil
}

// GithubRepo returns the Github repository path of the production using
// the Demozoo struct. It searches the external links for a link class that
// matches GithubRepo.
func (d Demozoo) GithubRepo() string {
	for _, link := range d.ExternalLinks {
		if link.LinkClass != "GithubRepo" {
			continue
		}
		url, err := url.Parse(link.URL)
		if err != nil {
			continue
		}
		if url.Host != "github.com" {
			continue
		}
		return url.Path
	}
	return ""
}

// PouetProd returns the Pouet ID of the production using
// the Demozoo struct. It searches the external links for a
// link class that matches PouetProduction.
// A 0 is returned whenever the production does not have a recognized
// Pouet production link.
func (d Demozoo) PouetProd() int {
	for _, link := range d.ExternalLinks {
		if link.LinkClass != "PouetProduction" {
			continue
		}
		url, err := url.Parse(link.URL)
		if err != nil {
			continue
		}
		id, err := strconv.Atoi(url.Query().Get("which"))
		if err != nil {
			continue
		}
		return id
	}
	return 0
}

// Unmarshal parses the JSON-encoded data and stores the result
// in the Demozoo production struct. It returns an error if the JSON data is
// invalid or the production ID is invalid.
func (d *Demozoo) Unmarshal(data []byte) error {
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}
	if d.ID < 1 {
		return fmt.Errorf("%w: %d", ErrID, d.ID)
	}
	return nil
}

// YouTubeVideo returns the ID of a video on YouTube. It searches the external links
// for a link class that matches YoutubeVideo.
// An empty string is returned whenever the production does not have a recognized
// YouTube video link.
func (d Demozoo) YouTubeVideo() string {
	for _, link := range d.ExternalLinks {
		if link.LinkClass != "YoutubeVideo" {
			continue
		}
		url, err := url.Parse(link.URL)
		if err != nil {
			continue
		}
		if url.Host != "youtube.com" && url.Host != "www.youtube.com" {
			continue
		}
		if url.Path != "/watch" {
			continue
		}
		return url.Query().Get("v")
	}
	return ""
}

// URI is a the URL slug of the releaser.
type URI string

// GroupID is the Demozoo ID of the group.
type GroupID uint

// Groups is a map of releasers URIs mapped to their Demozoo IDs.
type Groups map[URI]GroupID

// groups returns a map of releasers URIs mapped to their Demozoo IDs.
func groups() Groups {
	return Groups{
		"acid-productions":                  7647,
		"class":                             16508,
		"defacto2":                          10000,
		"fairlight":                         239,
		"international-network-of-crackers": 12175,
		"insane-creators-enterprise":        2169,
		"mirage":                            45887,
		"paradigm":                          26612,
		"razor-1911":                        519,
		"silicon-dream-artists":             25795,
		"superior-art-creations":            7050,
		"the-dream-team":                    20609,
		"the-humble-guys":                   7421,
		"the-silents":                       101,
		"tristar-ampersand-red-sector-inc":  69,
	}
}

// Find returns the Demozoo group ID for the given uri.
// It returns 0 if the uri is not known.
func Find(uri string) GroupID {
	if _, ok := groups()[URI(uri)]; ok {
		return groups()[URI(uri)]
	}
	return 0
}
