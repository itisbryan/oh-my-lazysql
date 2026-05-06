package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jorgerojas26/lazysql/app"
	"github.com/jorgerojas26/lazysql/models"
)

type ConnectionFormModel struct {
	name         textinput.Model
	providerIdx  int
	hostname     textinput.Model
	port         textinput.Model
	username     textinput.Model
	password     textinput.Model
	database     textinput.Model
	sslEnabled   bool
	showSSL      bool
	sslCert      textinput.Model
	sslKey       textinput.Model
	sslCA        textinput.Model
	readOnly     bool
	showAdvanced bool
	url          textinput.Model

	profiles     ProfileSelectorModel
	action       string
	editConn     *models.Connection

	focusIndex   int
	status       string
	statusErr    bool
	width        int
	height       int
}

var providers = []string{"MySQL", "PostgreSQL", "SQLite", "MSSQL"}

func NewConnectionFormModel(data any) *ConnectionFormModel {
	tiName := textinput.New()
	tiName.Placeholder = "Connection name"
	tiName.Width = 40

	tiHost := textinput.New()
	tiHost.Placeholder = "localhost"
	tiHost.Width = 40

	tiPort := textinput.New()
	tiPort.Placeholder = "3306"
	tiPort.Width = 10

	tiUser := textinput.New()
	tiUser.Placeholder = "root"
	tiUser.Width = 20

	tiPass := textinput.New()
	tiPass.Placeholder = "password"
	tiPass.Width = 20
	tiPass.EchoMode = textinput.EchoPassword

	tiDB := textinput.New()
	tiDB.Placeholder = "database"
	tiDB.Width = 40

	tiCert := textinput.New()
	tiCert.Placeholder = "ssl-cert.pem"
	tiCert.Width = 40

	tiKey := textinput.New()
	tiKey.Placeholder = "ssl-key.pem"
	tiKey.Width = 40

	tiCA := textinput.New()
	tiCA.Placeholder = "ca.pem"
	tiCA.Width = 40

	tiURL := textinput.New()
	tiURL.Placeholder = "mysql://user:pass@host:port/db"
	tiURL.Width = 50

	m := ConnectionFormModel{
		name:         tiName,
		providerIdx:  0,
		hostname:     tiHost,
		port:         tiPort,
		username:     tiUser,
		password:     tiPass,
		database:     tiDB,
		sslEnabled:   false,
		showSSL:      false,
		sslCert:      tiCert,
		sslKey:       tiKey,
		sslCA:        tiCA,
		readOnly:     false,
		showAdvanced: false,
		url:          tiURL,
		profiles:     NewProfileSelector(),
		action:       "new",
		focusIndex:   1,
	}

	m.name.Focus()

	if conn, ok := data.(models.Connection); ok && conn.Name != "" {
		m.action = "edit"
		m.editConn = &conn
		m.name.SetValue(conn.Name)
		m.providerIdx = m.providerIndex(conn.Provider)
		m.hostname.SetValue(conn.Hostname)
		m.port.SetValue(conn.Port)
		m.username.SetValue(conn.Username)
		m.password.SetValue(conn.Password)
		m.database.SetValue(conn.DBName)
		m.sslEnabled = len(conn.Profiles) > 0 && conn.Profiles[0].SSLEnabled
		m.showSSL = m.sslEnabled
		if len(conn.Profiles) > 0 {
			m.profiles.SetProfiles(conn.Profiles)
		}
	}

	return &m
}

func (m *ConnectionFormModel) Init() tea.Cmd {
	return nil
}

func (m ConnectionFormModel) providerIndex(provider string) int {
	for i, p := range providers {
		if p == provider {
			return i
		}
	}
	return 0
}

func (m *ConnectionFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.focusIndex = (m.focusIndex + 1) % 10
			m.updateFocus()
		case "shift+tab":
			m.focusIndex = (m.focusIndex - 1 + 10) % 10
			m.updateFocus()
		case "up":
			if m.focusIndex > 1 {
				m.focusIndex--
				m.updateFocus()
			}
		case "down":
			if m.focusIndex < 8 {
				m.focusIndex++
				m.updateFocus()
			}
		case "left":
			if m.focusIndex == 0 {
				m.providerIdx = (m.providerIdx - 1 + len(providers)) % len(providers)
			}
		case "right":
			if m.focusIndex == 0 {
				m.providerIdx = (m.providerIdx + 1) % len(providers)
			}
		case "ctrl+s":
			return m, m.saveConnection
		case "esc", "q":
			return m, func() tea.Msg {
				return ScreenChangeMsg{Screen: ScreenConnectionList, Data: nil}
			}
		case "enter":
			if m.focusIndex == 9 {
				return m, m.saveConnection
			}
		}
	}

	oldFocusedField := m.focusIndexToField()
	m.name, cmd = m.name.Update(msg)
	m.hostname, cmd = m.hostname.Update(msg)
	m.port, cmd = m.port.Update(msg)
	m.username, cmd = m.username.Update(msg)
	m.password, cmd = m.password.Update(msg)
	m.database, cmd = m.database.Update(msg)
	m.sslCert, cmd = m.sslCert.Update(msg)
	m.sslKey, cmd = m.sslKey.Update(msg)
	m.sslCA, cmd = m.sslCA.Update(msg)
	m.url, cmd = m.url.Update(msg)

	if oldFocusedField != m.focusIndexToField() {
		m.updateFocus()
	}

	return m, cmd
}

