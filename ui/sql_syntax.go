package ui

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/sql"
)

var sqlKeywords = []string{
	"SELECT", "FROM", "WHERE", "AND", "OR", "ORDER", "BY", "GROUP", "HAVING",
	"LIMIT", "OFFSET", "INSERT", "UPDATE", "DELETE", "JOIN", "LEFT", "RIGHT", "INNER",
	"OUTER", "ON", "AS", "DISTINCT", "NULL", "IS", "NOT", "LIKE", "IN", "BETWEEN",
}

func sqlTreeSitterAvailable() bool {
	return sql.GetLanguage() != nil
}

func highlightSQL(source string) string {
	if source == "" {
		return ""
	}

	language := sql.GetLanguage()
	if language == nil {
		return source
	}
	root := sitter.Parse([]byte(source), language)
	if root == nil || root.IsNull() || root.HasError() {
		return highlightSQLLexical(source)
	}
	return highlightSQLLexical(source)
}

func highlightSQLLexical(sql string) string {
	var out strings.Builder
	for i := 0; i < len(sql); {
		r := rune(sql[i])
		if r == '\'' || r == '"' {
			quote := sql[i]
			start := i
			i++
			for i < len(sql) && sql[i] != quote {
				i++
			}
			if i < len(sql) {
				i++
			}
			out.WriteString(sqlStringStyle().Render(sql[start:i]))
			continue
		}

		if isSQLOperatorByte(sql[i]) {
			out.WriteString(sqlOperatorStyle().Render(sql[i : i+1]))
			i++
			continue
		}

		if unicode.IsLetter(r) || r == '_' {
			start := i
			for i < len(sql) {
				current := rune(sql[i])
				if !unicode.IsLetter(current) && !unicode.IsDigit(current) && current != '_' {
					break
				}
				i++
			}
			word := sql[start:i]
			if isSQLKeyword(word) {
				out.WriteString(sqlKeywordStyle().Render(word))
			} else {
				out.WriteString(sqlIdentifierStyle().Render(word))
			}
			continue
		}

		out.WriteByte(sql[i])
		i++
	}
	return out.String()
}

func isSQLKeyword(word string) bool {
	upper := strings.ToUpper(word)
	for _, keyword := range sqlKeywords {
		if upper == keyword {
			return true
		}
	}
	return false
}

func isSQLOperatorByte(b byte) bool {
	return strings.ContainsRune("=<>+-*/(),.;", rune(b))
}

type completionKind int

const (
	keywordSuggestion completionKind = iota
	tableSuggestion
	columnSuggestion
)

type completionItem struct {
	Text string
	Kind completionKind
}

type CompletionState struct {
	Visible       bool
	Suggestions   []completionItem
	SelectedIndex int
	Prefix        string
	PrefixStart   int
	TableNames    []string
	ColumnNames   []string
}

func NewCompletionState() CompletionState {
	return CompletionState{}
}

func (cs *CompletionState) Update(input string) {
	prefix := currentSQLPrefix(input)
	prefixStart := len(input) - len(prefix)
	suggestions := sqlCompletionItems(input, cs.TableNames, cs.ColumnNames)
	if len(suggestions) == 0 {
		cs.Visible = false
		cs.Suggestions = nil
		cs.SelectedIndex = 0
		cs.Prefix = ""
		cs.PrefixStart = 0
		return
	}
	cs.Prefix = prefix
	cs.PrefixStart = prefixStart
	cs.Suggestions = suggestions
	cs.Visible = true
	if cs.SelectedIndex >= len(suggestions) {
		cs.SelectedIndex = 0
	}
}

func (cs *CompletionState) Accept(input string) string {
	if !cs.Visible || len(cs.Suggestions) == 0 {
		return input
	}
	replacement := cs.Suggestions[cs.SelectedIndex].Text
	result := input[:cs.PrefixStart] + replacement
	cs.Visible = false
	cs.Suggestions = nil
	cs.SelectedIndex = 0
	cs.Prefix = ""
	cs.PrefixStart = 0
	return result
}

func (cs *CompletionState) Cycle(direction int) {
	if !cs.Visible || len(cs.Suggestions) == 0 {
		return
	}
	cs.SelectedIndex = (cs.SelectedIndex + direction + len(cs.Suggestions)) % len(cs.Suggestions)
}

func (cs *CompletionState) Dismiss() {
	cs.Visible = false
	cs.Suggestions = nil
	cs.SelectedIndex = 0
	cs.Prefix = ""
	cs.PrefixStart = 0
}

func (cs *CompletionState) GhostSuffix(input string) string {
	if !cs.Visible || len(cs.Suggestions) == 0 || cs.SelectedIndex >= len(cs.Suggestions) {
		return ""
	}
	suggestion := cs.Suggestions[cs.SelectedIndex].Text
	prefix := cs.Prefix
	if len(prefix) >= len(suggestion) {
		return ""
	}
	return suggestion[len(prefix):]
}

