package reporter

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/stackrox/sensor-metrics-analyzer/internal/rules"
)

// GenerateConsole creates a console-formatted report with colors
func GenerateConsole(report rules.AnalysisReport) string {
	var result strings.Builder

	// Header
	result.WriteString(color.New(color.Bold).Sprint("Automated Metrics Analysis Report\n\n"))
	result.WriteString(fmt.Sprintf("Cluster: %s\n", report.ClusterName))
	result.WriteString(fmt.Sprintf("ACS Version: %s\n", report.ACSVersion))
	result.WriteString(fmt.Sprintf("Load Level: %s\n", report.LoadLevel))
	result.WriteString(fmt.Sprintf("Generated: %s\n\n", report.Timestamp.Format("2006-01-02 15:04:05")))

	// Summary table
	result.WriteString(color.New(color.Bold).Sprint("Summary\n"))
	t := table.NewWriter()
	var tableBuf bytes.Buffer
	t.SetOutputMirror(&tableBuf)
	t.AppendHeader(table.Row{"Status", "Count", "Percentage"})

	total := report.Summary.TotalAnalyzed
	if total > 0 {
		redPct := float64(report.Summary.RedCount) / float64(total) * 100
		yellowPct := float64(report.Summary.YellowCount) / float64(total) * 100
		greenPct := float64(report.Summary.GreenCount) / float64(total) * 100

		t.AppendRow(table.Row{
			color.RedString("🔴 RED"),
			report.Summary.RedCount,
			fmt.Sprintf("%.1f%%", redPct),
		})
		t.AppendRow(table.Row{
			color.YellowString("🟡 YELLOW"),
			report.Summary.YellowCount,
			fmt.Sprintf("%.1f%%", yellowPct),
		})
		t.AppendRow(table.Row{
			color.GreenString("🟢 GREEN"),
			report.Summary.GreenCount,
			fmt.Sprintf("%.1f%%", greenPct),
		})
	}

	t.SetStyle(table.StyleRounded)
	t.Render()
	result.WriteString(tableBuf.String())
	result.WriteString("\n")

	// Critical Issues
	redResults := filterByStatus(report.Results, rules.StatusRed)
	if len(redResults) > 0 {
		result.WriteString(color.New(color.Bold, color.FgRed).Sprint("🔴 Critical Issues\n\n"))
		for _, r := range redResults {
			result.WriteString(color.New(color.Bold).Sprintf("%s\n", r.RuleName))
			result.WriteString(color.RedString("  Status: RED\n"))
			result.WriteString(fmt.Sprintf("  Message: %s\n", r.Message))
			if len(r.Details) > 0 {
				result.WriteString(color.New(color.FgYellow).Sprint("  Details:\n"))
				for _, detail := range r.Details {
					result.WriteString(fmt.Sprintf("    %s\n", detail))
				}
			}
			if r.PotentialActionUser != "" {
				result.WriteString(fmt.Sprintf("  %s %s\n",
					color.New(color.FgYellow).Sprint("Potential action:"),
					r.PotentialActionUser))
			}
			if r.PotentialActionDeveloper != "" {
				result.WriteString(fmt.Sprintf("  %s %s\n",
					color.New(color.FgYellow).Sprint("Potential action (developer):"),
					r.PotentialActionDeveloper))
			}
			result.WriteString("\n")
		}
	}

	// Warnings
	yellowResults := filterByStatus(report.Results, rules.StatusYellow)
	if len(yellowResults) > 0 {
		result.WriteString(color.New(color.Bold, color.FgYellow).Sprint("🟡 Warnings\n\n"))
		for _, r := range yellowResults {
			result.WriteString(color.New(color.Bold).Sprintf("%s\n", r.RuleName))
			result.WriteString(color.YellowString("  Status: YELLOW\n"))
			result.WriteString(fmt.Sprintf("  Message: %s\n", r.Message))
			if len(r.Details) > 0 {
				result.WriteString(color.New(color.FgYellow).Sprint("  Details:\n"))
				for _, detail := range r.Details {
					result.WriteString(fmt.Sprintf("    %s\n", detail))
				}
			}
			if r.PotentialActionUser != "" {
				result.WriteString(fmt.Sprintf("  %s %s\n",
					color.New(color.FgYellow).Sprint("Potential action:"),
					r.PotentialActionUser))
			}
			if r.PotentialActionDeveloper != "" {
				result.WriteString(fmt.Sprintf("  %s %s\n",
					color.New(color.FgYellow).Sprint("Potential action (developer):"),
					r.PotentialActionDeveloper))
			}
			result.WriteString("\n")
		}
	}

	// Healthy Metrics (compact)
	greenResults := filterByStatus(report.Results, rules.StatusGreen)
	if len(greenResults) > 0 {
		result.WriteString(color.New(color.Bold, color.FgGreen).Sprint("🟢 Healthy Metrics\n\n"))
		for _, r := range greenResults {
			result.WriteString(color.GreenString("  ✓ "))
			result.WriteString(fmt.Sprintf("%s: %s\n", r.RuleName, r.Message))
		}
	}

	return result.String()
}

// PrintConsole prints console report to stdout
func PrintConsole(report rules.AnalysisReport) {
	output := GenerateConsole(report)
	fmt.Fprint(os.Stdout, output)
}
