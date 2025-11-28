package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil
	}

	// Handle text input updates when filtering
	if m.filtering {
		var cmd tea.Cmd
		m.filterInput, cmd = m.filterInput.Update(msg)
		m.filterText = m.filterInput.Value()
		m.applyFilter()
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys that work in any mode
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	}

	// Handle filtering mode
	if m.filtering {
		switch msg.String() {
		case "esc":
			m.filtering = false
			m.filterInput.Blur()
			return m, nil
		case "enter":
			m.filtering = false
			m.filterInput.Blur()
			return m, nil
		}
		// Let the text input handle other keys
		var cmd tea.Cmd
		m.filterInput, cmd = m.filterInput.Update(msg)
		m.filterText = m.filterInput.Value()
		m.applyFilter()
		return m, cmd
	}

	// Mode-specific handling
	switch m.viewMode {
	case ViewList:
		return m.handleListKeys(msg)
	case ViewDetail:
		return m.handleDetailKeys(msg)
	case ViewHelp:
		return m.handleHelpKeys(msg)
	}

	return m, nil
}

func (m Model) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.filteredResults)-1 {
			m.cursor++
		}

	case "home", "g":
		m.cursor = 0

	case "end", "G":
		m.cursor = max(0, len(m.filteredResults)-1)

	case "pgup":
		m.cursor = max(0, m.cursor-10)

	case "pgdown":
		m.cursor = min(len(m.filteredResults)-1, m.cursor+10)

	case "enter", "l", "right":
		if len(m.filteredResults) > 0 {
			m.viewMode = ViewDetail
		}

	case "/":
		m.filtering = true
		m.filterInput.Focus()
		return m, nil

	case "?":
		m.viewMode = ViewHelp

	case "1":
		m.filterMode = FilterAll
		m.applyFilter()

	case "2":
		m.filterMode = FilterRed
		m.applyFilter()

	case "3":
		m.filterMode = FilterYellow
		m.applyFilter()

	case "4":
		m.filterMode = FilterGreen
		m.applyFilter()

	case "esc":
		// Clear filter
		m.filterText = ""
		m.filterInput.SetValue("")
		m.filterMode = FilterAll
		m.applyFilter()
	}

	return m, nil
}

func (m Model) handleDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "esc", "h", "left", "backspace":
		m.viewMode = ViewList

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.filteredResults)-1 {
			m.cursor++
		}

	case "?":
		m.viewMode = ViewHelp
	}

	return m, nil
}

func (m Model) handleHelpKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "esc", "?", "enter":
		m.viewMode = ViewList
	}

	return m, nil
}
