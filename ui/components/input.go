package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jorgerojas26/lazysql/ui"
)

type InputModel struct {
	Label   string
	Input   textinput.Model
	focused bool
}

func NewInput(label string, placeholder string) InputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Width = 30
	return InputModel{
		Label: label,
		Input: ti,
	}
}

func (m InputModel) Focus() InputModel {
	m.focused = true
	m.Input.Focus()
	return m
}

func (m InputModel) Blur() InputModel {
	m.focused = false
	m.Input.Blur()
	return m
}

func (m InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	var cmd tea.Cmd
	m.Input, cmd = m.Input.Update(msg)
	return m, cmd
}

func (m InputModel) View() string {
	label := ui.InputLabelStyle.Render(m.Label + ":")
	return lipgloss.JoinHorizontal(lipgloss.Top, label, " ", m.Input.View())
}

func (m InputModel) Value() string {
	return m.Input.Value()
}

func (m InputModel) SetValue(val string) InputModel {
	m.Input.SetValue(val)
	return m
}