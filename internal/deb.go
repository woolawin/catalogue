package internal

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func DebMultiLine(lines []string) string {
	str := strings.Builder{}
	for idx, line := range lines {
		if idx > 0 {
			str.WriteString(" ")
		}
		str.WriteString(line)
		if idx < (len(lines) - 1) {
			str.WriteString("\n")
		}
	}
	return str.String()
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
			deb.WriteString(":")
			if len(value) == 0 {
				deb.WriteString("\n")
				continue
			}
			if value[0] != '\n' {
				deb.WriteString(" ")
			}
			deb.WriteString(value)
			deb.WriteString("\n")
		}
		deb.WriteString("\n")
	}

	return deb.String()
}
