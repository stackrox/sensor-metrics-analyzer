package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRule(t *testing.T) {
	tests := map[string]struct {
		filename  string
		wantError bool
		checkFunc func(*testing.T, Rule)
	}{
		"should load valid gauge rule successfully": {
			filename:  "../../testdata/fixtures/test_gauge.toml",
			wantError: false,
			checkFunc: func(t *testing.T, r Rule) {
				if r.RuleType != RuleTypeGauge {
					t.Errorf("LoadRule() ruleType = %v, want %v", r.RuleType, RuleTypeGauge)
				}
				if r.MetricName != "rox_sensor_network_flow_buffer_size" {
					t.Errorf("LoadRule() metricName = %v, want rox_sensor_network_flow_buffer_size", r.MetricName)
				}
			},
		},
		"should load valid percentage rule successfully": {
			filename:  "../../testdata/fixtures/test_percentage.toml",
			wantError: false,
			checkFunc: func(t *testing.T, r Rule) {
				if r.RuleType != RuleTypePercentage {
					t.Errorf("LoadRule() ruleType = %v, want %v", r.RuleType, RuleTypePercentage)
				}
				if r.PercentageConfig == nil {
					t.Error("LoadRule() PercentageConfig is nil")
				}
			},
		},
		"should return error for invalid rule file": {
			filename:  "../../testdata/fixtures/nonexistent.toml",
			wantError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			absPath, err := filepath.Abs(tt.filename)
			if err != nil {
				t.Fatalf("Failed to get absolute path: %v", err)
			}

			if !tt.wantError {
				if _, err := os.Stat(absPath); os.IsNotExist(err) {
					t.Skipf("Test file %s does not exist, skipping", absPath)
					return
				}
			}

			rule, err := LoadRule(absPath)

			if tt.wantError {
				if err == nil {
					t.Error("LoadRule() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("LoadRule() error = %v", err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, rule)
			}
		})
	}
}

func TestValidateRule(t *testing.T) {
	tests := map[string]struct {
		rule      Rule
		wantError bool
	}{
		"should validate gauge rule correctly": {
			rule: Rule{
				RuleType:   RuleTypeGauge,
				MetricName: "test_metric",
				Thresholds: Thresholds{
					Low:  10,
					High: 100,
				},
			},
			wantError: false,
		},
		"should return error for missing rule type": {
			rule: Rule{
				MetricName: "test_metric",
			},
			wantError: true,
		},
		"should return error for invalid rule type": {
			rule: Rule{
				RuleType: RuleType("invalid"),
			},
			wantError: true,
		},
		"should return error when low threshold >= high threshold": {
			rule: Rule{
				RuleType:   RuleTypeGauge,
				MetricName: "test_metric",
				Thresholds: Thresholds{
					Low:  100,
					High: 50,
				},
			},
			wantError: true,
		},
		"should allow zero-check rule (low=0, high=0)": {
			rule: Rule{
				RuleType:   RuleTypeGauge,
				MetricName: "test_metric",
				Thresholds: Thresholds{
					Low:  0,
					High: 0,
				},
			},
			wantError: false,
		},
		"should validate ACS version format correctly": {
			rule: Rule{
				RuleType:   RuleTypeGauge,
				MetricName: "test_metric",
				Thresholds: Thresholds{
					Low:  10,
					High: 100,
				},
				ACSVersions: []string{"4.7+", "4.8+"},
			},
			wantError: false,
		},
		"should return error for invalid ACS version format": {
			rule: Rule{
				RuleType:   RuleTypeGauge,
				MetricName: "test_metric",
				Thresholds: Thresholds{
					Low:  10,
					High: 100,
				},
				ACSVersions: []string{"invalid-version"},
			},
			wantError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := ValidateRule(tt.rule)

			if (err != nil) != tt.wantError {
				t.Errorf("ValidateRule() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestLoadRules(t *testing.T) {
	tests := map[string]struct {
		dir       string
		wantCount int
		wantError bool
	}{
		"should load all rules from directory": {
			dir:       "../../testdata/fixtures",
			wantCount: 7, // Test rule files (excluding load detection which fails validation)
			wantError: false,
		},
		"should handle non-existent directory gracefully": {
			dir:       "../../testdata/nonexistent",
			wantCount: 0, // Glob returns empty slice, not error
			wantError: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Tests run from project root, so use path directly
			absPath, err := filepath.Abs(tt.dir)
			if err != nil {
				t.Fatalf("Failed to get absolute path: %v", err)
			}

			rules, err := LoadRules(absPath)

			if tt.wantError {
				if err == nil {
					t.Error("LoadRules() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("LoadRules() error = %v (path: %s)", err, absPath)
			}

			if len(rules) < tt.wantCount {
				t.Errorf("LoadRules() got %d rules, want at least %d (path: %s)", len(rules), tt.wantCount, absPath)
			}
		})
	}
}
