package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/itisbryan/oh-my-lazysql/drivers"
)

type editorMode int

const (
	normalMode editorMode = iota
	insertMode
)

type EditorModel struct {
	lines        []string
	cursorRow    int
	cursorCol    int
	scrollRow    int
	mode         editorMode
	driver       drivers.Driver
	results      *ResultsModel
	width        int
	height       int
	focused      bool
	executing    bool
	spinnerFrame int
	completion   CompletionState
	pendingKey   string
	preferCol    int
}

type editorQueryExecutedMsg struct {
	results  [][]string
	rowCount int
	err      error
}

func NewEditorModel() *EditorModel {
	return &EditorModel{
		lines:     []string{""},
		mode:      normalMode,
		focused:   true,
		preferCol: 0,
	}
}

func (m *EditorModel) SetDriver(driver drivers.Driver) {
	m.driver = driver
}

func (m *EditorModel) SetResults(results *ResultsModel) {
	m.results = results
}

func (m *EditorModel) Init() tea.Cmd {
	return tea.Tick(time.Second/7, func(time.Time) tea.Msg {
		return spinnerTick{}
	})
}

func (m *EditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinnerTick:
		if m.executing {
			m.spinnerFrame = (m.spinnerFrame + 1) % len(spinnerFrames)
		}
		return m, tea.Tick(time.Second/7, func(time.Time) tea.Msg {
			return spinnerTick{}
		})
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if m.mode == normalMode {
			return m.updateNormal(msg)
		}
		return m.updateInsert(msg)
	}
	return m, nil
}

func (m *EditorModel) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.pendingKey != "" {
		return m.resolvePending(msg)
	}

	switch msg.String() {
	case "h":
		m.cursorCol = max(0, m.cursorCol-1)
		m.preferCol = m.cursorCol
	case "l":
		if m.cursorRow < len(m.lines) {
			m.cursorCol = min(len(m.lines[m.cursorRow]), m.cursorCol+1)
		}
		m.preferCol = m.cursorCol
	case "j":
		m.cursorDown()
	case "k":
		m.cursorUp()
	case "g":
		m.pendingKey = "g"
		return m, nil
	case "G":
		m.cursorRow = max(0, len(m.lines)-1)
		m.cursorCol = 0
		m.clampCursor()
		m.preferCol = m.cursorCol
	case "0":
		m.cursorCol = 0
		m.preferCol = 0
	case "$":
		if m.cursorRow < len(m.lines) {
			m.cursorCol = max(0, len(m.lines[m.cursorRow]))
		}
		m.preferCol = m.cursorCol
	case "i":
		m.mode = insertMode
		m.completion.Update(m.text())
	case "a":
		m.mode = insertMode
		if m.cursorRow < len(m.lines) && m.cursorCol < len(m.lines[m.cursorRow]) {
			m.cursorCol++
		}
		m.preferCol = m.cursorCol
		m.completion.Update(m.text())
	case "A":
		m.mode = insertMode
		if m.cursorRow < len(m.lines) {
			m.cursorCol = len(m.lines[m.cursorRow])
		}
		m.preferCol = m.cursorCol
		m.completion.Update(m.text())
	case "o":
		m.mode = insertMode
		m.openLineBelow()
	case "O":
		m.mode = insertMode
		m.openLineAbove()
	case "d":
		m.pendingKey = "d"
		return m, nil
	case "x":
		m.deleteCharUnderCursor()
	case "ctrl+r":
		if !m.executing && m.canExecute() {
			m.executing = true
			return m, m.executeQuery
		}
	case "/":
		m.mode = insertMode
		m.completion.Update(m.text())
	}
	m.scrollToCursor()
	return m, nil
}

func (m *EditorModel) resolvePending(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	pending := m.pendingKey
	m.pendingKey = ""
	switch pending {
	case "g":
		if msg.String() == "g" {
			m.cursorRow = 0
			m.cursorCol = 0
			m.clampCursor()
			m.preferCol = 0
			m.scrollToCursor()
		}
	case "d":
		if msg.String() == "d" {
			m.deleteCurrentLine()
		}
	}
	return m, nil
}

