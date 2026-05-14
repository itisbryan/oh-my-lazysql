package ui

import "github.com/charmbracelet/lipgloss"

var (
	BaseStyle = lipgloss.NewStyle().
			Padding(0, 1)

	BorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#666A7E"))

	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7AA2F7")).
			Bold(true)

	HighlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7AA2F7"))

	SelectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#283457")).
			Foreground(lipgloss.Color("#C0CAF5"))

	StatusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00"))

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	ErrorColor = lipgloss.Color("#FF0000")

	InputLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Width(10).
			Align(lipgloss.Right)

	InputFieldStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#FFFFFF")).
			Padding(0, 1)

	ProfileActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")).
				Bold(true)

	ProfileInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888"))

	ProviderButtonStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Background(lipgloss.Color("#FFFFFF")).
				Foreground(lipgloss.Color("#000000"))

	ProviderSelectedStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Background(lipgloss.Color("#283457")).
				Foreground(lipgloss.Color("#C0CAF5")).
				Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	KeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7AA2F7"))

	SidebarTitleColor = lipgloss.Color("#666A7E")

	PrimaryTextColor   = lipgloss.Color("#C0CAF5")
	SecondaryTextColor = lipgloss.Color("#7AA2F7")
	TertiaryTextColor  = lipgloss.Color("#9ECE6A")
	InverseTextColor   = lipgloss.Color("#888888")

	// Connection Form Styles
	formCardStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7AA2F7")).
			Padding(1, 3).
			Width(60)

	formTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BB9AF7")).
			Bold(true).
			MarginBottom(1)

	formSectionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7AA2F7")).
			MarginTop(1)

	formDecoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B4261"))
)
