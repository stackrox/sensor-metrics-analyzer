package reporter

import (
	"fmt"
	"sort"

	"github.com/stackrox/sensor-metrics-analyzer/internal/rules"
)

// GenerateMarkdown creates a markdown report from analysis results.
// The markdown template is the single source of truth; if it is missing
// or fails to render, return an error.
func GenerateMarkdown(report rules.AnalysisReport, templatePath string) (string, error) {
	if templatePath == "" {
		return "", fmt.Errorf("markdown template path is empty")
	}
	result, err := GenerateMarkdownFromTemplate(report, templatePath)
	if err != nil {
		return "", err
	}
	if result == "" {
		return "", fmt.Errorf("markdown template returned empty content")
	}
	return result, nil
}

func filterByStatus(results []rules.EvaluationResult, status rules.Status) []rules.EvaluationResult {
	var filtered []rules.EvaluationResult
	for _, r := range results {
		if r.Status == status {
			filtered = append(filtered, r)
		}
	}

	// Sort by rule name
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].RuleName < filtered[j].RuleName
	})

	return filtered
}
