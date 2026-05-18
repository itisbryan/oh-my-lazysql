package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/itisbryan/oh-my-lazysql/app"
	"github.com/itisbryan/oh-my-lazysql/models"
)

type ConnectionListModel struct {
	connections []models.Connection
	cursor      int
	width       int
	height      int
	status      string
	statusErr   bool
}

func NewConnectionListModel() *ConnectionListModel {
	return &ConnectionListModel{
		connections: app.App.Connections(),
	}
}

func (m *ConnectionListModel) Init() tea.Cmd {
	return nil
}

func (m *ConnectionListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.connections)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.connections) > 0 {
				conn := m.connections[m.cursor]
				return m, func() tea.Msg {
					return ScreenChangeMsg{Screen: ScreenHome, Data: conn}
				}
			}
		case "n":
			return m, func() tea.Msg {
				return ScreenChangeMsg{Screen: ScreenConnectionForm, Data: nil}
			}
		case "e":
			if len(m.connections) > 0 {
				conn := m.connections[m.cursor]
				return m, func() tea.Msg {
					return ScreenChangeMsg{Screen: ScreenConnectionForm, Data: conn}
				}
			}
		case "d":
			if len(m.connections) > 0 {
				conns := append(m.connections[:m.cursor], m.connections[m.cursor+1:]...)
				if err := app.App.SaveConnections(conns); err != nil {
					m.status = err.Error()
					m.statusErr = true
				} else {
					m.connections = conns
					if m.cursor >= len(m.connections) {
						m.cursor = len(m.connections) - 1
					}
				}
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *ConnectionListModel) View() string {
	title := m.renderWelcomeTitle()
	subtitle := lipgloss.NewStyle().Foreground(PurpleColor).Render("database console")
	bootMenu := m.renderBootMenu()
	commandBar := m.renderCommandBar()

	statusBar := ""
	if m.status != "" {
		style := StatusStyle
		if m.statusErr {
			style = ErrorStyle
		}
		statusBar = style.Render(m.status)
	}

	spacer := lipgloss.NewStyle().Height(1).Render("")
	content := lipgloss.JoinVertical(lipgloss.Left,
		spacer,
		title,
		"  ",
		subtitle,
		"  ",
		bootMenu,
		"  ",
		commandBar,
		"  ",
		statusBar,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m *ConnectionListModel) renderWelcomeTitle() string {
	name := lipgloss.NewStyle().
		Foreground(PurpleColor).
		Bold(true).
		Render("om-lazysql")
	art := lipgloss.NewStyle().
		Foreground(SecondaryTextColor).
		Bold(true).
		Render(strings.Join([]string{
			`              _                       _ `,
			` ___ _ __ ___| |__ _ ____  _ ___ __ _| |`,
			`/ _ \ '  \___| / _` + "`" + ` |_ / || (_-</ _` + "`" + ` | |`,
			`\___/_|_|_|  |_\__,_/__|\_, /__/\__, |_|`,
			`                        |__/       |_|  `,
		}, "\n"))
	return lipgloss.JoinVertical(lipgloss.Left, name, art)
}

func (m *ConnectionListModel) renderBootMenu() string {
	innerWidth := 66
	if m.width > 0 {
		innerWidth = min(74, max(52, m.width-18))
	}

	divider := lipgloss.NewStyle().
		Foreground(MutedBorderColor).
		Width(innerWidth).
		Render(strings.Repeat("─", innerWidth))

	menuRows := []string{}
	for i, conn := range m.connections {
		name := conn.Name
		if name == "" {
			name = conn.URL
		}

		cursor := " "
		style := lipgloss.NewStyle().
			Foreground(PrimaryTextColor).
			Padding(0, 1).
			Width(innerWidth)
		if i == m.cursor {
			cursor = ">"
			style = style.
				Foreground(SecondaryTextColor).
				Background(SelectionColor).
				Bold(true)
		}

		provider := providerBadge(conn.Provider)
		database := ""
		if conn.DBName != "" {
			database = lipgloss.NewStyle().Foreground(InverseTextColor).Render(" // " + conn.DBName)
		}
		row := fmt.Sprintf("%s %-24s %s%s", cursor, name, provider, database)
		if conn.ReadOnly {
			row += " " + lipgloss.NewStyle().Foreground(TertiaryTextColor).Bold(true).Render("READ")
		}
		menuRows = append(menuRows, style.Render(row))
		if i < len(m.connections)-1 {
			menuRows = append(menuRows, divider)
		}
	}

	if len(menuRows) == 0 {
		emptyStyle := lipgloss.NewStyle().Padding(0, 1).Width(innerWidth)
		menuRows = append(menuRows,
			emptyStyle.Render(lipgloss.NewStyle().Foreground(RedColor).Bold(true).Render("NO CONNECTION PROFILES FOUND")),
			emptyStyle.Render(lipgloss.NewStyle().Foreground(InverseTextColor).Render("Press N to initialize a new profile")),
		)
	}

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).
		Padding(1, 3).
		Render(lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Foreground(TertiaryTextColor).Bold(true).Render("BOOT MENU"),
			lipgloss.NewStyle().Foreground(InverseTextColor).Render("Select profile and press ENTER"),
			"",
			lipgloss.JoinVertical(lipgloss.Left, menuRows...),
		))
}

func (m *ConnectionListModel) renderCommandBar() string {
	return lipgloss.NewStyle().
		Foreground(PrimaryTextColor).
		Background(SurfaceColor).
		Padding(0, 2).
		Render(lipgloss.JoinHorizontal(lipgloss.Left,
			KeyStyle.Render("N"), HelpStyle.Render(":new   "),
			KeyStyle.Render("E"), HelpStyle.Render(":edit   "),
			KeyStyle.Render("D"), HelpStyle.Render(":delete   "),
			KeyStyle.Render("ENTER"), HelpStyle.Render(":connect   "),
			KeyStyle.Render("Q"), HelpStyle.Render(":quit"),
		))
}

func providerBadge(provider string) string {
	color := SecondaryTextColor
	switch provider {
	case "PostgreSQL", "postgres":
		color = CyanColor
	case "MySQL", "mysql":
		color = YellowColor
	case "SQLite", "sqlite3":
		color = GreenColor
	case "MSSQL", "sqlserver":
		color = PurpleColor
	}
	return lipgloss.NewStyle().Foreground(color).Bold(true).Render("[" + provider + "]")
}
