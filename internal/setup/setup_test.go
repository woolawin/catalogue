package setup

import (
	"testing"

	"github.com/woolawin/catalogue/internal"
)

func TestAddress(t *testing.T) {
	config := internal.Config{Port: 3465}

	actual := address(&config, "catalogue")
	expected := "http://localhost:3465/repositories/catalogue"

	if actual != expected {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
}
