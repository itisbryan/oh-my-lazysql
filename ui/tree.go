package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jorgerojas26/lazysql/drivers"
	"github.com/jorgerojas26/lazysql/helpers/logger"
)

type TreeNodeType int

const (
	NodeTypeRoot TreeNodeType = iota
	NodeTypeDatabase
	NodeTypeSection
	NodeTypeTable
	NodeTypeColumn
	NodeTypeFunction
	NodeTypeProcedure
	NodeTypeView
)

type TreeNode struct {
	Type     TreeNodeType
	Database string
	Schema   string
	Name     string
	Children []*TreeNode
	Expanded bool
}

type FlatNode struct {
	Name     string
	Depth    int
	Type     TreeNodeType
	Database string
	Schema   string
}

type TreeModel struct {
	driver    drivers.Driver
	root      *TreeNode
	cursor    int
	flattened []FlatNode
	width     int
	height    int
	focused   bool
	status    string
	dbName    string
	viewport  viewport.Model
	searching bool
	search    string
}

var treeIcons = map[TreeNodeType]string{
	NodeTypeDatabase:  "󰆼",
	NodeTypeSection:   "󰙅",
	NodeTypeTable:     "󰓱",
	NodeTypeColumn:    "󰋫",
	NodeTypeFunction:  "󰊕",
	NodeTypeProcedure: "󰡱",
	NodeTypeView:      "󰈙",
}

var treeColors = map[TreeNodeType]string{
	NodeTypeDatabase:  "#00BFFF",
	NodeTypeSection:   "#FFFF00",
	NodeTypeTable:     "#FFFFFF",
	NodeTypeColumn:    "#888888",
	NodeTypeFunction:  "#9370DB",
	NodeTypeProcedure: "#FFA500",
	NodeTypeView:      "#32CD32",
}

func NewTreeModel() *TreeModel {
	return &TreeModel{
		root: &TreeNode{
			Type:     NodeTypeRoot,
			Name:     "databases",
			Children: []*TreeNode{},
			Expanded: true,
		},
		cursor:    0,
		flattened: []FlatNode{},
		focused:   true,
		viewport:  viewport.New(0, 0),
	}
}

func (m *TreeModel) SetDriver(driver drivers.Driver) {
	m.driver = driver
}

type databasesLoadedMsg struct {
	databases []string
	err       error
}

type tablesLoadedMsg struct {
	database string
	tables   map[string][]string
	err      error
}

func (m *TreeModel) Init() tea.Cmd {
	if m.driver == nil {
		logger.Warn("Tree Init: no driver set", nil)
		return nil
	}
	if m.dbName != "" {
		logger.Info("Tree Init: loading tables for specific database", map[string]any{"database": m.dbName})
		return func() tea.Msg {
			tables, err := m.driver.GetTables(m.dbName)
			logger.Info("Tree Init: GetTables result", map[string]any{"database": m.dbName, "count": len(tables), "error": err})
			return tablesLoadedMsg{database: m.dbName, tables: tables, err: err}
		}
	}
	logger.Info("Tree Init: loading all databases", nil)
	return func() tea.Msg {
		databases, err := m.driver.GetDatabases()
		logger.Info("Tree Init: GetDatabases result", map[string]any{"count": len(databases), "error": err})
		return databasesLoadedMsg{databases: databases, err: err}
	}
}

