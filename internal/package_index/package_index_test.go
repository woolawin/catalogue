package internal

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBasicMeta(t *testing.T) {
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

	actual, err := ReadPackageIndex(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	expected := PackageIndex{
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

func TestTargetAndAll(t *testing.T) {
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

	actual, err := ReadPackageIndex(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	expected := PackageIndex{
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
