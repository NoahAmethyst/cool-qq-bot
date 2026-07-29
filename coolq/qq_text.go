package coolq

import (
	"regexp"
	"strings"
)

var (
	markdownImagePattern  = regexp.MustCompile(`!\[([^\]]*)\]\([^)]+\)`)
	markdownLinkPattern   = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	markdownHeadingPrefix = regexp.MustCompile(`^\s{0,3}#{1,6}\s*`)
	markdownQuotePrefix   = regexp.MustCompile(`^\s{0,3}>\s?`)
	markdownListPrefix    = regexp.MustCompile(`^\s{0,3}(?:[-+*]|\d+[.)])\s+`)
	markdownTableDivider  = regexp.MustCompile(`^\s*\|?[\s:|-]+\|?\s*$`)
)

func toPlainQQText(answer string) string {
	answer = strings.ReplaceAll(answer, "\r\n", "\n")
	answer = strings.ReplaceAll(answer, "\r", "\n")
	answer = markdownImagePattern.ReplaceAllString(answer, "$1")
	answer = markdownLinkPattern.ReplaceAllString(answer, "$1")

	lines := strings.Split(answer, "\n")
	result := make([]string, 0, len(lines))
	previousBlank := false
	replacer := strings.NewReplacer(
		"```", "",
		"**", "",
		"__", "",
		"~~", "",
		"`", "",
		"*", "",
		"|", " ",
	)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || markdownTableDivider.MatchString(trimmed) {
			continue
		}
		line = markdownHeadingPrefix.ReplaceAllString(line, "")
		line = markdownQuotePrefix.ReplaceAllString(line, "")
		line = markdownListPrefix.ReplaceAllString(line, "")
		line = strings.TrimSpace(replacer.Replace(line))

		if line == "" {
			if previousBlank || len(result) == 0 {
				continue
			}
			result = append(result, "")
			previousBlank = true
			continue
		}

		result = append(result, line)
		previousBlank = false
	}

	for len(result) > 0 && result[len(result)-1] == "" {
		result = result[:len(result)-1]
	}
	return strings.TrimSpace(strings.Join(result, "\n"))
}
