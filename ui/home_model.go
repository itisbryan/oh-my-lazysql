package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jorgerojas26/lazysql/drivers"
	"github.com/jorgerojas26/lazysql/helpers/logger"
	"github.com/jorgerojas26/lazysql/models"
)

type HomeModel struct {
	connection models.Connection
	driver     drivers.Driver
	tree       *TreeModel
	editor     *EditorModel
	results    *ResultsModel
	width      int
	height     int
	focus      string
}

func NewHomeModel(data any) *HomeModel {
	conn, ok := data.(models.Connection)
	if !ok {
		conn = models.Connection{}
	}

	var driver drivers.Driver
	switch conn.Provider {
	case "MySQL":
		driver = &drivers.MySQL{}
	case "PostgreSQL":
		driver = &drivers.Postgres{}
	case "SQLite":
		driver = &drivers.SQLite{}
	case "MSSQL":
		driver = &drivers.MSSQL{}
	default:
		driver = &drivers.MySQL{}
	}

	if conn.URL != "" {
		if err := driver.Connect(conn.URL); err != nil {
			logger.Error("Failed to connect", map[string]any{"error": err})
		}
	}

	home := &HomeModel{
		connection: conn,
		driver:     driver,
		tree:       NewTreeModel(),
		editor:     NewEditorModel(),
		results:    NewResultsModel(),
		focus:      "tree",
	}

	home.tree.driver = driver
	home.editor.driver = driver
	home.editor.results = home.results

	return home
}

func (m *HomeModel) Init() tea.Cmd {
	return nil
}

func (m *HomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	var cmd tea.Cmd
	var _ tea.Model
	_, _ = m.tree.Update(msg)
	_, _ = m.editor.Update(msg)
	return m, cmd
}

func (m *HomeModel) View() string {
	treePanel := lipgloss.NewStyle().
		Width(m.width/3 - 1).
		Height(m.height - 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666A7E")).
		Render(m.tree.View())

	editorPanel := lipgloss.NewStyle().
		Width(m.width*2/3 - 1).
		Height(m.height/2 - 1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666A7E")).
		Render(m.editor.View())

	resultsPanel := lipgloss.NewStyle().
		Width(m.width*2/3 - 1).
		Height(m.height/2 - 1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666A7E")).
		Render("Results Table")

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, treePanel, lipgloss.JoinVertical(lipgloss.Left, editorPanel, resultsPanel))

	header := TitleStyle.Render("Home - " + m.connection.Name)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, lipgloss.JoinVertical(lipgloss.Left, header, mainContent))
}