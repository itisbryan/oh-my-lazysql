# Interactive WHERE Filter Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add an interactive TablePlus/DataGrip-style WHERE filter to the records grid.

**Architecture:** `ResultsModel` owns filter UI state (`filterEditing`, `filterInput`, `whereFilter`) and emits a typed message when the user applies a new predicate. `HomeModel` receives that message, resets pagination, and reloads the current table with the predicate passed to `driver.GetRecords`.

**Tech Stack:** Go, Bubble Tea, Lip Gloss, existing driver API.

---

### Task 1: Tests

**Files:**
- Modify: `ui/results_test.go`

**Steps:**
1. Add a test that `/` enters filter editing and renders `Enter Apply` / `Esc Cancel`.
2. Add a test that typing text then pressing `Enter` emits a filter-applied message and persists the active filter.
3. Run `rtk go test ./ui` and confirm failure before implementation.

### Task 2: ResultsModel Filter State

**Files:**
- Modify: `ui/results.go`

**Steps:**
1. Add `filterEditing`, `filterInput`, and `whereFilter` fields.
2. Add `whereFilterAppliedMsg`.
3. Route `/`, text input, backspace, Enter, and Esc while editing.
4. Render inactive, editing, and active filter bar states.

### Task 3: HomeModel Reload Wiring

**Files:**
- Modify: `ui/home_model.go`

**Steps:**
1. Handle `whereFilterAppliedMsg`.
2. Reset page/row to 0 and call `loadCurrentTableRecords`.
3. Pass `m.results.whereFilter` into `driver.GetRecords`.

### Task 4: Verification

**Steps:**
1. Run `rtk go build -o lazysql .`.
2. Run `rtk go test ./...`.
