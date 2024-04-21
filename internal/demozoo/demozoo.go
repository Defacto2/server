// Package demozoo handles the retrieval of [production records] from the
// [Demozoo] API and the extraction of relevant data for the Defacto2 website.
//
// An example of a API v1 production call:
// As HTML, https://demozoo.org/api/v1/productions/185828/
// As JSONP, https://demozoo.org/api/v1/productions/185828/?format=jsonp
// As JSON,	https://demozoo.org/api/v1/productions/185828/?format=json
//
// [production records]: https://demozoo.org/api/v1/productions/
// [Demozoo]: https://demozoo.org
package demozoo

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
	"github.com/Defacto2/server/internal/tags"
)

const (
	ProdURL = "https://demozoo.org/api/v1/productions/" // ProdURL is the base URL for the Demozoo production API.
	Timeout = 10 * time.Second                          // HTTP client timeout, Demozoo replies can be slow.
	Sanity  = 450000                                    // Sanity is to check the maximum permitted production ID.
)

// Production is a Demozoo production record.
// Only the fields required for the Defacto2 website are included,
// with everything else being ignored.
type Production struct {
	// Title is the production title.
	Title string `json:"title"`
	// ReleaseDate is the production release date.
	ReleaseDate string `json:"release_date"`
	// Supertype is the production type.
	Supertype string `json:"supertype"`
	// Authors
	Authors []struct {
		Name     string `json:"name"`
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
// A status code is returned when the response status is not OK.
//
// [Demozoo API]: https://demozoo.org/api/v1/productions/
func (d *Production) Get(id int) (int, error) {
	if id < 1 {
		return 0, fmt.Errorf("%w: %d", ErrID, id)
	}
	client := http.Client{
		Timeout: Timeout,
	}
	url := ProdURL + strconv.Itoa(id)
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", helper.UserAgent)
	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return res.StatusCode, fmt.Errorf("%w: %s", ErrStatus, res.Status)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	err = json.Unmarshal(body, &d)
	if err != nil {
		return 0, err
	}
	if d.ID != id {
		return 0, fmt.Errorf("%w: %d", ErrSuccess, id)
	}
	return 0, nil
}

// GithubRepo returns the Github repository path of the production using
// the Production struct. It searches the external links for a link class that
// matches GithubRepo.
func (d Production) GithubRepo() string {
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
// the Production struct. It searches the external links for a
// link class that matches PouetProduction.
// A 0 is returned whenever the production does not have a recognized
// Pouet production link.
func (d Production) PouetProd() int {
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
// in the Production struct. It returns an error if the JSON data is
// invalid or the production ID is invalid.
func (d *Production) Unmarshal(r io.Reader) error {
	if err := json.NewDecoder(r).Decode(d); err != nil {
		return err
	}
	if d.ID < 1 {
		return fmt.Errorf("%w: %d", ErrID, d.ID)
	}
	return nil
}

// SuperType parses the Demozoo "production", "graphics" and "music" supertypes
// and returns the corresponding platform and section tags.
// It returns -1 for an unknown platform or section, in which case the
// caller should invalidate the Demozoo production.
func (d Production) SuperType() (tags.Tag, tags.Tag) {
	ok := func(s, p tags.Tag) bool {
		return s > -1 && p > -1
	}
	var platform tags.Tag = -1
	var section tags.Tag = -1

	platform, section = d.platforms(platform, section)
	if ok(section, platform) {
		return platform, section
	}

	platform, section = d.prodSuperType(platform, section)
	if ok(section, platform) {
		return platform, section
	}

	platform, section = d.graphicsSuperType(platform, section)
	if ok(section, platform) {
		return platform, section
	}

	platform, section = d.musicSuperType(platform, section)
	return platform, section
}

// platforms returns the platform and section tags for "platforms".
// A list of the types can be found at https://demozoo.org/api/v1/platforms/?ordering=id
func (d Production) platforms(platform, section tags.Tag) (tags.Tag, tags.Tag) {
	const (
		Windows = 1
		MsDos   = 4
		Linux   = 7
		MacOS   = 10
		Browser = 12
		// Javascript = 46 was removed from the api list of platforms.
		AdobeFlash = 47
		Java       = 48
		Macintosh  = 94
	)
	// Handle platforms.
	for _, p := range d.Platforms {
		switch p.ID {
		case Windows:
			platform = tags.Windows
		case MsDos:
			platform = tags.DOS
		case Linux:
			platform = tags.Linux
		case MacOS, Macintosh:
			platform = tags.Mac
		case Browser, AdobeFlash, Java:
			platform = tags.Markup
		}
		if platform > -1 {
			break
		}
	}
	return platform, section
}

// prodSuperType returns the platform and section tags for the "production" supertype.
// A list of the types can be found at https://demozoo.org/api/v1/production_types/?ordering=id
func (d Production) prodSuperType(platform, section tags.Tag) (tags.Tag, tags.Tag) {
	const (
		Demo           = 1
		Intro64K       = 2
		Intro4K        = 3
		Intro          = 4
		DiskMag        = 5
		Tool           = 6
		MusicDisk      = 7
		ProductionPack = 9
		Intro40K       = 10
		ChipMusicPack  = 12
		Cracktro       = 13
		Music          = 14
		Intro32b       = 15
		Intro64b       = 16
		Intro128b      = 18
		Intro256b      = 19
		Intro512b      = 20
		Intro1K        = 21
		Intro32K       = 22
		Game           = 33
		Intro16K       = 35
		Intro2K        = 37
		Intro100K      = 39
		BBStro         = 41
		Intro8K        = 43
		Magazine       = 47
		TextMag        = 49
		Intro96K       = 50
		Intro8b        = 54
		Intro16b       = 55
	)
	for _, p := range d.Platforms {
		switch p.ID {
		case Demo:
			section = tags.Demo
		case Intro64K, Intro4K, Intro, Intro40K, Intro32b,
			Intro64b, Intro128b, Intro256b, Intro512b, Intro1K,
			Intro32K, Intro16K, Intro2K, Intro100K, Intro8K,
			Intro96K, Intro8b, Intro16b:
			section = tags.Intro
		case DiskMag, Magazine, TextMag:
			section = tags.Mag
		case Tool:
			section = tags.Tool
		case MusicDisk, ChipMusicPack:
			section = tags.Pack
			platform = tags.Audio
		case ProductionPack:
			section = tags.Pack
		case Cracktro:
			section = tags.Intro
		case Music:
			platform = tags.Audio
			section = tags.Intro
		case Game:
			section = tags.Demo
		case BBStro:
			section = tags.BBS
		}
		if section > -1 {
			break
		}
	}
	return platform, section
}

// graphicsSuperType returns the platform and section tags for the "graphics" supertype.
func (d Production) graphicsSuperType(platform, section tags.Tag) (tags.Tag, tags.Tag) {
	const (
		Graphics   = 23
		Ascii      = 24
		PackAscii  = 25
		Ansi       = 26
		ExeGFX     = 27
		ExeGFX4K   = 28
		ArtPack    = 51
		ExeGFX256b = 56
		ExeGFX1K   = 58
	)
	for _, p := range d.Platforms {
		switch p.ID {
		case Graphics:
			platform = tags.Image
			section = tags.Logo
		case Ascii:
			platform = tags.Text
			section = tags.Logo
		case PackAscii:
			platform = tags.Text
			section = tags.Pack
		case Ansi:
			platform = tags.Text
			section = tags.Logo
		case ExeGFX, ExeGFX4K, ExeGFX256b, ExeGFX1K:
			section = tags.Logo
		case ArtPack:
			platform = tags.Image
			section = tags.Pack
		}
		if section > -1 {
			break
		}
	}
	return platform, section
}

// musicSuperType returns the platform and section tags for the "music" supertype.
func (d Production) musicSuperType(platform, section tags.Tag) (tags.Tag, tags.Tag) {
	const (
		ChipMusic   = 29
		ExeMusic    = 31
		ExeMusic32K = 32
		ExeMusic64K = 38
		MusicPack   = 52
	)
	for _, p := range d.Platforms {
		switch p.ID {
		case ChipMusic:
			platform = tags.Audio
			section = tags.Intro
		case ExeMusic, ExeMusic32K, ExeMusic64K:
			section = tags.Intro
		case MusicPack:
			platform = tags.Audio
			section = tags.Pack
		}
		if section > -1 {
			break
		}
	}
	return platform, section
}

// YouTubeVideo returns the ID of a video on YouTube. It searches the external links
// for a link class that matches YoutubeVideo.
// An empty string is returned whenever the production does not have a recognized
// YouTube video link.
func (d Production) YouTubeVideo() string {
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
		"quartex":                           1430,
		"razor-1911":                        519,
		"silicon-dream-artists":             25795,
		"scoopex":                           361,
		"superior-art-creations":            7050,
		"titan":                             2883,
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
