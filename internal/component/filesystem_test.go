package component

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/ext"
)

func TestLoadFileSystems(t *testing.T) {
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

	disk := ext.MockDisk{
		Dirs: []string{
			"filesystem",
			"filesystem/root.all",
			"filesystem/root.amd64",
			"filesystem/root.amd64-ubuntu",
		},
	}

	actual, err := loadFileSystems(targets, &disk)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string][]*FileSystem{
		"root": {
			{
				ID:     "root.all",
				Anchor: "root",
				Target: internal.Target{Name: "all", All: true},
			},
			{
				ID:     "root.amd64",
				Anchor: "root",
				Target: internal.Target{Name: "amd64", Architecture: internal.AMD64},
			},
			{
				ID:     "root.amd64-ubuntu",
				Anchor: "root",
				Target: internal.Target{Name: "amd64-ubuntu", Architecture: internal.AMD64, OSReleaseID: "ubuntu"},
			},
		},
	}

	if diff := cmp.Diff(actual, expected, cmpopts.SortSlices(sortByTarget)); diff != "" {
		t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
	}

}
