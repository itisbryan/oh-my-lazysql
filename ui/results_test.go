package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestResultsViewRendersDatabaseClientChrome(t *testing.T) {
	model := NewResultsModel()
	model.width = 100
	model.height = 24
	model.columns = []GridColumn{{Title: "id"}, {Title: "email"}, {Title: "created_at"}}
	model.rows = [][]string{{"2", "ada@example.com", "2026-05-07"}, {"1", "NULL&", "2026-05-06"}}
	model.totalRows = 2
	model.row = 1
	model.col = 1

	view := model.View()

	for _, expected := range []string{
		"RECORDS",
		"WHERE",
		"filter rows by SQL predicate",
		"/ Edit",
		"R2:C2",
		"▸ email",
		"┼",
		"▌NULL",
		"▐",
	} {
		if !strings.Contains(view, expected) {
			t.Fatalf("expected view to contain %q\n%s", expected, view)
		}
	}
}

func TestResultsFilterEditingRendersApplyAndCancelHints(t *testing.T) {
	model := NewResultsModel()
	model.width = 100
	model.height = 24
	model.columns = []GridColumn{{Title: "id"}}
	model.rows = [][]string{{"1"}}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})

	view := model.View()
	for _, expected := range []string{"Enter Apply", "Esc Cancel", "filter rows by SQL predicate"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("expected editing filter view to contain %q\n%s", expected, view)
		}
	}
}

func TestResultsFilterBarHasSpaciousHeightAndGutters(t *testing.T) {
	model := NewResultsModel()
	model.width = 100
	model.height = 24
	model.columns = []GridColumn{{Title: "id"}}
	model.rows = [][]string{{"1"}}

	filter := model.renderFilterBar()
	if got := strings.Count(filter, "\n") + 1; got < 3 {
		t.Fatalf("expected filter bar to be at least 3 lines tall, got %d\n%s", got, filter)
	}
	if !strings.Contains(filter, "WHERE      filter rows by SQL predicate") {
		t.Fatalf("expected wider gutter between WHERE and filter input\n%s", filter)
	}
}

func TestResultsTabsRenderNavigationHint(t *testing.T) {
	model := NewResultsModel()
	model.width = 100

	tabs := model.renderTabs()
	for _, expected := range []string{"[/] tabs", "1-5 jump"} {
		if !strings.Contains(tabs, expected) {
			t.Fatalf("expected tabs to include navigation hint %q\n%s", expected, tabs)
		}
	}
}

func TestResultsTabsRenderNumberlessActiveTab(t *testing.T) {
	model := NewResultsModel()
	model.width = 100
	model.activeTab = 2

	tabs := model.renderTabs()
	if !strings.Contains(tabs, "▸ Constraints") {
		t.Fatalf("expected tabs to visibly mark active tab\n%s", tabs)
	}
	if strings.Contains(tabs, "1 Records") || strings.Contains(tabs, "2 Columns") || strings.Contains(tabs, "3 Constraints") {
		t.Fatalf("expected tab labels without numeric prefixes\n%s", tabs)
	}
}

func TestResultsTabsRenderDividersBetweenLabels(t *testing.T) {
	model := NewResultsModel()
	model.width = 100

	tabs := model.renderTabs()
	if !strings.Contains(tabs, "Records │ Columns") || !strings.Contains(tabs, "Columns │ Constraints") {
		t.Fatalf("expected dividers between tabs\n%s", tabs)
	}
}

func TestResultsViewRendersTabsBelowFilterToolbar(t *testing.T) {
	model := NewResultsModel()
	model.width = 100
	model.height = 24
	model.columns = []GridColumn{{Title: "id"}}
	model.rows = [][]string{{"1"}}

	view := model.View()
	whereIndex := strings.Index(view, "WHERE")
	tabIndex := strings.Index(view, "▸ Records")
	if whereIndex == -1 || tabIndex == -1 {
		t.Fatalf("expected view to contain filter and tabs\n%s", view)
	}
	if tabIndex < whereIndex {
		t.Fatalf("expected tabs below filter toolbar so they stay visible\n%s", view)
	}
}

