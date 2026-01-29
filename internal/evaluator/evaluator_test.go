package evaluator

import (
	"strings"
	"testing"

	"github.com/stackrox/sensor-metrics-analyzer/internal/parser"
	"github.com/stackrox/sensor-metrics-analyzer/internal/rules"
)

func TestEvaluateGauge(t *testing.T) {
	tests := map[string]struct {
		rule       rules.Rule
		metrics    parser.MetricsData
		loadLevel  rules.LoadLevel
		wantStatus rules.Status
		wantError  bool
	}{
		"should return green status when value below low threshold": {
			rule: rules.Rule{
				RuleType:   rules.RuleTypeGauge,
				MetricName: "test_metric",
				Thresholds: rules.Thresholds{
					Low:           50,
					High:          100,
					HigherIsWorse: true,
				},
				Messages: rules.Messages{
					Green: "Value {value:.0f} is healthy",
				},
			},
			metrics: parser.MetricsData{
				"test_metric": &parser.Metric{
					Name: "test_metric",
					Values: []parser.MetricValue{
						{Value: 30, Labels: make(map[string]string)},
					},
				},
			},
			loadLevel:  rules.LoadLevelMedium,
			wantStatus: rules.StatusGreen,
		},
		"should return yellow status when value between thresholds": {
			rule: rules.Rule{
				RuleType:   rules.RuleTypeGauge,
				MetricName: "test_metric",
				Thresholds: rules.Thresholds{
					Low:           50,
					High:          100,
					HigherIsWorse: true,
				},
				Messages: rules.Messages{
					Yellow: "Value {value:.0f} is elevated",
				},
			},
			metrics: parser.MetricsData{
				"test_metric": &parser.Metric{
					Name: "test_metric",
					Values: []parser.MetricValue{
						{Value: 75, Labels: make(map[string]string)},
					},
				},
			},
			loadLevel:  rules.LoadLevelMedium,
			wantStatus: rules.StatusYellow,
		},
		"should return red status when value above high threshold": {
			rule: rules.Rule{
				RuleType:   rules.RuleTypeGauge,
				MetricName: "test_metric",
				Thresholds: rules.Thresholds{
					Low:           50,
					High:          100,
					HigherIsWorse: true,
				},
				Messages: rules.Messages{
					Red: "Value {value:.0f} is critical",
				},
			},
			metrics: parser.MetricsData{
				"test_metric": &parser.Metric{
					Name: "test_metric",
					Values: []parser.MetricValue{
						{Value: 150, Labels: make(map[string]string)},
					},
				},
			},
			loadLevel:  rules.LoadLevelMedium,
			wantStatus: rules.StatusRed,
		},
		"should use load-aware thresholds when available": {
			rule: rules.Rule{
				RuleType:   rules.RuleTypeGauge,
				MetricName: "test_metric",
				Thresholds: rules.Thresholds{
					Low:           50,
					High:          100,
					HigherIsWorse: true,
				},
				LoadLevelThresholds: &rules.LoadLevelThresholds{
					High: &rules.Thresholds{
						Low:           100,
						High:          200,
						HigherIsWorse: true,
					},
				},
				Messages: rules.Messages{
					Green: "Value {value:.0f} is healthy",
				},
			},
			metrics: parser.MetricsData{
				"test_metric": &parser.Metric{
					Name: "test_metric",
					Values: []parser.MetricValue{
						{Value: 120, Labels: make(map[string]string)},
					},
				},
			},
			loadLevel:  rules.LoadLevelHigh,
			wantStatus: rules.StatusYellow, // 120 is between 100 and 200 (yellow zone)
		},
		"should handle zero-check rule correctly": {
			rule: rules.Rule{
				RuleType:   rules.RuleTypeGauge,
				MetricName: "test_metric",
				Thresholds: rules.Thresholds{
					Low:           0,
					High:          0,
					HigherIsWorse: false,
				},
				Messages: rules.Messages{
					Green: "Value {value:.0f} exists",
					Red:   "Zero value detected",
				},
			},
			metrics: parser.MetricsData{
				"test_metric": &parser.Metric{
					Name: "test_metric",
					Values: []parser.MetricValue{
						{Value: 5, Labels: make(map[string]string)},
					},
				},
			},
			loadLevel:  rules.LoadLevelMedium,
			wantStatus: rules.StatusGreen,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := EvaluateGauge(tt.rule, tt.metrics, tt.loadLevel)

			if result.Status != tt.wantStatus {
				t.Errorf("EvaluateGauge() status = %v, want %v", result.Status, tt.wantStatus)
			}

			if result.RuleName != tt.rule.MetricName {
				t.Errorf("EvaluateGauge() ruleName = %v, want %v", result.RuleName, tt.rule.MetricName)
			}
		})
	}
}

