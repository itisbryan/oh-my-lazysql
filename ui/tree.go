package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
}

type TreeModel struct {
	root       *TreeNode
	cursor     int
	flattened  []FlatNode
	width      int
	height     int
	focused    bool
	status     string
}

var treeIcons = map[TreeNodeType]string{
	NodeTypeDatabase:  "[D]",
	NodeTypeSection:   "[S]",
	NodeTypeTable:     "[T]",
	NodeTypeColumn:    "[C]",
	NodeTypeFunction:  "[F]",
	NodeTypeProcedure: "[P]",
	NodeTypeView:      "[V]",
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
	}
}

func (m *TreeModel) Init() tea.Cmd {
	return nil
}

func (m *TreeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.flattened)-1 {
				m.cursor++
			}
		case "enter", "right", "l":
			m.toggleExpanded(m.cursor)
		case "left", "h":
			if m.cursor > 0 {
				m.cursor--
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *TreeModel) toggleExpanded(idx int) {
	if idx < 0 || idx >= len(m.flattened) {
		return
	}
	_ = idx
}

func (m *TreeModel) View() string {
	if len(m.flattened) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Render("  No databases found")
	}

	lines := make([]string, 0, len(m.flattened))
	for i, node := range m.flattened {
		indent := strings.Repeat("  ", node.Depth)
		icon := treeIcons[node.Type]

		prefix := "  "
		if i == m.cursor {
			prefix = lipgloss.NewStyle().
				Background(lipgloss.Color("#0000FF")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Render(">")
		}

		style := lipgloss.NewStyle()
		switch node.Type {
		case NodeTypeDatabase:
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00BFFF"))
		case NodeTypeSection:
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Bold(true)
		case NodeTypeTable:
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
		case NodeTypeColumn:
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
		case NodeTypeFunction:
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#9370DB"))
		case NodeTypeProcedure:
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
		case NodeTypeView:
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#32CD32"))
		}

		line := fmt.Sprintf("%s%s %s %s", prefix, indent, icon, style.Render(node.Name))
		lines = append(lines, line)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)

	border := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666A7E")).
		Width(m.width - 2).
		Height(m.height - 2)
	return border.Render(content)
}

func (m *TreeModel) SetDatabases(databases []string) {
	m.root.Children = make([]*TreeNode, 0, len(databases))
	for _, db := range databases {
		m.root.Children = append(m.root.Children, &TreeNode{
			Type:     NodeTypeDatabase,
			Name:     db,
			Children: []*TreeNode{},
			Expanded: false,
		})
	}
	m.rebuildFlattened()
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
		})
	}

	if !node.Expanded && node.Type != NodeTypeRoot {
		return
	}

	for _, child := range node.Children {
		m.flattenNode(child, depth+1)
	}
}

func (m *TreeModel) SelectedNode() (nodeType TreeNodeType, database, name string) {
	if m.cursor < 0 || m.cursor >= len(m.flattened) {
		return NodeTypeRoot, "", ""
	}
	node := m.flattened[m.cursor]
	return node.Type, node.Database, node.Name
}