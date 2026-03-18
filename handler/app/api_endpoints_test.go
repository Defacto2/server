package app_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/nalgeon/be"
)

// TestAllAPIEndpoints tests all API endpoints from apiinfo.tmpl.
func TestAllAPIEndpoints(t *testing.T) { //nolint:gocognit
	endpoints := []struct {
		name        string
		path        string
		expectCount bool
		expectArray bool
	}{
		// Basic endpoints
		{"boards", "/boards", true, true},
		{"categories", "/categories", true, true},
		{"websites", "/websites", true, false},
		{"demozoo", "/demozoo", true, false},
		{"groups", "/groups", true, false},
		{"magazines", "/magazines", true, true},
		{"platforms", "/platforms", true, true},
		{"sceners", "/sceners", true, false},
		{"sceners/artist", "/sceners/artist", true, false},
		{"sceners/coder", "/sceners/coder", true, false},
		{"sceners/musician", "/sceners/musician", true, false},
		{"sceners/writer", "/sceners/writer", true, false},
		{"areacodes", "/areacodes", true, true},
		{"areacodes/regions", "/areacodes/regions", true, true},
		{"milestones", "/milestones", true, true},
		{"milestones/highlights", "/milestones/highlights", true, true},
		// Category endpoints
		{"category/announcements", "/category/announcements", true, false},
		{"category/demo", "/category/demo", true, false},
		// Platform endpoints
		{"platform/ansi", "/platform/ansi", true, false},
		{"platform/dos", "/platform/dos", true, false},
		// File endpoints
		{"files", "/files", true, false},
		{"files/new", "/files/new", true, false},
		// Specific releaser endpoints
		{"releaser/defacto2", "/releaser/defacto2", true, false},
		{"releaser/razor-1911", "/releaser/razor-1911", true, false},
		// Area code specific endpoints
		{"areacodes/212", "/areacodes/212", false, false},
		{"areacodes/regions/CA", "/areacodes/regions/CA", false, false},
		{"areacodes/search/california", "/areacodes/search/california", true, false},
		// Milestone specific endpoints
		{"milestones/year/1995", "/milestones/year/1995", true, true},
		{"milestones/years/1990-2000", "/milestones/years/1990-2000", true, true},
		{"milestones/decade/1990s", "/milestones/decade/1990s", true, true},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.name, func(t *testing.T) {
			url := api + endpoint.path
			client := http.Client{}
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
			be.Equal(t, err, nil)

			resp, err := client.Do(req)
			be.Equal(t, err, nil)
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Log(err)
				}
			}()

			// Check HTTP status
			be.Equal(t, http.StatusOK, resp.StatusCode)

			// Try to parse as object first, fall back to array if that fails
			var result map[string]any
			err = json.NewDecoder(resp.Body).Decode(&result)
			if err != nil {
				// Try parsing as array
				var arrayResult []map[string]any
				if err := resp.Body.Close(); err != nil {
					t.Log(err)
				}
				req2, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
				resp2, err2 := client.Do(req2)
				be.Equal(t, err2, nil)
				defer func() {
					if err := resp2.Body.Close(); err != nil {
						t.Log(err)
					}
				}()
				err = json.NewDecoder(resp2.Body).Decode(&arrayResult)
				be.Equal(t, err, nil)
				be.True(t, len(arrayResult) > 0)

				// Check for expected fields based on endpoint type for array responses
				switch endpoint.name {
				case "categories", "platforms", "areacodes", "areacodes/regions", "milestones", "milestones/highlights", "magazines":
					be.True(t, len(arrayResult) > 0)
				case "milestones/year/1995", "milestones/years/1990-2000", "milestones/decade/1990s":
					be.True(t, len(arrayResult) > 0)
				}
				return
			}

			be.True(t, len(result) > 0)

			// Check for count field if expected
			if endpoint.expectCount {
				count, ok := result["count"].(float64)
				if ok {
					be.True(t, int(count) >= 0)
				}
			}

			// Check for expected fields based on endpoint type for object responses
			switch endpoint.name {
			case "boards":
				be.True(t, result["releasers"] != nil)
				be.True(t, result["page"] != nil)
			case "websites":
				be.True(t, result["websites"] != nil)
				be.True(t, result["count"] != nil)
			case "demozoo":
				be.True(t, result["groups"] != nil)
				be.True(t, result["count"] != nil)
			case "groups":
				be.True(t, result["releasers"] != nil)
				be.True(t, result["page"] != nil)
				be.True(t, result["totalPages"] != nil)
			case "sceners":
				be.True(t, result["sceners"] != nil)
				be.True(t, result["page"] != nil)
			case "sceners/artist", "sceners/coder", "sceners/musician", "sceners/writer":
				be.True(t, result["sceners"] != nil)
				be.True(t, result["page"] != nil)
			case "category/announcements", "category/demo":
				be.True(t, result["files"] != nil)
			case "platform/ansi", "platform/dos":
				be.True(t, result["files"] != nil)
			case "files", "files/new":
				be.True(t, result["files"] != nil)
				be.True(t, result["page"] != nil)
			case "releaser/defacto2", "releaser/razor-1911":
				be.True(t, result["group"] != nil)
				be.True(t, result["files"] != nil)
			case "areacodes/212":
				be.True(t, result["code"] != nil)
				be.True(t, result["territories"] != nil)
			case "areacodes/regions/CA":
				be.True(t, result["name"] != nil)
				be.True(t, result["abbreviation"] != nil)
				be.True(t, result["areaCodes"] != nil)
			case "areacodes/search/california":
				be.True(t, result["areacodes"] != nil)
				be.True(t, result["territories"] != nil)
			}
		})
	}
}

