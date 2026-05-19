# Active Gradient Table Headers Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Give records table headers stronger colors with an active-column gradient feel.

**Architecture:** Keep the existing `renderGridHeader` helper and adjust only header cell/separator colors based on distance from the selected column. No data loading or navigation behavior changes.

**Tech Stack:** Go, Bubble Tea, Lip Gloss ANSI styling.

---

### Task 1: Header Color Regression Test

**Files:**
- Modify: `ui/results_test.go`
- Modify: `ui/results.go`

**Step 1: Write the failing test**

Assert rendered output includes the active header marker and ANSI color escape sequences for the active gradient header.

**Step 2: Run test to verify it fails**

Run: `rtk go test ./ui -run TestResultsViewRendersDatabaseClientChrome -count=1`

Expected: FAIL because current color assertions are not present.

**Step 3: Implement active gradient header colors**

Update `renderGridHeader` to use brighter background/foreground for active header, subtle tint for neighboring headers, and brighter separators near the active column.

**Step 4: Run focused test**

Run: `rtk go test ./ui -run TestResultsViewRendersDatabaseClientChrome -count=1`

Expected: PASS.

### Task 2: Verify Full Project

**Files:**
- Modify: `ui/results.go`
- Modify: `ui/results_test.go`

**Step 1: Format code**

Run: `gofmt -w ui/results.go ui/results_test.go`

**Step 2: Build and test**

Run: `rtk go build -o lazysql . && rtk go test ./...`

Expected: Build succeeds and all packages pass.
