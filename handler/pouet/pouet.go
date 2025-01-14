// Package pouet provides production, user voting data sourced from the Pouet website API.
package pouet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
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

var client = http.Client{
	Timeout: 10 * time.Second,
}

const (
	ProdURL     = "https://api.pouet.net/v1/prod/?id=" // ProdURL is the base URL for the Pouet production API.
	StarRounder = 0.5                                  // StarRounder is the rounding value for the stars rating.
	Sanity      = 200000                               // Sanity is to check the maximum permitted production ID.
	firstID     = 1                                    // firstID is the first production ID on Pouet.
)

// Production is the production data from the Pouet API.
// The Pouet API returns values as null or string, so this struct
// is used to normalize the data types.
type Production struct {
	ID          int    `json:"id"`           // ID is the prod ID.
	Title       string `json:"title"`        // Title is the prod title.
	ReleaseDate string `json:"release_date"` // ReleaseDate is the prod release date.
	Download    string `json:"download"`     // Download is the first download link.
	Demozoo     string `json:"demozoo"`      // Demozoo is the Demozoo identifier.
	Groups      []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"groups"` // Groups are the releasers that produced the prod.
	Platforms Platforms `json:"platforms"` // Platforms are the platforms the prod runs on.
	Platform  string    `json:"platform"`  // Platform is the prod platforms as a string.
	Types     Types     `json:"types"`     // Types are the prod types.
	Links     []struct {
		Type string `json:"type"`
		Link string `json:"link"`
	} `json:"downloads"` // Downloads are the additional download links.
	Valid bool `json:"valid"` // Valid is true if this prod is a supported type and platform.
}

// Get requests data for a production record from the [Pouet API].
// It returns an error if the production ID is invalid, when the request
// reaches a [Timeout] or fails.
// A status code is returned when the response status is not OK.
//
// [Pouet API]: https://api.pouet.net/v1/prod/?id=
func (p *Production) Get(id int) (int, error) {
	if id < firstID {
		return 0, fmt.Errorf("get pouet production %w: %d", ErrID, id)
	}
	resp := Response{}
	if code, err := resp.Get(id); err != nil {
		return code, fmt.Errorf("pouet uploader get %w", err)
	}
	id, err := strconv.Atoi(resp.Prod.ID)
	if err != nil {
		return 0, fmt.Errorf("pouet uploader atoi %w", err)
	}
	platOkay := PlatformsValid(resp.Prod.Platforms.String())
	typeOkay := TypesValid(resp.Prod.Types.String())
	p.ID = id
	p.Title = resp.Prod.Title
	p.ReleaseDate = resp.Prod.ReleaseDate
	p.Download = resp.Prod.Download
	p.Demozoo = resp.Prod.Demozoo
	p.Groups = resp.Prod.Groups
	p.Platforms = resp.Prod.Platforms
	p.Types = resp.Prod.Types
	p.Links = resp.Prod.DownloadLinks
	p.Valid = platOkay && typeOkay
	return 0, nil
}

func PlatformsValid(s string) bool {
	platforms := strings.Split(strings.ToLower(s), ",")
	for _, platform := range platforms {
		switch strings.TrimSpace(platform) {
		case "msdosgus", "msdos", "windows":
			return true
		}
	}
	return false
}

func TypesValid(s string) bool {
	types := strings.Split(strings.ToLower(s), ",")
	for _, t := range types {
		s := Type(strings.TrimSpace(t))
		if s.Valid() {
			return true
		}
	}
	return false
}

// Releasers returns the first two names in the production that have is_group as true.
// The one exception is if the production title contains a reference to a BBS or FTP site name.
// Then that title will be used as the first group returned.
func (p Production) Releasers() (string, string) {
	// find any reference to BBS or FTP in the production title to
	// obtain a possible site name.
	var a, b string
	// range through author nicks for any group matches
	for _, group := range p.Groups {
		if a == "" {
			a = group.Name
			continue
		}
		if b == "" {
			b = group.Name
			break
		}
	}
	return a, b
}

// Released returns the production's release date as date_issued_ year, month, day values.
func (p Production) Released() (int16, int16, int16) {
	return helper.Released(p.ReleaseDate)
}

// PlatformType parses the Pouet "platform" and "type" data
// and returns the corresponding platform and section tags.
// It returns -1 for an unknown platform or section.
func (p Production) PlatformType() (tags.Tag, tags.Tag) {
	var platform tags.Tag = -1
	switch {
	case p.Platforms.Windows.Slug != "":
		platform = tags.Windows
	case p.Platforms.MSDos.Slug != "":
		platform = tags.DOS
	case p.Platforms.DosGus.Slug != "":
		platform = tags.DOS
	}
	var section tags.Tag = -1
	types := strings.Split(p.Types.String(), ",")
	for _, t := range types {
		switch strings.TrimSpace(t) {
		case "artpack":
			section = tags.Pack
		case "bbstro":
			section = tags.BBS
		case "demo":
			section = tags.Demo
		case "diskmag":
			section = tags.Mag
		default:
			section = tags.Intro
		}
		if section != -1 {
			break
		}
	}
	return platform, section
}

// Votes is the production voting data from the Pouet API.
// The Pouet API returns values as null or string, so this struct
// is used to normalize the data types.
type Votes struct {
	// ID is the production ID.
	ID int `json:"id"`
	// Stars is the production rating using the average votes multiplied by 5.
	Stars float64 `json:"stars"`
	// VotesAvg is the average votes, the maximum value is 1.0.
	VotesAvg float64 `json:"votes_avg"`
	// VotesUp is the number of thumbs up votes.
	VotesUp uint64 `json:"votes_up"`
	// VotesMeh is the number of meh votes otherwise called piggies.
	VotesMeh uint64 `json:"votes_meh"`
	// VotesDown is the number of thumbs down votes.
	VotesDown uint64 `json:"votes_down"`
}

// useful for json data to struct creation,
// https://mholt.github.io/json-to-go/

// Response is the JSON response from the Pouet API with production voting data.
type Response struct {
	Prod struct {
		ID          string `json:"id"`          // ID is the prod ID.
		Title       string `json:"name"`        // Title is the prod title.
		ReleaseDate string `json:"releaseDate"` // ReleaseDate is the prod release date.
		Voteup      string `json:"voteup"`      // Voteup is the number of thumbs up votes.
		Votepig     string `json:"votepig"`     // Votepig is the number of meh votes.
		Votedown    string `json:"votedown"`    // Votedown is the number of thumbs down votes.
		Voteavg     string `json:"voteavg"`     // Voteavg is the average votes, the maximum value is 1.0.
		Download    string `json:"download"`    // Download is the first download link.
		Demozoo     string `json:"demozoo"`     // Demozoo is the first Demozoo link.
		Groups      []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"groups"` // Groups are the releasers that produced the prod.
		Platforms     Platforms `json:"platforms"` // Platforms are the platforms the prod runs on.
		Types         Types     `json:"types"`     // Types are the prod types.
		DownloadLinks []struct {
			Type string `json:"type"`
			Link string `json:"link"`
		} `json:"downloadLinks"` // DownloadLinks are the additional download links.
	} `json:"prod"` // Prod is the production data.
	Success bool `json:"success"` // Success is true if the prod data was found.
}