func TestEvaluatePercentage(t *testing.T) {
	tests := map[string]struct {
		rule       rules.Rule
		metrics    parser.MetricsData
		loadLevel  rules.LoadLevel
		wantStatus rules.Status
	}{
		"should calculate percentage correctly and return green": {
			rule: rules.Rule{
				RuleType: rules.RuleTypePercentage,
				PercentageConfig: &rules.PercentageConfig{
					Numerator:   "numerator_metric",
					Denominator: "denominator_metric",
				},
				Thresholds: rules.Thresholds{
					Low:  2.0,
					High: 10.0,
				},
				Messages: rules.Messages{
					Green: "Percentage: {value:.1f}%",
				},
			},
			metrics: parser.MetricsData{
				"numerator_metric": &parser.Metric{
					Name: "numerator_metric",
					Values: []parser.MetricValue{
						{Value: 5, Labels: make(map[string]string)},
					},
				},
				"denominator_metric": &parser.Metric{
					Name: "denominator_metric",
					Values: []parser.MetricValue{
						{Value: 100, Labels: make(map[string]string)},
					},
				},
			},
			loadLevel:  rules.LoadLevelMedium,
			wantStatus: rules.StatusYellow, // 5/100 = 5%, which is between 2% and 10% (yellow zone)
		},
		"should return green when denominator is zero": {
			rule: rules.Rule{
				RuleType: rules.RuleTypePercentage,
				PercentageConfig: &rules.PercentageConfig{
					Numerator:   "numerator_metric",
					Denominator: "denominator_metric",
				},
				Thresholds: rules.Thresholds{},
				Messages: rules.Messages{
					ZeroActivity: "No activity yet",
				},
			},
			metrics: parser.MetricsData{
				"numerator_metric": &parser.Metric{
					Name: "numerator_metric",
					Values: []parser.MetricValue{
						{Value: 0, Labels: make(map[string]string)},
					},
				},
				"denominator_metric": &parser.Metric{
					Name: "denominator_metric",
					Values: []parser.MetricValue{
						{Value: 0, Labels: make(map[string]string)},
					},
				},
			},
			loadLevel:  rules.LoadLevelMedium,
			wantStatus: rules.StatusGreen,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := EvaluatePercentage(tt.rule, tt.metrics, tt.loadLevel)

			if result.Status != tt.wantStatus {
				t.Errorf("EvaluatePercentage() status = %v, want %v", result.Status, tt.wantStatus)
			}
		})
	}
}

func TestEvaluateQueue(t *testing.T) {
	tests := map[string]struct {
		rule       rules.Rule
		metrics    parser.MetricsData
		loadLevel  rules.LoadLevel
		wantStatus rules.Status
	}{
		"should calculate queue balance correctly": {
			rule: rules.Rule{
				RuleType:   rules.RuleTypeQueue,
				MetricName: "queue_operations",
				QueueConfig: &rules.QueueConfig{
					OperationLabel: "Operation",
					AddValue:       "Add",
					RemoveValue:    "Remove",
				},
				Thresholds: rules.Thresholds{
					Low:  10,
					High: 100,
				},
				Messages: rules.Messages{
					Green: "Queue balanced: Diff={diff:.0f}",
				},
			},
			metrics: parser.MetricsData{
				"queue_operations": &parser.Metric{
					Name: "queue_operations",
					Values: []parser.MetricValue{
						{Labels: map[string]string{"Operation": "Add"}, Value: 1000},
						{Labels: map[string]string{"Operation": "Remove"}, Value: 950},
					},
				},
			},
			loadLevel:  rules.LoadLevelMedium,
			wantStatus: rules.StatusYellow, // Diff = 50, which is between 10 and 100 (yellow zone)
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := EvaluateQueue(tt.rule, tt.metrics, tt.loadLevel)

			if result.Status != tt.wantStatus {
				t.Errorf("EvaluateQueue() status = %v, want %v", result.Status, tt.wantStatus)
			}
		})
	}
}

