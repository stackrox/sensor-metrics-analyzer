package loadlevel

import (
	"testing"

	"github.com/stackrox/sensor-metrics-analyzer/internal/parser"
	"github.com/stackrox/sensor-metrics-analyzer/internal/rules"
)

func TestDetect(t *testing.T) {
	tests := map[string]struct {
		rule      rules.LoadDetectionRule
		metrics   parser.MetricsData
		wantLevel rules.LoadLevel
		wantError bool
	}{
		"should detect low load level correctly": {
			rule: rules.LoadDetectionRule{
				Metrics: []rules.LoadDetectionMetric{
					{Source: "containers", Weight: 1.0},
					{Source: "pods", Weight: 0.5},
				},
				Thresholds: []rules.LoadDetectionThreshold{
					{Level: rules.LoadLevelLow, MaxValue: 100},
					{Level: rules.LoadLevelMedium, MinValue: 100, MaxValue: 500},
					{Level: rules.LoadLevelHigh, MinValue: 500},
				},
			},
			metrics: parser.MetricsData{
				"containers": &parser.Metric{
					Name: "containers",
					Values: []parser.MetricValue{
						{Value: 50, Labels: make(map[string]string)},
					},
				},
				"pods": &parser.Metric{
					Name: "pods",
					Values: []parser.MetricValue{
						{Value: 20, Labels: make(map[string]string)},
					},
				},
			},
			wantLevel: rules.LoadLevelLow,
			wantError: false,
		},
		"should detect medium load level correctly": {
			rule: rules.LoadDetectionRule{
				Metrics: []rules.LoadDetectionMetric{
					{Source: "containers", Weight: 1.0},
					{Source: "pods", Weight: 0.5},
				},
				Thresholds: []rules.LoadDetectionThreshold{
					{Level: rules.LoadLevelLow, MaxValue: 100},
					{Level: rules.LoadLevelMedium, MinValue: 100, MaxValue: 500},
					{Level: rules.LoadLevelHigh, MinValue: 500},
				},
			},
			metrics: parser.MetricsData{
				"containers": &parser.Metric{
					Name: "containers",
					Values: []parser.MetricValue{
						{Value: 200, Labels: make(map[string]string)},
					},
				},
				"pods": &parser.Metric{
					Name: "pods",
					Values: []parser.MetricValue{
						{Value: 50, Labels: make(map[string]string)},
					},
				},
			},
			wantLevel: rules.LoadLevelMedium,
			wantError: false,
		},
		"should detect high load level correctly": {
			rule: rules.LoadDetectionRule{
				Metrics: []rules.LoadDetectionMetric{
					{Source: "containers", Weight: 1.0},
					{Source: "pods", Weight: 0.5},
				},
				Thresholds: []rules.LoadDetectionThreshold{
					{Level: rules.LoadLevelLow, MaxValue: 100},
					{Level: rules.LoadLevelMedium, MinValue: 100, MaxValue: 500},
					{Level: rules.LoadLevelHigh, MinValue: 500},
				},
			},
			metrics: parser.MetricsData{
				"containers": &parser.Metric{
					Name: "containers",
					Values: []parser.MetricValue{
						{Value: 800, Labels: make(map[string]string)},
					},
				},
				"pods": &parser.Metric{
					Name: "pods",
					Values: []parser.MetricValue{
						{Value: 200, Labels: make(map[string]string)},
					},
				},
			},
			wantLevel: rules.LoadLevelHigh,
			// weightedSum = (800 * 1.0) + (200 * 0.5) = 900
			// totalWeight = 1.5
			// normalizedValue = 900 / 1.5 = 600 → HIGH
			wantError: false,
		},
		"should default to medium when no metrics found": {
			rule: rules.LoadDetectionRule{
				Metrics: []rules.LoadDetectionMetric{
					{Source: "missing_metric", Weight: 1.0},
				},
				Thresholds: []rules.LoadDetectionThreshold{
					{Level: rules.LoadLevelLow, MaxValue: 100},
					{Level: rules.LoadLevelMedium, MinValue: 100, MaxValue: 500},
					{Level: rules.LoadLevelHigh, MinValue: 500},
				},
			},
			metrics:   parser.MetricsData{},
			wantLevel: rules.LoadLevelMedium,
			wantError: false,
		},
		"should calculate weighted average correctly": {
			rule: rules.LoadDetectionRule{
				Metrics: []rules.LoadDetectionMetric{
					{Source: "containers", Weight: 1.0},
					{Source: "pods", Weight: 0.5},
				},
				Thresholds: []rules.LoadDetectionThreshold{
					{Level: rules.LoadLevelLow, MaxValue: 100},
					{Level: rules.LoadLevelMedium, MinValue: 100, MaxValue: 500},
					{Level: rules.LoadLevelHigh, MinValue: 500},
				},
			},
			metrics: parser.MetricsData{
				"containers": &parser.Metric{
					Name: "containers",
					Values: []parser.MetricValue{
						{Value: 200, Labels: make(map[string]string)},
					},
				},
				"pods": &parser.Metric{
					Name: "pods",
					Values: []parser.MetricValue{
						{Value: 100, Labels: make(map[string]string)},
					},
				},
			},
			wantLevel: rules.LoadLevelMedium,
			// weightedSum = (200 * 1.0) + (100 * 0.5) = 250
			// totalWeight = 1.5
			// normalizedValue = 250 / 1.5 = 166.67 → MEDIUM
			wantError: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			detector := NewDetector([]rules.LoadDetectionRule{tt.rule})
			level, err := detector.Detect(tt.metrics)

			if (err != nil) != tt.wantError {
				t.Errorf("Detect() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if level != tt.wantLevel {
				t.Errorf("Detect() level = %v, want %v", level, tt.wantLevel)
			}
		})
	}
}

func TestDetectWithOverride(t *testing.T) {
	tests := map[string]struct {
		metrics   parser.MetricsData
		override  rules.LoadLevel
		wantLevel rules.LoadLevel
		wantError bool
	}{
		"should return override when provided": {
			metrics:   parser.MetricsData{},
			override:  rules.LoadLevelHigh,
			wantLevel: rules.LoadLevelHigh,
			wantError: false,
		},
		"should detect when no override provided": {
			metrics: parser.MetricsData{
				"containers": &parser.Metric{
					Name: "containers",
					Values: []parser.MetricValue{
						{Value: 200, Labels: make(map[string]string)},
					},
				},
			},
			override:  "",
			wantLevel: rules.LoadLevelMedium, // Default when no rules
			wantError: false,
		},
		"should return error for invalid override": {
			metrics:   parser.MetricsData{},
			override:  rules.LoadLevel("invalid"),
			wantError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			detector := NewDetector([]rules.LoadDetectionRule{})
			level, err := DetectWithOverride(tt.metrics, detector, tt.override)

			if (err != nil) != tt.wantError {
				t.Errorf("DetectWithOverride() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && level != tt.wantLevel {
				t.Errorf("DetectWithOverride() level = %v, want %v", level, tt.wantLevel)
			}
		})
	}
}
