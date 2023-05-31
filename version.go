package version

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strconv"
	"strings"

	version "github.com/Masterminds/semver/v3"
)

type constraintHandler struct {
	cs *version.Constraints
	h  http.Handler
}

func (c constraintHandler) String() string {
	return c.cs.String()
}

type Map map[string]http.Handler

// https://restfulapi.net/content-negotiation/
// https://restfulapi.net/versioning/
func Version(vendor string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accept := r.Header.Get("Accept")

			ver, err := parseVersion(accept, vendor)
			if err != nil {
				// not acceptable
				http.Error(w, "not acceptable", http.StatusNotAcceptable)
				return
			}

			ctx := context.WithValue(r.Context(), apiVersion, ver)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Match(cm Map) http.Handler {
	cs, err := build(cm)
	if err != nil {
		panic(err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ver, ok := r.Context().Value(apiVersion).(*version.Version)
		if !ok {
			// bad request, not really
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		h, ok := match(cs, ver)
		if !ok {
			// not found
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		// This should always be valid
		major := strconv.Itoa(int(ver.Major()))
		w.Header().Set("X-API-Version", major)

		h.ServeHTTP(w, r)
	})
}

func parseVersion(mediaTyp, vendor string) (*version.Version, error) {
	mediaTyp, params, err := mime.ParseMediaType(mediaTyp)
	if err != nil {
		return nil, err
	}

	if !strings.Contains(mediaTyp, vendor) {
		return nil, errors.New("not a valid media type")
	}

	v, ok := params["version"]
	if !ok {
		return nil, errors.New("no version")
	}

	return version.NewVersion(v)
}

func build(cm Map) ([]constraintHandler, error) {
	var cs []constraintHandler
	for s, h := range cm {
		cc, err := version.NewConstraint(s)
		if err != nil {
			return nil, err
		}

		cs = append(cs, constraintHandler{cc, h})
	}

	return cs, nil
}

func match(cs []constraintHandler, ver *version.Version) (http.Handler, bool) {
	for _, c := range cs {
		if c.cs.Check(ver) {
			return c.h, true
		}
	}

	return nil, false
}

type contextKey string

const (
	apiVersion contextKey = "version"
)

func (c contextKey) String() string {
	return "context key " + string(c)
}
