package pouet

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"
)

const (
	// ProdURL is the base URL for the Pouet production API.
	ProdURL = "https://api.pouet.net/v1/prod/?id="
	// Timeout is the HTTP client timeout.
	Timeout = 5 * time.Second
	// StarRounder is the rounding value for the stars rating.
	StarRounder = 0.5
	// firstID is the first production ID on Pouet.
	firstID = 1
)

var (
	ErrID      = errors.New("pouet production id is invalid")
	ErrSuccess = errors.New("pouet production not found")
	ErrStatus  = errors.New("pouet production status is not ok")
)

// Pouet is the production voting data from the Pouet API.
type Pouet struct {
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
	Success bool `json:"success"`
	Prod    struct {
		ID       string `json:"id"`
		Voteup   string `json:"voteup"`
		Votepig  string `json:"votepig"`
		Votedown string `json:"votedown"`
		Voteavg  string `json:"voteavg"`
	} `json:"prod"`
}

// Votes retrieves the production voting data from the Pouet API.
// The id value is the Pouet production ID and must be greater than 0.
func (p *Pouet) Votes(id int) error {
	if id < firstID {
		return fmt.Errorf("%w: %d", ErrID, id)
	}
	client := http.Client{
		Timeout: Timeout,
	}
	url := ProdURL + strconv.Itoa(id)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent",
		"Defacto2 2023 app under construction (thanks!)")
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
	r := Response{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return err
	}
	if !r.Success {
		return fmt.Errorf("%w: %d", ErrSuccess, id)
	}
	p.ID, err = strconv.Atoi(r.Prod.ID)
	if err != nil {
		return err
	}
	const base, bitSize = 10, 64
	p.VotesUp, err = strconv.ParseUint(r.Prod.Voteup, base, bitSize)
	if err != nil {
		return err
	}
	p.VotesMeh, err = strconv.ParseUint(r.Prod.Votepig, base, bitSize)
	if err != nil {
		return err
	}
	p.VotesDown, err = strconv.ParseUint(r.Prod.Votedown, base, bitSize)
	if err != nil {
		return err
	}
	p.VotesAvg, err = strconv.ParseFloat(r.Prod.Voteavg, 64)
	if err != nil {
		return err
	}
	p.Stars = Stars(p.VotesUp, p.VotesMeh, p.VotesDown)
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
