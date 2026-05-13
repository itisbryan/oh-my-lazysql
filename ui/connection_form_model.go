package ui

import (
	"fmt"

	"github.com/charmbracelet/huh"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jorgerojas26/lazysql/app"
	"github.com/jorgerojas26/lazysql/drivers"
	"github.com/jorgerojas26/lazysql/helpers/logger"
	"github.com/jorgerojas26/lazysql/models"
)

var providers = []string{"MySQL", "PostgreSQL", "SQLite", "MSSQL"}

func normalizeProvider(p string) string {
	for _, name := range providers {
		if p == name {
			return p
		}
	}
	switch p {
	case "mysql":
		return "MySQL"
	case "postgres":
		return "PostgreSQL"
	case "sqlite3":
		return "SQLite"
	case "sqlserver":
		return "MSSQL"
	default:
		return "MySQL"
	}
}

func driverForProvider(provider string) drivers.Driver {
	switch provider {
	case "MySQL", "mysql":
		return &drivers.MySQL{}
	case "PostgreSQL", "postgres":
		return &drivers.Postgres{}
	case "SQLite", "sqlite3":
		return &drivers.SQLite{}
	case "MSSQL", "sqlserver":
		return &drivers.MSSQL{}
	default:
		return &drivers.MySQL{}
	}
}

const (
	formStateEditing = iota
	formStateTesting
	formStateTestDone
)

type connectionTestResult struct {
	err error
}

type ConnectionFormModel struct {
	form       *huh.Form
	provider   string
	name       string
	hostname   string
	port       string
	username   string
	password   string
	database   string
	url        string
	width      int
	height     int
	status     string
	statusErr  bool
	editConn   *models.Connection
	action     string
	formState  int
	testErr    error
	saved      bool
}

func NewConnectionFormModel(data any) *ConnectionFormModel {
	m := &ConnectionFormModel{
		provider:  "MySQL",
		action:    "new",
		formState: formStateEditing,
	}

	if conn, ok := data.(models.Connection); ok && conn.Name != "" {
		m.action = "edit"
		m.editConn = &conn
		m.name = conn.Name
		m.provider = normalizeProvider(conn.Provider)
		m.hostname = conn.Hostname
		m.port = conn.Port
		m.username = conn.Username
		m.password = conn.Password
		m.database = conn.DBName
		m.url = conn.URL
	}

	m.buildForm()

	return m
}

func (m *ConnectionFormModel) buildForm() {
	titleText := "New Connection"
	if m.action == "edit" {
		titleText = "Edit Connection"
	}

	providerOptions := make([]huh.Option[string], len(providers))
	for i, p := range providers {
		providerOptions[i] = huh.NewOption(p, p)
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(titleText).
				Description("Provider").
				Options(providerOptions...).
				Value(&m.provider),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Name").
				Placeholder("Connection name").
				Value(&m.name),
			huh.NewInput().
				Title("Host").
				Placeholder("localhost").
				Value(&m.hostname),
			huh.NewInput().
				Title("Port").
				Placeholder("3306").
				Value(&m.port),
			huh.NewInput().
				Title("User").
				Placeholder("root").
				Value(&m.username),
			huh.NewInput().
				Title("Password").
				Placeholder("password").
				EchoMode(huh.EchoModePassword).
				Value(&m.password),
			huh.NewInput().
				Title("Database").
				Placeholder("mydb").
				Value(&m.database),
		).Title("Connection Details"),
		huh.NewGroup(
			huh.NewInput().
				Title("URL").
				Placeholder("mysql://user:pass@host:port/db").
				Description("Override all fields above if set").
				Value(&m.url),
		).Title("Advanced"),
	).WithTheme(huh.ThemeDracula())
}

func (m *ConnectionFormModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m *ConnectionFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case connectionTestResult:
		m.formState = formStateTestDone
		m.testErr = msg.err
		if msg.err != nil {
			m.status = fmt.Sprintf("Connection failed: %v", msg.err)
			m.statusErr = true
		} else {
			m.status = "Connection successful!"
			m.statusErr = false
			conn := m.buildConnection()
			conn.URL = app.BuildConnectionURL(&conn)
			connections := append(app.App.Connections(), conn)
			if err := app.App.SaveConnections(connections); err != nil {
				m.status = fmt.Sprintf("Save failed: %v", err)
				m.statusErr = true
				return m, nil
			}
			return m, func() tea.Msg {
				return ScreenChangeMsg{Screen: ScreenHome, Data: conn}
			}
		}
		return m, nil
	}

	if m.formState == formStateTesting {
		return m, nil
	}

	if m.formState == formStateTestDone {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "r", "R":
				m.formState = formStateEditing
				m.status = ""
				m.buildForm()
				return m, m.form.Init()
			case "q", "esc", "ctrl+c":
				return m, func() tea.Msg {
					return ScreenChangeMsg{Screen: ScreenConnectionList, Data: nil}
				}
			}
		}
		return m, nil
	}

	model, cmd := m.form.Update(msg)
	m.form = model.(*huh.Form)

	if m.form.State == huh.StateCompleted {
		conn := m.buildConnection()
		conn.URL = app.BuildConnectionURL(&conn)
		if conn.Name == "" {
			conn.Name = conn.Hostname
		}

		m.formState = formStateTesting
		m.status = "Testing connection..."
		m.statusErr = false
		return m, m.testConnection(conn)
	}

	if m.form.State == huh.StateAborted {
		return m, func() tea.Msg {
			return ScreenChangeMsg{Screen: ScreenConnectionList, Data: nil}
		}
	}

	return m, cmd
}

func (m *ConnectionFormModel) testConnection(conn models.Connection) tea.Cmd {
	return func() tea.Msg {
		driver := driverForProvider(conn.Provider)
		url := conn.URL
		if url == "" {
			url = app.BuildConnectionURL(&conn)
		}
		logger.Info("TestConnection", map[string]any{"provider": conn.Provider, "url": url})
		err := driver.TestConnection(url)
		if err != nil {
			logger.Error("TestConnection failed", map[string]any{"error": err})
		} else {
			logger.Info("TestConnection succeeded", nil)
		}
		return connectionTestResult{err: err}
	}
}

func (m *ConnectionFormModel) buildConnection() models.Connection {
	conn := models.Connection{
		Name:     m.name,
		Provider: m.provider,
		Hostname: m.hostname,
		Port:     m.port,
		Username: m.username,
		Password: m.password,
		DBName:   m.database,
		URL:      m.url,
	}
	if m.editConn != nil {
		conn.ReadOnly = m.editConn.ReadOnly
		conn.Profiles = m.editConn.Profiles
	}
	return conn
}

func (m *ConnectionFormModel) View() string {
	if m.formState == formStateTestDone && m.testErr != nil {
		errContent := lipgloss.JoinVertical(lipgloss.Left,
			"",
			ErrorStyle.Bold(true).Render("Connection Failed"),
			"",
			ErrorStyle.Render(m.testErr.Error()),
			"",
			HelpStyle.Render("[R] Retry (edit form)   [Q] Cancel"),
			"",
		)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, errContent)
	}

	if m.formState == formStateTesting {
		testingContent := lipgloss.JoinVertical(lipgloss.Left,
			"",
			StatusStyle.Bold(true).Render("Testing connection..."),
			"",
		)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, testingContent)
	}

	formView := m.form.View()

	statusBar := ""
	if m.status != "" {
		style := StatusStyle
		if m.statusErr {
			style = ErrorStyle
		}
		statusBar = style.Render(m.status)
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		formView,
		"",
		statusBar,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}