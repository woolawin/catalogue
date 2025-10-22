package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReadRecord(t *testing.T) {
	value := `
[origin]
type='git'
url='https://github.com/foo/bar.git'
`

	actual, err := ReadRecord(strings.NewReader(value))
	if err != nil {
		t.Fatal(err)
	}

	expected := Record{
		Origin: Origin{
			Type: Git,
			URL:  u("https://github.com/foo/bar.git"),
		},
	}

	if diff := cmp.Diff(actual, expected); diff != "" {
		fmt.Printf("Mismatch (-actual +expected):\n%s", diff)
	}

}
