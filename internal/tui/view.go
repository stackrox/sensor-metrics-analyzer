package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const logo = `
 ___  ___ _ __  ___  ___  _ __ 
/ __|/ _ \ '_ \/ __|/ _ \| '__|
\__ \  __/ | | \__ \ (_) | |   
|___/\___|_| |_|___/\___/|_|   
      metrics analyzer 🔍
`

// View renders the UI
func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	switch m.viewMode {
	case ViewHelp:
		return m.viewHelp()
	case ViewDetail:
		return m.viewDetail()
	default:
		return m.viewList()
	}
}

func (m Model) viewList() string {
	var b strings.Builder

	// Header with logo
	b.WriteString(logoStyle.Render(logo))
	b.WriteString("\n")

	// Info box
	infoContent := fmt.Sprintf(
		"%s %s  │  %s %s  │  %s %s",
		detailLabelStyle.Render("Cluster:"),
		detailValueStyle.Render(m.report.ClusterName),
		detailLabelStyle.Render("ACS:"),
		detailValueStyle.Render(m.report.ACSVersion),
		detailLabelStyle.Render("Load:"),
		detailValueStyle.Render(string(m.report.LoadLevel)),
	)
	b.WriteString(headerBoxStyle.Render(infoContent))
	b.WriteString("\n")

	// Summary
	summary := fmt.Sprintf(
		"  %s %s   %s %s   %s %s   │   Total: %d",
		"🔴",
		redCountStyle.Render(fmt.Sprintf("%d", m.report.Summary.RedCount)),
		"🟡",
		yellowCountStyle.Render(fmt.Sprintf("%d", m.report.Summary.YellowCount)),
		"🟢",
		greenCountStyle.Render(fmt.Sprintf("%d", m.report.Summary.GreenCount)),
		m.report.Summary.TotalAnalyzed,
	)
	b.WriteString(summaryStyle.Render(summary))
	b.WriteString("\n")

	// Filter tabs
	tabs := m.renderFilterTabs()
	b.WriteString(tabs)
	b.WriteString("\n")

	// Filter input (if active)
	if m.filtering {
		b.WriteString(filterPromptStyle.Render("🔍 "))
		b.WriteString(m.filterInput.View())
		b.WriteString("\n")
	} else if m.filterText != "" {
		b.WriteString(filterPromptStyle.Render(fmt.Sprintf("🔍 Filter: %s", m.filterText)))
		b.WriteString("\n")
	}

	// Results list
	listContent := m.renderResultsList()
	b.WriteString(listStyle.Render(listContent))
	b.WriteString("\n")

	// Help bar
	b.WriteString(m.renderHelp())

	return appStyle.Render(b.String())
}