// TestAPIEndpointExamples tests the specific examples from apiinfo.tmpl.
func TestAPIEndpointExamples(t *testing.T) {
	examples := []struct {
		name string
		url  string
	}{
		{"files", api + "/files"},
		{"files/new", api + "/files/new"},
		{"categories", api + "/categories"},
		{"category/announcements", api + "/category/announcements"},
		{"magazines", api + "/magazines"},
		{"releaser/defacto2", api + "/releaser/defacto2"},
		{"milestones/year/1971", api + "/milestones/year/1971"},
		{"areacodes/regions/CA", api + "/areacodes/regions/CA"},
		{"sceners/artist", api + "/sceners/artist"},
		{"scener/dubmood", api + "/scener/dubmood"},
		{"websites", api + "/websites"},
		{"demozoo", api + "/demozoo"},
	}

	for _, example := range examples {
		t.Run(example.name, func(t *testing.T) {
			client := http.Client{}
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, example.url, nil)
			be.Equal(t, err, nil)

			resp, err := client.Do(req)
			be.Equal(t, err, nil)
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Log(err)
				}
			}()

			// Check HTTP status
			be.Equal(t, http.StatusOK, resp.StatusCode)

			// Try to parse as object first, fall back to array if that fails
			var result map[string]any
			err = json.NewDecoder(resp.Body).Decode(&result)
			if err != nil {
				// Try parsing as array
				var arrayResult []map[string]any
				if err := resp.Body.Close(); err != nil {
					t.Log(err)
				}
				req2, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, example.url, nil)
				resp2, err2 := client.Do(req2)
				be.Equal(t, err2, nil)
				defer func() {
					if err := resp2.Body.Close(); err != nil {
						t.Log(err)
					}
				}()
				err = json.NewDecoder(resp2.Body).Decode(&arrayResult)
				be.Equal(t, err, nil)
				be.True(t, len(arrayResult) > 0)
				// Verify content based on endpoint type for array responses
				switch example.name {
				case "categories", "magazines", "milestones/year/1971":
					be.True(t, len(arrayResult) > 0)
				}
				return
			}

			be.True(t, len(result) > 0)

			// Verify content based on endpoint type for object responses
			switch example.name {
			case "files", "files/new":
				be.True(t, result["files"] != nil)
				be.True(t, result["page"] != nil)
			case "category/announcements":
				be.True(t, result["files"] != nil)
			case "releaser/defacto2":
				be.True(t, result["group"] != nil)
				be.True(t, result["files"] != nil)
			case "areacodes/regions/CA":
				be.True(t, result["name"] != nil)
				be.True(t, result["abbreviation"] != nil)
			case "sceners/artist":
				be.True(t, result["sceners"] != nil)
			case "scener/dubmood":
				be.True(t, result["scener"] != nil)
			case "websites":
				be.True(t, result["websites"] != nil)
				be.True(t, result["count"] != nil)
			case "demozoo":
				be.True(t, result["groups"] != nil)
				be.True(t, result["count"] != nil)
			}
		})
	}
}

