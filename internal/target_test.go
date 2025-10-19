package internal

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRank(t *testing.T) {
	t.Run("first", func(t *testing.T) {
		targets := []*Target{
			{Name: "amd64", Architecture: AMD64},
			{Name: "arm64", Architecture: ARM64},
		}

		system := System{Architecture: AMD64}

		actual := Ranked(system, targets)
		expected := []*Target{targets[0]}

		if diff := cmp.Diff(actual, expected); diff != "" {
			t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
		}
	})

	t.Run("second", func(t *testing.T) {
		targets := []*Target{
			{Name: "amd64", Architecture: AMD64},
			{Name: "arm64", Architecture: ARM64},
		}

		system := System{Architecture: ARM64}

		actual := Ranked(system, targets)
		expected := []*Target{targets[1]}

		if diff := cmp.Diff(actual, expected); diff != "" {
			t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
		}
	})

	t.Run("all_is_last", func(t *testing.T) {
		targets := []*Target{
			{Name: "all", All: true},
			{Name: "amd64", Architecture: AMD64},
			{Name: "arm64", Architecture: ARM64},
		}

		system := System{Architecture: ARM64}

		actual := Ranked(system, targets)
		expected := []*Target{targets[2], targets[0]}

		if diff := cmp.Diff(actual, expected); diff != "" {
			t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
		}
	})
}

func TestMergeTargets(t *testing.T) {
	t.Run("compatible", func(t *testing.T) {
		a := Target{Name: "a", Architecture: AMD64}
		b := Target{Name: "b", OSReleaseID: "17"}
		c := Target{Name: "c", OSReleaseVersionCodeName: "dingo"}

		actual, err := mergeTargets([]Target{a, b, c})
		if err != nil {
			t.Fatal(err)
		}
		expected := Target{
			Name:                     "a-b-c",
			Architecture:             AMD64,
			OSReleaseID:              "17",
			OSReleaseVersionCodeName: "dingo",
		}

		if diff := cmp.Diff(actual, expected); diff != "" {
			t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
		}
	})

	t.Run("can_not_merge_all", func(t *testing.T) {
		a := Target{Name: "a", Architecture: AMD64}
		b := Target{Name: "b", OSReleaseID: "17"}
		c := Target{Name: "all", All: true}

		_, err := mergeTargets([]Target{a, b, c})
		if err == nil {
			t.Fatal("expected to FAIL")
		}
	})

	t.Run("conflict", func(t *testing.T) {
		a := Target{Name: "a", Architecture: AMD64}
		b := Target{Name: "b", OSReleaseID: "17"}
		c := Target{Name: "c", OSReleaseID: "16"}

		_, err := mergeTargets([]Target{a, b, c})
		if err == nil {
			t.Fatal("expected to FAIL")
		}
	})
}

