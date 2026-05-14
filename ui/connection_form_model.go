package ui

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/charmbracelet/huh"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jorgerojas26/lazysql/app"
	"github.com/jorgerojas26/lazysql/drivers"
	"github.com/jorgerojas26/lazysql/helpers/logger"
	"github.com/jorgerojas26/lazysql/models"
)

var providers = []string{"MySQL", "PostgreSQL", "SQLite", "MSSQL"}

const (
	connectionModeFields = "Guided fields"
	connectionModeURL    = "Paste URL"
)

func providerIcon(provider string) string {
	// Real database logo glyphs from Nerd Fonts / Devicons.
	// Source verified from devicon SVGs and Nerd Fonts glyphnames.json.
	switch normalizeProvider(provider) {
	case "MySQL":
		return "\ue704"
	case "PostgreSQL":
		return "\ue76e"
	case "SQLite":
		return "\ue7c4"
	case "MSSQL":
		return "\ue82e"
	default:
		return "●"
	}
}

func providerOptionLabel(provider string) string {
	return fmt.Sprintf("%s  %s", providerIcon(provider), provider)
}

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

func defaultPortForProvider(provider string) string {
	switch normalizeProvider(provider) {
	case "PostgreSQL":
		return "5432"
	case "MSSQL":
		return "1433"
	case "SQLite":
		return ""
	default:
		return "3306"
	}
}

func hostPlaceholderForProvider(provider string) string {
	if normalizeProvider(provider) == "SQLite" {
		return "leave empty for SQLite"
	}
	return "localhost or db.example.com"
}

func userPlaceholderForProvider(provider string) string {
	if normalizeProvider(provider) == "SQLite" {
		return "not needed"
	}
	return "root, postgres, or app_user"
}

func databasePlaceholderForProvider(provider string) string {
	if normalizeProvider(provider) == "SQLite" {
		return "~/data/app.db"
	}
	return "app_development"
}

func urlPlaceholderForProvider(provider string) string {
	switch normalizeProvider(provider) {
	case "PostgreSQL":
		return "postgres://user:pass@localhost:5432/app_development?sslmode=disable"
	case "SQLite":
		return "sqlite3:~/data/app.db"
	case "MSSQL":
		return "sqlserver://user:pass@localhost:1433?database=app_development"
	default:
		return "mysql://user:pass@localhost:3306/app_development"
	}
}

// parseURL extracts connection fields from a URL string
func parseURL(urlStr string) (provider, hostname, port, username, password, database string) {
	if urlStr == "" {
		return
	}

	// Parse the URL
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return
	}

	// Determine provider from scheme
	switch parsed.Scheme {
	case "mysql":
		provider = "MySQL"
	case "postgres", "postgresql":
		provider = "PostgreSQL"
	case "sqlite", "sqlite3":
		provider = "SQLite"
	case "sqlserver", "mssql":
		provider = "MSSQL"
	default:
		provider = strings.Title(parsed.Scheme)
	}

	// Extract host and port
	hostname = parsed.Hostname()
	port = parsed.Port()

	// Extract username and password
	username = parsed.User.Username()
	if parsed.User != nil {
		if pwd, ok := parsed.User.Password(); ok {
			password = pwd
		}
	}

	// Extract database (path without leading slash). SQLite URLs often keep
	// the file path in Opaque, and SQL Server URLs often use ?database=.
	database = strings.TrimPrefix(parsed.Path, "/")
	database = strings.TrimSuffix(database, "/")
	if database == "" && parsed.Opaque != "" {
		database = parsed.Opaque
	}
	if database == "" {
		database = parsed.Query().Get("database")
	}

	return
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
	form           *huh.Form
	provider       string
	name           string
	hostname       string
	port           string
	username       string
	password       string
	database       string
	url            string
	connectionMode string
	width          int
	height         int
	status         string
	statusErr      bool
	editConn       *models.Connection
	action         string
	formState      int
	testErr        error
	saved          bool
	hasURL         bool // tracks if URL was provided
	lastParsedURL  string
	lastProvider   string
}