func (m *TreeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case databasesLoadedMsg:
		if msg.err != nil {
			m.status = "Error loading databases: " + msg.err.Error()
			logger.Error("databasesLoadedMsg", map[string]any{"error": msg.err})
		} else {
			logger.Info("databasesLoadedMsg", map[string]any{"count": len(msg.databases)})
			m.SetDatabases(msg.databases)
		}
	case tablesLoadedMsg:
		if msg.err != nil {
			m.status = "Error loading tables: " + msg.err.Error()
		} else {
			var dbNode *TreeNode
			for _, child := range m.root.Children {
				if child.Name == msg.database {
					dbNode = child
					break
				}
			}
			if dbNode == nil {
				dbNode = &TreeNode{
					Type:     NodeTypeDatabase,
					Name:     msg.database,
					Database: msg.database,
					Children: []*TreeNode{},
					Expanded: true,
				}
				m.root.Children = append(m.root.Children, dbNode)
			}
			m.setTables(dbNode, msg.database, msg.tables)
			m.rebuildFlattened()
		}
	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}
		if m.searching {
			m.updateSearch(msg)
			return m, nil
		}
		switch msg.String() {
		case "/":
			m.searching = true
			m.search = ""
			m.cursor = 0
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.visibleNodes())-1 {
				m.cursor++
			}
		case "enter":
			nodes := m.visibleNodes()
			if m.cursor < len(nodes) {
				node := nodes[m.cursor]
				m.toggleFlatExpanded(node)
				if node.Type == NodeTypeDatabase && m.driver != nil {
					dbName := node.Name
					return m, func() tea.Msg {
						tables, err := m.driver.GetTables(dbName)
						return tablesLoadedMsg{database: dbName, tables: tables, err: err}
					}
				}
			}
		case "left", "h":
			nodes := m.visibleNodes()
			if m.cursor < len(nodes) && m.isFlatExpanded(nodes[m.cursor]) {
				m.toggleFlatExpanded(nodes[m.cursor])
			} else if m.cursor > 0 {
				m.cursor--
			}
		case "right", "l":
			nodes := m.visibleNodes()
			if m.cursor < len(nodes) {
				node := nodes[m.cursor]
				if (node.Type == NodeTypeDatabase || node.Type == NodeTypeSection) && !m.isFlatExpanded(node) {
					m.toggleFlatExpanded(node)
					if node.Type == NodeTypeDatabase && m.driver != nil {
						dbName := node.Name
						return m, func() tea.Msg {
							tables, err := m.driver.GetTables(dbName)
							return tablesLoadedMsg{database: dbName, tables: tables, err: err}
						}
					}
				} else if m.cursor < len(nodes)-1 {
					m.cursor++
				}
			}
		}
	}
	return m, nil
}

func (m *TreeModel) updateSearch(msg tea.KeyMsg) {
	switch msg.String() {
	case "enter":
		m.searching = false
	case "esc":
		m.searching = false
		m.search = ""
		m.cursor = 0
	case "backspace", "ctrl+h":
		if len(m.search) > 0 {
			runes := []rune(m.search)
			m.search = string(runes[:len(runes)-1])
		}
		m.cursor = min(m.cursor, max(0, len(m.visibleNodes())-1))
	default:
		if len(msg.Runes) > 0 {
			m.search += string(msg.Runes)
			m.cursor = 0
		}
	}
}

func (m *TreeModel) visibleNodes() []FlatNode {
	if strings.TrimSpace(m.search) == "" {
		return m.flattened
	}

	query := strings.ToLower(strings.TrimSpace(m.search))
	nodes := []FlatNode{}
	for _, child := range m.root.Children {
		m.collectSearchNodes(child, 0, query, &nodes)
	}
	return nodes
}

func (m *TreeModel) collectSearchNodes(node *TreeNode, depth int, query string, nodes *[]FlatNode) bool {
	selfMatches := strings.Contains(strings.ToLower(node.Name), query)
	start := len(*nodes)
	childMatched := false
	for _, child := range node.Children {
		childDepth := depth + 1
		if m.collectSearchNodes(child, childDepth, query, nodes) {
			childMatched = true
		}
	}

	if !selfMatches && !childMatched {
		return false
	}

	flat := FlatNode{
		Name:     node.Name,
		Depth:    depth,
		Type:     node.Type,
		Database: node.Database,
		Schema:   node.Schema,
	}
	*nodes = append((*nodes)[:start], append([]FlatNode{flat}, (*nodes)[start:]...)...)
	return true
}

