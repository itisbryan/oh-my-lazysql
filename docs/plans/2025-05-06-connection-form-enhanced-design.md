# Connection Form Enhanced Design

## Overview

Enhance the connection creation/edit form to support individual field inputs (host, port, user, password, database) alongside the existing connection string URL. Introduce connection profiles for switching between environments.

## Requirements

1. **Individual Fields** - Separate inputs for Host, Port, Username, Password, Database, SSL mode
2. **URL Field Hidden by Default** - Only shown via "Show Advanced" toggle
3. **SSL Mode Support** - Add SSL checkbox/options for applicable drivers
4. **SQLite Simplified Form** - Show only Database field, hide network-related fields
5. **Connection Profiles** - Multiple profiles (dev/staging/prod) per connection entry

---

## Design

### Layout

```
┌─ Connection Name ─────────────────────────────────────────┐
│ [myapp-db________________]                                │
├─ Provider ────────────────────────────────────────────────┤
│ [MySQL ▼]                                                 │
├─ Connection Details ──────────────────────────────────────┤
│ Host:     [localhost__________]                           │
│ Port:     [3306_____________]                            │
│ Username: [root_____________]                            │
│ Password: [********_________]                            │
│ Database: [mydb______________]                            │
│ SSL Mode: [ ] Enabled  Cert: [________] Key: [________]  │
├─ Profiles ────────────────────────────────────────────────┤
│ ┌─────────────────┐ ┌─────────────────┐ ┌───────────────┐ │
│ │ ● dev          │ │ ○ staging       │ │ ○ production  │ │
│ └─────────────────┘ └─────────────────┘ └───────────────┘ │
│ [yellow]+[-]                                      [Show Advanced]│
├───────────────────────────────────────────────────────────┤
│ [ ] Read-Only                                              │
├───────────────────────────────────────────────────────────┤
│ [F1] Save  [F2] Test  [F3] Connect  [Esc] Cancel           │
└───────────────────────────────────────────────────────────┘
```

### Driver-Specific Fields

| Driver    | Visible Fields                                      |
|-----------|------------------------------------------------------|
| MySQL     | Host, Port, Username, Password, Database, SSL       |
| PostgreSQL| Host, Port, Username, Password, Database, SSL       |
| SQLite    | Database (file path) only                           |
| MSSQL     | Host, Port, Username, Password, Database, SSL      |

### Connection Profiles

- Each connection can have multiple named profiles
- Profiles store environment-specific values for: Host, Port, Username, Password, Database, SSL
- The "Name" field becomes the connection group name
- Selecting a profile auto-fills the form fields
- Only one profile is "active" at a time (marked with ●)

**Profile Storage (in Connection model):**
```go
type Connection struct {
    Name      string
    URL       string
    Provider  string
    ReadOnly  bool
    Profiles  []Profile  // Multiple environments
}

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
```

When connecting, use the active profile to build the URL.

### URL Building Logic

1. If URL field is filled in Advanced section → use it directly
2. Otherwise → build from active profile fields:
   - MySQL: `mysql://${username}:${password}@${hostname}:${port}/${dbname}?sslmode=${sslmode}`
   - PostgreSQL: `postgres://${username}:${password}@${hostname}:${port}/${dbname}?sslmode=${sslmode}`
   - MSSQL: `sqlserver://${username}:${password}@${hostname}:${port}/${dbname}`
   - SQLite: `sqlite3:${dbname}`

### SSL Mode

For MySQL/PostgreSQL:
- Enable SSL checkbox
- When enabled, show additional fields: SSL Cert, SSL Key, SSL CA (all optional)

### Data Flow

**On Load (Edit mode):**
1. Parse existing URL to detect provider
2. Populate Provider dropdown
3. Populate Name field
4. If profiles exist → show them in profile tabs
5. If no profiles → create default from URL fields
6. ReadOnly checkbox populated

**On Save:**
1. Validate required fields (Name, Provider, Host/Port/Database depending on provider)
2. Build URL from profile fields (or use Advanced URL if provided)
3. Save Connection with Profiles array

---

## Components Affected

- `components/connection_form.go` - Main form logic
- `models/models.go` - Add Profile struct, modify Connection
- `helpers/utils.go` - Add URL builder function
- `app/config.go` - Handle profile serialization

---

## Backward Compatibility

- Existing connections without Profiles: auto-create default profile from URL
- URL field in Advanced section: still supported for complex connection strings