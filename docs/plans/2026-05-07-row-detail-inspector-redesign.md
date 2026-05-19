# Row Detail Inspector Redesign Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Redesign the `o` row detail panel into a polished database-inspector view.

**Architecture:** Keep the existing row-detail state and edit behavior, but replace the panel composition in `renderRowDetail`. The new renderer uses a compact header bar, column-like labels (`FIELD`, `TYPE`, `VALUE`), selected-field accent rail, dirty markers, and statusbar-style footer hints.

**Tech Stack:** Go, Bubble Tea, Lip Gloss.

---

### Task 1: Render Tests

**Files:**
- Modify: `ui/results_test.go`

**Steps:**
1. Add failing expectations for new inspector labels: `FIELD`, `TYPE`, `VALUE`.
2. Add expectations for `fields` count and footer hint segments.
3. Run `rtk go test ./ui`; expect failure.

### Task 2: Renderer Redesign

**Files:**
- Modify: `ui/results.go`

**Steps:**
1. Replace `renderRowDetail` title/rule/body/footer layout.
2. Add compact header bar with `Row N`, field count, and pending count.
3. Add a row header line: `FIELD`, `TYPE`, `VALUE`.
4. Render selected row with `▌` rail and full-row background.
5. Render dirty rows with `*` marker and orange accent.
6. Render footer as segmented status hints.

### Task 3: Verification

**Steps:**
1. Run `rtk go build -o lazysql .`.
2. Run `rtk go test ./...`.
