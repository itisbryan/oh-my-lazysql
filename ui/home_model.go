package ui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jorgerojas26/lazysql/app"
	"github.com/jorgerojas26/lazysql/drivers"
	"github.com/jorgerojas26/lazysql/helpers/logger"
	"github.com/jorgerojas26/lazysql/models"
)

type HomeModel struct {
	connection  models.Connection
	driver      drivers.Driver
	tree        *TreeModel
	editor      *EditorModel
	results     *ResultsModel
	width       int
	height      int
	focus       string
	lastFocus   string
	showEditor  bool
	showSidebar bool

	currentDatabase string
	currentSchema   string
	currentTable    string

	navStack       []navigationEntry
	loadingRecords bool
}

type navigationEntry struct {
	database  string
	schema    string
	table     string
	where     string
	row       int
	col       int
	totalRows int
	page      int
}

type recordsLoadedMsg struct {
	columns      []string
	colTypes     map[string]string
	nullableCols map[string]bool
	enumValues   map[string][]string
	fkRefs       map[string]ForeignKeyRef
	pkCols       map[string]bool
	fkCols       map[string]bool
	rows         [][]string
	totalRows    int
	query        string
	err          error
}

type metadataLoadedMsg struct {
	tab  int
	data [][]string
	err  error
}

type pendingChangesSavedMsg struct {
	err error
}

func NewHomeModel(data any) *HomeModel {
	conn, ok := data.(models.Connection)
	if !ok {
		conn = models.Connection{}
	}

	logger.Info("NewHomeModel received", map[string]any{
		"provider": conn.Provider,
		"name":     conn.Name,
		"dbName":   conn.DBName,
		"url":      conn.URL,
	})

	var driver drivers.Driver
	switch conn.Provider {
	case "MySQL", "mysql":
		driver = &drivers.MySQL{}
	case "PostgreSQL", "postgres":
		driver = &drivers.Postgres{}
	case "SQLite", "sqlite3":
		driver = &drivers.SQLite{}
	case "MSSQL", "sqlserver":
		driver = &drivers.MSSQL{}
	default:
		driver = &drivers.MySQL{}
	}

	if conn.URL != "" {
		logger.Info("Connecting driver", map[string]any{"url": conn.URL})
		if err := driver.Connect(conn.URL); err != nil {
			logger.Error("Failed to connect", map[string]any{"error": err})
		} else {
			logger.Info("Connected successfully", nil)
		}
	} else {
		logger.Error("No URL provided for connection", map[string]any{"name": conn.Name})
	}

	home := &HomeModel{
		connection:  conn,
		driver:      driver,
		tree:        NewTreeModel(),
		editor:      NewEditorModel(),
		results:     NewResultsModel(),
		focus:       "tree",
		showSidebar: true,
	}

	home.tree.driver = driver
	home.tree.dbName = conn.DBName
	home.editor.driver = driver
	home.editor.results = home.results

	return home
}

func (m *HomeModel) Init() tea.Cmd {
	return m.tree.Init()
}

