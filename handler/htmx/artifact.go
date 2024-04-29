package htmx

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// badRequest returns a JSON response with a 400 status code,
// the server cannot or will not process the request due to something that is perceived to be a client error.
func badRequest(c echo.Context, err error) error {
	return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request " + err.Error()})
}

// RecordToggle handles the post submission for the File artifact is online and public toggle.
func RecordToggle(c echo.Context, state bool) error {
	return c.String(http.StatusBadRequest, "bad request")
	// key := c.FormValue("artifact-editor-key")
	// fmt.Println("key: ", key)
	// // var f Form
	// // if err := c.Bind(&f); err != nil {
	// // 	return badRequest(c, err)
	// // }
	// // if state {
	// // 	if err := model.UpdateOnline(c, int64(f.ID)); err != nil {
	// // 		return badRequest(c, err)
	// // 	}
	// // 	return c.JSON(http.StatusOK, f)
	// // }
	// // if err := model.UpdateOffline(c, int64(f.ID)); err != nil {
	// // 	return badRequest(c, err)
	// // }
	// return c.String(http.StatusOK, "keyyer: "+key)
}
