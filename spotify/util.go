package spotify

import (
	"regexp"
	"strings"
)

var (
	starRegex = regexp.MustCompile(`\*[^\s]*`)
	featRegex = regexp.MustCompile(`[(\[]feat.*?[)\]] ?`)
	extrRegex = regexp.MustCompile(`[(\[].*?[)\]] ?`)
)

// Remove problematic elements from search queries.
func sanitizeQuery(str string) string {
	return strings.TrimSpace(starRegex.ReplaceAllStringFunc(featRegex.ReplaceAllString(str, ""), func(s string) string {
		return strings.Repeat("*", len(s))
	}))
}

func sanitizeString(str string) string {
	return strings.TrimSpace(starRegex.ReplaceAllStringFunc(extrRegex.ReplaceAllString(str, ""), func(s string) string {
		return strings.Repeat("*", len(s))
	}))
}