// TestAPIResponseValidation tests JSON validation and expected values for key endpoints.
func TestAPIResponseValidation(t *testing.T) {
	client := http.Client{}

	// Test websites endpoint
	t.Run("websites", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, api+"/websites", nil)
		resp, err := client.Do(req)
		be.Equal(t, err, nil)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Log(err)
			}
		}()
		be.Equal(t, http.StatusOK, resp.StatusCode)

		var result struct {
			Websites []struct {
				Title    string `json:"title"`
				URL      string `json:"url"`
				Info     string `json:"info"`
				Category string `json:"category"`
				Working  bool   `json:"working"`
			} `json:"websites"`
			Count int `json:"count"`
		}

		body, err := io.ReadAll(resp.Body)
		be.Equal(t, err, nil)

		err = json.Unmarshal(body, &result)
		be.Equal(t, err, nil)

		be.True(t, result.Count > 0)
		be.True(t, len(result.Websites) > 0)

		// Validate first website
		if len(result.Websites) > 0 {
			website := result.Websites[0]
			be.True(t, len(website.Title) > 0)
			be.True(t, len(website.URL) > 0)
			be.True(t, len(website.Category) > 0)
			be.True(t, strings.HasPrefix(website.URL, "http"))
		}
	})

	// Test demozoo endpoint
	t.Run("demozoo", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, api+"/demozoo", nil)
		resp, err := client.Do(req)
		be.Equal(t, err, nil)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Log(err)
			}
		}()

		be.Equal(t, http.StatusOK, resp.StatusCode)

		var result struct {
			Groups []struct {
				URI string `json:"uri"`
				ID  int    `json:"id"`
				URL string `json:"url"`
			} `json:"groups"`
			Count int `json:"count"`
		}

		body, err := io.ReadAll(resp.Body)
		be.Equal(t, err, nil)

		err = json.Unmarshal(body, &result)
		be.Equal(t, err, nil)

		be.True(t, result.Count > 0)
		be.True(t, len(result.Groups) > 0)

		// Validate first group
		if len(result.Groups) > 0 {
			group := result.Groups[0]
			be.True(t, len(group.URI) > 0)
			be.True(t, group.ID > 0)
			be.True(t, strings.HasPrefix(group.URL, "https://demozoo.org"))
		}
	})

	// Test groups endpoint
	t.Run("groups", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, api+"/groups", nil)
		resp, err := client.Do(req)
		be.Equal(t, err, nil)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Log(err)
			}
		}()

		be.Equal(t, http.StatusOK, resp.StatusCode)

		var result struct {
			Releasers []struct {
				ID    string `json:"id"`
				Name  string `json:"name"`
				Title string `json:"title"`
				URLs  struct {
					API  string `json:"api"`
					HTML string `json:"html"`
				} `json:"urls"`
				Statistics struct {
					TotalFiles     int    `json:"totalFiles"`
					TotalSize      string `json:"totalSize"`
					TotalSizeBytes int    `json:"totalSizeBytes"`
				} `json:"statistics"`
			} `json:"releasers"`
			Page       int `json:"page"`
			Totals     int `json:"totals"`
			TotalPages int `json:"totalPages"`
		}

		body, err := io.ReadAll(resp.Body)
		be.Equal(t, err, nil)

		err = json.Unmarshal(body, &result)
		be.Equal(t, err, nil)

		be.True(t, len(result.Releasers) > 0)
		be.Equal(t, 1, result.Page)
		be.True(t, result.TotalPages >= 1)

		// Validate first releaser
		if len(result.Releasers) > 0 {
			releaser := result.Releasers[0]
			be.True(t, len(releaser.ID) > 0)
			be.True(t, len(releaser.Name) > 0)
			// be.True(t, len(releaser.Title) > 0) // Title may be empty for some releasers
			be.True(t, strings.HasPrefix(releaser.URLs.API, app.APIBase+"/releaser/"))
			be.True(t, strings.HasPrefix(releaser.URLs.HTML, "/g/"))
			be.True(t, releaser.Statistics.TotalFiles >= 0)
			be.True(t, releaser.Statistics.TotalSizeBytes >= 0)
		}
	})

	// Test areacodes endpoint
	t.Run("areacodes", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, api+"/areacodes", nil)
		resp, err := client.Do(req)
		be.Equal(t, err, nil)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Log(err)
			}
		}()

		be.Equal(t, http.StatusOK, resp.StatusCode)

		var result []struct {
			Code        int      `json:"code"`
			Territories []string `json:"territories"`
			Notes       string   `json:"notes,omitempty"`
		}

		body, err := io.ReadAll(resp.Body)
		be.Equal(t, err, nil)

		err = json.Unmarshal(body, &result)
		be.Equal(t, err, nil)

		be.True(t, len(result) > 0)

		// Validate first area code
		if len(result) > 0 {
			areacode := result[0]
			be.True(t, areacode.Code > 0)
			be.True(t, len(areacode.Territories) > 0)
		}
	})

	// Test milestones endpoint
	t.Run("milestones", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, api+"/milestones", nil)
		resp, err := client.Do(req)
		be.Equal(t, err, nil)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Log(err)
			}
		}()

		be.Equal(t, http.StatusOK, resp.StatusCode)

		var result []struct {
			Title     string `json:"Title"`
			Year      int    `json:"Year"`
			Month     int    `json:"Month"`
			Day       int    `json:"Day"`
			Lead      string `json:"Lead"`
			Content   string `json:"Content"`
			Highlight bool   `json:"Highlight"`
			Picture   struct {
				Title       string `json:"Title"`
				Alt         string `json:"Alt"`
				Attribution string `json:"Attribution"`
				License     string `json:"License"`
			} `json:"Picture"`
		}

		body, err := io.ReadAll(resp.Body)
		be.Equal(t, err, nil)

		err = json.Unmarshal(body, &result)
		be.Equal(t, err, nil)

		be.True(t, len(result) > 0)

		// Validate first milestone
		if len(result) > 0 {
			milestone := result[0]
			be.True(t, len(milestone.Title) > 0)
			be.True(t, milestone.Year > 0)
			be.True(t, len(milestone.Lead) > 0)
		}
	})
}