func (m *HomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if msg.String() == "ctrl+p" {
			return m, func() tea.Msg {
				return ScreenChangeMsg{Screen: ScreenConnectionList, Data: nil}
			}
		}
		if msg.String() == "[" {
			if cmd := m.navigateBack(); cmd != nil {
				return m, cmd
			}
		}
		if m.focus == "results" && m.results.filterEditing {
			m.propagateCompletionContext()
			_, resultsCmd := m.results.Update(msg)
			return m, resultsCmd
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "ctrl+e":
			if m.showEditor && m.focus == "editor" {
				m.showEditor = false
				if m.lastFocus == "" {
					m.lastFocus = "results"
				}
				m.setFocus(m.lastFocus)
				return m, nil
			}
			if m.focus != "editor" {
				m.lastFocus = m.focus
			}
			m.showEditor = true
			m.setFocus("editor")
			m.propagateCompletionContext()
			return m, nil
		case "\\", "ctrl+b":
			m.showSidebar = !m.showSidebar
			if !m.showSidebar && m.focus == "tree" {
				m.setFocus("results")
			} else {
				m.tree.focused = m.showSidebar && m.focus == "tree"
			}
			return m, nil
		case "ctrl+r":
			if m.focus == "editor" {
				break
			}
			if m.focus == "results" && m.results.pendingChangeCount() > 0 {
				return m, m.savePendingChanges()
			}
			return m, nil
		case "esc":
			if m.showEditor && m.focus == "editor" {
				break
			}
		case "ctrl+h":
			m.focusLeft()
			return m, nil
		case "ctrl+l":
			m.focusRight()
			return m, nil
		case "tab":
			if m.focus == "editor" {
				break
			}
			m.focusRight()
			return m, nil
		case "shift+tab":
			m.focusLeft()
			return m, nil
		case "enter":
			if m.focus == "tree" {
				nodeType, database, schema, name := m.tree.SelectedNode()
				if nodeType == NodeTypeTable {
					m.openTable(database, schema, name, "")
					m.setFocus("results")
					m.loadingRecords = true
					m.results.loading = true
					return m, m.loadTableRecords(database, schema, name)
				}
			}
			if m.focus == "results" {
				if cmd := m.navigateForeignKeyCell(); cmd != nil {
					return m, cmd
				}
			}
		case "ctrl+n", ">":
			if m.focus == "results" {
				return m, m.loadResultsPage(m.results.page + 1)
			}
		case "<":
			if m.focus == "results" {
				return m, m.loadResultsPage(m.results.page - 1)
			}
		}
	}

	if msg, ok := msg.(recordsLoadedMsg); ok {
		m.loadingRecords = false
		m.results.loading = false
		if msg.err != nil {
			m.results.SetStatus("Error: " + msg.err.Error())
		} else {
			m.results.SetData(msg.columns, msg.colTypes, msg.nullableCols, msg.enumValues, msg.fkRefs, msg.pkCols, msg.fkCols, msg.rows)
			m.results.totalRows = msg.totalRows
			m.results.SetStatus(fmt.Sprintf("%d rows | %s", msg.totalRows, msg.query))
			m.propagateCompletionContext()
			return m, m.loadMetadata()
		}
		return m, nil
	}

	if msg, ok := msg.(editorQueryExecutedMsg); ok {
		m.editor.executing = false
		if msg.err != nil {
			m.results.SetStatus("Error: " + msg.err.Error())
			return m, nil
		}
		if len(msg.results) > 0 {
			columns := msg.results[0]
			rows := msg.results[1:]
			m.results.SetData(columns, nil, nil, nil, nil, nil, nil, rows)
			m.results.SetStatus(fmt.Sprintf("Got %d rows", msg.rowCount))
		}
		return m, nil
	}

	if msg, ok := msg.(metadataLoadedMsg); ok {
		if msg.err == nil && msg.tab >= 0 && msg.tab < len(m.results.metadata) {
			m.results.metadata[msg.tab] = msg.data
		}
		return m, nil
	}

	if msg, ok := msg.(pendingChangesSavedMsg); ok {
		if msg.err != nil {
			m.results.SetStatus("Error saving changes: " + msg.err.Error())
		} else {
			m.results.pendingEdits = map[cellPosition]pendingEdit{}
			m.results.pendingDeletes = map[int]bool{}
			if m.results.insertingRow && m.results.insertRow != nil {
				m.results.insertingRow = false
				m.results.insertRow = nil
			}
			m.results.SetStatus("Changes saved")
		}
		return m, nil
	}

	if _, ok := msg.(whereFilterAppliedMsg); ok {
		m.results.page = 0
		m.results.row = 0
		m.loadingRecords = true
		m.results.loading = true
		return m, m.loadCurrentTableRecords()
	}

	if _, ok := msg.(sortAppliedMsg); ok {
		m.results.page = 0
		m.results.row = 0
		m.loadingRecords = true
		m.results.loading = true
		return m, m.loadCurrentTableRecords()
	}

	var cmds []tea.Cmd
	_, treeCmd := m.tree.Update(msg)
	cmds = append(cmds, treeCmd)
	m.propagateCompletionContext()
	if m.showEditor && m.focus == "editor" {
		_, editorCmd := m.editor.Update(msg)
		cmds = append(cmds, editorCmd)
	}
	if m.focus == "results" {
		_, resultsCmd := m.results.Update(msg)
		cmds = append(cmds, resultsCmd)
	}
	return m, tea.Batch(cmds...)
}

func (m *HomeModel) canLoadResultsPage(page int) bool {
	if m.currentTable == "" || page < 0 {
		return false
	}
	if m.results.totalRows == 0 {
		return page == 0
	}
	return page*m.results.pageSize < m.results.totalRows
}

