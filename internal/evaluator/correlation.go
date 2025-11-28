package evaluator

import (
	"github.com/stackrox/sensor-metrics-analyzer/internal/parser"
	"github.com/stackrox/sensor-metrics-analyzer/internal/rules"
)

// EvaluateCorrelation evaluates correlation conditions and modifies status if needed
func EvaluateCorrelation(rule rules.Rule, metrics parser.MetricsData, result rules.EvaluationResult) rules.EvaluationResult {
	if rule.Correlation == nil {
		return result
	}

	// Evaluate suppress_if conditions
	for _, cond := range rule.Correlation.SuppressIf {
		if evaluateCondition(cond, metrics) {
			// Suppress: downgrade status
			result.Status = suppressStatus(result.Status)
			if cond.Status != "" {
				result.Status = cond.Status
			}
		}
	}

	// Evaluate elevate_if conditions
	for _, cond := range rule.Correlation.ElevateIf {
		if evaluateCondition(cond, metrics) {
			// Elevate: upgrade status
			result.Status = elevateStatus(result.Status)
			if cond.Status != "" {
				result.Status = cond.Status
			}
		}
	}

	return result
}

// evaluateCondition checks if a correlation condition is met
func evaluateCondition(cond rules.CorrelationCondition, metrics parser.MetricsData) bool {
	metric, exists := metrics.GetMetric(cond.MetricName)
	if !exists || len(metric.Values) == 0 {
		return false
	}

	// Get metric value (use first value or sum)
	var value float64
	if len(metric.Values) == 1 {
		value = metric.Values[0].Value
	} else {
		value = metric.SumValues()
	}

	// Evaluate operator
	switch cond.Operator {
	case "gt":
		return value > cond.Value
	case "lt":
		return value < cond.Value
	case "eq":
		return value == cond.Value
	case "gte":
		return value >= cond.Value
	case "lte":
		return value <= cond.Value
	default:
		return false
	}
}

// suppressStatus downgrades a status (RED -> YELLOW, YELLOW -> GREEN)
func suppressStatus(status rules.Status) rules.Status {
	switch status {
	case rules.StatusRed:
		return rules.StatusYellow
	case rules.StatusYellow:
		return rules.StatusGreen
	default:
		return status
	}
}

// elevateStatus upgrades a status (GREEN -> YELLOW, YELLOW -> RED)
func elevateStatus(status rules.Status) rules.Status {
	switch status {
	case rules.StatusGreen:
		return rules.StatusYellow
	case rules.StatusYellow:
		return rules.StatusRed
	default:
		return status
	}
}
