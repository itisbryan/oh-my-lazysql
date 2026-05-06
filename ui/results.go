package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ResultsModel struct {
	table     table.Model
	columns   []table.Column
	rows      []table.Row
	width     int
	height    int
	status    string
	page      int
	pageSize  int
	totalRows int
}

func NewResultsModel() *ResultsModel {
	cols := []table.Column{
		{Title: "Column 1", Width: 20},
		{Title: "Column 2", Width: 20},
		{Title: "Column 3", Width: 20},
	}

	t := table.New()
	t.SetColumns(cols)
	t.SetRows([]table.Row{
		{"Row 1 Col 1", "Row 1 Col 2", "Row 1 Col 3"},
		{"Row 2 Col 1", "Row 2 Col 2", "Row 2 Col 3"},
	})

	return &ResultsModel{
		columns:  cols,
		pageSize: 100,
	}
}

func (m *ResultsModel) Init() tea.Cmd {
	return nil
}

func (m *ResultsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.table.MoveUp(1)
		case "down", "j":
			m.table.MoveDown(1)
		case "pgup":
			m.table.MoveUp(10)
		case "pgdown":
			m.table.MoveDown(10)
		case "ctrl+p":
			if m.page > 0 {
				m.page--
			}
		case "ctrl+n":
			m.page++
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *ResultsModel) View() string {
	pagination := ""
	if m.totalRows > 0 {
		pagination = fmt.Sprintf("Page %d | Rows %d-%d of %d",
			m.page+1,
			m.page*m.pageSize+1,
			min((m.page+1)*m.pageSize, m.totalRows),
			m.totalRows)
	}

	statusBar := ""
	if m.status != "" {
		statusBar = StatusStyle.Render(m.status)
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		TitleStyle.Render("Results"),
		m.table.View(),
		HelpStyle.Render("[j/k] Navigate  [pgup/pgdn] Page  [ctrl+p/n] Prev/Next Page"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(pagination),
		statusBar,
	)

	box := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666A7E")).
		Width(m.width - 2).
		Height(m.height - 2)

	return box.Render(content)
}

func (m *ResultsModel) SetData(columns []string, rows [][]string) {
	tableCols := make([]table.Column, len(columns))
	for i, col := range columns {
		tableCols[i] = table.Column{
			Title: col,
			Width: 20,
		}
	}
	m.columns = tableCols
	m.rows = make([]table.Row, len(rows))
	for i, row := range rows {
		m.rows[i] = row
	}
	m.table.SetColumns(m.columns)
	m.table.SetRows(m.rows)
	m.totalRows = len(rows)
}

func (m *ResultsModel) SetStatus(msg string) {
	m.status = msg
}

func (m *ResultsModel) Clear() {
	m.rows = []table.Row{}
	m.table.SetRows(m.rows)
	m.totalRows = 0
	m.status = ""
}