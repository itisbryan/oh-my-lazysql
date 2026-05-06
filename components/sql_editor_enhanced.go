package components

import (
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/jorgerojas26/lazysql/app"
)

type acItem struct {
	text string
	icon string
}

type acState struct {
	active   bool
	prefix   string
	matches  []acItem
	selected int
}

type SyntaxHighlightedEditor struct {
	*tview.Box
	text             string
	placeholder      string
	cursorX          int
	cursorY          int
	scrollY          int
	scrollX          int
	inputCapture     func(*tcell.EventKey) *tcell.EventKey
	tokens           []sqlToken
	ac               acState
	extraCompletions []string
	blinkFrame       int
}

func NewSyntaxHighlightedEditor() *SyntaxHighlightedEditor {
	e := &SyntaxHighlightedEditor{
		Box: tview.NewBox(),
	}
	e.Box.SetBorder(true)
	return e
}

func (e *SyntaxHighlightedEditor) GetText() string {
	return e.text
}

func (e *SyntaxHighlightedEditor) SetText(text string) {
	e.text = text
	e.tokens = highlightSQL(text)
	e.clampCursor()
}

func (e *SyntaxHighlightedEditor) SetPlaceholder(text string) {
	e.placeholder = text
}

func (e *SyntaxHighlightedEditor) SetExtraCompletions(items []string) {
	e.extraCompletions = items
}

func (e *SyntaxHighlightedEditor) SetInputCapture(f func(*tcell.EventKey) *tcell.EventKey) {
	e.inputCapture = f
}

func (e *SyntaxHighlightedEditor) Draw(screen tcell.Screen) {
	e.Box.DrawForSubclass(screen, e)
	x, y, width, height := e.Box.GetInnerRect()

	if e.text == "" && e.placeholder != "" && !e.Box.HasFocus() {
		tview.Print(screen, e.placeholder, x, y, width, tview.AlignLeft, app.Styles.InverseTextColor)
		return
	}

	lines := strings.Split(e.text, "\n")

	for lineIdx := 0; lineIdx < height && (lineIdx+e.scrollY) < len(lines); lineIdx++ {
		absLine := lineIdx + e.scrollY
		if absLine >= len(lines) {
			break
		}
		lineStr := lines[absLine]
		drawY := y + lineIdx
		drawX := x

		lineStart := byteOffsetForLine(e.text, absLine)
		lineEnd := lineStart + uint32(len(lineStr))

		colored := false
		for _, tok := range e.tokens {
			if tok.end <= lineStart || tok.start >= lineEnd {
				continue
			}
			colored = true
			tokStart := maxU32(tok.start, lineStart) - lineStart
			tokEnd := minU32(tok.end, lineEnd) - lineStart

			if int(tokStart) > len(lineStr) {
				tokStart = uint32(len(lineStr))
			}
			if int(tokEnd) > len(lineStr) {
				tokEnd = uint32(len(lineStr))
			}

			color := sqlTokenColors[tok.kind]
			if color == tcell.ColorDefault {
				color = app.Styles.PrimaryTextColor
			}

			segment := lineStr[tokStart:tokEnd]
			printX := drawX + utf8.RuneCountInString(lineStr[:tokStart])
			if e.scrollX > printX-drawX {
				continue
			}
			tview.Print(screen, segment, printX-e.scrollX, drawY, width-(printX-e.scrollX-drawX), tview.AlignLeft, color)
		}

		if !colored {
			tview.Print(screen, lineStr, drawX-e.scrollX, drawY, width, tview.AlignLeft, app.Styles.PrimaryTextColor)
		}

	}

	lineInView := e.cursorY >= e.scrollY && e.cursorY < e.scrollY+height

	e.blinkFrame++
	showCursor := e.blinkFrame%20 < 10

	if lineInView && showCursor {
		cy := y + (e.cursorY - e.scrollY)
		cx := x + e.cursorX - e.scrollX
		if cx >= x && cx < x+width {
			char := cursorChar(lines[e.cursorY], e.cursorX)
			style := tcell.StyleDefault.Background(tview.Styles.PrimitiveBackgroundColor).Foreground(tview.Styles.PrimitiveBackgroundColor)
			screen.SetContent(cx, cy, ' ', nil, style)
			screen.SetContent(cx, cy, char, nil, style)
		}
	}

	if lineInView {
		cy := y + (e.cursorY - e.scrollY)
		cx := x + e.cursorX - e.scrollX
		screen.ShowCursor(cx, cy)
	}

	e.drawAutocomplete(screen, x, y, width, height)
}