func (m *HomeModel) loadCurrentTableRecords() tea.Cmd {
	return m.loadTableRecords(m.currentDatabase, m.currentSchema, m.currentTable)
}

func (m *HomeModel) loadResultsPage(page int) tea.Cmd {
	if !m.canLoadResultsPage(page) {
		return nil
	}
	m.results.page = page
	m.results.row = 0
	m.loadingRecords = true
	m.results.loading = true
	return m.loadCurrentTableRecords()
}

func (m *HomeModel) openTable(database, schema, table, where string) {
	m.currentDatabase = database
	m.currentSchema = schema
	m.currentTable = table
	m.results.page = 0
	m.results.row = 0
	m.results.col = 0
	m.results.colOffset = 0
	m.results.sortCol = -1
	m.results.sortDir = ""
	m.results.filterEditing = false
	m.results.filterInput = where
	m.results.whereFilter = where
}

func (m *HomeModel) navigateBack() tea.Cmd {
	if len(m.navStack) == 0 {
		return nil
	}

	entry := m.navStack[len(m.navStack)-1]
	m.navStack = m.navStack[:len(m.navStack)-1]

	m.currentDatabase = entry.database
	m.currentSchema = entry.schema
	m.currentTable = entry.table
	m.results.whereFilter = entry.where
	m.results.filterInput = entry.where
	m.results.filterEditing = false
	m.results.page = 0
	m.results.row = 0
	m.results.col = 0
	m.results.colOffset = 0
	m.results.sortCol = -1
	m.results.sortDir = ""

	if len(m.navStack) == 0 {
		m.results.SetStatus(fmt.Sprintf("Returned to %s.%s", entry.schema, entry.table))
	} else {
		m.results.SetStatus(fmt.Sprintf("Back to %s.%s (%d more)", entry.schema, entry.table, len(m.navStack)))
	}

	m.loadingRecords = true
	m.results.loading = true
	return m.loadTableRecords(entry.database, entry.schema, entry.table)
}

func (m *HomeModel) pushNavigation() {
	entry := navigationEntry{
		database:  m.currentDatabase,
		schema:    m.currentSchema,
		table:     m.currentTable,
		where:     m.results.whereFilter,
		row:       m.results.row,
		col:       m.results.col,
		totalRows: m.results.totalRows,
		page:      m.results.page,
	}
	m.navStack = append(m.navStack, entry)
}

func (m *HomeModel) navigateForeignKeyCell() tea.Cmd {
	if m.results == nil || m.results.editingCell || m.results.activeTab != 0 {
		return nil
	}
	if m.results.row < 0 || m.results.row >= len(m.results.rows) || m.results.col < 0 || m.results.col >= len(m.results.columns) {
		return nil
	}
	col := m.results.columns[m.results.col]
	if col.ForeignKey == nil || col.ForeignKey.Table == "" || col.ForeignKey.Column == "" {
		return nil
	}
	value := ""
	if m.results.col < len(m.results.rows[m.results.row]) {
		value = m.results.rows[m.results.row][m.results.col]
	}
	where := foreignKeyWhereClause(col.ForeignKey.Column, value, m.driver)
	schema := col.ForeignKey.Schema
	if schema == "" {
		schema = m.currentSchema
	}

	m.pushNavigation()
	m.openTable(m.currentDatabase, schema, col.ForeignKey.Table, where)
	m.results.SetStatus(fmt.Sprintf("Following %s → %s.%s (%d back)", col.Title, col.ForeignKey.Table, col.ForeignKey.Column, len(m.navStack)))
	m.loadingRecords = true
	m.results.loading = true
	return m.loadTableRecords(m.currentDatabase, schema, col.ForeignKey.Table)
}

func foreignKeyWhereClause(column, value string, driver drivers.Driver) string {
	ref := column
	if driver != nil {
		ref = driver.FormatReference(column)
	}
	if strings.EqualFold(strings.TrimSpace(value), "NULL") || strings.EqualFold(strings.TrimSpace(value), "NULL&") {
		return fmt.Sprintf("%s IS NULL", ref)
	}
	formatted := value
	if driver != nil {
		formatted = driver.FormatArgForQueryString(value)
	}
	return fmt.Sprintf("%s = %s", ref, formatted)
}

