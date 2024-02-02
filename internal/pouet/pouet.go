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
)

const (
	// ProdURL is the base URL for the Pouet production API.
	ProdURL = "https://api.pouet.net/v1/prod/?id="
	// Timeout is the HTTP client timeout.
	Timeout = 5 * time.Second
	// StarRounder is the rounding value for the stars rating.
	StarRounder = 0.5
	// User-Agent to send with the HTTP request.
	UserAgent = "Defacto2 2024 app, under construction (thanks!)"
	// firstID is the first production ID on Pouet.
	firstID = 1
)

var (
	ErrID      = errors.New("pouet production id is invalid")
	ErrSuccess = errors.New("pouet production not found")
	ErrStatus  = errors.New("pouet production status is not ok")
)

// Pouet is the production data from the Pouet API.
// The Pouet API returns values as null or string, so this struct
// is used to normalize the data types.
type Pouet struct {
	// Platforms are the platforms the prod runs on.
	Platforms Platforms `json:"platforms"`
	// Title is the prod title.
	Title string `json:"title"`
	// ReleaseDate is the prod release date.
	ReleaseDate string `json:"release_date"`
	// Platform is the prod platforms as a string.
	// If the string is empty then the prod is not supported.
	Platform string `json:"platform"`
	// Groups are the releasers that produced the prod.
	Groups []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"groups"`
	// Types are the prod types.
	Types Types `json:"types"`
	// ID is the prod ID.
	ID int `json:"id"`
	// Valid is true if this prod is a supported type and platform.
	Valid bool `json:"valid"`
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
		// used by uploader and voter
		ID string `json:"id"`
		// used by voter
		Voteup   string `json:"voteup"`
		Votepig  string `json:"votepig"`
		Votedown string `json:"votedown"`
		Voteavg  string `json:"voteavg"`
		// used by uploader
		Title       string `json:"name"`
		ReleaseDate string `json:"releaseDate"`
		Groups      []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"groups"`
		Platforms Platforms `json:"platforms"`
		Types     Types     `json:"types"`
	} `json:"prod"`
	Success bool `json:"success"`
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
		s = append(s, p.DosGus.Name)
	}
	if p.MSDos.Name != "" {
		s = append(s, p.MSDos.Name)
	}
	if p.Windows.Name != "" {
		s = append(s, p.Windows.Name)
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
func (r *Response) Get(id int) error {
	if id < firstID {
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
	req.Header.Set("User-Agent", UserAgent)
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
	err = json.Unmarshal(body, &r)
	if err != nil {
		return err
	}
	if !r.Success {
		return fmt.Errorf("%w: %d", ErrSuccess, id)
	}
	return nil
}

// Uploader retrieves and parses the production data from the Pouet API.
// The id value is the Pouet production ID and must be greater than 0.
// The data is intended for the Pouet Uploader.
func (p *Pouet) Uploader(id int) error {
	if id < firstID {
		return fmt.Errorf("%w: %d", ErrID, id)
	}
	r := Response{}
	err := r.Get(id)
	if err != nil {
		return err
	}
	p.ID, err = strconv.Atoi(r.Prod.ID)
	if err != nil {
		return err
	}
	p.Title = r.Prod.Title
	p.ReleaseDate = r.Prod.ReleaseDate
	p.Groups = r.Prod.Groups
	p.Platforms = r.Prod.Platforms
	p.Types = r.Prod.Types
	p.Platform = r.Prod.Platforms.String()
	p.Valid = r.Prod.Platforms.Valid() && r.Prod.Types.Valid()
	return nil
}

// Votes retrieves the production voting data from the Pouet API.
// The id value is the Pouet production ID and must be greater than 0.
// The data is intended for the About file page, Pouët reviews section.
func (v *Votes) Votes(id int) error {
	if id < firstID {
		return fmt.Errorf("%w: %d", ErrID, id)
	}
	r := Response{}
	err := r.Get(id)
	if err != nil {
		return err
	}
	v.ID, err = strconv.Atoi(r.Prod.ID)
	if err != nil {
		return err
	}
	const base, bitSize = 10, 64
	v.VotesUp, err = strconv.ParseUint(r.Prod.Voteup, base, bitSize)
	if err != nil {
		return err
	}
	v.VotesMeh, err = strconv.ParseUint(r.Prod.Votepig, base, bitSize)
	if err != nil {
		return err
	}
	v.VotesDown, err = strconv.ParseUint(r.Prod.Votedown, base, bitSize)
	if err != nil {
		return err
	}
	v.VotesAvg, err = strconv.ParseFloat(r.Prod.Voteavg, 64)
	if err != nil {
		return err
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
