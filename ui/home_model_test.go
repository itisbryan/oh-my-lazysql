package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/itisbryan/oh-my-lazysql/models"
)

func TestLoadTableRecordsDefaultsToLatestCreatedAt(t *testing.T) {
	driver := &fakeHomeDriver{
		columns: [][]string{
			{"column_name", "data_type"},
			{"id", "integer"},
			{"created_at", "timestamp"},
		},
		records: [][]string{
			{"id", "created_at"},
			{"2", "2025-05-07"},
		},
	}
	model := &HomeModel{driver: driver, results: NewResultsModel()}

	msg := model.loadTableRecords("app", "public", "users")()
	if loaded, ok := msg.(recordsLoadedMsg); !ok || loaded.err != nil {
		t.Fatalf("expected recordsLoadedMsg without error, got %#v", msg)
	}

	if driver.sort != "created_at DESC" {
		t.Fatalf("expected default sort by latest created_at, got %q", driver.sort)
	}
}

func TestResultsNextPageLoadsRecordsWithPageOffset(t *testing.T) {
	driver := &fakeHomeDriver{
		columns: [][]string{
			{"column_name", "data_type"},
			{"id", "integer"},
		},
		records: [][]string{
			{"id"},
			{"101"},
		},
		totalRows: 250,
	}
	results := NewResultsModel()
	results.totalRows = 250
	model := &HomeModel{
		driver:          driver,
		results:         results,
		focus:           "results",
		currentDatabase: "app",
		currentSchema:   "public",
		currentTable:    "users",
	}

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	if cmd == nil {
		t.Fatal("expected next page to trigger a records reload")
	}
	msg := cmd()
	if loaded, ok := msg.(recordsLoadedMsg); !ok || loaded.err != nil {
		t.Fatalf("expected recordsLoadedMsg without error, got %#v", msg)
	}

	if driver.offset != 100 {
		t.Fatalf("expected next page to load offset 100, got %d", driver.offset)
	}
}

func TestResultsNextPageKeyLoadsRecordsWithPageOffset(t *testing.T) {
	driver := &fakeHomeDriver{
		columns:   [][]string{{"column_name", "data_type"}, {"id", "integer"}},
		records:   [][]string{{"id"}, {"101"}},
		totalRows: 250,
	}
	results := NewResultsModel()
	results.totalRows = 250
	model := &HomeModel{
		driver:          driver,
		results:         results,
		focus:           "results",
		currentDatabase: "app",
		currentSchema:   "public",
		currentTable:    "users",
	}

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'>'}})
	if cmd == nil {
		t.Fatal("expected > to trigger a records reload")
	}
	msg := cmd()
	if loaded, ok := msg.(recordsLoadedMsg); !ok || loaded.err != nil {
		t.Fatalf("expected recordsLoadedMsg without error, got %#v", msg)
	}

	if driver.offset != 100 {
		t.Fatalf("expected next page to load offset 100, got %d", driver.offset)
	}
}

func TestCtrlPSwitchesBackToConnectionList(t *testing.T) {
	model := NewHomeModel(models.Connection{Name: "app"})

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	if cmd == nil {
		t.Fatal("expected ctrl+p to switch to connection list")
	}
	msg := cmd()
	screenMsg, ok := msg.(ScreenChangeMsg)
	if !ok {
		t.Fatalf("expected ScreenChangeMsg, got %#v", msg)
	}
	if screenMsg.Screen != ScreenConnectionList {
		t.Fatalf("expected connection list screen, got %v", screenMsg.Screen)
	}
}

func TestResultsPreviousPageUsesLeftChevronKey(t *testing.T) {
	driver := &fakeHomeDriver{
		columns:   [][]string{{"column_name", "data_type"}, {"id", "integer"}},
		records:   [][]string{{"id"}, {"1"}},
		totalRows: 250,
	}
	results := NewResultsModel()
	results.totalRows = 250
	results.page = 1
	model := &HomeModel{
		driver:          driver,
		results:         results,
		focus:           "results",
		currentDatabase: "app",
		currentSchema:   "public",
		currentTable:    "users",
	}

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'<'}})
	if cmd == nil {
		t.Fatal("expected < to trigger previous page reload")
	}
	msg := cmd()
	if loaded, ok := msg.(recordsLoadedMsg); !ok || loaded.err != nil {
		t.Fatalf("expected recordsLoadedMsg without error, got %#v", msg)
	}
	if driver.offset != 0 {
		t.Fatalf("expected previous page to load offset 0, got %d", driver.offset)
	}
}

