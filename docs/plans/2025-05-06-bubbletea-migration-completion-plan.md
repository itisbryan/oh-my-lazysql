# Bubbletea Migration Completion Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Complete bubbletea migration with full feature parity - integrate drivers, populate tree, wire queries, remove tview.

**Architecture:** Incremental integration - modify HomeModel to store Driver, populate Tree on init, wire Editor to ExecuteQuery, display Results, then remove tview.

**Tech Stack:** Bubble Tea, Lipgloss, Bubbles (textarea, table), drivers.Driver interface

---

## Phase 1: Driver Integration

### Task 1: Add Driver to HomeModel

**Files:**
- Modify: `ui/home_model.go`

**Step 1: Read current HomeModel structure**

Run: `cat ui/home_model.go`

**Step 2: Add driver field and import**

Modify `HomeModel` struct:
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

Add import:
```go
import (
    "github.com/jorgerojas26/lazysql/drivers"
    // ... existing imports
)
```

**Step 3: Update NewHomeModel to create driver**

Modify `NewHomeModel(data any) *HomeModel`:
```go
func NewHomeModel(data any) *HomeModel {
    conn, ok := data.(models.Connection)
    if !ok {
        conn = models.Connection{}
    }

    // Create driver based on provider
    var driver drivers.Driver
    switch conn.Provider {
    case "MySQL":
        driver = &drivers.MySQLDriver{}
    case "PostgreSQL":
        driver = &drivers.PostgresDriver{}
    case "SQLite":
        driver = &drivers.SQLiteDriver{}
    case "MSSQL":
        driver = &drivers.MSSQLDriver{}
    default:
        driver = &drivers.MySQLDriver{}
    }

    // Connect if URL provided
    if conn.URL != "" {
        if err := driver.Connect(conn.URL); err != nil {
            // Log but continue - tree will show error
            logger.Error("Failed to connect: %v", err)
        }
    }

    home := &HomeModel{
        connection: conn,
        driver:     driver,
        tree:       NewTreeModel(),
        editor:     NewEditorModel(),
        results:    NewResultsModel(),
        focus:      "tree",
    }

    // Set driver on sub-models
    home.tree.driver = driver
    home.editor.driver = driver
    home.editor.results = home.results

    return home
}
```

**Step 4: Update TreeModel and EditorModel to have driver fields**

Modify `ui/tree.go`:
```go
type TreeModel struct {
    driver    drivers.Driver
    root      *TreeNode
    cursor    int
    flattened []FlatNode
    width     int
    height    int
    focused   bool
    status   string
}
```

Modify `ui/editor.go`:
```go
type EditorModel struct {
    textarea *textarea.Model
    driver  drivers.Driver
    results *ResultsModel
    width   int
    height  int
    focused bool
}
```

**Step 5: Verify build**

Run: `go build -tags bubbletea -o lazysql-bt .`
Expected: SUCCESS

**Step 6: Commit**

```bash
git add ui/home_model.go ui/tree.go ui/editor.go
git commit -m "feat(bt): add driver integration to HomeModel"
```

---

### Task 2: Populate Tree with Databases

**Files:**
- Modify: `ui/tree.go`

**Step 1: Add SetDriver method to TreeModel**

Add after `NewTreeModel()`:
```go
func (m *TreeModel) SetDriver(driver drivers.Driver) {
    m.driver = driver
}
```

**Step 2: Modify Init to load databases**

Change `Init()`:
```go
func (m *TreeModel) Init() tea.Cmd {
    if m.driver == nil {
        return nil
    }
    return func() tea.Msg {
        databases, err := m.driver.GetDatabases()
        if err != nil {
            m.status = "Error loading databases: " + err.Error()
            return nil
        }
        m.SetDatabases(databases)
        return nil
    }
}
```

**Step 3: Update SetDatabases to create expandable nodes**

Replace current `SetDatabases`:
```go
func (m *TreeModel) SetDatabases(databases []string) {
    m.root.Children = make([]*TreeNode, 0, len(databases))
    for _, db := range databases {
        m.root.Children = append(m.root.Children, &TreeNode{
            Type:     NodeTypeDatabase,
            Name:     db,
            Database: db,
            Children: []*TreeNode{},
            Expanded: false,
        })
    }
    m.rebuildFlattened()
}
```