func (e *SyntaxHighlightedEditor) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if e.inputCapture != nil {
			if captured := e.inputCapture(event); captured == nil {
				return
			}
		}

		switch event.Key() {
		case tcell.KeyLeft:
			e.dismissAutocomplete()
			e.moveCursorLeft()
		case tcell.KeyRight:
			e.dismissAutocomplete()
			e.moveCursorRight()
		case tcell.KeyUp:
			if e.ac.active {
				e.ac.selected--
				if e.ac.selected < 0 {
					e.ac.selected = len(e.ac.matches) - 1
				}
			} else {
				e.moveCursorUp()
			}
		case tcell.KeyDown:
			if e.ac.active {
				e.ac.selected++
				if e.ac.selected >= len(e.ac.matches) {
					e.ac.selected = 0
				}
			} else {
				e.moveCursorDown()
			}
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			e.deleteLeft()
			e.triggerAutocomplete()
		case tcell.KeyDelete:
			e.deleteRight()
			e.triggerAutocomplete()
		case tcell.KeyEnter:
			if e.ac.active {
				e.acceptAutocomplete()
			} else {
				e.insertAtCursor("\n")
			}
		case tcell.KeyTab:
			if e.ac.active {
				e.acceptAutocomplete()
			}
			return
		case tcell.KeyEscape:
			if e.ac.active {
				e.ac.active = false
				return
			}
			e.dismissAutocomplete()
		case tcell.KeyRune:
			e.insertAtCursor(string(event.Rune()))
			e.triggerAutocomplete()
		case tcell.KeyHome:
			e.dismissAutocomplete()
			e.cursorX = 0
		case tcell.KeyEnd:
			e.dismissAutocomplete()
			lines := strings.Split(e.text, "\n")
			if e.cursorY < len(lines) {
				e.cursorX = utf8.RuneCountInString(lines[e.cursorY])
			}
		case tcell.KeyPgUp:
			e.dismissAutocomplete()
			e.cursorY -= 10
			e.clampCursor()
		case tcell.KeyPgDn:
			e.dismissAutocomplete()
			e.cursorY += 10
			e.clampCursor()
		default:
			return
		}

		e.ensureCursorVisible()
	}
}

func (e *SyntaxHighlightedEditor) moveCursorLeft() {
	if e.cursorX > 0 {
		e.cursorX--
	} else if e.cursorY > 0 {
		e.cursorY--
		lines := strings.Split(e.text, "\n")
		if e.cursorY < len(lines) {
			e.cursorX = utf8.RuneCountInString(lines[e.cursorY])
		}
	}
}

func (e *SyntaxHighlightedEditor) moveCursorRight() {
	lines := strings.Split(e.text, "\n")
	if e.cursorY < len(lines) {
		lineLen := utf8.RuneCountInString(lines[e.cursorY])
		if e.cursorX < lineLen {
			e.cursorX++
		} else if e.cursorY < len(lines)-1 {
			e.cursorY++
			e.cursorX = 0
		}
	}
}

func (e *SyntaxHighlightedEditor) moveCursorUp() {
	if e.cursorY > 0 {
		e.cursorY--
		e.clampCursorX()
	}
}

func (e *SyntaxHighlightedEditor) moveCursorDown() {
	lines := strings.Split(e.text, "\n")
	if e.cursorY < len(lines)-1 {
		e.cursorY++
		e.clampCursorX()
	}
}

func (e *SyntaxHighlightedEditor) insertAtCursor(s string) {
	lines := strings.Split(e.text, "\n")
	if e.cursorY >= len(lines) {
		lines = append(lines, "")
	}

	line := lines[e.cursorY]
	byteIdx := runeIndexToByte(line, e.cursorX)
	newLine := line[:byteIdx] + s + line[byteIdx:]
	lines[e.cursorY] = newLine
	e.text = strings.Join(lines, "\n")

	if s == "\n" {
		e.cursorY++
		e.cursorX = 0
	} else {
		e.cursorX += utf8.RuneCountInString(s)
	}

	e.tokens = highlightSQL(e.text)
}

func (e *SyntaxHighlightedEditor) deleteLeft() {
	if e.text == "" {
		return
	}
	lines := strings.Split(e.text, "\n")
	if e.cursorY >= len(lines) {
		return
	}

	if e.cursorX > 0 {
		line := lines[e.cursorY]
		byteIdx := runeIndexToByte(line, e.cursorX)
		_, size := utf8.DecodeLastRuneInString(line[:byteIdx])
		newLine := line[:byteIdx-size] + line[byteIdx:]
		lines[e.cursorY] = newLine
		e.cursorX--
	} else if e.cursorY > 0 {
		prevLine := lines[e.cursorY-1]
		prevLen := utf8.RuneCountInString(prevLine)
		lines[e.cursorY-1] = prevLine + lines[e.cursorY]
		lines = append(lines[:e.cursorY], lines[e.cursorY+1:]...)
		e.cursorY--
		e.cursorX = prevLen
	} else {
		return
	}

	e.text = strings.Join(lines, "\n")
	e.tokens = highlightSQL(e.text)
}

