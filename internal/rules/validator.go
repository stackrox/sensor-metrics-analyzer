package rules

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidateRule validates a rule's structure and values
func ValidateRule(rule Rule) error {
	// Check required fields
	if rule.RuleType == "" {
		return fmt.Errorf("rule_type is required")
	}

	// Validate rule type
	validTypes := []RuleType{
		RuleTypeGauge, RuleTypePercentage, RuleTypeQueue,
		RuleTypeHistogram, RuleTypeCacheHit, RuleTypeComposite,
	}
	isValid := false
	for _, vt := range validTypes {
		if rule.RuleType == vt {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid rule_type: %s", rule.RuleType)
	}

	// Validate ACS versions if specified
	if len(rule.ACSVersions) > 0 {
		for _, v := range rule.ACSVersions {
			if err := validateACSVersion(v); err != nil {
				return fmt.Errorf("invalid ACS version %s: %w", v, err)
			}
		}
	}

	if rule.MinACSVersion != "" {
		if err := validateACSVersion(rule.MinACSVersion); err != nil {
			return fmt.Errorf("invalid min_acs_version: %w", err)
		}
	}

	if rule.MaxACSVersion != "" {
		if err := validateACSVersion(rule.MaxACSVersion); err != nil {
			return fmt.Errorf("invalid max_acs_version: %w", err)
		}
	}

	// Validate load level thresholds if specified
	if rule.LoadLevelThresholds != nil {
		if err := validateLoadLevelThresholds(rule.LoadLevelThresholds); err != nil {
			return fmt.Errorf("invalid load_level_thresholds: %w", err)
		}
	}

	// Validate correlation config if specified
	if rule.Correlation != nil {
		if err := validateCorrelationConfig(rule.Correlation); err != nil {
			return fmt.Errorf("invalid correlation config: %w", err)
		}
	}

	// Type-specific validation
	switch rule.RuleType {
	case RuleTypeGauge:
		return validateGaugeRule(rule)
	case RuleTypePercentage:
		return validatePercentageRule(rule)
	case RuleTypeQueue:
		return validateQueueRule(rule)
	case RuleTypeHistogram:
		return validateHistogramRule(rule)
	case RuleTypeCacheHit:
		return validateCacheRule(rule)
	case RuleTypeComposite:
		return validateCompositeRule(rule)
	}

	return nil
}

func validateGaugeRule(rule Rule) error {
	if rule.MetricName == "" {
		return fmt.Errorf("metric_name is required for gauge rules")
	}
	// Allow low == high == 0 for zero-check rules (e.g., pods > 0 check)
	if rule.Thresholds.Low == 0 && rule.Thresholds.High == 0 {
		return nil
	}
	if rule.Thresholds.Low >= rule.Thresholds.High {
		return fmt.Errorf("low threshold must be less than high threshold")
	}
	return nil
}

func validatePercentageRule(rule Rule) error {
	if rule.PercentageConfig == nil {
		return fmt.Errorf("percentage_config is required")
	}
	if rule.PercentageConfig.Numerator == "" || rule.PercentageConfig.Denominator == "" {
		return fmt.Errorf("numerator and denominator are required")
	}
	if rule.Thresholds.Low >= rule.Thresholds.High {
		return fmt.Errorf("low threshold must be less than high threshold")
	}
	return nil
}

func validateQueueRule(rule Rule) error {
	if rule.MetricName == "" {
		return fmt.Errorf("metric_name is required for queue rules")
	}
	if rule.QueueConfig == nil {
		return fmt.Errorf("queue_config is required")
	}
	if rule.QueueConfig.OperationLabel == "" {
		return fmt.Errorf("operation_label is required")
	}
	return nil
}

func validateHistogramRule(rule Rule) error {
	if rule.MetricName == "" {
		return fmt.Errorf("metric_name is required for histogram rules")
	}
	if rule.Thresholds.P95Good >= rule.Thresholds.P95Warn {
		return fmt.Errorf("p95_good must be less than p95_warn")
	}
	return nil
}

func validateCacheRule(rule Rule) error {
	if rule.CacheConfig == nil {
		return fmt.Errorf("cache_config is required")
	}
	if rule.CacheConfig.HitsMetric == "" || rule.CacheConfig.MissesMetric == "" {
		return fmt.Errorf("hits_metric and misses_metric are required")
	}
	return nil
}

func validateCompositeRule(rule Rule) error {
	if rule.CompositeConfig == nil {
		return fmt.Errorf("composite_config is required")
	}
	if len(rule.CompositeConfig.Metrics) == 0 {
		return fmt.Errorf("at least one metric is required")
	}
	return nil
}

func validateLoadLevelThresholds(thresholds *LoadLevelThresholds) error {
	if thresholds.Low != nil {
		if thresholds.Low.Low >= thresholds.Low.High {
			return fmt.Errorf("low load level: low threshold must be less than high threshold")
		}
	}
	if thresholds.Medium != nil {
		if thresholds.Medium.Low >= thresholds.Medium.High {
			return fmt.Errorf("medium load level: low threshold must be less than high threshold")
		}
	}
	if thresholds.High != nil {
		if thresholds.High.Low >= thresholds.High.High {
			return fmt.Errorf("high load level: low threshold must be less than high threshold")
		}
	}
	return nil
}

func validateCorrelationConfig(config *CorrelationConfig) error {
	for i, cond := range config.SuppressIf {
		if err := validateCorrelationCondition(cond); err != nil {
			return fmt.Errorf("suppress_if[%d]: %w", i, err)
		}
	}
	for i, cond := range config.ElevateIf {
		if err := validateCorrelationCondition(cond); err != nil {
			return fmt.Errorf("elevate_if[%d]: %w", i, err)
		}
	}
	return nil
}

func validateCorrelationCondition(cond CorrelationCondition) error {
	if cond.MetricName == "" {
		return fmt.Errorf("metric_name is required")
	}
	validOps := []string{"gt", "lt", "eq", "gte", "lte"}
	isValid := false
	for _, op := range validOps {
		if cond.Operator == op {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid operator: %s (must be one of: gt, lt, eq, gte, lte)", cond.Operator)
	}
	return nil
}

func validateACSVersion(version string) error {
	// Support formats: "4.7", "4.7+", "4.7.1", ">=4.7", "4.7-4.9"
	versionPattern := regexp.MustCompile(`^(\d+\.\d+(?:\.\d+)?)(\+|$|-\d+\.\d+(?:\.\d+)?|>=|<=)?$`)
	if !versionPattern.MatchString(version) {
		return fmt.Errorf("invalid version format: %s (expected format: 4.7, 4.7+, >=4.7, or 4.7-4.9)", version)
	}
	return nil
}

// ValidateMessageTemplates checks if message templates reference valid placeholders
func ValidateMessageTemplates(rule Rule) error {
	// Extract placeholders from messages
	messages := []string{rule.Messages.Green, rule.Messages.Yellow, rule.Messages.Red}

	for _, msg := range messages {
		if msg == "" {
			continue
		}

		// Find all {placeholder} patterns
		placeholders := findPlaceholders(msg)

		// Validate placeholders exist for rule type
		for _, ph := range placeholders {
			if !isValidPlaceholder(ph, rule.RuleType) {
				return fmt.Errorf("invalid placeholder {%s} in message template", ph)
			}
		}
	}

	return nil
}

func findPlaceholders(msg string) []string {
	var placeholders []string
	start := -1

	for i, ch := range msg {
		if ch == '{' {
			start = i
		} else if ch == '}' && start != -1 {
			placeholders = append(placeholders, msg[start+1:i])
			start = -1
		}
	}

	return placeholders
}

func isValidPlaceholder(placeholder string, ruleType RuleType) bool {
	commonPlaceholders := []string{"value", "status"}

	for _, cp := range commonPlaceholders {
		if placeholder == cp {
			return true
		}
	}

	// Type-specific placeholders
	switch ruleType {
	case RuleTypeQueue:
		return strings.Contains(placeholder, "add") || strings.Contains(placeholder, "remove") || strings.Contains(placeholder, "diff")
	case RuleTypeHistogram:
		return strings.Contains(placeholder, "p95") || strings.Contains(placeholder, "p99")
	case RuleTypeComposite:
		// Any placeholder is valid for composite (metric names)
		return true
	}

	return false
}
