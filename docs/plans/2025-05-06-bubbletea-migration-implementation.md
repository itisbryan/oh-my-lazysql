# Bubbletea Migration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Migrate the entire TUI layer from tview to Charm Bubbletea/Lipgloss/Bubbles for better UI polish and styling control.

**Architecture:** Single `tea.Program` with screens as nested `tea.Model`s. Lipgloss handles styling. Bubbles provides input/table components. Screen switching via `ScreenChangeMsg`.

**Tech Stack:** Bubbletea, Lipgloss, Bubbles (textinput, textarea, table, list, spinner), Glamour (markdown rendering)

---

## Phase 1: Foundation

### Task 1: Install Bubbletea Dependencies

**Files:**
- Modify: `go.mod`

**Step 1: Add dependencies**

```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles/v2
go get github.com/charmbracelet/glamour
```

**Step 2: Verify**

Run: `go build ./...`
Expected: SUCCESS (no import errors yet since we haven't used them)

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add bubbletea, lipgloss, bubbles dependencies"
```

---

### Task 2: Create Root Model and Screen Router

**Files:**
- Create: `ui/root.go`
- Create: `ui/screen.go`
- Create: `ui/styles.go`

**Step 1: Create screen enum**

`ui/screen.go`:
```go
package ui

type Screen int

const (
	ScreenConnectionList Screen = iota
	ScreenConnectionForm
	ScreenHome
)

type ScreenChangeMsg struct {
	Screen Screen
	Data   any
}

type ConnectionSelectedMsg struct {
	Connection any
}

type ErrorMsg struct {
	Err error
}
```

**Step 2: Create global styles**

`ui/styles.go`:
```go
package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Base colors matching current tview theme
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
```

**Step 3: Create root model**

`ui/root.go`:
```go
package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type RootModel struct {
	screen            Screen
	connectionList    tea.Model
	connectionForm    tea.Model
	home              tea.Model
	width             int
	height            int
}

func NewRootModel() RootModel {
	return RootModel{
		screen:         ScreenConnectionList,
		connectionList: NewConnectionListModel(),
	}
}

func (m RootModel) Init() tea.Cmd {
	return m.connectionList.Init()
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case ScreenChangeMsg:
		m.screen = msg.Screen
		switch msg.Screen {
		case ScreenConnectionForm:
			m.connectionForm = NewConnectionFormModel(msg.Data)
			return m, m.connectionForm.Init()
		case ScreenHome:
			m.home = NewHomeModel(msg.Data)
			return m, m.home.Init()
		}
	}

	switch m.screen {
	case ScreenConnectionList:
		m.connectionList, cmd = m.connectionList.Update(msg)
	case ScreenConnectionForm:
		m.connectionForm, cmd = m.connectionForm.Update(msg)
	case ScreenHome:
		m.home, cmd = m.home.Update(msg)
	}

	return m, cmd
}

func (m RootModel) View() string {
	switch m.screen {
	case ScreenConnectionList:
		return m.connectionList.View()
	case ScreenConnectionForm:
		return m.connectionForm.View()
	case ScreenHome:
		return m.home.View()
	default:
		return m.connectionList.View()
	}
}
```

**Step 4: Update main.go (add bubbletea entry, keep tview temporarily)**

Create a new `main_bubbletea.go` file so we can run both side by side during development:

`main_bubbletea.go`:
```go
//go:build bubbletea

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/itisbryan/oh-my-lazysql/app"
	"github.com/itisbryan/oh-my-lazysql/helpers/logger"
	"github.com/itisbryan/oh-my-lazysql/ui"
)

var version = "dev"

