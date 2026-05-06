# Bubbletea Migration Completion Design

> **Date:** 2025-05-06
> **Goal:** Complete bubbletea migration with full feature parity

## Overview

The bubbletea migration scaffold is complete but lacks database integration. This design outlines an incremental approach to add driver integration, achieve feature parity, and remove the tview dependency.

## Current State

### Bubbletea Scaffold (Working)
- Connection list screen with navigation
- Connection form with focus management
- Home screen with tree/editor/results panels
- Modal components (confirm, help, error)
- CLI args wiring

### Missing Integration
- No `drivers.Driver` storage in HomeModel
- Tree not populated with real databases/tables
- Editor doesn't execute queries
- Results shows hardcoded sample data

### tview Version (Reference)
See `components/home.go`, `components/tree.go`, `components/sql_editor.go` for driver integration patterns.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      RootModel                               │
│  ┌─────────────────┐  ┌──────────────────┐                │
│  │ ConnectionList   │→ │ HomeModel        │                │
│  │ (exists)        │  │ ├─ TreeModel     │                │
│  └─────────────────┘  │ ├─ EditorModel   │                │
│                       │ ├─ ResultsModel   │                │
│                       │ └─ driver.Driver  │────────────────│
└─────────────────────────────────────────────────────────────┘
```

## Integration Steps

### Step 1: HomeModel Gets Driver

**File:** `ui/home_model.go`

Add `driver` field to `HomeModel`:
```go
type HomeModel struct {
    connection models.Connection
    driver    drivers.Driver
    tree      *TreeModel
    editor    *EditorModel
    results   *ResultsModel
    width     int
    height    int
    focus     string
}
```

In `NewHomeModel(data any)`:
1. Extract connection from data
2. Create driver based on connection.Provider
3. Call `driver.Connect(connection.URL)`
4. Pass driver to TreeModel for data loading

### Step 2: Tree Loads Databases

**File:** `ui/tree.go`

Modify `TreeModel`:
- Add `driver drivers.Driver` field
- Add `database string` field (current selected DB)
- In `Init()` or when database node selected, call `driver.GetDatabases()`
- Populate root with database nodes

Add expansion logic:
- On Enter on database node, call `driver.GetTables(dbName)`
- Add tables as children of database node
- Support expand/collapse

### Step 3: Editor Executes Queries

**File:** `ui/editor.go`

Modify `EditorModel`:
```go
type EditorModel struct {
    textarea *textarea.Model
    driver  drivers.Driver
    results *ResultsModel
    // ...
}
```

In `executeQuery()`:
```go
func (m *EditorModel) executeQuery() tea.Msg {
    sql := m.textarea.Value()
    cols, rows, err := m.driver.ExecuteQuery(sql)
    if err != nil {
        return ErrorMsg{Err: err}
    }
    m.results.SetData(cols, rows)
    return nil
}
```

### Step 4: Results Displays Data

**File:** `ui/results.go`

`ResultsModel` already has `SetData(cols []string, rows [][]string)`. Ensure it's wired from editor query results.

### Step 5: Remove tview

**Files to delete:**
- `components/` (entire directory, ~32 files)
- `main.go`
- `app/app.go` (may need to keep config functions)

**Files to modify:**
- `main_bubbletea.go` - Remove build tag, make primary
- `go.mod` - Remove `github.com/rivo/tview`, `github.com/gdamore/tcell/v2`

## Implementation Order

1. HomeModel gets driver (add field, create in NewHomeModel)
2. Tree loads databases on init
3. Tree expands to show tables
4. Editor executes query → Results displays
5. Delete tview components and main.go
6. Clean up go.mod

## Testing

After each step:
```bash
go build -tags bubbletea -o lazysql-bt .
./lazysql-bt
```

Manual test checklist:
- [ ] Create connection
- [ ] Connect to database
- [ ] Browse tables in tree
- [ ] Expand database to see tables
- [ ] Click table to see columns
- [ ] Enter SQL query
- [ ] See results in table
- [ ] Export CSV works
- [ ] All keyboard shortcuts work

## Notes

- Keep `go test ./...` passing throughout
- Build both versions initially, then switch to bubbletea-only
- Driver creation follows pattern in `components/connection_selection.go:doConnect()`
