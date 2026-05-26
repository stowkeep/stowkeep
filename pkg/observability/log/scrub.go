package log

import (
	"strings"
)

// ScrubQuery scrubs sensitive query parameter values from a URL query string.
func ScrubQuery(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}
	parts := strings.Split(rawQuery, "&")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		key, _, ok := strings.Cut(part, "=")
		if !ok {
			out = append(out, part)
			continue
		}
		if isSensitiveKey(key) {
			out = append(out, key+"=[REDACTED]")
			continue
		}
		out = append(out, part)
	}
	return strings.Join(out, "&")
}

func isSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	sensitive := []string{"token", "key", "secret", "password", "auth", "session"}
	for _, s := range sensitive {
		if strings.Contains(lower, s) {
			return true
		}
	}
	return false
}

// ScrubHeaderValue returns a redacted placeholder for sensitive HTTP headers.
func ScrubHeaderValue(name, value string) string {
	if value == "" {
		return ""
	}
	switch strings.ToLower(name) {
	case "authorization", "cookie", "set-cookie", "x-api-key":
		return "[REDACTED]"
	default:
		return value
	}
}
