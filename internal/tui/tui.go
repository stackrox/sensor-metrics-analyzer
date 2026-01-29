package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stackrox/sensor-metrics-analyzer/internal/rules"
	"golang.org/x/term"
)

// IsTerminal returns true if stdout is a terminal
func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// Run starts the interactive TUI with the given analysis report
func Run(report rules.AnalysisReport) error {
	// Check if we're running in a terminal
	if !IsTerminal() {
		return fmt.Errorf("TUI mode requires an interactive terminal (stdout is not a TTY)")
	}

	model := NewModel(report)

	p := tea.NewProgram(
		model,
		// Avoid alternate screen and mouse capture so text is copyable
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}

// RunWithOutput runs the TUI and returns when the user quits
// This is useful for testing or when you need to do something after the TUI exits
func RunWithOutput(report rules.AnalysisReport) (*Model, error) {
	model := NewModel(report)

	p := tea.NewProgram(
		model,
		// Avoid alternate screen and mouse capture so text is copyable
	)

	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("error running TUI: %w", err)
	}

	if m, ok := finalModel.(Model); ok {
		return &m, nil
	}

	return nil, fmt.Errorf("unexpected model type")
}

// Preview renders a static preview of the TUI (useful for screenshots/docs)
func Preview(report rules.AnalysisReport, width, height int) string {
	model := NewModel(report)
	model.width = width
	model.height = height
	model.ready = true
	return model.View()
}

// PrintPreview prints a preview to stdout (non-interactive)
func PrintPreview(report rules.AnalysisReport) {
	preview := Preview(report, 100, 40)
	fmt.Fprint(os.Stdout, preview)
}
