package ext

import (
	"testing"
)

func TestFilePath(t *testing.T) {
	disk := diskImpl{base: "/foo"}
	path := disk.Path("/tmp/1234567", "foo")
	if path != "/foo/tmp/1234567/foo" {
		t.Fatalf("'%s' not correct", path)
	}

	path = disk.Path("/tmp/1234567", "foo", "bar")
	if path != "/foo/tmp/1234567/foo/bar" {
		t.Fatalf("'%s' not correct", path)
	}
}

func TestArchiveFilePath(t *testing.T) {
	actual, err := arhiveFilePath("/foo/bar/baz", "/foo/bar/baz/doh/data.txt")
	if err != nil {
		t.Fatal(err)
	}
	expected := "/doh/data.txt"
	if actual != expected {
		t.Fatalf("expected '%s' to be '%s'", actual, expected)
	}
}
