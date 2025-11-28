package parser

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// HistogramBucket represents a histogram bucket
type HistogramBucket struct {
	Le    float64 // Less than or equal to value
	Count float64 // Cumulative count
}

// MetricValue represents a single metric data point
type MetricValue struct {
	Labels map[string]string
	Value  float64
}

// Metric represents a Prometheus metric with metadata
type Metric struct {
	Name   string
	Help   string
	Type   string
	Values []MetricValue
}

// MetricsData is a map of metric names to their data
type MetricsData map[string]*Metric

var (
	helpRegex         = regexp.MustCompile(`^# HELP\s+(\S+)\s+(.*)$`)
	typeRegex         = regexp.MustCompile(`^# TYPE\s+(\S+)\s+(\S+)$`)
	metricRegex       = regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_:]*)\{([^}]*)\}\s+(.+)$`)
	simpleMetricRegex = regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_:]*)\s+(.+)$`)
)

// ParseFile parses a Prometheus metrics file
func ParseFile(filepath string) (MetricsData, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	metrics := make(MetricsData)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "# EOF") {
			continue
		}

		// Parse HELP
		if matches := helpRegex.FindStringSubmatch(line); matches != nil {
			metricName := matches[1]
			helpText := matches[2]
			if metrics[metricName] == nil {
				metrics[metricName] = &Metric{Name: metricName}
			}
			metrics[metricName].Help = helpText
			continue
		}

		// Parse TYPE
		if matches := typeRegex.FindStringSubmatch(line); matches != nil {
			metricName := matches[1]
			metricType := matches[2]
			if metrics[metricName] == nil {
				metrics[metricName] = &Metric{Name: metricName}
			}
			metrics[metricName].Type = metricType
			continue
		}

		// Parse metric with labels
		if matches := metricRegex.FindStringSubmatch(line); matches != nil {
			metricName := matches[1]
			labelsStr := matches[2]
			valueStr := matches[3]

			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				continue // Skip non-numeric values
			}

			labels := parseLabels(labelsStr)

			if metrics[metricName] == nil {
				metrics[metricName] = &Metric{Name: metricName}
			}
			metrics[metricName].Values = append(metrics[metricName].Values, MetricValue{
				Labels: labels,
				Value:  value,
			})
			continue
		}

		// Parse simple metric (no labels)
		if matches := simpleMetricRegex.FindStringSubmatch(line); matches != nil {
			metricName := matches[1]
			valueStr := matches[2]

			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				continue
			}

			if metrics[metricName] == nil {
				metrics[metricName] = &Metric{Name: metricName}
			}
			metrics[metricName].Values = append(metrics[metricName].Values, MetricValue{
				Labels: make(map[string]string),
				Value:  value,
			})
		}
	}

	return metrics, scanner.Err()
}

// parseLabels parses label string like: key1="value1",key2="value2"
func parseLabels(labelsStr string) map[string]string {
	labels := make(map[string]string)

	if labelsStr == "" {
		return labels
	}

	pairs := strings.Split(labelsStr, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), `"`)
			labels[key] = value
		}
	}

	return labels
}

// GetMetric retrieves a metric by name
func (md MetricsData) GetMetric(name string) (*Metric, bool) {
	metric, exists := md[name]
	return metric, exists
}

// GetSingleValue gets the first value of a metric (for gauges/counters)
func (m *Metric) GetSingleValue() (float64, bool) {
	if len(m.Values) == 0 {
		return 0, false
	}
	return m.Values[0].Value, true
}

// GetValuesByLabel gets values grouped by a specific label
func (m *Metric) GetValuesByLabel(labelKey string) map[string]float64 {
	result := make(map[string]float64)
	for _, v := range m.Values {
		if labelValue, exists := v.Labels[labelKey]; exists {
			result[labelValue] = v.Value
		}
	}
	return result
}

// SumValues sums all values of a metric
func (m *Metric) SumValues() float64 {
	sum := 0.0
	for _, v := range m.Values {
		sum += v.Value
	}
	return sum
}

// GetHistogramBuckets returns buckets for a histogram metric
func (m *Metric) GetHistogramBuckets() []HistogramBucket {
	var buckets []HistogramBucket
	for _, v := range m.Values {
		if leStr, exists := v.Labels["le"]; exists && leStr != "+Inf" {
			if le, err := strconv.ParseFloat(leStr, 64); err == nil {
				buckets = append(buckets, HistogramBucket{
					Le:    le,
					Count: v.Value,
				})
			}
		}
	}
	return buckets
}

// GetHistogramSum returns the _sum value for a histogram metric
func (md MetricsData) GetHistogramSum(baseName string) (float64, bool) {
	sumMetric := baseName + "_sum"
	metric, exists := md.GetMetric(sumMetric)
	if !exists || len(metric.Values) == 0 {
		return 0, false
	}
	return metric.GetSingleValue()
}

// GetHistogramCount returns the _count value for a histogram metric
func (md MetricsData) GetHistogramCount(baseName string) (float64, bool) {
	countMetric := baseName + "_count"
	metric, exists := md.GetMetric(countMetric)
	if !exists || len(metric.Values) == 0 {
		return 0, false
	}
	return metric.GetSingleValue()
}

// DetectACSVersion attempts to detect ACS version from metrics
func (md MetricsData) DetectACSVersion() (string, bool) {
	// Try common version metrics
	versionMetrics := []string{
		"rox_sensor_version_info",
		"rox_central_version_info",
		"rox_version",
	}

	for _, metricName := range versionMetrics {
		metric, exists := md.GetMetric(metricName)
		if exists {
			// Look for version in labels or values
			for _, v := range metric.Values {
				// Check labels for version
				if version, ok := v.Labels["version"]; ok && version != "" {
					return version, true
				}
				if version, ok := v.Labels["rox_version"]; ok && version != "" {
					return version, true
				}
			}
		}
	}

	return "", false
}