func NewConnectionFormModel(data any) *ConnectionFormModel {
	m := &ConnectionFormModel{
		provider:       "MySQL",
		port:           defaultPortForProvider("MySQL"),
		connectionMode: connectionModeFields,
		action:         "new",
		formState:      formStateEditing,
		lastProvider:   "MySQL",
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
		m.hasURL = conn.URL != ""
		m.lastParsedURL = conn.URL
		m.lastProvider = m.provider
	}

	m.buildForm()

	return m
}

func (m *ConnectionFormModel) buildForm() {
	titleText := "Let's add a database"
	if m.action == "edit" {
		titleText = "Update this connection"
	}

	providerOptions := make([]huh.Option[string], len(providers))
	for i, p := range providers {
		providerOptions[i] = huh.NewOption(providerOptionLabel(p), p)
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Connection setup").
				Description("Use guided fields, or paste a full connection URL instead.").
				Options(
					huh.NewOption("Guided fields (recommended)", connectionModeFields),
					huh.NewOption("Paste connection URL", connectionModeURL),
				).
				Value(&m.connectionMode),
			huh.NewSelect[string]().
				Title("Database type").
				Description("Pick the engine LazySQL should speak to. We'll pre-fill sensible defaults.").
				Options(providerOptions...).
				Value(&m.provider),
		).Title(titleText),
		huh.NewGroup(
			huh.NewInput().
				Title("Connection nickname").
				Placeholder("Local app database").
				Description("Shown in the connection list. Use something you'll recognize later.").
				Value(&m.name),
			huh.NewInput().
				Title("Host").
				Placeholder(hostPlaceholderForProvider(m.provider)).
				Description("Server address. For local Docker databases this is often localhost.").
				Value(&m.hostname),
			huh.NewInput().
				Title("Port").
				Placeholder(defaultPortForProvider(m.provider)).
				Description("Leave the default unless your database listens somewhere else.").
				Value(&m.port),
			huh.NewInput().
				Title("Username").
				Placeholder(userPlaceholderForProvider(m.provider)).
				Value(&m.username),
			huh.NewInput().
				Title("Password").
				Placeholder("stored locally in your LazySQL config").
				Description("Hidden while typing.").
				EchoMode(huh.EchoModePassword).
				Value(&m.password),
			huh.NewInput().
				Title("Database").
				Placeholder(databasePlaceholderForProvider(m.provider)).
				Description("For SQLite, enter the database file path.").
				Value(&m.database),
		).Title("Connection details").
			WithHideFunc(func() bool { return m.connectionMode != connectionModeFields }),
		huh.NewGroup(
			huh.NewInput().
				Title("Connection URL").
				Placeholder(urlPlaceholderForProvider(m.provider)).
				Description("Paste a full URL. LazySQL will parse and save the matching fields.").
				Value(&m.url),
		).Title("Paste a connection URL").
			WithHideFunc(func() bool { return m.connectionMode != connectionModeURL }),
	).WithTheme(huh.ThemeDracula())
}

func (m *ConnectionFormModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m *ConnectionFormModel) syncProviderDefaults() {
	if m.provider == m.lastProvider {
		return
	}

	previousDefaultPort := defaultPortForProvider(m.lastProvider)
	newDefaultPort := defaultPortForProvider(m.provider)
	if m.port == "" || m.port == previousDefaultPort {
		m.port = newDefaultPort
	}
	m.lastProvider = m.provider
}

func (m *ConnectionFormModel) syncURLFields() {
	urlValue := strings.TrimSpace(m.url)
	if urlValue == "" {
		m.hasURL = false
		m.lastParsedURL = ""
		return
	}
	if urlValue == m.lastParsedURL {
		return
	}

	prov, host, port, user, pass, db := parseURL(urlValue)
	if prov != "" {
		m.provider = normalizeProvider(prov)
	}
	if host != "" {
		m.hostname = host
	}
	if port != "" {
		m.port = port
	}
	if user != "" {
		m.username = user
	}
	if pass != "" {
		m.password = pass
	}
	if db != "" {
		m.database = db
	}
	if m.name == "" {
		m.name = suggestedConnectionName(m.provider, host, db)
	}

	m.hasURL = true
	m.lastParsedURL = urlValue
}

