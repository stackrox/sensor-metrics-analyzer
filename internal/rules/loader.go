package rules

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// LoadRules loads all TOML rules from a directory
func LoadRules(rulesDir string) ([]Rule, error) {
	var rules []Rule

	files, err := filepath.Glob(filepath.Join(rulesDir, "*.toml"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob rules directory: %w", err)
	}

	for _, file := range files {
		rule, err := LoadRule(file)
		if err != nil {
			return nil, fmt.Errorf("failed to load rule %s: %w", file, err)
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// LoadRule loads a single TOML rule file
func LoadRule(filepath string) (Rule, error) {
	var rule Rule

	data, err := os.ReadFile(filepath)
	if err != nil {
		return rule, fmt.Errorf("failed to read file: %w", err)
	}

	if err := toml.Unmarshal(data, &rule); err != nil {
		return rule, fmt.Errorf("failed to parse TOML: %w", err)
	}

	// Validate the rule
	if err := ValidateRule(rule); err != nil {
		return rule, fmt.Errorf("validation failed: %w", err)
	}

	return rule, nil
}

// LoadLoadDetectionRules loads load detection rules from a directory
func LoadLoadDetectionRules(rulesDir string) ([]LoadDetectionRule, error) {
	var rules []LoadDetectionRule

	files, err := filepath.Glob(filepath.Join(rulesDir, "*.toml"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob load detection rules directory: %w", err)
	}

	for _, file := range files {
		var rule LoadDetectionRule
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", file, err)
		}

		if err := toml.Unmarshal(data, &rule); err != nil {
			return nil, fmt.Errorf("failed to parse TOML %s: %w", file, err)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}
