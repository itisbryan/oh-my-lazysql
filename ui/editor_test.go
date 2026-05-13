package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jorgerojas26/lazysql/models"
)

func TestEditorCtrlRExecutesQuery(t *testing.T) {
	driver := &fakeEditorDriver{results: [][]string{{"id"}, {"1"}}, rowCount: 1}
	results := NewResultsModel()
	model := NewEditorModel()
	model.driver = driver
	model.results = results
	model.lines = []string{"select 1"}
	model.cursorRow = 0
	model.cursorCol = 0
	model.mode = normalMode

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	if cmd == nil {
		t.Fatal("expected ctrl+r to execute query")
	}
	msg := cmd()
	executed, ok := msg.(editorQueryExecutedMsg)
	if !ok {
		t.Fatalf("expected editorQueryExecutedMsg, got %T", msg)
	}

	if driver.executedSQL != "select 1" {
		t.Fatalf("expected ctrl+r to execute query, got %q", driver.executedSQL)
	}
	if executed.rowCount != 1 || len(executed.results) != 2 || executed.err != nil {
		t.Fatalf("expected query completion message with results, got %#v", executed)
	}
}

func TestEditorEnterDoesNotExecuteQuery(t *testing.T) {
	driver := &fakeEditorDriver{results: [][]string{{"id"}, {"1"}}, rowCount: 1}
	model := NewEditorModel()
	model.driver = driver
	model.lines = []string{"select 1"}
	model.cursorCol = 7
	model.mode = normalMode

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if driver.executedSQL != "" {
		t.Fatalf("expected enter not to execute query, got %q", driver.executedSQL)
	}
}

func TestEditorNormalModeHIssueMovesCursor(t *testing.T) {
	model := NewEditorModel()
	model.lines = []string{"hello world"}
	model.cursorRow = 0
	model.cursorCol = 5
	model.mode = normalMode

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if model.cursorCol != 4 {
		t.Fatalf("expected h to move cursor left, got col %d", model.cursorCol)
	}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if model.cursorCol != 5 {
		t.Fatalf("expected l to move cursor right, got col %d", model.cursorCol)
	}
}

func TestEditorNormalModeJKMovesCursor(t *testing.T) {
	model := NewEditorModel()
	model.lines = []string{"line 1", "line 2", "line 3"}
	model.cursorRow = 1
	model.cursorCol = 0
	model.mode = normalMode

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if model.cursorRow != 2 {
		t.Fatalf("expected j to move cursor down, got row %d", model.cursorRow)
	}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if model.cursorRow != 1 {
		t.Fatalf("expected k to move cursor up, got row %d", model.cursorRow)
	}
}

func TestEditorInsertModeTypesText(t *testing.T) {
	model := NewEditorModel()
	model.lines = []string{""}
	model.cursorRow = 0
	model.cursorCol = 0
	model.mode = normalMode

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	if model.mode != insertMode {
		t.Fatal("expected i to enter insert mode")
	}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	if model.lines[0] != "hi" {
		t.Fatalf("expected typing to produce 'hi', got %q", model.lines[0])
	}
}

func TestEditorEscReturnsToNormalMode(t *testing.T) {
	model := NewEditorModel()
	model.lines = []string{"hello"}
	model.cursorCol = 3
	model.mode = insertMode

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if model.mode != normalMode {
		t.Fatal("expected Esc to return to normal mode")
	}
	if model.cursorCol != 2 {
		t.Fatalf("expected Esc to move cursor back one, got col %d", model.cursorCol)
	}
}

func TestEditorTabAcceptsSQLAutocomplete(t *testing.T) {
	model := NewEditorModel()
	model.lines = []string{"sel"}
	model.cursorRow = 0
	model.cursorCol = 3
	model.mode = insertMode
	model.completion.Update(model.text())

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})

	if model.lines[0] != "SELECT" {
		t.Fatalf("expected tab to accept SELECT suggestion, got %q", model.lines[0])
	}
}

func TestEditorViewShowsGhostTextWhenSuggestionsAvailable(t *testing.T) {
	model := NewEditorModel()
	model.width = 80
	model.height = 12
	model.lines = []string{"sel"}
	model.cursorRow = 0
	model.cursorCol = 3
	model.mode = insertMode
	model.completion.Update(model.text())

	view := model.View()
	if !strings.Contains(view, "-- INSERT --") {
		t.Fatalf("expected editor view to show INSERT mode indicator\n%s", view)
	}
}

func TestEditorViewShowsNormalModeIndicator(t *testing.T) {
	model := NewEditorModel()
	model.width = 80
	model.height = 12
	model.lines = []string{"select 1"}
	model.cursorRow = 0
	model.cursorCol = 0
	model.mode = normalMode

	view := model.View()
	if !strings.Contains(view, "-- NORMAL --") {
		t.Fatalf("expected editor view to show NORMAL mode indicator\n%s", view)
	}
}