func (m *ConnectionFormModel) focusIndexToField() int {
	switch m.focusIndex {
	case 1:
		return 1
	case 2:
		return 2
	case 3:
		return 3
	case 4:
		return 4
	case 5:
		return 5
	case 6:
		return 6
	case 7:
		return 7
	case 8:
		return 8
	default:
		return 0
	}
}

func (m *ConnectionFormModel) updateFocus() {
	m.name.Blur()
	m.hostname.Blur()
	m.port.Blur()
	m.username.Blur()
	m.password.Blur()
	m.database.Blur()
	m.sslCert.Blur()
	m.sslKey.Blur()
	m.sslCA.Blur()

	switch m.focusIndex {
	case 1:
		m.name.Focus()
	case 2:
		m.hostname.Focus()
	case 3:
		m.port.Focus()
	case 4:
		m.username.Focus()
	case 5:
		m.password.Focus()
	case 6:
		m.database.Focus()
	case 7:
		m.sslCert.Focus()
	case 8:
		m.sslKey.Focus()
	}
}

func (m ConnectionFormModel) saveConnection() tea.Msg {
	conn := models.Connection{
		Name:     m.name.Value(),
		Provider: providers[m.providerIdx],
		Hostname: m.hostname.Value(),
		Port:     m.port.Value(),
		Username: m.username.Value(),
		Password: m.password.Value(),
		DBName:   m.database.Value(),
		ReadOnly: m.readOnly,
		Profiles: m.profiles.profiles,
	}

	conn.URL = app.BuildConnectionURL(&conn)

	if conn.Name == "" {
		conn.Name = conn.Hostname
	}

	connections := append(app.App.Connections(), conn)
	if err := app.App.SaveConnections(connections); err != nil {
		m.status = err.Error()
		m.statusErr = true
	}

	return ScreenChangeMsg{Screen: ScreenConnectionList, Data: nil}
}

func (m *ConnectionFormModel) View() string {
	providerButtons := make([]string, len(providers))
	for i, p := range providers {
		style := ProviderButtonStyle
		if i == m.providerIdx {
			style = ProviderSelectedStyle
		}
		providerButtons[i] = style.Render(p)
	}

	nameLabel := InputLabelStyle.Render("Name:")
	nameField := m.name.View()

	hostLabel := InputLabelStyle.Render("Host:")
	hostField := m.hostname.View()

	portLabel := InputLabelStyle.Render("Port:")
	portField := m.port.View()

	userLabel := InputLabelStyle.Render("User:")
	userField := m.username.View()

	passLabel := InputLabelStyle.Render("Pass:")
	passField := m.password.View()

	dbLabel := InputLabelStyle.Render("Database:")
	dbField := m.database.View()

	sslToggle := "[ ] SSL"
	if m.sslEnabled {
		sslToggle = "[x] SSL"
	}
	if m.focusIndex == 7 {
		sslToggle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render(sslToggle)
	}

	titleText := "New Connection"
	if m.action == "edit" {
		titleText = "Edit Connection"
	}

	form := lipgloss.JoinVertical(lipgloss.Left,
		TitleStyle.Render(titleText),
		"",
		lipgloss.JoinHorizontal(lipgloss.Top, providerButtons...),
		"",
		lipgloss.JoinHorizontal(lipgloss.Top, nameLabel, " ", nameField),
		lipgloss.JoinHorizontal(lipgloss.Top, hostLabel, " ", hostField),
		lipgloss.JoinHorizontal(lipgloss.Top, portLabel, " ", portField),
		lipgloss.JoinHorizontal(lipgloss.Top, userLabel, " ", userField),
		lipgloss.JoinHorizontal(lipgloss.Top, passLabel, " ", passField),
		lipgloss.JoinHorizontal(lipgloss.Top, dbLabel, " ", dbField),
		"",
		sslToggle,
	)

	help := HelpStyle.Render("[Tab] Next  [Enter] Save  [Q] Cancel")

	statusBar := ""
	if m.status != "" {
		style := StatusStyle
		if m.statusErr {
			style = ErrorStyle
		}
		statusBar = style.Render(m.status)
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		form,
		"",
		help,
		statusBar,
	)

	box := BorderStyle.Width(m.width - 4).Height(m.height - 4).
		Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}