func TestResultsCannotAdvancePastLastPage(t *testing.T) {
	results := NewResultsModel()
	results.totalRows = 250
	results.page = 2
	model := &HomeModel{
		driver:          &fakeHomeDriver{},
		results:         results,
		focus:           "results",
		currentDatabase: "app",
		currentSchema:   "public",
		currentTable:    "users",
	}

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	if cmd != nil {
		t.Fatal("expected no reload past the final page")
	}
	if model.results.page != 2 {
		t.Fatalf("expected page to stay on final page 2, got %d", model.results.page)
	}
}

func TestWhereFilterReloadsCurrentTable(t *testing.T) {
	driver := &fakeHomeDriver{
		columns: [][]string{{"column_name", "data_type"}, {"id", "integer"}},
		records: [][]string{{"id"}, {"2"}},
	}
	results := NewResultsModel()
	results.whereFilter = "id > 1"
	results.page = 3
	model := &HomeModel{
		driver:          driver,
		results:         results,
		focus:           "results",
		currentDatabase: "app",
		currentSchema:   "public",
		currentTable:    "users",
	}

	_, cmd := model.Update(whereFilterAppliedMsg{where: "id > 1"})
	if cmd == nil {
		t.Fatal("expected filter apply to reload records")
	}
	msg := cmd()
	if loaded, ok := msg.(recordsLoadedMsg); !ok || loaded.err != nil {
		t.Fatalf("expected recordsLoadedMsg without error, got %#v", msg)
	}
	if driver.where != "WHERE id > 1" {
		t.Fatalf("expected driver to receive where filter, got %q", driver.where)
	}
	if model.results.page != 0 || model.results.row != 0 {
		t.Fatalf("expected page and row reset, got page=%d row=%d", model.results.page, model.results.row)
	}
}

func TestNormalizeWhereFilterPreservesExplicitWhere(t *testing.T) {
	if got := normalizeWhereFilter("WHERE id = '2350'"); got != "WHERE id = '2350'" {
		t.Fatalf("unexpected normalized filter: %q", got)
	}
	if got := normalizeWhereFilter("  id = '2350'  "); got != "WHERE id = '2350'" {
		t.Fatalf("unexpected normalized predicate: %q", got)
	}
}

func TestBackslashTogglesSidebarAndMovesFocusToResults(t *testing.T) {
	model := NewHomeModel(models.Connection{})
	model.focus = "tree"
	model.tree.focused = true

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\\'}})
	if cmd != nil {
		t.Fatal("expected sidebar toggle not to emit command")
	}
	if model.showSidebar {
		t.Fatal("expected sidebar to be hidden")
	}
	if model.focus != "results" || model.tree.focused {
		t.Fatalf("expected focus moved to results, got focus=%q treeFocused=%v", model.focus, model.tree.focused)
	}
}

func TestEnterOnSelectedTableMovesFocusToResults(t *testing.T) {
	driver := &fakeHomeDriver{
		columns: [][]string{{"column_name", "data_type"}, {"id", "integer"}},
		records: [][]string{{"id"}, {"1"}},
	}
	model := NewHomeModel(models.Connection{})
	model.driver = driver
	model.tree = NewTreeModel()
	model.tree.focused = true
	model.tree.flattened = []FlatNode{{
		Type:     NodeTypeTable,
		Database: "app",
		Schema:   "public",
		Name:     "users",
	}}
	model.focus = "tree"

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected selected table to load records")
	}

	if model.focus != "results" || model.tree.focused {
		t.Fatalf("expected focus moved to results, got focus=%q treeFocused=%v", model.focus, model.tree.focused)
	}
}

func TestBackslashTogglesSidebarBackOn(t *testing.T) {
	model := NewHomeModel(models.Connection{})
	model.showSidebar = false
	model.focus = "results"

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\\'}})
	if !model.showSidebar {
		t.Fatal("expected sidebar to be shown")
	}
}

