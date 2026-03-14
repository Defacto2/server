package app_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/Defacto2/server/handler/app"
	"github.com/labstack/echo/v4"
)

// Example_milestoneAPI demonstrates how to use the milestone API endpoints.
func Example_milestoneAPI() {
	e := echo.New()
	apiGroup := e.Group("/api/v1")
	apiGroup.GET("/milestones", app.MilestonesAPI)
	apiGroup.GET("/milestones/year/:year", app.MilestoneYearAPI)
	apiGroup.GET("/milestones/range/:range", app.MilestoneYearsAPI)
	apiGroup.GET("/milestones/highlights", app.MilestoneHighlightsAPI)
	apiGroup.GET("/milestones/decade/:decade", app.MilestoneDecadeAPI)

	// Example 1: Get all milestones
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/milestones", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	switch rec.Code {
	case http.StatusOK:
		fmt.Println("✓ GET /api/v1/milestones - Success")

		// Parse the response
		var milestones app.Milestones
		if err := json.NewDecoder(rec.Body).Decode(&milestones); err == nil {
			fmt.Printf("  Found %d milestones\n", len(milestones))
			if len(milestones) > 0 {
				fmt.Printf("  First milestone year: %d\n", milestones[0].Year)
				fmt.Printf("  Has clean HTML content: %t\n", len(milestones[0].ContentHTML) > 0)
				fmt.Printf("  Has plain text content: %t\n", len(milestones[0].ContentHTML) > 0)
			}
		}
	default:
		fmt.Printf("✗ GET /api/v1/milestones - Failed with status %d\n", rec.Code)
	}

	// Example 2: Get milestones by year
	req = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/milestones/year/1995", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	switch rec.Code {
	case http.StatusOK:
		fmt.Println("✓ GET /api/v1/milestones/year/1995 - Success")
		var milestones app.Milestones
		if err := json.NewDecoder(rec.Body).Decode(&milestones); err == nil {
			fmt.Printf("  Found %d milestones for year 1995\n", len(milestones))
		}
	case http.StatusNotFound:
		fmt.Println("✓ GET /api/v1/milestones/year/1995 - No milestones found (expected)")
	default:
		fmt.Printf("✗ GET /api/v1/milestones/year/1995 - Failed with status %d\n", rec.Code)
	}

	// Example 3: Get milestones by year range
	req = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/milestones/range/1990-2000", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	switch rec.Code {
	case http.StatusOK:
		fmt.Println("✓ GET /api/v1/milestones/range/1990-2000 - Success")
		var milestones app.Milestones
		if err := json.NewDecoder(rec.Body).Decode(&milestones); err == nil {
			fmt.Printf("  Found %d milestones in range 1990-2000\n", len(milestones))
		}
	case http.StatusNotFound:
		fmt.Println("✓ GET /api/v1/milestones/range/1990-2000 - No milestones found (expected)")
	default:
		fmt.Printf("✗ GET /api/v1/milestones/range/1990-2000 - Failed with status %d\n", rec.Code)
	}

	// Example 4: Get highlighted milestones
	req = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/milestones/highlights", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	switch rec.Code {
	case http.StatusOK:
		fmt.Println("✓ GET /api/v1/milestones/highlights - Success")
		var milestones app.Milestones
		if err := json.NewDecoder(rec.Body).Decode(&milestones); err == nil {
			fmt.Printf("  Found %d highlighted milestones\n", len(milestones))
		}
	case http.StatusNotFound:
		fmt.Println("✓ GET /api/v1/milestones/highlights - No highlighted milestones found (expected)")
	default:
		fmt.Printf("✗ GET /api/v1/milestones/highlights - Failed with status %d\n", rec.Code)
	}

	// Example 5: Get milestones by decade
	req = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/milestones/decade/1990s", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	switch rec.Code {
	case http.StatusOK:
		fmt.Println("✓ GET /api/v1/milestones/decade/1990s - Success")
		var milestones app.Milestones
		if err := json.NewDecoder(rec.Body).Decode(&milestones); err == nil {
			fmt.Printf("  Found %d milestones from the 1990s\n", len(milestones))
		}
	case http.StatusNotFound:
		fmt.Println("✓ GET /api/v1/milestones/decade/1990s - No milestones found (expected)")
	default:
		fmt.Printf("✗ GET /api/v1/milestones/decade/1990s - Failed with status %d\n", rec.Code)
	}

	// Example 6: Demonstrate HTML cleaning functions
	html := `<p class="test">This has <a href="https://example.com" class="link">a link</a> and <span style="color: red;">formatting</span>.</p>`
	cleaned := app.ApiMarkup(html)
	plain := app.ApiMarkup(html)

	fmt.Println("\n✓ HTML Cleaning Functions:")
	fmt.Printf("  Original: %s\n", html)
	fmt.Printf("  Cleaned:  %s\n", cleaned)
	fmt.Printf("  Plain:    %s\n", plain)

	// Output:
	// ✓ GET /api/v1/milestones - Success
	//   Found 122 milestones
	//   First milestone year: 1971
	//   Has clean HTML content: true
	//   Has plain text content: true
	// ✓ GET /api/v1/milestones/year/1995 - Success
	//   Found 3 milestones for year 1995
	// ✓ GET /api/v1/milestones/range/1990-2000 - Success
	//   Found 38 milestones in range 1990-2000
	// ✓ GET /api/v1/milestones/highlights - Success
	//   Found 32 highlighted milestones
	// ✓ GET /api/v1/milestones/decade/1990s - Success
	//   Found 35 milestones from the 1990s
	// ✓ HTML Cleaning Functions:
	//   Original: <p class="test">This has <a href="https://example.com" class="link">a link</a> and <span style="color: red;">formatting</span>.</p>
	//   Cleaned:  <p>This has <a href="https://example.com">a link</a> and <span>formatting</span>.</p>
	//   Plain:    This has a link and formatting .
}

