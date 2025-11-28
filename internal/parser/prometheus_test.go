package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFile(t *testing.T) {
	tests := map[string]struct {
		filename     string
		wantMetrics  int
		wantError    bool
		checkMetric  string
		wantValue    float64
		wantHasValue bool
	}{
		"should parse valid metrics file successfully": {
			filename:    "testdata/fixtures/sample_metrics.txt",
			wantMetrics: 20,
			wantError:   false,
		},
		"should extract gauge metric value correctly": {
			filename:     "testdata/fixtures/sample_metrics.txt",
			checkMetric:  "rox_sensor_num_containers_in_clusterentities_store",
			wantValue:    150,
			wantHasValue: true,
		},
		"should return error for non-existent file": {
			filename:  "testdata/fixtures/nonexistent.txt",
			wantError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Get absolute path relative to project root
			absPath := filepath.Join("..", "..", "..", tt.filename)
			absPath, err := filepath.Abs(absPath)
			if err != nil {
				t.Fatalf("Failed to get absolute path: %v", err)
			}

			// Check if file exists (except for error test case)
			if !tt.wantError {
				if _, err := os.Stat(absPath); os.IsNotExist(err) {
					t.Skipf("Test file %s does not exist, skipping", absPath)
					return
				}
			}

			metrics, err := ParseFile(absPath)

			if tt.wantError {
				if err == nil {
					t.Errorf("ParseFile() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseFile() error = %v", err)
			}

			if tt.wantMetrics > 0 && len(metrics) != tt.wantMetrics {
				t.Errorf("ParseFile() got %d metrics, want %d", len(metrics), tt.wantMetrics)
			}

			if tt.checkMetric != "" {
				metric, exists := metrics.GetMetric(tt.checkMetric)
				if !exists && tt.wantHasValue {
					t.Errorf("ParseFile() metric %s not found", tt.checkMetric)
					return
				}
				if exists {
					value, hasValue := metric.GetSingleValue()
					if hasValue != tt.wantHasValue {
						t.Errorf("ParseFile() metric %s hasValue = %v, want %v", tt.checkMetric, hasValue, tt.wantHasValue)
					}
					if hasValue && value != tt.wantValue {
						t.Errorf("ParseFile() metric %s value = %v, want %v", tt.checkMetric, value, tt.wantValue)
					}
				}
			}
		})
	}
}

func TestParseLabels(t *testing.T) {
	tests := map[string]struct {
		input    string
		want     map[string]string
		wantKeys []string
	}{
		"should parse single label correctly": {
			input:    `key1="value1"`,
			wantKeys: []string{"key1"},
			want:     map[string]string{"key1": "value1"},
		},
		"should parse multiple labels correctly": {
			input:    `key1="value1",key2="value2"`,
			wantKeys: []string{"key1", "key2"},
			want:     map[string]string{"key1": "value1", "key2": "value2"},
		},
		"should handle empty label string": {
			input:    "",
			wantKeys: []string{},
			want:     map[string]string{},
		},
		"should trim whitespace from labels": {
			input:    ` key1 = "value1" , key2 = "value2" `,
			wantKeys: []string{"key1", "key2"},
			want:     map[string]string{"key1": "value1", "key2": "value2"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := parseLabels(tt.input)

			if len(got) != len(tt.want) {
				t.Errorf("parseLabels() got %d keys, want %d", len(got), len(tt.want))
			}

			for key, wantValue := range tt.want {
				if gotValue, exists := got[key]; !exists {
					t.Errorf("parseLabels() missing key %s", key)
				} else if gotValue != wantValue {
					t.Errorf("parseLabels() key %s = %v, want %v", key, gotValue, wantValue)
				}
			}
		})
	}
}