func (m *HomeModel) savePendingChanges() tea.Cmd {
	return func() tea.Msg {
		if m.driver == nil {
			return pendingChangesSavedMsg{err: fmt.Errorf("no driver")}
		}
		changes := m.buildPendingChanges()
		if len(changes) == 0 {
			return pendingChangesSavedMsg{}
		}
		return pendingChangesSavedMsg{err: m.driver.ExecutePendingChanges(changes)}
	}
}

func (m *HomeModel) buildPendingChanges() []models.DBDMLChange {
	if m.currentTable == "" {
		return nil
	}
	tableName := m.currentTable
	if m.currentSchema != "" && m.driver != nil && m.driver.UseSchemas() && !strings.Contains(tableName, ".") {
		tableName = m.currentSchema + "." + tableName
	}

	changes := make([]models.DBDMLChange, 0, len(m.results.pendingEdits)+len(m.results.pendingDeletes)+1)

	for pos, edit := range m.results.pendingEdits {
		if pos.row < 0 || pos.row >= len(m.results.rows) || pos.col < 0 || pos.col >= len(m.results.columns) {
			continue
		}
		pkInfo := m.primaryKeyInfoForRow(pos.row)
		if len(pkInfo) == 0 {
			continue
		}
		value, cellType := dmlCellValue(edit.value, m.results.columns[pos.col].Type)
		changes = append(changes, models.DBDMLChange{
			Database:       m.currentDatabase,
			Table:          tableName,
			PrimaryKeyInfo: pkInfo,
			Values: []models.CellValue{{
				Column:           m.results.columns[pos.col].Title,
				Value:            value,
				TableColumnIndex: pos.col,
				TableRowIndex:    pos.row,
				Type:             cellType,
			}},
			Type: models.DMLUpdateType,
		})
	}

	for rowIdx := range m.results.pendingDeletes {
		if rowIdx < 0 || rowIdx >= len(m.results.rows) {
			continue
		}
		pkInfo := m.primaryKeyInfoForRow(rowIdx)
		if len(pkInfo) == 0 {
			continue
		}
		changes = append(changes, models.DBDMLChange{
			Database:       m.currentDatabase,
			Table:          tableName,
			PrimaryKeyInfo: pkInfo,
			Type:           models.DMLDeleteType,
		})
	}

	if m.results.insertingRow && m.results.insertRow != nil {
		values := make([]models.CellValue, 0, len(m.results.columns))
		for colIdx, col := range m.results.columns {
			value := ""
			if colIdx < len(m.results.insertRow) {
				value = m.results.insertRow[colIdx]
			}
			typedValue, cellType := dmlCellValue(value, col.Type)
			values = append(values, models.CellValue{
				Column:           col.Title,
				Value:            typedValue,
				TableColumnIndex: colIdx,
				TableRowIndex:    len(m.results.rows),
				Type:             cellType,
			})
		}
		changes = append(changes, models.DBDMLChange{
			Database: m.currentDatabase,
			Table:    tableName,
			Values:   values,
			Type:     models.DMLInsertType,
		})
	}

	return changes
}

func dmlCellValue(value, columnType string) (any, models.CellValueType) {
	switch value {
	case "NULL", "NULL&":
		return "NULL", models.Null
	case "DEFAULT":
		return value, models.Default
	case "":
		return value, models.Empty
	}

	lowerType := strings.ToLower(columnType)
	lowerValue := strings.ToLower(value)
	if strings.Contains(lowerType, "bool") || strings.Contains(lowerType, "tinyint(1)") || lowerType == "bit" {
		switch lowerValue {
		case "true", "t", "1":
			return true, models.String
		case "false", "f", "0":
			return false, models.String
		}
	}

	if isIntegerColumnType(lowerType) {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return parsed, models.String
		}
	}

	if isFloatColumnType(lowerType) {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed, models.String
		}
	}

	return value, models.String
}

func isIntegerColumnType(columnType string) bool {
	return strings.Contains(columnType, "int") || strings.Contains(columnType, "serial") || strings.Contains(columnType, "identity")
}

func isFloatColumnType(columnType string) bool {
	return strings.Contains(columnType, "numeric") ||
		strings.Contains(columnType, "decimal") ||
		strings.Contains(columnType, "real") ||
		strings.Contains(columnType, "double") ||
		strings.Contains(columnType, "float")
}

