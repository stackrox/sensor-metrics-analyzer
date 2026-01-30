package evaluator

import (
	"testing"

	"github.com/stackrox/sensor-metrics-analyzer/internal/parser"
	"github.com/stackrox/sensor-metrics-analyzer/internal/rules"
)

func TestGaugeThresholdsGeneric(t *testing.T) {
	tests := []struct {
		name     string
		rule     rules.Rule
		value    float64
		expected rules.Status
	}{
		{
			name: "higher is worse below low is green",
			rule: rules.Rule{
				RuleType:   rules.RuleTypeGauge,
				MetricName: "test_metric",
				Thresholds: rules.Thresholds{
					Low:           100,
					High:          200,
					HigherIsWorse: true,
				},
				Messages: rules.Messages{
					Green:  "green",
					Yellow: "yellow",
					Red:    "red",
				},
			},
			value:    10,
			expected: rules.StatusGreen,
		},
		{
			name: "higher is worse between low/high is yellow",
			rule: rules.Rule{
				RuleType:   rules.RuleTypeGauge,
				MetricName: "test_metric",
				Thresholds: rules.Thresholds{
					Low:           100,
					High:          200,
					HigherIsWorse: true,
				},
				Messages: rules.Messages{
					Green:  "green",
					Yellow: "yellow",
					Red:    "red",
				},
			},
			value:    150,
			expected: rules.StatusYellow,
		},
		{
			name: "higher is worse at/above high is red",
			rule: rules.Rule{
				RuleType:   rules.RuleTypeGauge,
				MetricName: "test_metric",
				Thresholds: rules.Thresholds{
					Low:           100,
					High:          200,
					HigherIsWorse: true,
				},
				Messages: rules.Messages{
					Green:  "green",
					Yellow: "yellow",
					Red:    "red",
				},
			},
			value:    200,
			expected: rules.StatusRed,
		},
		{
			name: "lower is worse above high is green",
			rule: rules.Rule{
				RuleType:   rules.RuleTypeGauge,
				MetricName: "test_metric",
				Thresholds: rules.Thresholds{
					Low:           100,
					High:          200,
					HigherIsWorse: false,
				},
				Messages: rules.Messages{
					Green:  "green",
					Yellow: "yellow",
					Red:    "red",
				},
			},
			value:    250,
			expected: rules.StatusGreen,
		},
		{
			name: "lower is worse between low/high is yellow",
			rule: rules.Rule{
				RuleType:   rules.RuleTypeGauge,
				MetricName: "test_metric",
				Thresholds: rules.Thresholds{
					Low:           100,
					High:          200,
					HigherIsWorse: false,
				},
				Messages: rules.Messages{
					Green:  "green",
					Yellow: "yellow",
					Red:    "red",
				},
			},
			value:    150,
			expected: rules.StatusYellow,
		},
		{
			name: "lower is worse below low is red",
			rule: rules.Rule{
				RuleType:   rules.RuleTypeGauge,
				MetricName: "test_metric",
				Thresholds: rules.Thresholds{
					Low:           100,
					High:          200,
					HigherIsWorse: false,
				},
				Messages: rules.Messages{
					Green:  "green",
					Yellow: "yellow",
					Red:    "red",
				},
			},
			value:    50,
			expected: rules.StatusRed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := parser.MetricsData{
				"test_metric": {
					Name: "test_metric",
					Values: []parser.MetricValue{
						{Labels: map[string]string{}, Value: tt.value},
					},
				},
			}

			result := EvaluateGauge(tt.rule, metrics, rules.LoadLevelMedium)
			if result.Status != tt.expected {
				t.Fatalf("expected %s, got %s (value %.2f)", tt.expected, result.Status, tt.value)
			}
		})
	}
}
