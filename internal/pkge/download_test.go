package pkge

import (
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func u(value string) *url.URL {
	res, err := url.Parse(value)
	if err != nil {
		panic(err)
	}
	return res
}

func TestValidateDownload(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		raw := RawDownload{
			Source:      "https://foo.com/bar.txt",
			Destination: "path://root/etc/foo/bar.txt",
		}

		actual, err := raw.validate()
		if err != nil {
			t.Fatal(err)
		}
		expected := Download{
			Source:      u("https://foo.com/bar.txt"),
			Destination: u("path://root/etc/foo/bar.txt"),
		}
		if diff := cmp.Diff(actual, expected); diff != "" {
			t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
		}
	})
}
