# Table Header Refresh Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make records table headers feel more like a polished database-client grid.

**Architecture:** Keep the existing custom grid renderer and replace only the header-row construction. Extract a small helper so header formatting remains isolated from row rendering.

**Tech Stack:** Go, Bubble Tea, Lip Gloss.

---

### Task 1: Header Render Test

**Files:**
- Modify: `ui/results_test.go`
- Modify: `ui/results.go`

**Step 1: Write the failing test**

Assert the rendered table includes the active header marker `▸ email` and structured header separators.

**Step 2: Run test to verify it fails**

Run: `rtk go test ./ui -run TestResultsViewRendersDatabaseClientChrome -count=1`

Expected: FAIL because headers currently render only the bare column name.

**Step 3: Implement header helper**

Add a helper in `ui/results.go` that pads header labels, uses `▸` for active column, and uses stronger separator glyphs in the header band.

**Step 4: Run focused test**

Run: `rtk go test ./ui -run TestResultsViewRendersDatabaseClientChrome -count=1`

Expected: PASS.

### Task 2: Full Verification

**Files:**
- Modify: `ui/results.go`
- Modify: `ui/results_test.go`

**Step 1: Format code**

Run: `gofmt -w ui/results.go ui/results_test.go`

**Step 2: Build and test**

Run: `rtk go build -o lazysql . && rtk go test ./...`

Expected: Build succeeds and all packages pass.
