package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	v1 "github.com/adoublef-go/rest-api/internal/http/v1"
	v2 "github.com/adoublef-go/rest-api/internal/http/v2"
	"github.com/go-chi/chi/v5"
	"github.com/hashicorp/go-version"
	v "github.com/kataras/versioning"
	is "github.com/stretchr/testify/require"
)

type constraintHandler struct {
	constraints version.Constraints
	handler     http.Handler
}

func (c constraintHandler) String() string {
	return c.constraints.String()
}

// constraintHandlerSort will sort with the highest version first.
func constraintHandlerSort(cm map[string]http.Handler) (cs []constraintHandler, err error) {
	for c, h := range cm {
		constraints, err := version.NewConstraint(c)
		if err != nil {
			return nil, err
		}
		cs = append(cs, constraintHandler{constraints, h})
	}

	// NOTE may check if two constraints are the same, then return an error

	sort.Slice(cs, func(i, j int) bool {
		return cs[i].constraints.String() > cs[j].constraints.String()
	})

	return cs, nil
}

func getMediaTypeVersion(mediaType, vendor string) (*version.Version, error) {
	mediaType, params, err := mime.ParseMediaType(mediaType)
	if err != nil {
		return nil, err
	}

	// TODO make `vnd.api+json` a function argument
	if !strings.Contains(mediaType, vendor) {
		return nil, errors.New("not a valid media type")
	}

	v, ok := params["version"]
	if !ok {
		return nil, errors.New("no version")
	}

	return version.NewVersion(v)
}

func matchVer(cs []constraintHandler, ver *version.Version) (http.Handler, bool) {
	for _, c := range cs {
		if c.constraints.Check(ver) {
			return c.handler, true
		}
	}

	return nil, false
}

func TestCheckVersion(t *testing.T) {
	vm := map[string]http.Handler{
		">=1": http.NotFoundHandler(),
		"2":   http.NotFoundHandler(),
		"3":   http.NotFoundHandler(),
	}

	// called during mux.Mount
	cs, err := constraintHandlerSort(vm)
	is.NoError(t, err)

	acceptHeader := "application/vnd.api+json; version=1.2"
	// called during mux.Use
	ver, err := getMediaTypeVersion(acceptHeader, "vnd.api+json")
	is.NoError(t, err)

	// called during mux.Mount
	_, ok := matchVer(cs, ver)
	is.True(t, ok)
}

func TestAPIConstraints(t *testing.T) {
	vm := map[string]http.Handler{
		">=3": http.NotFoundHandler(),
		"1":   http.NotFoundHandler(),
		"<2":  http.NotFoundHandler(),
	}

	cs, err := constraintHandlerSort(vm)
	is.NoError(t, err)

	for _, c := range cs {
		t.Log(c.String())
	}
}

func TestMIMEType(t *testing.T) {
	type testcase struct {
		name    string
		media   string
		version string
	}

	for _, tc := range []testcase{
		{
			name:    "version 1",
			media:   "application/vnd.api+json; version=1.0",
			version: "1.0.0",
		},
		{
			name:    "version 2",
			media:   "application/vnd.api+json ; version=2.0",
			version: "2.0.0",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ver, err := getMediaTypeVersion(tc.media, "vnd.api+json")
			is.NoError(t, err)

			is.Equal(t, tc.version, ver.String())
		})
	}
}

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
			accept:  "application/vnd.api+json;version=1.0",
			version: "v1",
		},
		{
			name:    "version 1",
			accept:  "application/vnd.api+json; version=1.0",
			version: "v1",
		},
		{
			name:    "version 2",
			accept:  "application/vnd.api+json ;version=2.0",
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
	mux.Use(setter("vnd.api+json"))
	mux.Mount("/", getter(v.Map{"1": v1, "2": v2}))

	return httptest.NewServer(mux)
}

// https://restfulapi.net/content-negotiation/
// https://restfulapi.net/versioning/
func setter(vendor string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accept := r.Header.Get("Accept")

			ver, err := getMediaTypeVersion(accept, vendor)
			if err != nil {
				// not acceptable
				http.Error(w, "not acceptable", http.StatusNotAcceptable)
				return
			}

			ctx := context.WithValue(r.Context(), keyVersion, ver)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getter(vm map[string]http.Handler) http.Handler {
	cs, err := constraintHandlerSort(vm)
	if err != nil {
		panic(err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ver, ok := r.Context().Value(keyVersion).(*version.Version)
		if !ok {
			// bad request, not really
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		h, ok := matchVer(cs, ver)
		if !ok {
			// not found
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		h.ServeHTTP(w, r)
	})
}

type contextKey string

var (
	keyVersion contextKey = "version"
)

func (c contextKey) String() string {
	return "context key " + string(c)
}

