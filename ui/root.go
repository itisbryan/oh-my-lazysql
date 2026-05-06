package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type RootModel struct {
	screen            Screen
	connectionList    tea.Model
	connectionForm    tea.Model
	home              tea.Model
	width             int
	height            int
}

func NewRootModel() *RootModel {
	return &RootModel{
		screen:         ScreenConnectionList,
		connectionList: NewConnectionListModel(),
	}
}

func (m *RootModel) Init() tea.Cmd {
	return m.connectionList.Init()
}

func (m *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case ScreenChangeMsg:
		m.screen = msg.Screen
		switch msg.Screen {
		case ScreenConnectionForm:
			m.connectionForm = NewConnectionFormModel(msg.Data)
			return m, m.connectionForm.Init()
		case ScreenHome:
			m.home = NewHomeModel(msg.Data)
			return m, m.home.Init()
		}
	}

	switch m.screen {
	case ScreenConnectionList:
		m.connectionList, cmd = m.connectionList.Update(msg)
	case ScreenConnectionForm:
		m.connectionForm, cmd = m.connectionForm.Update(msg)
	case ScreenHome:
		m.home, cmd = m.home.Update(msg)
	}

	return m, cmd
}

func (m *RootModel) View() string {
	switch m.screen {
	case ScreenConnectionList:
		return m.connectionList.View()
	case ScreenConnectionForm:
		return m.connectionForm.View()
	case ScreenHome:
		return m.home.View()
	default:
		return m.connectionList.View()
	}
}