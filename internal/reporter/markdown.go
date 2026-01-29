package reporter

import (
	"fmt"
	"sort"

	"github.com/stackrox/sensor-metrics-analyzer/internal/rules"
)

// GenerateMarkdown creates a markdown report from analysis results
func GenerateMarkdown(report rules.AnalysisReport, templatePath string) string {
	// Try to use template if available
	if templatePath != "" {
		result, err := GenerateMarkdownFromTemplate(report, templatePath)
		if err == nil && result != "" {
			return result
		}
		// Fall back to default if template fails (don't log error, just fall back)
	}

	// Default markdown generation (fallback)
	return generateMarkdownDefault(report)
}

// generateMarkdownDefault generates markdown without template (fallback)
func generateMarkdownDefault(report rules.AnalysisReport) string {
	var result string

	// Header
	result += "# Automated Metrics Analysis Report\n\n"
	result += "**Cluster:** " + report.ClusterName + "\n"
	result += "**ACS Version:** " + report.ACSVersion + "\n"
	result += "**Load Level:** " + string(report.LoadLevel) + "\n"
	result += "**Generated:** " + report.Timestamp.Format("2006-01-02 15:04:05") + "\n\n"

	// Summary
	result += "## Summary\n\n"
	result += "- 🔴 **RED:** " + formatInt(report.Summary.RedCount) + " metrics\n"
	result += "- 🟡 **YELLOW:** " + formatInt(report.Summary.YellowCount) + " metrics\n"
	result += "- 🟢 **GREEN:** " + formatInt(report.Summary.GreenCount) + " metrics\n\n"

	// Critical Issues (RED metrics)
	redResults := filterByStatus(report.Results, rules.StatusRed)
	if len(redResults) > 0 {
		result += "## 🔴 Critical Issues\n\n"
		for _, r := range redResults {
			result += "### " + r.RuleName + "\n\n"
			result += "**Status:** RED\n"
			result += "**Message:** " + r.Message + "\n"
			if len(r.Details) > 0 {
				result += "**Details:**\n"
				for _, detail := range r.Details {
					result += "- " + detail + "\n"
				}
			}
			if r.PotentialActionUser != "" {
				result += "**Potential action:** " + r.PotentialActionUser + "\n"
			}
			if r.PotentialActionDeveloper != "" {
				result += "**Potential action (developer):** " + r.PotentialActionDeveloper + "\n"
			}
			result += "\n"
		}
	}

	// Warnings (YELLOW metrics)
	yellowResults := filterByStatus(report.Results, rules.StatusYellow)
	if len(yellowResults) > 0 {
		result += "## 🟡 Warnings\n\n"
		for _, r := range yellowResults {
			result += "### " + r.RuleName + "\n\n"
			result += "**Status:** YELLOW\n"
			result += "**Message:** " + r.Message + "\n"
			if len(r.Details) > 0 {
				result += "**Details:**\n"
				for _, detail := range r.Details {
					result += "- " + detail + "\n"
				}
			}
			if r.PotentialActionUser != "" {
				result += "**Potential action:** " + r.PotentialActionUser + "\n"
			}
			if r.PotentialActionDeveloper != "" {
				result += "**Potential action (developer):** " + r.PotentialActionDeveloper + "\n"
			}
			result += "\n"
		}
	}

	// Healthy Metrics (GREEN)
	greenResults := filterByStatus(report.Results, rules.StatusGreen)
	if len(greenResults) > 0 {
		result += "## 🟢 Healthy Metrics\n\n"
		for _, r := range greenResults {
			result += "- **" + r.RuleName + ":** " + r.Message + "\n"
		}
	}

	return result
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

func formatInt(i int) string {
	// Use fmt.Sprintf for simplicity
	return fmt.Sprintf("%d", i)
}
