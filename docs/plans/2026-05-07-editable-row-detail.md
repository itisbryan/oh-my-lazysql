# Editable Row Detail Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Improve the `o` row detail view with field selection and inline editing.

**Architecture:** Extend `ResultsModel` row-detail state with selected field index and edit mode/input. Reuse the existing pending edit map and cell commit behavior so detail edits update the same row data and save through the existing `Ctrl+R` path.

**Tech Stack:** Go, Bubble Tea, Lip Gloss.

---

### Task 1: Detail Field Navigation Tests

**Files:**
- Modify: `ui/results_test.go`

**Steps:**
1. Add a failing test that `j/k` move selected detail field.
2. Add a failing render assertion that selected field has visible content and footer hints.
3. Run `rtk go test ./ui`; expect failure.

### Task 2: Detail Field Editing Tests

**Files:**
- Modify: `ui/results_test.go`

**Steps:**
1. Add a failing test that `Enter` edits the selected detail field.
2. Type a new value and press `Enter`; assert row data updates and pending count increments.
3. Add an `Esc` cancel assertion.

### Task 3: Implementation

**Files:**
- Modify: `ui/results.go`

**Steps:**
1. Add `detailRow`, `detailEditing`, and `detailInput` state.
2. Route keys while `showRowDetail` is true: `j/k`, `Enter`, `Esc`, printable input, backspace.
3. Commit detail edits into `rows[row][detailRow]` and `pendingEdits`.

### Task 4: Visual Polish

**Files:**
- Modify: `ui/results.go`

**Steps:**
1. Render selected detail field with a full-row highlight.
2. Render editing value with readable foreground and cursor.
3. Show pending count and footer hints: `j/k field  Enter edit  Esc close  Ctrl+R save`.

### Task 5: Verification

**Steps:**
1. Run `rtk go build -o lazysql .`.
2. Run `rtk go test ./...`.