func TestEditorDdDeletesLine(t *testing.T) {
	model := NewEditorModel()
	model.lines = []string{"line 1", "line 2", "line 3"}
	model.cursorRow = 1
	model.cursorCol = 0
	model.mode = normalMode

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	if len(model.lines) != 2 {
		t.Fatalf("expected dd to delete line, got %d lines", len(model.lines))
	}
	if model.lines[0] != "line 1" || model.lines[1] != "line 3" {
		t.Fatalf("expected remaining lines to be 'line 1' and 'line 3', got %v", model.lines)
	}
}

func TestEditorGGoesToTop(t *testing.T) {
	model := NewEditorModel()
	model.lines = []string{"line 1", "line 2", "line 3"}
	model.cursorRow = 2
	model.cursorCol = 3

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})

	if model.cursorRow != 0 || model.cursorCol != 0 {
		t.Fatalf("expected gg to go to top, got row %d col %d", model.cursorRow, model.cursorCol)
	}
}

func TestEditorOInsertsLineBelow(t *testing.T) {
	model := NewEditorModel()
	model.lines = []string{"line 1", "line 3"}
	model.cursorRow = 0
	model.cursorCol = 0
	model.mode = normalMode

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})

	if model.mode != insertMode {
		t.Fatal("expected o to enter insert mode")
	}
	if model.cursorRow != 1 {
		t.Fatalf("expected cursor on new line below, got row %d", model.cursorRow)
	}
	if model.cursorCol != 0 {
		t.Fatalf("expected cursor at col 0, got col %d", model.cursorCol)
	}
}

func TestEditorEscDismissesCompletionBeforeMode(t *testing.T) {
	model := NewEditorModel()
	model.lines = []string{"sel"}
	model.cursorRow = 0
	model.cursorCol = 3
	model.mode = insertMode
	model.completion.Update(model.text())

	if !model.completion.Visible {
		t.Fatal("expected completion visible")
	}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if model.completion.Visible {
		t.Fatal("expected first Esc to dismiss completion")
	}
	if model.mode != insertMode {
		t.Fatal("expected first Esc not to change mode")
	}
}

type fakeEditorDriver struct {
	results     [][]string
	rowCount    int
	executedSQL string
}

func (d *fakeEditorDriver) Connect(string) error                               { return nil }
func (d *fakeEditorDriver) TestConnection(string) error                        { return nil }
func (d *fakeEditorDriver) GetDatabases() ([]string, error)                    { return nil, nil }
func (d *fakeEditorDriver) GetTables(string) (map[string][]string, error)      { return nil, nil }
func (d *fakeEditorDriver) GetTableColumns(string, string) ([][]string, error) { return nil, nil }
func (d *fakeEditorDriver) GetConstraints(string, string) ([][]string, error)  { return nil, nil }
func (d *fakeEditorDriver) GetForeignKeys(string, string) ([][]string, error)  { return nil, nil }
func (d *fakeEditorDriver) GetIndexes(string, string) ([][]string, error)      { return nil, nil }
func (d *fakeEditorDriver) GetRecords(string, string, string, string, int, int) ([][]string, int, string, error) {
	return nil, 0, "", nil
}
func (d *fakeEditorDriver) UpdateRecord(string, string, string, string, string, string) error {
	return nil
}
func (d *fakeEditorDriver) DeleteRecord(string, string, string, string) error { return nil }
func (d *fakeEditorDriver) ExecuteDMLStatement(string) (string, error)        { return "", nil }
func (d *fakeEditorDriver) ExecuteQuery(sql string) ([][]string, int, error) {
	d.executedSQL = sql
	return d.results, d.rowCount, nil
}
func (d *fakeEditorDriver) ExecutePendingChanges([]models.DBDMLChange) error { return nil }
func (d *fakeEditorDriver) GetProvider() string                              { return "test" }
func (d *fakeEditorDriver) GetPrimaryKeyColumnNames(string, string) ([]string, error) {
	return nil, nil
}
func (d *fakeEditorDriver) SupportsProgramming() bool                                 { return false }
func (d *fakeEditorDriver) UseSchemas() bool                                          { return true }
func (d *fakeEditorDriver) GetFunctions(string) (map[string][]string, error)          { return nil, nil }
func (d *fakeEditorDriver) GetProcedures(string) (map[string][]string, error)         { return nil, nil }
func (d *fakeEditorDriver) GetViews(string) (map[string][]string, error)              { return nil, nil }
func (d *fakeEditorDriver) GetFunctionDefinition(string, string) (string, error)      { return "", nil }
func (d *fakeEditorDriver) GetProcedureDefinition(string, string) (string, error)     { return "", nil }
func (d *fakeEditorDriver) GetViewDefinition(string, string) (string, error)          { return "", nil }
func (d *fakeEditorDriver) FormatArg(any, models.CellValueType) any                   { return nil }
func (d *fakeEditorDriver) FormatArgForQueryString(any) string                        { return "" }
func (d *fakeEditorDriver) FormatReference(reference string) string                   { return reference }
func (d *fakeEditorDriver) FormatPlaceholder(int) string                              { return "" }
func (d *fakeEditorDriver) DMLChangeToQueryString(models.DBDMLChange) (string, error) { return "", nil }
func (d *fakeEditorDriver) SetProvider(string)                                        {}
