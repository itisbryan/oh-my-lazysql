# Retro Terminal Welcome Screen Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the plain connection launcher with a retro terminal-style welcome screen.

**Architecture:** Keep `ConnectionListModel` behavior unchanged and refactor only `View()` rendering into small helpers for title, boot menu, system panel, and command bar.

**Tech Stack:** Go, Bubble Tea, Lip Gloss.

---

### Task 1: Welcome Screen Render Test

**Files:**
- Create: `ui/connection_list_test.go`
- Modify: `ui/connection_list.go`

**Step 1: Write the failing test**

Assert the connection list view includes retro terminal markers: ASCII title, `DATABASE CONSOLE`, `BOOT MENU`, `SYSTEM`, and command-bar text.

**Step 2: Run test to verify it fails**

Run: `rtk go test ./ui -run TestConnectionListRendersRetroWelcomeScreen -count=1`

Expected: FAIL because the current view renders only `LazySQL` and a plain connection table.

**Step 3: Implement retro rendering helpers**

Update `ui/connection_list.go` with helper functions for title, connection boot rows, system panel, empty state, and command bar.

**Step 4: Run focused test**

Run: `rtk go test ./ui -run TestConnectionListRendersRetroWelcomeScreen -count=1`

Expected: PASS.

### Task 2: Full Verification

**Files:**
- Modify: `ui/connection_list.go`
- Test: `ui/connection_list_test.go`

**Step 1: Format code**

Run: `gofmt -w ui/connection_list.go ui/connection_list_test.go`

**Step 2: Build and test**

Run: `rtk go build -o lazysql . && rtk go test ./...`

Expected: Build succeeds and all packages pass.
