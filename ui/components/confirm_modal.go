package components

import (
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

type ConfirmModalModel struct {
	title     string
	message   string
	open      bool
	active    int
	focused   bool
	width     int
	height    int
	onConfirm func() tea.Msg
	onCancel  func() tea.Msg
}

func NewConfirmModal(title, message string) ConfirmModalModel {
	return ConfirmModalModel{
		title:   title,
		message: message,
		open:    true,
		active:  0,
		width:   40,
		height:  10,
	}
}

func (m ConfirmModalModel) Init() tea.Cmd {
	return nil
}

func (m ConfirmModalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if !m.open {
			return m, nil
		}
		switch msg.String() {
		case "left", "h", "tab":
			if m.active > 0 {
				m.active--
			}
		case "right", "l", "shift+tab":
			if m.active < 1 {
				m.active++
			}
		case "enter":
			if m.active == 0 && m.onConfirm != nil {
				return m, m.onConfirm
			}
			if m.active == 1 && m.onCancel != nil {
				return m, m.onCancel
			}
		case "q", "esc":
			if m.onCancel != nil {
				return m, m.onCancel
			}
			m.open = false
		}
	}
	return m, nil
}

func (m ConfirmModalModel) View() string {
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

	yesBtn := "[ Yes ]"
	noBtn := "[ No ]"
	if m.active == 0 {
		yesBtn = lipgloss.NewStyle().Background(lipgloss.Color("#0000FF")).Foreground(lipgloss.Color("#FFFFFF")).Render(" Yes ")
	}
	if m.active == 1 {
		noBtn = lipgloss.NewStyle().Background(lipgloss.Color("#0000FF")).Foreground(lipgloss.Color("#FFFFFF")).Render(" No ")
	}

	content := title + "\n\n" + m.message + "\n\n  " + yesBtn + "  " + noBtn

	return lipgloss.Place(m.width+4, m.height+4, lipgloss.Center, lipgloss.Center, box.Render(content))
}

func (m *ConfirmModalModel) SetOpen(open bool) {
	m.open = open
}

func (m *ConfirmModalModel) SetOnConfirm(f func() tea.Msg) {
	m.onConfirm = f
}

func (m *ConfirmModalModel) SetOnCancel(f func() tea.Msg) {
	m.onCancel = f
}