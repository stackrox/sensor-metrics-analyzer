package evaluator

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/stackrox/sensor-metrics-analyzer/internal/parser"
	"github.com/stackrox/sensor-metrics-analyzer/internal/rules"
)

const (
	infOverflowYellowThresholdPercent = 25.0
	infOverflowRedThresholdPercent    = 50.0
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

	// Calculate P50, P75, P95 and P99
	p50Threshold := totalCount * 0.50
	p75Threshold := totalCount * 0.75
	p95Threshold := totalCount * 0.95
	p99Threshold := totalCount * 0.99

	var p50, p75, p95, p99 float64
	for _, bucket := range buckets {
		if p50 == 0 && bucket.Count >= p50Threshold {
			p50 = bucket.Le
		}
		if p75 == 0 && bucket.Count >= p75Threshold {
			p75 = bucket.Le
		}
		if p95 == 0 && bucket.Count >= p95Threshold {
			p95 = bucket.Le
		}
		if p99 == 0 && bucket.Count >= p99Threshold {
			p99 = bucket.Le
		}
		if p50 > 0 && p75 > 0 && p95 > 0 && p99 > 0 {
			break
		}
	}

	result.Value = p95
	unit := ""
	if rule.HistogramConfig != nil {
		unit = strings.TrimSpace(rule.HistogramConfig.Unit)
	}
	result.Details = append(result.Details,
		fmt.Sprintf("p50: %s (i.e., 50%% of the observations are below this value)", formatHistogramValue(p50, unit)),
		fmt.Sprintf("p75: %s (i.e., 75%% of the observations are below this value)", formatHistogramValue(p75, unit)),
		fmt.Sprintf("p95: %s (i.e., 95%% of the observations are below this value)", formatHistogramValue(p95, unit)),
		fmt.Sprintf("p99: %s (i.e., 99%% of the observations are below this value)", formatHistogramValue(p99, unit)),
		fmt.Sprintf("count: %s", formatHumanInteger(totalCount)),
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
func EvaluateHistogramInfOverflow(metrics parser.MetricsData) []rules.EvaluationResult {
	var results []rules.EvaluationResult

	// Get all histogram base names
	histogramBases := metrics.GetHistogramBaseNames()

	for _, baseName := range histogramBases {
		results = append(results, evaluateSingleHistogramInfOverflow(baseName, metrics)...)
	}

	return results
}

// evaluateSingleHistogramInfOverflow evaluates a single histogram metric for +Inf overflow
// Handles multiple label combinations (series) by emitting one result per problematic series.
func evaluateSingleHistogramInfOverflow(baseName string, metrics parser.MetricsData) []rules.EvaluationResult {
	// Get histogram buckets
	bucketMetricName := baseName + "_bucket"
	bucketMetric, exists := metrics.GetMetric(bucketMetricName)
	if !exists || len(bucketMetric.Values) == 0 {
		return nil
	}
	metricHelp := resolveMetricHelp(baseName, metrics)

	// Group buckets by label combination (excluding "le" label)
	// Each label combination represents a separate time series
	seriesBuckets := make(map[string][]parser.MetricValue)
	seriesLabels := make(map[string]map[string]string)
	for _, v := range bucketMetric.Values {
		// Create a key from all labels except "le"
		seriesKey := getSeriesKey(v.Labels)
		seriesBuckets[seriesKey] = append(seriesBuckets[seriesKey], v)
		if _, seen := seriesLabels[seriesKey]; !seen {
			seriesLabels[seriesKey] = extractSeriesLabels(v.Labels)
		}
	}

	type seriesEvaluation struct {
		labels          map[string]string
		infPercentage   float64
		infObservations float64
		totalCount      float64
		highestFiniteLe float64
		status          rules.Status
	}
	var evaluations []seriesEvaluation

	hasAnyData := false
	var worstOverall *seriesEvaluation
	var worstSeriesKey string

	// Evaluate each series separately
	for seriesKey, buckets := range seriesBuckets {
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

		status := rules.StatusGreen
		if infPercentage > infOverflowRedThresholdPercent {
			status = rules.StatusRed
		} else if infPercentage > infOverflowYellowThresholdPercent {
			status = rules.StatusYellow
		}

		eval := seriesEvaluation{
			labels:          seriesLabels[seriesKey],
			infPercentage:   infPercentage,
			infObservations: infObservations,
			totalCount:      infCount,
			highestFiniteLe: highestFiniteLe,
			status:          status,
		}
		evaluations = append(evaluations, eval)

		if worstOverall == nil || infPercentage > worstOverall.infPercentage ||
			(infPercentage == worstOverall.infPercentage && seriesKey < worstSeriesKey) {
			worstCopy := eval
			worstOverall = &worstCopy
			worstSeriesKey = seriesKey
		}
	}

	if !hasAnyData || len(evaluations) == 0 {
		return nil
	}

	var results []rules.EvaluationResult
	for _, eval := range evaluations {
		if eval.status == rules.StatusGreen {
			continue
		}
		result := rules.EvaluationResult{
			RuleName:     formatSeriesRuleName(baseName, eval.labels),
			Status:       eval.status,
			MetricHelp:   metricHelp,
			Details:      []string{},
			Timestamp:    time.Now(),
			ReviewStatus: "Automatically generated rule; reviewed by the code author at the time of implementation.",
			PotentialActionUser: fmt.Sprintf("Further investigation is required to understand why values exceed %s. "+
				"Check if there are other alerts for this specific metric with more precise context.", formatHumanNumber(eval.highestFiniteLe)),
			PotentialActionDeveloper: "Review code paths and metric instrumentation to confirm whether observed latencies are expected.",
		}
		result.Details = append(result.Details,
			"Total Number of Observations: "+formatHumanNumber(eval.totalCount),
			"Observations in +Inf bucket: "+formatHumanNumber(eval.infObservations),
			"Percentage of observations in +Inf bucket: "+formatHumanNumber(eval.infPercentage)+" %",
			"Highest non-infinity bucket: "+formatHumanNumber(eval.highestFiniteLe)+" unit",
		)
		result.Message = fmt.Sprintf("%s%% of observations are in +Inf bucket (%s out of %s). "+
			"This indicates the metric designer likely didn't expect the values to be so high. "+
			"This may indicate a problem in the system or a problem with metrics design. "+
			"Highest non-infinity bucket: %s",
			formatHumanNumber(eval.infPercentage),
			formatHumanNumber(eval.infObservations),
			formatHumanNumber(eval.totalCount),
			formatHumanNumber(eval.highestFiniteLe))
		results = append(results, result)
	}

	if len(results) > 0 {
		sort.Slice(results, func(i, j int) bool {
			if results[i].Status != results[j].Status {
				return results[i].Status == rules.StatusRed
			}
			return results[i].RuleName < results[j].RuleName
		})
		return results
	}

	greenResult := rules.EvaluationResult{
		RuleName:     baseName + " (+Inf overflow check)",
		Status:       rules.StatusGreen,
		MetricHelp:   metricHelp,
		Details:      []string{},
		Timestamp:    time.Now(),
		ReviewStatus: "Automatically generated rule; reviewed by the code author at the time of implementation.",
	}
	greenResult.Details = append(greenResult.Details,
		"Total Number of Observations: "+formatHumanNumber(worstOverall.totalCount),
		"Observations in +Inf bucket: "+formatHumanNumber(worstOverall.infObservations),
		"Percentage of observations in +Inf bucket: "+formatHumanNumber(worstOverall.infPercentage)+" %",
		"Highest non-infinity bucket: "+formatHumanNumber(worstOverall.highestFiniteLe)+" unit",
	)
	greenResult.Message = fmt.Sprintf("%s%% of observations in +Inf bucket (acceptable). Highest non-infinity bucket: %s",
		formatHumanNumber(worstOverall.infPercentage), formatHumanNumber(worstOverall.highestFiniteLe))
	return []rules.EvaluationResult{greenResult}
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

func extractSeriesLabels(labels map[string]string) map[string]string {
	result := make(map[string]string)
	for key, value := range labels {
		if key == "le" {
			continue
		}
		result[key] = value
	}
	return result
}

func formatSeriesRuleName(baseName string, labels map[string]string) string {
	if len(labels) == 0 {
		return baseName + " (+Inf overflow check)"
	}
	var keys []string
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var parts []string
	for _, key := range keys {
		parts = append(parts, key+"=\""+labels[key]+"\"")
	}
	return baseName + "{" + strings.Join(parts, ",") + "} (+Inf overflow check)"
}

func resolveMetricHelp(baseName string, metrics parser.MetricsData) string {
	if baseMetric, ok := metrics.GetMetric(baseName); ok && baseMetric.Help != "" {
		return baseMetric.Help
	}
	if bucketMetric, ok := metrics.GetMetric(baseName + "_bucket"); ok && bucketMetric.Help != "" {
		return bucketMetric.Help
	}
	return ""
}

func formatHumanNumber(value float64) string {
	return formatHumanNumberWithPrecision(value, 2)
}

func formatHumanInteger(value float64) string {
	return formatHumanNumberWithPrecision(value, 0)
}

func formatHumanNumberWithPrecision(value float64, precision int) string {
	raw := strconv.FormatFloat(value, 'f', precision, 64)
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

func formatHistogramValue(value float64, unit string) string {
	formatted := formatHumanNumber(value)
	if value == math.Trunc(value) {
		formatted = formatHumanInteger(value)
	}
	unit = strings.TrimSpace(strings.ToLower(unit))
	if unit == "" {
		return formatted
	}
	switch unit {
	case "milliseconds", "millisecond", "ms":
		return formatted + " ms"
	case "seconds", "second", "s":
		return formatted + " s"
	}
	return formatted + " " + unit
}