func (m *EditorModel) updateInsert(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.completion.Visible {
			m.completion.Dismiss()
			return m, nil
		}
		m.mode = normalMode
		m.completion.Dismiss()
		if m.cursorCol > 0 {
			m.cursorCol--
		}
		m.preferCol = m.cursorCol
		return m, nil
	case "tab":
		if m.completion.Visible {
			m.acceptCompletion()
		} else {
			m.typeText("    ")
		}
		return m, nil
	case "shift+tab":
		if m.completion.Visible {
			m.completion.Cycle(-1)
			return m, nil
		}
	case "up":
		if m.completion.Visible {
			m.completion.Cycle(-1)
			return m, nil
		}
		m.cursorUp()
	case "down":
		if m.completion.Visible {
			m.completion.Cycle(1)
			return m, nil
		}
		m.cursorDown()
	case "left":
		m.cursorCol = max(0, m.cursorCol-1)
		m.preferCol = m.cursorCol
	case "right":
		if m.cursorRow < len(m.lines) {
			m.cursorCol = min(len(m.lines[m.cursorRow]), m.cursorCol+1)
		}
		m.preferCol = m.cursorCol
	case "enter":
		m.splitLine()
	case "backspace", "ctrl+h":
		m.deleteCharBefore()
	case "ctrl+r":
		if !m.executing && m.canExecute() {
			m.executing = true
			return m, m.executeQuery
		}
	default:
		if len(msg.Runes) > 0 {
			m.typeText(string(msg.Runes))
		}
	}
	m.scrollToCursor()
	m.completion.Update(m.text())
	return m, nil
}

func (m *EditorModel) text() string {
	return strings.Join(m.lines, "\n")
}

func (m *EditorModel) canExecute() bool {
	return m.driver != nil && strings.TrimSpace(m.text()) != ""
}

func (m *EditorModel) typeText(s string) {
	if len(m.lines) == 0 {
		m.lines = []string{s}
		m.cursorRow = 0
		m.cursorCol = len(s)
		return
	}
	line := m.lines[m.cursorRow]
	before := line[:m.cursorCol]
	after := ""
	if m.cursorCol < len(line) {
		after = line[m.cursorCol:]
	}
	m.lines[m.cursorRow] = before + s + after
	m.cursorCol += len(s)
	m.preferCol = m.cursorCol
}

func (m *EditorModel) splitLine() {
	if len(m.lines) == 0 {
		m.lines = []string{"", ""}
		m.cursorRow = 1
		m.cursorCol = 0
		return
	}
	line := m.lines[m.cursorRow]
	before := line[:m.cursorCol]
	after := ""
	if m.cursorCol < len(line) {
		after = line[m.cursorCol:]
	}
	m.lines[m.cursorRow] = before
	newLines := make([]string, len(m.lines)+1)
	copy(newLines, m.lines[:m.cursorRow+1])
	newLines[m.cursorRow+1] = after
	copy(newLines[m.cursorRow+2:], m.lines[m.cursorRow+1:])
	m.lines = newLines
	m.cursorRow++
	m.cursorCol = 0
	m.preferCol = 0
}

func (m *EditorModel) deleteCharBefore() {
	if m.cursorCol > 0 {
		line := m.lines[m.cursorRow]
		m.lines[m.cursorRow] = line[:m.cursorCol-1] + line[m.cursorCol:]
		m.cursorCol--
		m.preferCol = m.cursorCol
	} else if m.cursorRow > 0 {
		prevLine := m.lines[m.cursorRow-1]
		joinCol := len(prevLine)
		m.lines[m.cursorRow-1] = prevLine + m.lines[m.cursorRow]
		m.lines = append(m.lines[:m.cursorRow], m.lines[m.cursorRow+1:]...)
		m.cursorRow--
		m.cursorCol = joinCol
		m.preferCol = m.cursorCol
	}
}

