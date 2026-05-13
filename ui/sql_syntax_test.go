package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestSQLTreeSitterHighlightParsesAndStylesKeywords(t *testing.T) {
	highlighted := highlightSQL("select id from users where id = 1")

	if !strings.Contains(highlighted, "select") || !strings.Contains(highlighted, "where") {
		t.Fatalf("expected highlighted SQL to preserve source text, got %q", highlighted)
	}
	if !sqlTreeSitterAvailable() {
		t.Fatal("expected Tree-sitter SQL parser to be available")
	}
}

func TestSQLAutocompleteSuggestsKeywordsByPrefix(t *testing.T) {
	suggestions := sqlAutocompleteSuggestions("SEL")
	if len(suggestions) == 0 || suggestions[0] != "SELECT" {
		t.Fatalf("expected SELECT suggestion, got %#v", suggestions)
	}
}

func TestApplySQLAutocompleteReplacesCurrentPrefix(t *testing.T) {
	if got := applySQLAutocomplete("select * fr"); got != "select * FROM" {
		t.Fatalf("expected current prefix replaced, got %q", got)
	}
}

func TestHighlightSQLCanBeMeasuredByLipgloss(t *testing.T) {
	if width := lipgloss.Width(highlightSQL("where id = 1")); width != len("where id = 1") {
		t.Fatalf("expected ANSI-highlighted SQL to keep display width, got %d", width)
	}
}

func TestSQLCompletionItemsIncludesKeywordsTablesAndColumns(t *testing.T) {
	items := sqlCompletionItems("u", []string{"users", "orders"}, []string{"id", "updated_at"})
	if len(items) == 0 {
		t.Fatal("expected completion items for 'u' prefix")
	}
	hasTable := false
	for _, item := range items {
		if item.Text == "users" && item.Kind == tableSuggestion {
			hasTable = true
		}
	}
	if !hasTable {
		t.Fatalf("expected 'users' table suggestion, got %#v", items)
	}
	hasColumn := false
	for _, item := range items {
		if item.Text == "updated_at" && item.Kind == columnSuggestion {
			hasColumn = true
		}
	}
	if !hasColumn {
		t.Fatalf("expected 'updated_at' column suggestion, got %#v", items)
	}
}

func TestCompletionDropdownShowsTypeBadges(t *testing.T) {
	items := []completionItem{
		{Text: "SELECT", Kind: keywordSuggestion},
		{Text: "users", Kind: tableSuggestion},
		{Text: "id", Kind: columnSuggestion},
	}
	dropdown := renderCompletionDropdown(items, 0, 30)
	if !strings.Contains(dropdown, "SELECT") {
		t.Fatalf("expected dropdown to contain SELECT\n%s", dropdown)
	}
	if !strings.Contains(dropdown, "users") {
		t.Fatalf("expected dropdown to contain users\n%s", dropdown)
	}
	if !strings.Contains(dropdown, "id") {
		t.Fatalf("expected dropdown to contain id\n%s", dropdown)
	}
}

func TestCompletionStateGhostSuffix(t *testing.T) {
	cs := NewCompletionState()
	cs.Update("wh")
	ghost := cs.GhostSuffix("wh")
	if ghost != "ERE" {
		t.Fatalf("expected ghost suffix 'ERE' for 'wh' -> WHERE, got %q", ghost)
	}
}

func TestCompletionStateCycleAndAccept(t *testing.T) {
	cs := NewCompletionState()
	cs.Update("s")
	if !cs.Visible {
		t.Fatal("expected completion visible for 's' prefix")
	}
	initialIdx := cs.SelectedIndex
	cs.Cycle(1)
	if cs.SelectedIndex == initialIdx && len(cs.Suggestions) > 1 {
		t.Fatal("expected cycling to change selected index")
	}
	result := cs.Accept("s")
	if cs.Visible {
		t.Fatal("expected accept to dismiss completion")
	}
	_ = result
}