func main() {
	// Same flag parsing as current main.go...
	defaultConfigPath, err := app.DefaultConfigFile()
	if err != nil {
		log.Fatalf("Error getting default config file: %v", err)
	}
	configFile := flag.String("config", defaultConfigPath, "config file to use")
	printVersion := flag.Bool("version", false, "Show version")
	logLevel := flag.String("loglevel", "info", "Log level")
	logFile := flag.String("logfile", "", "Log file")
	readOnly := flag.Bool("read-only", false, "Connect in read-only mode")
	flag.Parse()

	if *printVersion {
		println("LazySQL version: ", version)
		os.Exit(0)
	}

	slogLevel, err := logger.ParseLogLevel(*logLevel)
	if err != nil {
		log.Fatalf("Error parsing log level: %v", err)
	}
	logger.SetLevel(slogLevel)

	if *logFile != "" {
		if err := logger.SetFile(*logFile); err != nil {
			log.Fatalf("Error setting log file: %v", err)
		}
	}

	logger.Info("Starting LazySQL (Bubbletea)...", nil)

	if err := mysql.SetLogger(log.New(io.Discard, "", 0)); err != nil {
		log.Fatalf("Error setting MySQL logger: %v", err)
	}

	if err := app.LoadConfig(*configFile); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	args := flag.Args()
	_ = args
	_ = readOnly  // TODO: handle CLI args

	p := tea.NewProgram(ui.NewRootModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running app: %v", err)
	}
}
```

**Step 5: Verify build**

Run: `go build -tags bubbletea -o lazysql-bt .`
Expected: FAILS (missing ui package implementations - that's OK, we're scaffolding)

**Step 6: Commit**

```bash
git add ui/ main_bubbletea.go
git commit -m "feat: scaffold bubbletea root model, screen router, and styles"
```

---

### Task 3: Create Reusable UI Components

**Files:**
- Create: `ui/components/input.go`
- Create: `ui/components/button.go`
- Create: `ui/components/toggle.go`
- Create: `ui/components/modal.go`

**Step 1: Create styled input component**

`ui/components/input.go`:
```go
package components

import (
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/itisbryan/oh-my-lazysql/ui"
)

type InputModel struct {
	Label string
	Input textinput.Model
	Focus bool
}

func NewInput(label string, placeholder string) InputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Width = 30
	return InputModel{
		Label: label,
		Input: ti,
	}
}

func (m InputModel) Focus() InputModel {
	m.Focus = true
	m.Input.Focus()
	return m
}

func (m InputModel) Blur() InputModel {
	m.Focus = false
	m.Input.Blur()
	return m
}

func (m InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	var cmd tea.Cmd
	m.Input, cmd = m.Input.Update(msg)
	return m, cmd
}

func (m InputModel) View() string {
	label := ui.InputLabelStyle.Render(m.Label + ":")
	return lipgloss.JoinHorizontal(lipgloss.Top, label, " ", m.Input.View())
}

func (m InputModel) Value() string {
	return m.Input.Value()
}

func (m InputModel) SetValue(val string) InputModel {
	m.Input.SetValue(val)
	return m
}
```

**Step 2: Create button component**

`ui/components/button.go`:
```go
package components

import (
	"github.com/charmbracelet/lipgloss"
)

type ButtonModel struct {
	Label   string
	Style   lipgloss.Style
	Active  bool
}

func NewButton(label string, style lipgloss.Style) ButtonModel {
	return ButtonModel{
		Label: label,
		Style: style,
	}
}

func (m ButtonModel) View() string {
	if m.Active {
		return m.Style.Bold(true).Render(m.Label)
	}
	return m.Style.Render(m.Label)
}

func (m ButtonModel) SetActive(active bool) ButtonModel {
	m.Active = active
	return m
}
```

**Step 3: Create toggle (checkbox) component**

`ui/components/toggle.go`:
```go
package components

import "github.com/charmbracelet/lipgloss"

type ToggleModel struct {
	Label   string
	Checked bool
	Focused bool
}

func NewToggle(label string, checked bool) ToggleModel {
	return ToggleModel{
		Label:   label,
		Checked: checked,
	}
}

