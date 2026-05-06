package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/jorgerojas26/lazysql/models"
)

type ProfileSelectorModel struct {
	profiles     []models.Profile
	activeIndex  int
	focused      bool
	width        int
}

func NewProfileSelector() ProfileSelectorModel {
	return ProfileSelectorModel{
		profiles:    []models.Profile{{Name: "default"}},
		activeIndex: 0,
	}
}

func (m ProfileSelectorModel) Init() {}

func (m ProfileSelectorModel) Update(msg any) (ProfileSelectorModel, any) {
	return m, nil
}

func (m ProfileSelectorModel) View() string {
	elements := make([]string, 0, len(m.profiles)*2+4)

	elements = append(elements, ProfileActiveStyle.Render("+"))
	elements = append(elements, " ")
	if len(m.profiles) > 1 {
		elements = append(elements, ErrorStyle.Render("-"))
		elements = append(elements, " ")
	}

	for i, p := range m.profiles {
		indicator := "○"
		if i == m.activeIndex {
			indicator = "●"
		}
		style := ProfileInactiveStyle
		if i == m.activeIndex {
			style = ProfileActiveStyle
		}
		elements = append(elements, style.Render(fmt.Sprintf("%s %s", indicator, p.Name)))
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, elements...)
}

func (m *ProfileSelectorModel) AddProfile(name string) {
	current := m.getCurrentFormValues()
	current.Name = name
	m.profiles = append(m.profiles[:m.activeIndex+1], current)
	m.activeIndex++
}

func (m *ProfileSelectorModel) DeleteActive() {
	if len(m.profiles) <= 1 {
		return
	}
	m.profiles = append(m.profiles[:m.activeIndex], m.profiles[m.activeIndex+1:]...)
	if m.activeIndex >= len(m.profiles) {
		m.activeIndex = len(m.profiles) - 1
	}
}

func (m *ProfileSelectorModel) SetProfiles(profiles []models.Profile) {
	m.profiles = profiles
	if len(m.profiles) == 0 {
		m.profiles = []models.Profile{{Name: "default"}}
	}
	m.activeIndex = 0
}

func (m *ProfileSelectorModel) GetActive() models.Profile {
	return m.profiles[m.activeIndex]
}

func (m *ProfileSelectorModel) getCurrentFormValues() models.Profile {
	return models.Profile{Name: "new"}
}