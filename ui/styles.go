package ui

import "github.com/charmbracelet/lipgloss"

var (
	BaseStyle = lipgloss.NewStyle().
			Padding(0, 1)

	BorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#666A7E"))

	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	HighlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00"))

	SelectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#0000FF")).
			Foreground(lipgloss.Color("#FFFFFF"))

	StatusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00"))

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	InputLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
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
				Background(lipgloss.Color("#0000FF")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))
)