func TestCtrlBTogglesSidebarAsReliableAlias(t *testing.T) {
	model := NewHomeModel(models.Connection{})
	model.focus = "tree"

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlB})
	if model.showSidebar {
		t.Fatal("expected sidebar to be hidden with ctrl+b")
	}
}

func TestCtrlETogglesEditorFromTreeAndRestoresTreeFocus(t *testing.T) {
	model := NewHomeModel(models.Connection{})
	model.focus = "tree"
	model.tree.focused = true

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlE})
	if !model.showEditor || model.focus != "editor" || !model.editor.focused {
		t.Fatalf("expected ctrl+e from tree to open and focus editor, showEditor=%v focus=%q editorFocused=%v", model.showEditor, model.focus, model.editor.focused)
	}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlE})
	if model.showEditor || model.focus != "tree" || !model.tree.focused {
		t.Fatalf("expected ctrl+e to close editor and restore tree focus, showEditor=%v focus=%q treeFocused=%v", model.showEditor, model.focus, model.tree.focused)
	}
}

func TestCtrlETogglesEditorFromResultsAndRestoresResultsFocus(t *testing.T) {
	model := NewHomeModel(models.Connection{})
	model.focus = "results"

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlE})
	if !model.showEditor || model.focus != "editor" || !model.editor.focused {
		t.Fatalf("expected ctrl+e from results to open and focus editor, showEditor=%v focus=%q editorFocused=%v", model.showEditor, model.focus, model.editor.focused)
	}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlE})
	if model.showEditor || model.focus != "results" || model.tree.focused {
		t.Fatalf("expected ctrl+e to close editor and restore results focus, showEditor=%v focus=%q treeFocused=%v", model.showEditor, model.focus, model.tree.focused)
	}
}

func TestEscInEditorNormalModeDoesNotCloseEditor(t *testing.T) {
	model := NewHomeModel(models.Connection{})
	model.showEditor = true
	model.focus = "editor"
	model.editor.focused = true
	model.lastFocus = "results"
	model.editor.mode = normalMode

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd != nil {
		t.Fatal("expected esc in editor normal mode not to emit command")
	}
	if !model.showEditor || model.focus != "editor" {
		t.Fatalf("expected esc not to close focused editor, showEditor=%v focus=%q", model.showEditor, model.focus)
	}
}

func TestEscInFocusedEditorInsertModeReturnsToNormalMode(t *testing.T) {
	model := NewHomeModel(models.Connection{})
	model.showEditor = true
	model.focus = "editor"
	model.editor.focused = true
	model.lastFocus = "results"
	model.editor.mode = insertMode

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd != nil {
		t.Fatal("expected esc in editor insert mode not to emit command")
	}
	if !model.showEditor || model.focus != "editor" {
		t.Fatalf("expected editor to remain open and focused, showEditor=%v focus=%q", model.showEditor, model.focus)
	}
	if model.editor.mode != normalMode {
		t.Fatalf("expected editor esc to switch to normal mode, got %v", model.editor.mode)
	}
}

func TestCtrlRExecutesQueryWhenEditorFocused(t *testing.T) {
	driver := &fakeHomeDriver{queryResults: [][]string{{"id"}, {"1"}}, queryRowCount: 1}
	model := NewHomeModel(models.Connection{})
	model.driver = driver
	model.editor.driver = driver
	model.focus = "editor"
	model.showEditor = true
	model.editor.lines = []string{"select 1"}
	model.editor.cursorRow = 0
	model.editor.cursorCol = 0

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	if cmd == nil {
		t.Fatal("expected ctrl+r to reach editor and execute query")
	}
	msg := cmd()
	if msg == nil {
		t.Fatal("expected query execution command to return completion message")
	}
	_, _ = model.Update(msg)

	if driver.executedSQL != "select 1" {
		t.Fatalf("expected query execution through editor, got %q", driver.executedSQL)
	}
	if model.editor.executing {
		t.Fatal("expected editor executing state to be cleared after query completion")
	}
	if model.results.status != "Got 1 rows" {
		t.Fatalf("expected results status to reflect query completion, got %q", model.results.status)
	}
}

