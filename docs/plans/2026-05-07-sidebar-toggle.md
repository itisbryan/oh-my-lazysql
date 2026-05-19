# Sidebar Toggle Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a `Ctrl+S` keybinding to hide/show the schema sidebar.

**Architecture:** `HomeModel` owns a session-local `showSidebar` flag. The update loop toggles it globally, moves focus to results when hiding the focused tree, and the view layout either renders tree + right panel or full-width right panel.

**Tech Stack:** Go, Bubble Tea, Lip Gloss, existing `HomeModel` layout.

---

### Task 1: Tests

**Files:**
- Modify: `ui/home_model_test.go`

**Steps:**
1. Add a test that `Ctrl+S` hides the sidebar and moves focus from tree to results.
2. Add a test that pressing `Ctrl+S` again shows the sidebar.
3. Run `rtk go test ./ui` and confirm failure before implementation.

### Task 2: Model State and Keybinding

**Files:**
- Modify: `ui/home_model.go`

**Steps:**
1. Add `showSidebar bool` to `HomeModel` and default it to `true` in `NewHomeModel`.
2. Handle `ctrl+s` in `HomeModel.Update` before focus routing.
3. When hiding while focused on tree, switch focus to results.

### Task 3: Layout

**Files:**
- Modify: `ui/home_model.go`

**Steps:**
1. If `showSidebar` is false, skip `tree.View()` and give the right panel the full width.
2. If `showSidebar` is true, preserve existing tree/right-panel layout.
3. Add `Ctrl+S sidebar` to the bottom statusline hints.

### Task 4: Verification

**Steps:**
1. Run `rtk go build -o lazysql .`.
2. Run `rtk go test ./...`.
