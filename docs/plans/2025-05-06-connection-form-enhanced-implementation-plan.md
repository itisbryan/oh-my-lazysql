# Connection Form Enhanced Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add individual field inputs (host, port, user, password, database, SSL) to connection form, with connection profiles for environment switching.

**Architecture:** Connection model gains a `Profiles` array. Form builds URLs from profile fields or accepts raw URL in advanced mode. Profiles are per-connection, not global.

**Tech Stack:** Go, tview, dburl

---

## Task 1: Add Profile Model

**Files:**
- Modify: `models/models.go:19-50`

**Step 1: Add Profile struct and modify Connection**

```go
// Add after Command struct (around line 50)

type Profile struct {
    Name       string
    Hostname   string
    Port       string
    Username   string
    Password   string
    DBName     string
    SSLEnabled bool
    SSLCert    string
    SSLKey     string
    SSLCA      string
}

// Modify Connection struct (line 19-41)
type Connection struct {
    Name     string
    URL      string // Advanced/manual connection string
    Provider string
    ReadOnly bool

    // New: Profiles for environment switching
    Profiles []Profile

    // Legacy fields (kept for backward compat when loading old configs)
    Username  string
    Password  string
    Hostname  string
    Port      string
    DBName    string
    URLParams string

    Schemas   []string
    Commands  []*Command
}
```

**Step 2: Run to verify**

Run: `go build ./...`
Expected: SUCCESS

**Step 3: Commit**

```bash
git add models/models.go
git commit -m "feat: add Profile struct to Connection model"
```

---

## Task 2: Add URL Builder Helper

**Files:**
- Modify: `helpers/utils.go`

**Step 1: Add BuildConnectionURL function**

```go
// Add after ParseConnectionString (line 17)

func BuildConnectionURL(provider, username, password, hostname, port, dbname string, sslEnabled bool, sslCert, sslKey, sslCA string) (string, error) {
    switch provider {
    case DriverMySQL:
        url := fmt.Sprintf("mysql://%s:%s@%s:%s/%s", url.QueryEscape(username), url.QueryEscape(password), hostname, port, dbname)
        if sslEnabled {
            url += "?tls=true"
            if sslCert != "" || sslKey != "" || sslCA != "" {
                url += "&tls=true"
            }
        }
        return url, nil

    case DriverPostgres:
        url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", url.QueryEscape(username), url.QueryEscape(password), hostname, port, dbname)
        if sslEnabled {
            url += "?sslmode=require"
        }
        return url, nil

    case DriverSqlite:
        if dbname == "" {
            return "", errors.New("database name required for SQLite")
        }
        return fmt.Sprintf("sqlite3:%s", dbname), nil

    case DriverMSSQL:
        url := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s", url.QueryEscape(username), url.QueryEscape(password), hostname, port, dbname)
        return url, nil

    default:
        return "", errors.New("unsupported provider: " + provider)
    }
}
```

**Step 2: Import "fmt" if not present**

Check imports and add `fmt` if missing.

**Step 3: Run to verify**

Run: `go build ./...`
Expected: SUCCESS

**Step 4: Commit**

```bash
git add helpers/utils.go
git commit -m "feat: add BuildConnectionURL helper function"
```

---

## Task 3: Rewrite Connection Form UI

**Files:**
- Modify: `components/connection_form.go`

**Step 1: Rewrite connection form to support profiles**

This is a significant UI change. The form needs to:
1. Add provider dropdown
2. Add individual field inputs
3. Add profile selector buttons
4. Add SSL checkbox and cert fields
5. Add "Show Advanced" toggle for URL field

Key form indices after rewrite:
- 0: Name
- 1: Provider (dropdown: MySQL, PostgreSQL, SQLite, MSSQL)
- 2: Hostname
- 3: Port
- 4: Username
- 5: Password
- 6: Database
- 7: SSL Enabled (checkbox)
- 8: SSL Cert (shown only when SSL enabled)
- 9: SSL Key (shown only when SSL enabled)
- 10: SSL CA (shown only when SSL enabled)
- 11: Read-Only (checkbox)
- 12: URL (Advanced, hidden by default)

