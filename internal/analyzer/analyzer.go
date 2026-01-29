package analyzer

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/stackrox/sensor-metrics-analyzer/internal/evaluator"
	"github.com/stackrox/sensor-metrics-analyzer/internal/loadlevel"
	"github.com/stackrox/sensor-metrics-analyzer/internal/parser"
	"github.com/stackrox/sensor-metrics-analyzer/internal/rules"
)

// Options controls analysis behavior and logging.
type Options struct {
	RulesDir           string
	LoadLevelDir       string
	ClusterName        string
	LoadLevelOverride  string
	ACSVersionOverride string
	Logger             io.Writer
}

// AnalyzeFile parses metrics and evaluates rules, returning the analysis report.
func AnalyzeFile(metricsFile string, opts Options) (rules.AnalysisReport, error) {
	logOut := opts.Logger
	if logOut == nil {
		logOut = io.Discard
	}

	rulesDir := opts.RulesDir
	if rulesDir == "" {
		return rules.AnalysisReport{}, fmt.Errorf("rules directory is required")
	}

	loadLevelDir := opts.LoadLevelDir
	if loadLevelDir == "" {
		loadLevelDir = filepath.Join(rulesDir, "load-level")
	}

	clusterName := opts.ClusterName
	if clusterName == "" {
		clusterName = ExtractClusterName(metricsFile)
	}

	fmt.Fprintf(logOut, "Loading load detection rules from %s...\n", loadLevelDir)
	loadRules, err := rules.LoadLoadDetectionRules(loadLevelDir)
	if err != nil {
		fmt.Fprintf(logOut, "Warning: Failed to load load detection rules: %v\n", err)
		loadRules = []rules.LoadDetectionRule{}
	}

	fmt.Fprintf(logOut, "Loading rules from %s...\n", rulesDir)
	rulesList, err := rules.LoadRules(rulesDir)
	if err != nil {
		return rules.AnalysisReport{}, fmt.Errorf("failed to load rules: %w", err)
	}
	fmt.Fprintf(logOut, "Loaded %d rules\n", len(rulesList))

	fmt.Fprintf(logOut, "Parsing metrics from %s...\n", metricsFile)
	metrics, err := parser.ParseFile(metricsFile)
	if err != nil {
		return rules.AnalysisReport{}, fmt.Errorf("failed to parse metrics: %w", err)
	}
	fmt.Fprintf(logOut, "Parsed %d metrics\n", len(metrics))

	acsVersion := opts.ACSVersionOverride
	if acsVersion == "" {
		if detected, ok := metrics.DetectACSVersion(); ok {
			acsVersion = detected
			fmt.Fprintf(logOut, "Detected ACS version: %s\n", acsVersion)
		} else {
			fmt.Fprintf(logOut, "Warning: Could not detect ACS version\n")
		}
	}

	loadDetector := loadlevel.NewDetector(loadRules)
	detectedLoadLevel, err := loadlevel.DetectWithOverride(metrics, loadDetector, rules.LoadLevel(opts.LoadLevelOverride))
	if err != nil {
		fmt.Fprintf(logOut, "Warning: Load level detection failed: %v\n", err)
		detectedLoadLevel = rules.LoadLevelMedium
	}
	fmt.Fprintf(logOut, "Detected load level: %s\n", detectedLoadLevel)

	fmt.Fprintf(logOut, "Evaluating rules...\n")
	report := evaluator.EvaluateAllRules(rulesList, metrics, detectedLoadLevel, acsVersion)
	report.ClusterName = clusterName

	return report, nil
}

// ExtractClusterName derives a cluster name from a file name.
func ExtractClusterName(filename string) string {
	base := filepath.Base(filename)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	name = strings.TrimSuffix(name, "-sensor-metrics")
	name = strings.TrimSuffix(name, "-metrics")
	return name
}
