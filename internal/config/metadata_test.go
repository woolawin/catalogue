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

type DoNothingLogger struct {
}

func (log *DoNothingLogger) Log(stmt *internal.LogStatement) {

}

func TestMergeMeta(t *testing.T) {
	log := internal.NewLog(&DoNothingLogger{})
	system := internal.System{Architecture: internal.AMD64}

	metadatas := []*TargetMetadata{
		{
			Target: internal.Target{Name: "all", All: true},
			Metadata: Metadata{
				Dependencies: "foo,bar",
				Category:     "utilities",
				Homepage:     "https://foobar.com",
				Description:  "foo bar",
				Maintainer:   "Bob Doe",
			},
		},
		{
			Target: internal.Target{Name: "amd64", Architecture: internal.AMD64},
			Metadata: Metadata{
				Architecture: "amd64",
				Maintainer:   "Jane Doe",
			},
		},
		{
			Target: internal.Target{Name: "arm64", Architecture: internal.ARM64},
			Metadata: Metadata{
				Homepage: "https://foobar.com/amd64",
			},
		},
	}

	actual, err := BuildMetadata(metadatas, log, system)
	if err != nil {
		t.Fatal(err)
	}
	expected := Metadata{
		Dependencies: "foo,bar",
		Category:     "utilities",
		Homepage:     "https://foobar.com",
		Description:  "foo bar",
		Maintainer:   "Jane Doe",
		Architecture: "amd64",
	}

	if diff := cmp.Diff(actual.Metadata, expected); diff != "" {
		t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
	}
}
