package app

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"hash/fnv"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/handler/app/internal/filerecord"
	"github.com/Defacto2/server/internal/panics"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
)

// hashString creates a stable hash ID from a string and returns it as base64.
func hashString(s string) string {
	h := fnv.New64a()
	h.Write([]byte(s))
	hashBytes := h.Sum(nil)
	// Use URLEncoding to avoid special characters
	return base64.URLEncoding.EncodeToString(hashBytes)
}

// ReleaserAPI represents a releaser/group for API responses.
type ReleaserAPI struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Title string `json:"title"`
	URLs  struct {
		API   string `json:"api"`
		HTML3 string `json:"html3"`
		HTML  string `json:"html"`
	} `json:"urls"`
	Statistics struct {
		TotalFiles     int64  `json:"totalFiles"`
		TotalSize      string `json:"totalSize"`
		TotalSizeBytes int64  `json:"totalSizeBytes"`
	} `json:"statistics"`
}

// getReleasersCount returns the total number of releasers efficiently.
func getReleasersCount(ctx context.Context, db *sql.DB) (int, error) {
	// Use DistinctGroups to get just the releaser names that have files
	// This is much more efficient than loading full data and matches what Limit() returns
	var names model.ReleaserNames
	if err := names.DistinctGroups(ctx, db); err != nil {
		return 0, fmt.Errorf("count releasers: %w", err)
	}
	return len(names), nil
}

// getReleasersWithStats builds the ReleaserAPI list from model data.
func getReleasersWithStats(releasers model.Releasers) []ReleaserAPI {
	results := make([]ReleaserAPI, 0, len(releasers))
	for _, rel := range releasers {
		relname := releaser.Link(rel.Unique.Name)
		relURI := releaser.Obfuscate(rel.Unique.Name)

		// Use the data already in the Releaser struct (much faster than querying again)
		// The Count and Bytes are already loaded by the Limit() method
		count := rel.Unique.Count
		bytes := rel.Unique.Bytes

		// Create stable ID from the URL (obfuscated name)
		stableID := hashString(relURI)

		result := ReleaserAPI{
			ID:    stableID,
			Name:  relURI,
			Title: relname,
			URLs: struct {
				API   string `json:"api"`
				HTML3 string `json:"html3"`
				HTML  string `json:"html"`
			}{
				API:   "/api/group/" + relURI,
				HTML3: "/html3/group/" + relURI,
				HTML:  "/g/" + relURI,
			},
			Statistics: struct {
				TotalFiles     int64  `json:"totalFiles"`
				TotalSize      string `json:"totalSize"`
				TotalSizeBytes int64  `json:"totalSizeBytes"`
			}{
				TotalFiles:     int64(count),
				TotalSize:      helper.ByteCount(int64(bytes)),
				TotalSizeBytes: int64(bytes),
			},
		}
		results = append(results, result)
	}
	return results
}

// ReleasersAPI returns a list of all releasers/groups with pagination.
func ReleasersAPI(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	const msg = "releasers api"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	// Parse page parameter, default to page 1
	page := 1
	pageStr := c.QueryParam("page")
	if pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid page parameter",
			})
		}
	}

	const limit = 1000 // Fixed limit per page
	ctx := context.Background()

	// Get total count for pagination efficiently (without loading all data)
	totalReleasers, err := getReleasersCount(ctx, db)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to count releasers",
		})
	}

	totalPages := (totalReleasers + limit - 1) / limit // Ceiling division

	// Get only the releasers needed for the current page
	paginatedReleasers := model.Releasers{}
	if err := paginatedReleasers.Limit(ctx, db, model.Alphabetical, limit, page); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to query releasers",
		})
	}

	// Check if no releasers were returned (page out of range)
	if len(paginatedReleasers) == 0 {
		return c.JSON(http.StatusOK, map[string]any{
			"releasers":  []ReleaserAPI{},
			"page":       page,
			"totalPages": totalPages,
		})
	}

	// Only load statistics for the releasers on this page
	releasersWithStats := getReleasersWithStats(paginatedReleasers)

	return c.JSON(http.StatusOK, map[string]any{
		"releasers":  releasersWithStats,
		"page":       page,
		"totalPages": totalPages,
	})
}

