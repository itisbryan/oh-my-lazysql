package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ResultsModel struct {
	columns        []GridColumn
	rows           [][]string
	width          int
	height         int
	status         string
	page           int
	pageSize       int
	totalRows      int
	row            int
	col            int
	colOffset      int
	activeTab      int
	focused        bool
	metadata       [5][][]string
	filterEditing  bool
	filterInput    string
	whereFilter    string
	editingCell    bool
	editInput      string
	pendingEdits   map[cellPosition]pendingEdit
	pendingDeletes map[int]bool
	insertingRow   bool
	insertRow      []string
	showRowDetail  bool
	detailRow      int
	detailEditing  bool
	detailInput    string
	completion     CompletionState
	sortCol        int
	sortDir        string
}

type cellPosition struct {
	row int
	col int
}

type pendingEdit struct {
	original string
	value    string
}

type whereFilterAppliedMsg struct {
	where string
}

type sortAppliedMsg struct{}

type GridColumn struct {
	Title string
	Width int
	Type  string
	IsPK  bool
	IsFK  bool
}

func NewResultsModel() *ResultsModel {
	return &ResultsModel{
		pageSize:       100,
		pendingEdits:   map[cellPosition]pendingEdit{},
		pendingDeletes: map[int]bool{},
		sortCol:        -1,
	}
}

func (m *ResultsModel) Init() tea.Cmd {
	return nil
}

func (m *ResultsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.clampSelection()
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if m.showRowDetail {
			m.updateRowDetail(msg)
			return m, nil
		}
		if m.filterEditing {
			return m, m.updateFilterInput(msg)
		}
		if m.editingCell {
			m.updateCellEdit(msg)
			return m, nil
		}
		switch msg.String() {
		case "o":
			if m.activeTab == 0 && len(m.rows) > 0 {
				m.showRowDetail = true
				m.detailRow = min(max(0, m.col), max(0, len(m.columns)-1))
			}
		case "a":
			if m.activeTab == 0 && len(m.columns) > 0 {
				m.startInsertRow()
			}
		case "ctrl+x":
			if m.activeTab == 0 && len(m.rows) > 0 && m.row >= 0 && m.row < len(m.rows) {
				m.pendingDeletes[m.row] = true
				m.status = fmt.Sprintf("Row %d marked for deletion. Press Ctrl+R to save or u to restore.", m.row+1)
			}
		case "u":
			if m.activeTab == 0 && len(m.rows) > 0 && m.row >= 0 && m.row < len(m.rows) {
				delete(m.pendingDeletes, m.row)
				m.status = fmt.Sprintf("Row %d restored", m.row+1)
			}
		case "s":
			if m.activeTab == 0 && len(m.columns) > 0 && m.col >= 0 && m.col < len(m.columns) {
				if m.sortCol == m.col {
					if m.sortDir == "ASC" {
						m.sortDir = "DESC"
					} else {
						m.sortCol = -1
						m.sortDir = ""
					}
				} else {
					m.sortCol = m.col
					m.sortDir = "ASC"
				}
				return m, func() tea.Msg { return sortAppliedMsg{} }
			}
		case "enter":
			m.startCellEdit()
		case "/":
			m.filterEditing = true
			m.filterInput = m.whereFilter
			m.completion.Update(m.filterInput)
		case "[":
			if m.activeTab > 0 {
				m.activeTab--
			}
		case "]":
			if m.activeTab < len(tabNames)-1 {
				m.activeTab++
			}
		case "1":
			m.activeTab = 0
		case "2":
			m.activeTab = 1
		case "3":
			m.activeTab = 2
		case "4":
			m.activeTab = 3
		case "5":
			m.activeTab = 4
		case "up", "k":
			if m.row > 0 {
				m.row--
			}
		case "down", "j":
			if m.row < len(m.rows)-1 {
				m.row++
			}
		case "left", "h":
			if m.col > 0 {
				m.col--
			}
		case "right", "l":
			if m.col < len(m.columns)-1 {
				m.col++
			}
		case "e":
			if len(m.columns) > 0 {
				m.col = (m.col + 1) % len(m.columns)
			}
		case "pgup":
			m.row = max(0, m.row-10)
		case "pgdown":
			if len(m.rows) > 0 {
				m.row = min(len(m.rows)-1, m.row+10)
			}
		case "ctrl+d":
			if len(m.rows) > 0 {
				m.row = min(len(m.rows)-1, m.row+5)
			}
		case "ctrl+u":
			m.row = max(0, m.row-5)
		}
	}
	m.clampSelection()
	return m, cmd
}

func (m *ResultsModel) updateFilterInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "enter":
		m.filterEditing = false
		m.whereFilter = strings.TrimSpace(m.filterInput)
		m.completion.Dismiss()
		return func() tea.Msg { return whereFilterAppliedMsg{where: m.whereFilter} }
	case "tab":
		if m.completion.Visible {
			m.filterInput = m.completion.Accept(m.filterInput)
		}
		m.completion.Update(m.filterInput)
	case "shift+tab":
		if m.completion.Visible {
			m.completion.Cycle(-1)
		}
	case "down":
		if m.completion.Visible {
			m.completion.Cycle(1)
		}
	case "up":
		if m.completion.Visible {
			m.completion.Cycle(-1)
		}
	case "esc":
		if m.completion.Visible {
			m.completion.Dismiss()
		} else {
			m.filterEditing = false
			m.filterInput = m.whereFilter
		}
	case "backspace", "ctrl+h":
		if len(m.filterInput) > 0 {
			runes := []rune(m.filterInput)
			m.filterInput = string(runes[:len(runes)-1])
		}
		m.completion.Update(m.filterInput)
	default:
		if len(msg.Runes) > 0 {
			m.filterInput += string(msg.Runes)
			m.completion.Update(m.filterInput)
		}
	}
	return nil
}

