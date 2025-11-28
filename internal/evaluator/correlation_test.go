package evaluator

import (
	"testing"

	"github.com/stackrox/sensor-metrics-analyzer/internal/parser"
	"github.com/stackrox/sensor-metrics-analyzer/internal/rules"
)

func TestEvaluateCorrelation(t *testing.T) {
	tests := map[string]struct {
		rule       rules.Rule
		metrics    parser.MetricsData
		initial    rules.EvaluationResult
		wantStatus rules.Status
	}{
		"should suppress status when correlation condition met": {
			rule: rules.Rule{
				Correlation: &rules.CorrelationConfig{
					SuppressIf: []rules.CorrelationCondition{
						{
							MetricName: "go_goroutines",
							Operator:   "lt",
							Value:      1000,
							Status:     rules.StatusYellow,
						},
					},
				},
			},
			metrics: parser.MetricsData{
				"go_goroutines": &parser.Metric{
					Name: "go_goroutines",
					Values: []parser.MetricValue{
						{Value: 500, Labels: make(map[string]string)},
					},
				},
			},
			initial: rules.EvaluationResult{
				Status: rules.StatusRed,
			},
			wantStatus: rules.StatusYellow, // Suppressed from RED to YELLOW
		},
		"should elevate status when correlation condition met": {
			rule: rules.Rule{
				Correlation: &rules.CorrelationConfig{
					ElevateIf: []rules.CorrelationCondition{
						{
							MetricName: "go_goroutines",
							Operator:   "gt",
							Value:      10000,
							Status:     rules.StatusRed,
						},
					},
				},
			},
			metrics: parser.MetricsData{
				"go_goroutines": &parser.Metric{
					Name: "go_goroutines",
					Values: []parser.MetricValue{
						{Value: 15000, Labels: make(map[string]string)},
					},
				},
			},
			initial: rules.EvaluationResult{
				Status: rules.StatusYellow,
			},
			wantStatus: rules.StatusRed, // Elevated from YELLOW to RED
		},
		"should not modify status when correlation condition not met": {
			rule: rules.Rule{
				Correlation: &rules.CorrelationConfig{
					SuppressIf: []rules.CorrelationCondition{
						{
							MetricName: "go_goroutines",
							Operator:   "lt",
							Value:      1000,
						},
					},
				},
			},
			metrics: parser.MetricsData{
				"go_goroutines": &parser.Metric{
					Name: "go_goroutines",
					Values: []parser.MetricValue{
						{Value: 2000, Labels: make(map[string]string)},
					},
				},
			},
			initial: rules.EvaluationResult{
				Status: rules.StatusRed,
			},
			wantStatus: rules.StatusRed, // Unchanged
		},
		"should handle missing correlation metric gracefully": {
			rule: rules.Rule{
				Correlation: &rules.CorrelationConfig{
					SuppressIf: []rules.CorrelationCondition{
						{
							MetricName: "missing_metric",
							Operator:   "lt",
							Value:      1000,
						},
					},
				},
			},
			metrics: parser.MetricsData{},
			initial: rules.EvaluationResult{
				Status: rules.StatusRed,
			},
			wantStatus: rules.StatusRed, // Unchanged when metric missing
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := EvaluateCorrelation(tt.rule, tt.metrics, tt.initial)

			if result.Status != tt.wantStatus {
				t.Errorf("EvaluateCorrelation() status = %v, want %v", result.Status, tt.wantStatus)
			}
		})
	}
}