func (e *SyntaxHighlightedEditor) deleteRight() {
	lines := strings.Split(e.text, "\n")
	if e.cursorY >= len(lines) {
		return
	}

	line := lines[e.cursorY]
	byteIdx := runeIndexToByte(line, e.cursorX)

	if byteIdx < len(line) {
		_, size := utf8.DecodeRuneInString(line[byteIdx:])
		lines[e.cursorY] = line[:byteIdx] + line[byteIdx+size:]
	} else if e.cursorY < len(lines)-1 {
		lines[e.cursorY] = line + lines[e.cursorY+1]
		lines = append(lines[:e.cursorY+1], lines[e.cursorY+2:]...)
	} else {
		return
	}

	e.text = strings.Join(lines, "\n")
	e.tokens = highlightSQL(e.text)
}

func (e *SyntaxHighlightedEditor) clampCursor() {
	e.clampCursorY()
	e.clampCursorX()
}

func (e *SyntaxHighlightedEditor) clampCursorY() {
	lines := strings.Split(e.text, "\n")
	if e.cursorY >= len(lines) {
		e.cursorY = maxInt(0, len(lines)-1)
	}
	if e.cursorY < 0 {
		e.cursorY = 0
	}
}

func (e *SyntaxHighlightedEditor) clampCursorX() {
	lines := strings.Split(e.text, "\n")
	if e.cursorY >= len(lines) {
		return
	}
	lineLen := utf8.RuneCountInString(lines[e.cursorY])
	if e.cursorX > lineLen {
		e.cursorX = lineLen
	}
	if e.cursorX < 0 {
		e.cursorX = 0
	}
}

func (e *SyntaxHighlightedEditor) ensureCursorVisible() {
	_, _, _, height := e.Box.GetInnerRect()
	if e.cursorY < e.scrollY {
		e.scrollY = e.cursorY
	}
	if e.cursorY >= e.scrollY+height {
		e.scrollY = e.cursorY - height + 1
	}
}

func byteOffsetForLine(text string, lineIdx int) uint32 {
	if lineIdx <= 0 {
		return 0
	}
	offset := 0
	for i := 0; i < lineIdx; i++ {
		nl := strings.Index(text[offset:], "\n")
		if nl < 0 {
			return uint32(len(text))
		}
		offset += nl + 1
	}
	return uint32(offset)
}

func runeIndexToByte(s string, runeIdx int) int {
	pos := 0
	for i := 0; i < runeIdx && pos < len(s); i++ {
		_, size := utf8.DecodeRuneInString(s[pos:])
		pos += size
	}
	return pos
}

func cursorChar(line string, cursorX int) rune {
	pos := 0
	for i := 0; i < cursorX; i++ {
		if pos >= len(line) {
			return ' '
		}
		_, size := utf8.DecodeRuneInString(line[pos:])
		pos += size
	}
	if pos >= len(line) {
		return ' '
	}
	r, _ := utf8.DecodeRuneInString(line[pos:])
	return r
}

var sqlKeywords = []string{
	"SELECT", "FROM", "WHERE", "INSERT", "INTO", "VALUES", "UPDATE", "SET",
	"DELETE", "CREATE", "TABLE", "DROP", "ALTER", "ADD", "COLUMN", "INDEX",
	"VIEW", "FUNCTION", "PROCEDURE", "TRIGGER", "PRIMARY", "KEY", "FOREIGN",
	"REFERENCES", "NOT", "NULL", "DEFAULT", "UNIQUE", "CHECK", "CONSTRAINT",
	"JOIN", "LEFT", "RIGHT", "INNER", "OUTER", "FULL", "CROSS", "ON",
	"USING", "NATURAL", "AND", "OR", "IN", "EXISTS", "BETWEEN", "LIKE",
	"IS", "ORDER", "BY", "ASC", "DESC", "GROUP", "HAVING", "LIMIT",
	"OFFSET", "UNION", "ALL", "INTERSECT", "EXCEPT", "AS", "DISTINCT",
	"CASE", "WHEN", "THEN", "ELSE", "END", "BEGIN", "COMMIT", "ROLLBACK",
	"SAVEPOINT", "RELEASE", "GRANT", "REVOKE", "IF", "INT", "INTEGER",
	"BIGINT", "SMALLINT", "TINYINT", "FLOAT", "DOUBLE", "DECIMAL",
	"NUMERIC", "REAL", "CHAR", "VARCHAR", "TEXT", "BOOLEAN", "DATE",
	"TIMESTAMP", "DATETIME", "BLOB", "TRUE", "FALSE", "COUNT", "SUM",
	"AVG", "MIN", "MAX", "EXPLAIN", "DESCRIBE", "SHOW", "USE",
}

