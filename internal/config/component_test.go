package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBasicDeserialize(t *testing.T) {
	input := `
name='  foobar '
type='  package  '

[metadata.all]
dependencies='foo, bar'
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

	expected := ComponentTOML{
		Name: "  foobar ",
		Type: "  package  ",
		Metadata: map[string]MetadataTOML{
			"all": {
				Dependencies: "foo, bar",
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
name='foobar'
type='package'
supported_targets=['foo', 'bar']

[target.ubuntu]
os_release_id='ubuntu'

[metadata.all]
dependencies='foo,bar'
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

	expected := ComponentTOML{
		Name:             "foobar",
		Type:             "package",
		SupportedTargets: []string{"foo", "bar"},
		Metadata: map[string]MetadataTOML{
			"all": {
				Dependencies: "foo,bar",
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