// Platforms are the supported platforms from the Pouet API.
type Platforms struct {
	DosGus  Platform `json:"69"` // MS-Dos with GUS
	Windows Platform `json:"68"` // Windows
	MSDos   Platform `json:"67"` // MS-Dos
}

func (p Platforms) String() string {
	s := []string{}
	if p.DosGus.Name != "" {
		s = append(s, p.DosGus.Slug)
	}
	if p.MSDos.Name != "" {
		s = append(s, p.MSDos.Slug)
	}
	if p.Windows.Name != "" {
		s = append(s, p.Windows.Slug)
	}
	return strings.Join(s, ", ")
}

func (p Platforms) Valid() bool {
	if p.DosGus.Slug == "msdosgus" {
		return true
	}
	if p.Windows.Slug == "windows" {
		return true
	}
	if p.MSDos.Slug == "msdos" {
		return true
	}
	return false
}

// Platform is the production platform data from the Pouet API.
type Platform struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// Type is the production type from the Pouet API.
type Type string

func (t Type) Valid() bool {
	switch t {
	case "dentro", "fastdemo", "invitation", "liveact", "musicdisk",
		"procedural graphics", "report", "slideshow", "votedisk", "wild":
		return false
	default:
		return true
	}
}

// Types are the production types from the Pouet API.
type Types []Type