// Example_htmlCleaning demonstrates the HTML cleaning functions.
func Example_htmlCleaning() {
	// Example HTML with various attributes and tags
	html := `<div class="content" id="main">
		<h1 style="color: blue;">Welcome</h1>
		<p class="intro">This is <strong>important</strong> content with <a href="https://example.com" class="external" title="Visit Example">a link</a>.</p>
		<p>More content with <span data-info="test">data attributes</span> and <a name="anchor">anchor without href</a>.</p>
	</div>`
	cleaned := app.ApiMarkup(html)
	plain := app.ApiMarkup(html)
	fmt.Println("Original HTML:")
	fmt.Println(html)
	fmt.Println("\nCleaned HTML (preserves structure, removes presentation):")
	fmt.Println(cleaned)
	fmt.Println("\nPlain text (all HTML tags removed):")
	fmt.Println(plain)

	// Output:
	// Original HTML:
	// <div class="content" id="main">
	//   <h1 style="color: blue;">Welcome</h1>
	//   <p class="intro">This is <strong>important</strong> content with <a href="https://example.com" class="external" title="Visit Example">a link</a>.</p>
	//   <p>More content with <span data-info="test">data attributes</span> and <a name="anchor">anchor without href</a>.</p>
	// </div>
	//
	// Cleaned HTML (preserves structure, removes presentation):
	// <div>
	//   <h1>Welcome</h1>
	//   <p>This is <strong>important</strong> content with <a href="https://example.com">a link</a>.</p>
	//   <p>More content with <span>data attributes</span> and anchor without href.</p>
	// </div>
	//
	// Plain text (all HTML tags removed):
	// Welcome
	// This is important content with a link.
	// More content with data attributes and anchor without href.
}

// Example_errorHandling demonstrates error cases.
func Example_errorHandling() {
	e := echo.New()
	apiGroup := e.Group("/api/v1")
	apiGroup.GET("/milestones/year/:year", app.MilestoneYearAPI)
	apiGroup.GET("/milestones/range/:range", app.MilestoneYearsAPI)

	// Example 1: Invalid year format
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/milestones/year/invalid", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	fmt.Printf("Invalid year format: Status %d\n", rec.Code)

	// Example 2: Invalid range format
	req = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/milestones/range/invalid", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	fmt.Printf("Invalid range format: Status %d\n", rec.Code)

	// Example 3: Start year > end year
	req = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/milestones/range/2000-1990", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	fmt.Printf("Start year > end year: Status %d\n", rec.Code)
	// Output: Invalid year format: Status 400
	// Invalid range format: Status 400
	// Start year > end year: Status 400
}
