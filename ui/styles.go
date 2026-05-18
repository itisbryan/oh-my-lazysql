package ui

import (
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type ThemePalette struct {
	Name string

	Background string
	Surface    string
	SurfaceAlt string
	Overlay    string
	Selection  string

	Border      string
	MutedBorder string

	PrimaryText   string
	SecondaryText string
	MutedText     string
	InverseText   string

	Accent string
	Blue   string
	Cyan   string
	Green  string
	Yellow string
	Orange string
	Purple string
	Red    string
}

var themes = map[string]ThemePalette{
	"tokyonight": {
		Name: "tokyonight", Background: "#1A1B26", Surface: "#161B2D", SurfaceAlt: "#151A2B", Overlay: "#1F2335", Selection: "#283457",
		Border: "#7AA2F7", MutedBorder: "#3B4261", PrimaryText: "#C0CAF5", SecondaryText: "#7AA2F7", MutedText: "#565F89", InverseText: "#1A1B26",
		Accent: "#7AA2F7", Blue: "#7AA2F7", Cyan: "#7DCFFF", Green: "#9ECE6A", Yellow: "#E0AF68", Orange: "#FF9E64", Purple: "#BB9AF7", Red: "#F7768E",
	},
	"dracula": {
		Name: "dracula", Background: "#282A36", Surface: "#21222C", SurfaceAlt: "#1E1F29", Overlay: "#343746", Selection: "#44475A",
		Border: "#BD93F9", MutedBorder: "#6272A4", PrimaryText: "#F8F8F2", SecondaryText: "#BD93F9", MutedText: "#6272A4", InverseText: "#282A36",
		Accent: "#BD93F9", Blue: "#8BE9FD", Cyan: "#8BE9FD", Green: "#50FA7B", Yellow: "#F1FA8C", Orange: "#FFB86C", Purple: "#BD93F9", Red: "#FF5555",
	},
	"catppuccin-mocha": {
		Name: "catppuccin-mocha", Background: "#1E1E2E", Surface: "#181825", SurfaceAlt: "#11111B", Overlay: "#313244", Selection: "#45475A",
		Border: "#89B4FA", MutedBorder: "#585B70", PrimaryText: "#CDD6F4", SecondaryText: "#89B4FA", MutedText: "#6C7086", InverseText: "#1E1E2E",
		Accent: "#89B4FA", Blue: "#89B4FA", Cyan: "#74C7EC", Green: "#A6E3A1", Yellow: "#F9E2AF", Orange: "#FAB387", Purple: "#CBA6F7", Red: "#F38BA8",
	},
	"nord": {
		Name: "nord", Background: "#2E3440", Surface: "#252B36", SurfaceAlt: "#242933", Overlay: "#3B4252", Selection: "#434C5E",
		Border: "#88C0D0", MutedBorder: "#4C566A", PrimaryText: "#D8DEE9", SecondaryText: "#88C0D0", MutedText: "#67748A", InverseText: "#2E3440",
		Accent: "#88C0D0", Blue: "#81A1C1", Cyan: "#8FBCBB", Green: "#A3BE8C", Yellow: "#EBCB8B", Orange: "#D08770", Purple: "#B48EAD", Red: "#BF616A",
	},
	"gruvbox-dark": {
		Name: "gruvbox-dark", Background: "#282828", Surface: "#1D2021", SurfaceAlt: "#1B1B1B", Overlay: "#3C3836", Selection: "#504945",
		Border: "#83A598", MutedBorder: "#665C54", PrimaryText: "#EBDBB2", SecondaryText: "#83A598", MutedText: "#928374", InverseText: "#282828",
		Accent: "#83A598", Blue: "#83A598", Cyan: "#8EC07C", Green: "#B8BB26", Yellow: "#FABD2F", Orange: "#FE8019", Purple: "#D3869B", Red: "#FB4934",
	},
	"terminal": {
		Name: "terminal", Background: "", Surface: "", SurfaceAlt: "", Overlay: "", Selection: "7",
		Border: "4", MutedBorder: "8", PrimaryText: "15", SecondaryText: "4", MutedText: "8", InverseText: "0",
		Accent: "4", Blue: "4", Cyan: "6", Green: "2", Yellow: "3", Orange: "3", Purple: "5", Red: "1",
	},
}

var CurrentTheme = themes["tokyonight"]

var (
	BackgroundColor = lipgloss.Color(CurrentTheme.Background)
	SurfaceColor    = lipgloss.Color(CurrentTheme.Surface)
	SurfaceAltColor = lipgloss.Color(CurrentTheme.SurfaceAlt)
	OverlayColor    = lipgloss.Color(CurrentTheme.Overlay)
	SelectionColor  = lipgloss.Color(CurrentTheme.Selection)

	BorderColor      = lipgloss.Color(CurrentTheme.Border)
	MutedBorderColor = lipgloss.Color(CurrentTheme.MutedBorder)

	PrimaryTextColor   = lipgloss.Color(CurrentTheme.PrimaryText)
	SecondaryTextColor = lipgloss.Color(CurrentTheme.SecondaryText)
	TertiaryTextColor  = lipgloss.Color(CurrentTheme.Green)
	MutedTextColor     = lipgloss.Color(CurrentTheme.MutedText)
	InverseTextColor   = lipgloss.Color(CurrentTheme.InverseText)

	AccentColor = lipgloss.Color(CurrentTheme.Accent)
	BlueColor   = lipgloss.Color(CurrentTheme.Blue)
	CyanColor   = lipgloss.Color(CurrentTheme.Cyan)
	GreenColor  = lipgloss.Color(CurrentTheme.Green)
	YellowColor = lipgloss.Color(CurrentTheme.Yellow)
	OrangeColor = lipgloss.Color(CurrentTheme.Orange)
	PurpleColor = lipgloss.Color(CurrentTheme.Purple)
	RedColor    = lipgloss.Color(CurrentTheme.Red)

	ErrorColor        = RedColor
	SidebarTitleColor = MutedBorderColor
)

var (
	BaseStyle             lipgloss.Style
	BorderStyle           lipgloss.Style
	TitleStyle            lipgloss.Style
	HighlightStyle        lipgloss.Style
	SelectedStyle         lipgloss.Style
	StatusStyle           lipgloss.Style
	ErrorStyle            lipgloss.Style
	InputLabelStyle       lipgloss.Style
	InputFieldStyle       lipgloss.Style
	ProfileActiveStyle    lipgloss.Style
	ProfileInactiveStyle  lipgloss.Style
	ProviderButtonStyle   lipgloss.Style
	ProviderSelectedStyle lipgloss.Style
	HelpStyle             lipgloss.Style
	KeyStyle              lipgloss.Style

	// Connection Form Styles
	formCardStyle    lipgloss.Style
	formTitleStyle   lipgloss.Style
	formSectionStyle lipgloss.Style
	formDecoStyle    lipgloss.Style
)

func init() {
	ApplyTheme("tokyonight")
}

func ApplyTheme(name string) string {
	key := normalizeThemeName(name)
	palette, ok := themes[key]
	if !ok {
		palette = themes["tokyonight"]
	}
	CurrentTheme = palette

	BackgroundColor = lipgloss.Color(palette.Background)
	SurfaceColor = lipgloss.Color(palette.Surface)
	SurfaceAltColor = lipgloss.Color(palette.SurfaceAlt)
	OverlayColor = lipgloss.Color(palette.Overlay)
	SelectionColor = lipgloss.Color(palette.Selection)
	BorderColor = lipgloss.Color(palette.Border)
	MutedBorderColor = lipgloss.Color(palette.MutedBorder)
	PrimaryTextColor = lipgloss.Color(palette.PrimaryText)
	SecondaryTextColor = lipgloss.Color(palette.SecondaryText)
	TertiaryTextColor = lipgloss.Color(palette.Green)
	MutedTextColor = lipgloss.Color(palette.MutedText)
	InverseTextColor = lipgloss.Color(palette.InverseText)
	AccentColor = lipgloss.Color(palette.Accent)
	BlueColor = lipgloss.Color(palette.Blue)
	CyanColor = lipgloss.Color(palette.Cyan)
	GreenColor = lipgloss.Color(palette.Green)
	YellowColor = lipgloss.Color(palette.Yellow)
	OrangeColor = lipgloss.Color(palette.Orange)
	PurpleColor = lipgloss.Color(palette.Purple)
	RedColor = lipgloss.Color(palette.Red)
	ErrorColor = RedColor
	SidebarTitleColor = MutedBorderColor

	BaseStyle = lipgloss.NewStyle().Padding(0, 1)
	BorderStyle = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(MutedBorderColor)
	TitleStyle = lipgloss.NewStyle().Foreground(AccentColor).Bold(true)
	HighlightStyle = lipgloss.NewStyle().Foreground(AccentColor)
	SelectedStyle = lipgloss.NewStyle().Background(SelectionColor).Foreground(PrimaryTextColor)
	StatusStyle = lipgloss.NewStyle().Foreground(GreenColor)
	ErrorStyle = lipgloss.NewStyle().Foreground(ErrorColor)
	InputLabelStyle = lipgloss.NewStyle().Foreground(GreenColor).Width(10).Align(lipgloss.Right)
	InputFieldStyle = lipgloss.NewStyle().Foreground(InverseTextColor).Background(PrimaryTextColor).Padding(0, 1)
	ProfileActiveStyle = lipgloss.NewStyle().Foreground(GreenColor).Bold(true)
	ProfileInactiveStyle = lipgloss.NewStyle().Foreground(MutedTextColor)
	ProviderButtonStyle = lipgloss.NewStyle().Padding(0, 1).Background(PrimaryTextColor).Foreground(InverseTextColor)
	ProviderSelectedStyle = lipgloss.NewStyle().Padding(0, 1).Background(SelectionColor).Foreground(PrimaryTextColor).Bold(true)
	HelpStyle = lipgloss.NewStyle().Foreground(MutedTextColor)
	KeyStyle = lipgloss.NewStyle().Foreground(AccentColor)

	formCardStyle = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(BorderColor).Padding(1, 3).Width(60)
	formTitleStyle = lipgloss.NewStyle().Foreground(PurpleColor).Bold(true).MarginBottom(1)
	formSectionStyle = lipgloss.NewStyle().Foreground(AccentColor).MarginTop(1)
	formDecoStyle = lipgloss.NewStyle().Foreground(MutedBorderColor)

	return palette.Name
}

func normalizeThemeName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, "_", "-")
	if name == "" || name == "tokyo-night" || name == "tokyo" {
		return "tokyonight"
	}
	if name == "catppuccin" || name == "mocha" {
		return "catppuccin-mocha"
	}
	if name == "gruvbox" {
		return "gruvbox-dark"
	}
	return name
}

func AvailableThemes() []string {
	return []string{"tokyonight", "dracula", "catppuccin-mocha", "nord", "gruvbox-dark", "terminal"}
}

func FormTheme() *huh.Theme {
	theme := huh.ThemeDracula()
	theme.Focused.Title = lipgloss.NewStyle().Foreground(AccentColor).Bold(true)
	theme.Focused.Description = lipgloss.NewStyle().Foreground(MutedTextColor)
	theme.Focused.Base = lipgloss.NewStyle().Foreground(PrimaryTextColor)
	theme.Focused.SelectedOption = lipgloss.NewStyle().Foreground(GreenColor).Bold(true)
	theme.Focused.UnselectedOption = lipgloss.NewStyle().Foreground(MutedTextColor)
	theme.Blurred.Title = lipgloss.NewStyle().Foreground(MutedTextColor)
	theme.Blurred.Description = lipgloss.NewStyle().Foreground(MutedTextColor)
	theme.Blurred.Base = lipgloss.NewStyle().Foreground(MutedTextColor)
	return theme
}
