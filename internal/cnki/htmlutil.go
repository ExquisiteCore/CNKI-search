package cnki

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
)

func textOnly(fragment string) string {
	if fragment == "" {
		return ""
	}
	noTags := regexp.MustCompile(`(?is)<[^>]+>`).ReplaceAllString(fragment, " ")
	return collapseSpace(html.UnescapeString(noTags))
}

func collapseSpace(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}

func attrValue(tag, attr string) string {
	if tag == "" {
		return ""
	}
	pattern := fmt.Sprintf(`(?is)\b%s\s*=\s*["']([^"']*)["']`, regexp.QuoteMeta(attr))
	m := regexp.MustCompile(pattern).FindStringSubmatch(tag)
	if len(m) < 2 {
		return ""
	}
	return html.UnescapeString(m[1])
}

func firstMatch(s, pattern string) string {
	m := regexp.MustCompile(pattern).FindStringSubmatch(s)
	if len(m) < 2 {
		if len(m) == 1 {
			return m[0]
		}
		return ""
	}
	return m[1]
}

func allMatches(s, pattern string) []string {
	matches := regexp.MustCompile(pattern).FindAllStringSubmatch(s, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) > 1 {
			out = append(out, m[1])
		}
	}
	return out
}

func firstYear(s string) int {
	re := regexp.MustCompile(`\b(19|20|21)\d{2}\b`)
	m := re.FindString(s)
	return intOrZero(m)
}

func intOrZero(s string) int {
	digits := regexp.MustCompile(`\D+`).ReplaceAllString(s, "")
	if digits == "" {
		return 0
	}
	n, err := strconv.Atoi(digits)
	if err != nil {
		return 0
	}
	return n
}
