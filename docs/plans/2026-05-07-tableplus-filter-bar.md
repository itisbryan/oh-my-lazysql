# TablePlus Filter Bar Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Restyle the WHERE filter bar to feel closer to TablePlus/DataGrip.

**Architecture:** Keep the existing one-line `renderFilterBar` and only adjust its visual composition. Render a muted `WHERE` pill, a wide input-like field, and a right-aligned `Enter Apply` action hint.

**Tech Stack:** Go, Bubble Tea, Lip Gloss.

---

### Task 1: Update Render Expectations

**Files:**
- Modify: `ui/results_test.go`

**Step 1:** Update expected filter placeholder to `filter rows by SQL predicate` and action hint to `Enter Apply`.

**Step 2:** Run `rtk go test ./ui`; expect failure before implementation.

### Task 2: Restyle Filter Bar

**Files:**
- Modify: `ui/results.go`

**Step 1:** Change `renderFilterBar` to compose a muted WHERE pill, expanded field, and right action hint.

**Step 2:** Keep one-line height and preserve existing layout.

### Task 3: Verify

**Files:**
- Test: `ui/results_test.go`

**Step 1:** Run `rtk go build -o lazysql .`.

**Step 2:** Run `rtk go test ./...`.
