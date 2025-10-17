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
				Maintainer: "Riley Doe",
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
		fmt.Printf("Mismatch (-actual +expected):\n%s", diff)
	}
}
