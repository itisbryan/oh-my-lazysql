package components

import (
	"github.com/charmbracelet/lipgloss"
)

type ButtonModel struct {
	Label   string
	Style   lipgloss.Style
	Active  bool
}

func NewButton(label string, style lipgloss.Style) ButtonModel {
	return ButtonModel{
		Label: label,
		Style: style,
	}
}

func (m ButtonModel) View() string {
	if m.Active {
		return m.Style.Bold(true).Render(m.Label)
	}
	return m.Style.Render(m.Label)
}

func (m ButtonModel) SetActive(active bool) ButtonModel {
	m.Active = active
	return m
}