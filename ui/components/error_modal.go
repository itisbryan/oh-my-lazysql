package components

import (
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

type ErrorModalModel struct {
	title     string
	message   string
	open      bool
	focused   bool
	width     int
	height    int
	onClose   func() tea.Msg
}

func NewErrorModal(title, message string) ErrorModalModel {
	return ErrorModalModel{
		title:   title,
		message: message,
		open:    true,
		width:   50,
		height:  12,
	}
}

func (m ErrorModalModel) Init() tea.Cmd {
	return nil
}

func (m ErrorModalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if !m.open {
			return m, nil
		}
		switch msg.String() {
		case "enter", "q", "esc":
			if m.onClose != nil {
				return m, m.onClose
			}
			m.open = false
		}
	}
	return m, nil
}

func (m ErrorModalModel) View() string {
	if !m.open {
		return ""
	}

	box := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF0000")).
		Padding(1, 2).
		Background(lipgloss.Color("#1E1E1E"))

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF0000")).Render(m.title)
	errIcon := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render("✗")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		errIcon+" "+m.message,
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("[Enter/Q] Close"),
	)

	return lipgloss.Place(m.width+4, m.height+4, lipgloss.Center, lipgloss.Center, box.Render(content))
}

func (m *ErrorModalModel) SetMessage(msg string) {
	m.message = msg
}

func (m *ErrorModalModel) SetOpen(open bool) {
	m.open = open
}

func (m *ErrorModalModel) SetOnClose(f func() tea.Msg) {
	m.onClose = f
}