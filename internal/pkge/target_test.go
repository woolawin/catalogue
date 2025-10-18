package pkge

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/woolawin/catalogue/internal/target"
)

func TestLoadTargets(t *testing.T) {
	raw := map[string]TargetTOML{
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

	actual, err := loadTargets(raw)
	if err != nil {
		t.Fatal(actual)
	}

	expected := []target.Target{
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
