package evaluator

import (
	"fmt"
	"sort"
	"time"

	"github.com/stackrox/sensor-metrics-analyzer/internal/parser"
	"github.com/stackrox/sensor-metrics-analyzer/internal/rules"
)

// EvaluateHistogram evaluates a histogram rule
func EvaluateHistogram(rule rules.Rule, metrics parser.MetricsData, loadLevel rules.LoadLevel) rules.EvaluationResult {
	result := rules.EvaluationResult{
		RuleName:  rule.MetricName,
		Status:    rules.StatusGreen,
		Details:   make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	// Get histogram buckets
	bucketMetricName := rule.MetricName + "_bucket"
	bucketMetric, exists := metrics.GetMetric(bucketMetricName)
	if !exists || len(bucketMetric.Values) == 0 {
		result.Message = fmt.Sprintf("Histogram buckets for %s not found", rule.MetricName)
		return result
	}

	buckets := bucketMetric.GetHistogramBuckets()
	if len(buckets) == 0 {
		result.Message = "No histogram buckets found"
		return result
	}

	// Sort buckets by le value
	sort.Slice(buckets, func(i, j int) bool {
		return buckets[i].Le < buckets[j].Le
	})

	// Get total count (from last bucket before +Inf)
	totalCount := buckets[len(buckets)-1].Count
	if totalCount == 0 {
		result.Status = rules.StatusGreen
		result.Message = "No histogram data yet"
		return result
	}

	// Calculate P95 and P99
	p95Threshold := totalCount * 0.95
	p99Threshold := totalCount * 0.99

	var p95, p99 float64
	for _, bucket := range buckets {
		if p95 == 0 && bucket.Count >= p95Threshold {
			p95 = bucket.Le
		}
		if p99 == 0 && bucket.Count >= p99Threshold {
			p99 = bucket.Le
		}
		if p95 > 0 && p99 > 0 {
			break
		}
	}

	result.Value = p95
	result.Details["p95"] = p95
	result.Details["p99"] = p99
	result.Details["count"] = totalCount

	// Select thresholds based on load level
	thresholds := selectThresholds(rule, loadLevel)

	// Evaluate thresholds based on P95
	if p95 < thresholds.P95Good {
		result.Status = rules.StatusGreen
		result.Message = interpolate(rule.Messages.Green, p95, map[string]interface{}{
			"p95": p95,
			"p99": p99,
		})
	} else if p95 < thresholds.P95Warn {
		result.Status = rules.StatusYellow
		result.Message = interpolate(rule.Messages.Yellow, p95, map[string]interface{}{
			"p95": p95,
			"p99": p99,
		})
	} else {
		result.Status = rules.StatusRed
		result.Message = interpolate(rule.Messages.Red, p95, map[string]interface{}{
			"p95": p95,
			"p99": p99,
		})
	}

	return result
}