func TestPropagationPutsTableNamesInCompletionContext(t *testing.T) {
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

	model.results.columns = []GridColumn{{Title: "id"}, {Title: "email"}, {Title: "created_at"}}
	model.results.rows = [][]string{{"1", "a@b.com", "2026-01-01"}}

	model.propagateCompletionContext()

	tables := model.results.completion.TableNames
	if len(tables) == 0 {
		t.Fatalf("expected table names in completion context, got empty. tree root children: %d", len(model.tree.root.Children))
	}
	found := false
	for _, tn := range tables {
		if tn == "users" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected 'users' in table names, got %v", tables)
	}

	cols := model.results.completion.ColumnNames
	foundCol := false
	for _, c := range cols {
		if c == "email" {
			foundCol = true
		}
	}
	if !foundCol {
		t.Fatalf("expected 'email' in column names, got %v", cols)
	}

	model.results.filterEditing = true
	model.results.filterInput = "u"
	model.results.completion.Update(model.results.filterInput)

	if !model.results.completion.Visible {
		t.Fatalf("expected completion visible for 'u' prefix, got Visible=%v, Suggestions=%v", model.results.completion.Visible, model.results.completion.Suggestions)
	}

	hasUsers := false
	for _, s := range model.results.completion.Suggestions {
		if s.Text == "users" && s.Kind == tableSuggestion {
			hasUsers = true
		}
	}
	if !hasUsers {
		t.Fatalf("expected 'users' table suggestion for 'u' prefix, got %v", model.results.completion.Suggestions)
	}
}

func TestTableNameCollectionWithNonSchemaDriver(t *testing.T) {
	model := NewHomeModel(models.Connection{})
	model.driver = &fakeHomeDriverNoSchemas{}
	model.tree.driver = &fakeHomeDriverNoSchemas{}

	model.tree.SetDatabases([]string{"mydb"})
	model.tree.root.Children[0].Expanded = true
	model.tree.setTables(model.tree.root.Children[0], "mydb", map[string][]string{"mydb": {"users", "orders", "products"}}, nil, nil)
	model.tree.rebuildFlattened()

	names := model.tree.TableNames()
	has := func(n string) bool {
		for _, x := range names {
			if x == n {
				return true
			}
		}
		return false
	}
	if !has("users") {
		t.Fatalf("expected 'users' in TableNames for non-schema driver, got %v", names)
	}
	if !has("orders") {
		t.Fatalf("expected 'orders' in TableNames for non-schema driver, got %v", names)
	}
	if has("mydb") {
		t.Fatalf("did not expect database name 'mydb' in TableNames, got %v", names)
	}
}

type fakeHomeDriverNoSchemas struct {
	fakeHomeDriver
}

func (d *fakeHomeDriverNoSchemas) UseSchemas() bool { return false }

func TestTabAcceptsAutocompleteWhenEditorFocused(t *testing.T) {
	driver := &fakeHomeDriver{}
	results := NewResultsModel()
	results.columns = []GridColumn{{Title: "id", IsPK: true}, {Title: "email"}}
	results.rows = [][]string{{"1", "new@example.com"}}
	results.pendingEdits = map[cellPosition]pendingEdit{
		{row: 0, col: 1}: {original: "old@example.com", value: "new@example.com"},
	}
	model := &HomeModel{
		driver:          driver,
		results:         results,
		focus:           "results",
		currentDatabase: "app",
		currentSchema:   "public",
		currentTable:    "users",
	}

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	if cmd == nil {
		t.Fatal("expected save command")
	}
	msg := cmd()
	if saved, ok := msg.(pendingChangesSavedMsg); !ok || saved.err != nil {
		t.Fatalf("expected pendingChangesSavedMsg without error, got %#v", msg)
	}
	if len(driver.pendingChanges) != 1 {
		t.Fatalf("expected 1 pending change, got %d", len(driver.pendingChanges))
	}
	change := driver.pendingChanges[0]
	if change.Database != "app" || change.Table != "public.users" || change.Type != models.DMLUpdateType {
		t.Fatalf("unexpected change target/type: %#v", change)
	}
	if len(change.PrimaryKeyInfo) != 1 || change.PrimaryKeyInfo[0].Name != "id" || change.PrimaryKeyInfo[0].Value != "1" {
		t.Fatalf("unexpected primary key info: %#v", change.PrimaryKeyInfo)
	}
	if len(change.Values) != 1 || change.Values[0].Column != "email" || change.Values[0].Value != "new@example.com" {
		t.Fatalf("unexpected values: %#v", change.Values)
	}
}

