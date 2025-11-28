package evaluator

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/stackrox/sensor-metrics-analyzer/internal/rules"
)

// FilterRulesByVersion filters rules applicable to given ACS version
func FilterRulesByVersion(rulesList []rules.Rule, acsVersion string) []rules.Rule {
	if acsVersion == "" {
		// No version specified, return all rules
		return rulesList
	}

	var filtered []rules.Rule
	for _, rule := range rulesList {
		if IsRuleApplicable(rule, acsVersion) {
			filtered = append(filtered, rule)
		}
	}
	return filtered
}

// IsRuleApplicable checks if rule applies to ACS version
func IsRuleApplicable(rule rules.Rule, acsVersion string) bool {
	// If no version constraints, rule applies to all versions
	if len(rule.ACSVersions) == 0 && rule.MinACSVersion == "" && rule.MaxACSVersion == "" {
		return true
	}

	// Check ACSVersions array
	if len(rule.ACSVersions) > 0 {
		for _, versionSpec := range rule.ACSVersions {
			if matchesVersion(acsVersion, versionSpec) {
				return true
			}
		}
		return false
	}

	// Check min/max version constraints
	if rule.MinACSVersion != "" {
		if !versionGreaterOrEqual(acsVersion, rule.MinACSVersion) {
			return false
		}
	}

	if rule.MaxACSVersion != "" {
		if !versionLessOrEqual(acsVersion, rule.MaxACSVersion) {
			return false
		}
	}

	return true
}

// matchesVersion checks if a version matches a version specifier
// Supports: "4.7", "4.7+", ">=4.7", "4.7-4.9"
func matchesVersion(version, spec string) bool {
	// Parse version
	ver, err := parseVersion(version)
	if err != nil {
		return false
	}

	// Handle range: "4.7-4.9"
	if strings.Contains(spec, "-") {
		parts := strings.Split(spec, "-")
		if len(parts) == 2 {
			minVer, err1 := parseVersion(parts[0])
			maxVer, err2 := parseVersion(parts[1])
			if err1 != nil || err2 != nil {
				return false
			}
			return compareVersions(ver, minVer) >= 0 && compareVersions(ver, maxVer) <= 0
		}
	}

	// Handle ">=" prefix
	if strings.HasPrefix(spec, ">=") {
		minVer, err := parseVersion(strings.TrimPrefix(spec, ">="))
		if err != nil {
			return false
		}
		return compareVersions(ver, minVer) >= 0
	}

	// Handle "+" suffix: "4.7+" means >= 4.7
	if strings.HasSuffix(spec, "+") {
		minVer, err := parseVersion(strings.TrimSuffix(spec, "+"))
		if err != nil {
			return false
		}
		return compareVersions(ver, minVer) >= 0
	}

	// Exact match
	exactVer, err := parseVersion(spec)
	if err != nil {
		return false
	}
	return compareVersions(ver, exactVer) == 0
}

// parseVersion parses a version string into major.minor.patch
func parseVersion(version string) ([3]int, error) {
	var result [3]int

	// Remove any non-numeric prefixes/suffixes
	version = strings.TrimSpace(version)

	// Extract version numbers
	re := regexp.MustCompile(`(\d+)\.(\d+)(?:\.(\d+))?`)
	matches := re.FindStringSubmatch(version)
	if len(matches) < 3 {
		return result, strconv.ErrSyntax
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch := 0
	if len(matches) > 3 && matches[3] != "" {
		patch, _ = strconv.Atoi(matches[3])
	}

	result[0] = major
	result[1] = minor
	result[2] = patch
	return result, nil
}

// compareVersions compares two versions
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareVersions(v1, v2 [3]int) int {
	if v1[0] != v2[0] {
		if v1[0] < v2[0] {
			return -1
		}
		return 1
	}
	if v1[1] != v2[1] {
		if v1[1] < v2[1] {
			return -1
		}
		return 1
	}
	if v1[2] != v2[2] {
		if v1[2] < v2[2] {
			return -1
		}
		return 1
	}
	return 0
}

// versionGreaterOrEqual checks if v1 >= v2
func versionGreaterOrEqual(v1, v2 string) bool {
	ver1, err1 := parseVersion(v1)
	ver2, err2 := parseVersion(v2)
	if err1 != nil || err2 != nil {
		return false
	}
	return compareVersions(ver1, ver2) >= 0
}

// versionLessOrEqual checks if v1 <= v2
func versionLessOrEqual(v1, v2 string) bool {
	ver1, err1 := parseVersion(v1)
	ver2, err2 := parseVersion(v2)
	if err1 != nil || err2 != nil {
		return false
	}
	return compareVersions(ver1, ver2) <= 0
}
