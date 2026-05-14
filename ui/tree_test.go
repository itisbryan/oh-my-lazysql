package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTreeSlashFiltersSidebarNodes(t *testing.T) {
	model := treeWithSearchFixture()

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

	view := model.View()
	if !strings.Contains(view, "/ user") {
		t.Fatalf("expected sidebar search prompt to show query\n%s", view)
	}
	if !strings.Contains(view, "users") || !strings.Contains(view, "public") || !strings.Contains(view, "app") {
		t.Fatalf("expected matching table with parent context\n%s", view)
	}
	if strings.Contains(view, "orders") {
		t.Fatalf("expected non-matching table to be filtered out\n%s", view)
	}
}

func TestTreeSearchEscRestoresFullTree(t *testing.T) {
	model := treeWithSearchFixture()

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	view := model.View()
	if strings.Contains(view, "/ u") {
		t.Fatalf("expected search prompt to close\n%s", view)
	}
	if !strings.Contains(view, "users") || !strings.Contains(view, "orders") {
		t.Fatalf("expected full tree restored\n%s", view)
	}
}

func TestTreeCtrlDAndCtrlUScrollLikeResults(t *testing.T) {
	model := treeWithManyTablesFixture(12)

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	if model.cursor != 5 {
		t.Fatalf("expected ctrl+d to move cursor down 5 rows, got %d", model.cursor)
	}
	if model.scanPrefix != "T" {
		t.Fatalf("expected ctrl+d to expose selected row prefix T, got %q", model.scanPrefix)
	}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	if model.cursor != 0 {
		t.Fatalf("expected ctrl+u to move cursor up 5 rows, got %d", model.cursor)
	}
	if model.scanPrefix != "A" {
		t.Fatalf("expected ctrl+u to expose selected row prefix A, got %q", model.scanPrefix)
	}
}

func TestTreeSetTablesSortsTablesAscendingAZ(t *testing.T) {
	model := NewTreeModel()
	dbNode := &TreeNode{Type: NodeTypeDatabase, Name: "app", Database: "app", Expanded: true}

	model.setTables(dbNode, "app", map[string][]string{
		"public": {"zebra", "accounts", "Orders", "billing"},
	})

	got := make([]string, 0, len(dbNode.Children))
	for _, child := range dbNode.Children {
		got = append(got, child.Name)
	}
	want := []string{"accounts", "billing", "Orders", "zebra"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("expected tables sorted A-Z, got %v", got)
	}
}

func treeWithSearchFixture() *TreeModel {
	model := NewTreeModel()
	model.width = 60
	model.height = 20
	model.focused = true
	model.root.Children = []*TreeNode{{
		Type:     NodeTypeDatabase,
		Name:     "app",
		Database: "app",
		Expanded: true,
		Children: []*TreeNode{{
			Type:     NodeTypeSection,
			Name:     "public",
			Database: "app",
			Schema:   "public",
			Expanded: true,
			Children: []*TreeNode{
				{Type: NodeTypeTable, Name: "users", Database: "app", Schema: "public"},
				{Type: NodeTypeTable, Name: "orders", Database: "app", Schema: "public"},
			},
		}},
	}}
	model.rebuildFlattened()
	return model
}

func treeWithManyTablesFixture(count int) *TreeModel {
	model := treeWithSearchFixture()
	tables := make([]*TreeNode, 0, count)
	for i := 0; i < count; i++ {
		tables = append(tables, &TreeNode{Type: NodeTypeTable, Name: "table_" + string(rune('a'+i)), Database: "app", Schema: "public"})
	}
	model.root.Children[0].Children[0].Children = tables
	model.rebuildFlattened()
	return model
}
