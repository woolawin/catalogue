package component

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/woolawin/catalogue/internal/target"
)

func TestLoadMetadata(t *testing.T) {

	targets := []target.Target{
		{
			Name:         "amd64",
			Architecture: target.AMD64,
		},
		{
			Name:         "arm64",
			Architecture: target.ARM64,
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
			Dependencies:    []string{"foo ", " bar "},
			Section:         "  other ",
			Priority:        "  normal  ",
			Homepage:        "   https://foo.com/bar   ",
			Maintainer:      " me  ",
			Description:     "  Foo Bar ",
			Architecture:    " all ",
			Recommendations: []string{"  baz "},
		},
		"ubuntu": {
			Maintainer: " canonical ",
		},
		"arm64-ubuntu": {
			Homepage:        "https://arm.com/foo",
			Recommendations: []string{"driver"},
		},
	}

	actual, err := loadMetadata(deserialized, targets)
	if err != nil {
		t.Fatal(err)
	}

	expected := []*Metadata{
		{
			Target:     target.Target{Name: "ubuntu", OSReleaseID: "ubuntu"},
			Maintainer: "canonical",
		},
		{
			Target:          target.Target{Name: "all", All: true},
			Dependencies:    []string{"foo", "bar"},
			Section:         "other",
			Priority:        "normal",
			Homepage:        "https://foo.com/bar",
			Maintainer:      "me",
			Description:     "Foo Bar",
			Architecture:    "all",
			Recommendations: []string{"baz"},
		},
		{
			Target: target.Target{
				Name:         "arm64-ubuntu",
				Architecture: target.ARM64,
				OSReleaseID:  "ubuntu",
			},
			Homepage:        "https://arm.com/foo",
			Recommendations: []string{"driver"},
		},
	}

	if diff := cmp.Diff(actual, expected, cmpopts.SortSlices(sortByTarget)); diff != "" {
		t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
	}
}
func sortByTarget(a, b target.GetTarget) int {
	return strings.Compare(a.GetTarget().Name, b.GetTarget().Name)
}
