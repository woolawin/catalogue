package config

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/woolawin/catalogue/internal"
)

func TestLoadTargetMetadata(t *testing.T) {

	targets := []internal.Target{
		{
			Name:         "amd64",
			Architecture: internal.AMD64,
		},
		{
			Name:         "arm64",
			Architecture: internal.ARM64,
		},
		{
			Name: "all",
			All:  true,
		},
		{
			Name:        "ubuntu",
			OSReleaseID: "ubuntu",
		},
	}

	deserialized := map[string]MetadataTOML{
		"all": {
			Dependencies: " foo,bar ",
			Category:     "  other ",
			Homepage:     "   https://foo.com/bar   ",
			Maintainer:   " me  ",
			Description:  "  Foo Bar ",
			Architecture: " all ",
		},
		"ubuntu": {
			Maintainer: " canonical ",
		},
		"arm64-ubuntu": {
			Homepage: "https://arm.com/foo",
		},
	}

	actual, err := loadTargetMetadata(deserialized, targets)
	if err != nil {
		t.Fatal(err)
	}

	expected := []*TargetMetadata{
		{
			Target:   internal.Target{Name: "ubuntu", OSReleaseID: "ubuntu"},
			Metadata: Metadata{Maintainer: "canonical"},
		},
		{
			Target: internal.Target{Name: "all", All: true},
			Metadata: Metadata{
				Dependencies: "foo,bar",
				Category:     "other",
				Homepage:     "https://foo.com/bar",
				Maintainer:   "me",
				Description:  "Foo Bar",
				Architecture: "all",
			},
		},
		{
			Target: internal.Target{
				Name:         "arm64-ubuntu",
				Architecture: internal.ARM64,
				OSReleaseID:  "ubuntu",
			},
			Metadata: Metadata{
				Homepage: "https://arm.com/foo",
			},
		},
	}

	if diff := cmp.Diff(actual, expected, cmpopts.SortSlices(sortByTarget)); diff != "" {
		t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
	}
}
func sortByTarget(a, b internal.GetTarget) int {
	return strings.Compare(a.GetTarget().Name, b.GetTarget().Name)
}
