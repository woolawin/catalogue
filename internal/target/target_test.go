package target

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRank(t *testing.T) {
	t.Run("first", func(t *testing.T) {
		targets := []Target{
			{Name: "amd64", Architecture: AMD64},
			{Name: "arm64", Architecture: ARM64},
		}

		system := System{Architecture: AMD64}

		actual := system.Rank(targets)
		expected := []int{0}

		if diff := cmp.Diff(actual, expected); diff != "" {
			t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
		}
	})

	t.Run("second", func(t *testing.T) {
		targets := []Target{
			{Name: "amd64", Architecture: AMD64},
			{Name: "arm64", Architecture: ARM64},
		}

		system := System{Architecture: ARM64}

		actual := system.Rank(targets)
		expected := []int{1}

		if diff := cmp.Diff(actual, expected); diff != "" {
			t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
		}
	})

	t.Run("all_is_last", func(t *testing.T) {
		targets := []Target{
			{Name: "all", All: true},
			{Name: "amd64", Architecture: AMD64},
			{Name: "arm64", Architecture: ARM64},
		}

		system := System{Architecture: ARM64}

		actual := system.Rank(targets)
		expected := []int{2, 0}

		if diff := cmp.Diff(actual, expected); diff != "" {
			t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
		}
	})

}
