package evaluator

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/stackrox/sensor-metrics-analyzer/internal/parser"
	"github.com/stackrox/sensor-metrics-analyzer/internal/rules"
)

// EvaluateHistogram evaluates a histogram rule
func EvaluateHistogram(rule rules.Rule, metrics parser.MetricsData, loadLevel rules.LoadLevel) rules.EvaluationResult {
	result := rules.EvaluationResult{
		RuleName:  rule.MetricName,
		Status:    rules.StatusGreen,
		Details:   []string{},
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
	result.Details = append(result.Details,
		fmt.Sprintf("p95: %.3f", p95),
		fmt.Sprintf("p99: %.3f", p99),
		fmt.Sprintf("count: %.0f", totalCount),
	)

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

// EvaluateHistogramInfOverflow evaluates all histogram metrics for +Inf bucket overflow
// This is a general rule that applies to any histogram metric
func EvaluateHistogramInfOverflow(metrics parser.MetricsData, loadLevel rules.LoadLevel) []rules.EvaluationResult {
	var results []rules.EvaluationResult

	// Get all histogram base names
	histogramBases := metrics.GetHistogramBaseNames()

	for _, baseName := range histogramBases {
		result := evaluateSingleHistogramInfOverflow(baseName, metrics)
		if result != nil {
			results = append(results, *result)
		}
	}

	return results
}

// evaluateSingleHistogramInfOverflow evaluates a single histogram metric for +Inf overflow
// Handles multiple label combinations (series) by evaluating each separately and reporting the worst case
func evaluateSingleHistogramInfOverflow(baseName string, metrics parser.MetricsData) *rules.EvaluationResult {
	result := &rules.EvaluationResult{
		RuleName:  baseName + " (+Inf overflow check)",
		Status:    rules.StatusGreen,
		Details:   []string{},
		Timestamp: time.Now(),
	}
	result.ReviewStatus = "Automatically generated rule; review by the code author"

	// Get histogram buckets
	bucketMetricName := baseName + "_bucket"
	bucketMetric, exists := metrics.GetMetric(bucketMetricName)
	if !exists || len(bucketMetric.Values) == 0 {
		return nil // Skip if no buckets found
	}

	// Group buckets by label combination (excluding "le" label)
	// Each label combination represents a separate time series
	seriesBuckets := make(map[string][]parser.MetricValue)
	for _, v := range bucketMetric.Values {
		// Create a key from all labels except "le"
		seriesKey := getSeriesKey(v.Labels)
		seriesBuckets[seriesKey] = append(seriesBuckets[seriesKey], v)
	}

	// Track worst case across all series
	var worstInfPercentage float64
	var worstInfObservations float64
	var worstTotalCount float64
	var worstHighestFiniteLe float64
	var worstStatus rules.Status

	hasAnyData := false

	// Evaluate each series separately
	for _, buckets := range seriesBuckets {
		// Find +Inf bucket and highest finite bucket for this series
		var infCount float64
		var highestFiniteLe float64
		var highestFiniteCount float64
		hasInf := false
		hasFinite := false

		for _, v := range buckets {
			if leStr, exists := v.Labels["le"]; exists {
				if leStr == "+Inf" {
					infCount = v.Value
					hasInf = true
				} else if le, err := strconv.ParseFloat(leStr, 64); err == nil {
					if !hasFinite || le > highestFiniteLe {
						highestFiniteLe = le
						highestFiniteCount = v.Value
						hasFinite = true
					}
				}
			}
		}

		if !hasInf || !hasFinite || infCount == 0 {
			continue
		}

		hasAnyData = true

		// Calculate the percentage of observations in +Inf bucket for this series
		// Observations in +Inf = totalCount - highestFiniteCount
		infObservations := infCount - highestFiniteCount
		if infObservations < 0 {
			// This can happen if there are data inconsistencies, skip this series
			continue
		}
		infPercentage := (infObservations / infCount) * 100.0

		// Track worst case
		if infPercentage > worstInfPercentage {
			worstInfPercentage = infPercentage
			worstInfObservations = infObservations
			worstTotalCount = infCount
			worstHighestFiniteLe = highestFiniteLe

			// Determine status for this series
			if infPercentage > 50.0 {
				worstStatus = rules.StatusRed
			} else if infPercentage > 25.0 {
				worstStatus = rules.StatusYellow
			} else {
				worstStatus = rules.StatusGreen
			}
		}
	}

	if !hasAnyData {
		return nil
	}

	result.Status = worstStatus
	if baseMetric, ok := metrics.GetMetric(baseName); ok && baseMetric.Help != "" {
		result.Details = append(result.Details, "Metric Description: "+baseMetric.Help)
	} else if bucketMetric, ok := metrics.GetMetric(baseName + "_bucket"); ok && bucketMetric.Help != "" {
		result.Details = append(result.Details, "Metric Description: "+bucketMetric.Help)
	}
	result.Details = append(result.Details,
		"Total Number of Observations: "+formatHumanNumber(worstTotalCount)+" unit",
		"Observations in +Inf bucket: "+formatHumanNumber(worstInfObservations)+" unit",
		"Percentage of observations in +Inf bucket: "+formatHumanNumber(worstInfPercentage)+" %",
		"Highest non-infinity bucket: "+formatHumanNumber(worstHighestFiniteLe)+" unit",
	)

	// Build message based on worst case
	result.Message = fmt.Sprintf("%s%% of observations in +Inf bucket (acceptable). Highest non-infinity bucket: %s",
		formatHumanNumber(worstInfPercentage), formatHumanNumber(worstHighestFiniteLe))
	if worstInfPercentage > 25.0 {
		result.Message = fmt.Sprintf("%s%% of observations are in +Inf bucket (%s out of %s). "+
			"This indicates the metric designer likely didn't expect processing durations to be so high. "+
			"Highest non-infinity bucket: %s",
			formatHumanNumber(worstInfPercentage),
			formatHumanNumber(worstInfObservations),
			formatHumanNumber(worstTotalCount),
			formatHumanNumber(worstHighestFiniteLe))
		result.PotentialActionUser = fmt.Sprintf("Further investigation is required to understand why values exceed %s. "+
			"Check if there are other alerts for this specific metric with more precise context.", formatHumanNumber(worstHighestFiniteLe))
		result.PotentialActionDeveloper = "Review code paths and metric instrumentation to confirm whether observed latencies are expected."
	}
	return result
}

// getSeriesKey creates a key from labels excluding "le" to group buckets by series
func getSeriesKey(labels map[string]string) string {
	var keys []string
	for k, v := range labels {
		if k != "le" {
			keys = append(keys, k+"="+v)
		}
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

func formatHumanNumber(value float64) string {
	raw := strconv.FormatFloat(value, 'f', 2, 64)
	sign := ""
	if strings.HasPrefix(raw, "-") {
		sign = "-"
		raw = strings.TrimPrefix(raw, "-")
	}
	parts := strings.SplitN(raw, ".", 2)
	intPart := parts[0]
	var fracPart string
	if len(parts) == 2 && parts[1] != "" {
		fracPart = parts[1]
	}
	var grouped strings.Builder
	for i, r := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			grouped.WriteString(" ")
		}
		grouped.WriteRune(r)
	}
	if fracPart != "" {
		return sign + grouped.String() + "." + fracPart
	}
	return sign + grouped.String()
}