func (m *HomeModel) primaryKeyInfoForRow(rowIdx int) []models.PrimaryKeyInfo {
	if rowIdx < 0 || rowIdx >= len(m.results.rows) {
		return nil
	}
	row := m.results.rows[rowIdx]
	info := []models.PrimaryKeyInfo{}
	for colIdx, col := range m.results.columns {
		if !col.IsPK || colIdx >= len(row) {
			continue
		}
		info = append(info, models.PrimaryKeyInfo{Name: col.Title, Value: row[colIdx]})
	}
	return info
}

func (m *HomeModel) focusLeft() {
	if m.focus == "results" && m.showEditor {
		m.setFocus("editor")
		return
	}
	if !m.showSidebar {
		m.setFocus("results")
		return
	}
	m.setFocus("tree")
}

func (m *HomeModel) focusRight() {
	if m.focus == "tree" && m.showEditor && m.showSidebar {
		m.setFocus("editor")
		return
	}
	m.setFocus("results")
}

func (m *HomeModel) setFocus(focus string) {
	m.focus = focus
	m.tree.focused = focus == "tree"
	m.editor.focused = focus == "editor"
	m.propagateCompletionContext()
}

func (m *HomeModel) propagateCompletionContext() {
	tableNames := m.tree.TableNames()
	columnNames := make([]string, 0, len(m.results.columns))
	for _, col := range m.results.columns {
		columnNames = append(columnNames, col.Title)
	}
	m.results.completion.TableNames = tableNames
	m.results.completion.ColumnNames = columnNames
	m.editor.completion.TableNames = tableNames
	m.editor.completion.ColumnNames = columnNames
}

func (m *HomeModel) loadTableRecords(database, schema, table string) tea.Cmd {
	return func() tea.Msg {
		if m.driver == nil {
			return recordsLoadedMsg{err: fmt.Errorf("no driver")}
		}

		tableName := table
		if schema != "" && m.driver.UseSchemas() && !strings.Contains(table, ".") {
			tableName = schema + "." + table
		}

		offset := m.results.page * m.results.pageSize
		records, totalRows, query, err := m.driver.GetRecords(database, tableName, normalizeWhereFilter(m.results.whereFilter), m.defaultSort(database, tableName), offset, m.results.pageSize)
		if err != nil {
			return recordsLoadedMsg{err: err}
		}
		if len(records) == 0 {
			return recordsLoadedMsg{columns: []string{}, colTypes: map[string]string{}, nullableCols: map[string]bool{}, enumValues: map[string][]string{}, fkRefs: map[string]ForeignKeyRef{}, pkCols: map[string]bool{}, fkCols: map[string]bool{}, rows: [][]string{}, totalRows: totalRows, query: query}
		}

		colTypes := map[string]string{}
		nullableCols := map[string]bool{}
		enumValues := map[string][]string{}
		colData, colErr := m.driver.GetTableColumns(database, tableName)
		if colErr == nil && len(colData) > 0 {
			headerRow := colData[0]
			typeIdx := columnIndex(headerRow, "data_type", "type")
			nameIdx := columnIndex(headerRow, "column_name", "name", "field")
			nullableIdx := columnIndex(headerRow, "is_nullable", "null", "nullable")
			notNullIdx := columnIndex(headerRow, "notnull")
			enumIdx := columnIndex(headerRow, "enum_values")
			if nameIdx >= 0 {
				for _, row := range colData[1:] {
					if nameIdx >= len(row) {
						continue
					}
					name := row[nameIdx]
					typeName := ""
					if typeIdx >= 0 && typeIdx < len(row) {
						typeName = row[typeIdx]
						colTypes[name] = typeName
					}
					if enumIdx >= 0 && enumIdx < len(row) {
						enumValues[name] = parseEnumValues(row[enumIdx])
					} else if values := parseEnumType(typeName); len(values) > 0 {
						enumValues[name] = values
					}
					if nullableIdx >= 0 && nullableIdx < len(row) {
						nullableCols[name] = isNullableValue(row[nullableIdx])
					} else if notNullIdx >= 0 && notNullIdx < len(row) {
						nullableCols[name] = !isTruthyValue(row[notNullIdx])
					}
				}
			}
		}

		pkCols := map[string]bool{}
		pkNames, pkErr := m.driver.GetPrimaryKeyColumnNames(database, tableName)
		if pkErr == nil {
			for _, name := range pkNames {
				pkCols[name] = true
			}
		}

		fkCols := map[string]bool{}
		fkRefs := map[string]ForeignKeyRef{}
		fkData, fkErr := m.driver.GetForeignKeys(database, tableName)
		if fkErr == nil && len(fkData) > 0 {
			fkRefs = parseForeignKeyRefs(fkData)
			for colName := range fkRefs {
				fkCols[colName] = true
			}
		}

		return recordsLoadedMsg{
			columns:      records[0],
			colTypes:     colTypes,
			nullableCols: nullableCols,
			enumValues:   enumValues,
			fkRefs:       fkRefs,
			pkCols:       pkCols,
			fkCols:       fkCols,
			rows:         records[1:],
			totalRows:    totalRows,
			query:        query,
		}
	}
}

