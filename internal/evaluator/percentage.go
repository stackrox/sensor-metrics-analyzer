package evaluator

import (
	"fmt"
	"time"

	"github.com/stackrox/sensor-metrics-analyzer/internal/parser"
	"github.com/stackrox/sensor-metrics-analyzer/internal/rules"
)

// EvaluatePercentage evaluates a percentage/ratio rule
func EvaluatePercentage(rule rules.Rule, metrics parser.MetricsData, loadLevel rules.LoadLevel) rules.EvaluationResult {
	result := rules.EvaluationResult{
		RuleName:  rule.DisplayName,
		Status:    rules.StatusGreen,
		Details:   []string{},
		Timestamp: time.Now(),
	}

	if rule.PercentageConfig == nil {
		result.Message = "Percentage config not specified"
		return result
	}

	// Get numerator metric
	numeratorMetric, exists := metrics.GetMetric(rule.PercentageConfig.Numerator)
	if !exists || len(numeratorMetric.Values) == 0 {
		result.Message = fmt.Sprintf("Numerator metric %s not found", rule.PercentageConfig.Numerator)
		return result
	}

	// Get denominator metric
	denominatorMetric, exists := metrics.GetMetric(rule.PercentageConfig.Denominator)
	if !exists || len(denominatorMetric.Values) == 0 {
		result.Message = fmt.Sprintf("Denominator metric %s not found", rule.PercentageConfig.Denominator)
		return result
	}

	numerator, _ := numeratorMetric.GetSingleValue()
	denominator, _ := denominatorMetric.GetSingleValue()

	if denominator == 0 {
		result.Status = rules.StatusGreen
		result.Message = rule.Messages.ZeroActivity
		if result.Message == "" {
			result.Message = "No activity yet (denominator is zero)"
		}
		return result
	}

	percentage := (numerator / denominator) * 100
	result.Value = percentage

	// Select thresholds based on load level
	thresholds := selectThresholds(rule, loadLevel)

	// Evaluate thresholds
	if percentage < thresholds.Low {
		result.Status = rules.StatusGreen
		result.Message = interpolate(rule.Messages.Green, percentage, map[string]interface{}{
			"numerator":   numerator,
			"denominator": denominator,
		})
	} else if percentage < thresholds.High {
		result.Status = rules.StatusYellow
		result.Message = interpolate(rule.Messages.Yellow, percentage, map[string]interface{}{
			"numerator":   numerator,
			"denominator": denominator,
		})
	} else {
		result.Status = rules.StatusRed
		result.Message = interpolate(rule.Messages.Red, percentage, map[string]interface{}{
			"numerator":   numerator,
			"denominator": denominator,
		})
	}

	return result
}
