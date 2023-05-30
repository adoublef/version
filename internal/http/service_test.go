package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "github.com/adoublef-go/rest-api/internal/http/v1"
	v2 "github.com/adoublef-go/rest-api/internal/http/v2"
	v "github.com/kataras/versioning"
	"github.com/go-chi/chi/v5"
	is "github.com/stretchr/testify/require"
)

func TestVersioning(t *testing.T) {
	// "Accept" header, i.e Accept: "application/json; version=1.0" - "Accept-Version" header, i.e Accept-Version: "1.0"
	// req.Header.Set("Accept", "application/vnd.api.v1+json")
	srv := newTestServer(t)

	t.Cleanup(func() { srv.Close() })

	for _, tc := range []struct {
		name    string
		accept  string
		version string
	}{
		{
			name:    "version 1",
			accept:  "application/json; version=1.0",
			version: "v1",
		},
		{
			name:    "version 2",
			accept:  "application/json; version=2.0",
			version: "v2",
		},
		// {
		// 	name: "default version",
		// 	version: "v1",
		// },
	} {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, srv.URL+"/", nil)
			is.NoError(t, err)

			req.Header.Set("Accept", tc.accept)
			res, err := srv.Client().Do(req)
			is.NoError(t, err)

			is.Equal(t, http.StatusOK, res.StatusCode)
			defer res.Body.Close()

			// decode the response and get the version from the body
			var body map[string]any
			err = json.NewDecoder(res.Body).Decode(&body)
			is.NoError(t, err)

			// check the version
			is.Equal(t, tc.version, body["version"])
		})
	}

}

// mux.Mount(<prefix>, <version_handler>(map[string|number]http.Handler{"1": <handler>, "2": <handler>})))
func newTestServer(t *testing.T) (srv *httptest.Server) {
	v1, v2 := v1.NewService(), v2.NewService()

	mux := chi.NewMux()
	mux.Mount("/", v.NewMatcher(v.Map{"1": v1, "2": v2}))

	// TODO: middleware to get the version from the header and set it in the context
	// mux.Use(<version_middleware>)
	// TODO: handler to use context and get the version from the context to match the handler
	// mux.Get("/", <version_handler>(map[string]http.Handler{<version_number>:<handler_one>, <version_number>:<handler_two>}))

	srv = httptest.NewServer(mux)
	return
}