func (m *EditorModel) deleteCharUnderCursor() {
	if m.cursorRow >= len(m.lines) {
		return
	}
	line := m.lines[m.cursorRow]
	if m.cursorCol < len(line) {
		m.lines[m.cursorRow] = line[:m.cursorCol] + line[m.cursorCol+1:]
	} else if m.cursorRow < len(m.lines)-1 {
		m.lines[m.cursorRow] = line + m.lines[m.cursorRow+1]
		m.lines = append(m.lines[:m.cursorRow+1], m.lines[m.cursorRow+2:]...)
	}
}

func (m *EditorModel) deleteCurrentLine() {
	if len(m.lines) == 1 {
		m.lines[0] = ""
		m.cursorCol = 0
		m.preferCol = 0
		return
	}
	m.lines = append(m.lines[:m.cursorRow], m.lines[m.cursorRow+1:]...)
	if m.cursorRow >= len(m.lines) {
		m.cursorRow = len(m.lines) - 1
	}
	m.clampCursor()
	m.preferCol = m.cursorCol
}

func (m *EditorModel) openLineBelow() {
	if len(m.lines) == 0 {
		m.lines = []string{"", ""}
		m.cursorRow = 1
		m.cursorCol = 0
		m.preferCol = 0
		return
	}
	newLines := make([]string, len(m.lines)+1)
	copy(newLines, m.lines[:m.cursorRow+1])
	newLines[m.cursorRow+1] = ""
	copy(newLines[m.cursorRow+2:], m.lines[m.cursorRow+1:])
	m.lines = newLines
	m.cursorRow++
	m.cursorCol = 0
	m.preferCol = 0
}

func (m *EditorModel) openLineAbove() {
	newLines := make([]string, len(m.lines)+1)
	copy(newLines, m.lines[:m.cursorRow])
	newLines[m.cursorRow] = ""
	copy(newLines[m.cursorRow+1:], m.lines[m.cursorRow:])
	m.lines = newLines
	m.cursorCol = 0
	m.preferCol = 0
}

func (m *EditorModel) acceptCompletion() {
	if !m.completion.Visible || len(m.completion.Suggestions) == 0 {
		return
	}
	replacement := m.completion.Suggestions[m.completion.SelectedIndex].Text
	line := m.lines[m.cursorRow]
	prefixLen := m.completion.PrefixStart
	if m.completion.PrefixStart > len(line) {
		prefixLen = len(line)
	}
	before := ""
	if prefixLen <= len(line) {
		before = line[:prefixLen]
	}
	after := ""
	if m.cursorCol < len(line) {
		after = line[m.cursorCol:]
	}
	m.lines[m.cursorRow] = before + replacement + after
	m.cursorCol = len(before) + len(replacement)
	m.preferCol = m.cursorCol
	m.completion.Dismiss()
	m.completion.Update(m.text())
}

func (m *EditorModel) cursorDown() {
	if m.cursorRow < len(m.lines)-1 {
		m.cursorRow++
		m.clampCursorToLine()
	}
}

func (m *EditorModel) cursorUp() {
	if m.cursorRow > 0 {
		m.cursorRow--
		m.clampCursorToLine()
	}
}

func (m *EditorModel) clampCursorToLine() {
	if m.cursorRow >= len(m.lines) {
		m.cursorRow = len(m.lines) - 1
	}
	lineLen := 0
	if m.cursorRow < len(m.lines) {
		lineLen = len(m.lines[m.cursorRow])
	}
	maxCol := lineLen
	if m.mode == normalMode {
		if lineLen > 0 {
			maxCol = lineLen - 1
		}
	}
	m.cursorCol = min(m.preferCol, maxCol)
	if m.cursorCol < 0 {
		m.cursorCol = 0
	}
}

func (m *EditorModel) clampCursor() {
	if len(m.lines) == 0 {
		m.lines = []string{""}
	}
	if m.cursorRow < 0 {
		m.cursorRow = 0
	}
	if m.cursorRow >= len(m.lines) {
		m.cursorRow = len(m.lines) - 1
	}
	m.clampCursorToLine()
}

