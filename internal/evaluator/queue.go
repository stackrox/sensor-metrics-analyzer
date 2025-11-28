package evaluator

import (
	"fmt"
	"time"

	"github.com/stackrox/sensor-metrics-analyzer/internal/parser"
	"github.com/stackrox/sensor-metrics-analyzer/internal/rules"
)

// EvaluateQueue evaluates a queue operations rule
func EvaluateQueue(rule rules.Rule, metrics parser.MetricsData, loadLevel rules.LoadLevel) rules.EvaluationResult {
	result := rules.EvaluationResult{
		RuleName:  rule.MetricName,
		Status:    rules.StatusGreen,
		Details:   make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	if rule.QueueConfig == nil {
		result.Message = "Queue config not specified"
		return result
	}

	metric, exists := metrics.GetMetric(rule.MetricName)
	if !exists || len(metric.Values) == 0 {
		result.Message = fmt.Sprintf("Metric %s not found", rule.MetricName)
		return result
	}

	// Group values by operation label
	valuesByOp := metric.GetValuesByLabel(rule.QueueConfig.OperationLabel)

	addValue := valuesByOp[rule.QueueConfig.AddValue]
	removeValue := valuesByOp[rule.QueueConfig.RemoveValue]

	diff := addValue - removeValue
	result.Value = diff

	// Select thresholds based on load level
	thresholds := selectThresholds(rule, loadLevel)

	// Evaluate thresholds
	if diff < thresholds.Low {
		result.Status = rules.StatusGreen
		result.Message = interpolate(rule.Messages.Green, diff, map[string]interface{}{
			"add":    addValue,
			"remove": removeValue,
			"diff":   diff,
		})
	} else if diff < thresholds.High {
		result.Status = rules.StatusYellow
		result.Message = interpolate(rule.Messages.Yellow, diff, map[string]interface{}{
			"add":    addValue,
			"remove": removeValue,
			"diff":   diff,
		})
	} else {
		result.Status = rules.StatusRed
		result.Message = interpolate(rule.Messages.Red, diff, map[string]interface{}{
			"add":    addValue,
			"remove": removeValue,
			"diff":   diff,
		})
	}

	return result
}