**Step 4: Add LoadTables method**

Add after `SetDatabases`:
```go
func (m *TreeModel) LoadTables(dbName string) error {
    if m.driver == nil {
        return fmt.Errorf("no driver")
    }

    tables, err := m.driver.GetTables(dbName)
    if err != nil {
        return err
    }

    // Find database node and add tables
    for _, child := range m.root.Children {
        if child.Name == dbName {
            child.Children = make([]*TreeNode, 0, len(tables))
            for tableName, views := range tables {
                // Create table node
                tableNode := &TreeNode{
                    Type:     NodeTypeTable,
                    Name:     tableName,
                    Database: dbName,
                    Children: []*TreeNode{},
                    Expanded: false,
                }

                // Add views as children if present
                for _, view := range views {
                    tableNode.Children = append(tableNode.Children, &TreeNode{
                        Type:     NodeTypeView,
                        Name:     view,
                        Database: dbName,
                    })
                }

                child.Children = append(child.Children, tableNode)
            }
            break
        }
    }

    m.rebuildFlattened()
    return nil
}
```

**Step 5: Handle expand in Update**

Modify key handling in `Update()`:
```go
case "enter", "right", "l":
    if m.cursor < len(m.flattened) {
        node := m.flattened[m.cursor]
        if node.Type == NodeTypeDatabase && !m.isExpanded(m.cursor) {
            // Expand - load tables
            if m.driver != nil {
                if err := m.LoadTables(node.Name); err != nil {
                    m.status = err.Error()
                }
            }
        }
        m.toggleExpanded(m.cursor)
    }
```

Add helper:
```go
func (m *TreeModel) isExpanded(idx int) bool {
    if idx < 0 || idx >= len(m.flattened) {
        return false
    }
    node := m.flattened[idx]
    for _, child := range m.root.Children {
        if child.Name == node.Name && child.Type == NodeTypeDatabase {
            return child.Expanded
        }
    }
    return false
}
```

**Step 6: Update toggleExpanded to set Expanded flag**

```go
func (m *TreeModel) toggleExpanded(idx int) {
    if idx < 0 || idx >= len(m.flattened) {
        return
    }
    node := &m.flattened[idx]

    // Find in tree and toggle
    for _, child := range m.root.Children {
        if child.Name == node.Name && child.Type == NodeTypeDatabase {
            child.Expanded = !child.Expanded
            break
        }
    }
    m.rebuildFlattened()
}
```

**Step 7: Verify build**

Run: `go build -tags bubbletea -o lazysql-bt .`
Expected: SUCCESS

**Step 8: Commit**

```bash
git add ui/tree.go
git commit -m "feat(bt): populate tree with databases from driver"
```

---

### Task 3: Wire Editor to ExecuteQuery

**Files:**
- Modify: `ui/editor.go`

**Step 1: Update EditorModel struct**

```go
type EditorModel struct {
    textarea *textarea.Model
    driver  drivers.Driver
    results *ResultsModel
    width   int
    height  int
    focused bool
    executing bool
}
```

**Step 2: Add SetDriver and SetResults methods**

```go
func (m *EditorModel) SetDriver(driver drivers.Driver) {
    m.driver = driver
}

func (m *EditorModel) SetResults(results *ResultsModel) {
    m.results = results
}
```

**Step 3: Update executeQuery to use driver**

Replace `executeQuery()`:
```go
func (m *EditorModel) executeQuery() tea.Msg {
    if m.driver == nil {
        return nil
    }
    sql := m.textarea.Value()
    if sql == "" {
        return nil
    }

    m.executing = true
    cols, rows, err := m.driver.ExecuteQuery(sql)
    m.executing = false

    if err != nil {
        if m.results != nil {
            m.results.SetStatus("Error: " + err.Error())
        }
        return nil
    }

    if m.results != nil {
        m.results.SetData(cols, rows)
        m.results.SetStatus(fmt.Sprintf("Got %d rows", len(rows)))
    }
    return nil
}
```

**Step 4: Update Update to handle execute**

Modify key handling:
```go
case "ctrl+e", "enter":
    if m.textarea.Focused() && !m.executing {
        return m, m.executeQuery
    }
```

**Step 5: Add running indicator in View**

