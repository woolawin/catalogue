package internal

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSerializeDebFile(t *testing.T) {
	t.Run("one_paragraph", func(t *testing.T) {
		in := []map[string]string{
			{
				"Name": "Bob",
				"Age":  "65",
			},
		}

		actual := SerializeDebFile(in)
		fmt.Println(strings.TrimSpace(actual))

	})

	t.Run("two_paragraphs", func(t *testing.T) {
		in := []map[string]string{
			{
				"Name": "Bob",
				"Age":  "65",
			},
			{
				"Job":     "Engineer",
				"Company": "FooBar Inc",
			},
		}

		actual := SerializeDebFile(in)
		fmt.Println(strings.TrimSpace(actual))

	})
}

func TestDeserialzeDebFile(t *testing.T) {
	t.Run("one_paragraph", func(t *testing.T) {
		in := `
Name: Bob
Age: 65
`

		actual, err := DeserializeDebFile(strings.NewReader(in))
		if err != nil {
			t.Fatal(err)
		}

		expected := []map[string]string{
			{
				"Name": "Bob",
				"Age":  "65",
			},
		}

		if diff := cmp.Diff(actual, expected); diff != "" {
			t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
		}
	})

	t.Run("two_paragraphs", func(t *testing.T) {
		in := `
Name: Bob
Age: 65

Job: Engineer
Company: FooBar Inc
`

		actual, err := DeserializeDebFile(strings.NewReader(in))
		if err != nil {
			t.Fatal(err)
		}

		expected := []map[string]string{
			{
				"Name": "Bob",
				"Age":  "65",
			},
			{
				"Job":     "Engineer",
				"Company": "FooBar Inc",
			},
		}

		if diff := cmp.Diff(actual, expected); diff != "" {
			t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
		}
	})

	t.Run("two_paragraphs_with_multiline", func(t *testing.T) {
		in := `
Name: Bob
Age: 65

Job: Software
 Engineer
Company: FooBar Inc
`

		actual, err := DeserializeDebFile(strings.NewReader(in))
		if err != nil {
			t.Fatal(err)
		}

		expected := []map[string]string{
			{
				"Name": "Bob",
				"Age":  "65",
			},
			{
				"Job":     "Software\nEngineer",
				"Company": "FooBar Inc",
			},
		}

		if diff := cmp.Diff(actual, expected); diff != "" {
			t.Fatalf("Mismatch (-actual +expected):\n%s", diff)
		}
	})
}
