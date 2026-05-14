package components

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

// ConnectionFormComponent is a styled connection form with a card/container design
type ConnectionFormComponent struct {
	form          *huh.Form
	width         int
	height        int
	provider      string
	name          string
	hostname      string
	port          string
	username      string
	password      string
	database      string
	url           string
	showAdvanced  bool
	titleText     string
	action        string // "new" or "edit"
}

var (
	providerButtonBase = lipgloss.NewStyle().
				Padding(0, 2).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#3B4261"))

	providerButtonSelected = lipgloss.NewStyle().
				Padding(0, 2).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7AA2F7")).
				Foreground(lipgloss.Color("#7AA2F7"))
)

func NewConnectionFormComponent(action string) *ConnectionFormComponent {
	c := &ConnectionFormComponent{
		provider:     "MySQL",
		action:       action,
		showAdvanced: false,
	}
	if action == "edit" {
		c.titleText = "Edit Connection"
	} else {
		c.titleText = "New Connection"
	}
	c.buildForm()
	return c
}

func (c *ConnectionFormComponent) buildForm() {
	providers := []string{"MySQL", "PostgreSQL", "SQLite", "MSSQL"}
	providerOptions := make([]huh.Option[string], len(providers))
	for i, p := range providers {
		providerOptions[i] = huh.NewOption(p, p)
	}

	c.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Connection Name").
				Placeholder("my-database").
				Value(&c.name),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Provider").
				Options(providerOptions...).
				Value(&c.provider),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Host").
				Placeholder("localhost").
				Value(&c.hostname),
			huh.NewInput().
				Title("Port").
				Placeholder("3306").
				Value(&c.port),
			huh.NewInput().
				Title("Username").
				Placeholder("root").
				Value(&c.username),
			huh.NewInput().
				Title("Password").
				Placeholder("••••••••").
				EchoMode(huh.EchoModePassword).
				Value(&c.password),
			huh.NewInput().
				Title("Database").
				Placeholder("mydb").
				Value(&c.database),
		).Title("Connection Details"),
		huh.NewGroup(
			huh.NewInput().
				Title("URL").
				Placeholder("mysql://user:pass@host:port/db").
				Description("Override all fields above if set").
				Value(&c.url),
		).Title("Advanced"),
	).WithTheme(huh.ThemeDracula())
}

func (c *ConnectionFormComponent) Init() tea.Cmd {
	return c.form.Init()
}

func (c *ConnectionFormComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
	}

	model, cmd := c.form.Update(msg)
	c.form = model.(*huh.Form)
	return c, cmd
}

func (c *ConnectionFormComponent) View() string {
	// Build the header with ASCII art style
	header := c.buildHeader()

	// Build provider selector
	providerSection := c.buildProviderSection()

	// Build connection details section
	detailsSection := c.buildDetailsSection()

	// Build advanced section
	advancedSection := c.buildAdvancedSection()

	// Footer help
	helpBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666A7E")).
		MarginTop(1).
		Render("[Enter] Next   [Tab] Skip   [Esc] Cancel")

	// Combine everything
	mainContent := lipgloss.JoinVertical(lipgloss.Center,
		header,
		"",
		providerSection,
		"",
		detailsSection,
		"",
		advancedSection,
		"",
		helpBar,
	)

	// Wrap in card
	cardStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7AA2F7")).
		Padding(1, 3).
		Width(58)

	card := cardStyle.Render(mainContent)

	// Center the card on screen
	return lipgloss.Place(c.width, c.height, lipgloss.Center, lipgloss.Center, card)
}

func (c *ConnectionFormComponent) buildHeader() string {
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#BB9AF7")).
		Bold(true).
		Render(c.titleText)

	decoration := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3B4261")).
		Render("──────────────────────────")

	return lipgloss.JoinHorizontal(lipgloss.Center,
		decoration,
		" ",
		title,
		" ",
		decoration,
	)
}

func (c *ConnectionFormComponent) buildProviderSection() string {
	providers := []string{"MySQL", "PostgreSQL", "SQLite", "MSSQL"}
	providerColors := map[string]lipgloss.Color{
		"MySQL":      lipgloss.Color("#E0AF68"),
		"PostgreSQL": lipgloss.Color("#7DCFFF"),
		"SQLite":     lipgloss.Color("#9ECE6A"),
		"MSSQL":      lipgloss.Color("#BB9AF7"),
	}

	var buttons []string
	for _, p := range providers {
		color := providerColors[p]
		style := providerButtonBase.Copy().Foreground(color)
		if c.provider == p {
			buttons = append(buttons, style.Render("[ "+p+" ]"))
		} else {
			buttons = append(buttons, style.Render(" "+p+" "))
		}
	}

	label := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9ECE6A")).
		Render("Provider:")

	providerRow := lipgloss.JoinHorizontal(lipgloss.Left, buttons...)

	return lipgloss.JoinVertical(lipgloss.Left,
		label,
		" "+providerRow,
	)
}

func (c *ConnectionFormComponent) buildDetailsSection() string {
	label := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7AA2F7")).
		Bold(true).
		Render("── Connection Details ──")

	fields := []string{
		c.formatField("Name", c.name, "my-database"),
		c.formatField("Host", c.hostname, "localhost"),
		c.formatField("Port", c.port, "3306"),
		c.formatField("User", c.username, "root"),
		c.formatField("Pass", "••••••••", "••••••••"),
		c.formatField("DB", c.database, "mydb"),
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		label,
		lipgloss.JoinVertical(lipgloss.Left, fields...),
	)
}

func (c *ConnectionFormComponent) formatField(label, value, placeholder string) string {
	lblStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9ECE6A")).
		Width(10).
		Align(lipgloss.Right)

	fieldStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#C0CAF5")).
		Background(lipgloss.Color("#1A1B26")).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3B4261")).
		Padding(0, 2).
		Width(25)

	lbl := lblStyle.Render(label + ":")
	var displayValue string
	if value == "" {
		displayValue = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B4261")).
			Render(placeholder)
	} else {
		displayValue = fieldStyle.Render(value)
	}

	return lbl + "  " + displayValue
}

func (c *ConnectionFormComponent) buildAdvancedSection() string {
	label := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666A7E")).
		Render("── Advanced ──")

	urlStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#C0CAF5")).
		Background(lipgloss.Color("#1A1B26")).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3B4261")).
		Padding(0, 2).
		Width(35)

	var displayURL string
	if c.url == "" {
		displayURL = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B4261")).
			Render("mysql://user:pass@host:port/db")
	} else {
		displayURL = urlStyle.Render(c.url)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		label,
		"  URL: "+displayURL,
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B4261")).
			MarginTop(1).
			Render("   Override all fields above if set"),
	)
}

// Form methods to expose state
func (c *ConnectionFormComponent) GetForm() *huh.Form {
	return c.form
}

func (c *ConnectionFormComponent) GetConnectionData() (name, provider, hostname, port, username, password, database, url string) {
	return c.name, c.provider, c.hostname, c.port, c.username, c.password, c.database, c.url
}

func (c *ConnectionFormComponent) SetValues(name, provider, hostname, port, username, password, database, url string) {
	c.name = name
	c.provider = provider
	c.hostname = hostname
	c.port = port
	c.username = username
	c.password = password
	c.database = database
	c.url = url
	c.buildForm()
}