func (m *ResultsModel) startCellEdit() {
	if m.activeTab != 0 || m.row < 0 || m.col < 0 {
		return
	}
	if m.insertingRow {
		m.editingCell = true
		m.editInput = m.insertRow[m.col]
		return
	}
	if m.row >= len(m.rows) || m.col >= len(m.rows[m.row]) {
		return
	}
	m.editingCell = true
	m.editInput = ""
}

func (m *ResultsModel) updateCellEdit(msg tea.KeyMsg) {
	switch msg.String() {
	case "enter":
		m.commitCellEdit()
	case "esc":
		if m.insertingRow {
			m.insertingRow = false
			m.insertRow = nil
			if len(m.rows) > 0 {
				m.row = len(m.rows) - 1
			} else {
				m.row = 0
			}
		}
		m.editingCell = false
		m.editInput = ""
	case "tab":
		m.commitCellEdit()
		if m.insertingRow && m.col < len(m.insertRow)-1 {
			m.col++
			m.editingCell = true
			m.editInput = m.insertRow[m.col]
		}
	case "backspace", "ctrl+h":
		if len(m.editInput) > 0 {
			runes := []rune(m.editInput)
			m.editInput = string(runes[:len(runes)-1])
			if m.insertingRow {
				m.insertRow[m.col] = m.editInput
			}
		}
	default:
		if len(msg.Runes) > 0 {
			m.editInput += string(msg.Runes)
			if m.insertingRow {
				m.insertRow[m.col] = m.editInput
			}
		}
	}
}

func (m *ResultsModel) commitCellEdit() {
	if m.insertingRow {
		if m.row < 0 || m.col < 0 || m.col >= len(m.insertRow) {
			m.editingCell = false
			m.editInput = ""
			return
		}
		m.insertRow[m.col] = m.editInput
		m.editingCell = false
		m.editInput = ""
		if m.col < len(m.insertRow)-1 {
			m.col++
			m.editingCell = true
			m.editInput = m.insertRow[m.col]
		}
		return
	}
	if m.row < 0 || m.row >= len(m.rows) || m.col < 0 || m.col >= len(m.rows[m.row]) {
		m.editingCell = false
		return
	}
	pos := cellPosition{row: m.row, col: m.col}
	original := m.rows[m.row][m.col]
	if existing, ok := m.pendingEdits[pos]; ok {
		original = existing.original
	}
	newValue := m.editInput
	m.rows[m.row][m.col] = newValue
	if newValue == original {
		delete(m.pendingEdits, pos)
	} else {
		m.pendingEdits[pos] = pendingEdit{original: original, value: newValue}
	}
	m.editingCell = false
	m.editInput = ""
}

func (m *ResultsModel) updateRowDetail(msg tea.KeyMsg) {
	if m.detailEditing {
		m.updateDetailEdit(msg)
		return
	}
	switch msg.String() {
	case "esc":
		m.showRowDetail = false
	case "up", "k":
		if m.detailRow > 0 {
			m.detailRow--
		}
	case "down", "j":
		if m.detailRow < len(m.columns)-1 {
			m.detailRow++
		}
	case "enter":
		m.startDetailEdit()
	}
}

func (m *ResultsModel) startDetailEdit() {
	if m.row < 0 || m.row >= len(m.rows) || m.detailRow < 0 || m.detailRow >= len(m.columns) {
		return
	}
	if m.detailRow >= len(m.rows[m.row]) {
		return
	}
	m.detailEditing = true
	m.detailInput = ""
}

func (m *ResultsModel) updateDetailEdit(msg tea.KeyMsg) {
	switch msg.String() {
	case "enter":
		m.commitDetailEdit()
	case "esc":
		m.detailEditing = false
		m.detailInput = ""
	case "backspace", "ctrl+h":
		if len(m.detailInput) > 0 {
			runes := []rune(m.detailInput)
			m.detailInput = string(runes[:len(runes)-1])
		}
	default:
		if len(msg.Runes) > 0 {
			m.detailInput += string(msg.Runes)
		}
	}
}

func (m *ResultsModel) commitDetailEdit() {
	if m.row < 0 || m.row >= len(m.rows) || m.detailRow < 0 || m.detailRow >= len(m.rows[m.row]) {
		m.detailEditing = false
		return
	}
	pos := cellPosition{row: m.row, col: m.detailRow}
	original := m.rows[m.row][m.detailRow]
	if existing, ok := m.pendingEdits[pos]; ok {
		original = existing.original
	}
	m.rows[m.row][m.detailRow] = m.detailInput
	if m.detailInput == original {
		delete(m.pendingEdits, pos)
	} else {
		m.pendingEdits[pos] = pendingEdit{original: original, value: m.detailInput}
	}
	m.detailEditing = false
	m.detailInput = ""
}

func (m *ResultsModel) pendingChangeCount() int {
	count := len(m.pendingEdits) + len(m.pendingDeletes)
	if m.insertingRow && m.insertRow != nil {
		count++
	}
	return count
}