// ReleaserDetailAPI returns details for a specific releaser/group.
func ReleaserDetailAPI(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	const msg = "releaser detail api"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	uri := c.Param("id")
	if uri == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Releaser ID parameter is required",
		})
	}

	ctx := context.Background()

	// Get the releaser info
	r := model.Releasers{}
	files, err := r.Where(ctx, db, uri)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Releaser not found",
		})
	}
	if len(files) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Releaser not found",
		})
	}

	// Get statistics for this releaser
	m := model.Summary{ //nolint:exhaustruct // Fields are set by ByReleaser method
	}
	if err := m.ByReleaser(ctx, db, uri); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get releaser statistics",
		})
	}

	// Convert files to API format
	apiFiles := make([]AnnouncementFile, 0, len(files))
	for _, f := range files {
		// Get tags for the file
		category := filerecord.TagCategory(f)
		platform := filerecord.TagProgram(f)
		
		// Get humanized description for the tags
		categoryTag := tags.TagByURI(category)
		platformTag := tags.TagByURI(platform)
		humanized := tags.Humanize(platformTag, categoryTag)

		// Create a file record for the description function
		fileRecord := &models.File{ //nolint:exhaustruct
			Filename:       f.Filename,
			RecordTitle:    f.RecordTitle,
			GroupBrandBy:   f.GroupBrandBy,
			GroupBrandFor:  f.GroupBrandFor,
			DateIssuedYear: f.DateIssuedYear,
		}

		apiFile := AnnouncementFile{
			ID:          f.ID,
			Filename:    f.Filename.String,
			Size:        struct{ Formatted string `json:"formatted"`; Bytes int64 `json:"bytes"` }{Formatted: helper.ByteCount(f.Filesize.Int64), Bytes: f.Filesize.Int64},
			Description: filerecord.Description(fileRecord),
			Tags: struct {
				Category    string `json:"category"`
				Platform    string `json:"platform"`
				Description string `json:"description"`
			}{
				Category:    category,
				Platform:    platform,
				Description: humanized,
			},
			URLs: struct {
				Download  string `json:"download"`
				HTML      string `json:"html"`
				Thumbnail string `json:"thumbnail,omitempty"`
			}{
				Download:  "/d/" + helper.ObfuscateID(f.ID),
				HTML:      "/f/" + helper.ObfuscateID(f.ID),
				Thumbnail: "/public/image/thumb/" + f.UUID.String,
			},
		}

		// Set dates if available
		if f.DateIssuedYear.Valid {
			apiFile.DatePublished = struct{ Year int16 `json:"year,omitempty"`; Month int16 `json:"month,omitempty"`; Day int16 `json:"day,omitempty"` }{
				Year:  f.DateIssuedYear.Int16,
				Month: f.DateIssuedMonth.Int16,
				Day:   f.DateIssuedDay.Int16,
			}
		}
		if f.Createdat.Valid {
			apiFile.PostedDate = &f.Createdat.Time
		}

		apiFiles = append(apiFiles, apiFile)
	}

	relname := releaser.Link(uri)
	stableID := hashString(uri)

	// Return response with both group info and files
	return c.JSON(http.StatusOK, map[string]any{
		"group": ReleaserAPI{
			ID:    stableID,
			Name:  uri,
			Title: relname,
			URLs: struct {
				API   string `json:"api"`
				HTML3 string `json:"html3"`
				HTML  string `json:"html"`
			}{
				API:   "/api/group/" + uri,
				HTML3: "/html3/group/" + uri,
				HTML:  "/g/" + uri,
			},
			Statistics: struct {
				TotalFiles     int64  `json:"totalFiles"`
				TotalSize      string `json:"totalSize"`
				TotalSizeBytes int64  `json:"totalSizeBytes"`
			}{
				TotalFiles:     m.SumCount.Int64,
				TotalSize:      helper.ByteCount(m.SumBytes.Int64),
				TotalSizeBytes: m.SumBytes.Int64,
			},
		},
		"files": apiFiles,
	})
}
