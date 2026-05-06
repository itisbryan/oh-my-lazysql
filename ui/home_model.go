package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jorgerojas26/lazysql/models"
)

type HomeModel struct {
	connection models.Connection
	width      int
	height     int
	focus      string
}

func NewHomeModel(data any) HomeModel {
	conn, ok := data.(models.Connection)
	if !ok {
		conn = models.Connection{}
	}
	return HomeModel{
		connection: conn,
		focus:      "tree",
	}
}

func (m HomeModel) Init() tea.Cmd {
	return nil
}

func (m HomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m HomeModel) View() string {
	title := TitleStyle.Render("Home - " + m.connection.Name)
	help := HelpStyle.Render("[Q] Quit")

	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("Database tree, SQL editor, and results will go here"),
		"",
		help,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}