func (m *ResultsModel) startInsertRow() {
	if m.insertingRow {
		return
	}
	m.insertingRow = true
	m.insertRow = make([]string, len(m.columns))
	for i := range m.columns {
		m.insertRow[i] = ""
	}
	m.row = len(m.rows)
	m.col = 0
	m.editingCell = true
	m.editInput = ""
}

func (m *ResultsModel) View() string {
	if len(m.columns) == 0 {
		return lipgloss.NewStyle().
			Foreground(InverseTextColor).
			Render("Select a table to load records")
	}

	pagination := "0 rows"
	if m.totalRows > 0 {
		pagination = fmt.Sprintf("%d-%d of %d", m.page*m.pageSize+1,
			min((m.page+1)*m.pageSize, m.totalRows), m.totalRows)
	}

	tabs := m.renderTabs()
	filter := m.renderFilterBar()

	contentHeight := max(1, m.height-6)
	if m.filterEditing && m.completion.Visible {
		dropdownLines := min(6, len(m.completion.Suggestions)) + 2
		contentHeight = max(1, contentHeight-dropdownLines)
	}

	var content string
	switch m.activeTab {
	case 0:
		if m.showRowDetail {
			content = m.renderRowDetail(max(1, m.width-2), contentHeight)
		} else {
			content = m.renderGrid(max(1, m.width-2), contentHeight)
		}
	default:
		data := m.metadata[m.activeTab]
		if len(data) == 0 {
			content = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#565F89")).
				Width(max(1, m.width-2)).
				Height(contentHeight).
				Render("Loading " + tabNames[m.activeTab] + "...")
		} else {
			content = m.renderMetadataGrid(data, max(1, m.width-2), contentHeight)
		}
	}

	statusline := m.renderStatusline(pagination)

	parts := []string{filter, tabs, content, statusline}
	if m.status != "" {
		parts = append(parts, StatusStyle.Render(m.status))
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m *ResultsModel) renderFilterBar() string {
	activeBadge := ""
	if m.whereFilter != "" && !m.filterEditing {
		activeBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1A1B26")).
			Background(lipgloss.Color("#FF9E64")).
			Bold(true).
			Padding(0, 1).
			Render("ACTIVE")
	}

	badge := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7AA2F7")).
		Background(lipgloss.Color("#1F2335")).
		Bold(true).
		Padding(0, 1).
		Render("WHERE")

	placeholder := "filter rows by SQL predicate"
	fieldText := placeholder
	fieldColor := lipgloss.Color("#565F89")
	hintText := "Enter Apply"
	if m.filterEditing {
		fieldText = highlightSQL(m.filterInput) + "▌"
		if m.filterInput == "" {
			fieldText = placeholder + " ▌"
		}
		fieldColor = PrimaryTextColor
		hintText = "Enter Apply  Esc Cancel"
	} else if m.whereFilter != "" {
		fieldText = highlightSQL(m.whereFilter)
		fieldColor = PrimaryTextColor
		hintText = "/ Edit  Enter Apply"
	} else {
		hintText = "/ Edit"
	}

	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#565F89")).
		Render(hintText)

	badges := badge
	if activeBadge != "" {
		badges = lipgloss.JoinHorizontal(lipgloss.Left, badge, " ", activeBadge)
	}

	gap := "    "
	fieldWidth := max(12, m.width-lipgloss.Width(badges)-lipgloss.Width(hint)-14)

	ghostSuffix := ""
	if m.filterEditing {
		ghostSuffix = m.completion.GhostSuffix(m.filterInput)
	}
	ghostStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#565F89"))

	if m.filterEditing && (ghostSuffix != "" || m.completion.Visible) {
		renderedText := fieldText
		if ghostSuffix != "" {
			ghostTextWidth := max(0, fieldWidth-lipgloss.Width(truncateDisplay(fieldText, fieldWidth)))
			if ghostTextWidth > 0 {
				renderedText = truncateDisplay(fieldText, fieldWidth) + ghostStyle.Render(truncateDisplay(ghostSuffix, ghostTextWidth))
			}
		}
		field := lipgloss.NewStyle().
			Foreground(fieldColor).
			Background(lipgloss.Color("#151A2B")).
			Padding(0, 1).
			Width(fieldWidth).
			Render(renderedText)
		left := lipgloss.JoinHorizontal(lipgloss.Left, badges, gap, field)
		spacer := strings.Repeat(" ", max(2, m.width-lipgloss.Width(left)-lipgloss.Width(hint)-6))

		filterBar := lipgloss.NewStyle().
			Background(lipgloss.Color("#161B2D")).
			Padding(0, 1, 0, 2).
			Width(max(1, m.width-2)).
			Render(left + spacer + hint)

		if m.completion.Visible {
			cursorPos := lipgloss.Width(badges) + len(gap) + 2 + min(len(m.filterInput), fieldWidth)
			dropWidth := min(24, max(1, m.width-cursorPos-4))
			dropdown := renderCompletionDropdown(m.completion.Suggestions, m.completion.SelectedIndex, dropWidth)
			indented := strings.Repeat(" ", cursorPos) + dropdown
			return filterBar + "\n" + indented
		}
		return filterBar
	}

	suggestionText := ""
	field := lipgloss.NewStyle().
		Foreground(fieldColor).
		Background(lipgloss.Color("#151A2B")).
		Padding(0, 1).
		Width(fieldWidth).
		Render(truncateDisplay(fieldText, fieldWidth) + suggestionText)

	left := lipgloss.JoinHorizontal(lipgloss.Left, badges, gap, field)
	spacer := strings.Repeat(" ", max(2, m.width-lipgloss.Width(left)-lipgloss.Width(hint)-6))

	return lipgloss.NewStyle().
		Background(lipgloss.Color("#161B2D")).
		Padding(1, 2).
		Width(max(1, m.width-2)).
		Render(left + spacer + hint)
}

func (m *ResultsModel) renderTabs() string {
	var parts []string
	divider := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3B4261")).
		Render("│")
	for i, name := range tabNames {
		if i > 0 {
			parts = append(parts, divider)
		}
		label := name
		if i == m.activeTab {
			label = "▸ " + label
			parts = append(parts, lipgloss.NewStyle().
				Background(lipgloss.Color("#20304F")).
				Foreground(lipgloss.Color("#7DCFFF")).
				Bold(true).
				Padding(0, 1).
				Render(label))
		} else {
			parts = append(parts, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#BB9AF7")).
				Padding(0, 1).
				Render(label))
		}
	}
	tabList := lipgloss.JoinHorizontal(lipgloss.Left, parts...)
	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#565F89")).
		Render("[/] tabs  1-5 jump")
	spacer := strings.Repeat(" ", max(1, m.width-lipgloss.Width(tabList)-lipgloss.Width(hint)-4))

	return lipgloss.NewStyle().
		Background(lipgloss.Color("#1A1B26")).
		Padding(0, 1).
		Width(max(1, m.width-2)).
		Render(tabList + spacer + hint)
}

func (m *ResultsModel) renderStatusline(pagination string) string {
	modeBg := lipgloss.Color("#7AA2F7")
	modeFg := lipgloss.Color("#1A1B26")

	mode := lipgloss.NewStyle().
		Background(modeBg).
		Foreground(modeFg).
		Bold(true).
		Padding(0, 1).
		Render("RECORDS")

	var info []string

	sep := lipgloss.NewStyle().Foreground(lipgloss.Color("#3B4261")).Render(" │ ")

	if m.col < len(m.columns) {
		col := m.columns[m.col]
		if col.Type != "" {
			info = append(info, lipgloss.NewStyle().Foreground(lipgloss.Color("#7DCFFF")).Render(abbrevType(col.Type)))
		}
		if col.IsPK {
			info = append(info, lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9E64")).Render("PK"))
		}
		if col.IsFK {
			info = append(info, lipgloss.NewStyle().Foreground(lipgloss.Color("#9ECE6A")).Render("FK"))
		}
	}

	pos := fmt.Sprintf("R%d:C%d", m.row+1, m.col+1)
	info = append(info, lipgloss.NewStyle().Foreground(lipgloss.Color("#C0CAF5")).Render(pos))

	pending := m.pendingChangeCount()
	if pending > 0 {
		info = append(info, lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9E64")).Render(fmt.Sprintf("%d pending", pending)))
	}

	if m.sortCol >= 0 && m.sortCol < len(m.columns) && m.sortDir != "" {
		sortInfo := fmt.Sprintf("↕ %s %s", m.columns[m.sortCol].Title, m.sortDir)
		info = append(info, lipgloss.NewStyle().Foreground(lipgloss.Color("#BB9AF7")).Render(sortInfo))
	}

	leftStr := mode + sep + strings.Join(info, sep)

	rightParts := []string{
		lipgloss.NewStyle().Foreground(lipgloss.Color("#7AA2F7")).Render(pagination),
	}
	rightStr := lipgloss.JoinHorizontal(lipgloss.Left, rightParts...)

	spacerWidth := max(1, m.width-lipgloss.Width(leftStr)-lipgloss.Width(rightStr)-4)
	spacer := strings.Repeat(" ", spacerWidth)

	return lipgloss.NewStyle().
		Background(lipgloss.Color("#1A1B26")).
		Padding(0, 1).
		Width(max(1, m.width-2)).
		Render(leftStr + spacer + rightStr)
}

func (m *ResultsModel) renderGrid(width, height int) string {
	if len(m.columns) == 0 {
		return ""
	}
	m.clampSelection()

	rowNumberWidth := 5
	m.resizeColumnsFromVisibleData()
	m.ensureCellVisible(width - rowNumberWidth)

	visibleCols := []int{}
	usedWidth := rowNumberWidth + 1
	for i := m.colOffset; i < len(m.columns); i++ {
		colWidth := m.columns[i].Width
		if usedWidth+colWidth+1 > width && len(visibleCols) > 0 {
			break
		}
		visibleCols = append(visibleCols, i)
		usedWidth += colWidth + 1
	}

	startRow := 0
	gridHeight := max(1, height-4)
	linesPerRow := 2
	visibleRowCount := max(1, gridHeight/linesPerRow)
	if m.row >= startRow+visibleRowCount {
		startRow = m.row - visibleRowCount + 1
	}
	endRow := min(len(m.rows), startRow+visibleRowCount)

	divider := lipgloss.NewStyle().Foreground(lipgloss.Color("#3B4261")).Render("│")
	dimmedDivider := lipgloss.NewStyle().Foreground(dimColor).Render("│")

	header := m.renderGridHeader(visibleCols, rowNumberWidth)

	lines := []string{header}
	for rowIdx := startRow; rowIdx < endRow; rowIdx++ {
		isActiveRow := rowIdx == m.row
		isDeleted := m.pendingDeletes[rowIdx]
		sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3B4261"))
		if !isActiveRow {
			sepStyle = lipgloss.NewStyle().Foreground(dimColor)
		}
		sepParts := []string{sepStyle.Render(strings.Repeat("─", rowNumberWidth))}
		for _, colIdx := range visibleCols {
			sepParts = append(sepParts, sepStyle.Render("┼"), sepStyle.Render(strings.Repeat("─", m.columns[colIdx].Width)))
		}
		sepParts = append(sepParts, sepStyle.Render("┼"))
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Left, sepParts...))

		row := m.rows[rowIdx]
		rowStyle := lipgloss.NewStyle()
		if !isDeleted && rowIdx%2 == 1 {
			rowStyle = rowStyle.Background(lipgloss.Color("#111827"))
		}
		rowNumberStyle := lipgloss.NewStyle().Width(rowNumberWidth)
		if isDeleted {
			rowNumberStyle = rowNumberStyle.Foreground(lipgloss.Color("#F7768E")).Bold(true)
		} else if isActiveRow {
			rowNumberStyle = rowNumberStyle.Foreground(SecondaryTextColor).Bold(true)
		} else {
			rowNumberStyle = rowNumberStyle.Foreground(lipgloss.Color("#565F89"))
		}
		rowLabel := fmt.Sprintf("%d", rowIdx+1)
		if isDeleted {
			rowLabel = "✗" + rowLabel
		}
		cells := []string{rowNumberStyle.Render(rowLabel)}
		for _, colIdx := range visibleCols {
			value := ""
			if colIdx < len(row) {
				value = row[colIdx]
			}
			displayValue, special := formatCellValue(value)
			colWidth := m.columns[colIdx].Width
			if isDeleted {
				cell := renderDeletedCell(displayValue, colWidth, special)
				cells = append(cells, divider, cell)
			} else if isActiveRow {
				cell := renderPlainCell(displayValue, colWidth)
				if special {
					cell = renderSpecialCell(displayValue, colWidth)
				}
				if colIdx == m.col {
					if m.editingCell {
						cell = renderEditingCell(m.editInput, colWidth)
					} else {
						cell = renderSelectedCell(displayValue, colWidth, special)
					}
				} else if _, dirty := m.pendingEdits[cellPosition{row: rowIdx, col: colIdx}]; dirty {
					cell = renderDirtyCell(displayValue, colWidth)
				}
				cells = append(cells, divider, cell)
			} else {
				cell := renderDimmedCell(displayValue, colWidth)
				if special {
					cell = renderDimmedSpecialCell(displayValue, colWidth)
				}
				if _, dirty := m.pendingEdits[cellPosition{row: rowIdx, col: colIdx}]; dirty {
					cell = renderDirtyCell(displayValue, colWidth)
				}
				cells = append(cells, dimmedDivider, cell)
			}
		}
		lines = append(lines, rowStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, cells...)))
	}

	if m.insertingRow && m.insertRow != nil {
		sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9ECE6A"))
		sepParts := []string{sepStyle.Render(strings.Repeat("─", rowNumberWidth))}
		for _, colIdx := range visibleCols {
			sepParts = append(sepParts, sepStyle.Render("┼"), sepStyle.Render(strings.Repeat("─", m.columns[colIdx].Width)))
		}
		sepParts = append(sepParts, sepStyle.Render("┼"))
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Left, sepParts...))

		rowNumberStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9ECE6A")).Bold(true).Width(rowNumberWidth)
		cells := []string{rowNumberStyle.Render("+NEW")}
		for _, colIdx := range visibleCols {
			value := ""
			if colIdx < len(m.insertRow) {
				value = m.insertRow[colIdx]
				if value == "" {
					value = "NULL"
				}
			}
			colWidth := m.columns[colIdx].Width
			if m.editingCell && colIdx == m.col {
				cells = append(cells, divider, renderEditingCell(m.editInput+"▌", colWidth))
			} else {
				cells = append(cells, divider, renderInsertCell(value, colWidth))
			}
		}
		insertStyle := lipgloss.NewStyle().Background(lipgloss.Color("#162820"))
		lines = append(lines, insertStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, cells...)))
	}

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3B4261")).
		Height(height).
		Width(max(1, width)).
		Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m *ResultsModel) renderMetadataGrid(data [][]string, width, height int) string {
	if len(data) == 0 {
		return ""
	}

	headers := data[0]
	rows := data[1:]

	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = lipgloss.Width(h) + 2
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) {
				cellW := lipgloss.Width(cell) + 2
				if cellW > colWidths[i] {
					colWidths[i] = cellW
				}
			}
		}
	}

	visibleCols := []int{}
	usedWidth := 0
	for i, w := range colWidths {
		if usedWidth+w+1 > width && len(visibleCols) > 0 {
			break
		}
		visibleCols = append(visibleCols, i)
		usedWidth += w + 1
	}

	hdrCells := []string{}
	sepParts := []string{}
	for _, i := range visibleCols {
		w := colWidths[i]
		if i >= len(headers) {
			break
		}
		hdrCells = append(hdrCells,
			lipgloss.NewStyle().Foreground(lipgloss.Color("#7AA2F7")).Background(lipgloss.Color("#1F2335")).Width(w).Bold(true).Render(" "+headers[i]),
		)
		sepParts = append(sepParts,
			lipgloss.NewStyle().Foreground(lipgloss.Color("#3B4261")).Render(strings.Repeat("─", w)),
		)
	}
	lines := []string{
		lipgloss.JoinHorizontal(lipgloss.Left, hdrCells...),
		lipgloss.JoinHorizontal(lipgloss.Left, sepParts...),
	}

	for rowIdx, row := range rows {
		rowStyle := lipgloss.NewStyle()
		if rowIdx%2 == 1 {
			rowStyle = rowStyle.Background(lipgloss.Color("#111827"))
		}
		cells := []string{}
		for _, i := range visibleCols {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			w := colWidths[i]
			cells = append(cells, lipgloss.NewStyle().Foreground(PrimaryTextColor).Width(w).Render(" "+cell))
		}
		lines = append(lines, rowStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, cells...)))
		if len(lines) >= height-2 {
			break
		}
	}

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3B4261")).
		Height(height).
		Width(max(1, width)).
		Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m *ResultsModel) renderRowDetail(width, height int) string {
	if m.row < 0 || m.row >= len(m.rows) {
		return ""
	}
	row := m.rows[m.row]
	nameWidth := 0
	for _, col := range m.columns {
		nameWidth = max(nameWidth, lipgloss.Width(col.Title))
	}
	nameWidth = min(max(8, nameWidth), max(8, width/3))

	titleText := fmt.Sprintf("Row %d", m.row+1)
	metaText := fmt.Sprintf("%d fields", len(m.columns))
	if count := m.pendingChangeCount(); count > 0 {
		metaText += fmt.Sprintf(" · %d pending", count)
	}
	headerLeft := lipgloss.NewStyle().Foreground(lipgloss.Color("#1A1B26")).Background(SecondaryTextColor).Bold(true).Padding(0, 1).Render(titleText)
	headerRight := lipgloss.NewStyle().Foreground(lipgloss.Color("#565F89")).Render(metaText)
	headerGap := strings.Repeat(" ", max(1, width-lipgloss.Width(headerLeft)-lipgloss.Width(headerRight)-6))

	typeWidth := 12
	valueWidth := max(12, width-nameWidth-typeWidth-12)
	columnHeader := lipgloss.JoinHorizontal(lipgloss.Left,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#565F89")).Width(nameWidth+2).Bold(true).Render("FIELD"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#565F89")).Width(typeWidth).Bold(true).Render("TYPE"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#565F89")).Width(valueWidth).Bold(true).Render("VALUE"),
	)

	lines := []string{
		headerLeft + headerGap + headerRight,
		columnHeader,
	}

	for i, col := range m.columns {
		value := ""
		if i < len(row) {
			value = row[i]
		}
		displayValue, special := formatCellValue(value)
		nameText := col.Title
		if _, dirty := m.pendingEdits[cellPosition{row: m.row, col: i}]; dirty {
			nameText = "*" + nameText
		}
		name := lipgloss.NewStyle().Foreground(lipgloss.Color("#7AA2F7")).Width(nameWidth).Render(truncateDisplay(nameText, nameWidth))
		badges := lipgloss.NewStyle().Width(typeWidth).Render(truncateDisplay(rowDetailBadges(col), typeWidth))
		valueStyle := lipgloss.NewStyle().Foreground(PrimaryTextColor)
		if special {
			valueStyle = lipgloss.NewStyle().Foreground(InverseTextColor)
		}
		availableValueWidth := max(1, valueWidth)
		valueText := truncateDisplay(displayValue, availableValueWidth)
		if i == m.detailRow && m.detailEditing {
			valueText = truncateDisplay(m.detailInput+"▌", availableValueWidth)
			valueStyle = lipgloss.NewStyle().Foreground(PrimaryTextColor).Bold(true)
		} else if _, dirty := m.pendingEdits[cellPosition{row: m.row, col: i}]; dirty {
			valueText = truncateDisplay("*"+displayValue, availableValueWidth)
			valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9E64"))
		}
		rail := " "
		if i == m.detailRow {
			rail = lipgloss.NewStyle().Foreground(SecondaryTextColor).Render("▌")
		}
		line := lipgloss.JoinHorizontal(lipgloss.Left,
			rail,
			" ",
			name,
			"  ",
			badges,
			valueStyle.Render(valueText),
		)
		if i == m.detailRow {
			pad := max(0, width-lipgloss.Width(line)-4)
			line = lipgloss.NewStyle().Background(lipgloss.Color("#283457")).Render(line + strings.Repeat(" ", pad))
		}
		lines = append(lines, line)
		if len(lines) >= height-3 {
			break
		}
	}

	footerText := "j/k field  Enter edit  Esc close  Ctrl+R save"
	if m.detailEditing {
		footerText = "Enter apply  Esc cancel"
	}
	footer := renderDetailFooter(footerText)
	lines = append(lines, footer)

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3B4261")).
		Padding(0, 1).
		Height(height).
		Width(max(1, width)).
		Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func rowDetailBadges(col GridColumn) string {
	parts := []string{}
	if col.Type != "" {
		parts = append(parts, abbrevType(col.Type))
	}
	if col.IsPK {
		parts = append(parts, "PK")
	}
	if col.IsFK {
		parts = append(parts, "FK")
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " ")
}

func renderDetailFooter(text string) string {
	parts := strings.Split(text, "  ")
	styled := make([]string, 0, len(parts))
	for _, part := range parts {
		styled = append(styled, lipgloss.NewStyle().Foreground(lipgloss.Color("#565F89")).Render(part))
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, styled...)
}

func (m *ResultsModel) renderGridHeader(visibleCols []int, rowNumberWidth int) string {
	_, rowNumberBackground, rowNumberSeparator := headerColorsForDistance(2)
	headerDivider := lipgloss.NewStyle().Foreground(lipgloss.Color(rowNumberSeparator)).Render("┼")
	headerCells := []string{
		lipgloss.NewStyle().
			Foreground(InverseTextColor).
			Background(lipgloss.Color(rowNumberBackground)).
			Width(rowNumberWidth).
			Height(2).
			Bold(true).
			Render("#"),
	}

	for _, colIdx := range visibleCols {
		col := m.columns[colIdx]
		distance := colIdx - m.col
		if distance < 0 {
			distance = -distance
		}
		foreground, background, separator := headerColorsForDistance(distance)
		headerDivider = lipgloss.NewStyle().Foreground(lipgloss.Color(separator)).Render("┼")

		keyIcon := ""
		if col.IsPK {
			keyIcon = " 󰌽"
		} else if col.IsFK {
			keyIcon = " 󰌿"
		}

		label := "  " + col.Title
		if colIdx == m.col {
			label = "▸ " + col.Title
		}
		if m.sortCol == colIdx && m.sortDir != "" {
			if m.sortDir == "ASC" {
				label += "↑"
			} else {
				label += "↓"
			}
		}

		typeLabel := ""
		if col.Type != "" {
			short := abbrevType(col.Type)
			if lipgloss.Width(short) > col.Width-2 {
				short = short[:col.Width-2]
			}
			typeLabel = short
		}

		nameLine := truncateDisplay(label+keyIcon, col.Width)
		typeLine := ""
		if typeLabel != "" {
			typeLine = "\n" + lipgloss.NewStyle().
				Foreground(lipgloss.Color("#565F89")).
				Background(lipgloss.Color(background)).
				Render(truncateDisplay("  "+typeLabel, col.Width))
		}

		headerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(foreground)).
			Background(lipgloss.Color(background)).
			Width(col.Width).
			Bold(true)

		if typeLine != "" {
			headerStyle = headerStyle.Height(2)
		}

		headerCells = append(headerCells, headerDivider, headerStyle.Render(nameLine+typeLine))
	}

	return lipgloss.NewStyle().
		Foreground(PrimaryTextColor).
		Background(lipgloss.Color(rowNumberBackground)).
		Render(lipgloss.JoinHorizontal(lipgloss.Left, headerCells...))
}

