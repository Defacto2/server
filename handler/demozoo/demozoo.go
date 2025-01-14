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
	"strings"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/tags"
)

var (
	ErrID      = errors.New("id is invalid")
	ErrSuccess = errors.New("not found")
	ErrStatus  = errors.New("status is not ok")
)

func client() http.Client {
	const ten = 10
	return http.Client{
		Timeout: ten * time.Second,
	}
}

const (
	ProdURL = "https://demozoo.org/api/v1/productions/" // ProdURL is the base URL for the Demozoo production API.
	Sanity  = 450000                                    // Sanity is to check the maximum permitted production ID.
	firstID = 1                                         // firstID is the first production ID on Pouet.
)

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
	Credits []struct {
		Nick struct {
			Name         string `json:"name"`
			Abbreviation string `json:"abbreviation"`
			Releaser     struct {
				URL     string `json:"url"`
				ID      int    `json:"id"`
				Name    string `json:"name"`
				IsGroup bool   `json:"is_group"`
			} `json:"releaser"`
		} `json:"nick"`
		Category string `json:"category"`
		Role     string `json:"role"`
	} `json:"credits"`
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

// Get requests data for a production record from the [Demozoo API].
// It returns an error if the production ID is invalid, when the request
// reaches a [Timeout] or fails.
// A status code is returned when the response status is not OK.
//
// [Demozoo API]: https://demozoo.org/api/v1/productions/
func (p *Production) Get(id int) (int, error) {
	if id < firstID {
		return 0, fmt.Errorf("get demozoo production %w: %d", ErrID, id)
	}
	url := ProdURL + strconv.Itoa(id)
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("get demozoo production new request %w", err)
	}
	req.Header.Set("User-Agent", helper.UserAgent)
	c := client()
	res, err := c.Do(req)
	if err != nil {
		return 0, fmt.Errorf("get demozoo production client do %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, res.Body)
		res.Body.Close()
		return res.StatusCode, fmt.Errorf("get demozoo production %w: %s", ErrStatus, res.Status)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		_, _ = io.Copy(io.Discard, res.Body)
		res.Body.Close()
		return 0, fmt.Errorf("get demozoo production read all %w", err)
	}
	err = json.Unmarshal(body, &p)
	clear(body)
	if err != nil {
		return 0, fmt.Errorf("get demozoo production json unmarshal %w", err)
	}
	if p.ID != id {
		return 0, fmt.Errorf("get demozoo production %w: %d", ErrSuccess, id)
	}
	return 0, nil
}

