package internal

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Deb822 struct {
	builder strings.Builder
}

func (deb *Deb822) Add(key string, value string) *Deb822 {
	if len(strings.TrimSpace(value)) == 0 {
		return deb
	}
	deb.builder.WriteString(key)
	deb.builder.WriteString(": ")
	deb.builder.WriteString(value)
	deb.builder.WriteString("\n")
	return deb
}

func (deb *Deb822) AddAll(data map[string]any) *Deb822 {
	for key, value := range data {
		deb.builder.WriteString(key)
		deb.builder.WriteString(": ")
		deb.builder.WriteString(fmt.Sprintf("%v", value))
		deb.builder.WriteString("\n")
	}

	return deb
}

func (deb *Deb822) String() string {
	return deb.builder.String()
}

func DeserializeDebFile(src io.Reader) ([]map[string]string, error) {
	var output []map[string]string

	paragraph := make(map[string]string)
	var key string
	value := strings.Builder{}

	scanner := bufio.NewScanner(src)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			if key != "" {
				paragraph[key] = value.String()
			}
			if len(paragraph) != 0 {
				output = append(output, paragraph)
			}
			paragraph = make(map[string]string)
			key = ""
			value.Reset()
			continue
		}

		if line[0] == ' ' || line[0] == '\t' {
			value.WriteString("\n")
			value.WriteString(line[1:])
			continue
		}

		if key != "" {
			paragraph[key] = value.String()
			key = ""
			value.Reset()
		}

		index := strings.Index(line, ":")
		if index == -1 {
			return nil, fmt.Errorf("line not valid, missing ':' : %s", line)
		}
		key = line[:index]
		valueStart := index + 1
		if valueStart != index {
			if line[valueStart] == ' ' || line[valueStart] == '\t' {
				valueStart++
			}
			lineValue := line[valueStart:]
			value.WriteString(lineValue)
		}
	}

	if key != "" {
		paragraph[key] = value.String()
	}

	if len(paragraph) != 0 {
		output = append(output, paragraph)
	}

	return output, nil
}

func SerializeDebParagraph(data map[string]string) string {
	if len(data) == 0 {
		return ""
	}

	deb := strings.Builder{}
	for key, value := range data {
		deb.WriteString(key)
		deb.WriteString(": ")
		deb.WriteString(strings.ReplaceAll(value, "\n", " "))
		deb.WriteString("\n")
	}
	deb.WriteString("\n")

	return deb.String()
}

func SerializeDebFile(data []map[string]string) string {
	deb := strings.Builder{}

	for _, paragraph := range data {
		if len(paragraph) == 0 {
			continue
		}
		for key, value := range paragraph {
			deb.WriteString(key)
			deb.WriteString(": ")
			deb.WriteString(strings.ReplaceAll(value, "\n", " "))
			deb.WriteString("\n")
		}
		deb.WriteString("\n")
	}

	return deb.String()
}
