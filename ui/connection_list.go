package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jorgerojas26/lazysql/app"
	"github.com/jorgerojas26/lazysql/models"
)

type ConnectionListModel struct {
	connections []models.Connection
	cursor      int
	width       int
	height      int
	status      string
	statusErr   bool
}

func NewConnectionListModel() ConnectionListModel {
	return ConnectionListModel{
		connections: app.App.Connections(),
	}
}

func (m ConnectionListModel) Init() tea.Cmd {
	return nil
}

func (m ConnectionListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.connections)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.connections) > 0 {
				conn := m.connections[m.cursor]
				return m, func() tea.Msg {
					return ScreenChangeMsg{Screen: ScreenHome, Data: conn}
				}
			}
		case "n":
			return m, func() tea.Msg {
				return ScreenChangeMsg{Screen: ScreenConnectionForm, Data: nil}
			}
		case "e":
			if len(m.connections) > 0 {
				conn := m.connections[m.cursor]
				return m, func() tea.Msg {
					return ScreenChangeMsg{Screen: ScreenConnectionForm, Data: conn}
				}
			}
		case "d":
			if len(m.connections) > 0 {
				conns := append(m.connections[:m.cursor], m.connections[m.cursor+1:]...)
				if err := app.App.SaveConnections(conns); err != nil {
					m.status = err.Error()
					m.statusErr = true
				} else {
					m.connections = conns
					if m.cursor >= len(m.connections) {
						m.cursor = len(m.connections) - 1
					}
				}
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m ConnectionListModel) View() string {
	header := TitleStyle.Render("Connections")
	subtitle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("Select a connection to start")

	rows := make([]string, len(m.connections))
	for i, conn := range m.connections {
		name := conn.Name
		if name == "" {
			name = conn.URL
		}
		profileCount := ""
		if len(conn.Profiles) > 1 {
			profileCount = fmt.Sprintf(" (%d profiles)", len(conn.Profiles))
		}

		provider := conn.Provider
		style := lipgloss.NewStyle()
		if i == m.cursor {
			style = SelectedStyle
		}
		rows[i] = style.Render(fmt.Sprintf("  %-20s %-12s%s", name, provider, profileCount))
	}

	if len(rows) == 0 {
		rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("  No connections. Press N to add one."))
	}

	table := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666A7E")).
		Padding(0, 1).
		Render(lipgloss.JoinVertical(lipgloss.Left, rows...))

	help := HelpStyle.Render("[N] New  [E] Edit  [D] Delete  [Enter] Connect  [Q] Quit")

	statusBar := ""
	if m.status != "" {
		style := StatusStyle
		if m.statusErr {
			style = ErrorStyle
		}
		statusBar = style.Render(m.status)
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		subtitle,
		"",
		table,
		"",
		help,
		statusBar,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}