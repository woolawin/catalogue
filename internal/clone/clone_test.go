package clone

import (
	"testing"
)

func TestIsInPath(t *testing.T) {
	t.Run("is_in", func(t *testing.T) {
		if !isInPath("/foo/bar", "/foo/bar/baz") {
			t.Fatal("expected to be IN path")
		}

		if !isInPath("foo/bar", "foo/bar/baz") {
			t.Fatal("expected to be IN path")
		}
	})

	t.Run("self", func(t *testing.T) {
		if !isInPath("/foo/bar", "/foo/bar") {
			t.Fatal("expected to be IN path")
		}
	})

	t.Run("is_out", func(t *testing.T) {
		if isInPath("/foo/bar", "/foo/baz/doh") {
			t.Fatal("expected NOT to be in path")
		}
	})

	t.Run("sibling", func(t *testing.T) {
		if isInPath("/foo/bar", "/foo/baz") {
			t.Fatal("expected NOT to be in path")
		}
	})

	t.Run("parent", func(t *testing.T) {
		if isInPath("/foo/bar", "/foo") {
			t.Fatal("expected NOT to be in path")
		}
	})
}