func (m *TreeModel) toggleExpanded(idx int) {
	if idx < 0 || idx >= len(m.flattened) {
		return
	}
	flat := m.flattened[idx]
	m.toggleFlatExpanded(flat)
}

func (m *TreeModel) toggleFlatExpanded(flat FlatNode) {
	node := m.findNode(m.root, flat)
	if node != nil && len(node.Children) > 0 {
		node.Expanded = !node.Expanded
	}
	m.rebuildFlattened()
}

func (m *TreeModel) isExpanded(idx int) bool {
	if idx < 0 || idx >= len(m.flattened) {
		return false
	}
	flat := m.flattened[idx]
	return m.isFlatExpanded(flat)
}

func (m *TreeModel) isFlatExpanded(flat FlatNode) bool {
	node := m.findNode(m.root, flat)
	if node != nil {
		return node.Expanded
	}
	return false
}

func (m *TreeModel) findNode(node *TreeNode, flat FlatNode) *TreeNode {
	if node.Type == flat.Type && node.Name == flat.Name && node.Database == flat.Database && node.Schema == flat.Schema {
		return node
	}
	for _, child := range node.Children {
		if found := m.findNode(child, flat); found != nil {
			return found
		}
	}
	return nil
}

func (m *TreeModel) View() string {
	if len(m.flattened) == 0 && m.status == "" {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Render("  Loading...")
	}

	if len(m.flattened) == 0 {
		return lipgloss.NewStyle().
			Foreground(ErrorColor).
			Render("  " + m.status)
	}

	borderColor := lipgloss.Color("#666A7E")
	titleColor := lipgloss.Color("#FFFFFF")
	if m.focused {
		borderColor = SecondaryTextColor
		titleColor = SecondaryTextColor
	}

	contentWidth := max(1, m.width-2)
	contentHeight := max(1, m.height-2)
	nodes := m.visibleNodes()
	if m.cursor >= len(nodes) {
		m.cursor = max(0, len(nodes)-1)
	}
	m.viewport.Width = contentWidth
	m.viewport.Height = contentHeight
	if m.cursor < m.viewport.YOffset {
		m.viewport.SetYOffset(m.cursor)
	} else if m.cursor >= m.viewport.YOffset+contentHeight {
		m.viewport.SetYOffset(m.cursor - contentHeight + 1)
	}

	lines := make([]string, 0, len(nodes)+1)
	if m.searching || m.search != "" {
		query := m.search
		if m.searching {
			query += "▌"
		}
		lines = append(lines, lipgloss.NewStyle().
			Foreground(SecondaryTextColor).
			Background(lipgloss.Color("#1F2335")).
			Padding(0, 1).
			Width(contentWidth).
			Render("/ "+query))
	}

	for i, node := range nodes {
		indent := strings.Repeat("  ", node.Depth)
		icon := treeIcons[node.Type]
		color := treeColors[node.Type]

		if !m.focused {
			color = "#888888"
		}

		expand := " "
		if node.Type == NodeTypeDatabase || node.Type == NodeTypeSection {
			if m.isFlatExpanded(node) {
				expand = "▾"
			} else {
				expand = "▸"
			}
		}

		availableNameWidth := max(1, contentWidth-lipgloss.Width(indent)-lipgloss.Width(expand)-lipgloss.Width(icon)-4)
		nameText := truncateDisplay(node.Name, availableNameWidth)
		name := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Inline(true).Render(nameText)
		line := fmt.Sprintf("%s%s %s %s", indent, expand, icon, name)

		if i == m.cursor && m.focused {
			padWidth := max(0, contentWidth-lipgloss.Width(line)-2)
			line = lipgloss.NewStyle().
				Background(lipgloss.Color("#283457")).
				Foreground(lipgloss.Color("#C0CAF5")).
				Inline(true).
				Render(" " + line + strings.Repeat(" ", padWidth) + " ")
		} else {
			line = truncateDisplay(line, contentWidth)
		}

		lines = append(lines, line)
	}

	m.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Left, lines...))

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(max(1, m.width-2)).
		Height(m.height - 2).
		Foreground(titleColor).
		Render(m.viewport.View())
}

