package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReadRecord(t *testing.T) {
	value := `
name='Foo Bar'

[latest_pin]
version_name='v0.54.2'
commit_hash='c7t43c374c34yh43fc43'

[metadata]
dependencies='foo,bar'
section='utilities'
priority='normal'
homepage='https://foobar.com'
description='foo bar'
maintainer='Bob Doe'
architecture='amd64'

[remote]
protocol='git'
url='https://github.com/foo/bar.git'
`

	actual, err := DeserializeRecord(strings.NewReader(value))
	if err != nil {
		t.Fatal(err)
	}

	expected := Record{
		Name:      "Foo Bar",
		LatestPin: Pin{VersionName: "v0.54.2", CommitHash: "c7t43c374c34yh43fc43"},
		Remote: Remote{
			Protocol: Git,
			URL:      u("https://github.com/foo/bar.git"),
		},
		Metadata: Metadata{
			Dependencies: "foo,bar",
			Section:      "utilities",
			Priority:     "normal",
			Homepage:     "https://foobar.com",
			Description:  "foo bar",
			Maintainer:   "Bob Doe",
			Architecture: "amd64",
		},
	}

	if diff := cmp.Diff(actual, expected); diff != "" {
		fmt.Printf("Mismatch (-actual +expected):\n%s", diff)
	}

}
