package ui

import "testing"

func TestApplyThemeSwitchesPalette(t *testing.T) {
	previous := CurrentTheme.Name
	t.Cleanup(func() { ApplyTheme(previous) })

	applied := ApplyTheme("catppuccin")
	if applied != "catppuccin-mocha" {
		t.Fatalf("expected catppuccin alias to apply catppuccin-mocha, got %q", applied)
	}
	if CurrentTheme.Name != "catppuccin-mocha" {
		t.Fatalf("expected current theme catppuccin-mocha, got %q", CurrentTheme.Name)
	}
	if PrimaryTextColor != "#CDD6F4" {
		t.Fatalf("expected catppuccin primary text, got %s", PrimaryTextColor)
	}
}

func TestApplyThemeFallsBackToDefault(t *testing.T) {
	previous := CurrentTheme.Name
	t.Cleanup(func() { ApplyTheme(previous) })

	applied := ApplyTheme("missing-theme")
	if applied != "tokyonight" {
		t.Fatalf("expected fallback theme tokyonight, got %q", applied)
	}
}