func truncateDisplay(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	runes := []rune(s)
	for len(runes) > 0 && lipgloss.Width(string(runes))+1 > width {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}

func (m *TreeModel) SetDatabases(databases []string) {
	m.root.Children = make([]*TreeNode, 0, len(databases))
	for _, db := range databases {
		m.root.Children = append(m.root.Children, &TreeNode{
			Type:     NodeTypeDatabase,
			Name:     db,
			Database: db,
			Children: []*TreeNode{},
			Expanded: false,
		})
	}
	m.rebuildFlattened()
}

func (m *TreeModel) setTables(dbNode *TreeNode, database string, tables map[string][]string) {
	dbNode.Children = make([]*TreeNode, 0, len(tables))

	if m.driver != nil && m.driver.UseSchemas() {
		for schemaName, tableNames := range tables {
			schemaNode := &TreeNode{
				Type:     NodeTypeSection,
				Name:     schemaName,
				Database: database,
				Schema:   schemaName,
				Children: make([]*TreeNode, 0, len(tableNames)),
				Expanded: false,
			}
			for _, tableName := range tableNames {
				schemaNode.Children = append(schemaNode.Children, &TreeNode{
					Type:     NodeTypeTable,
					Name:     tableName,
					Database: database,
					Schema:   schemaName,
				})
			}
			dbNode.Children = append(dbNode.Children, schemaNode)
		}
		return
	}

	for _, tableNames := range tables {
		for _, tableName := range tableNames {
			dbNode.Children = append(dbNode.Children, &TreeNode{
				Type:     NodeTypeTable,
				Name:     tableName,
				Database: database,
				Children: []*TreeNode{},
				Expanded: false,
			})
		}
	}
}

func (m *TreeModel) LoadTables(dbName string) error {
	if m.driver == nil {
		return fmt.Errorf("no driver")
	}

	tables, err := m.driver.GetTables(dbName)
	if err != nil {
		return err
	}

	for _, child := range m.root.Children {
		if child.Name == dbName {
			m.setTables(child, dbName, tables)
			break
		}
	}

	m.rebuildFlattened()
	return nil
}

func (m *TreeModel) rebuildFlattened() {
	m.flattened = []FlatNode{}
	m.flattenNode(m.root, 0)
}

func (m *TreeModel) flattenNode(node *TreeNode, depth int) {
	if node.Type != NodeTypeRoot {
		m.flattened = append(m.flattened, FlatNode{
			Name:     node.Name,
			Depth:    depth,
			Type:     node.Type,
			Database: node.Database,
			Schema:   node.Schema,
		})
	}

	if !node.Expanded && node.Type != NodeTypeRoot {
		return
	}

	for _, child := range node.Children {
		childDepth := depth + 1
		if node.Type == NodeTypeRoot {
			childDepth = 0
		}
		m.flattenNode(child, childDepth)
	}
}

func (m *TreeModel) SelectedNode() (nodeType TreeNodeType, database, schema, name string) {
	nodes := m.visibleNodes()
	if m.cursor < 0 || m.cursor >= len(nodes) {
		return NodeTypeRoot, "", "", ""
	}
	node := nodes[m.cursor]
	return node.Type, node.Database, node.Schema, node.Name
}

func (m *TreeModel) TableNames() []string {
	names := []string{}
	for _, child := range m.root.Children {
		collectTableNames(child, &names)
	}
	return names
}

func collectTableNames(node *TreeNode, names *[]string) {
	if node.Type == NodeTypeTable {
		*names = append(*names, node.Name)
	}
	for _, child := range node.Children {
		collectTableNames(child, names)
	}
}
