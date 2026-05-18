package components

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/itisbryan/oh-my-lazysql/ui"
)

type ModalModel struct {
	Title   string
	Content string
	Width   int
	Height  int
	Open    bool
	Buttons []string
	Active  int
}

func NewModal(title string) ModalModel {
	return ModalModel{
		Title:  title,
		Width:  60,
		Height: 10,
		Open:   true,
	}
}

func (m ModalModel) View() string {
	if !m.Open {
		return ""
	}

	box := lipgloss.NewStyle().
		Width(m.Width).
		Height(m.Height).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666A7E")).
		Padding(1, 2)

	title := ui.TitleStyle.Render(m.Title)
	content := m.Content + "\n\n"

	buttons := ""
	for i, btn := range m.Buttons {
		style := lipgloss.NewStyle().Padding(0, 2)
		if i == m.Active {
			style = style.Background(lipgloss.Color("#0000FF")).Foreground(lipgloss.Color("#FFFFFF"))
		}
		buttons += style.Render(btn) + "  "
	}

	return lipgloss.Place(m.Width+4, m.Height+4,
		lipgloss.Center, lipgloss.Center,
		box.Render(title+"\n\n"+content+buttons))
}