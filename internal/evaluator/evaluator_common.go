package evaluator

import (
	"fmt"
	"time"

	"github.com/stackrox/sensor-metrics-analyzer/internal/parser"
	"github.com/stackrox/sensor-metrics-analyzer/internal/rules"
)

// selectThresholds selects appropriate thresholds based on load level
func selectThresholds(rule rules.Rule, loadLevel rules.LoadLevel) rules.Thresholds {
	// If load level thresholds are not configured, use default thresholds
	if rule.LoadLevelThresholds == nil {
		return rule.Thresholds
	}

	var selected *rules.Thresholds
	switch loadLevel {
	case rules.LoadLevelLow:
		selected = rule.LoadLevelThresholds.Low
	case rules.LoadLevelMedium:
		selected = rule.LoadLevelThresholds.Medium
	case rules.LoadLevelHigh:
		selected = rule.LoadLevelThresholds.High
	}

	// If no threshold for this load level, fall back to default
	if selected == nil {
		return rule.Thresholds
	}

	// Merge with defaults (use load level specific if set, otherwise default)
	result := rule.Thresholds
	if selected.Low > 0 || selected.High > 0 {
		if selected.Low > 0 {
			result.Low = selected.Low
		}
		if selected.High > 0 {
			result.High = selected.High
		}
		result.HigherIsWorse = selected.HigherIsWorse
	}
	if selected.P95Good > 0 {
		result.P95Good = selected.P95Good
	}
	if selected.P95Warn > 0 {
		result.P95Warn = selected.P95Warn
	}
	if selected.MinRatio > 0 {
		result.MinRatio = selected.MinRatio
	}

	return result
}

// getRemediation returns the remediation message for a given status
func getRemediation(rule rules.Rule, status rules.Status) string {
	if rule.Remediation == nil {
		return ""
	}

	switch status {
	case rules.StatusRed:
		return rule.Remediation.Red
	case rules.StatusYellow:
		return rule.Remediation.Yellow
	case rules.StatusGreen:
		return rule.Remediation.Green
	default:
		return ""
	}
}

func applyReviewMetadata(rule rules.Rule) string {
	review := rule.Reviewed
	by := rule.LastReviewBy
	on := rule.LastReviewOn

	switch {
	case review != "" && by != "" && on != "":
		return fmt.Sprintf("%s (last review: %s on %s)", review, by, on)
	case review != "" && by != "":
		return fmt.Sprintf("%s (last review: %s)", review, by)
	case review != "" && on != "":
		return fmt.Sprintf("%s (last review: %s)", review, on)
	case review != "":
		return review
	case by != "" && on != "":
		return fmt.Sprintf("Last review: %s on %s", by, on)
	case by != "":
		return fmt.Sprintf("Last review: %s", by)
	case on != "":
		return fmt.Sprintf("Last review: %s", on)
	}

	return "review status unavailable"
}

// EvaluateAllRules evaluates all rules against metrics
func EvaluateAllRules(rulesList []rules.Rule, metrics parser.MetricsData, loadLevel rules.LoadLevel, acsVersion string) rules.AnalysisReport {
	report := rules.AnalysisReport{
		ACSVersion: acsVersion,
		LoadLevel:  loadLevel,
		Timestamp:  time.Now(),
	}

	// Filter rules by ACS version
	filteredRules := FilterRulesByVersion(rulesList, acsVersion)

	for _, rule := range filteredRules {
		var result rules.EvaluationResult

		// Evaluate based on rule type
		switch rule.RuleType {
		case rules.RuleTypeGauge:
			result = EvaluateGauge(rule, metrics, loadLevel)
		case rules.RuleTypePercentage:
			result = EvaluatePercentage(rule, metrics, loadLevel)
		case rules.RuleTypeQueue:
			result = EvaluateQueue(rule, metrics, loadLevel)
		case rules.RuleTypeHistogram:
			result = EvaluateHistogram(rule, metrics, loadLevel)
		case rules.RuleTypeCacheHit:
			result = EvaluateCacheHit(rule, metrics, loadLevel)
		case rules.RuleTypeComposite:
			result = EvaluateComposite(rule, metrics, loadLevel)
		default:
			continue // Skip unknown rule types
		}

		// Apply correlation if configured
		if rule.Correlation != nil {
			result = EvaluateCorrelation(rule, metrics, result)
		}

		result.ReviewStatus = applyReviewMetadata(rule)

		// Add potential actions (user-facing)
		result.Remediation = getRemediation(rule, result.Status)
		result.PotentialActionUser = result.Remediation
		result.Timestamp = time.Now()

		report.Results = append(report.Results, result)

		// Update summary
		switch result.Status {
		case rules.StatusRed:
			report.Summary.RedCount++
		case rules.StatusYellow:
			report.Summary.YellowCount++
		case rules.StatusGreen:
			report.Summary.GreenCount++
		}
	}

	// Apply general histogram +Inf overflow rule to all histogram metrics
	infOverflowResults := EvaluateHistogramInfOverflow(metrics, loadLevel)
	for _, result := range infOverflowResults {
		report.Results = append(report.Results, result)

		// Update summary
		switch result.Status {
		case rules.StatusRed:
			report.Summary.RedCount++
		case rules.StatusYellow:
			report.Summary.YellowCount++
		case rules.StatusGreen:
			report.Summary.GreenCount++
		}
	}

	report.Summary.TotalAnalyzed = len(report.Results)

	return report
}