func suggestedConnectionName(provider, host, database string) string {
	if database != "" {
		return database
	}
	if host != "" {
		return fmt.Sprintf("%s on %s", normalizeProvider(provider), host)
	}
	return ""
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
			m.enrichConnectionFromURL(&conn)

			logger.Info("Saving connection", map[string]any{
				"name":     conn.Name,
				"url":      conn.URL,
				"dbName":   conn.DBName,
				"provider": conn.Provider,
			})

			if err := m.saveConnection(conn); err != nil {
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
	if m.connectionMode == connectionModeURL {
		m.syncURLFields()
	}
	m.syncProviderDefaults()

	if m.form.State == huh.StateCompleted {
		conn := m.buildConnection()
		// Only build URL if not already set (preserve custom URLs)
		if conn.URL == "" {
			conn.URL = app.BuildConnectionURL(&conn)
		}
		if conn.Name == "" {
			conn.Name = suggestedConnectionName(conn.Provider, conn.Hostname, conn.DBName)
		}

		m.formState = formStateTesting
		m.status = fmt.Sprintf("Testing %s...", conn.Name)
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
		urlStr := conn.URL

		// Only build URL if not already set
		if urlStr == "" {
			urlStr = app.BuildConnectionURL(&conn)
		}

		// If DBName is empty but URL has a database path, extract it
		if conn.DBName == "" && urlStr != "" {
			parsedURL, err := url.Parse(urlStr)
			if err == nil {
				dbFromURL := strings.TrimPrefix(parsedURL.Path, "/")
				dbFromURL = strings.TrimSuffix(dbFromURL, "/")
				if dbFromURL != "" {
					conn.DBName = dbFromURL
				}
			}
		}

		logger.Info("TestConnection", map[string]any{"provider": conn.Provider, "url": urlStr, "dbName": conn.DBName})
		err := driver.TestConnection(urlStr)
		if err != nil {
			logger.Error("TestConnection failed", map[string]any{"error": err})
		} else {
			logger.Info("TestConnection succeeded", nil)
		}
		return connectionTestResult{err: err}
	}
}

func (m *ConnectionFormModel) buildConnection() models.Connection {
	urlValue := ""
	if m.connectionMode == connectionModeURL {
		urlValue = strings.TrimSpace(m.url)
	}

	conn := models.Connection{
		Name:     strings.TrimSpace(m.name),
		Provider: normalizeProvider(m.provider),
		Hostname: strings.TrimSpace(m.hostname),
		Port:     strings.TrimSpace(m.port),
		Username: strings.TrimSpace(m.username),
		Password: m.password,
		DBName:   strings.TrimSpace(m.database),
		URL:      urlValue,
	}
	if conn.Name == "" {
		conn.Name = suggestedConnectionName(conn.Provider, conn.Hostname, conn.DBName)
	}
	if conn.Port == "" {
		conn.Port = defaultPortForProvider(conn.Provider)
	}
	if m.editConn != nil {
		conn.ReadOnly = m.editConn.ReadOnly
		conn.Profiles = m.editConn.Profiles
	}
	return conn
}

func (m *ConnectionFormModel) enrichConnectionFromURL(conn *models.Connection) {
	if conn.URL == "" {
		return
	}

	prov, host, port, user, pass, db := parseURL(conn.URL)
	if prov != "" {
		conn.Provider = normalizeProvider(prov)
	}
	if host != "" {
		conn.Hostname = host
	}
	if port != "" {
		conn.Port = port
	}
	if user != "" {
		conn.Username = user
	}
	if pass != "" {
		conn.Password = pass
	}
	if db != "" {
		conn.DBName = db
	}
	if conn.Name == "" {
		conn.Name = suggestedConnectionName(conn.Provider, conn.Hostname, conn.DBName)
	}
}

func (m *ConnectionFormModel) saveConnection(conn models.Connection) error {
	connections := app.App.Connections()
	if m.action == "edit" && m.editConn != nil {
		for i := range connections {
			if connections[i].Name == m.editConn.Name {
				connections[i] = conn
				return app.App.SaveConnections(connections)
			}
		}
	}

	for i := range connections {
		if connections[i].Name == conn.Name {
			connections[i] = conn
			return app.App.SaveConnections(connections)
		}
	}

	connections = append(connections, conn)
	return app.App.SaveConnections(connections)
}

// formTitle returns the styled title for the form
func (m *ConnectionFormModel) formTitle() string {
	title := "Add a friendly connection"
	if m.action == "edit" {
		title = "Update connection"
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#BB9AF7")).
		Bold(true).
		Render(title)
}

// providerBadge returns a styled badge showing current provider
func (m *ConnectionFormModel) providerBadge() string {
	color := SecondaryTextColor
	switch m.provider {
	case "PostgreSQL", "postgres":
		color = lipgloss.Color("#7DCFFF")
	case "MySQL", "mysql":
		color = lipgloss.Color("#E0AF68")
	case "SQLite", "sqlite3":
		color = lipgloss.Color("#9ECE6A")
	case "MSSQL", "sqlserver":
		color = lipgloss.Color("#BB9AF7")
	}
	return lipgloss.NewStyle().Foreground(color).Bold(true).Render(providerOptionLabel(m.provider))
}

func (m *ConnectionFormModel) formWidth() int {
	width := 76
	if m.width > 0 {
		width = min(width, max(42, m.width-26))
	}
	return width
}

func (m *ConnectionFormModel) formHeight() int {
	height := 20
	if m.height > 0 {
		height = min(height, max(12, m.height-14))
	}
	return height
}

func (m *ConnectionFormModel) cardWidth(formWidth int) int {
	width := min(96, formWidth+18)
	if m.width > 0 {
		width = min(width, max(50, m.width-10))
	}
	return width
}

func (m *ConnectionFormModel) cardHeight() int {
	height := 30
	if m.height > 0 {
		height = min(height, max(18, m.height-6))
	}
	return height
}

func (m *ConnectionFormModel) View() string {
	// Error state - styled error display
	if m.formState == formStateTestDone && m.testErr != nil {
		errBox := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ErrorColor).
			Padding(2, 4).
			Width(m.width - 20)

		errContent := lipgloss.JoinVertical(lipgloss.Center,
			ErrorStyle.Bold(true).Render("✗ We couldn't connect yet"),
			"",
			ErrorStyle.Render(m.testErr.Error()),
			"",
			HelpStyle.Render("Check host, port, credentials, or paste a full URL in Advanced."),
			HelpStyle.Render("[R] Edit details   [Q] Cancel"),
		)

		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			errBox.Render(errContent))
	}

	// Testing state - styled loading display
	if m.formState == formStateTesting {
		loadingBox := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(TertiaryTextColor).
			Padding(2, 4).
			Width(40)

		loadingContent := lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.NewStyle().
				Foreground(TertiaryTextColor).
				Bold(true).
				Render("⟳ Checking this connection..."),
			HelpStyle.Render("LazySQL will save it after the test succeeds."),
		)

		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			loadingBox.Render(loadingContent))
	}

	formWidth := m.formWidth()
	m.form.WithWidth(formWidth).WithHeight(m.formHeight())

	// Get the raw form view from huh after constraining it so it can be centered
	formView := m.form.View()

	// Build header with title and provider badge
	header := lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B4261")).
			Render("────────────────"),
		" ",
		m.formTitle(),
		" ",
		lipgloss.NewStyle().
			Foreground(TertiaryTextColor).
			Render("["+m.providerBadge()+"]"),
		" ",
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B4261")).
			Render("────────────────"),
	)

	// Status bar
	statusBar := ""
	if m.status != "" {
		style := StatusStyle
		if m.statusErr {
			style = ErrorStyle
		}
		statusBar = style.Render(m.status)
	}

	// Help footer
	helpBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#565F89")).
		Render("[↑↓] Move   [Enter] Continue/Test   [Esc] Back")

	// Combine everything with proper spacing
	content := lipgloss.JoinVertical(lipgloss.Center,
		header,
		"",
		formView,
		"",
		statusBar,
		helpBar,
	)

	// Wrap in a centered card with border
	cardStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7AA2F7")).
		Padding(2, 5).
		Width(m.cardWidth(formWidth)).
		Height(m.cardHeight())

	card := cardStyle.Render(content)

	// Place in center of screen
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		card)
}
