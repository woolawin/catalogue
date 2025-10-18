package pkge

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/woolawin/catalogue/internal/target"
)

func TestBasicDeserialize(t *testing.T) {
	input := `
[meta.all]
name='FooBar'
dependencies=['foo', 'bar']
section='utilities'
priority='normal'
homepage='https://foobar.com'
description='foo bar'
maintainer='Bob Doe'
architecture='amd64'
	`

	actual, err := deserialize(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	expected := Raw{
		Meta: map[string]MetadataTOML{
			"all": {
				Name:         "FooBar",
				Dependencies: []string{"foo", "bar"},
				Section:      "utilities",
				Priority:     "normal",
				Homepage:     "https://foobar.com",
				Description:  "foo bar",
				Maintainer:   "Bob Doe",
				Architecture: "amd64",
			},
		},
	}

	if diff := cmp.Diff(actual, expected); diff != "" {
		fmt.Printf("Mismatch (-actual +expected):\n%s", diff)
	}

}

func TestDeserializeFull(t *testing.T) {
	input := `

[target.ubuntu]
os_release_id='ubuntu'

[meta.all]
name='FooBar'
dependencies=['foo', 'bar']
section='utilities'
priority='normal'
homepage='https://foobar.com'
description='foo bar'
maintainer='Bob Doe'

[meta.amd64]
architecture='amd64'
maintainer='Jane Doe'

[download.bin.all]
src="https://foo.com/bin"
dst="path://root/usr/bin"
`

	actual, err := deserialize(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	expected := Raw{
		Meta: map[string]MetadataTOML{
			"all": {
				Name:         "FooBar",
				Dependencies: []string{"foo", "bar"},
				Section:      "utilities",
				Priority:     "normal",
				Homepage:     "https://foobar.com",
				Description:  "foo bar",
				Maintainer:   "Bob Doe",
			},
			"amd64": {
				Architecture: "amd64",
				Maintainer:   "Jane Doe",
			},
		},
		Target: map[string]TargetTOML{
			"ubuntu": {
				OSReleaseID: "ubuntu",
			},
		},
		Download: map[string]map[string]RawDownload{
			"bin": {
				"all": {
					Source:      "https://foo.com/bin",
					Destination: "path://root/usr/bin",
				},
			},
		},
	}

	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
	}

}

func TestConstruct(t *testing.T) {
	raw := Raw{
		Target: map[string]TargetTOML{
			"ubuntu": {
				Architecture:             "amd64",
				OSReleaseID:              "ubuntu",
				OSReleaseVersion:         "22",
				OSReleaseVersionID:       "22.04",
				OSReleaseVersionCodeName: "cody cod",
			},
		},
		Meta: map[string]MetadataTOML{
			"all": {
				Name:            "foo",
				Dependencies:    []string{"bar", "baz"},
				Section:         "other",
				Priority:        "normal",
				Homepage:        "https://foo.com",
				Maintainer:      "me",
				Description:     "example",
				Architecture:    "amd64",
				Recommendations: []string{"baz"},
			},
		},
	}
	actual, err := construct(&raw)
	if err != nil {
		t.Fatal(err)
	}

	expected := Index{
		Targets: []target.Target{
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
				Name:                     "ubuntu",
				All:                      false,
				Architecture:             "amd64",
				OSReleaseID:              "ubuntu",
				OSReleaseVersion:         "22",
				OSReleaseVersionID:       "22.04",
				OSReleaseVersionCodeName: "cody cod",
			},
		},
		Metadata: []*Metadata{
			{
				Target:          target.Target{Name: "all", All: true},
				Name:            "foo",
				Dependencies:    []string{"bar", "baz"},
				Section:         "other",
				Priority:        "normal",
				Homepage:        "https://foo.com",
				Maintainer:      "me",
				Description:     "example",
				Architecture:    "amd64",
				Recommendations: []string{"baz"},
			},
		},
	}

	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
	}
}

func TestNormalizeList(t *testing.T) {
	actual := []string{"foo", "bar"}
	actual = normalizeList(actual)
	expected := []string{"foo", "bar"}

	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
	}

	actual = []string{" foo ", " bar "}
	actual = normalizeList(actual)
	expected = []string{"foo", "bar"}

	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
	}

	actual = []string{" foo ", " ", "baz"}
	actual = normalizeList(actual)
	expected = []string{"foo", "baz"}

	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
	}

	actual = []string{"  ", " "}
	actual = normalizeList(actual)
	expected = nil

	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
	}
}
