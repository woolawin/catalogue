package build

import (
	"testing"
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