func (m ToggleModel) Toggle() ToggleModel {
	m.Checked = !m.Checked
	return m
}

func (m ToggleModel) View() string {
	checkbox := "[ ]"
	if m.Checked {
		checkbox = "[✓]"
	}
	label := m.Label
	if m.Focused {
		checkbox = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render(checkbox)
	}
	return checkbox + " " + label
}

func (m ToggleModel) Value() bool {
	return m.Checked
}
```

**Step 4: Create modal component**

`ui/components/modal.go`:
```go
package components

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/itisbryan/oh-my-lazysql/ui"
)

type ModalModel struct {
	Title   string
	Content string
	Width   int
	Height  int
	Open    bool
	Buttons []string
	Active  int
}

func NewModal(title string) ModalModel {
	return ModalModel{
		Title:  title,
		Width:  60,
		Height: 10,
		Open:   true,
	}
}

func (m ModalModel) View() string {
	if !m.Open {
		return ""
	}

	box := lipgloss.NewStyle().
		Width(m.Width).
		Height(m.Height).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666A7E")).
		Padding(1, 2)

	title := ui.TitleStyle.Render(m.Title)
	content := m.Content + "\n\n"

	buttons := ""
	for i, btn := range m.Buttons {
		style := lipgloss.NewStyle().Padding(0, 2)
		if i == m.Active {
			style = style.Background(lipgloss.Color("#0000FF")).Foreground(lipgloss.Color("#FFFFFF"))
		}
		buttons += style.Render(btn) + "  "
	}

	return lipgloss.Place(m.Width+4, m.Height+4,
		lipgloss.Center, lipgloss.Center,
		box.Render(title+"\n\n"+content+buttons))
}
```

**Step 5: Verify build**

Run: `go build -tags bubbletea ./ui/...`
Expected: SUCCESS

**Step 6: Commit**

```bash
git add ui/components/
git commit -m "feat: add reusable bubbletea UI components (input, button, toggle, modal)"
```

---

## Phase 2: Connection Screens

### Task 4: Implement Connection List Screen

**Files:**
- Create: `ui/connection/list.go`

This replaces `components/connection_selection.go`, `components/connections_table.go`, `components/connection_page.go`.

**Step 1: Implement connection list model**

The connection list shows saved connections in a styled table, with F1/Enter to connect, E to edit, N for new, D to delete.

`ui/connection/list.go`:
```go
package connection

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/itisbryan/oh-my-lazysql/app"
	"github.com/itisbryan/oh-my-lazysql/models"
	ui "github.com/itisbryan/oh-my-lazysql/ui"
	uiComponents "github.com/itisbryan/oh-my-lazysql/ui/components"
)

type ConnectionListModel struct {
	connections []models.Connection
	cursor      int
	width       int
	height      int
	status      string
	statusErr   bool
}

func NewConnectionListModel() ConnectionListModel {
	return ConnectionListModel{
		connections: app.App.Connections(),
	}
}

func (m ConnectionListModel) Init() tea.Cmd {
	return nil
}

