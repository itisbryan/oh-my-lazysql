# Inline Pending Edits Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add inline cell editing with pending changes and `Ctrl+R` save.

**Architecture:** `ResultsModel` owns edit state and a pending-change map keyed by row/column. Confirming an edit updates the visible cell and marks it dirty locally. `HomeModel` converts pending edits into `models.DBDMLChange` using primary key columns and calls `driver.ExecutePendingChanges` on `Ctrl+R`.

**Tech Stack:** Go, Bubble Tea, Lip Gloss, existing `models.DBDMLChange`, existing driver write API.

---

### Task 1: Local Edit Lifecycle

**Files:**
- Modify: `ui/results_test.go`
- Modify: `ui/results.go`

**Steps:**
1. Write failing tests for `Enter` starting edit, typed input replacing the current value, `Enter` committing locally, and `Esc` canceling.
2. Add edit state fields to `ResultsModel`.
3. While editing, route printable keys/backspace/Enter/Esc to edit input instead of navigation.
4. Render editing cell with cursor and dirty cells with accent.

### Task 2: Pending Change Tracking

**Files:**
- Modify: `ui/results_test.go`
- Modify: `ui/results.go`

**Steps:**
1. Write failing tests that committed edits increment pending count and appear in the statusline.
2. Store pending edits keyed by row/column with original/new values.
3. Add `pendingChangeCount()` helper.

### Task 3: Save Pending Changes

**Files:**
- Modify: `ui/home_model_test.go`
- Modify: `ui/home_model.go`

**Steps:**
1. Write a failing test that `Ctrl+R` calls `ExecutePendingChanges` with update changes.
2. Build `models.DBDMLChange` records from pending edits using PK columns from `GridColumn.IsPK` and current row values.
3. On save success, clear pending edits and statusline pending count.
4. On save failure, keep pending edits and show error status.

### Task 4: Verification

**Steps:**
1. Run `rtk go build -o lazysql .`.
2. Run `rtk go test ./...`.
