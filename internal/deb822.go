package internal

import "strings"

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

func (deb *Deb822) String() string {
	return deb.builder.String()
}
