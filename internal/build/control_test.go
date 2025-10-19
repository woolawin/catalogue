package build

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/woolawin/catalogue/internal/component"
	"github.com/woolawin/catalogue/internal/target"
)

func TestControlDataString(t *testing.T) {
	data := ControlData{
		Package:      "foo-bar",
		Version:      "2.3.0",
		Depends:      []string{"baz", "doh"},
		Recommends:   []string{"hello", "world"},
		Section:      "other",
		Priority:     "normal",
		Homepage:     "https://foobar.com",
		Architecture: "amd64",
		Maintainer:   "me",
		Description:  "meh",
	}

	actual := data.String()
	expected := `Package: foo-bar
Version: 2.3.0
Depends: baz,doh
Recommends: hello|world
Section: other
Priority: normal
Homepage: https://foobar.com
Architecture: amd64
Maintainer: me
Description: meh
`
	if actual != expected {
		t.Fatalf("'%s' was not '%s'\n", actual, expected)
	}
}

func TestMergeMeta(t *testing.T) {
	system := target.System{Architecture: target.AMD64}

	metadatas := []*component.Metadata{
		{
			Target:       target.Target{Name: "all", All: true},
			Dependencies: []string{"foo", "bar"},
			Section:      "utilities",
			Priority:     "normal",
			Homepage:     "https://foobar.com",
			Description:  "foo bar",
			Maintainer:   "Bob Doe",
		},
		{
			Target:       target.Target{Name: "amd64", Architecture: target.AMD64},
			Architecture: "amd64",
			Maintainer:   "Jane Doe",
		},
		{
			Target:          target.Target{Name: "arm64", Architecture: target.ARM64},
			Recommendations: []string{"happy", "puppy"},
		},
	}

	actual, err := Metadata(metadatas, system)
	if err != nil {
		t.Fatal(err)
	}
	expected := component.Metadata{
		Dependencies: []string{"foo", "bar"},
		Section:      "utilities",
		Priority:     "normal",
		Homepage:     "https://foobar.com",
		Description:  "foo bar",
		Maintainer:   "Jane Doe",
		Architecture: "amd64",
	}

	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
	}
}