func (t Types) Valid() bool {
	for _, t := range t {
		if t.Valid() {
			return true
		}
	}
	return false
}

func (t Types) String() string {
	s := []string{}
	for _, t := range t {
		s = append(s, string(t))
	}
	return strings.Join(s, ", ")
}

// Get retrieves the production voting data from the Pouet API.
// The id value is the Pouet production ID and must be greater than 0.
func (r *Response) Get(id int) (int, error) {
	if id < firstID {
		return 0, fmt.Errorf("%w: %d", ErrID, id)
	}
	url := ProdURL + strconv.Itoa(id)
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("get pouet production new request %w", err)
	}
	req.Header.Set("User-Agent", helper.UserAgent)
	res, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("get pouet production client do %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, res.Body)
		res.Body.Close()
		return res.StatusCode, fmt.Errorf("get pouet production %w: %s", ErrStatus, res.Status)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		_, _ = io.Copy(io.Discard, res.Body)
		res.Body.Close()
		return 0, fmt.Errorf("get pouet production read all %w", err)
	}
	err = json.Unmarshal(body, &r)
	clear(body)
	if err != nil {
		return 0, fmt.Errorf("get pouet production json unmarshal %w", err)
	}
	if !r.Success {
		return 0, fmt.Errorf("get pouet production %w: %d", ErrSuccess, id)
	}
	return 0, nil
}

// Votes retrieves the production voting data from the Pouet API.
// The id value is the Pouet production ID and must be greater than 0.
// The data is intended for the Artifact page, Pouët reviews section.
func (v *Votes) Votes(id int) error {
	if id < firstID {
		return fmt.Errorf("%w: %d", ErrID, id)
	}
	r := Response{}
	_, err := r.Get(id)
	if err != nil {
		return fmt.Errorf("pouet votes get %w", err)
	}
	v.ID, err = strconv.Atoi(r.Prod.ID)
	if err != nil {
		return fmt.Errorf("pouet votes atoi %w", err)
	}
	const base, bitSize = 10, 64
	v.VotesUp, err = strconv.ParseUint(r.Prod.Voteup, base, bitSize)
	if err != nil {
		return fmt.Errorf("pouet votes parse up %w", err)
	}
	v.VotesMeh, err = strconv.ParseUint(r.Prod.Votepig, base, bitSize)
	if err != nil {
		return fmt.Errorf("pouet votes parse pig %w", err)
	}
	v.VotesDown, err = strconv.ParseUint(r.Prod.Votedown, base, bitSize)
	if err != nil {
		return fmt.Errorf("pouet votes parse down %w", err)
	}
	v.VotesAvg, err = strconv.ParseFloat(r.Prod.Voteavg, 64)
	if err != nil {
		return fmt.Errorf("pouet votes parse average %w", err)
	}
	v.Stars = Stars(v.VotesUp, v.VotesMeh, v.VotesDown)
	return nil
}

// Stars returns the number of stars for the average votes.
// The value of votesAvg must be a valid float64 value and not greater than 1.0.
func Stars(up, ok, down uint64) float64 {
	if up+ok+down == 0 {
		return 0
	}
	const (
		scoreUp = 5
		scoreOk = 3
		scoreDn = 1
	)
	stars := float64(scoreUp*up+scoreOk*ok+scoreDn*down) / float64(up+ok+down)
	return math.Round(stars/StarRounder) * StarRounder
}