func sqlCompletionItems(input string, tableNames []string, columnNames []string) []completionItem {
	prefix := strings.ToUpper(currentSQLPrefix(input))
	if prefix == "" {
		return nil
	}
	items := []completionItem{}
	for _, keyword := range sqlKeywords {
		if strings.HasPrefix(keyword, prefix) {
			items = append(items, completionItem{Text: keyword, Kind: keywordSuggestion})
		}
	}
	for _, table := range tableNames {
		if strings.HasPrefix(strings.ToUpper(table), prefix) {
			items = append(items, completionItem{Text: table, Kind: tableSuggestion})
		}
	}
	for _, col := range columnNames {
		if strings.HasPrefix(strings.ToUpper(col), prefix) {
			items = append(items, completionItem{Text: col, Kind: columnSuggestion})
		}
	}
	return items
}

func sqlAutocompleteSuggestions(input string) []string {
	prefix := strings.ToUpper(currentSQLPrefix(input))
	if prefix == "" {
		return nil
	}
	suggestions := []string{}
	for _, keyword := range sqlKeywords {
		if strings.HasPrefix(keyword, prefix) {
			suggestions = append(suggestions, keyword)
		}
	}
	return suggestions
}

func applySQLAutocomplete(input string) string {
	suggestions := sqlAutocompleteSuggestions(input)
	if len(suggestions) == 0 {
		return input
	}
	prefix := currentSQLPrefix(input)
	return input[:len(input)-len(prefix)] + suggestions[0]
}

func currentSQLPrefix(input string) string {
	end := len(input)
	start := end
	for start > 0 {
		r := rune(input[start-1])
		if !unicode.IsLetter(r) && r != '_' {
			break
		}
		start--
	}
	return input[start:end]
}

func renderCompletionDropdown(items []completionItem, selectedIndex int, maxWidth int) string {
	if len(items) == 0 {
		return ""
	}
	maxVisible := 6
	if len(items) < maxVisible {
		maxVisible = len(items)
	}
	startIdx := 0
	if selectedIndex >= maxVisible {
		startIdx = selectedIndex - maxVisible + 1
	}

	itemWidth := 4
	for _, it := range items {
		labelW := len(it.Text) + 4
		if labelW > itemWidth {
			itemWidth = labelW
		}
	}
	if itemWidth > maxWidth-2 {
		itemWidth = maxWidth - 2
	}
	if itemWidth < 8 {
		itemWidth = 8
	}

	lines := make([]string, 0, maxVisible)
	for i := 0; i < maxVisible; i++ {
		idx := startIdx + i
		if idx >= len(items) {
			break
		}
		item := items[idx]

		var kindBadge string
		switch item.Kind {
		case keywordSuggestion:
			kindBadge = lipgloss.NewStyle().
				Foreground(InverseTextColor).
				Background(CyanColor).
				Bold(true).
				Padding(0, 1).
				Render("K")
		case tableSuggestion:
			kindBadge = lipgloss.NewStyle().
				Foreground(InverseTextColor).
				Background(GreenColor).
				Bold(true).
				Padding(0, 1).
				Render("T")
		case columnSuggestion:
			kindBadge = lipgloss.NewStyle().
				Foreground(InverseTextColor).
				Background(OrangeColor).
				Bold(true).
				Padding(0, 1).
				Render("C")
		}

		textWidth := itemWidth - lipgloss.Width(kindBadge) - 1
		if textWidth < 2 {
			textWidth = 2
		}
		text := truncateDisplay(item.Text, textWidth)

		if idx == selectedIndex {
			textRendered := lipgloss.NewStyle().
				Foreground(PrimaryTextColor).
				Background(SelectionColor).
				Width(textWidth).
				Render(text)
			lines = append(lines, textRendered+" "+kindBadge)
		} else {
			textRendered := lipgloss.NewStyle().
				Foreground(PrimaryTextColor).
				Background(BackgroundColor).
				Width(textWidth).
				Render(text)
			lines = append(lines, textRendered+" "+kindBadge)
		}
	}

	box := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(MutedBorderColor).
		Background(BackgroundColor).
		Padding(0, 1).
		Width(max(1, itemWidth+4)).
		Render(lipgloss.JoinVertical(lipgloss.Left, lines...))

	if len(items) > maxVisible {
		more := lipgloss.NewStyle().
			Foreground(MutedTextColor).
			Render(fmt.Sprintf("  ↓ %d more", len(items)-maxVisible))
		box = lipgloss.JoinVertical(lipgloss.Left, box, more)
	}
	return box
}

func sqlKeywordStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(CyanColor).Bold(true)
}

func sqlIdentifierStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(PrimaryTextColor)
}

func sqlStringStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GreenColor)
}

func sqlOperatorStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(OrangeColor)
}