func TestFilterRulesByVersion(t *testing.T) {
	tests := map[string]struct {
		rules      []rules.Rule
		acsVersion string
		wantCount  int
	}{
		"should filter rules by ACS version correctly": {
			rules: []rules.Rule{
				{
					RuleType:    rules.RuleTypeGauge,
					MetricName:  "metric1",
					ACSVersions: []string{"4.7+", "4.8+"},
				},
				{
					RuleType:    rules.RuleTypeGauge,
					MetricName:  "metric2",
					ACSVersions: []string{"4.9+"},
				},
				{
					RuleType:   rules.RuleTypeGauge,
					MetricName: "metric3",
					// No version constraint - applies to all
				},
			},
			acsVersion: "4.8.0",
			wantCount:  2, // metric1 (4.8+) and metric3 (no constraint)
		},
		"should return all rules when no version specified": {
			rules: []rules.Rule{
				{
					RuleType:    rules.RuleTypeGauge,
					MetricName:  "metric1",
					ACSVersions: []string{"4.7+"},
				},
			},
			acsVersion: "",
			wantCount:  1, // All rules when no version filter
		},
		"should handle version range correctly": {
			rules: []rules.Rule{
				{
					RuleType:    rules.RuleTypeGauge,
					MetricName:  "metric1",
					ACSVersions: []string{"4.7-4.9"},
				},
			},
			acsVersion: "4.8.0",
			wantCount:  1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			filtered := FilterRulesByVersion(tt.rules, tt.acsVersion)

			if len(filtered) != tt.wantCount {
				t.Errorf("FilterRulesByVersion() got %d rules, want %d", len(filtered), tt.wantCount)
			}
		})
	}
}

