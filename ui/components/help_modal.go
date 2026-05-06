package components

import (
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

type HelpModalModel struct {
	title     string
	items     [][]string
	open      bool
	focused   bool
	width     int
	height    int
}

func NewHelpModal(title string) HelpModalModel {
	return HelpModalModel{
		title:   title,
		items:   [][]string{},
		open:    false,
		width:   60,
		height:  20,
	}
}

func (m HelpModalModel) Init() tea.Cmd {
	return nil
}

func (m HelpModalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if !m.open {
			return m, nil
		}
		switch msg.String() {
		case "q", "esc", "h", "?":
			m.open = false
		}
	}
	return m, nil
}

func (m HelpModalModel) View() string {
	if !m.open {
		return ""
	}

	box := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666A7E")).
		Padding(1, 2).
		Background(lipgloss.Color("#1E1E1E"))

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render(m.title)

	lines := []string{title, ""}
	for _, item := range m.items {
		if len(item) >= 2 {
			key := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render(item[0])
			desc := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(item[1])
			lines = append(lines, "  "+key+"  "+desc)
		}
	}

	lines = append(lines, "", lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("[Q] Close"))

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)

	return lipgloss.Place(m.width+4, m.height+4, lipgloss.Center, lipgloss.Center, box.Render(content))
}

func (m *HelpModalModel) AddItem(key, description string) {
	m.items = append(m.items, []string{key, description})
}

func (m *HelpModalModel) SetOpen(open bool) {
	m.open = open
}