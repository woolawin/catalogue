package internal

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestIsAPTTarget(t *testing.T) {
	if !isAPTTarget("foobar") {
		t.Fatal("expected TRUE")
	}

	if !isAPTTarget("foo_bar") {
		t.Fatal("expected TRUE")
	}

	if !isAPTTarget("foo-bar") {
		t.Fatal("expected TRUE")
	}

	if !isAPTTarget("foo+bar") {
		t.Fatal("expected TRUE")
	}
}

func TestGithubTarget(t *testing.T) {
	if !isGithubTarget("github/foo/bar") {
		t.Fatal("expected TRUE")
	}

	if !isGithubTarget("github/foo_bar/baz_doh") {
		t.Fatal("expected TRUE")
	}
}

func TestParseTarget(t *testing.T) {
	value := "foo"
	actual, _ := ParseTarget(value)
	expected := Target{APT: &value}

	if diff := cmp.Diff(actual, expected); diff != "" {
		fmt.Printf("Mismatch (-actual +expected):\n%s", diff)
	}

	value = "github/foo/bar"
	actual, _ = ParseTarget(value)
	expected = Target{GitHub: &value}

	if diff := cmp.Diff(actual, expected); diff != "" {
		fmt.Printf("Mismatch (-actual +expected):\n%s", diff)
	}
}
