//nolint:exhaustruct_v5
package testutil

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

// Echo package helpers

type Input map[string]string

// Middleware returns an [*echo.Context] for a GET request to target.
// If no target is provided, it is set to root "/".
//
// The echo.Echo can be provided using [echo.New].
func Middleware(tb testing.TB, e *echo.Echo, target string) *echo.Context {
	tb.Helper()
	if target == "" {
		target = "/"
	}

	req := httptest.NewRequestWithContext(tb.Context(), http.MethodGet, target, nil)
	rec := httptest.NewRecorder()

	return e.NewContext(req, rec)
}

// MiddlewareStatus returns an [*echo.Context] for a GET request to target
// with the response status header already written.
//
// The echo.Echo can be provided using [echo.New].
func MiddlewareStatus(tb testing.TB, e *echo.Echo, target string, status int) *echo.Context {
	tb.Helper()
	if target == "" {
		target = "/" // httptest.NewRequest panics on an empty target.
	}

	req := httptest.NewRequestWithContext(tb.Context(), http.MethodGet, target, nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.Response().WriteHeader(status)

	return c
}

// NewContext creates a new instance of Echo and returns an [*echo.Context] for a GET request to target.
// If no target is provided, it is set to root "/". Query params can also be provided
// using target, for example:
//
//	target="/myurl?id=1&name=hello"
func NewContext(tb testing.TB, target string) *echo.Context {
	tb.Helper()
	if target == "" {
		target = "/"
	}

	e := echo.New()
	return Middleware(tb, e, target)
}

// NewFile creates a new Echo context with a multipart POST containing an embedded textfile upload.
//
//   - fieldname is the named input of the form containing the upload.
//   - filename is the original filename of the file upload.
//   - The upload contains the following string, "Hello world!".
func NewFile(tb testing.TB, target, fieldname, filename string) *echo.Context {
	tb.Helper()
	return newMultipartContext(tb, target, fieldname, filename, nil, nil)
}

// NewFileInputs functions the same as [NewFile] but also allows for additional
// form inputs using [Input], for example:
//
//	inputs := testutil.Input{
//		"id":   "1",
//		"uuid": "123e4567-e89b-12d3-a456-426614174000",
//	}
func NewFileInputs(tb testing.TB, target, fieldname, filename string, formInputs Input,
) *echo.Context {
	tb.Helper()
	return newMultipartContext(tb, target, fieldname, filename, formInputs, nil)
}

// NewPath creates a new Echo context for a GET request with a collection of [echo.PathValues],
// providing path parameters using Name and Value pairs.
func NewPath(tb testing.TB, target string, pathValues echo.PathValues) *echo.Context {
	tb.Helper()

	c := NewContext(tb, target)
	if len(pathValues) > 0 {
		c.SetPathValues(pathValues)
	}
	return c
}

// NewInput creates a new instance of Echo and returns an [*echo.Context] for a POST request
// with a single application/x-www-form-urlencoded form key and value.
// If no target is provided, it is set to root "/".
func NewInput(tb testing.TB, target, key, value string) *echo.Context {
	tb.Helper()
	if target == "" {
		target = "/"
	}

	form := url.Values{}
	form.Set(key, value)

	body := strings.NewReader(form.Encode())
	req := httptest.NewRequestWithContext(tb.Context(), http.MethodPost, target, body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	rec := httptest.NewRecorder()
	e := echo.New()
	return e.NewContext(req, rec)
}

// NewInputs creates a new Echo context for a multipart/form-data POST request
// populated with the form fields in formInputs.
func NewInputs(tb testing.TB, target string, formInputs Input) *echo.Context {
	tb.Helper()

	return newMultipartContext(tb, target, "", "", formInputs, nil)
}

// NewInputsPath functions as both [NewInputs] and [NewPath].
func NewInputsPath(tb testing.TB, target string, formInputs Input, pathValues echo.PathValues) *echo.Context {
	tb.Helper()

	return newMultipartContext(tb, target, "", "", formInputs, pathValues)
}

func newMultipartContext(tb testing.TB,
	target, fieldname, filename string, formInputs Input, pathValues echo.PathValues,
) *echo.Context {
	tb.Helper()
	if target == "" {
		target = "/"
	}

	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	// must be done before closing
	for key, value := range formInputs {
		if err := w.WriteField(key, value); err != nil {
			tb.Fatal(err)
		}
	}

	if fieldname != "" && filename != "" {
		p := []byte("Hello world!")
		part, err := w.CreateFormFile(fieldname, filename)
		if err != nil {
			tb.Fatal(err)
		}
		if _, err = part.Write(p); err != nil {
			tb.Fatal(err)
		}
	}

	// closing the file/form writer must always be done last
	if err := w.Close(); err != nil {
		tb.Fatal(err)
	}

	req := httptest.NewRequestWithContext(tb.Context(), http.MethodPost, target, &body)
	req.Header.Set(echo.HeaderContentType, w.FormDataContentType())

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	if len(pathValues) > 0 {
		c.SetPathValues(pathValues)
	}

	return c
}
