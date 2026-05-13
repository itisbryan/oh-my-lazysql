package ui

import (
	"strings"
	"testing"

	"github.com/jorgerojas26/lazysql/models"
)

func TestConnectionListRendersRetroWelcomeScreen(t *testing.T) {
	model := &ConnectionListModel{
		connections: []models.Connection{
			{Name: "local", Provider: "PostgreSQL", DBName: "app_development"},
		},
		width:  100,
		height: 30,
	}

	view := model.View()

	for _, expected := range []string{
		"om-lazysql",
		"database console",
		"BOOT MENU",
		"> local",
		"N:new",
		"ENTER:connect",
	} {
		if !strings.Contains(view, expected) {
			t.Fatalf("expected view to contain %q\n%s", expected, view)
		}
	}

	if strings.Contains(view, "SYSTEM") {
		t.Fatal("expected no SYSTEM panel after redesign")
	}
}
