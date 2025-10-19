package component

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/woolawin/catalogue/internal"
)

func TestLoadTargets(t *testing.T) {
	deserialized := map[string]TargetTOML{
		"ubuntu": {
			Architecture: "    ",
			OSReleaseID:  "  ubuntu  ",
		},
		"ubuntu22": {
			OSReleaseID:              "  ubuntu  ",
			OSReleaseVersion:         "22.04",
			OSReleaseVersionID:       "  22",
			OSReleaseVersionCodeName: "Jammy",
		},
	}

	actual, err := loadTargets(deserialized)
	if err != nil {
		t.Fatal(actual)
	}

	expected := []internal.Target{
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
		{
			Name:                     "ubuntu22",
			OSReleaseID:              "ubuntu",
			OSReleaseVersion:         "22.04",
			OSReleaseVersionID:       "22",
			OSReleaseVersionCodeName: "Jammy",
		},
	}

	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
	}
}
