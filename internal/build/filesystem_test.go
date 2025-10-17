package build

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseFileSystemRef(t *testing.T) {
	t.Run("invalid_no_targets", func(t *testing.T) {
		_, err := parseFileSystemRef("root")
		if err == nil {
			t.Fatal("expetced to FAIL")
		}
	})

	t.Run("valid_1_target", func(t *testing.T) {
		actual, err := parseFileSystemRef("root.foo")
		if err != nil {
			t.Fatal(err)
		}

		expected := FileSystemRef{Anchor: "root", Targets: []string{"foo"}, Target: "foo"}
		if diff := cmp.Diff(actual, expected); diff != "" {
			t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
		}
	})

	t.Run("valid_2_targets", func(t *testing.T) {
		actual, err := parseFileSystemRef("root.foo-bar")
		if err != nil {
			t.Fatal(err)
		}

		expected := FileSystemRef{Anchor: "root", Targets: []string{"foo", "bar"}, Target: "foo-bar"}
		if diff := cmp.Diff(actual, expected); diff != "" {
			t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
		}
	})

	t.Run("invalid_anchor", func(t *testing.T) {
		_, err := parseFileSystemRef("b2b-ccc")
		if err == nil {
			t.Fatal("expected to FAIL")
		}
	})
}
