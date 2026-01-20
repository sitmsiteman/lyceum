package tlgcore

import (
	"regexp"
	"strings"
)

func NormalizeLatin(s string) string {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	s = fields[0]
	re := regexp.MustCompile(`[\^_\d\#]`)
	return strings.ToLower(re.ReplaceAllString(s, ""))
}

func NormalizeStrict(s string) string {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	s = fields[0]
	re := regexp.MustCompile(`[/\(\)\\=\|\+\^_\d]`)
	return strings.ToLower(re.ReplaceAllString(s, ""))
}

func NormalizeFuzzy(s string) string {
	s = NormalizeStrict(s)
	vowelFuzzer := strings.NewReplacer("e", "a", "h", "a", "o", "a", "w", "a")
	return vowelFuzzer.Replace(s)
}

func NormalizeBetaCode(s string) string {
	return strings.ReplaceAll(s, "\\", "/")
}
