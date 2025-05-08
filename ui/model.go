package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// BubbleteaModel represents the bubbletea UI state
type BubbleteaModel struct {
	Options         []string
	FilteredOptions []string
	Cursor          int
	Selected        int
	DirMap          map[string]string
	SearchQuery     string
	ShowSearch      bool
}

// Init is the bubbletea initialization function
func (m *BubbleteaModel) Init() tea.Cmd {
	// Initialize FilteredOptions with all options
	m.FilteredOptions = m.Options
	// Show search if we have more than 3 options
	m.ShowSearch = len(m.Options) > 3
	return nil
}

// FilterOptions filters the options based on search query
func (m *BubbleteaModel) FilterOptions() {
	if m.SearchQuery == "" {
		m.FilteredOptions = m.Options
		return
	}

	m.FilteredOptions = []string{}
	lowerQuery := strings.ToLower(m.SearchQuery)

	for _, opt := range m.Options {
		if strings.Contains(strings.ToLower(opt), lowerQuery) {
			m.FilteredOptions = append(m.FilteredOptions, opt)
		}
	}

	// Reset cursor if it's out of bounds
	if len(m.FilteredOptions) > 0 && m.Cursor >= len(m.FilteredOptions) {
		m.Cursor = len(m.FilteredOptions) - 1
	}
}

// Update is the bubbletea update function that handles messages
func (m *BubbleteaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.FilteredOptions)-1 {
				m.Cursor++
			}
		case "enter", " ":
			if len(m.FilteredOptions) > 0 {
				// Find the original option index that corresponds to the filtered selection
				selectedOption := m.FilteredOptions[m.Cursor]
				for i, opt := range m.Options {
					if opt == selectedOption {
						m.Selected = i
						break
					}
				}
				return m, tea.Quit
			}
		case "backspace":
			if m.ShowSearch && len(m.SearchQuery) > 0 {
				m.SearchQuery = m.SearchQuery[:len(m.SearchQuery)-1]
				m.FilterOptions()
			}
		default:
			if m.ShowSearch {
				// Check if it's a printable character
				if len(msg.String()) == 1 {
					m.SearchQuery += msg.String()
					m.FilterOptions()
					// Reset cursor position when search changes
					if len(m.FilteredOptions) > 0 {
						m.Cursor = 0
					}
				}
			}
		}
	}
	return m, nil
}

// View is the bubbletea view function that renders the UI
func (m *BubbleteaModel) View() string {
	s := "Select a repository:"

	// Show search box if enabled
	if m.ShowSearch {
		s += fmt.Sprintf("\nSearch: %s", m.SearchQuery)
	}

	s += "\n\n"

	if len(m.FilteredOptions) == 0 {
		s += "No matching repositories found.\n"
	} else {
		for i, option := range m.FilteredOptions {
			cursor := " "
			if m.Cursor == i {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, option)
		}
	}

	s += "\nPress q to quit."
	if m.ShowSearch {
		s += " Type to search."
	}
	s += "\n"
	return s
}

// InitializeModel initializes the bubbletea model
func InitializeModel(options []string, dirMap map[string]string) BubbleteaModel {
	showSearch := len(options) > 3

	return BubbleteaModel{
		Options:         options,
		FilteredOptions: options,
		Cursor:          0,
		Selected:        -1,
		DirMap:          dirMap,
		SearchQuery:     "",
		ShowSearch:      showSearch,
	}
}
