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

// Remove certain elements from search queries that I personally found to be potentially problematic in finding the exact match for a song.
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
