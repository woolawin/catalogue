package internal

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestValidateName(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		err := ValidateName("abcd")
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("valid_with_number", func(t *testing.T) {
		err := ValidateName("foo123")
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("valid_with_number_and_underscore", func(t *testing.T) {
		err := ValidateName("foo_123")
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("invalid_dash", func(t *testing.T) {
		err := ValidateName("foo-123")
		if err == nil {
			t.Fatal("expected to FAIL")
		}
	})
}

func TestValidateNameList(t *testing.T) {
	t.Run("single", func(t *testing.T) {
		actual, err := ValidateNameList("one")
		if err != nil {
			t.Fatal(err)
		}
		expected := []string{"one"}
		if diff := cmp.Diff(actual, expected); diff != "" {
			fmt.Printf("Mismatch (-actual +expected):\n%s", diff)
		}
	})

	t.Run("multple", func(t *testing.T) {
		actual, err := ValidateNameList("one-two-three")
		if err != nil {
			t.Fatal(err)
		}
		expected := []string{"one", "two", "three"}
		if diff := cmp.Diff(actual, expected); diff != "" {
			fmt.Printf("Mismatch (-actual +expected):\n%s", diff)
		}
	})
}

func TestValidateNameAndTarget(t *testing.T) {
	t.Run("single_target", func(t *testing.T) {
		name, target, err := ValidateNameAndTarget("foo.bar")
		if err != nil {
			t.Fatal(err)
		}

		if name != "foo" {
			t.Fatal("expected name TO Be foo")
		}
		if diff := cmp.Diff(target, []string{"bar"}); diff != "" {
			fmt.Printf("Mismatch (-actual +expected):\n%s", diff)
		}
	})

	t.Run("multi_target", func(t *testing.T) {
		name, target, err := ValidateNameAndTarget("foo.bar-baz")
		if err != nil {
			t.Fatal(err)
		}

		if name != "foo" {
			t.Fatal("expected name TO Be foo")
		}
		if diff := cmp.Diff(target, []string{"bar", "baz"}); diff != "" {
			fmt.Printf("Mismatch (-actual +expected):\n%s", diff)
		}
	})
}
