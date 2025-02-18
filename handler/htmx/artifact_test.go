package htmx

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func newContext() echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr error
	}{
		{
			name:    "absolute path",
			path:    "/absolute/path",
			wantErr: ErrPath,
		},
		{
			name:    "clean path",
			path:    "relative/path",
			wantErr: nil,
		},
		{
			name:    "clean path",
			path:    "relative/path/",
			wantErr: nil,
		},
		{
			name:    "unclean path 1",
			path:    "relative/../path",
			wantErr: ErrPath,
		},
		{
			name:    "unclean path 2",
			path:    "./relative/path",
			wantErr: ErrPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.path)
			if err != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && tt.wantErr != nil {
				t.Errorf("Validate() expected error = %v, got nil", tt.wantErr)
			}
		})
	}
}

func TestPath(t *testing.T) {
	tests := []struct {
		name     string
		unid     string
		path     string
		wantUnid string
		wantName string
		wantErr  bool
	}{
		{
			name:     "valid unid and path",
			unid:     "123e4567-e89b-12d3-a456-426614174000",
			path:     "relative/path",
			wantUnid: "123e4567-e89b-12d3-a456-426614174000",
			wantName: "relative/path",
			wantErr:  false,
		},
		{
			name:     "invalid unid",
			unid:     "invalid-unid",
			path:     "relative/path",
			wantUnid: "",
			wantName: "",
			wantErr:  true,
		},
		{
			name:     "invalid path",
			unid:     "123e4567-e89b-12d3-a456-426614174000",
			path:     "/absolute/path",
			wantUnid: "",
			wantName: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newContext()
			c.SetParamNames("unid", "path")
			c.SetParamValues(tt.unid, url.QueryEscape(tt.path))

			gotUnid, gotName, err := Path(c)
			got := (err != nil)
			if !assert.Equal(t, tt.wantErr, got) {
				t.Errorf("Path() error = %v, wantErr %v", got, tt.wantErr)
			}
			assert.Equal(t, tt.wantUnid, gotUnid)
			assert.Equal(t, tt.wantName, gotName)
		})
	}
}
