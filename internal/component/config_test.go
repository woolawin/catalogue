package component

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBasicDeserialize(t *testing.T) {
	input := `
[metadata.all]
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

	expected := ConfigTOML{
		Metadata: map[string]MetadataTOML{
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

[metadata.all]
name='FooBar'
dependencies=['foo', 'bar']
section='utilities'
priority='normal'
homepage='https://foobar.com'
description='foo bar'
maintainer='Bob Doe'

[metadata.amd64]
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

	expected := ConfigTOML{
		Metadata: map[string]MetadataTOML{
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
		Download: map[string]map[string]DownloadTOML{
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