func headerColorsForDistance(distance int) (foreground, background, separator lipgloss.Color) {
	if distance == 0 {
		return SecondaryTextColor, lipgloss.Color("#283457"), SecondaryTextColor
	}
	if distance == 1 {
		return lipgloss.Color("#BB9AF7"), lipgloss.Color("#1F2335"), lipgloss.Color("#7DCFFF")
	}
	return PrimaryTextColor, lipgloss.Color("#1A1F33"), lipgloss.Color("#565F89")
}

var dimColor = lipgloss.Color("#565F89")

var tabNames = []string{"Records", "Columns", "Constraints", "Foreign Keys", "Indexes"}

func renderPlainCell(value string, width int) string {
	return lipgloss.NewStyle().Width(width).Render(truncateDisplay(value, width))
}

func renderDimmedCell(value string, width int) string {
	return lipgloss.NewStyle().Foreground(dimColor).Width(width).Render(truncateDisplay(value, width))
}

func renderSpecialCell(value string, width int) string {
	text := truncateDisplay(value, width)
	styled := lipgloss.NewStyle().Foreground(InverseTextColor).Render(text)
	padding := max(0, width-lipgloss.Width(text))
	return styled + strings.Repeat(" ", padding)
}

func renderDimmedSpecialCell(value string, width int) string {
	text := truncateDisplay(value, width)
	styled := lipgloss.NewStyle().Foreground(dimColor).Render(text)
	padding := max(0, width-lipgloss.Width(text))
	return styled + strings.Repeat(" ", padding)
}