func (m ConnectionListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
					return ui.ScreenChangeMsg{Screen: ui.ScreenHome, Data: conn}
				}
			}
		case "n":
			return m, func() tea.Msg {
				return ui.ScreenChangeMsg{Screen: ui.ScreenConnectionForm, Data: nil}
			}
		case "e":
			if len(m.connections) > 0 {
				conn := m.connections[m.cursor]
				return m, func() tea.Msg {
					return ui.ScreenChangeMsg{Screen: ui.ScreenConnectionForm, Data: conn}
				}
			}
		case "d":
			if len(m.connections) > 0 {
				// TODO: confirmation modal
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

func (m ConnectionListModel) View() string {
	// Header
	header := ui.TitleStyle.Render("Connections")
	subtitle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("Select a connection to start")

	// Table
	rows := make([]string, len(m.connections))
	for i, conn := range m.connections {
		name := conn.Name
		if name == "" {
			name = conn.URL
		}
		profileCount := ""
		if len(conn.Profiles) > 1 {
			profileCount = fmt.Sprintf(" (%d profiles)", len(conn.Profiles))
		}

		provider := conn.Provider
		style := lipgloss.NewStyle()
		if i == m.cursor {
			style = ui.SelectedStyle
		}
		rows[i] = style.Render(fmt.Sprintf("  %-20s %-12s%s", name, provider, profileCount))
	}

	if len(rows) == 0 {
		rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("  No connections. Press N to add one."))
	}

	table := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666A7E")).
		Padding(0, 1).
		Render(lipgloss.JoinVertical(lipgloss.Left, rows...))

	// Help bar
	help := ui.HelpStyle.Render("[N] New  [E] Edit  [D] Delete  [Enter] Connect  [Q] Quit")

	// Status bar
	statusBar := ""
	if m.status != "" {
		style := ui.StatusStyle
		if m.statusErr {
			style = ui.ErrorStyle
		}
		statusBar = style.Render(m.status)
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		subtitle,
		"",
		table,
		"",
		help,
		statusBar,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
```

**Step 2: Verify**

Run: `go build -tags bubbletea ./ui/connection/`
Expected: May have import issues, fix them

**Step 3: Commit**

```bash
git add ui/connection/
git commit -m "feat: add bubbletea connection list screen"
```

---

### Task 5: Implement Connection Form Screen (with Profiles)

**Files:**
- Create: `ui/connection/form.go`
- Create: `ui/connection/profiles.go`

This replaces `components/connection_form.go` with a polished Bubbletea version that has proper styling for the provider selector, form fields, and profile buttons.

**Key differences from tview version:**
- Provider selector is a row of styled buttons (not ugly dropdown)
- SSL fields only show when SSL enabled
- Profile buttons are pill-shaped with ●/○ indicators
- All styling uses Lipgloss for pixel-perfect control
- Clean separation of form state from rendering

**Step 1: Implement profile selector**

`ui/connection/profiles.go`:
```go
package connection

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	ui "github.com/itisbryan/oh-my-lazysql/ui"
	"github.com/itisbryan/oh-my-lazysql/models"
)

type ProfileSelectorModel struct {
	profiles     []models.Profile
	activeIndex  int
	focused      bool
	width        int
}

func NewProfileSelector() ProfileSelectorModel {
	return ProfileSelectorModel{
		profiles:    []models.Profile{{Name: "default"}},
		activeIndex: 0,
	}
}

func (m ProfileSelectorModel) Init() tea.Cmd { return nil }

func (m ProfileSelectorModel) Update(msg tea.Msg) (ProfileSelectorModel, tea.Cmd) {
	return m, nil
}

func (m ProfileSelectorModel) View() string {
	elements := make([]string, 0, len(m.profiles)*2+4)

	elements = append(elements, ui.ProfileActiveStyle.Render("+"))
	elements = append(elements, " ")
	if len(m.profiles) > 1 {
		elements = append(elements, ui.ErrorStyle.Render("-"))
		elements = append(elements, " ")
	}

	for i, p := range m.profiles {
		indicator := "○"
		if i == m.activeIndex {
			indicator = "●"
		}
		style := ui.ProfileInactiveStyle
		if i == m.activeIndex {
			style = ui.ProfileActiveStyle
		}
		elements = append(elements, style.Render(fmt.Sprintf("%s %s", indicator, p.Name)))
	}

	return lipgloss.JoinHorizontal(lipgloss.Middle, elements...)
}

func (m *ProfileSelectorModel) AddProfile(name string) {
	current := m.getCurrentFormValues()
	current.Name = name
	m.profiles = append(m.profiles, m.activeIndex+1, current)
	m.activeIndex++
}

func (m *ProfileSelectorModel) DeleteActive() {
	if len(m.profiles) <= 1 {
		return
	}
	m.profiles = append(m.profiles[:m.activeIndex], m.profiles[m.activeIndex+1:]...)
	if m.activeIndex >= len(m.profiles) {
		m.activeIndex = len(m.profiles) - 1
	}
}

func (m *ProfileSelectorModel) SetProfiles(profiles []models.Profile) {
	m.profiles = profiles
	if len(m.profiles) == 0 {
		m.profiles = []models.Profile{{Name: "default"}}
	}
	m.activeIndex = 0
}

func (m *ProfileSelectorModel) GetActive() models.Profile {
	return m.profiles[m.activeIndex]
}

func (m *ProfileSelectorModel) getCurrentFormValues() models.Profile {
	// This will be populated from the form state
	return models.Profile{Name: "new"}
}
```

**Step 2: Implement connection form**

`ui/connection/form.go` - Full form model with provider buttons, input fields, SSL toggle, profile selector, and advanced URL. This is ~400 lines. Key structure:

```go
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
	action       string // "new" or "edit"
	editConn     *models.Connection

	focusIndex   int // which field is focused
	status       string
	statusErr    bool
	width        int
	height       int
}
```

**Step 3: Verify**

Run: `go build -tags bubbletea ./ui/connection/`

**Step 4: Commit**

```bash
git add ui/connection/
git commit -m "feat: add bubbletea connection form with profiles"
```

---

## Phase 3: Home Screen

### Task 6: Implement Home Screen Layout

**Files:**
- Create: `ui/home/home.go`

This replaces `components/home.go` with a Lipgloss layout that joins tree, editor, and results panels.

**Step 1: Implement home model**

`ui/home/home.go`:
```go
package home

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	ui "github.com/itisbryan/oh-my-lazysql/ui"
	"github.com/itisbryan/oh-my-lazysql/models"
	"github.com/itisbryan/oh-my-lazysql/drivers"
)

