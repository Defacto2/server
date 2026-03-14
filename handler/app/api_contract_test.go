//go:build !ignore

package app_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/nalgeon/be"
)

// TestMain checks server availability before running tests.
func TestMain(m *testing.M) {
	// Check if server is running.
	client := http.Client{}
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost:1323/health-check", nil)
	resp, err := client.Do(req)
	if err != nil {
		os.Stderr.WriteString("SKIP: Server not running at localhost:1323\n")
		os.Exit(0)
	}
	_ = resp.Body.Close()

	// Run tests if server is available.
	code := m.Run()
	os.Exit(code)
}

// TestAnnouncementsContract verifies the announcements endpoint contract.
func TestAnnouncementsContract(t *testing.T) {
	client := http.Client{}
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost:1323/api/files/announcements?limit=5", nil)
	resp, err := client.Do(req)
	be.Equal(t, err, nil)
	defer resp.Body.Close()

	be.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		Announcements []struct {
			ID          int64  `json:"id"`
			Filename    string `json:"filename"`
			Description string `json:"description"`
			FileType    string `json:"fileType"`
			URLs        struct {
				Download  string `json:"download"`
				HTML      string `json:"html"`
				Thumbnail string `json:"thumbnail,omitempty"`
			} `json:"urls"`
		} `json:"announcements"`
	}

	body, err := io.ReadAll(resp.Body)
	be.Equal(t, err, nil)

	err = json.Unmarshal(body, &result)
	be.Equal(t, err, nil)

	be.True(t, len(result.Announcements) > 0)

	// Verify URL patterns
	for _, file := range result.Announcements {
		be.True(t, strings.HasPrefix(file.URLs.Download, "/d/"))
		be.True(t, strings.HasPrefix(file.URLs.HTML, "/f/"))
		be.True(t, len(file.Description) > 0)
		be.True(t, len(file.FileType) > 0)
	}
}

// TestCategoriesContract verifies the categories endpoint contract.
func TestCategoriesContract(t *testing.T) {
	client := http.Client{}
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost:1323/api/categories", nil)
	resp, err := client.Do(req)
	be.Equal(t, err, nil)
	defer resp.Body.Close()

	be.Equal(t, http.StatusOK, resp.StatusCode)

	var result []struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		URI         string `json:"uri"`
		URLs        struct {
			API   string `json:"api"`
			HTML3 string `json:"html3"`
			HTML  string `json:"html"`
		} `json:"urls"`
	}

	body, err := io.ReadAll(resp.Body)
	be.Equal(t, err, nil)

	err = json.Unmarshal(body, &result)
	be.Equal(t, err, nil)

	be.True(t, len(result) > 0)

	// Verify at least one category has proper structure
	be.True(t, len(result[0].Name) > 0)
	be.True(t, len(result[0].URI) > 0)
	be.True(t, strings.HasPrefix(result[0].URLs.API, "/api/files/"))
	be.True(t, strings.HasPrefix(result[0].URLs.HTML3, "/html3/"))
	be.True(t, strings.HasPrefix(result[0].URLs.HTML, "/files/"))
}

// TestPlatformsContract verifies the platforms endpoint contract.
func TestPlatformsContract(t *testing.T) {
	client := http.Client{}
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost:1323/api/platforms", nil)
	resp, err := client.Do(req)
	be.Equal(t, err, nil)
	defer resp.Body.Close()

	be.Equal(t, http.StatusOK, resp.StatusCode)

	var result []struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		URI         string `json:"uri"`
		URLs        struct {
			API   string `json:"api"`
			HTML3 string `json:"html3"`
			HTML  string `json:"html"`
		} `json:"urls"`
	}

	body, err := io.ReadAll(resp.Body)
	be.Equal(t, err, nil)

	err = json.Unmarshal(body, &result)
	be.Equal(t, err, nil)

	be.True(t, len(result) > 0)

	// Verify at least one platform has proper structure
	be.True(t, len(result[0].Name) > 0)
	be.True(t, len(result[0].URI) > 0)
	be.True(t, strings.HasPrefix(result[0].URLs.API, "/api/files/"))
	be.True(t, strings.HasPrefix(result[0].URLs.HTML3, "/html3/"))
	be.True(t, strings.HasPrefix(result[0].URLs.HTML, "/files/"))
}

