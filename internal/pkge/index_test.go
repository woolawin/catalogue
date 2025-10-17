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

func TestDeserializeMetaTarget(t *testing.T) {
	input := `
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
	}

	if diff := cmp.Diff(actual, expected); diff != "" {
		fmt.Printf("Mismatch (-actual +expected):\n%s", diff)
	}

}

func TestMergeMeta(t *testing.T) {
	targets := []target.Target{
		{Name: "arm64", Architecture: target.ARM64},
		{Name: "amd64", Architecture: target.AMD64},
		{Name: "all", All: true},
	}
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

	actual := MergeMeta(&raw, system, targets)
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
	}

	actual.Clean()

	expected := Raw{
		Meta: map[string]Meta{
			"all": Meta{Name: "fooo"},
		},
	}

	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
	}
}
