package evaluator

import (
	"fmt"
	"strings"
	"time"

	"github.com/stackrox/sensor-metrics-analyzer/internal/parser"
	"github.com/stackrox/sensor-metrics-analyzer/internal/rules"
)

// EvaluateComposite evaluates a composite (multi-metric) rule
func EvaluateComposite(rule rules.Rule, metrics parser.MetricsData, loadLevel rules.LoadLevel) rules.EvaluationResult {
	result := rules.EvaluationResult{
		RuleName:  rule.DisplayName,
		Status:    rules.StatusGreen,
		Details:   []string{},
		Timestamp: time.Now(),
	}

	if rule.CompositeConfig == nil {
		result.Message = "Composite config not specified"
		return result
	}

	// Collect metric values
	metricValues := make(map[string]float64)
	for _, metricDef := range rule.CompositeConfig.Metrics {
		metric, exists := metrics.GetMetric(metricDef.Source)
		if !exists || len(metric.Values) == 0 {
			result.Message = fmt.Sprintf("Metric %s not found", metricDef.Source)
			return result
		}
		value, _ := metric.GetSingleValue()
		metricValues[metricDef.Name] = value
		result.Details = append(result.Details, fmt.Sprintf("%s: %.3f", metricDef.Name, value))
	}

	// Evaluate checks in order
	for _, check := range rule.CompositeConfig.Checks {
		if evaluateCompositeCheck(check, metricValues) {
			// Check condition met, apply status and message
			status := parseStatus(check.Status)
			if status != "" {
				result.Status = status
			}
			result.Message = interpolateCompositeMessage(check.Message, metricValues)
			return result
		}
	}

	// All checks passed, use green message
	result.Message = interpolateCompositeMessage(rule.Messages.Green, metricValues)
	return result
}

// evaluateCompositeCheck evaluates a composite check condition
func evaluateCompositeCheck(check rules.CompositeCheck, metricValues map[string]float64) bool {
	switch check.CheckType {
	case "not_zero":
		// Check if any metric is zero
		for _, metricName := range check.Metrics {
			if value, exists := metricValues[metricName]; exists && value == 0 {
				return true
			}
		}
		return false

	case "ratio":
		// Check if ratio is below minimum
		numerator, numExists := metricValues[check.Numerator]
		denominator, denExists := metricValues[check.Denominator]
		if !numExists || !denExists || denominator == 0 {
			return false
		}
		ratio := numerator / denominator
		return ratio < check.MinRatio

	default:
		return false
	}
}

// interpolateCompositeMessage replaces placeholders in composite message
func interpolateCompositeMessage(template string, metricValues map[string]float64) string {
	result := template
	for name, value := range metricValues {
		placeholder := "{" + name + "}"
		replacement := fmt.Sprintf("%.0f", value)
		result = strings.ReplaceAll(result, placeholder, replacement)
	}
	return result
}

// parseStatus parses a status string
func parseStatus(statusStr string) rules.Status {
	switch statusStr {
	case "RED", "red":
		return rules.StatusRed
	case "YELLOW", "yellow":
		return rules.StatusYellow
	case "GREEN", "green":
		return rules.StatusGreen
	default:
		return ""
	}
}
