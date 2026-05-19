# Filter Chip Redesign Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Redesign the Records WHERE filter row as a compact database-client-style filter chip.

**Architecture:** Keep the existing one-line `renderFilterBar` in `ui/results.go`, but restyle it as a segmented filter row: a `WHERE` badge, an input-like capsule, and a right-aligned apply hint. Preserve the current layout height and only verify rendered text/style-affordance through existing UI tests.

**Tech Stack:** Go, Bubble Tea, Lip Gloss, existing `ResultsModel` rendering.

---

### Task 1: Update Filter Bar Render Test

**Files:**
- Modify: `ui/results_test.go`

**Step 1:** Extend `TestResultsViewRendersDatabaseClientChrome` expectations to include the filter placeholder and apply hint.

**Step 2:** Run `rtk go test ./ui` and confirm the test fails before implementation.

### Task 2: Redesign Filter Bar

**Files:**
- Modify: `ui/results.go`

**Step 1:** Replace the current plain WHERE row with a segmented row.

**Step 2:** Use a green `WHERE` badge, dark input capsule, muted placeholder, and right-aligned `enter apply` hint.

**Step 3:** Keep the row one terminal line high.

### Task 3: Verify

**Files:**
- Test: `ui/results_test.go`

**Step 1:** Run `rtk go build -o lazysql .`.

**Step 2:** Run `rtk go test ./...`.

**Step 3:** Confirm all tests pass.