func TestResultsFilterApplyEmitsMessage(t *testing.T) {
	model := NewResultsModel()
	model.width = 100
	model.height = 24
	model.columns = []GridColumn{{Title: "id"}}
	model.rows = [][]string{{"1"}}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'>'}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("expected filter apply command")
	}
	msg := cmd()
	applied, ok := msg.(whereFilterAppliedMsg)
	if !ok {
		t.Fatalf("expected whereFilterAppliedMsg, got %T", msg)
	}
	if applied.where != "id > 1" || model.whereFilter != "id > 1" || model.filterEditing {
		t.Fatalf("unexpected filter state: msg=%q active=%q editing=%v", applied.where, model.whereFilter, model.filterEditing)
	}
}

func TestResultsFilterTabAcceptsSQLAutocomplete(t *testing.T) {
	model := NewResultsModel()
	model.filterEditing = true
	model.filterInput = "wh"
	model.completion.Update(model.filterInput)

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})

	if model.filterInput != "WHERE" {
		t.Fatalf("expected tab to accept WHERE suggestion, got %q", model.filterInput)
	}
}

func TestResultsFilterShowsSQLAutocompleteDropdown(t *testing.T) {
	model := NewResultsModel()
	model.width = 100
	model.filterEditing = true
	model.filterInput = "wh"
	model.completion.Update(model.filterInput)

	filter := model.renderFilterBar()
	if !strings.Contains(filter, "WHERE") {
		t.Fatalf("expected filter bar dropdown to show WHERE suggestion\n%s", filter)
	}
}

func TestResultsInlineEditCommitsPendingChange(t *testing.T) {
	model := NewResultsModel()
	model.width = 100
	model.height = 24
	model.columns = []GridColumn{{Title: "id"}, {Title: "email"}}
	model.rows = [][]string{{"1", "old@example.com"}}
	model.col = 1

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if model.rows[0][1] != "new" {
		t.Fatalf("expected cell to be updated locally, got %q", model.rows[0][1])
	}
	if model.pendingChangeCount() != 1 {
		t.Fatalf("expected 1 pending change, got %d", model.pendingChangeCount())
	}
	if model.editingCell {
		t.Fatal("expected editing to stop after commit")
	}
}

func TestResultsDeleteMarksRowAsPendingChange(t *testing.T) {
	model := NewResultsModel()
	model.columns = []GridColumn{{Title: "id"}, {Title: "email"}}
	model.rows = [][]string{{"1", "ada@example.com"}}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlX})

	if !model.pendingDeletes[0] {
		t.Fatalf("expected row 0 to be marked for deletion, got %#v", model.pendingDeletes)
	}
	if !strings.Contains(model.status, "marked for deletion") {
		t.Fatalf("expected delete status feedback, got %q", model.status)
	}
	if model.pendingChangeCount() != 1 {
		t.Fatalf("expected 1 pending change, got %d", model.pendingChangeCount())
	}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	if model.pendingDeletes[0] {
		t.Fatal("expected u to unmark pending deletion")
	}
	if !strings.Contains(model.status, "restored") {
		t.Fatalf("expected restore status feedback, got %q", model.status)
	}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlX})
	if !model.pendingDeletes[0] {
		t.Fatal("expected repeated ctrl+x to leave row marked for deletion")
	}
}

