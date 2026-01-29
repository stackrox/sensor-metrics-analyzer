package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette - Cyberpunk/Neon theme
var (
	// Primary colors
	colorPink       = lipgloss.Color("#FF6AC1")
	colorCyan       = lipgloss.Color("#9AEDFE")
	colorYellow     = lipgloss.Color("#F1FA8C")
	colorGreen      = lipgloss.Color("#50FA7B")
	colorRed        = lipgloss.Color("#FF5555")
	colorOrange     = lipgloss.Color("#FFB86C")
	colorPurple     = lipgloss.Color("#BD93F9")
	colorComment    = lipgloss.Color("#6272A4")
	colorForeground = lipgloss.Color("#F8F8F2")
	colorBackground = lipgloss.Color("#282A36")
	colorSelection  = lipgloss.Color("#44475A")

	// Status colors
	statusRed    = lipgloss.Color("#FF5555")
	statusYellow = lipgloss.Color("#F1FA8C")
	statusGreen  = lipgloss.Color("#50FA7B")
)

// Styles
var (
	// App frame
	appStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Header styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPink).
			Background(colorSelection).
			Padding(0, 2).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorComment).
			Italic(true)

	// Box for header info
	headerBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPurple).
			Padding(0, 1).
			MarginBottom(1)

	// Summary stats
	summaryStyle = lipgloss.NewStyle().
			Padding(0, 1).
			MarginBottom(1)

	redCountStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(statusRed)

	yellowCountStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(statusYellow)

	greenCountStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(statusGreen)

	// List styles
	listStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorCyan).
			Padding(0, 1).
			MarginBottom(1)

	selectedItemStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorBackground).
				Background(colorCyan).
				Padding(0, 1)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(colorForeground).
			Padding(0, 1)

	// Status badges
	redBadgeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(statusRed).
			Padding(0, 1)

	yellowBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorBackground).
				Background(statusYellow).
				Padding(0, 1)

	greenBadgeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBackground).
			Background(statusGreen).
			Padding(0, 1)

	// Detail panel
	detailBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorOrange).
			Padding(1, 2)

	detailTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorOrange).
				MarginBottom(1)

	detailLabelStyle = lipgloss.NewStyle().
				Foreground(colorComment)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(colorForeground)

	remediationStyle = lipgloss.NewStyle().
				Foreground(colorYellow).
				Italic(true)

	// Help bar
	helpStyle = lipgloss.NewStyle().
			Foreground(colorComment).
			MarginTop(1)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(colorPink).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(colorComment)

	// Filter input
	filterPromptStyle = lipgloss.NewStyle().
				Foreground(colorCyan).
				Bold(true)

	filterInputStyle = lipgloss.NewStyle().
				Foreground(colorForeground)

	// Logo/ASCII art
	logoStyle = lipgloss.NewStyle().
			Foreground(colorPink).
			Bold(true)

	// Tab styles
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBackground).
			Background(colorCyan).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(colorComment).
				Padding(0, 2)

	// Progress indicator
	spinnerStyle = lipgloss.NewStyle().
			Foreground(colorPink)
)

// StatusBadge returns a styled badge for the given status
func StatusBadge(status string) string {
	switch status {
	case "RED":
		return redBadgeStyle.Render("RED")
	case "YELLOW":
		return yellowBadgeStyle.Render("YEL")
	case "GREEN":
		return greenBadgeStyle.Render("GRN")
	default:
		return status
	}
}

// StatusEmoji returns an emoji for the given status
func StatusEmoji(status string) string {
	switch status {
	case "RED":
		return "🔴"
	case "YELLOW":
		return "🟡"
	case "GREEN":
		return "🟢"
	default:
		return "⚪"
	}
}
