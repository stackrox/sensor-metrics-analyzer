package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/stackrox/sensor-metrics-analyzer/internal/rules"
)

// ViewMode represents the current view
type ViewMode int

const (
	ViewList ViewMode = iota
	ViewDetail
	ViewHelp
)

// FilterMode represents what to filter by
type FilterMode int

const (
	FilterAll FilterMode = iota
	FilterRed
	FilterYellow
	FilterGreen
)

// Model represents the application state
type Model struct {
	// Data
	report  rules.AnalysisReport
	results []rules.EvaluationResult

	// UI state
	cursor      int
	viewMode    ViewMode
	filterMode  FilterMode
	filterInput textinput.Model
	filtering   bool
	filterText  string
	width       int
	height      int
	ready       bool

	// Filtered results
	filteredResults []rules.EvaluationResult
}

// NewModel creates a new TUI model with the given report
func NewModel(report rules.AnalysisReport) Model {
	ti := textinput.New()
	ti.Placeholder = "Type to filter..."
	ti.CharLimit = 50
	ti.Width = 30

	m := Model{
		report:      report,
		results:     report.Results,
		filterInput: ti,
		filterMode:  FilterAll,
		viewMode:    ViewList,
	}

	m.applyFilter()
	return m
}

// applyFilter filters results based on current filter settings
func (m *Model) applyFilter() {
	m.filteredResults = make([]rules.EvaluationResult, 0)

	for _, r := range m.results {
		// Status filter
		if m.filterMode != FilterAll {
			switch m.filterMode {
			case FilterRed:
				if r.Status != rules.StatusRed {
					continue
				}
			case FilterYellow:
				if r.Status != rules.StatusYellow {
					continue
				}
			case FilterGreen:
				if r.Status != rules.StatusGreen {
					continue
				}
			}
		}

		// Text filter
		if m.filterText != "" {
			if !containsIgnoreCase(r.RuleName, m.filterText) &&
				!containsIgnoreCase(r.Message, m.filterText) {
				continue
			}
		}

		m.filteredResults = append(m.filteredResults, r)
	}

	// Reset cursor if out of bounds
	if m.cursor >= len(m.filteredResults) {
		m.cursor = max(0, len(m.filteredResults)-1)
	}
}

// selectedResult returns the currently selected result
func (m *Model) selectedResult() *rules.EvaluationResult {
	if len(m.filteredResults) == 0 || m.cursor >= len(m.filteredResults) {
		return nil
	}
	return &m.filteredResults[m.cursor]
}

// Helper functions
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(substr) == 0 ||
			containsLower(toLower(s), toLower(substr)))
}

func containsLower(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
