# Vertical Row Detail Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a vertical detail view for the currently highlighted records row.

**Architecture:** `ResultsModel` owns a `showRowDetail` boolean. Pressing `o` on the Records tab opens a detail panel rendered instead of the grid content; `Esc` closes it. The panel reuses current `columns` and selected `rows[row]` data, showing one field per line with column metadata badges.

**Tech Stack:** Go, Bubble Tea, Lip Gloss.

---

### Task 1: Tests

**Files:**
- Modify: `ui/results_test.go`

**Steps:**
1. Add a failing test that pressing `o` opens row detail and renders `Row 1 Detail`, column names, and values.
2. Add a failing test that pressing `Esc` closes row detail.
3. Run `rtk go test ./ui` and confirm failure before implementation.

### Task 2: State and Key Handling

**Files:**
- Modify: `ui/results.go`

**Steps:**
1. Add `showRowDetail bool` to `ResultsModel`.
2. Handle `o` to open detail on Records tab.
3. Handle `Esc` to close detail before normal grid/filter behavior.

### Task 3: Detail Rendering

**Files:**
- Modify: `ui/results.go`

**Steps:**
1. Add `renderRowDetail(width, height int) string`.
2. Render a bordered panel with title, `column`, `type/key` badges, and full cell value.
3. Show `Esc close` footer hint.
4. Ensure NULL/EMPTY/DEFAULT formatting remains readable.

### Task 4: Verification

**Steps:**
1. Run `rtk go build -o lazysql .`.
2. Run `rtk go test ./...`.
