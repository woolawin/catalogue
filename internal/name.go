package internal

import (
	"strings"
	"unicode"
)

func ValidateName(value string) error {
	return validateNameWithOffset(value, 0)
}

func validateNameWithOffset(value string, offset int) error {
	for i, r := range value {
		if unicode.IsLower(r) || unicode.IsNumber(r) || string(r) == "_" {
			continue
		}
		return Err("invalid caharcter '%s' at %d", string(r), offset+i)
	}
	return nil
}

func ValidateNameList(value string) ([]string, error) {
	var names []string
	current := strings.Builder{}
	currentIdx := 0
	for i, r := range value {
		if string(r) == "-" {
			if current.Len() == 0 {
				return nil, Err("name seperator '-' at position %d out of place", i)
			}
			err := validateNameWithOffset(current.String(), currentIdx)
			if err != nil {
				return nil, err
			}
			names = append(names, current.String())
			current.Reset()
			currentIdx = i
			continue
		}

		current.WriteRune(r)
	}
	if current.Len() != 0 {
		err := validateNameWithOffset(current.String(), currentIdx)
		if err != nil {
			return nil, err
		}
		names = append(names, current.String())
	}
	return names, nil
}
