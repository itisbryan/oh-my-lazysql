package components

import "github.com/charmbracelet/lipgloss"

type ToggleModel struct {
	Label   string
	Checked bool
	Focused bool
}

func NewToggle(label string, checked bool) ToggleModel {
	return ToggleModel{
		Label:   label,
		Checked: checked,
	}
}

func (m ToggleModel) Toggle() ToggleModel {
	m.Checked = !m.Checked
	return m
}

func (m ToggleModel) View() string {
	checkbox := "[ ]"
	if m.Checked {
		checkbox = "[x]"
	}
	label := m.Label
	if m.Focused {
		checkbox = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render(checkbox)
	}
	return checkbox + " " + label
}

func (m ToggleModel) Value() bool {
	return m.Checked
}