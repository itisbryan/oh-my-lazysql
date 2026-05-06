# Tree Icons And Colors Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add Nerd Font icons and semantic colors to the database schema tree.

**Architecture:** Keep the change localized to `components/tree.go`. Add small label/color helpers and use them where `tview.TreeNode` instances are created, while preserving raw `SetReference` values for search, selection, and database actions.

**Tech Stack:** Go, `rivo/tview`, `gdamore/tcell/v2`.

---

### Task 1: Add Tree Label And Color Helpers

**Files:**
- Modify: `components/tree.go:40-56`
- Create: `components/tree_test.go`

**Step 1: Write the failing tests**

Add focused tests for icon label generation and system schema detection:

```go
package components

import (
	"strings"
	"testing"
)

func TestTreeNodeLabelAddsNerdFontPrefix(t *testing.T) {
	label := treeNodeLabel(NodeTypeTable, "addresses")

	if !strings.Contains(label, "addresses") {
		t.Fatalf("expected label to keep raw name, got %q", label)
	}

	if label == "addresses" {
		t.Fatal("expected label to include an icon prefix")
	}
}

func TestIsSystemSchema(t *testing.T) {
	for _, schema := range []string{"information_schema", "pg_catalog", "mysql", "sys"} {
		if !isSystemSchema(schema) {
			t.Fatalf("expected %q to be treated as a system schema", schema)
		}
	}

	if isSystemSchema("public") {
		t.Fatal("expected public to remain a normal schema")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./components -run 'TestTreeNodeLabel|TestIsSystemSchema'`

Expected: FAIL because helpers do not exist yet.

**Step 3: Implement the helpers**

Add constants/helpers near `TreeNodeType`:

```go
const (
	treeIconDatabase  = "󰆼"
	treeIconSchema    = "󰙅"
	treeIconTables    = "󰓫"
	treeIconTable     = "󰓱"
	treeIconView      = "󰈙"
	treeIconFunction  = "󰊕"
	treeIconProcedure = "󰡱"
)

func treeNodeLabel(nodeType TreeNodeType, label string) string {
	switch nodeType {
	case NodeTypeDatabase:
		return fmt.Sprintf("%s %s", treeIconDatabase, label)
	case NodeTypeTable:
		return fmt.Sprintf("%s %s", treeIconTable, label)
	case NodeTypeFunction:
		return fmt.Sprintf("%s %s", treeIconFunction, label)
	case NodeTypeProcedure:
		return fmt.Sprintf("%s %s", treeIconProcedure, label)
	case NodeTypeView:
		return fmt.Sprintf("%s %s", treeIconView, label)
	default:
		return label
	}
}

func treeSectionLabel(label string) string {
	switch label {
	case "tables":
		return fmt.Sprintf("%s %s", treeIconTables, label)
	case "functions":
		return fmt.Sprintf("%s %s", treeIconFunction, label)
	case "procedures":
		return fmt.Sprintf("%s %s", treeIconProcedure, label)
	case "views":
		return fmt.Sprintf("%s %s", treeIconView, label)
	default:
		return fmt.Sprintf("%s %s", treeIconSchema, label)
	}
}

func isSystemSchema(name string) bool {
	switch name {
	case "information_schema", "pg_catalog", "mysql", "sys":
		return true
	default:
		return false
	}
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./components -run 'TestTreeNodeLabel|TestIsSystemSchema'`

Expected: PASS.

---

### Task 2: Apply Helpers To Tree Nodes

**Files:**
- Modify: `components/tree.go:326-468`
- Modify: `components/tree.go:819-824`

**Step 1: Update database/schema/table nodes**

Replace plain labels with helper-generated labels:

```go
rootNode = tview.NewTreeNode(treeSectionLabel(key))
tablesNode := tview.NewTreeNode(treeSectionLabel("tables"))
childNode := tview.NewTreeNode(treeNodeLabel(NodeTypeTable, child))
```

For top-level databases:

```go
childNode := tview.NewTreeNode(treeNodeLabel(NodeTypeDatabase, database))
```

Keep every existing `SetReference(...)` unchanged.

**Step 2: Update programming nodes**

Use section labels and object labels:

```go
functionsNode = tview.NewTreeNode(treeSectionLabel("functions"))
functionNode := tview.NewTreeNode(treeNodeLabel(NodeTypeFunction, function))

proceduresNode = tview.NewTreeNode(treeSectionLabel("procedures"))
procedureNode := tview.NewTreeNode(treeNodeLabel(NodeTypeProcedure, procedure))

viewsNode = tview.NewTreeNode(treeSectionLabel("views"))
viewNode := tview.NewTreeNode(treeNodeLabel(NodeTypeView, view))
```

**Step 3: Fix database lookup in programming nodes**

Because node labels now include icons, change:

```go
database := node.GetText()
```

to:

```go
database := node.GetReference().(string)
```

**Step 4: Run package tests**

Run: `go test ./components`

Expected: PASS.

---

### Task 3: Add Semantic Colors

**Files:**
- Modify: `components/tree.go:40-56`
- Modify: `components/tree.go:326-468`
- Modify: `components/tree.go:819-824`

**Step 1: Add color helper**

```go
func treeNodeColor(nodeType TreeNodeType, label string) tcell.Color {
	switch nodeType {
	case NodeTypeDatabase:
		return tcell.ColorDeepSkyBlue
	case NodeTypeFunction:
		return tcell.ColorMediumPurple
	case NodeTypeProcedure:
		return tcell.ColorOrange
	case NodeTypeView:
		return tcell.ColorGreen
	case NodeTypeSection:
		if isSystemSchema(label) {
			return tcell.ColorGray
		}
		return app.Styles.SecondaryTextColor
	default:
		return app.Styles.PrimaryTextColor
	}
}
```

**Step 2: Use semantic colors during node creation**

Replace selected `SetColor(app.Styles.PrimaryTextColor)` calls with `treeNodeColor(...)` based on the node type and raw label.

**Step 3: Preserve focus/blur behavior**

Keep focus and selected node text tag behavior unchanged. Do not change `SetReference` or selected-node event logic.

**Step 4: Run full tests**

Run: `go test ./...`

Expected: PASS.

---

### Task 4: Manual Smoke Test

**Files:**
- No code changes unless the smoke test reveals an issue.

**Step 1: Run the app**

Run the local app with an existing connection configuration:

```bash
go run .
```

Expected: schema tree renders with Nerd Font icons and semantic colors.

**Step 2: Verify interactions**

Check these behaviors:

- Expanding/collapsing databases and sections still works.
- Selecting a table still loads records.
- Search still finds table names.
- Function/procedure/view selection still works where supported.
- Focus highlight still applies to the current tree node.

**Step 3: Final verification**

Run: `go test ./...`

Expected: PASS.
