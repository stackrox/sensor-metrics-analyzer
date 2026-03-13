package evaluator

import (
	"regexp"
	"strings"
)

var (
	unitTokenToCanonical = map[string]string{
		"seconds":      "seconds",
		"second":       "seconds",
		"s":            "seconds",
		"sec":          "seconds",
		"secs":         "seconds",
		"milliseconds": "milliseconds",
		"millisecond":  "milliseconds",
		"ms":           "milliseconds",
		"msec":         "milliseconds",
		"msecs":        "milliseconds",
		"millis":       "milliseconds",
		"bytes":        "bytes",
		"byte":         "bytes",
	}

	helpUnitMatchers = map[string]*regexp.Regexp{
		"milliseconds": regexp.MustCompile(`(?i)\bmilliseconds?\b|\bms\b|\bmsecs?\b|\bmillis\b`),
		"seconds":      regexp.MustCompile(`(?i)\bseconds?\b|\bsecs?\b|\bsec\b`),
		"bytes":        regexp.MustCompile(`(?i)\bbytes?\b`),
	}
)

// guessMetricUnit infers unit from metric name first, then HELP text.
// HELP text is only used when exactly one unit candidate is detected.
func guessMetricUnit(metricName, helpText string) string {
	if unit := guessUnitFromMetricName(metricName); unit != "" {
		return unit
	}
	return guessUnitFromHelpText(helpText)
}

func guessUnitFromMetricName(metricName string) string {
	metricName = strings.TrimSpace(strings.ToLower(metricName))
	if metricName == "" {
		return ""
	}

	parts := strings.Split(metricName, "_")
	if len(parts) == 0 {
		return ""
	}

	// Prometheus best-practice suffixes like *_seconds_total.
	if len(parts) >= 2 && parts[len(parts)-1] == "total" {
		if unit, ok := unitTokenToCanonical[parts[len(parts)-2]]; ok {
			return unit
		}
	}

	// Common suffixes like *_seconds and *_bytes.
	if unit, ok := unitTokenToCanonical[parts[len(parts)-1]]; ok {
		return unit
	}

	// Handle *_timestamp_seconds.
	if len(parts) >= 2 && parts[len(parts)-2] == "timestamp" {
		if unit, ok := unitTokenToCanonical[parts[len(parts)-1]]; ok {
			return unit
		}
	}

	return ""
}

func guessUnitFromHelpText(helpText string) string {
	helpText = strings.TrimSpace(helpText)
	if helpText == "" {
		return ""
	}

	var matched []string
	for unit, matcher := range helpUnitMatchers {
		if matcher.MatchString(helpText) {
			matched = append(matched, unit)
		}
	}

	// Use HELP only if it points to exactly one unit.
	if len(matched) != 1 {
		return ""
	}
	return matched[0]
}
