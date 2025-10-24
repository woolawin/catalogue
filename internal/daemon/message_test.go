package daemon

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	msgpacklib "github.com/vmihailenco/msgpack/v5"
	"github.com/woolawin/catalogue/internal"
)

func TestSerializeLog(t *testing.T) {

	now := time.Now().UTC().Truncate(time.Second)

	expected := internal.NewLogStatement("server", 6, now, "Hello", map[string]any{"to": "world"}, true)

	buffer := bytes.NewBuffer([]byte{})

	err := msgpacklib.NewEncoder(buffer).Encode(&expected)
	if err != nil {
		t.Fatal(err)
	}

	actual := internal.LogStatement{}

	err = msgpacklib.NewDecoder(buffer).Decode(&actual)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(actual, expected, cmpopts.IgnoreUnexported()); diff != "" {
		fmt.Printf("Mismatch (-actual +expected):\n%s", diff)
	}
}
