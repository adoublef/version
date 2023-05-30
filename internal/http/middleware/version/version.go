package version

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"
)

type constraintHandler struct {
	constraints version.Constraints
	handler     http.Handler
}

func (c constraintHandler) String() string {
	return c.constraints.String()
}

type Map map[string]http.Handler

// constraintHandlerSort will sort with the highest version first.
func (cm Map) Slice() (cs []constraintHandler, err error) {
	for c, h := range cm {
		constraints, err := version.NewConstraint(c)
		if err != nil {
			return nil, err
		}
		cs = append(cs, constraintHandler{constraints, h})
	}

	// NOTE may check if two constraints are the same, then return an error

	sort.Slice(cs, func(i, j int) bool {
		return !cs[i].constraints.Equals(cs[j].constraints)
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

// https://restfulapi.net/content-negotiation/
// https://restfulapi.net/versioning/
func Vendor(vendor string) func(h http.Handler) http.Handler {
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

func Match(vm Map) http.Handler {
	cs, err := vm.Slice()
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

		// set the version header
		w.Header().Set("X-API-Version", ver.String())

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
