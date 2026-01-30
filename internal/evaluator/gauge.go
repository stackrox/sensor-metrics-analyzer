package evaluator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/stackrox/sensor-metrics-analyzer/internal/parser"
	"github.com/stackrox/sensor-metrics-analyzer/internal/rules"
)

// EvaluateGauge evaluates a gauge threshold rule
func EvaluateGauge(rule rules.Rule, metrics parser.MetricsData, loadLevel rules.LoadLevel) rules.EvaluationResult {
	result := rules.EvaluationResult{
		RuleName:  rule.MetricName,
		Status:    rules.StatusGreen,
		Details:   []string{},
		Timestamp: time.Now(),
	}

	if rule.MetricName == "" {
		result.Message = "Metric name not specified"
		return result
	}

	metric, exists := metrics.GetMetric(rule.MetricName)
	if !exists || len(metric.Values) == 0 {
		result.Message = fmt.Sprintf("Metric %s not found", rule.MetricName)
		return result
	}

	value, _ := metric.GetSingleValue()
	result.Value = value

	// Select thresholds based on load level
	thresholds := selectThresholds(rule, loadLevel)

	extras := map[string]interface{}{
		"value_human": formatHumanNumberGauge(value),
	}

	// Evaluate thresholds
	if thresholds.HigherIsWorse {
		if value < thresholds.Low {
			result.Status = rules.StatusGreen
			result.Message = interpolate(rule.Messages.Green, value, extras)
		} else if value < thresholds.High {
			result.Status = rules.StatusYellow
			result.Message = interpolate(rule.Messages.Yellow, value, extras)
		} else {
			result.Status = rules.StatusRed
			result.Message = interpolate(rule.Messages.Red, value, extras)
		}
	} else {
		// Lower is worse (inverted) - special case for zero checks
		if thresholds.Low == 0 && thresholds.High == 0 {
			// Zero check: > 0 is good, == 0 is bad
			if value > 0 {
				result.Status = rules.StatusGreen
				result.Message = interpolate(rule.Messages.Green, value, extras)
			} else {
				result.Status = rules.StatusRed
				result.Message = interpolate(rule.Messages.Red, value, extras)
			}
		} else {
			// Normal inverted logic
			if value >= thresholds.High {
				result.Status = rules.StatusGreen
				result.Message = interpolate(rule.Messages.Green, value, extras)
			} else if value >= thresholds.Low {
				result.Status = rules.StatusYellow
				result.Message = interpolate(rule.Messages.Yellow, value, extras)
			} else {
				result.Status = rules.StatusRed
				result.Message = interpolate(rule.Messages.Red, value, extras)
			}
		}
	}

	return result
}

// interpolate replaces placeholders in template with actual values
func interpolate(template string, value float64, extras map[string]interface{}) string {
	result := template

	// Replace {value} or {value:.0f} etc. with formatted value
	// Handle format specifiers like {value:.0f}, {value:.1f}, {value:.2f}
	valueRe := regexp.MustCompile(`\{value(?::[^}]+)?\}`)
	result = valueRe.ReplaceAllStringFunc(result, func(match string) string {
		// Extract format specifier if present
		if strings.Contains(match, ":") {
			// Extract format like ".0f", ".1f", ".2f"
			formatMatch := regexp.MustCompile(`:([^}]+)`)
			if fm := formatMatch.FindStringSubmatch(match); len(fm) == 2 {
				return fmt.Sprintf("%"+fm[1], value)
			}
		}
		// Default format
		return fmt.Sprintf("%.0f", value)
	})

	// Replace other placeholders from extras map
	for key, val := range extras {
		// Handle format specifiers
		keyRe := regexp.MustCompile(`\{` + regexp.QuoteMeta(key) + `(?::[^}]+)?\}`)
		result = keyRe.ReplaceAllStringFunc(result, func(match string) string {
			if strings.Contains(match, ":") {
				formatMatch := regexp.MustCompile(`:([^}]+)`)
				if fm := formatMatch.FindStringSubmatch(match); len(fm) == 2 {
					return fmt.Sprintf("%"+fm[1], val)
				}
			}
			return fmt.Sprintf("%v", val)
		})
	}

	return result
}

func formatHumanNumberGauge(value float64) string {
	raw := strconv.FormatFloat(value, 'f', 0, 64)
	sign := ""
	if strings.HasPrefix(raw, "-") {
		sign = "-"
		raw = strings.TrimPrefix(raw, "-")
	}
	var grouped strings.Builder
	for i, r := range raw {
		if i > 0 && (len(raw)-i)%3 == 0 {
			grouped.WriteString(" ")
		}
		grouped.WriteRune(r)
	}
	return sign + grouped.String()
}
