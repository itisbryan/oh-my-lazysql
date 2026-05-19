# Data Table UI Refresh Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Refresh the records table into a compact database-client style grid.

**Architecture:** Keep the existing `ResultsModel` state and keyboard navigation. Refactor rendering into smaller helpers for toolbar, filter bar, grid shell, rows, and footer without changing data loading.

**Tech Stack:** Go, Bubble Tea, Lip Gloss, existing custom grid renderer.

---

### Task 1: Add Render Test

**Files:**
- Create: `ui/results_test.go`
- Modify: `ui/results.go`

**Step 1: Write the failing test**

Add a test that builds a `ResultsModel` with columns, rows, dimensions, selected row/column, and asserts the rendered view contains database-client UI markers: active tab, WHERE placeholder, row count, selected cell position, and boxed selected cell rails.

**Step 2: Run test to verify it fails**

Run: `rtk go test ./ui -run TestResultsViewRendersDatabaseClientChrome -count=1`

Expected: FAIL because the current view does not render the new toolbar/position text.

**Step 3: Implement minimal render helpers**

Update `ui/results.go` to render a compact toolbar, filter row, bordered grid, and footer.

**Step 4: Run focused test**

Run: `rtk go test ./ui -run TestResultsViewRendersDatabaseClientChrome -count=1`

Expected: PASS.

### Task 2: Verify Full Project

**Files:**
- Modify: `ui/results.go`
- Test: `ui/results_test.go`

**Step 1: Format code**

Run: `gofmt -w ui/results.go ui/results_test.go`

**Step 2: Build**

Run: `rtk go build -o lazysql .`

Expected: Success.

**Step 3: Test**

Run: `rtk go test ./...`

Expected: All packages pass.