func TestEnterOnViewNodeLoadsRecords(t *testing.T) {
	driver := &fakeHomeDriver{
		columns: [][]string{{"column_name", "data_type"}, {"id", "integer"}, {"email", "text"}},
		records: [][]string{{"id", "email"}, {"1", "a@b.com"}},
	}
	model := NewHomeModel(models.Connection{})
	model.driver = driver
	model.tree.driver = driver
	model.tree.root.Children = []*TreeNode{{
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
			Children: []*TreeNode{{
				Type:     NodeTypeSection,
				Name:     "Views",
				Database: "app",
				Schema:   "public",
				Expanded: true,
				Children: []*TreeNode{{
					Type:     NodeTypeView,
					Name:     "active_users",
					Database: "app",
					Schema:   "public",
				}},
			}},
		}},
	}}
	model.tree.rebuildFlattened()
	model.focus = "tree"

	found := false
	for i, node := range model.tree.visibleNodes() {
		if node.Type == NodeTypeView && node.Name == "active_users" {
			model.tree.cursor = i
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected view node to be visible in tree")
	}

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected enter on view node to load records")
	}
	msg := cmd()
	if loaded, ok := msg.(recordsLoadedMsg); !ok || loaded.err != nil {
		t.Fatalf("expected recordsLoadedMsg without error, got %#v", msg)
	}
	if model.currentTable != "active_users" {
		t.Fatalf("expected current table active_users, got %q", model.currentTable)
	}
}

func TestEnterOnForeignKeyCellNavigatesToReferencedTable(t *testing.T) {
	driver := &fakeHomeDriver{
		records: [][]string{{"id", "name"}, {"42", "Ada"}},
	}
	results := NewResultsModel()
	results.columns = []GridColumn{
		{Title: "id", Type: "integer"},
		{Title: "user_id", Type: "integer", IsFK: true, ForeignKey: &ForeignKeyRef{Schema: "public", Table: "users", Column: "id"}},
	}
	results.rows = [][]string{{"1", "42"}}
	results.col = 1
	model := &HomeModel{
		driver:          driver,
		results:         results,
		focus:           "results",
		currentDatabase: "app",
		currentSchema:   "public",
		currentTable:    "orders",
	}

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected foreign key enter to load referenced table")
	}
	msg := cmd()
	if loaded, ok := msg.(recordsLoadedMsg); !ok || loaded.err != nil {
		t.Fatalf("expected recordsLoadedMsg without error, got %#v", msg)
	}
	if model.currentTable != "users" {
		t.Fatalf("expected current table users, got %q", model.currentTable)
	}
	if model.results.whereFilter != "id = '42'" || driver.where != "WHERE id = '42'" {
		t.Fatalf("expected FK where filter, model=%q driver=%q", model.results.whereFilter, driver.where)
	}
	if len(model.navStack) != 1 {
		t.Fatalf("expected navStack length 1 after FK navigation, got %d", len(model.navStack))
	}

	_, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	if cmd == nil {
		t.Fatal("expected [ to navigate back")
	}
	msg = cmd()
	if loaded, ok := msg.(recordsLoadedMsg); !ok || loaded.err != nil {
		t.Fatalf("expected recordsLoadedMsg when going back, got %#v", msg)
	}
	if model.currentTable != "orders" {
		t.Fatalf("expected back to orders table, got %q", model.currentTable)
	}
	if len(model.navStack) != 0 {
		t.Fatalf("expected navStack empty after navigating back, got %d", len(model.navStack))
	}
}