func TestResultsPlainXAndDDoNotMarkDeletion(t *testing.T) {
	model := NewResultsModel()
	model.columns = []GridColumn{{Title: "id"}, {Title: "email"}}
	model.rows = [][]string{{"1", "ada@example.com"}}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if model.pendingDeletes[0] {
		t.Fatalf("did not expect plain x to mark row for deletion, got %#v", model.pendingDeletes)
	}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if model.pendingDeletes[0] {
		t.Fatalf("did not expect d to mark row for deletion, got %#v", model.pendingDeletes)
	}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d', 'd'}})
	if model.pendingDeletes[0] {
		t.Fatalf("did not expect dd to mark row for deletion, got %#v", model.pendingDeletes)
	}
}

func TestResultsDeletedRowRendersVisibleMarker(t *testing.T) {
	model := NewResultsModel()
	model.width = 100
	model.height = 24
	model.columns = []GridColumn{{Title: "id"}, {Title: "email"}}
	model.rows = [][]string{{"1", "ada@example.com"}}
	model.pendingDeletes = map[int]bool{0: true}

	view := model.View()
	if !strings.Contains(view, "✗1") {
		t.Fatalf("expected deleted row marker in view\n%s", view)
	}
	if strings.Contains(view, "[38;2;") || strings.Contains(view, "[9m") {
		t.Fatalf("expected deleted row view not to expose raw ANSI styling\n%s", view)
	}
}

func TestResultsInsertRowCreatesEditablePendingInsert(t *testing.T) {
	model := NewResultsModel()
	model.columns = []GridColumn{{Title: "id"}, {Title: "email"}}
	model.rows = [][]string{{"1", "ada@example.com"}}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})

	if !model.insertingRow {
		t.Fatal("expected insert mode to remain pending")
	}
	if model.row != len(model.rows) {
		t.Fatalf("expected selection to stay on synthetic insert row, got row=%d", model.row)
	}
	if got := model.insertRow[0]; got != "2" {
		t.Fatalf("expected first inserted value to be committed, got %q", got)
	}
	if got := model.insertRow[1]; got != "b" {
		t.Fatalf("expected second inserted value to be live-updated, got %q", got)
	}
	if model.pendingChangeCount() != 1 {
		t.Fatalf("expected pending insert to count as a change, got %d", model.pendingChangeCount())
	}
}

func TestResultsSortKeyCyclesColumnSortAndEmitsMessage(t *testing.T) {
	model := NewResultsModel()
	model.columns = []GridColumn{{Title: "id"}, {Title: "email"}}
	model.rows = [][]string{{"1", "ada@example.com"}}
	model.col = 1

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if cmd == nil {
		t.Fatal("expected sort key to emit reload message")
	}
	msg := cmd()
	if _, ok := msg.(sortAppliedMsg); !ok {
		t.Fatalf("expected sortAppliedMsg, got %T", msg)
	}
	if model.sortCol != 1 || model.sortDir != "ASC" {
		t.Fatalf("expected email ASC sort, got col=%d dir=%q", model.sortCol, model.sortDir)
	}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if model.sortCol != 1 || model.sortDir != "DESC" {
		t.Fatalf("expected email DESC sort, got col=%d dir=%q", model.sortCol, model.sortDir)
	}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if model.sortCol != -1 || model.sortDir != "" {
		t.Fatalf("expected sort cleared, got col=%d dir=%q", model.sortCol, model.sortDir)
	}
}

func TestResultsInlineEditEscCancelsChange(t *testing.T) {
	model := NewResultsModel()
	model.width = 100
	model.height = 24
	model.columns = []GridColumn{{Title: "email"}}
	model.rows = [][]string{{"old@example.com"}}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if model.rows[0][0] != "old@example.com" {
		t.Fatalf("expected cell to remain unchanged, got %q", model.rows[0][0])
	}
	if model.pendingChangeCount() != 0 {
		t.Fatalf("expected no pending changes, got %d", model.pendingChangeCount())
	}
	if model.editingCell {
		t.Fatal("expected editing to stop after cancel")
	}
}

