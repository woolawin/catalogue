package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReadRecord(t *testing.T) {
	value := `
latest_known_version='v0.54.2'

[remote]
protocol='git'
url='https://github.com/foo/bar.git'
`

	actual, err := DeserializeRecord(strings.NewReader(value))
	if err != nil {
		t.Fatal(err)
	}

	expected := Record{
		LatestKnownVersion: "v0.54.2",
		Remote: Remote{
			Protocol: Git,
			URL:      u("https://github.com/foo/bar.git"),
		},
	}

	if diff := cmp.Diff(actual, expected); diff != "" {
		fmt.Printf("Mismatch (-actual +expected):\n%s", diff)
	}

}
