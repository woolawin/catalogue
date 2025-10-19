package ext

import "testing"

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