type HomeModel struct {
	connection models.Connection
	driver     drivers.Driver
	tree       TreeModel
	editor     EditorModel
	results    ResultsModel
	sidebar    SidebarModel
	width      int
	height     int
	focus      string // "tree", "editor", "results"
}

func NewHomeModel(data any) HomeModel {
	conn, ok := data.(models.Connection)
	if !ok {
		conn = models.Connection{}
	}
	return HomeModel{
		connection: conn,
		tree:       NewTreeModel(),
		editor:     NewEditorModel(),
		results:    NewResultsModel(),
		sidebar:    NewSidebarModel(),
		focus:      "tree",
	}
}
```

**Step 2: Verify**

Run: `go build -tags bubbletea ./ui/home/`

**Step 3: Commit**

```bash
git add ui/home/
git commit -m "feat: add bubbletea home screen scaffold"
```

---

### Task 7: Implement Database Tree

**Files:**
- Create: `ui/home/tree.go`

This replaces `components/tree.go` (976 lines). Use Lipgloss styled tree with expand/collapse.

**Step 1: Implement tree model**

Key features:
- Expandable/collapsible nodes (databases, schemas, tables, columns)
- Styled icons for each node type (📁, 📊, 📋)
- Navigation with j/k or up/down
- Enter to expand/collapse or select

**Step 2: Commit**

```bash
git add ui/home/tree.go
git commit -m "feat: add bubbletea database tree"
```

---

### Task 8: Implement SQL Editor

**Files:**
- Create: `ui/home/editor.go`

This replaces `components/sql_editor.go` and `components/sql_editor_enhanced.go` (786 lines combined).

Use `bubbles/textarea` with syntax highlighting via Lipgloss styles.

**Step 1: Implement editor model**

Key features:
- Multi-line text input via `bubbles/textarea`
- SQL syntax highlighting via Lipgloss
- History navigation (up/down)
- Execute with Enter or Ctrl+Enter

**Step 2: Commit**

```bash
git add ui/home/editor.go
git commit -m "feat: add bubbletea SQL editor"
```

---

### Task 9: Implement Results Table

**Files:**
- Create: `ui/home/results.go`
- Create: `ui/home/pagination.go`

This replaces `components/results_table.go` (1921 lines) and `components/pagination.go`.

Use `bubbles/table` with Lipgloss styling.

**Step 1: Implement results table model**

Key features:
- Sortable columns
- Row selection
- Pagination
- Cell editing
- JSON viewer toggle

**Step 2: Commit**

```bash
git add ui/home/
git commit -m "feat: add bubbletea results table with pagination"
```

---

## Phase 4: Modals and Polish

### Task 10: Implement Modals System

**Files:**
- Create: `ui/components/confirm_modal.go`
- Create: `ui/components/help_modal.go`
- Create: `ui/components/error_modal.go`

**Step 1: Implement modals**

Replaces: `components/confirmation_modal.go`, `components/help_modal.go`, `components/error_modal.go`, `components/csv_export_modal.go`, `components/save_query_modal.go`, `components/query_history_modal.go`, `components/query_preview_modal.go`

Each modal is a `tea.Model` with focused buttons, overlay rendering, and centered positioning via `lipgloss.Place`.

**Step 2: Commit**

```bash
git add ui/components/
git commit -m "feat: add bubbletea modal components"
```

---

### Task 11: Wire Up CLI Args and Config Loading

**Files:**
- Modify: `main.go` (or `main_bubbletea.go`)

**Step 1: Handle CLI args**

- Parse connection URL from args
- Handle `--config` flag
- Handle `--read-only` flag
- Skip connection list if URL provided

**Step 2: Handle config loading**

- Load config on startup
- Connect to database from CLI arg

**Step 3: Commit**

```bash
git add main.go
git commit -m "feat: wire up CLI args and config loading for bubbletea"
```

---

### Task 12: Remove tview Dependency

**Files:**
- Delete: `components/` directory (all 32 files)
- Delete: `app/app.go` (replace with bubbletea version)
- Modify: `go.mod` (remove tview, tcell)

**Step 1: Verify bubbletea version handles all features**

**Step 2: Delete old components**

```bash
rm -rf components/
```

**Step 3: Remove tview/tcell from go.mod**

```bash
go mod tidy
```

**Step 4: Commit**

```bash
git add -A
git commit -m "feat: remove tview dependency, fully on bubbletea"
```

---

### Task 13: Polish and Final Testing

**Files:**
- Modify: various `ui/` files for polish

**Step 1: Test all features**

- Create connection with individual fields
- Create connection with connection string (advanced mode)
- Add/delete/select profiles
- Edit existing connection
- Connect to database
- Browse tables
- Execute SQL queries
- Edit rows
- Export CSV

**Step 2: Fix styling issues**

- Check all screens render correctly in different terminal sizes
- Verify focus indicators are clear
- Verify colors contrast well

**Step 3: Final commit**

```bash
git add -A
git commit -m "feat: polish bubbletea UI"
```

---

## Summary

| Task | Description | Est. Lines |
|------|-------------|-----------|
| 1 | Install Bubbletea deps | ~5 |
| 2 | Root model + screen router + styles | ~200 |
| 3 | Reusable UI components | ~200 |
| 4 | Connection list screen | ~150 |
| 5 | Connection form with profiles | ~400 |
| 6 | Home screen layout | ~100 |
| 7 | Database tree | ~400 |
| 8 | SQL editor | ~200 |
| 9 | Results table + pagination | ~500 |
| 10 | Modals system | ~300 |
| 11 | CLI args + config wiring | ~100 |
| 12 | Remove tview dependency | -8600 |
| 13 | Polish + testing | ~200 |

**Total new code: ~2755 lines** (vs 8648 tview lines)
**Net reduction: ~5900 lines** (Bubbletea is more concise)