func TestBuildPendingChangesIncludesDeletesAndTypedValues(t *testing.T) {
	results := NewResultsModel()
	results.columns = []GridColumn{{Title: "id", Type: "integer", IsPK: true}, {Title: "active", Type: "boolean"}, {Title: "score", Type: "numeric"}, {Title: "deleted_at", Type: "timestamp"}}
	results.rows = [][]string{{"1", "false", "0", ""}}
	results.pendingEdits = map[cellPosition]pendingEdit{
		{row: 0, col: 1}: {original: "false", value: "true"},
		{row: 0, col: 2}: {original: "0", value: "42.5"},
		{row: 0, col: 3}: {original: "", value: "NULL"},
	}
	results.pendingDeletes = map[int]bool{0: true}
	model := &HomeModel{
		driver:          &fakeHomeDriver{},
		results:         results,
		currentDatabase: "app",
		currentSchema:   "public",
		currentTable:    "users",
	}

	changes := model.buildPendingChanges()
	if len(changes) != 4 {
		t.Fatalf("expected update and delete changes, got %#v", changes)
	}
	foundBool := false
	foundFloat := false
	foundNull := false
	foundDelete := false
	for _, change := range changes {
		if change.Type == models.DMLDeleteType {
			foundDelete = true
			continue
		}
		if change.Type != models.DMLUpdateType || len(change.Values) != 1 {
			t.Fatalf("unexpected change: %#v", change)
		}
		switch change.Values[0].Column {
		case "active":
			foundBool = change.Values[0].Value == true
		case "score":
			foundFloat = change.Values[0].Value == 42.5
		case "deleted_at":
			foundNull = change.Values[0].Type == models.Null
		}
	}
	if !foundBool || !foundFloat || !foundNull || !foundDelete {
		t.Fatalf("expected bool, float, null, and delete changes, got %#v", changes)
	}
}

func TestBuildPendingChangesIncludesInsertRow(t *testing.T) {
	results := NewResultsModel()
	results.columns = []GridColumn{{Title: "id", IsPK: true}, {Title: "email"}, {Title: "created_at"}}
	results.rows = [][]string{{"1", "ada@example.com", "2026-01-01"}}
	results.insertingRow = true
	results.insertRow = []string{"2", "grace@example.com", "DEFAULT"}
	model := &HomeModel{
		driver:          &fakeHomeDriver{},
		results:         results,
		currentDatabase: "app",
		currentSchema:   "public",
		currentTable:    "users",
	}

	changes := model.buildPendingChanges()
	if len(changes) != 1 {
		t.Fatalf("expected one insert change, got %#v", changes)
	}
	change := changes[0]
	if change.Type != models.DMLInsertType || change.Table != "public.users" {
		t.Fatalf("unexpected insert change: %#v", change)
	}
	if len(change.Values) != 3 || change.Values[2].Column != "created_at" || change.Values[2].Type != models.Default {
		t.Fatalf("expected DEFAULT typed created_at value, got %#v", change.Values)
	}
}

func TestSortAppliedReloadsCurrentTableWithSelectedColumnSort(t *testing.T) {
	driver := &fakeHomeDriver{
		columns: [][]string{{"column_name", "data_type"}, {"id", "integer"}, {"email", "text"}},
		records: [][]string{{"id", "email"}, {"1", "ada@example.com"}},
	}
	results := NewResultsModel()
	results.columns = []GridColumn{{Title: "id"}, {Title: "email"}}
	results.sortCol = 1
	results.sortDir = "DESC"
	model := &HomeModel{
		driver:          driver,
		results:         results,
		focus:           "results",
		currentDatabase: "app",
		currentSchema:   "public",
		currentTable:    "users",
	}

	_, cmd := model.Update(sortAppliedMsg{})
	if cmd == nil {
		t.Fatal("expected sort apply to reload records")
	}
	msg := cmd()
	if loaded, ok := msg.(recordsLoadedMsg); !ok || loaded.err != nil {
		t.Fatalf("expected recordsLoadedMsg without error, got %#v", msg)
	}
	if driver.sort != "email DESC" {
		t.Fatalf("expected driver sort by selected column, got %q", driver.sort)
	}
}

func TestHomeModelRoutesCtrlXKeyToResultsDeletion(t *testing.T) {
	model := NewHomeModel(models.Connection{})
	model.focus = "results"
	model.results.columns = []GridColumn{{Title: "id", IsPK: true}, {Title: "email"}}
	model.results.rows = [][]string{{"1", "ada@example.com"}}
	model.results.row = 0
	model.results.activeTab = 0

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlX})

	if !model.results.pendingDeletes[0] {
		t.Fatalf("expected HomeModel to route ctrl+x to results deletion, got %#v", model.results.pendingDeletes)
	}
}

