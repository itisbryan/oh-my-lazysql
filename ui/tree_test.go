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