Profile buttons rendered separately, not as form items.

**Step 2: Update inputCapture logic**

When F1/Save pressed:
1. Get Name, Provider from form
2. If advanced URL filled, use it directly
3. Otherwise build URL from form fields + active profile
4. If no profiles exist, create default profile from fields
5. Save connection with Profiles array

**Step 3: Update SetConnectionData**

When editing, populate form from connection's active profile (or first profile, or legacy URL fields).

**Step 4: Run to verify**

Run: `go build ./...`
Expected: SUCCESS

**Step 5: Commit**

```bash
git add components/connection_form.go
git commit -m "feat: rewrite connection form with individual fields and profiles"
```

---

## Task 4: Add Profile Management to Form

**Files:**
- Modify: `components/connection_form.go`

**Step 1: Add profile selection buttons**

In NewConnectionForm, after the form items, add profile display area:
- Show buttons for each profile (● for active, ○ for inactive)
- [yellow]+ button to add new profile (opens input modal)
- [-] button to delete selected profile

**Step 2: Add profile selection handler**

When profile button clicked:
- Set all form fields from selected profile
- Update button states (●/○)

**Step 3: Add profile creation modal**

When [+] clicked, show modal with:
- Text input for profile name
- Confirm/Cancel buttons
- After confirm, create profile with current form values

**Step 4: Run to verify**

Run: `go build ./...`
Expected: SUCCESS

**Step 5: Commit**

```bash
git add components/connection_form.go
git commit -m "feat: add profile management to connection form"
```

---

## Task 5: Update Connection Selection to Show Profiles

**Files:**
- Modify: `components/connection_selection.go`

**Step 1: Update connection table display**

When displaying connections, if a connection has multiple profiles:
- Show connection name with profile count: "WMS DB (3 profiles)"
- Or expand to show all profiles as separate rows

**Step 2: Handle profile selection**

When user selects a connection with profiles:
- Option A: Connect immediately with active profile
- Option B: Show profile picker submenu

Recommend Option B for better UX.

**Step 3: Run to verify**

Run: `go build ./...`
Expected: SUCCESS

**Step 4: Commit**

```bash
git add components/connection_selection.go
git commit -m "feat: show profiles in connection selection"
```

---

## Task 6: Update Config Load/Save for Profiles

**Files:**
- Modify: `app/config.go`

**Step 1: Handle backward compatibility**

In LoadConfig, existing connections without Profiles:
- Create default profile from legacy fields (Username, Password, Hostname, Port, DBName)
- Set profile name to "default"

**Step 2: Verify save works**

Verify SaveConnections properly serializes Profiles array to TOML.

**Step 3: Run to verify**

Run: `go build ./...`
Expected: SUCCESS

**Step 4: Commit**

```bash
git add app/config.go
git commit -m "feat: handle profile serialization in config"
```

---

## Task 7: Test End-to-End

**Manual Testing:**
1. Create new connection with individual fields
2. Add multiple profiles (dev, stag, prod)
3. Switch between profiles
4. Save and restart app
5. Verify profiles persist
6. Test connecting with different profiles
7. Test editing existing connection (backward compat)
8. Test SQLite (should only show database field)
9. Test SSL mode

**Step 2: Commit**

```bash
git add -A
git commit -m "feat: complete connection profiles feature"
```

---

## Summary

| Task | Description |
|------|-------------|
| 1 | Add Profile struct to models |
| 2 | Add BuildConnectionURL helper |
| 3 | Rewrite connection form UI |
| 4 | Add profile management (add/delete/select) |
| 5 | Update connection selection UI |
| 6 | Handle profile serialization in config |
| 7 | End-to-end testing |

---

**Plan complete and saved to `docs/plans/2025-05-06-connection-form-enhanced-design.md`. Two execution options:**

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

Which approach?