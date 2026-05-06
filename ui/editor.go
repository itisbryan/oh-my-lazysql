package ui

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type EditorModel struct {
	textarea textarea.Model
	width    int
	height   int
	focused  bool
}

func NewEditorModel() EditorModel {
	ta := textarea.New()
	ta.Placeholder = "Enter SQL query..."
	ta.SetWidth(80)
	ta.SetHeight(10)
	ta.ShowLineNumbers = true
	return EditorModel{
		textarea: ta,
		focused:  true,
	}
}

func (m EditorModel) Init() tea.Cmd {
	return nil
}

func (m EditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+e", "enter":
			if m.textarea.Focused() {
				return m, m.executeQuery
			}
		case "esc":
			if m.textarea.Focused() {
				m.textarea.Blur()
			}
		}
	}
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m EditorModel) executeQuery() tea.Msg {
	return nil
}

func (m EditorModel) View() string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		TitleStyle.Render("SQL Editor"),
		m.textarea.View(),
		HelpStyle.Render("[Ctrl+E] Execute  [Esc] Blur"),
	)

	box := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666A7E")).
		Width(m.width - 2).
		Height(m.height - 2)

	return box.Render(content)
}

func (m EditorModel) Value() string {
	return m.textarea.Value()
}

func (m EditorModel) SetValue(sql string) EditorModel {
	m.textarea.SetValue(sql)
	return m
}