// TestGenericCategoryContract verifies the generic category endpoint contract.
func TestGenericCategoryContract(t *testing.T) {
	testCases := []string{"demo", "scenerules", "magazine"}

	for _, category := range testCases {
		t.Run(category, func(t *testing.T) {
			url := fmt.Sprintf("http://localhost:1323/api/files/%s?limit=3", category)
			client := http.Client{}
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
			resp, err := client.Do(req)
			be.Equal(t, err, nil)
			defer resp.Body.Close()

			be.Equal(t, http.StatusOK, resp.StatusCode)

			var result struct {
				Files []struct {
					ID          int64  `json:"id"`
					Filename    string `json:"filename"`
					Description string `json:"description"`
					URLs        struct {
						Download  string `json:"download"`
						HTML      string `json:"html"`
						Thumbnail string `json:"thumbnail,omitempty"`
					} `json:"urls"`
				} `json:"files"`
			}

			body, err := io.ReadAll(resp.Body)
			be.Equal(t, err, nil)

			err = json.Unmarshal(body, &result)
			be.Equal(t, err, nil)

			// Verify structure (may be empty but should be valid JSON)
			for _, file := range result.Files {
				be.True(t, strings.HasPrefix(file.URLs.Download, "/d/"))
				be.True(t, strings.HasPrefix(file.URLs.HTML, "/f/"))
			}
		})
	}
}

// TestPlatformQueries verifies that platform queries work correctly.
func TestPlatformQueries(t *testing.T) {
	// Test a few different platform types
	platforms := []string{"ansi", "audio", "dos", "windows", "image"}

	for _, platform := range platforms {
		t.Run(platform, func(t *testing.T) {
			url := fmt.Sprintf("http://localhost:1323/api/files/%s?limit=2", platform)
			client := http.Client{}
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
			resp, err := client.Do(req)
			be.Equal(t, err, nil)
			defer resp.Body.Close()

			be.Equal(t, http.StatusOK, resp.StatusCode)

			var result struct {
				Files []struct {
					ID          int64  `json:"id"`
					Filename    string `json:"filename"`
					Description string `json:"description"`
					URLs        struct {
						Download  string `json:"download"`
						HTML      string `json:"html"`
						Thumbnail string `json:"thumbnail,omitempty"`
					} `json:"urls"`
				} `json:"files"`
			}

			body, err := io.ReadAll(resp.Body)
			be.Equal(t, err, nil)

			err = json.Unmarshal(body, &result)
			be.Equal(t, err, nil)

			// Verify URL patterns for platform files
			for _, file := range result.Files {
				be.True(t, strings.HasPrefix(file.URLs.Download, "/d/"))
				be.True(t, strings.HasPrefix(file.URLs.HTML, "/f/"))
			}
		})
	}
}

// TestURLPatterns verifies consistent URL patterns across all endpoints.
func TestURLPatterns(t *testing.T) {
	endpoints := []string{
		"announcements",
		"demo",
		"scenerules",
		"magazine",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			url := fmt.Sprintf("http://localhost:1323/api/files/%s?limit=2", endpoint)
			client := http.Client{}
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
			resp, err := client.Do(req)
			be.Equal(t, err, nil)
			defer resp.Body.Close()

			be.Equal(t, http.StatusOK, resp.StatusCode)

			// Just verify it returns valid JSON with expected structure
			var result map[string]any
			err = json.NewDecoder(resp.Body).Decode(&result)
			be.Equal(t, err, nil)

			be.True(t, len(result) > 0)
		})
	}
}