func TestResultsCtrlDScrollsDownFiveRows(t *testing.T) {
	model := NewResultsModel()
	model.height = 24
	model.columns = []GridColumn{{Title: "id"}}
	model.rows = make([][]string, 20)
	for i := range model.rows {
		model.rows[i] = []string{"1"}
	}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlD})

	if model.row != 5 {
		t.Fatalf("expected ctrl+d to move down 5 rows, got %d", model.row)
	}
}

func TestResultsCtrlUScrollsUpFiveRows(t *testing.T) {
	model := NewResultsModel()
	model.height = 24
	model.columns = []GridColumn{{Title: "id"}}
	model.rows = make([][]string, 20)
	for i := range model.rows {
		model.rows[i] = []string{"1"}
	}
	model.row = 12

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlU})

	if model.row != 7 {
		t.Fatalf("expected ctrl+u to move up 5 rows, got %d", model.row)
	}
}

func TestResultsEIteratesColumnsWithWraparound(t *testing.T) {
	model := NewResultsModel()
	model.columns = []GridColumn{{Title: "id"}, {Title: "email"}, {Title: "created_at"}}
	model.rows = [][]string{{"1", "ada@example.com", "2026-05-08"}}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if model.col != 1 {
		t.Fatalf("expected e to move to next column, got %d", model.col)
	}

	model.col = 2
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if model.col != 0 {
		t.Fatalf("expected e to wrap to first column, got %d", model.col)
	}
}

func TestResultsFilterTabCyclesCompletions(t *testing.T) {
	model := NewResultsModel()
	model.width = 100
	model.height = 24
	model.columns = []GridColumn{{Title: "id"}}
	model.filterEditing = true
	model.filterInput = "i"
	model.completion.Update(model.filterInput)

	count := len(model.completion.Suggestions)
	if count < 2 {
		t.Fatalf("expected at least 2 suggestions for 'i' prefix, got %d", count)
	}

	expected := model.completion.Suggestions[model.completion.SelectedIndex].Text
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	if model.filterInput != expected {
		t.Fatalf("expected tab to accept %q, got %q", expected, model.filterInput)
	}
}

func TestResultsFilterUpDownNavigatesCompletions(t *testing.T) {
	model := NewResultsModel()
	model.width = 100
	model.height = 24
	model.columns = []GridColumn{{Title: "id"}}
	model.filterEditing = true
	model.filterInput = "i"
	model.completion.Update(model.filterInput)

	if len(model.completion.Suggestions) < 2 {
		t.Fatalf("expected at least 2 suggestions for 'i' prefix, got %d", len(model.completion.Suggestions))
	}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	if model.completion.SelectedIndex != 1 {
		t.Fatalf("expected down to move to index 1, got %d", model.completion.SelectedIndex)
	}
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	if model.completion.SelectedIndex != 0 {
		t.Fatalf("expected up to move back to index 0, got %d", model.completion.SelectedIndex)
	}
}

func TestResultsFilterEscDismissesCompletion(t *testing.T) {
	model := NewResultsModel()
	model.width = 100
	model.height = 24
	model.columns = []GridColumn{{Title: "id"}}
	model.filterEditing = true
	model.filterInput = "wh"
	model.completion.Update(model.filterInput)

	if !model.completion.Visible {
		t.Fatal("expected completion to be visible after typing")
	}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if model.completion.Visible {
		t.Fatal("expected Esc to dismiss completion")
	}
	if !model.filterEditing {
		t.Fatal("expected first Esc to dismiss completion, not exit filter editing")
	}
}

func TestCompletionDropdownRenders(t *testing.T) {
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
}

func TestEditingCellUsesReadableForeground(t *testing.T) {
	if editingCellForeground() != PrimaryTextColor {
		t.Fatalf("expected editing cell foreground to be readable primary text, got %s", editingCellForeground())
	}
}