func TestHomeModelDoesNotRouteDOrDDToResultsDeletion(t *testing.T) {
	model := NewHomeModel(models.Connection{})
	model.focus = "results"
	model.results.columns = []GridColumn{{Title: "id", IsPK: true}, {Title: "email"}}
	model.results.rows = [][]string{{"1", "ada@example.com"}}
	model.results.row = 0
	model.results.activeTab = 0

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if model.results.pendingDeletes[0] {
		t.Fatalf("did not expect d to mark row for deletion, got %#v", model.results.pendingDeletes)
	}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d', 'd'}})
	if model.results.pendingDeletes[0] {
		t.Fatalf("did not expect dd to mark row for deletion, got %#v", model.results.pendingDeletes)
	}
}

type fakeHomeDriver struct {
	columns        [][]string
	records        [][]string
	sort           string
	where          string
	offset         int
	totalRows      int
	queryResults   [][]string
	queryRowCount  int
	executedSQL    string
	pendingChanges []models.DBDMLChange
}

func (d *fakeHomeDriver) Connect(string) error                               { return nil }
func (d *fakeHomeDriver) TestConnection(string) error                        { return nil }
func (d *fakeHomeDriver) GetDatabases() ([]string, error)                    { return nil, nil }
func (d *fakeHomeDriver) GetTables(string) (map[string][]string, error)      { return nil, nil }
func (d *fakeHomeDriver) GetTableColumns(string, string) ([][]string, error) { return d.columns, nil }
func (d *fakeHomeDriver) GetConstraints(string, string) ([][]string, error)  { return nil, nil }
func (d *fakeHomeDriver) GetForeignKeys(string, string) ([][]string, error)  { return nil, nil }
func (d *fakeHomeDriver) GetIndexes(string, string) ([][]string, error)      { return nil, nil }
func (d *fakeHomeDriver) GetRecords(_, _, where, sort string, offset, _ int) ([][]string, int, string, error) {
	d.where = where
	d.sort = sort
	d.offset = offset
	totalRows := d.totalRows
	if totalRows == 0 {
		totalRows = len(d.records) - 1
	}
	return d.records, totalRows, "select", nil
}
func (d *fakeHomeDriver) UpdateRecord(string, string, string, string, string, string) error {
	return nil
}
func (d *fakeHomeDriver) DeleteRecord(string, string, string, string) error { return nil }
func (d *fakeHomeDriver) ExecuteDMLStatement(string) (string, error)        { return "", nil }
func (d *fakeHomeDriver) ExecuteQuery(sql string) ([][]string, int, error) {
	d.executedSQL = sql
	return d.queryResults, d.queryRowCount, nil
}
func (d *fakeHomeDriver) ExecutePendingChanges(changes []models.DBDMLChange) error {
	d.pendingChanges = changes
	return nil
}
func (d *fakeHomeDriver) GetProvider() string                                       { return "test" }
func (d *fakeHomeDriver) GetPrimaryKeyColumnNames(string, string) ([]string, error) { return nil, nil }
func (d *fakeHomeDriver) SupportsProgramming() bool                                 { return false }
func (d *fakeHomeDriver) UseSchemas() bool                                          { return true }
func (d *fakeHomeDriver) GetFunctions(string) (map[string][]string, error)          { return nil, nil }
func (d *fakeHomeDriver) GetProcedures(string) (map[string][]string, error)         { return nil, nil }
func (d *fakeHomeDriver) GetViews(string) (map[string][]string, error)              { return nil, nil }
func (d *fakeHomeDriver) GetMaterializedViews(string) (map[string][]string, error)  { return nil, nil }
func (d *fakeHomeDriver) GetFunctionDefinition(string, string) (string, error)      { return "", nil }
func (d *fakeHomeDriver) GetProcedureDefinition(string, string) (string, error)     { return "", nil }
func (d *fakeHomeDriver) GetViewDefinition(string, string) (string, error)          { return "", nil }
func (d *fakeHomeDriver) FormatArg(any, models.CellValueType) any                   { return nil }
func (d *fakeHomeDriver) FormatArgForQueryString(arg any) string {
	if value, ok := arg.(string); ok {
		return "'" + value + "'"
	}
	return "''"
}
func (d *fakeHomeDriver) FormatReference(reference string) string                   { return reference }
func (d *fakeHomeDriver) FormatPlaceholder(int) string                              { return "" }
func (d *fakeHomeDriver) DMLChangeToQueryString(models.DBDMLChange) (string, error) { return "", nil }
func (d *fakeHomeDriver) SetProvider(string)                                        {}
