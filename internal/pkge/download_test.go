package pkge

import (
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/woolawin/catalogue/internal/target"
)

func TestLoadDownloads(t *testing.T) {
	targets := []target.Target{
		{
			Name:         "amd64",
			Architecture: target.AMD64,
		},
		{
			Name:         "arm64",
			Architecture: target.ARM64,
		},
		{
			Name: "all",
			All:  true,
		},
		{
			Name:        "ubuntu",
			OSReleaseID: "ubuntu",
		},
	}

	deserialized := map[string]map[string]DownloadTOML{
		"bin": {
			"amd64": {
				Source:      " https://foo.com/bar-x86-64  ",
				Destination: "  path://root/usr/bin/bar  ",
			},
			"arm64": {
				Source:      " https://foo.com/bar-arm  ",
				Destination: "  path://root/usr/bin/bar  ",
			},
		},
		"logo": {
			"amd64-ubuntu": {
				Source:      "https://foo.com/logo.svg",
				Destination: " path://root/usr/bin/foo.svg ",
			},
		},
	}

	actual, err := loadDownloads(deserialized, targets)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string][]*Download{
		"bin": {
			{
				ID:          "bin.amd64",
				Name:        "bin",
				Target:      target.Target{Name: "amd64", Architecture: target.AMD64},
				Source:      u("https://foo.com/bar-x86-64"),
				Destination: u("path://root/usr/bin/bar"),
			},
			{
				ID:          "bin.arm64",
				Name:        "bin",
				Target:      target.Target{Name: "arm64", Architecture: target.ARM64},
				Source:      u("https://foo.com/bar-arm"),
				Destination: u("path://root/usr/bin/bar"),
			},
		},
		"logo": {
			{
				ID:          "logo.amd64-ubuntu",
				Name:        "logo",
				Target:      target.Target{Name: "amd64-ubuntu", Architecture: target.AMD64, OSReleaseID: "ubuntu"},
				Source:      u("https://foo.com/logo.svg"),
				Destination: u("path://root/usr/bin/foo.svg"),
			},
		},
	}

	if diff := cmp.Diff(actual, expected, cmpopts.SortSlices(sortByTarget)); diff != "" {
		t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
	}

}

func u(value string) *url.URL {
	res, err := url.Parse(value)
	if err != nil {
		panic(err)
	}
	return res
}
