package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "example/internal/http/v1"
	v2 "example/internal/http/v2"

	v "github.com/adoublef-go/version"
	"github.com/go-chi/chi/v5"
)

func TestVersioning(t *testing.T) {
	srv := newTestServer(t)

	t.Cleanup(func() { srv.Close() })

	for _, tc := range []struct {
		name    string
		accept  string
		version string
	}{
		{
			name:    "version 1 - good formatted header",
			accept:  "application/vnd.api+json;version=1",
			version: "1",
		},
		{
			name:    "version 1 - ok formatted header",
			accept:  "application/vnd.api+json; version=1.2",
			version: "1",
		},
		{
			name:    "version 2 - poorly formatted header",
			accept:  "application/vnd.api+json ;version=2.0",
			version: "2",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, srv.URL+"/", nil)
			if err != nil {
				t.Fatalf("http.NewRequest: %v", err)
			}

			req.Header.Set("Accept", tc.accept)
			res, err := srv.Client().Do(req)
			if err != nil {
				t.Fatalf("srv.Client().Do: %v", err)
			}

			if res.StatusCode != http.StatusOK {
				t.Fatalf("res.StatusCode: %v", res.StatusCode)
			}

			if tc.version != res.Header.Get("X-API-Version") {
				t.Fatalf("res.Header.Get(X-API-Version): %v", res.Header.Get("X-API-Version"))
			}
		})
	}

}

// mux.Mount(<prefix>, <version_handler>(map[string|number]http.Handler{"1": <handler>, "2": <handler>})))
func newTestServer(t *testing.T) (srv *httptest.Server) {
	v1, v2 := v1.NewService(), v2.NewService()

	mux := chi.NewMux()
	mux.Use(v.Version("vnd.api+json"))
	mux.Mount("/", v.Match(v.Map{"^1": v1, "^2": v2}))

	return httptest.NewServer(mux)
}