func TestEvaluateHistogram(t *testing.T) {
	tests := map[string]struct {
		rule       rules.Rule
		metrics    parser.MetricsData
		loadLevel  rules.LoadLevel
		wantStatus rules.Status
	}{
		"should calculate P95 correctly and return green": {
			rule: rules.Rule{
				RuleType:   rules.RuleTypeHistogram,
				MetricName: "test_histogram",
				Thresholds: rules.Thresholds{
					P95Good: 0.1,
					P95Warn: 1.0,
				},
				Messages: rules.Messages{
					Green: "p95={p95:.3f}s (good)",
				},
			},
			metrics: parser.MetricsData{
				"test_histogram_bucket": &parser.Metric{
					Name: "test_histogram_bucket",
					Values: []parser.MetricValue{
						{Labels: map[string]string{"le": "0.005"}, Value: 100},
						{Labels: map[string]string{"le": "0.01"}, Value: 150},
						{Labels: map[string]string{"le": "0.025"}, Value: 180},
						{Labels: map[string]string{"le": "0.05"}, Value: 190},
						{Labels: map[string]string{"le": "+Inf"}, Value: 200},
					},
				},
			},
			loadLevel:  rules.LoadLevelMedium,
			wantStatus: rules.StatusGreen,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := EvaluateHistogram(tt.rule, tt.metrics, tt.loadLevel)

			if result.Status != tt.wantStatus {
				t.Errorf("EvaluateHistogram() status = %v, want %v", result.Status, tt.wantStatus)
			}
		})
	}
}

func TestEvaluateCacheHit(t *testing.T) {
	tests := map[string]struct {
		rule       rules.Rule
		metrics    parser.MetricsData
		loadLevel  rules.LoadLevel
		wantStatus rules.Status
	}{
		"should calculate cache hit rate correctly": {
			rule: rules.Rule{
				RuleType: rules.RuleTypeCacheHit,
				CacheConfig: &rules.CacheConfig{
					HitsMetric:   "cache_hits",
					MissesMetric: "cache_misses",
				},
				Thresholds: rules.Thresholds{
					Low:           40.0,
					High:          60.0,
					HigherIsWorse: false,
				},
				Messages: rules.Messages{
					Green: "{value:.1f}% hit rate (good)",
				},
			},
			metrics: parser.MetricsData{
				"cache_hits": &parser.Metric{
					Name: "cache_hits",
					Values: []parser.MetricValue{
						{Value: 80, Labels: make(map[string]string)},
					},
				},
				"cache_misses": &parser.Metric{
					Name: "cache_misses",
					Values: []parser.MetricValue{
						{Value: 20, Labels: make(map[string]string)},
					},
				},
			},
			loadLevel:  rules.LoadLevelMedium,
			wantStatus: rules.StatusGreen, // 80/(80+20) = 80% > 60%
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := EvaluateCacheHit(tt.rule, tt.metrics, tt.loadLevel)

			if result.Status != tt.wantStatus {
				t.Errorf("EvaluateCacheHit() status = %v, want %v", result.Status, tt.wantStatus)
			}
		})
	}
}

func TestEvaluateComposite(t *testing.T) {
	tests := map[string]struct {
		rule       rules.Rule
		metrics    parser.MetricsData
		loadLevel  rules.LoadLevel
		wantStatus rules.Status
	}{
		"should evaluate composite rule with not_zero check": {
			rule: rules.Rule{
				RuleType: rules.RuleTypeComposite,
				CompositeConfig: &rules.CompositeConfig{
					Metrics: []rules.CompositeMetric{
						{Name: "containers", Source: "containers"},
						{Name: "endpoints", Source: "endpoints"},
					},
					Checks: []rules.CompositeCheck{
						{
							CheckType: "not_zero",
							Metrics:   []string{"containers", "endpoints"},
							Status:    "red",
							Message:   "Zero detected",
						},
					},
				},
				Messages: rules.Messages{
					Green: "All healthy",
				},
			},
			metrics: parser.MetricsData{
				"containers": &parser.Metric{
					Name: "containers",
					Values: []parser.MetricValue{
						{Value: 100, Labels: make(map[string]string)},
					},
				},
				"endpoints": &parser.Metric{
					Name: "endpoints",
					Values: []parser.MetricValue{
						{Value: 50, Labels: make(map[string]string)},
					},
				},
			},
			loadLevel:  rules.LoadLevelMedium,
			wantStatus: rules.StatusGreen, // No zeros detected
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := EvaluateComposite(tt.rule, tt.metrics, tt.loadLevel)

			if result.Status != tt.wantStatus {
				t.Errorf("EvaluateComposite() status = %v, want %v", result.Status, tt.wantStatus)
			}
		})
	}
}
