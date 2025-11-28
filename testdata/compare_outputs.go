package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: compare_outputs <python-output> <go-output>\n")
		os.Exit(1)
	}

	pythonFile := os.Args[1]
	goFile := os.Args[2]

	// Parse Python output
	pythonStatuses, err := parsePythonOutput(pythonFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse Python output: %v\n", err)
		os.Exit(1)
	}

	// Parse Go output
	goStatuses, err := parseGoOutput(goFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse Go output: %v\n", err)
		os.Exit(1)
	}

	// Compare
	mismatches := compareStatuses(pythonStatuses, goStatuses)

	if len(mismatches) == 0 {
		fmt.Printf("✅ All statuses match!\n")
		fmt.Printf("Python: %d metrics\n", len(pythonStatuses))
		fmt.Printf("Go: %d metrics\n", len(goStatuses))
		os.Exit(0)
	}

	fmt.Fprintf(os.Stderr, "❌ Found %d mismatches:\n\n", len(mismatches))
	for metric, statuses := range mismatches {
		fmt.Fprintf(os.Stderr, "  %s: Python=%s, Go=%s\n", metric, statuses[0], statuses[1])
	}
	os.Exit(1)
}

func parsePythonOutput(filename string) (map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	statuses := make(map[string]string)
	scanner := bufio.NewScanner(file)

	// Python output format: "🔴 RED     | metric_name"
	re := regexp.MustCompile(`(🔴|🟡|🟢)\s+(RED|YELLOW|GREEN)\s+\|\s+(\S+)`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if len(matches) == 4 {
			metric := matches[3]
			status := matches[2]
			statuses[metric] = status
		}
	}

	return statuses, scanner.Err()
}

func parseGoOutput(filename string) (map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	statuses := make(map[string]string)
	scanner := bufio.NewScanner(file)

	// Go markdown format: "### metric_name" followed by "**Status:** RED"
	var currentMetric string
	statusRe := regexp.MustCompile(`\*\*Status:\*\*\s+(RED|YELLOW|GREEN)`)
	metricRe := regexp.MustCompile(`###\s+(\S+)`)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for metric name
		if matches := metricRe.FindStringSubmatch(line); len(matches) == 2 {
			currentMetric = matches[1]
		}

		// Check for status
		if matches := statusRe.FindStringSubmatch(line); len(matches) == 2 && currentMetric != "" {
			statuses[currentMetric] = matches[1]
		}
	}

	return statuses, scanner.Err()
}

func compareStatuses(python, goStatuses map[string]string) map[string][2]string {
	mismatches := make(map[string][2]string)

	// Check all Python metrics
	for metric, pythonStatus := range python {
		goStatus, exists := goStatuses[metric]
		if !exists {
			mismatches[metric] = [2]string{pythonStatus, "MISSING"}
		} else if pythonStatus != goStatus {
			mismatches[metric] = [2]string{pythonStatus, goStatus}
		}
	}

	// Check for metrics in Go but not in Python
	for metric, goStatus := range goStatuses {
		if _, exists := python[metric]; !exists {
			mismatches[metric] = [2]string{"MISSING", goStatus}
		}
	}

	return mismatches
}
