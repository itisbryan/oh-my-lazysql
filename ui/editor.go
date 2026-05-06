package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jorgerojas26/lazysql/drivers"
)

type EditorModel struct {
	textarea   textarea.Model
	driver     drivers.Driver
	results    *ResultsModel
	width      int
	height     int
	focused    bool
	executing  bool
}

func NewEditorModel() *EditorModel {
	ta := textarea.New()
	ta.Placeholder = "Enter SQL query..."
	ta.SetWidth(80)
	ta.SetHeight(10)
	ta.ShowLineNumbers = true
	return &EditorModel{
		textarea: ta,
		focused:  true,
	}
}

func (m *EditorModel) SetDriver(driver drivers.Driver) {
	m.driver = driver
}

func (m *EditorModel) SetResults(results *ResultsModel) {
	m.results = results
}

func (m *EditorModel) Init() tea.Cmd {
	return nil
}

func (m *EditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+e", "enter":
			if m.textarea.Focused() && !m.executing {
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

func (m *EditorModel) executeQuery() tea.Msg {
	if m.driver == nil {
		return nil
	}
	sql := m.textarea.Value()
	if sql == "" {
		return nil
	}

	m.executing = true
	results, rowCount, err := m.driver.ExecuteQuery(sql)
	m.executing = false

	if err != nil {
		if m.results != nil {
			m.results.SetStatus("Error: " + err.Error())
		}
		return nil
	}

	if m.results != nil && len(results) > 0 {
		columns := results[0]
		rows := results[1:]
		m.results.SetData(columns, rows)
		m.results.SetStatus(fmt.Sprintf("Got %d rows", rowCount))
	}
	return nil
}

func (m *EditorModel) View() string {
	executeHint := "[Ctrl+E] Execute  [Esc] Blur"
	if m.executing {
		executeHint = "Executing..."
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		TitleStyle.Render("SQL Editor"),
		m.textarea.View(),
		HelpStyle.Render(executeHint),
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