package services

import "strings"

// NormalizeEmail trims spaces and lowercases an email value.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// NormalizeText trims spaces from free-form text.
func NormalizeText(value string) string {
	return strings.TrimSpace(value)
}
