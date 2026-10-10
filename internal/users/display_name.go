package users

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

// MaxDisplayName is the longest display name, in characters.
const MaxDisplayName = 30

// CleanDisplayName tidies a display name: trims it, turns runs of spaces
// (and any control characters) into one space, and checks the length.
// "" is allowed and means "use my username".
func CleanDisplayName(raw string) (string, error) {
	fields := strings.FieldsFunc(raw, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) })
	name := strings.Join(fields, " ")
	if utf8.RuneCountInString(name) > MaxDisplayName {
		return "", errors.New("name can be at most 30 characters")
	}
	if name != "" && utf8.RuneCountInString(name) < 2 {
		return "", errors.New("name needs at least 2 characters")
	}
	return name, nil
}
