// Copyright © 2023-2026 Ben Garrett. All rights reserved.

package app

import (
	"embed"

	"github.com/labstack/echo/v4"
)

// RegisterOpenAPIRoutes sets up the OpenAPI documentation routes.
func RegisterOpenAPIRoutes(e *echo.Echo, public embed.FS) {
	// Serve OpenAPI specification from embedded public directory
	e.FileFS("/api/openapi.json", "public/json/openapi.json", public)
}