func (m *EditorModel) scrollToCursor() {
	visibleHeight := max(1, m.height-6)
	if m.cursorRow < m.scrollRow {
		m.scrollRow = m.cursorRow
	}
	if m.cursorRow >= m.scrollRow+visibleHeight {
		m.scrollRow = m.cursorRow - visibleHeight + 1
	}
	if m.scrollRow < 0 {
		m.scrollRow = 0
	}
}

func (m *EditorModel) executeQuery() tea.Msg {
	if m.driver == nil {
		return editorQueryExecutedMsg{err: fmt.Errorf("no driver")}
	}
	sql := strings.TrimSpace(m.text())
	if sql == "" {
		return editorQueryExecutedMsg{}
	}

	results, rowCount, err := m.driver.ExecuteQuery(sql)
	return editorQueryExecutedMsg{results: results, rowCount: rowCount, err: err}
}

func (m *EditorModel) View() string {
	modeLabel := "-- NORMAL --"
	modeFg := AccentColor
	modeBg := OverlayColor
	if m.mode == insertMode {
		modeLabel = "-- INSERT --"
		modeFg = GreenColor
		modeBg = SurfaceColor
	}
	if m.executing {
		modeLabel = "EXECUTING..."
		modeFg = OrangeColor
		modeBg = SurfaceColor
	}

	editorContent := m.renderEditorContent()

	modeBox := lipgloss.NewStyle().
		Foreground(modeFg).
		Background(modeBg).
		Bold(true).
		Padding(0, 1).
		Render(modeLabel)

	posInfo := fmt.Sprintf("%d:%d", m.cursorRow+1, m.cursorCol+1)
	lineInfo := fmt.Sprintf("L%d/%d", m.cursorRow+1, len(m.lines))
	posStyle := lipgloss.NewStyle().Foreground(MutedTextColor)

	sqlHint := lipgloss.NewStyle().Foreground(MutedTextColor).Render("[Ctrl+R] Run  [Ctrl+E] Toggle")

	rightInfo := posStyle.Render(posInfo + "  " + lineInfo + "  " + sqlHint)
	if m.executing {
		loadingLabel := lipgloss.NewStyle().Foreground(OrangeColor).Bold(true).Render(spinnerFrames[m.spinnerFrame] + " Running query")
		bar := indeterminateProgressBar(max(12, min(24, m.width/4)), m.spinnerFrame, OrangeColor, SelectionColor)
		rightInfo = lipgloss.JoinHorizontal(lipgloss.Left, loadingLabel, "  ", bar)
	}

	statusBar := lipgloss.NewStyle().
		Background(BackgroundColor).
		Padding(0, 1).
		Width(max(1, m.width-2)).
		Render(modeBox + strings.Repeat(" ", max(1, m.width-lipgloss.Width(modeBox)-lipgloss.Width(rightInfo)-6)) + rightInfo)

	parts := []string{editorContent, statusBar}

	box := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(panelBorderColor(m.focused)).
		Width(max(1, m.width-2))

	return box.Render(lipgloss.JoinVertical(lipgloss.Left, parts...))
}