Modify View() to show "Executing..." when running.

**Step 6: Verify build**

Run: `go build -tags bubbletea -o lazysql-bt .`
Expected: SUCCESS

**Step 7: Commit**

```bash
git add ui/editor.go
git commit -m "feat(bt): wire editor to driver.ExecuteQuery"
```

---

### Task 4: Wire Results Display

**Files:**
- Modify: `ui/results.go`

**Step 1: Review current SetData implementation**

Run: `grep -A20 "func.*SetData" ui/results.go`

**Step 2: Add SetStatus method if missing**

```go
func (m *ResultsModel) SetStatus(msg string) {
    m.status = msg
}
```

**Step 3: Verify build**

Run: `go build -tags bubbletea -o lazysql-bt .`
Expected: SUCCESS

**Step 4: Commit**

```bash
git add ui/results.go
git commit -m "feat(bt): results displays query data"
```

---

## Phase 2: Remove tview

### Task 5: Delete tview Components

**Files:**
- Delete: `components/` (entire directory)

**Step 1: Verify no remaining references**

Run: `grep -r "github.com/rivo/tview" --include="*.go" .`
Expected: Only in go.mod (will be cleaned)

**Step 2: Delete components directory**

```bash
rm -rf components/
```

**Step 3: Commit**

```bash
git add -A
git commit -m "feat: remove tview components"
```

---

### Task 6: Replace main.go with Bubbletea Version

**Files:**
- Delete: `main.go`
- Modify: `main_bubbletea.go`

**Step 1: Replace main_bubbletea.go content**

Copy `main_bubbletea.go` to `main.go` and remove build tag:

```go
package main

import (
    "flag"
    "io"
    "log"
    "os"

    "github.com/go-sql-driver/mysql"
    tea "github.com/charmbracelet/bubbletea"

    "github.com/jorgerojas26/lazysql/app"
    "github.com/jorgerojas26/lazysql/helpers/logger"
    "github.com/jorgerojas26/lazysql/models"
    "github.com/jorgerojas26/lazysql/ui"
)

var version = "dev"

func main() {
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

    logger.Info("Starting LazySQL...", nil)

    if err := mysql.SetLogger(log.New(io.Discard, "", 0)); err != nil {
        log.Fatalf("Error setting MySQL logger: %v", err)
    }

    if err := app.LoadConfig(*configFile); err != nil {
        log.Fatalf("Error loading config: %v", err)
    }

    var initModel tea.Model
    initModel = ui.NewRootModel()

    args := flag.Args()
    if len(args) == 1 {
        conn := models.Connection{
            Name:     "CLI Connection",
            URL:      args[0],
            ReadOnly: *readOnly,
        }
        initModel = ui.NewHomeModel(conn)
    } else if len(args) > 1 {
        log.Fatal("Only a single connection is allowed")
    }

    p := tea.NewProgram(initModel, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        log.Fatalf("Error running app: %v", err)
    }
}
```

**Step 2: Delete main_bubbletea.go and old main.go**

```bash
rm main_bubbletea.go main.go
```

**Step 3: Clean go.mod**

```bash
go mod tidy
```

**Step 4: Verify build**

Run: `go build -o lazysql .`
Expected: SUCCESS (single binary)

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: replace tview with bubbletea, single binary"
```

---

## Phase 3: Polish

### Task 7: Test All Features

**Manual test checklist:**
- [ ] Create and save connection
- [ ] Connect to MySQL database
- [ ] Browse databases in tree
- [ ] Expand database to see tables
- [ ] Run SELECT query
- [ ] See results in table
- [ ] Navigate results with j/k
- [ ] Error handling for bad queries
- [ ] Quit and reconnect

**Step 1: Fix any issues found**

**Step 2: Commit final polish**

```bash
git add -A
git commit -m "feat(bt): polish and final fixes"
```

---

## Summary

| Task | Description | Status |
|------|-------------|--------|
| 1 | Add Driver to HomeModel | pending |
| 2 | Populate Tree with Databases | pending |
| 3 | Wire Editor to ExecuteQuery | pending |
| 4 | Wire Results Display | pending |
| 5 | Delete tview Components | pending |
| 6 | Replace main.go | pending |
| 7 | Test All Features | pending |

**Total: 7 tasks**