func renderSelectedCell(value string, width int, special bool) string {
	if width <= 2 {
		return lipgloss.NewStyle().Foreground(SecondaryTextColor).Render(truncateDisplay(value, width))
	}

	innerWidth := width - 2
	text := truncateDisplay(value, innerWidth)
	style := lipgloss.NewStyle().Foreground(SecondaryTextColor).Bold(true)
	if special {
		style = lipgloss.NewStyle().Foreground(InverseTextColor)
	}

	left := lipgloss.NewStyle().Foreground(SecondaryTextColor).Render("▌")
	right := lipgloss.NewStyle().Foreground(SecondaryTextColor).Render("▐")
	styled := style.Render(text)
	padding := max(0, innerWidth-lipgloss.Width(text))
	return left + styled + strings.Repeat(" ", padding) + right
}

func renderEditingCell(value string, width int) string {
	text := truncateDisplay(value+"▌", width)
	return lipgloss.NewStyle().Foreground(editingCellForeground()).Background(lipgloss.Color("#283457")).Width(width).Render(text)
}

func editingCellForeground() lipgloss.Color {
	return PrimaryTextColor
}

func renderDirtyCell(value string, width int) string {
	if width <= 1 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9E64")).Render(truncateDisplay(value, width))
	}
	text := truncateDisplay("*"+value, width)
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9E64")).Width(width).Render(text)
}