func TestIsRuleApplicable(t *testing.T) {
	tests := map[string]struct {
		rule       rules.Rule
		acsVersion string
		want       bool
	}{
		"should return true for rule with matching version": {
			rule: rules.Rule{
				ACSVersions: []string{"4.7+", "4.8+"},
			},
			acsVersion: "4.8.0",
			want:       true,
		},
		"should return false for rule with non-matching version": {
			rule: rules.Rule{
				ACSVersions: []string{"4.9+"},
			},
			acsVersion: "4.8.0",
			want:       false,
		},
		"should return true for rule without version constraints": {
			rule:       rules.Rule{},
			acsVersion: "4.8.0",
			want:       true,
		},
		"should handle min_acs_version correctly": {
			rule: rules.Rule{
				MinACSVersion: "4.7",
			},
			acsVersion: "4.8.0",
			want:       true,
		},
		"should handle max_acs_version correctly": {
			rule: rules.Rule{
				MaxACSVersion: "4.9",
			},
			acsVersion: "4.8.0",
			want:       true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := IsRuleApplicable(tt.rule, tt.acsVersion)

			if got != tt.want {
				t.Errorf("IsRuleApplicable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluateHistogramInfOverflow(t *testing.T) {
	tests := map[string]struct {
		metrics    parser.MetricsData
		wantStatus map[string]rules.Status // metric name -> expected status
		wantCount  int                     // expected number of results
	}{
		"should return red status when >50% in +Inf": {
			metrics: parser.MetricsData{
				"test_histogram_bucket": &parser.Metric{
					Name: "test_histogram_bucket",
					Type: "histogram",
					Values: []parser.MetricValue{
						{Value: 10, Labels: map[string]string{"le": "0.1"}},
						{Value: 20, Labels: map[string]string{"le": "0.5"}},
						{Value: 30, Labels: map[string]string{"le": "1.0"}},
						{Value: 100, Labels: map[string]string{"le": "+Inf"}}, // 70 in +Inf (70%)
					},
				},
			},
			wantStatus: map[string]rules.Status{
				"test_histogram (+Inf overflow check)": rules.StatusRed,
			},
			wantCount: 1,
		},
		"should return yellow status when >25% but <=50% in +Inf": {
			metrics: parser.MetricsData{
				"test_histogram_bucket": &parser.Metric{
					Name: "test_histogram_bucket",
					Type: "histogram",
					Values: []parser.MetricValue{
						{Value: 10, Labels: map[string]string{"le": "0.1"}},
						{Value: 20, Labels: map[string]string{"le": "0.5"}},
						{Value: 30, Labels: map[string]string{"le": "1.0"}},
						{Value: 50, Labels: map[string]string{"le": "+Inf"}}, // 20 in +Inf (40%)
					},
				},
			},
			wantStatus: map[string]rules.Status{
				"test_histogram (+Inf overflow check)": rules.StatusYellow,
			},
			wantCount: 1,
		},
		"should return green status when <=25% in +Inf": {
			metrics: parser.MetricsData{
				"test_histogram_bucket": &parser.Metric{
					Name: "test_histogram_bucket",
					Type: "histogram",
					Values: []parser.MetricValue{
						{Value: 10, Labels: map[string]string{"le": "0.1"}},
						{Value: 20, Labels: map[string]string{"le": "0.5"}},
						{Value: 30, Labels: map[string]string{"le": "1.0"}},
						{Value: 35, Labels: map[string]string{"le": "+Inf"}}, // 5 in +Inf (14.3%)
					},
				},
			},
			wantStatus: map[string]rules.Status{
				"test_histogram (+Inf overflow check)": rules.StatusGreen,
			},
			wantCount: 1,
		},
		"should skip histogram without +Inf bucket": {
			metrics: parser.MetricsData{
				"test_histogram_bucket": &parser.Metric{
					Name: "test_histogram_bucket",
					Type: "histogram",
					Values: []parser.MetricValue{
						{Value: 10, Labels: map[string]string{"le": "0.1"}},
						{Value: 20, Labels: map[string]string{"le": "0.5"}},
					},
				},
			},
			wantStatus: map[string]rules.Status{},
			wantCount:  0,
		},
		"should handle multiple histograms": {
			metrics: parser.MetricsData{
				"hist1_bucket": &parser.Metric{
					Name: "hist1_bucket",
					Type: "histogram",
					Values: []parser.MetricValue{
						{Value: 10, Labels: map[string]string{"le": "1.0"}},
						{Value: 100, Labels: map[string]string{"le": "+Inf"}}, // 90 in +Inf (90%)
					},
				},
				"hist2_bucket": &parser.Metric{
					Name: "hist2_bucket",
					Type: "histogram",
					Values: []parser.MetricValue{
						{Value: 10, Labels: map[string]string{"le": "1.0"}},
						{Value: 15, Labels: map[string]string{"le": "+Inf"}}, // 5 in +Inf (33%)
					},
				},
			},
			wantStatus: map[string]rules.Status{
				"hist1 (+Inf overflow check)": rules.StatusRed,
				"hist2 (+Inf overflow check)": rules.StatusYellow,
			},
			wantCount: 2,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			results := EvaluateHistogramInfOverflow(tt.metrics, rules.LoadLevelMedium)

			if len(results) != tt.wantCount {
				t.Errorf("EvaluateHistogramInfOverflow() returned %d results, want %d", len(results), tt.wantCount)
			}

			for _, result := range results {
				wantStatus, exists := tt.wantStatus[result.RuleName]
				if !exists {
					t.Errorf("Unexpected result for metric %s", result.RuleName)
					continue
				}

				if result.Status != wantStatus {
					t.Errorf("EvaluateHistogramInfOverflow() for %s = %v, want %v", result.RuleName, result.Status, wantStatus)
				}

				// Verify message contains expected information
				if result.Status != rules.StatusGreen {
					if result.Message == "" {
						t.Error("Message should not be empty for non-green status")
					}
					if !strings.Contains(result.Message, "Highest non-infinity bucket") {
						t.Error("Message should contain 'Highest non-infinity bucket'")
					}
					if !strings.Contains(result.Message, "didn't expect") {
						t.Error("Message should contain explanation about designer expectations")
					}
				}
			}
		})
	}
}