// GithubRepo returns the Github repository path of the production using
// the Production struct. It searches the external links for a link class that
// matches GithubRepo.
func (p *Production) GithubRepo() string {
	for _, link := range p.ExternalLinks {
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
func (p *Production) PouetProd() int {
	for _, link := range p.ExternalLinks {
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
func (p *Production) Unmarshal(r io.Reader) error {
	if r == nil {
		return nil
	}
	if err := json.NewDecoder(r).Decode(p); err != nil {
		return fmt.Errorf("demozoo production json decode: %w", err)
	}
	if p.ID < firstID {
		return fmt.Errorf("demozoo production %w: %d", ErrID, p.ID)
	}
	return nil
}

// SuperType parses the Demozoo "production", "graphics" and "music" supertypes
// and returns the corresponding platform and section tags.
// It returns -1 for an unknown platform or section, in which case the
// caller should invalidate the Demozoo production.
func (p *Production) SuperType() (tags.Tag, tags.Tag) {
	superType := func(pl, se tags.Tag) bool {
		return pl > -1 && se > -1
	}
	var platform tags.Tag = -1
	var section tags.Tag = -1

	platform, section = p.platforms(platform, section)
	if superType(platform, section) {
		return platform, section
	}

	platform, section = p.prodSuperType(platform, section)
	if superType(platform, section) {
		return platform, section
	}

	platform, section = p.graphicsSuperType(platform, section)
	if superType(platform, section) {
		return platform, section
	}

	platform, section = p.musicSuperType(platform, section)
	return platform, section
}

// platforms returns the platform and section tags for "platforms".
// A list of the types can be found at https://demozoo.org/api/v1/platforms/?ordering=id
func (p *Production) platforms(platform, section tags.Tag) (tags.Tag, tags.Tag) {
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
	for _, item := range p.Platforms {
		switch item.ID {
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
func (p *Production) prodSuperType(platform, section tags.Tag) (tags.Tag, tags.Tag) {
	for _, item := range p.Platforms {
		switch item.ID {
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
func (p *Production) graphicsSuperType(platform, section tags.Tag) (tags.Tag, tags.Tag) {
	const (
		Graphics   = 23
		ASCII      = 24
		PackASCII  = 25
		Ansi       = 26
		ExeGFX     = 27
		ExeGFX4K   = 28
		ArtPack    = 51
		ExeGFX256b = 56
		ExeGFX1K   = 58
	)
	for _, item := range p.Platforms {
		switch item.ID {
		case Graphics:
			platform = tags.Image
			section = tags.Logo
		case ASCII:
			platform = tags.Text
			section = tags.Logo
		case PackASCII:
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
func (p *Production) musicSuperType(platform, section tags.Tag) (tags.Tag, tags.Tag) {
	const (
		ChipMusic   = 29
		ExeMusic    = 31
		ExeMusic32K = 32
		ExeMusic64K = 38
		MusicPack   = 52
	)
	for _, item := range p.Platforms {
		switch item.ID {
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
func (p *Production) YouTubeVideo() string {
	for _, link := range p.ExternalLinks {
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

// Released returns the production's release date as date_issued_year, month, day values.
func (p *Production) Released() (int16, int16, int16) {
	return helper.Released(p.ReleaseDate)
}

// Groups returns the first two names in the production that have is_group as true.
// The one exception is if the production title contains a reference to a BBS or FTP site name.
// Then that title will be used as the first group returned.
func (p *Production) Groups() (string, string) {
	// find any reference to BBS or FTP in the production title to
	// obtain a possible site name.
	var a, b string
	if s := Site(p.Title); s != "" {
		a = s
	}
	// range through author nicks for any group matches
	for _, nick := range p.Authors {
		if !nick.Releaser.IsGroup {
			continue
		}
		if a == "" {
			a = nick.Releaser.Name
			continue
		}
		if b == "" {
			b = nick.Releaser.Name
			break
		}
	}
	return a, b
}

// Site parses a production title to see if it is suitable as a BBS or FTP site name,
// otherwise an empty string is returned.
func Site(title string) string {
	s := strings.Split(title, " ")
	if s[0] == "The" {
		s = s[1:]
	}
	for i, n := range s {
		if n == "BBS" {
			return strings.Join(s[0:i], " ") + " BBS"
		}
		if n == "FTP" {
			return strings.Join(s[0:i], " ") + " FTP"
		}
	}
	return ""
}

// Authors parses Demozoo authors and reclassifies them into Defacto2 people rolls.
func (p *Production) Releasers() ([]string, []string, []string, []string) {
	tx, co, gx, mu := []string{}, []string{}, []string{}, []string{}
	for _, c := range p.Credits {
		if c.Nick.Releaser.IsGroup {
			continue
		}
		switch category(c.Category) {
		case TextC:
			tx = append(tx, c.Nick.Name)
		case CodeC:
			co = append(co, c.Nick.Name)
		case GraphicsC:
			gx = append(gx, c.Nick.Name)
		case MusicC:
			mu = append(mu, c.Nick.Name)
		case MagazineC:
			// do nothing.
		}
	}
	return tx, co, gx, mu
}

// Category are tags for production imports.
type Category int

const (
	TextC     Category = iota // Text based files.
	CodeC                     // Code are binary files.
	GraphicsC                 // Graphics are images.
	MusicC                    // Music is audio.
	MagazineC                 // Magazine are publications.
)

func (c Category) String() string {
	return [...]string{"text", "code", "graphics", "music", "magazine"}[c]
}

func category(s string) Category {
	switch strings.ToLower(s) {
	case TextC.String():
		return TextC
	case CodeC.String():
		return CodeC
	case GraphicsC.String():
		return GraphicsC
	case MusicC.String():
		return MusicC
	case MagazineC.String():
		return MagazineC
	}
	return -1
}

// URI is a the URL slug of the releaser.
type URI string

// GroupID is the Demozoo ID of the group.
type GroupID int

// Groups is a map of releasers URIs mapped to their Demozoo IDs.
type Groups map[URI]GroupID

// Find returns the Demozoo group ID for the given uri.
// It returns 0 if the uri is not known.
func Find(uri string) GroupID {
	if group, exist := groups()[URI(uri)]; exist {
		return group
	}
	return 0
}

const (
	cpc   = "corporation-for-public-cybercasting-2001"
	nappa = "north-american-pirate_phreak-association"
)

// groups returns a map of releasers URIs mapped to their Demozoo IDs.
func groups() Groups { //nolint:funlen
	return Groups{
		"2000ad":                                20,
		"aces-of-ansi-art":                      14208,
		"acid-productions":                      7647,
		"advanced-pirate-technology":            46652,
		"alpha-flight":                          1492,
		"adrenalin":                             46669,
		"anthrox":                               1218,
		"bentley-sidwell-productions":           46300,
		"bitchin-ansi-design":                   81373,
		"boys-from-company-c":                   47088,
		"canadian-pirates-inc":                  69325,
		"cascada":                               7926,
		"class":                                 16508,
		"c-ampersand-m":                         146439,
		cpc:                                     146445,
		"codex":                                 114419,
		"club-elan":                             82987,
		"crackers-in-action":                    59013,
		"creators-of-intense-art":               17338,
		"cybrix":                                70314,
		"damn-excellent-ansi-design":            25642,
		"defacto2":                              10000,
		"dead-memory":                           76576,
		"digital-noise-alliance":                75943,
		"dread":                                 76438,
		"drink-or-die":                          46616,
		"dynamix":                               68008,
		"dytec":                                 6698,
		"eclipse":                               67881,
		"esp-pirates":                           55436,
		"electro-magnetic-crackers":             76266,
		"electromotive-force":                   7702,
		"electronic-rats":                       17164,
		"extinct":                               131861,
		"fairlight":                             239,
		"friendship":                            76473,
		"five-o":                                123441,
		"future-brain-inc":                      59015,
		"future-crew":                           357,
		"genesis":                               37525,
		"graphic-revolution-in-progress":        23211,
		"graphics-rendered-in-magnificence":     25682,
		"kosmic-loader-foundation":              30739,
		"hype":                                  111124,
		"illuminatus":                           120174,
		"international-network-of-crackers":     12175,
		"insane-creators-enterprise":            2169,
		"insanity":                              130208,
		"katharsis":                             37053,
		"kgb":                                   69323,
		"knights-of-the-round-table":            47158,
		"lancelot":                              131757,
		"lancelot-2":                            131699,
		"legacy":                                86436,
		"legend":                                2075,
		"licensed-to-draw":                      25816,
		"lkcc":                                  904,
		"mai-review":                            145041,
		"masters-of-abstractions-and-illusions": 145041,
		"malicious-art-denomination":            86862,
		"malice":                                46350,
		"majic-12":                              870,
		"mea-culpa":                             76417,
		"mercury":                               113031,
		"miami-cracking-machine":                45877,
		"mirage":                                45887,
		"new-york-crackers":                     53704,
		nappa:                                   122784,
		"norwegian-cracking-company":            82964,
		"outlaws":                               2335,
		"originally-funny-guys":                 76260,
		"paradox":                               1853,
		"pyradical":                             83958,
		"paradigm":                              26612,
		"pentagram":                             46472,
		"pirates-sick-of-initials":              59019,
		"pirates-with-attitudes":                46360,
		"propaganda":                            145592,
		"prevues":                               130455,
		"ptl-club":                              53053,
		"quartex":                               1430,
		"razor-1911":                            519,
		"rebels":                                628,
		"relentless-pursuit-of-magnificence":    45917,
		"rise-in-superior-couriering":           45969,
		"skillion":                              46362,
		"skid-row":                              14943,
		"silicon-dream-artists":                 25795,
		"sma-posse":                             58173,
		"scoopex":                               361,
		"software-pirates-inc":                  123017,
		"sorcerers":                             37044,
		"superior-art-creations":                7050,
		"surprise-productions":                  1536,
		"sprint":                                112416,
		"technobrains":                          75071,
		"titan":                                 2883,
		"the-brain-slayer":                      59156,
		"the-dream-team":                        20609,
		"the-duplicators":                       146432,
		"the-humble-guys":                       7421,
		"the-firm":                              45892,
		"the-grand-council":                     84582,
		"the-north-west-connection":             131124,
		"the-phoney-coders":                     6627,
		"the-silents":                           101,
		"the-space-pigs":                        55023,
		"the-sysops-association-network":        76382,
		"the-underground-council":               68127,
		"the-untouchables":                      76042,
		"thg-fx":                                46356,
		"toads":                                 146433,
		"tristar-ampersand-red-sector-inc":      69,
		"triad":                                 131111,
		"untouchables":                          112780,
		"ultra-force":                           37076,
		"ultra-tech":                            75375,
		"union":                                 58739,
		"united-artist-association":             118271,
		"united-software-association":           45881,
		"united-software-association*fairlight": 45881,
		"velocity-couriers":                     83317,
		"visions-of-reality":                    86454,
		"vortex-software":                       146440,
		"xerox":                                 59161,
		"chicago-bbs":                           12584,
		"public-domain":                         146450,
		"pirate":                                146562,
		"psycho-corporate-productions":          127714,
		"copycats-inc":                          146659,
		"pirates-r-us":                          46502,
		"gpa":                                   146655,
		"pirates-club-inc":                      146463,
		"ipl":                                   146461,
		"scb":                                   146545,
		"imperial-warlords":                     122965,
		"the-illinois-pirates":                  146676,
		"silicon-valley-swappe-shoppe":          146458,
		"occult-network":                        146546,
		"myth-inc":                              12549,
		"crime-syndicate-net":                   131344,
		"eagle-soft-incorporated":               1540,
		"the-elementals-piratelist":             146549,
		"the-nameless-ones-1989":                146711,
		"the-stealth-pirate-network":            146712,
		"extasy":                                79316,
		"pirates-cove":                          146751,
		"classic-vocs":                          128541,
		"defjam":                                2909,
		"disassemblers-of-america":              363349,
		"mickey-mouse-club":                     146767,
		"interceptor":                           78186,
		"los-angeles-sysops-alliance":           128540,
		"national-elite-underground-alliance":   47090,
		"national-pirate-list":                  146548,
		"north-american-society-of-anarchists":  79217,
		"pc_cracking-service":                   69326,
		"phoenix":                               146791,
		"powr":                                  131690,
		"psycho":                                146810,
		"public-enemy":                          53703,
		"american-pirate-industries":            112171,
		"west-coast-cracking-production":        112222,
		"warez":                                 146839,
		"unit-173":                              111613,
		"tired-of-protection":                   112798,
		"the-warez-alliance":                    46910,
		"the-syndicate":                         112279,
		"the-canadian-crackers":                 112645,
		"the-alternative":                       46537,
		"spectrum":                              124103,
		"software-chronicles-digest":            46476,
		"scd_dox":                               46567,
		"rescue-raider":                         131347,
		"idiots-creations-unlimited":            147108,
		"more-stupid-initials":                  147104,
		"quantum":                               76350,
		"sniper":                                147175,
		"delirium-of-disorder":                  66642,
		"needful-things":                        147177,
		"police":                                76598,
		"netrunners":                            83424,
	}
}
