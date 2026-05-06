# Bubbletea Migration Design

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Migrate the entire TUI layer from tview to Charm Bubbletea/Lipgloss/Bubbles for better UI polish and styling control.

**Architecture:** Single `tea.Program` with screens as nested `tea.Model`s, switching via a `Screen` enum. Lipgloss handles all styling. Bubbles provides pre-built input/table components.

**Tech Stack:** Bubbletea (event loop), Lipgloss (styling), Bubbles (textinput, textarea, table, list, spinner)

---

## Architecture

```
main.go → tea.Program(RootModel)

RootModel {
    screen: Screen  // enum: ConnectionList, ConnectionForm, Home
    connectionList  // tea.Model
    connectionForm  // tea.Model
    home            // tea.Model
}

RootModel.Update(msg) → delegates to current screen's Update()
RootModel.View()       → delegates to current screen's View()

Screen switches happen via tea.Msg (ScreenChangeMsg)
```

## Screen Flow

```
ConnectionList → ConnectionForm → Home
                                    ├── Tree (sidebar)
                                    ├── SQLEditor (main)
                                    └── ResultsTable (bottom)
```

## Component Mapping

| tview Component | Bubbletea Equivalent |
|---|---|
| `tview.Form` | Custom model with `bubbles/textinput`, `lipgloss` layout |
| `tview.DropDown` | Custom button group with Lipgloss styling |
| `tview.Table` | `bubbles/table` |
| `tview.TreeView` | Custom Lipgloss tree model |
| `tview.TextView` | `lipgloss.NewStyle().Render()` + `glamour` for markdown |
| `tview.Flex` | `lipgloss.JoinHorizontal` / `lipgloss.JoinVertical` / `lipgloss.Place` |
| `tview.Pages` | Screen switching via `Screen` enum in RootModel |
| `tview.Modal` | Overlay model with Lipgloss borders + centering |
| `tview.InputField` | `bubbles/textinput` |
| `tview.TextArea` | `bubbles/textarea` |
| `tview.Button` | Lipgloss-styled key binding display |
| `tview.Checkbox` | Custom toggle model |

## Files Kept Unchanged

- `app/config.go` - config loading/saving
- `app/keymap.go` - keymapping
- `drivers/*` - all database drivers
- `helpers/*` - utilities
- `models/models.go` - data models (Profile, Connection, etc.)
- `commands/*` - command definitions
- `lib/clipboard.go` - clipboard
- `keymap/*` - keybind definitions

## Files Rewritten

- `main.go` - new `tea.Program` entry point
- `app/app.go` - adapt to Bubbletea
- `components/*` - all 32 files → new `ui/` package with `tea.Model`s

## Project Structure (New)

```
ui/
├── root.go              # RootModel (screen router)
├── screen.go            # Screen enum + ScreenChangeMsg
├── styles.go             # Global Lipgloss styles
├── connection/
│   ├── list.go          # Connection list screen
│   ├── form.go          # Connection form screen (with profiles)
│   ├── table.go         # Connection table component
│   └── profiles.go      # Profile selector buttons
├── home/
│   ├── home.go          # Home screen (tree + editor + results)
│   ├── tree.go          # Database tree sidebar
│   ├── editor.go         # SQL editor
│   ├── results.go        # Results table
│   ├── pagination.go     # Pagination controls
│   └── sidebar.go        # Sidebar component
├── components/
│   ├── input.go          # Styled text input wrapper
│   ├── button.go         # Styled button component
│   ├── toggle.go         # Checkbox/toggle component
│   ├── modal.go           # Modal overlay
│   ├── table.go           # Styled table wrapper
│   ├── help_bar.go       # Bottom help/status bar
│   └── tabs.go           # Tab bar component
└── keys.go               # Key bindings as tea.KeyMsg handlers
```

## Styling Philosophy

- **Borders:** Rounded borders with subtle colors, not heavy boxes
- **Colors:** Use the terminal's color palette via Lipgloss adaptive colors
- **Spacing:** Generous padding, don't cram everything together
- **Focus:** Clear visual distinction between focused/unfocused elements
- **Profile selector:** Pill-shaped buttons with ●/○ indicators, highlighted active

## Key Bindings

Maintain the same key bindings as current app (F1-F4, Esc, Tab, etc.) but express them as `key.Binding` with `help` descriptions using Bubbles' keybinding system.

## Data Flow

```
RootModel.Update(msg)
    → current_screen.Update(msg)
        → returns (model, cmd)
            → cmd may produce tea.Msg (ScreenChangeMsg, SaveConnectionMsg, etc.)
            → RootModel handles ScreenChangeMsg to switch screens
```

## Migration Order

1. Core infrastructure (root model, screen switching, global styles)
2. Connection screens (list, form with profiles)
3. Home screen layout (tree + editor + results)
4. Modals and overlays
5. Polish and fine-tuning

## Benefits Over tview

- Full control over rendering via Lipgloss
- Better testability (tea.Model is pure data)
- Easier styling with Lipgloss (no awkward tview style API)
- Better community support and documentation
- Consistent with modern Go TUI ecosystem
- Cleaner separation of concerns (Model/Update/View)