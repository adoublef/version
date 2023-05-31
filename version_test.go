package version

import (
	"testing"

	version "github.com/Masterminds/semver/v3"
	is "github.com/stretchr/testify/require"
)

func TestCheckVersion(t *testing.T) {
	vm := Map{
		">=1": nil,
		"2":   nil,
		"3":   nil,
	}

	var cs []constraintHandler
	for s, h := range vm {
		cc, err := version.NewConstraint(s)
		is.NoError(t, err)

		cs = append(cs, constraintHandler{cc, h})
	}

	// called during mux.Mount

	acceptHeader := "application/vnd.api+json; version=1.2"
	// called during mux.Use
	ver, err := parseVersion(acceptHeader, "vnd.api+json")
	if err != nil {
		t.Fatalf("parseVersion: %v", err)
	}

	// called during mux.Mount
	_, ok := match(cs, ver)
	if !ok {
		t.Fatalf("match: %v", err)
	}
}

func TestParseVersion(t *testing.T) {
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
			ver, err := parseVersion(tc.media, "vnd.api+json")
			if err != nil {
				t.Fatalf("parseVersion: %v", err)
			}

			if tc.version != ver.String() {
				t.Fatalf("expected %s, got %s", tc.version, ver.String())
			}
		})
	}
}
