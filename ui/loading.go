package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func indeterminateProgressBar(width, frame int, active, track lipgloss.Color) string {
	width = max(8, width)
	segment := max(3, width/5)
	travel := width + segment
	head := (frame * 2) % travel
	start := head - segment

	var b strings.Builder
	for i := 0; i < width; i++ {
		if i >= start && i < head {
			b.WriteString(lipgloss.NewStyle().Foreground(active).Render("█"))
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(track).Render("─"))
		}
	}
	return b.String()
}