func parseForeignKeyRefs(data [][]string) map[string]ForeignKeyRef {
	refs := map[string]ForeignKeyRef{}
	if len(data) == 0 {
		return refs
	}
	header := data[0]
	columnIdx := columnIndex(header, "column_name", "COLUMN_NAME", "from")
	tableIdx := columnIndex(header, "foreign_table_name", "referenced_table", "REFERENCED_TABLE_NAME", "table")
	targetColumnIdx := columnIndex(header, "foreign_column_name", "referenced_column", "REFERENCED_COLUMN_NAME", "to")
	schemaIdx := columnIndex(header, "foreign_table_schema", "referenced_schema", "schema")
	if columnIdx < 0 || tableIdx < 0 || targetColumnIdx < 0 {
		return refs
	}
	for _, row := range data[1:] {
		if columnIdx >= len(row) || tableIdx >= len(row) || targetColumnIdx >= len(row) {
			continue
		}
		ref := ForeignKeyRef{Table: row[tableIdx], Column: row[targetColumnIdx]}
		if schemaIdx >= 0 && schemaIdx < len(row) {
			ref.Schema = row[schemaIdx]
		}
		refs[row[columnIdx]] = ref
	}
	return refs
}

func parseEnumValues(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			values = append(values, part)
		}
	}
	return values
}

func parseEnumType(typeName string) []string {
	trimmed := strings.TrimSpace(typeName)
	lower := strings.ToLower(trimmed)
	if !strings.HasPrefix(lower, "enum(") || !strings.HasSuffix(trimmed, ")") {
		return nil
	}

	body := trimmed[len("enum(") : len(trimmed)-1]
	values := []string{}
	var current strings.Builder
	inQuote := false
	escaped := false
	for _, r := range body {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case r == '\'':
			inQuote = !inQuote
		case r == ',' && !inQuote:
			values = append(values, current.String())
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}
	values = append(values, current.String())
	return values
}

func columnIndex(headers []string, names ...string) int {
	for i, header := range headers {
		for _, name := range names {
			if strings.EqualFold(header, name) {
				return i
			}
		}
	}
	return -1
}

func isNullableValue(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	return lower == "yes" || lower == "true" || lower == "1" || lower == "y"
}

func isTruthyValue(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	return lower == "yes" || lower == "true" || lower == "1" || lower == "y"
}

func normalizeWhereFilter(filter string) string {
	filter = strings.TrimSpace(filter)
	if filter == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(filter), "where ") {
		return filter
	}
	return "WHERE " + filter
}

func (m *HomeModel) defaultSort(database, tableName string) string {
	if m.results.sortCol >= 0 && m.results.sortCol < len(m.results.columns) && m.results.sortDir != "" {
		colName := m.results.columns[m.results.sortCol].Title
		dir := "ASC"
		if m.results.sortDir == "DESC" {
			dir = "DESC"
		}
		return m.driver.FormatReference(colName) + " " + dir
	}

	columns, err := m.driver.GetTableColumns(database, tableName)
	if err != nil {
		logger.Error("defaultSort: failed to load columns", map[string]any{"table": tableName, "error": err})
		return ""
	}
	if len(columns) < 2 {
		return ""
	}

	columnNames := map[string]bool{}
	for _, row := range columns[1:] {
		if len(row) > 0 {
			columnNames[strings.ToLower(row[0])] = true
		}
	}

	for _, candidate := range []string{"created_at", "updated_at", "inserted_at", "id"} {
		if columnNames[candidate] {
			return m.driver.FormatReference(candidate) + " DESC"
		}
	}

	return ""
}