func TestRowDetailOpensForCurrentRow(t *testing.T) {
	model := NewResultsModel()
	model.width = 100
	model.height = 24
	model.columns = []GridColumn{{Title: "id", Type: "integer", IsPK: true}, {Title: "email", Type: "text"}}
	model.rows = [][]string{{"1", "ada@example.com"}}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})

	if !model.showRowDetail {
		t.Fatal("expected row detail to be open")
	}
	view := model.View()
	for _, expected := range []string{"Row 1", "2 fields", "FIELD", "TYPE", "VALUE", "▌", "id", "int", "PK", "1", "email", "ada@example.com", "j/k field", "Enter edit", "Esc close", "Ctrl+R save"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("expected row detail to contain %q\n%s", expected, view)
		}
	}
}

func TestRowDetailEscCloses(t *testing.T) {
	model := NewResultsModel()
	model.width = 100
	model.height = 24
	model.columns = []GridColumn{{Title: "id"}}
	model.rows = [][]string{{"1"}}
	model.showRowDetail = true

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if model.showRowDetail {
		t.Fatal("expected row detail to close on Esc")
	}
}

func TestRowDetailMovesSelectedField(t *testing.T) {
	model := NewResultsModel()
	model.columns = []GridColumn{{Title: "id"}, {Title: "email"}}
	model.rows = [][]string{{"1", "ada@example.com"}}
	model.showRowDetail = true

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if model.detailRow != 1 {
		t.Fatalf("expected detail row 1 after j, got %d", model.detailRow)
	}
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if model.detailRow != 0 {
		t.Fatalf("expected detail row 0 after k, got %d", model.detailRow)
	}
}

func TestRowDetailEditsSelectedField(t *testing.T) {
	model := NewResultsModel()
	model.columns = []GridColumn{{Title: "id"}, {Title: "email"}}
	model.rows = [][]string{{"1", "old@example.com"}}
	model.showRowDetail = true
	model.detailRow = 1

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if model.rows[0][1] != "new" {
		t.Fatalf("expected detail edit to update row, got %q", model.rows[0][1])
	}
	if model.pendingChangeCount() != 1 {
		t.Fatalf("expected pending detail edit, got %d", model.pendingChangeCount())
	}
	if model.detailEditing {
		t.Fatal("expected detail editing to stop after commit")
	}
}

func TestRowDetailEditEscCancelsFieldEdit(t *testing.T) {
	model := NewResultsModel()
	model.columns = []GridColumn{{Title: "email"}}
	model.rows = [][]string{{"old@example.com"}}
	model.showRowDetail = true

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if model.rows[0][0] != "old@example.com" {
		t.Fatalf("expected row value unchanged after cancel, got %q", model.rows[0][0])
	}
	if !model.showRowDetail {
		t.Fatal("expected Esc while editing to cancel edit, not close detail")
	}
	if model.detailEditing {
		t.Fatal("expected detail editing to stop after cancel")
	}
}

func TestHeaderColorsCreateActiveGradient(t *testing.T) {
	activeForeground, activeBackground, activeSeparator := headerColorsForDistance(0)
	neighborForeground, neighborBackground, neighborSeparator := headerColorsForDistance(1)
	inactiveForeground, inactiveBackground, inactiveSeparator := headerColorsForDistance(2)

	if activeForeground != SecondaryTextColor || activeBackground != "#283457" || activeSeparator != SecondaryTextColor {
		t.Fatalf("unexpected active header colors: %s %s %s", activeForeground, activeBackground, activeSeparator)
	}
	if neighborForeground != "#BB9AF7" || neighborBackground != "#1F2335" || neighborSeparator != "#7DCFFF" {
		t.Fatalf("unexpected neighbor header colors: %s %s %s", neighborForeground, neighborBackground, neighborSeparator)
	}
	if inactiveForeground != PrimaryTextColor || inactiveBackground != "#1A1F33" || inactiveSeparator != "#565F89" {
		t.Fatalf("unexpected inactive header colors: %s %s %s", inactiveForeground, inactiveBackground, inactiveSeparator)
	}
}
