package loadlevel

import (
	"fmt"

	"github.com/stackrox/sensor-metrics-analyzer/internal/parser"
	"github.com/stackrox/sensor-metrics-analyzer/internal/rules"
)

// Detector evaluates cluster load level based on weighted metrics
type Detector struct {
	rules []rules.LoadDetectionRule
}

// NewDetector creates a new load level detector
func NewDetector(loadRules []rules.LoadDetectionRule) *Detector {
	return &Detector{
		rules: loadRules,
	}
}

// Detect evaluates the cluster load level from metrics
func (d *Detector) Detect(metrics parser.MetricsData) (rules.LoadLevel, error) {
	if len(d.rules) == 0 {
		// Default to medium if no rules configured
		return rules.LoadLevelMedium, nil
	}

	// Use the first load detection rule
	rule := d.rules[0]

	// Calculate weighted sum of metrics
	weightedSum := 0.0
	totalWeight := 0.0

	for _, metricDef := range rule.Metrics {
		metric, exists := metrics.GetMetric(metricDef.Source)
		if !exists {
			continue // Skip missing metrics
		}

		// Get the metric value (sum all values if multiple)
		value := metric.SumValues()
		weightedSum += value * metricDef.Weight
		totalWeight += metricDef.Weight
	}

	if totalWeight == 0 {
		// No metrics found, default to medium
		return rules.LoadLevelMedium, nil
	}

	// Normalize by total weight (optional, but helps with consistency)
	normalizedValue := weightedSum / totalWeight

	// Find matching threshold
	for _, threshold := range rule.Thresholds {
		matches := true
		if threshold.MinValue > 0 && normalizedValue < threshold.MinValue {
			matches = false
		}
		if threshold.MaxValue > 0 && normalizedValue >= threshold.MaxValue {
			matches = false
		}
		if matches {
			return threshold.Level, nil
		}
	}

	// Default to medium if no threshold matches
	return rules.LoadLevelMedium, nil
}

// DetectWithOverride allows overriding the detected load level
func DetectWithOverride(metrics parser.MetricsData, detector *Detector, override rules.LoadLevel) (rules.LoadLevel, error) {
	if override != "" {
		// Validate override
		validLevels := []rules.LoadLevel{rules.LoadLevelLow, rules.LoadLevelMedium, rules.LoadLevelHigh}
		for _, level := range validLevels {
			if override == level {
				return override, nil
			}
		}
		return "", fmt.Errorf("invalid load level override: %s", override)
	}

	return detector.Detect(metrics)
}