func (m *HomeModel) loadMetadata() tea.Cmd {
	if m.driver == nil || m.currentTable == "" {
		return nil
	}

	tableName := m.currentTable
	if m.currentSchema != "" && m.driver.UseSchemas() && !strings.Contains(tableName, ".") {
		tableName = m.currentSchema + "." + tableName
	}
	database := m.currentDatabase

	type loader struct {
		tab int
		fn  func() ([][]string, error)
	}
	loaders := []loader{
		{1, func() ([][]string, error) { return m.driver.GetTableColumns(database, tableName) }},
		{2, func() ([][]string, error) { return m.driver.GetConstraints(database, tableName) }},
		{3, func() ([][]string, error) { return m.driver.GetForeignKeys(database, tableName) }},
		{4, func() ([][]string, error) { return m.driver.GetIndexes(database, tableName) }},
	}

	cmds := make([]tea.Cmd, len(loaders))
	for i, l := range loaders {
		tab, fn := l.tab, l.fn
		cmds[i] = func() tea.Msg {
			data, err := fn()
			return metadataLoadedMsg{tab: tab, data: data, err: err}
		}
	}
	return tea.Batch(cmds...)
}

func (m *HomeModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	treeWidth := 45
	if app.App != nil && app.App.Config() != nil && app.App.Config().TreeWidth > 0 {
		treeWidth = max(treeWidth, app.App.Config().TreeWidth)
	}
	if treeWidth > m.width/2 {
		treeWidth = m.width / 2
	}
	contentHeight := m.height - 2

	treePanel := ""
	availableWidth := m.width
	if m.showSidebar {
		m.tree.width = treeWidth
		m.tree.height = contentHeight
		m.tree.focused = m.focus == "tree"
		treePanel = m.tree.View()
		actualTreeWidth := lipgloss.Width(treePanel)
		availableWidth = m.width - actualTreeWidth
	} else {
		m.tree.focused = false
	}
	if availableWidth < 1 {
		availableWidth = 1
	}

	var rightPanel string
	if m.showEditor {
		editorHeight := contentHeight / 2
		resultsHeight := contentHeight - editorHeight

		m.editor.width = availableWidth
		m.editor.height = editorHeight

		m.results.width = availableWidth
		m.results.height = resultsHeight

		rightPanel = lipgloss.JoinVertical(lipgloss.Left, m.editor.View(), m.results.View())
	} else {
		m.results.width = availableWidth
		m.results.height = contentHeight

		rightPanel = m.results.View()
	}
	mainContent := rightPanel
	if m.showSidebar {
		mainContent = lipgloss.JoinHorizontal(lipgloss.Top, treePanel, rightPanel)
	}

	connName := m.connection.Name
	if m.connection.DBName != "" {
		connName += " / " + m.connection.DBName
	}

	mode := lipgloss.NewStyle().
		Background(lipgloss.Color("#7AA2F7")).
		Foreground(lipgloss.Color("#1A1B26")).
		Bold(true).
		Padding(0, 1).
		Render("NAV")

	leftStr := mode + "  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#C0CAF5")).Render(connName)

	helpKeys := lipgloss.JoinHorizontal(lipgloss.Left,
		KeyStyle.Render("Tab"),
		HelpStyle.Render(" switch "),
		KeyStyle.Render("\\"),
		HelpStyle.Render(" sidebar "),
		KeyStyle.Render("Ctrl+E"),
		HelpStyle.Render(" toggle sql "),
		KeyStyle.Render("Ctrl+P"),
		HelpStyle.Render(" connections "),
		KeyStyle.Render("q"),
		HelpStyle.Render(" quit"),
	)

	spacerWidth := max(1, m.width-lipgloss.Width(leftStr)-lipgloss.Width(helpKeys)-4)
	spacer := strings.Repeat(" ", spacerWidth)

	footer := lipgloss.NewStyle().
		Background(lipgloss.Color("#1A1B26")).
		Padding(0, 1).
		Width(m.width).
		Render(leftStr + spacer + helpKeys)

	return lipgloss.JoinVertical(lipgloss.Left, mainContent, footer)
}

func panelBorderColor(focused bool) lipgloss.Color {
	if focused {
		return SecondaryTextColor
	}
	return lipgloss.Color("#666A7E")
}