func TestScore(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		system := System{
			Architecture:             AMD64,
			OSReleaseID:              "ubuntu",
			OSReleaseVersion:         "4.0",
			OSReleaseVersionID:       "4",
			OSReleaseVersionCodeName: "dingo",
		}

		target := Target{
			Architecture:             "",
			OSReleaseID:              "",
			OSReleaseVersion:         "",
			OSReleaseVersionID:       "",
			OSReleaseVersionCodeName: "",
		}

		actual, applicable := score(system, target)
		if !applicable {
			t.Fatal("expected to be APPLICABLE")
		}
		expected := 0

		if actual != expected {
			t.Fatalf("expected '%d' to be '%d'", actual, expected)
		}
	})

	t.Run("one", func(t *testing.T) {
		system := System{
			Architecture:             AMD64,
			OSReleaseID:              "ubuntu",
			OSReleaseVersion:         "4.0",
			OSReleaseVersionID:       "4",
			OSReleaseVersionCodeName: "dingo",
		}

		target := Target{
			Architecture:             AMD64,
			OSReleaseID:              "",
			OSReleaseVersion:         "",
			OSReleaseVersionID:       "",
			OSReleaseVersionCodeName: "",
		}

		actual, applicable := score(system, target)
		if !applicable {
			t.Fatal("expected to be APPLICABLE")
		}
		expected := 1

		if actual != expected {
			t.Fatalf("expected '%d' to be '%d'", actual, expected)
		}
	})

	t.Run("two", func(t *testing.T) {
		system := System{
			Architecture:             AMD64,
			OSReleaseID:              "ubuntu",
			OSReleaseVersion:         "4.0",
			OSReleaseVersionID:       "4",
			OSReleaseVersionCodeName: "dingo",
		}

		target := Target{
			Architecture:             AMD64,
			OSReleaseID:              "",
			OSReleaseVersion:         "",
			OSReleaseVersionID:       "4",
			OSReleaseVersionCodeName: "",
		}

		actual, applicable := score(system, target)
		if !applicable {
			t.Fatal("expected to be APPLICABLE")
		}
		expected := 2

		if actual != expected {
			t.Fatalf("expected '%d' to be '%d'", actual, expected)
		}
	})

	t.Run("full", func(t *testing.T) {
		system := System{
			Architecture:             AMD64,
			OSReleaseID:              "ubuntu",
			OSReleaseVersion:         "4.0",
			OSReleaseVersionID:       "4",
			OSReleaseVersionCodeName: "dingo",
		}

		target := Target{
			Architecture:             AMD64,
			OSReleaseID:              "ubuntu",
			OSReleaseVersion:         "4.0",
			OSReleaseVersionID:       "4",
			OSReleaseVersionCodeName: "dingo",
		}

		actual, applicable := score(system, target)
		if !applicable {
			t.Fatal("expected to be APPLICABLE")
		}
		expected := 5

		if actual != expected {
			t.Fatalf("expected '%d' to be '%d'", actual, expected)
		}
	})

	t.Run("not_applicable", func(t *testing.T) {
		system := System{
			Architecture:             AMD64,
			OSReleaseID:              "ubuntu",
			OSReleaseVersion:         "4.0",
			OSReleaseVersionID:       "4",
			OSReleaseVersionCodeName: "dingo",
		}

		target := Target{
			Architecture:             ARM64,
			OSReleaseID:              "",
			OSReleaseVersion:         "",
			OSReleaseVersionID:       "",
			OSReleaseVersionCodeName: "",
		}

		_, applicable := score(system, target)
		if applicable {
			t.Fatal("expected NOT to be applicable")
		}
	})

	t.Run("not_applicable_other", func(t *testing.T) {
		system := System{
			Architecture:             AMD64,
			OSReleaseID:              "ubuntu",
			OSReleaseVersion:         "4.0",
			OSReleaseVersionID:       "4",
			OSReleaseVersionCodeName: "dingo",
		}

		target := Target{
			Architecture:             AMD64,
			OSReleaseID:              "",
			OSReleaseVersion:         "",
			OSReleaseVersionID:       "5",
			OSReleaseVersionCodeName: "",
		}

		_, applicable := score(system, target)
		if applicable {
			t.Fatal("expected NOT to be applicable")
		}
	})
}

func TestBuildTarget(t *testing.T) {

	t.Run("just_all", func(t *testing.T) {
		targets := []Target{
			{
				Name: "all",
				All:  true,
			},
			{
				Name:         "amd64",
				Architecture: AMD64,
			},
			{
				Name:         "arm64",
				Architecture: "arm64",
			},
			{
				Name:        "ubuntu",
				OSReleaseID: "ubuntu",
			},
		}

		actual, err := BuildTarget(targets, []string{"all"})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(actual, targets[0]); diff != "" {
			t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
		}
	})

	t.Run("multi", func(t *testing.T) {
		targets := []Target{
			{
				Name: "all",
				All:  true,
			},
			{
				Name:         "amd64",
				Architecture: AMD64,
			},
			{
				Name:         "arm64",
				Architecture: "arm64",
			},
			{
				Name:        "ubuntu",
				OSReleaseID: "ubuntu",
			},
		}

		actual, err := BuildTarget(targets, []string{"amd64", "ubuntu"})
		if err != nil {
			t.Fatal(err)
		}
		expected := Target{
			Name:         "amd64-ubuntu",
			Architecture: AMD64,
			OSReleaseID:  "ubuntu",
		}

		if diff := cmp.Diff(actual, expected); diff != "" {
			t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
		}
	})
}
