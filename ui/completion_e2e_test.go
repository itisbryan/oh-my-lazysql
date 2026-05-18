package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/itisbryan/oh-my-lazysql/models"
)

func TestCompletionContextAfterRecordLoad(t *testing.T) {
	model := NewResultsModel()
	model.width = 100
	model.height = 24
	model.columns = []GridColumn{{Title: "id"}, {Title: "email"}, {Title: "created_at"}}
	model.rows = [][]string{{"1", "a@b.com", "2026-01-01"}}
	model.totalRows = 1

	model.completion.TableNames = []string{"users", "orders"}
	model.completion.ColumnNames = []string{"id", "email", "created_at"}

	model.filterEditing = true
	model.filterInput = "e"
	model.completion.Update(model.filterInput)

	if !model.completion.Visible {
		t.Fatalf("expected completion visible after typing 'e', got Visible=%v, Suggestions=%v", model.completion.Visible, model.completion.Suggestions)
	}

	hasEmail := false
	for _, s := range model.completion.Suggestions {
		if s.Text == "email" && s.Kind == columnSuggestion {
			hasEmail = true
		}
	}
	if !hasEmail {
		t.Fatalf("expected 'email' column suggestion, got %v", model.completion.Suggestions)
	}
}

func TestCompletionDropdownVisibleInView(t *testing.T) {
	model := NewResultsModel()
	model.width = 100
	model.height = 24
	model.columns = []GridColumn{{Title: "id"}, {Title: "email"}}
	model.rows = [][]string{{"1", "test@example.com"}}
	model.totalRows = 1

	model.completion.TableNames = []string{"users"}
	model.completion.ColumnNames = []string{"id", "email"}

	model.filterEditing = true
	model.filterInput = "e"
	model.completion.Update(model.filterInput)

	if !model.completion.Visible {
		t.Fatal("expected completion visible")
	}

	view := model.View()
	if !strings.Contains(view, "email") {
		t.Fatalf("expected view to contain 'email' completion suggestion\n%s", view)
	}
}

func TestEditorCompletionDropdownVisibleInView(t *testing.T) {
	model := NewEditorModel()
	model.width = 80
	model.height = 20
	model.lines = []string{"sel"}
	model.cursorRow = 0
	model.cursorCol = 3
	model.mode = insertMode
	model.completion.TableNames = []string{"users"}
	model.completion.ColumnNames = []string{"id", "email"}
	model.completion.Update(model.text())

	if !model.completion.Visible {
		t.Fatal("expected completion visible")
	}

	foundSelect := false
	for _, s := range model.completion.Suggestions {
		if s.Text == "SELECT" {
			foundSelect = true
		}
	}
	if !foundSelect {
		t.Fatalf("expected SELECT suggestion, got %v", model.completion.Suggestions)
	}
}

func TestHomeModelFilterTypingShowsTables(t *testing.T) {
	driver := &fakeHomeDriver{
		columns:   [][]string{{"column_name", "data_type"}, {"id", "integer"}, {"email", "text"}},
		records:   [][]string{{"id", "email"}, {"1", "a@b.com"}},
		sort:      "id DESC",
		totalRows: 1,
	}
	model := NewHomeModel(models.Connection{})
	model.driver = driver
	model.tree.driver = driver
	model.tree.SetDatabases([]string{"testdb"})
	model.tree.root.Children[0].Expanded = true
	model.tree.setTables(model.tree.root.Children[0], "testdb", map[string][]string{"public": {"users", "orders"}}, nil, nil)
	model.tree.rebuildFlattened()
	model.results.columns = []GridColumn{{Title: "id"}, {Title: "email"}}
	model.results.rows = [][]string{{"1", "a@b.com"}}
	model.currentTable = "users"
	model.focus = "results"
	model.width = 120
	model.height = 40

	model.propagateCompletionContext()

	if len(model.results.completion.TableNames) == 0 {
		t.Fatalf("expected table names after propagation, got %v", model.results.completion.TableNames)
	}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if !model.results.filterEditing {
		t.Fatal("expected filter editing after /")
	}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})

	if !model.results.completion.Visible {
		t.Fatalf("expected completion visible after typing 'u', TableNames=%v, filterInput=%q",
			model.results.completion.TableNames, model.results.filterInput)
	}

	foundUsers := false
	for _, s := range model.results.completion.Suggestions {
		if s.Text == "users" && s.Kind == tableSuggestion {
			foundUsers = true
		}
	}
	if !foundUsers {
		t.Fatalf("expected 'users' table suggestion, got %v", model.results.completion.Suggestions)
	}
}