func (m *EditorModel) renderEditorContent() string {
	visibleHeight := max(1, m.height-5)
	renderLines := make([]string, 0, visibleHeight)

	dropdownLines := 0

	for i := 0; i < visibleHeight; i++ {
		lineIdx := m.scrollRow + i
		if lineIdx >= len(m.lines) {
			renderLines = append(renderLines, strings.Repeat(" ", max(1, m.width-4)))
			continue
		}

		lineNum := fmt.Sprintf("%4d ", lineIdx+1)
		lineNumStyle := lipgloss.NewStyle().Foreground(MutedBorderColor)
		if lineIdx == m.cursorRow {
			lineNumStyle = lipgloss.NewStyle().Foreground(AccentColor).Bold(true)
		}

		contentWidth := max(1, m.width-8)
		lineContent := m.lines[lineIdx]
		highlighted := highlightSQL(lineContent)
		displayLine := truncateDisplay(highlighted, contentWidth)

		if lineIdx == m.cursorRow {
			cursorCol := m.cursorCol
			plainRunes := []rune(m.lines[lineIdx])
			currentLineBg := lipgloss.NewStyle().Background(BackgroundColor)
			currentLineHighlight := lipgloss.NewStyle().Background(BackgroundColor)

			if m.mode == normalMode {
				if len(plainRunes) == 0 {
					displayLine = lipgloss.NewStyle().Background(SelectionColor).Render(" ")
				} else if cursorCol < len(plainRunes) {
					ch := string(plainRunes[cursorCol])
					before := ""
					if cursorCol > 0 {
						before = currentLineHighlight.Render(truncateDisplay(highlightSQL(string(plainRunes[:cursorCol])), contentWidth))
					}
					after := ""
					if cursorCol+1 < len(plainRunes) {
						after = currentLineHighlight.Render(truncateDisplay(highlightSQL(string(plainRunes[cursorCol+1:])), contentWidth))
					}
					cursorStyle := lipgloss.NewStyle().
						Background(SelectionColor).
						Foreground(PrimaryTextColor).Bold(true)
					selectedChar := cursorStyle.Render(ch)
					displayLine = before + selectedChar + after
				} else {
					before := currentLineHighlight.Render(truncateDisplay(highlightSQL(string(plainRunes)), contentWidth))
					eolMarker := lipgloss.NewStyle().Background(SelectionColor).Foreground(PrimaryTextColor).Bold(true).Render(" ")
					displayLine = before + eolMarker
				}
			} else {
				before := ""
				if cursorCol > 0 {
					before = currentLineBg.Render(truncateDisplay(highlightSQL(string(plainRunes[:cursorCol])), contentWidth))
				}
				after := ""
				if cursorCol < len(plainRunes) {
					after = currentLineBg.Render(truncateDisplay(highlightSQL(string(plainRunes[cursorCol:])), contentWidth))
				}
				cursor := lipgloss.NewStyle().Foreground(PrimaryTextColor).Bold(true).Render("▌")

				ghost := ""
				ghostSuffix := m.completion.GhostSuffix(m.text())
				if ghostSuffix != "" && lineIdx == m.cursorRow {
					ghostStyle := lipgloss.NewStyle().Foreground(MutedTextColor)
					ghost = ghostStyle.Render(truncateDisplay(ghostSuffix, max(1, contentWidth-lipgloss.Width(before+after))))
				}
				displayLine = before + after + cursor + ghost
			}
		}

		separator := lipgloss.NewStyle().Foreground(MutedBorderColor).Render("│")
		if lineIdx == m.cursorRow {
			separator = lipgloss.NewStyle().Foreground(AccentColor).Render("│")
		}

		renderLines = append(renderLines, lineNumStyle.Render(lineNum)+separator+" "+displayLine)

		if m.completion.Visible && lineIdx == m.cursorRow && dropdownLines > 0 {
			cursorScreenCol := 5 + min(m.cursorCol, contentWidth-1)
			dropdownMaxWidth := max(1, m.width-cursorScreenCol-2)
			dropdownWidth := min(24, dropdownMaxWidth)
			dropdown := renderCompletionDropdown(m.completion.Suggestions, m.completion.SelectedIndex, dropdownWidth)
			dropdownRows := strings.Split(dropdown, "\n")
			indent := strings.Repeat(" ", cursorScreenCol)
			linesToInsert := min(len(dropdownRows), visibleHeight-len(renderLines))
			for d := 0; d < linesToInsert; d++ {
				renderLines = append(renderLines, indent+dropdownRows[d])
			}
			dropdownLines = 0
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, renderLines...)
}

func (m EditorModel) Value() string {
	return m.text()
}

func (m EditorModel) SetValue(sql string) EditorModel {
	m.lines = strings.Split(sql, "\n")
	if len(m.lines) == 0 {
		m.lines = []string{""}
	}
	m.cursorRow = 0
	m.cursorCol = 0
	m.scrollRow = 0
	m.preferCol = 0
	m.clampCursor()
	return m
}
