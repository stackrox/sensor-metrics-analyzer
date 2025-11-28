package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2EAnalyzeCommand(t *testing.T) {
	tests := map[string]struct {
		args       []string
		wantOutput []string
		wantError  bool
	}{
		"should analyze metrics file successfully": {
			args: []string{
				"analyze",
				"--rules", "../../testdata/fixtures",
				"../../testdata/fixtures/sample_metrics.txt",
			},
			wantOutput: []string{"Summary", "RED", "YELLOW", "GREEN"},
			wantError:  false,
		},
		"should generate markdown output when format specified": {
			args: []string{
				"analyze",
				"--rules", "../../testdata/fixtures",
				"--format", "markdown",
				"--output", "/tmp/test_e2e_report.md",
				"../../testdata/fixtures/sample_metrics.txt",
			},
			wantOutput: []string{"Report written"},
			wantError:  false,
		},
		"should return error when metrics file not found": {
			args: []string{
				"analyze",
				"--rules", "../../testdata/fixtures",
				"nonexistent.txt",
			},
			wantError: true,
		},
		"should handle non-existent rules directory gracefully": {
			args: []string{
				"analyze",
				"--rules", "nonexistent",
				"../../testdata/fixtures/sample_metrics.txt",
			},
			wantOutput: []string{"Loaded 0 rules"}, // Should not error, just load 0 rules
			wantError:  false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Get binary path
			binPath := filepath.Join("..", "..", "bin", "metrics-analyzer")
			absPath, err := filepath.Abs(binPath)
			if err != nil {
				t.Fatalf("Failed to get absolute path: %v", err)
			}

			// Check if binary exists
			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				t.Skipf("Binary %s does not exist, skipping e2e test. Run 'make build' first.", absPath)
				return
			}

			cmd := exec.Command(absPath, tt.args...)
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none. Output: %s", outputStr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Command failed: %v\nOutput: %s", err, outputStr)
			}

			// Check for expected output strings
			for _, wantStr := range tt.wantOutput {
				if !strings.Contains(outputStr, wantStr) {
					t.Errorf("Output missing expected string '%s'. Full output:\n%s", wantStr, outputStr)
				}
			}
		})
	}
}

func TestE2EValidateCommand(t *testing.T) {
	tests := map[string]struct {
		args      []string
		wantError bool
	}{
		"should validate rules successfully": {
			args: []string{
				"validate",
				"../../testdata/fixtures",
			},
			wantError: false,
		},
		"should handle non-existent directory gracefully": {
			args: []string{
				"validate",
				"../../testdata/nonexistent",
			},
			wantError: false, // Returns 0 rules, not an error
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			binPath := filepath.Join("..", "..", "bin", "metrics-analyzer")
			absPath, err := filepath.Abs(binPath)
			if err != nil {
				t.Fatalf("Failed to get absolute path: %v", err)
			}

			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				t.Skipf("Binary %s does not exist, skipping e2e test", absPath)
				return
			}

			cmd := exec.Command(absPath, tt.args...)
			output, err := cmd.CombinedOutput()

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none. Output: %s", string(output))
				}
				return
			}

			if err != nil {
				t.Fatalf("Command failed: %v\nOutput: %s", err, string(output))
			}
		})
	}
}

func TestE2EListRulesCommand(t *testing.T) {
	tests := map[string]struct {
		args      []string
		wantError bool
	}{
		"should list rules successfully": {
			args: []string{
				"list-rules",
				"../../testdata/fixtures",
			},
			wantError: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			binPath := filepath.Join("..", "..", "bin", "metrics-analyzer")
			absPath, err := filepath.Abs(binPath)
			if err != nil {
				t.Fatalf("Failed to get absolute path: %v", err)
			}

			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				t.Skipf("Binary %s does not exist, skipping e2e test", absPath)
				return
			}

			cmd := exec.Command(absPath, tt.args...)
			output, err := cmd.CombinedOutput()

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none. Output: %s", string(output))
				}
				return
			}

			if err != nil {
				t.Fatalf("Command failed: %v\nOutput: %s", err, string(output))
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, "Found") && !strings.Contains(outputStr, "rules") {
				t.Errorf("Output missing expected content. Output: %s", outputStr)
			}
		})
	}
}

func TestE2EFullWorkflow(t *testing.T) {
	// Test complete workflow: load rules, parse metrics, evaluate, generate report
	binPath := filepath.Join("..", "..", "bin", "metrics-analyzer")
	absPath, err := filepath.Abs(binPath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("Binary %s does not exist, skipping e2e test", absPath)
		return
	}

	metricsPath := filepath.Join("..", "..", "testdata", "fixtures", "sample_metrics.txt")
	rulesPath := filepath.Join("..", "..", "testdata", "fixtures")

	cmd := exec.Command(absPath, "analyze",
		"--rules", rulesPath,
		"--format", "markdown",
		"--output", "/tmp/e2e_test_report.md",
		metricsPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Full workflow test failed: %v\nOutput: %s", err, string(output))
	}

	// Verify report was created
	if _, err := os.Stat("/tmp/e2e_test_report.md"); os.IsNotExist(err) {
		t.Error("Report file was not created")
	}

	// Verify report contains expected content
	reportContent, err := os.ReadFile("/tmp/e2e_test_report.md")
	if err != nil {
		t.Fatalf("Failed to read report file: %v", err)
	}

	reportStr := string(reportContent)
	expectedStrings := []string{"Automated Metrics Analysis Report", "Summary", "RED", "YELLOW", "GREEN"}
	for _, expected := range expectedStrings {
		if !strings.Contains(reportStr, expected) {
			t.Errorf("Report missing expected content: %s", expected)
		}
	}
}
