package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/stackrox/sensor-metrics-analyzer/internal/evaluator"
	"github.com/stackrox/sensor-metrics-analyzer/internal/loadlevel"
	"github.com/stackrox/sensor-metrics-analyzer/internal/parser"
	"github.com/stackrox/sensor-metrics-analyzer/internal/reporter"
	"github.com/stackrox/sensor-metrics-analyzer/internal/rules"
	"github.com/stackrox/sensor-metrics-analyzer/internal/tui"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "analyze":
		analyzeCommand()
	case "validate":
		validateCommand()
	case "list-rules":
		listRulesCommand()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func analyzeCommand() {
	fs := flag.NewFlagSet("analyze", flag.ExitOnError)
	rulesDir := fs.String("rules", ".", "Directory containing TOML rules (default: current directory)")
	loadLevelDir := fs.String("load-level-dir", "./load-level", "Directory containing load detection rules")
	output := fs.String("output", "", "Output file (default: stdout)")
	format := fs.String("format", "console", "Output format: console, markdown, tui (interactive)")
	clusterName := fs.String("cluster", "", "Cluster name (extracted from filename if not provided)")
	loadLevelOverride := fs.String("load-level", "", "Override detected load level (low/medium/high)")
	acsVersionOverride := fs.String("acs-version", "", "Override detected ACS version")
	templatePath := fs.String("template", "./templates/markdown.tmpl", "Path to markdown template")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: metrics-analyzer analyze [flags] <metrics-file>\n\n")
		fmt.Fprintf(os.Stderr, "Analyzes Prometheus metrics using declarative TOML rules.\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  metrics-file       Path to Prometheus metrics file\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n⚠️  Note: Flags must come BEFORE the metrics file!\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  metrics-analyzer analyze metrics.txt\n")
		fmt.Fprintf(os.Stderr, "  metrics-analyzer analyze --rules ./automated-rules metrics.txt\n")
		fmt.Fprintf(os.Stderr, "  metrics-analyzer analyze --format markdown --output report.md metrics.txt\n")
		fmt.Fprintf(os.Stderr, "  metrics-analyzer analyze --format tui --rules ./automated-rules metrics.txt\n")
	}

	fs.Parse(os.Args[2:])

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: missing metrics file\n")
		fmt.Fprintf(os.Stderr, "Usage: metrics-analyzer analyze [flags] <metrics-file>\n")
		os.Exit(1)
	}

	metricsFile := fs.Arg(0)

	// Check for flags after positional argument (common mistake)
	for i := 1; i < fs.NArg(); i++ {
		arg := fs.Arg(i)
		if strings.HasPrefix(arg, "-") {
			fmt.Fprintf(os.Stderr, "Error: flags must come before the metrics file, not after\n")
			fmt.Fprintf(os.Stderr, "  Found flag '%s' after '%s'\n\n", arg, metricsFile)
			fmt.Fprintf(os.Stderr, "Correct usage:\n")
			fmt.Fprintf(os.Stderr, "  metrics-analyzer analyze [flags] <metrics-file>\n\n")
			fmt.Fprintf(os.Stderr, "Example:\n")
			fmt.Fprintf(os.Stderr, "  metrics-analyzer analyze --format tui --rules ./automated-rules %s\n", metricsFile)
			os.Exit(1)
		}
	}

	// Extract cluster name from filename if not provided
	if *clusterName == "" {
		*clusterName = extractClusterName(metricsFile)
	}

	// Load load detection rules
	fmt.Fprintf(os.Stderr, "Loading load detection rules from %s...\n", *loadLevelDir)
	loadRules, err := rules.LoadLoadDetectionRules(*loadLevelDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load load detection rules: %v\n", err)
		loadRules = []rules.LoadDetectionRule{}
	}

	// Load evaluation rules
	fmt.Fprintf(os.Stderr, "Loading rules from %s...\n", *rulesDir)
	rulesList, err := rules.LoadRules(*rulesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load rules: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Loaded %d rules\n", len(rulesList))

	// Parse metrics
	fmt.Fprintf(os.Stderr, "Parsing metrics from %s...\n", metricsFile)
	metrics, err := parser.ParseFile(metricsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse metrics: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Parsed %d metrics\n", len(metrics))

	// Detect ACS version
	acsVersion := *acsVersionOverride
	if acsVersion == "" {
		if detected, ok := metrics.DetectACSVersion(); ok {
			acsVersion = detected
			fmt.Fprintf(os.Stderr, "Detected ACS version: %s\n", acsVersion)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: Could not detect ACS version\n")
		}
	}

	// Detect load level
	loadDetector := loadlevel.NewDetector(loadRules)
	detectedLoadLevel, err := loadlevel.DetectWithOverride(metrics, loadDetector, rules.LoadLevel(*loadLevelOverride))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Load level detection failed: %v\n", err)
		detectedLoadLevel = rules.LoadLevelMedium
	}
	fmt.Fprintf(os.Stderr, "Detected load level: %s\n", detectedLoadLevel)

	// Evaluate all rules
	fmt.Fprintf(os.Stderr, "Evaluating rules...\n")
	report := evaluator.EvaluateAllRules(rulesList, metrics, detectedLoadLevel, acsVersion)
	report.ClusterName = *clusterName

	// Generate report
	var outputContent string
	switch *format {
	case "tui":
		// Interactive TUI mode
		if *output != "" {
			fmt.Fprintf(os.Stderr, "Warning: --output is ignored in TUI mode\n")
		}
		if err := tui.Run(report); err != nil {
			fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
			os.Exit(1)
		}
		return
	case "console":
		// If output file specified, still use console format
		if *output != "" {
			outputContent = reporter.GenerateConsole(report)
		} else {
			reporter.PrintConsole(report)
			return
		}
	case "markdown":
		outputContent = reporter.GenerateMarkdown(report, *templatePath)
		if outputContent == "" {
			fmt.Fprintf(os.Stderr, "Warning: Markdown generation returned empty content\n")
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %s\n", *format)
		os.Exit(1)
	}

	// Write output
	if *output == "" {
		fmt.Print(outputContent)
	} else {
		err := os.WriteFile(*output, []byte(outputContent), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write output: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Report written to %s\n", *output)
	}
}

func validateCommand() {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: metrics-analyzer validate [rules-directory]\n\n")
		fmt.Fprintf(os.Stderr, "Validates TOML rule files in the specified directory.\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  rules-directory    Directory containing TOML rule files (default: ./automated-rules)\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  metrics-analyzer validate\n")
		fmt.Fprintf(os.Stderr, "  metrics-analyzer validate ./automated-rules\n")
		fmt.Fprintf(os.Stderr, "  metrics-analyzer validate --help\n")
	}

	fs.Parse(os.Args[2:])

	rulesDir := "./automated-rules"
	if fs.NArg() > 0 {
		rulesDir = fs.Arg(0)
	}

	fmt.Printf("Validating rules in %s...\n", rulesDir)

	rulesList, err := rules.LoadRules(rulesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ All %d rules are valid!\n", len(rulesList))
}

func listRulesCommand() {
	fs := flag.NewFlagSet("list-rules", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: metrics-analyzer list-rules [rules-directory]\n\n")
		fmt.Fprintf(os.Stderr, "Lists all available TOML rules in the specified directory.\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  rules-directory    Directory containing TOML rule files (default: ./automated-rules)\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  metrics-analyzer list-rules\n")
		fmt.Fprintf(os.Stderr, "  metrics-analyzer list-rules ./automated-rules\n")
		fmt.Fprintf(os.Stderr, "  metrics-analyzer list-rules --help\n")
	}

	fs.Parse(os.Args[2:])

	rulesDir := "./automated-rules"
	if fs.NArg() > 0 {
		rulesDir = fs.Arg(0)
	}

	rulesList, err := rules.LoadRules(rulesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load rules: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d rules:\n\n", len(rulesList))
	for _, rule := range rulesList {
		displayName := rule.DisplayName
		if displayName == "" {
			displayName = rule.MetricName
		}
		fmt.Printf("- %s (%s): %s\n", displayName, rule.RuleType, rule.Description)
	}
}

func extractClusterName(filename string) string {
	base := filepath.Base(filename)
	// Remove extension
	name := strings.TrimSuffix(base, filepath.Ext(base))
	// Remove common prefixes/suffixes
	name = strings.TrimSuffix(name, "-sensor-metrics")
	name = strings.TrimSuffix(name, "-metrics")
	return name
}

func printUsage() {
	fmt.Println("Usage: metrics-analyzer <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  analyze      Analyze a Prometheus metrics file")
	fmt.Println("  validate     Validate TOML rule files")
	fmt.Println("  list-rules   List all available rules")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  metrics-analyzer analyze metrics.txt")
	fmt.Println("  metrics-analyzer analyze --rules ./automated-rules metrics.txt")
	fmt.Println("  metrics-analyzer analyze --format markdown --output report.md metrics.txt")
	fmt.Println("  metrics-analyzer analyze --format tui --rules ./automated-rules metrics.txt")
	fmt.Println("  metrics-analyzer analyze --load-level high --acs-version 4.8 metrics.txt")
	fmt.Println("  metrics-analyzer validate")
	fmt.Println("  metrics-analyzer validate ./automated-rules")
	fmt.Println("  metrics-analyzer list-rules")
	fmt.Println()
	fmt.Println("Note: Flags must come BEFORE positional arguments!")
}
