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
		Meta: map[string]Meta{
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
		Meta: map[string]Meta{
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
		Target: map[string]RawTarget{
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
		Target: map[string]RawTarget{
			"ubuntu": {
				Architecture:             "amd64",
				OSReleaseID:              "ubuntu",
				OSReleaseVersion:         "22",
				OSReleaseVersionID:       "22.04",
				OSReleaseVersionCodeName: "cody cod",
			},
		},
		Meta: map[string]Meta{
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
	system := target.System{Architecture: target.AMD64}
	actual, err := construct(&raw, system)
	if err != nil {
		t.Fatal(err)
	}

	expected := Index{
		Targets: []target.Target{
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
		Meta: Meta{
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
	}

	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
	}
}

func TestMergeMeta(t *testing.T) {
	system := target.System{Architecture: target.AMD64}

	raw := Raw{
		Meta: map[string]Meta{
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
			"arm64": {
				Recommendations: []string{"happy", "puppy"},
			},
		},
	}

	actual, err := mergeMeta(&raw, system, target.NewRegistry(nil))
	if err != nil {
		t.Fatal(err)
	}
	expected := Meta{
		Name:         "FooBar",
		Dependencies: []string{"foo", "bar"},
		Section:      "utilities",
		Priority:     "normal",
		Homepage:     "https://foobar.com",
		Description:  "foo bar",
		Maintainer:   "Jane Doe",
		Architecture: "amd64",
	}

	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
	}
}

func TestCleanString(t *testing.T) {
	actual := "   foo   "
	cleanString(&actual)
	expected := "foo"

	if actual != expected {
		t.Fatalf("expected '%s' to be '%s'\n", actual, expected)
	}

	actual = "foo"
	cleanString(&actual)
	expected = "foo"

	if actual != expected {
		t.Fatalf("expected '%s' to be '%s'\n", actual, expected)
	}

	actual = ""
	cleanString(&actual)
	expected = ""

	if actual != expected {
		t.Fatalf("expected '%s' to be '%s'\n", actual, expected)
	}
}

func TestCleanList(t *testing.T) {
	actual := []string{"foo", "bar"}
	cleanList(&actual)
	expected := []string{"foo", "bar"}

	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
	}

	actual = []string{" foo ", " bar "}
	cleanList(&actual)
	expected = []string{"foo", "bar"}

	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
	}

	actual = []string{" foo ", " ", "baz"}
	cleanList(&actual)
	expected = []string{"foo", "baz"}

	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
	}

	actual = []string{"  ", " "}
	cleanList(&actual)
	expected = nil

	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
	}
}

func TestCleanRaw(t *testing.T) {
	actual := Raw{
		Meta: map[string]Meta{
			"all": Meta{Name: " fooo  "},
		},
		Target: map[string]RawTarget{
			"mine": {
				Architecture:             "  arch ",
				OSReleaseID:              "   id  ",
				OSReleaseVersion:         "  ver   ",
				OSReleaseVersionID:       "   ver_id   ",
				OSReleaseVersionCodeName: "   ",
			},
		},
		Download: map[string]map[string]RawDownload{
			"bin": {
				"all": {
					Source:      "   https://foo.com/bin  ",
					Destination: "  path://root/usr/bin   ",
				},
			},
		},
	}

	actual.Clean()

	expected := Raw{
		Meta: map[string]Meta{
			"all": Meta{Name: "fooo"},
		},
		Target: map[string]RawTarget{
			"mine": {
				Architecture:             "arch",
				OSReleaseID:              "id",
				OSReleaseVersion:         "ver",
				OSReleaseVersionID:       "ver_id",
				OSReleaseVersionCodeName: "",
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