func (m Model) renderFilterTabs() string {
	var tabs []string

	if m.filterMode == FilterAll {
		tabs = append(tabs, activeTabStyle.Render("1:All"))
	} else {
		tabs = append(tabs, inactiveTabStyle.Render("1:All"))
	}

	if m.filterMode == FilterRed {
		tabs = append(tabs, activeTabStyle.Render("2:🔴 Red"))
	} else {
		tabs = append(tabs, inactiveTabStyle.Render("2:🔴 Red"))
	}

	if m.filterMode == FilterYellow {
		tabs = append(tabs, activeTabStyle.Render("3:🟡 Yellow"))
	} else {
		tabs = append(tabs, inactiveTabStyle.Render("3:🟡 Yellow"))
	}

	if m.filterMode == FilterGreen {
		tabs = append(tabs, activeTabStyle.Render("4:🟢 Green"))
	} else {
		tabs = append(tabs, inactiveTabStyle.Render("4:🟢 Green"))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (m Model) renderResultsList() string {
	if len(m.filteredResults) == 0 {
		return detailLabelStyle.Render("  No results match the current filter")
	}

	var lines []string

	// Calculate visible window
	visibleHeight := 15 // Number of visible items
	start := 0
	if m.cursor >= visibleHeight {
		start = m.cursor - visibleHeight + 1
	}
	end := min(start+visibleHeight, len(m.filteredResults))

	// Show scroll indicator if needed
	if start > 0 {
		lines = append(lines, detailLabelStyle.Render(fmt.Sprintf("  ↑ %d more above", start)))
	}

	for i := start; i < end; i++ {
		r := m.filteredResults[i]

		// Truncate name if too long
		name := r.RuleName
		if len(name) > 35 {
			name = name[:32] + "..."
		}

		// Truncate message if too long
		msg := r.Message
		if len(msg) > 40 {
			msg = msg[:37] + "..."
		}

		line := fmt.Sprintf("%s %-35s %s", StatusEmoji(string(r.Status)), name, msg)

		if i == m.cursor {
			lines = append(lines, selectedItemStyle.Render(line))
		} else {
			lines = append(lines, normalItemStyle.Render(line))
		}
	}

	// Show scroll indicator if needed
	if end < len(m.filteredResults) {
		lines = append(lines, detailLabelStyle.Render(fmt.Sprintf("  ↓ %d more below", len(m.filteredResults)-end)))
	}

	return strings.Join(lines, "\n")
}

func (m Model) viewDetail() string {
	var b strings.Builder

	result := m.selectedResult()
	if result == nil {
		return "No result selected"
	}

	// Header
	b.WriteString(logoStyle.Render(logo))
	b.WriteString("\n")

	// Navigation info
	navInfo := fmt.Sprintf("  Result %d of %d", m.cursor+1, len(m.filteredResults))
	b.WriteString(detailLabelStyle.Render(navInfo))
	b.WriteString("\n\n")

	// Detail card
	var detail strings.Builder

	// Title with status
	titleLine := fmt.Sprintf("%s  %s", StatusBadge(string(result.Status)), result.RuleName)
	detail.WriteString(detailTitleStyle.Render(titleLine))
	detail.WriteString("\n\n")

	// Status
	detail.WriteString(detailLabelStyle.Render("Status:     "))
	switch result.Status {
	case "RED":
		detail.WriteString(redCountStyle.Render("● RED - Critical"))
	case "YELLOW":
		detail.WriteString(yellowCountStyle.Render("● YELLOW - Warning"))
	case "GREEN":
		detail.WriteString(greenCountStyle.Render("● GREEN - Healthy"))
	}
	detail.WriteString("\n\n")

	// Message (wrapped to screen width)
	detail.WriteString(detailLabelStyle.Render("Message:"))
	detail.WriteString("\n")
	messageWidth := 70
	if m.width > 0 {
		messageWidth = m.width - 10
		if messageWidth < 40 {
			messageWidth = 40
		}
	}
	wrappedMessage := wordWrap(result.Message, messageWidth)
	for _, line := range strings.Split(wrappedMessage, "\n") {
		detail.WriteString(fmt.Sprintf("  %s\n", line))
	}
	detail.WriteString("\n")

	// Value
	if result.Value != 0 {
		detail.WriteString(detailLabelStyle.Render("Value:      "))
		detail.WriteString(detailValueStyle.Render(fmt.Sprintf("%.4f", result.Value)))
		detail.WriteString("\n\n")
	}

	// Details map - display all details in plain text format for easy copying
	if len(result.Details) > 0 {
		detail.WriteString(detailLabelStyle.Render("Details:"))
		detail.WriteString("\n")

		// Sort keys for consistent ordering
		keys := make([]string, 0, len(result.Details))
		for k := range result.Details {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// Display details in plain text format (no ANSI codes) for easy selection/copying
		for _, k := range keys {
			v := result.Details[k]
			// Format the value nicely
			var formattedValue string
			switch val := v.(type) {
			case float64:
				// Format floats with appropriate precision
				if val >= 1000000 {
					formattedValue = fmt.Sprintf("%.0f", val)
				} else if val >= 1000 {
					formattedValue = fmt.Sprintf("%.0f", val)
				} else if val >= 1 {
					formattedValue = fmt.Sprintf("%.2f", val)
				} else {
					formattedValue = fmt.Sprintf("%.4f", val)
				}
			case int:
				formattedValue = fmt.Sprintf("%d", val)
			case int64:
				formattedValue = fmt.Sprintf("%d", val)
			default:
				// For other types, convert to string as-is (no truncation)
				formattedValue = fmt.Sprintf("%v", v)
			}

			// Plain text format: key: value (completely plain text for easy copying)
			detail.WriteString(fmt.Sprintf("  %s: %s\n", k, formattedValue))
		}
		detail.WriteString("\n")
	}

	// Potential actions (user/developer)
	if result.PotentialActionUser != "" && (result.Status == "RED" || result.Status == "YELLOW") {
		detail.WriteString(detailLabelStyle.Render("Potential action:"))
		detail.WriteString("\n")
		wrapped := wordWrap(result.PotentialActionUser, 60)
		for _, line := range strings.Split(wrapped, "\n") {
			detail.WriteString(remediationStyle.Render(fmt.Sprintf("  %s", line)))
			detail.WriteString("\n")
		}
	}
	if result.PotentialActionDeveloper != "" && (result.Status == "RED" || result.Status == "YELLOW") {
		detail.WriteString(detailLabelStyle.Render("Potential action (developer):"))
		detail.WriteString("\n")
		wrapped := wordWrap(result.PotentialActionDeveloper, 60)
		for _, line := range strings.Split(wrapped, "\n") {
			detail.WriteString(remediationStyle.Render(fmt.Sprintf("  %s", line)))
			detail.WriteString("\n")
		}
	}

	b.WriteString(detailBoxStyle.Render(detail.String()))
	b.WriteString("\n")

	// Help
	help := fmt.Sprintf(
		"%s back  %s/%s prev/next  %s quit",
		helpKeyStyle.Render("←/esc"),
		helpKeyStyle.Render("↑"),
		helpKeyStyle.Render("↓"),
		helpKeyStyle.Render("q"),
	)
	b.WriteString(helpStyle.Render(help))

	return appStyle.Render(b.String())
}

func (m Model) viewHelp() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(" 🎮 Keyboard Shortcuts "))
	b.WriteString("\n\n")

	helpItems := []struct {
		key  string
		desc string
	}{
		{"↑/k, ↓/j", "Navigate up/down"},
		{"Enter/→", "View details"},
		{"←/Esc", "Go back"},
		{"g/Home", "Go to top"},
		{"G/End", "Go to bottom"},
		{"PgUp/PgDn", "Page up/down"},
		{"/", "Search/filter"},
		{"1-4", "Filter by status (All/Red/Yellow/Green)"},
		{"Esc", "Clear filter"},
		{"?", "Toggle help"},
		{"q", "Quit"},
	}

	for _, item := range helpItems {
		b.WriteString(fmt.Sprintf("  %s  %s\n",
			helpKeyStyle.Render(fmt.Sprintf("%-12s", item.key)),
			helpDescStyle.Render(item.desc),
		))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Press any key to return..."))

	return appStyle.Render(b.String())
}

func (m Model) renderHelp() string {
	if m.filtering {
		return helpStyle.Render(
			fmt.Sprintf("%s confirm  %s cancel",
				helpKeyStyle.Render("Enter"),
				helpKeyStyle.Render("Esc"),
			),
		)
	}

	return helpStyle.Render(
		fmt.Sprintf("%s navigate  %s details  %s search  %s filter  %s help  %s quit",
			helpKeyStyle.Render("↑↓"),
			helpKeyStyle.Render("Enter"),
			helpKeyStyle.Render("/"),
			helpKeyStyle.Render("1-4"),
			helpKeyStyle.Render("?"),
			helpKeyStyle.Render("q"),
		),
	)
}

// wordWrap wraps text at the specified width
func wordWrap(text string, width int) string {
	if len(text) <= width {
		return text
	}

	var result strings.Builder
	words := strings.Fields(text)
	lineLen := 0

	for i, word := range words {
		if lineLen+len(word)+1 > width && lineLen > 0 {
			result.WriteString("\n")
			lineLen = 0
		}
		if lineLen > 0 {
			result.WriteString(" ")
			lineLen++
		}
		result.WriteString(word)
		lineLen += len(word)
		_ = i // unused
	}

	return result.String()
}
