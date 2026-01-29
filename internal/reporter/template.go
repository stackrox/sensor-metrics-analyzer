package reporter

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/stackrox/sensor-metrics-analyzer/internal/rules"
)

// LoadTemplate loads a template from file
func LoadTemplate(templatePath string) (*template.Template, error) {
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	tmpl, err := template.New(filepath.Base(templatePath)).Funcs(templateFuncs()).Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return tmpl, nil
}

// templateFuncs returns template helper functions
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"formatBytes":   formatBytes,
		"formatPercent": formatPercent,
		"formatValue":   formatValue,
	}
}

// formatBytes formats byte values
func formatBytes(bytes float64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%.0fB", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1fKB", bytes/1024)
	} else if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.1fMB", bytes/(1024*1024))
	} else {
		return fmt.Sprintf("%.2fGB", bytes/(1024*1024*1024))
	}
}

// formatPercent formats percentage values
func formatPercent(value float64) string {
	return fmt.Sprintf("%.1f%%", value)
}

// formatValue formats a float value
func formatValue(value float64) string {
	return fmt.Sprintf("%.2f", value)
}

// ExecuteTemplate executes a template with data
func ExecuteTemplate(tmpl *template.Template, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	return buf.String(), nil
}

// GenerateMarkdownFromTemplate generates markdown report using template
func GenerateMarkdownFromTemplate(report rules.AnalysisReport, templatePath string) (string, error) {
	tmpl, err := LoadTemplate(templatePath)
	if err != nil {
		return "", err
	}

	data := struct {
		rules.AnalysisReport
		RedResults    []rules.EvaluationResult
		YellowResults []rules.EvaluationResult
		GreenResults  []rules.EvaluationResult
	}{
		AnalysisReport: report,
		RedResults:     filterByStatus(report.Results, rules.StatusRed),
		YellowResults:  filterByStatus(report.Results, rules.StatusYellow),
		GreenResults:   filterByStatus(report.Results, rules.StatusGreen),
	}

	return ExecuteTemplate(tmpl, data)
}
