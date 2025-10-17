package build

import "testing"

func TestFilePath(t *testing.T) {
	path := filePath("/tmp/1234567", "foo")
	if path != "/tmp/1234567/foo" {
		t.Fatalf("'%s' not correct", path)
	}

	path = filePath("/tmp/1234567", "foo", "bar")
	if path != "/tmp/1234567/foo/bar" {
		t.Fatalf("'%s' not correct", path)
	}
}