func renderDeletedCell(value string, width int, special bool) string {
	if special {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#F7768E")).Width(width).Render(truncateDisplay(value, width))
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#F7768E")).Width(width).Render(truncateDisplay(value, width))
}

func renderInsertCell(value string, width int) string {
	_, special := formatCellValue(value)
	if special {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#9ECE6A")).Width(width).Render(truncateDisplay(value, width))
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#73DACA")).Width(width).Render(truncateDisplay(value, width))
}

func formatCellValue(value string) (string, bool) {
	switch value {
	case "NULL&":
		return "NULL", true
	case "EMPTY&":
		return "EMPTY", true
	case "DEFAULT&":
		return "DEFAULT", true
	default:
		return value, false
	}
}

func abbrevType(t string) string {
	lower := strings.ToLower(t)
	switch {
	case strings.HasPrefix(lower, "bigint"):
		return "bigint"
	case strings.HasPrefix(lower, "smallint"):
		return "smallint"
	case strings.HasPrefix(lower, "integer") || lower == "int" || lower == "int4" || lower == "int2":
		return "int"
	case strings.HasPrefix(lower, "boolean") || lower == "bool":
		return "bool"
	case strings.HasPrefix(lower, "timestamp"):
		return "tstamp"
	case strings.HasPrefix(lower, "character varying") || strings.HasPrefix(lower, "varchar"):
		return "varchar"
	case strings.HasPrefix(lower, "character") || strings.HasPrefix(lower, "char"):
		return "char"
	case strings.HasPrefix(lower, "text") || strings.HasPrefix(lower, "ntext") || strings.HasPrefix(lower, "citext"):
		return "text"
	case strings.HasPrefix(lower, "numeric") || strings.HasPrefix(lower, "decimal"):
		return "decimal"
	case strings.HasPrefix(lower, "real") || lower == "float4":
		return "float4"
	case strings.HasPrefix(lower, "double") || lower == "float8":
		return "float8"
	case strings.HasPrefix(lower, "float"):
		return "float"
	case strings.HasPrefix(lower, "date"):
		return "date"
	case strings.HasPrefix(lower, "time"):
		return "time"
	case strings.HasPrefix(lower, "uuid"):
		return "uuid"
	case strings.HasPrefix(lower, "json"):
		return "json"
	case strings.HasPrefix(lower, "bytea") || strings.HasPrefix(lower, "blob") || strings.HasPrefix(lower, "binary"):
		return "bytes"
	case strings.HasPrefix(lower, "serial") || strings.HasPrefix(lower, "identity"):
		return "serial"
	case strings.HasPrefix(lower, "auto_increment"):
		return "auto"
	default:
		if len(t) > 8 {
			return t[:7] + "…"
		}
		return t
	}
}

func (m *ResultsModel) resizeColumnsFromVisibleData() {
	for colIdx := range m.columns {
		width := headerMinWidth(m.columns[colIdx])
		for rowIdx := 0; rowIdx < len(m.rows) && rowIdx < 100; rowIdx++ {
			if colIdx >= len(m.rows[rowIdx]) {
				continue
			}
			width = max(width, lipgloss.Width(m.rows[rowIdx][colIdx])+2)
		}
		m.columns[colIdx].Width = max(headerMinWidth(m.columns[colIdx]), min(34, width))
	}
}

func headerMinWidth(col GridColumn) int {
	width := lipgloss.Width(col.Title) + 4
	if col.IsPK || col.IsFK {
		width += 2
	}
	if col.Type != "" {
		width = max(width, lipgloss.Width(abbrevType(col.Type))+2)
	}
	return max(8, width)
}

func (m *ResultsModel) ensureCellVisible(width int) {
	m.clampSelection()
	if m.col < m.colOffset {
		m.colOffset = m.col
	}
	for m.colOffset < m.col {
		used := 0
		for i := m.colOffset; i <= m.col && i < len(m.columns); i++ {
			used += m.columns[i].Width + 1
		}
		if used <= width {
			break
		}
		m.colOffset++
	}
}

func (m *ResultsModel) clampSelection() {
	if m.insertingRow && m.insertRow != nil {
		m.row = len(m.rows)
	} else if len(m.rows) == 0 {
		m.row = 0
	} else if m.row < 0 {
		m.row = 0
	} else if m.row >= len(m.rows) {
		m.row = len(m.rows) - 1
	}

	if len(m.columns) == 0 {
		m.col = 0
		m.colOffset = 0
		return
	}
	if m.col < 0 {
		m.col = 0
	} else if m.col >= len(m.columns) {
		m.col = len(m.columns) - 1
	}
	if m.colOffset < 0 {
		m.colOffset = 0
	} else if m.colOffset >= len(m.columns) {
		m.colOffset = len(m.columns) - 1
	}
}

func (m *ResultsModel) SetData(columns []string, colTypes map[string]string, pkCols, fkCols map[string]bool, rows [][]string) {
	tableCols := make([]GridColumn, len(columns))
	for i, col := range columns {
		width := max(8, min(34, lipgloss.Width(col)+2))
		tableCols[i] = GridColumn{
			Title: col,
			Width: width,
			Type:  colTypes[col],
			IsPK:  pkCols[col],
			IsFK:  fkCols[col],
		}
	}
	m.columns = tableCols
	m.rows = make([][]string, len(rows))
	for i, row := range rows {
		m.rows[i] = row
	}
	m.totalRows = len(rows)
	m.row = 0
	m.col = 0
	m.colOffset = 0
	m.editingCell = false
	m.editInput = ""
	m.pendingEdits = map[cellPosition]pendingEdit{}
	m.pendingDeletes = map[int]bool{}
	m.insertingRow = false
	m.insertRow = nil
	m.clampSelection()
}

func (m *ResultsModel) SetStatus(msg string) {
	m.status = msg
}

func (m *ResultsModel) Clear() {
	m.rows = [][]string{}
	m.totalRows = 0
	m.status = ""
	m.row = 0
	m.col = 0
	m.colOffset = 0
}