func (e *SyntaxHighlightedEditor) triggerAutocomplete() {
	lines := strings.Split(e.text, "\n")
	if e.cursorY >= len(lines) {
		e.ac.active = false
		return
	}
	line := lines[e.cursorY]
	byteIdx := runeIndexToByte(line, e.cursorX)

	prefix := ""
	for i := byteIdx - 1; i >= 0; i-- {
		c := line[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			prefix = string(c) + prefix
		} else {
			break
		}
	}

	if len(prefix) < 1 {
		e.ac.active = false
		return
	}

	upper := strings.ToUpper(prefix)
	seen := map[string]bool{}
	var matches []acItem

	for _, kw := range sqlKeywords {
		if strings.HasPrefix(kw, upper) && kw != upper {
			seen[kw] = true
			matches = append(matches, acItem{text: kw})
		}
	}

	for _, item := range e.extraCompletions {
		u := strings.ToUpper(item)
		if strings.HasPrefix(u, upper) && !seen[u] {
			seen[u] = true
			matches = append(matches, acItem{text: item, icon: treeIconTable})
		}
	}

	if len(matches) == 0 {
		e.ac.active = false
		return
	}

	e.ac = acState{
		active:   true,
		prefix:   prefix,
		matches:  matches,
		selected: 0,
	}
}

func (e *SyntaxHighlightedEditor) dismissAutocomplete() {
	e.ac.active = false
}

func (e *SyntaxHighlightedEditor) acceptAutocomplete() bool {
	if !e.ac.active || e.ac.selected >= len(e.ac.matches) {
		return false
	}
	completion := e.ac.matches[e.ac.selected].text

	lines := strings.Split(e.text, "\n")
	if e.cursorY >= len(lines) {
		return false
	}
	line := lines[e.cursorY]
	byteIdx := runeIndexToByte(line, e.cursorX)

	startByte := byteIdx - len(e.ac.prefix)
	rest := line[byteIdx:]
	newLine := line[:startByte] + completion + rest
	lines[e.cursorY] = newLine
	e.text = strings.Join(lines, "\n")
	e.cursorX += utf8.RuneCountInString(completion) - utf8.RuneCountInString(e.ac.prefix)

	e.tokens = highlightSQL(e.text)
	e.ac.active = false
	return true
}

const maxAutocompleteItems = 10

func (e *SyntaxHighlightedEditor) drawAutocomplete(screen tcell.Screen, x, y, width, height int) {
	if !e.ac.active {
		return
	}

	innerX, innerY, innerWidth, _ := e.Box.GetInnerRect()

	popupY := y + e.cursorY - e.scrollY + 1
	popupX := x + e.cursorX - e.scrollX

	if popupY >= innerY+height {
		popupY = innerY + height - 1
	}

	count := len(e.ac.matches)
	if count > maxAutocompleteItems {
		count = maxAutocompleteItems
	}

	maxW := 0
	for i := 0; i < count; i++ {
		displayW := len(e.ac.matches[i].text)
		if e.ac.matches[i].icon != "" {
			displayW += 2
		}
		if displayW > maxW {
			maxW = displayW
		}
	}
	popupW := maxW + 4

	if popupX+popupW > innerX+innerWidth {
		popupX = innerX + innerWidth - popupW
	}
	if popupX < innerX {
		popupX = innerX
	}

	for i := 0; i < count; i++ {
		bg := tcell.ColorDarkSlateGray
		fg := app.Styles.PrimaryTextColor
		if i == e.ac.selected {
			bg = app.Styles.SecondaryTextColor
			fg = tview.Styles.ContrastSecondaryTextColor
		}
		for j := 0; j < popupW; j++ {
			screen.SetContent(popupX+j, popupY+i, ' ', nil, tcell.StyleDefault.Background(bg).Foreground(fg))
		}
		item := e.ac.matches[i]
		display := item.text
		if item.icon != "" {
			display = item.icon + " " + item.text
		}
		tview.Print(screen, " "+display, popupX, popupY+i, popupW, tview.AlignLeft, fg)
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxU32(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

func minU32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}
