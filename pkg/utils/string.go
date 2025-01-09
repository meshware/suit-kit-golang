package utils

import "strings"

// IsNilOrEmpty The function is used to check if a string pointer is nil or if the string it points to is empty.
func IsNilOrEmpty(value *string) bool {
	if value == nil {
		return true
	}
	return strings.TrimSpace(*value) == "" || len(*value) == 0
}

// IsEmptyOrBlank The function is used to check if a string is empty or if it contains only whitespace characters.
func IsEmptyOrBlank(value string) bool {
	return strings.TrimSpace(value) == "" || len(value) == 0
}
