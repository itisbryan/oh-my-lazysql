package components

import (
	"context"
	"sort"

	"github.com/gdamore/tcell/v2"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/sql"
)

type sqlTokenType int

const (
	sqlTokenKeyword sqlTokenType = iota
	sqlTokenTypeKeyword
	sqlTokenIdentifier
	sqlTokenString
	sqlTokenNumber
	sqlTokenFunction
	sqlTokenOperator
	sqlTokenWildcard
	sqlTokenNormal
)

type sqlToken struct {
	text  string
	kind  sqlTokenType
	start uint32
	end   uint32
}

var sqlTokenColors = map[sqlTokenType]tcell.Color{
	sqlTokenKeyword:     tcell.ColorCornflowerBlue,
	sqlTokenTypeKeyword: tcell.ColorDarkCyan,
	sqlTokenIdentifier:  tcell.ColorDefault,
	sqlTokenString:      tcell.ColorGreen,
	sqlTokenNumber:      tcell.ColorOrange,
	sqlTokenFunction:    tcell.ColorMediumPurple,
	sqlTokenOperator:    tcell.ColorGray,
	sqlTokenWildcard:    tcell.ColorYellow,
	sqlTokenNormal:      tcell.ColorDefault,
}

var (
	sqlParser     *sitter.Parser
	sqlParserOnce bool
)

func ensureSQLParser() {
	if !sqlParserOnce {
		sqlParser = sitter.NewParser()
		sqlParser.SetLanguage(sql.GetLanguage())
		sqlParserOnce = true
	}
}

func highlightSQL(input string) []sqlToken {
	ensureSQLParser()

	if len(input) == 0 {
		return nil
	}

	tree, err := sqlParser.ParseCtx(context.Background(), nil, []byte(input))
	if err != nil || tree == nil {
		return []sqlToken{{text: input, kind: sqlTokenNormal, start: 0, end: uint32(len(input))}}
	}

	var tokens []sqlToken
	collectTokens(tree.RootNode(), []byte(input), &tokens)

	sort.Slice(tokens, func(i, j int) bool {
		return tokens[i].start < tokens[j].start
	})

	var result []sqlToken
	pos := uint32(0)
	for _, t := range tokens {
		if t.start > pos {
			result = append(result, sqlToken{
				text:  input[pos:t.start],
				kind:  sqlTokenNormal,
				start: pos,
				end:   t.start,
			})
		}
		if t.end > pos {
			result = append(result, t)
			pos = t.end
		}
	}
	if pos < uint32(len(input)) {
		result = append(result, sqlToken{
			text:  input[pos:],
			kind:  sqlTokenNormal,
			start: pos,
			end:   uint32(len(input)),
		})
	}

	return result
}

func collectTokens(node *sitter.Node, input []byte, tokens *[]sqlToken) {
	if node == nil {
		return
	}
	if node.ChildCount() == 0 {
		kind := classifyToken(node)
		*tokens = append(*tokens, sqlToken{
			text:  node.Content(input),
			kind:  kind,
			start: node.StartByte(),
			end:   node.EndByte(),
		})
	} else {
		for i := 0; i < int(node.ChildCount()); i++ {
			collectTokens(node.Child(i), input, tokens)
		}
	}
}

func classifyToken(node *sitter.Node) sqlTokenType {
	switch node.Type() {
	case "keyword_select", "keyword_from", "keyword_where", "keyword_order",
		"keyword_by", "keyword_limit", "keyword_insert", "keyword_into",
		"keyword_values", "keyword_update", "keyword_set", "keyword_delete",
		"keyword_create", "keyword_table", "keyword_key", "keyword_primary",
		"keyword_join", "keyword_on", "keyword_as",
		"keyword_having", "keyword_group", "keyword_distinct", "keyword_exists",
		"keyword_not", "keyword_in", "keyword_is", "keyword_null",
		"keyword_and", "keyword_or", "keyword_between", "keyword_like",
		"keyword_drop", "keyword_alter", "keyword_add", "keyword_column",
		"keyword_index", "keyword_view", "keyword_function", "keyword_procedure",
		"keyword_return", "keyword_if", "keyword_else", "keyword_then",
		"keyword_when", "keyword_case", "keyword_begin", "keyword_end",
		"keyword_commit", "keyword_rollback", "keyword_union",
		"keyword_all", "keyword_intersect", "keyword_except",
		"keyword_asc", "keyword_desc", "keyword_default",
		"keyword_left", "keyword_right", "keyword_inner", "keyword_outer",
		"keyword_full", "keyword_using":
		return sqlTokenKeyword
	case "keyword_int", "keyword_text", "keyword_boolean",
		"keyword_float", "keyword_double", "keyword_decimal", "keyword_date",
		"keyword_timestamp", "keyword_serial", "keyword_bigint", "keyword_smallint",
		"keyword_char", "keyword_blob", "keyword_numeric", "keyword_real",
		"keyword_integer", "keyword_tinyint", "keyword_mediumint",
		"keyword_varchar":
		return sqlTokenTypeKeyword
	case "invocation":
		return sqlTokenFunction
	case "identifier", "object_reference", "field", "relation", "column":
		return sqlTokenIdentifier
	case "literal":
		return sqlTokenString
	case "int", "float":
		return sqlTokenNumber
	case "all_fields":
		return sqlTokenWildcard
	case "assignment":
		return sqlTokenOperator
	default:
		return sqlTokenNormal
	}
}
