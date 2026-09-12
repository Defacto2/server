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
	"github.com/labstack/echo/v5/echotest"
)

// Echo package helpers

type Input map[string]string

// Middleware returns an echo response using the target url.
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

// MiddlewareStatus returns an echo response using http request using the target url and http status.
//
// The echo.Echo can be provided using [echo.New].
func MiddlewareStatus(tb testing.TB, e *echo.Echo, target string, status int) *echo.Context {
	tb.Helper()
	if target == "" {
		target = "/" // NOTE: target cannot be empty or a race condition may occur.
	}

	req := httptest.NewRequestWithContext(tb.Context(), http.MethodGet, target, nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.Response().WriteHeader(status)

	return c
}

// NewContext creates a new instance of Echo and returns a response using the target url.
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

// NewFile creates a new Echo instance with an embedded textfile upload.
//
//   - fieldname is the named input of the form containing the upload.
//   - filename is the original filename of the file upload.
//   - The upload contains the following string, "Hello world!".
func NewFile(t *testing.T, target, fieldname, filename string) *echo.Context {
	t.Helper()
	return newEchoTest(t, target, fieldname, filename, nil, nil)
}

// NewFileInputs functions the same as [NewFile] but also allows for additional
// form inputs using [Input], for example:
//
//	Input["id"] = 1
//	Input["uuid"] = 123e4567-e89b-12d3-a456-426614174000
func NewFileInputs(t *testing.T, target, fieldname, filename string, formInputs Input,
) *echo.Context {
	t.Helper()
	return newEchoTest(t, target, fieldname, filename, formInputs, nil)
}

// NewPath creates a new Echo instance with a collection of [echo.PathValues],
// that allow the provision path parameters using Name and Value pairs.
func NewPath(t *testing.T, target string, pathValues echo.PathValues) *echo.Context {
	t.Helper()

	return newEchoTest(t, target, "", "", nil, pathValues)
}

// NewInput creates a new instance of Echo and returns a response using the target url.
// It sets the key and value that are mapped as both query parameters and form values.
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

// NewInputs creates a new Echo instance and returns a response using the target url.
// It accepts an Input map using keys and values that are treated as both query parameters and form values.
func NewInputs(t *testing.T, target string, formInputs Input) *echo.Context {
	t.Helper()

	return newEchoTest(t, target, "", "", formInputs, nil)
}

// NewInputsPath functions as both [NewInputs] and [NewPath].
func NewInputsPath(t *testing.T, target string, formInputs Input, pathValues echo.PathValues) *echo.Context {
	t.Helper()

	return newEchoTest(t, target, "", "", formInputs, pathValues)
}

func newEchoTest(t *testing.T,
	target, fieldname, filename string, formInputs Input, pathValues echo.PathValues,
) *echo.Context {
	t.Helper()
	if target == "" {
		target = "/"
	}

	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	// must be done before closing
	if size := len(formInputs); size > 0 {
		for fieldname, value := range formInputs {
			err := w.WriteField(fieldname, value)
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	if fieldname != "" && filename != "" {
		p := []byte("Hello world!")
		part, err := w.CreateFormFile(fieldname, filename)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = part.Write(p); err != nil {
			t.Fatal(err)
		}
	}

	// closing the file/form writer must always be done last
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, target, &body)
	r.Header.Set(echo.HeaderContentType, w.FormDataContentType())

	c, _ := echotest.ContextConfig{}.ToContextRecorder(t)
	if len(pathValues) > 0 {
		c.SetPathValues(pathValues)
	}
	c.SetRequest(r)

	return c
}
