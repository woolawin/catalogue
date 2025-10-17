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

func TestParseValidTargetNames(t *testing.T) {
	t.Run("single_valid", func(t *testing.T) {
		actual, err := ParseTargetNamesString("abc")
		if err != nil {
			t.Fatal(err)
		}
		expected := []string{"abc"}
		if diff := cmp.Diff(actual, expected); diff != "" {
			t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
		}
	})

	t.Run("multi_valid", func(t *testing.T) {
		actual, err := ParseTargetNamesString("abc-def")
		if err != nil {
			t.Fatal(err)
		}
		expected := []string{"abc", "def"}
		if diff := cmp.Diff(actual, expected); diff != "" {
			t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
		}
	})

	t.Run("single_valid_number", func(t *testing.T) {
		actual, err := ParseTargetNamesString("abc123")
		if err != nil {
			t.Fatal(err)
		}
		expected := []string{"abc123"}
		if diff := cmp.Diff(actual, expected); diff != "" {
			t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
		}
	})

	t.Run("single_valid_underscore", func(t *testing.T) {
		actual, err := ParseTargetNamesString("abc_123")
		if err != nil {
			t.Fatal(err)
		}
		expected := []string{"abc_123"}
		if diff := cmp.Diff(actual, expected); diff != "" {
			t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := ParseTargetNamesString("abc%")
		if err == nil {
			t.Fatal("expected to FAIL")
		}
	})
}

func TestMergeTargets(t *testing.T) {
	t.Run("compatible", func(t *testing.T) {
		a := Target{Name: "a", Architecture: AMD64}
		b := Target{Name: "b", OSReleaseID: "17"}
		c := Target{Name: "c", OSReleaseVersionCodeName: "dingo"}

		actual, err := MergeTargets([]Target{a, b, c})
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

		_, err := MergeTargets([]Target{a, b, c})
		if err == nil {
			t.Fatal("expected to FAIL")
		}
	})

	t.Run("conflict", func(t *testing.T) {
		a := Target{Name: "a", Architecture: AMD64}
		b := Target{Name: "b", OSReleaseID: "17"}
		c := Target{Name: "c", OSReleaseID: "16"}

		_, err := MergeTargets([]Target{a, b, c})
		if err == nil {
			t.Fatal("expected to FAIL")
		}
	})
}

func TestFindOSReleaseValue(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		values := []string{
			"VERSION_ID=17",
			"VERSION_NAME=\"foobar\"",
		}
		actual, found := findOSReleaseValue(values, "VERSION_ID")
		expected := "17"
		if !found {
			t.Fatal("expected to be FOUND")
		}
		if actual != expected {
			t.Fatalf("expected '%s' to be '%s'", actual, expected)
		}
	})

	t.Run("found_quoted", func(t *testing.T) {
		values := []string{
			"VERSION_ID=17",
			"VERSION_NAME=\"foobar\"",
		}
		actual, found := findOSReleaseValue(values, "VERSION_NAME")
		expected := "foobar"
		if !found {
			t.Fatal("expected to be FOUND")
		}
		if actual != expected {
			t.Fatalf("expected '%s' to be '%s'", actual, expected)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		values := []string{
			"VERSION_ID=17",
			"VERSION_NAME=\"foobar\"",
		}
		_, found := findOSReleaseValue(values, "VERSION")
		if found {
			t.Fatal("expected NOT to be found")
		}
	})

	t.Run("empty_value", func(t *testing.T) {
		values := []string{
			"VERSION_ID=17",
			"VERSION_NAME=\"foobar\"",
			"FOO=",
		}
		_, found := findOSReleaseValue(values, "FOO")
		if found {
			t.Fatal("expected NOT to be found")
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

func TestSplitTargetNames(t *testing.T) {
	actual := splitTargetNames("foo")
	expected := []string{"foo"}

	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
	}

	actual = splitTargetNames("foo-bar")
	expected = []string{"foo", "bar"}

	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
	}
}

func TestRegistryLoad(t *testing.T) {
	reg := Registry{
		base: []Target{
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
		},
	}

	actual, err := reg.Load([]string{"all", "amd64-ubuntu", "amd64"})
	if err != nil {
		t.Fatal(err)
	}
	expected := []Target{
		{
			Name: "all",
			All:  true,
		},
		{
			Name:         "amd64-ubuntu",
			Architecture: AMD64,
			OSReleaseID:  "ubuntu",
		},
		{
			Name:         "amd64",
			Architecture: AMD64,
		},
	}

	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
	}
}
