package version

import (
	"net/http"
	"testing"

	is "github.com/stretchr/testify/require"
)



func TestCheckVersion(t *testing.T) {
	vm := Map{
		">=1": http.NotFoundHandler(),
		"2":   http.NotFoundHandler(),
		"3":   http.NotFoundHandler(),
	}

	// called during mux.Mount
	cs, err := vm.Slice()
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
	vm := Map{
		">=1":   http.NotFoundHandler(),
		"2":  http.NotFoundHandler(),
	}

	cs, err := vm.Slice()
	is.NoError(t, err)

	// first is the highest version
	is.Equal(t, "2", cs[0].String())
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