func TestGetHistogramBuckets(t *testing.T) {
	tests := map[string]struct {
		metric    *Metric
		wantCount int
		wantLE    []float64
	}{
		"should extract histogram buckets correctly": {
			metric: &Metric{
				Name: "test_histogram_bucket",
				Values: []MetricValue{
					{Labels: map[string]string{"le": "0.005"}, Value: 100},
					{Labels: map[string]string{"le": "0.01"}, Value: 150},
					{Labels: map[string]string{"le": "+Inf"}, Value: 200},
				},
			},
			wantCount: 2, // +Inf is excluded
			wantLE:    []float64{0.005, 0.01},
		},
		"should return empty slice for non-histogram metric": {
			metric: &Metric{
				Name: "test_gauge",
				Values: []MetricValue{
					{Labels: map[string]string{}, Value: 100},
				},
			},
			wantCount: 0,
			wantLE:    []float64{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buckets := tt.metric.GetHistogramBuckets()

			if len(buckets) != tt.wantCount {
				t.Errorf("GetHistogramBuckets() got %d buckets, want %d", len(buckets), tt.wantCount)
			}

			for i, wantLE := range tt.wantLE {
				if i >= len(buckets) {
					t.Errorf("GetHistogramBuckets() bucket %d missing", i)
					continue
				}
				if buckets[i].Le != wantLE {
					t.Errorf("GetHistogramBuckets() bucket[%d].Le = %v, want %v", i, buckets[i].Le, wantLE)
				}
			}
		})
	}
}

func TestDetectACSVersion(t *testing.T) {
	tests := map[string]struct {
		metrics      MetricsData
		wantVersion  string
		wantDetected bool
	}{
		"should detect version from rox_sensor_version_info": {
			metrics: MetricsData{
				"rox_sensor_version_info": &Metric{
					Name: "rox_sensor_version_info",
					Values: []MetricValue{
						{Labels: map[string]string{"version": "4.8.0"}, Value: 1},
					},
				},
			},
			wantVersion:  "4.8.0",
			wantDetected: true,
		},
		"should detect version from rox_version label": {
			metrics: MetricsData{
				"rox_sensor_version_info": &Metric{
					Name: "rox_sensor_version_info",
					Values: []MetricValue{
						{Labels: map[string]string{"rox_version": "4.9.1"}, Value: 1},
					},
				},
			},
			wantVersion:  "4.9.1",
			wantDetected: true,
		},
		"should return false when version not found": {
			metrics: MetricsData{
				"some_other_metric": &Metric{
					Name: "some_other_metric",
					Values: []MetricValue{
						{Labels: map[string]string{}, Value: 100},
					},
				},
			},
			wantDetected: false,
		},
		"should return false for empty metrics": {
			metrics:      MetricsData{},
			wantDetected: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			version, detected := tt.metrics.DetectACSVersion()

			if detected != tt.wantDetected {
				t.Errorf("DetectACSVersion() detected = %v, want %v", detected, tt.wantDetected)
			}

			if detected && version != tt.wantVersion {
				t.Errorf("DetectACSVersion() version = %v, want %v", version, tt.wantVersion)
			}
		})
	}
}

func TestGetValuesByLabel(t *testing.T) {
	tests := map[string]struct {
		metric    *Metric
		labelKey  string
		wantCount int
		wantValue map[string]float64
	}{
		"should group values by label correctly": {
			metric: &Metric{
				Name: "test_metric",
				Values: []MetricValue{
					{Labels: map[string]string{"Operation": "Add"}, Value: 100},
					{Labels: map[string]string{"Operation": "Remove"}, Value: 50},
					{Labels: map[string]string{"Operation": "Add"}, Value: 200},
				},
			},
			labelKey:  "Operation",
			wantCount: 2,
			wantValue: map[string]float64{"Add": 200, "Remove": 50}, // Last value for each key
		},
		"should return empty map for missing label": {
			metric: &Metric{
				Name: "test_metric",
				Values: []MetricValue{
					{Labels: map[string]string{"other": "value"}, Value: 100},
				},
			},
			labelKey:  "Operation",
			wantCount: 0,
			wantValue: map[string]float64{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.metric.GetValuesByLabel(tt.labelKey)

			if len(got) != tt.wantCount {
				t.Errorf("GetValuesByLabel() got %d entries, want %d", len(got), tt.wantCount)
			}

			for key, wantVal := range tt.wantValue {
				if gotVal, exists := got[key]; !exists {
					t.Errorf("GetValuesByLabel() missing key %s", key)
				} else if gotVal != wantVal {
					t.Errorf("GetValuesByLabel() key %s = %v, want %v", key, gotVal, wantVal)
				}
			}
		})
	}
}
