<a name="readme-top"></a>

# OhMyLazySQL — relationship-aware terminal database browser

A keyboard-first TUI database client written in Go. OhMyLazySQL builds on the original LazySQL and focuses on a friendlier connection flow, smoother table browsing, inline editing, and relationship-aware navigation.

> This repository is a fork of [jorgerojas26/lazysql](https://github.com/itisbryan/oh-my-lazysql). The original project deserves credit for the core idea, database driver foundation, TUI structure, configuration format, and much of the baseline functionality. This fork documents the behavior in this repository, which now differs from upstream in several UI and navigation areas.

![Connection selection screenshot][product-screenshot1]
![OhMyLazySQL screenshot][product-screenshot2]

## Contents

- [What this fork adds](#what-this-fork-adds)
- [Supported databases](#supported-databases)
- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Connections](#connections)
- [Keyboard workflow](#keyboard-workflow)
- [Keybindings](#keybindings)
- [External editor](#external-editor)
- [Clipboard support](#clipboard-support)
- [Development](#development)
- [Attribution](#attribution)
- [License](#license)

## What this fork adds

OhMyLazySQL keeps the original LazySQL terminal workflow and adds a more polished day-to-day database browsing experience:

- Friendly guided connection form with provider defaults and URL-paste mode
- Saved connection list with wider spacing and clearer selected state
- Database engine icons for MySQL, PostgreSQL, SQLite, and MSSQL
- A-Z sorted schemas and tables in the sidebar
- Faster sidebar movement with `Ctrl+D` and `Ctrl+U`
- Search/scan prefix highlighting in the tree
- Global return to connection list with `Ctrl+P`
- Loading spinner while records are being fetched
- Relationship-aware browsing: press `Enter` on a foreign-key cell to follow it
- Foreign-key navigation history: press `[` to travel back
- Smarter inline editing for booleans and enums
- Nullable boolean/enum cells cycle through `NULL`
- `Ctrl+A` select-all behavior while editing cells
- PostgreSQL enum metadata extraction
- Read-only connection mode
- Local project config via `.lazysql.toml`
- Custom keymaps through TOML config
- CSV export
- JSON viewer for rows and cells
- Query history and query preview workflows

> Some visual details use Nerd Font / Devicon glyphs. Use a Nerd Font-compatible terminal font if database icons appear as boxes.

## Supported databases

| Database | Status |
| --- | --- |
| MySQL | Supported |
| PostgreSQL | Supported |
| SQLite | Supported |
| MSSQL | Supported |
| MongoDB | Not supported |

## Installation

### Homebrew (macOS / Linux)

```bash
brew install itisbryan/tap/oh-my-lazysql
```

Or:

```bash
brew tap itisbryan/tap
brew install oh-my-lazysql
```

Once installed, just run:

```bash
oh-my-lazysql
```

### Curl (macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/itisbryan/oh-my-lazysql/main/install.sh | bash
```

Or download a release tarball directly:

```bash
# Apple Silicon
curl -L https://github.com/itisbryan/oh-my-lazysql/releases/latest/download/oh-my-lazysql_Darwin_arm64.tar.gz -o oh-my-lazysql.tar.gz

# Intel
curl -L https://github.com/itisbryan/oh-my-lazysql/releases/latest/download/oh-my-lazysql_Darwin_x86_64.tar.gz -o oh-my-lazysql.tar.gz

tar -xzf oh-my-lazysql.tar.gz
sudo mv oh-my-lazysql /usr/local/bin/oh-my-lazysql
oh-my-lazysql --version
```

> **Note:** Publishing a GitHub Release (e.g. `v0.1.0`) with tarball assets is required for Homebrew and curl install to work. Update the placeholders with your GitHub username once releases are published.


### Build from source

```bash
git clone https://github.com/itisbryan/oh-my-lazysql
cd oh-my-lazysql
go build -o oh-my-lazysql .
./oh-my-lazysql
```

Or install into your Go bin path:

```bash
go install .
```

## Usage

Open the connection picker:

```bash
oh-my-lazysql
```

Connect directly with a URL:

```bash
oh-my-lazysql 'postgres://user:pass@localhost:5432/app'
```

Open in read-only mode:

```bash
oh-my-lazysql --read-only 'postgres://user:pass@localhost:5432/app'
```

Use a specific config file:

```bash
oh-my-lazysql --config ./config.toml
```

Other flags:

```bash
oh-my-lazysql --version
oh-my-lazysql --loglevel debug
oh-my-lazysql --logfile ./oh-my-lazysql.log
```

## Configuration

OhMyLazySQL currently reuses the upstream LazySQL config locations:

- `$XDG_CONFIG_HOME/lazysql/config.toml` when `XDG_CONFIG_HOME` is set
- macOS: `~/Library/Application Support/lazysql/config.toml`
- Linux: `~/.config/lazysql/config.toml`
- Windows: `%APPDATA%\lazysql\config.toml`

### Application settings

```toml
[application]
theme = "tokyonight"
default_page_size = 300
disable_sidebar = false
sidebar_overlay = false
max_query_history_per_connection = 100
tree_width = 30
json_viewer_word_wrap = false
enter_opens_json_viewer = false
```

| Setting | Default | Description |
| --- | --- | --- |
| `theme` | `tokyonight` | Color palette: `tokyonight`, `dracula`, `catppuccin-mocha`, `nord`, `gruvbox-dark`, or `terminal` |
| `default_page_size` | `300` | Number of rows fetched per page |
| `disable_sidebar` | `false` | Start without the table tree sidebar |
| `sidebar_overlay` | `false` | Show sidebar as an overlay instead of a side panel |
| `max_query_history_per_connection` | `100` | Query history entries retained per connection |
| `tree_width` | `30` | Preferred tree/sidebar width |
| `json_viewer_word_wrap` | `false` | Wrap long JSON lines |
| `enter_opens_json_viewer` | `false` | Open JSON viewer with `Enter` on normal cells |

### Local project config

Place `.lazysql.toml` in a project directory to override the global config for that repo. OhMyLazySQL walks upward from the current directory until it finds `.lazysql.toml` or reaches the Git root.

Merge behavior:

| Section | Behavior |
| --- | --- |
| `[application]` | Deep merge; local values override global/default values |
| `[[database]]` | Replace; local connections replace global connections |
| `[keymap.*]` | Deep merge; local keybindings override matching commands |

Example:

```toml
[application]
theme = "tokyonight"
default_page_size = 500

[[database]]
Name = "Local development"
Provider = "postgres"
URL = "postgres://localhost/myproject_dev"
```

Saving connections from the UI writes to the global config file, not `.lazysql.toml`.

### Themes

Set the UI palette in `[application]`:

```toml
[application]
theme = "catppuccin-mocha"
```

Available themes: `tokyonight`, `dracula`, `catppuccin-mocha`, `nord`, `gruvbox-dark`, and `terminal`.

## Connections

You can create connections in the TUI or define them in TOML.

### Guided fields or URL mode

The connection form supports two modes:

- **Guided fields** — pick a provider and fill host, port, user, password, and database
- **Connection URL** — paste a full connection string

Provider defaults are filled automatically:

| Provider | Default port |
| --- | --- |
| MySQL | `3306` |
| PostgreSQL | `5432` |
| MSSQL | `1433` |
| SQLite | empty/path based |

### Example connection config

```toml
[[database]]
Name = "Production"
Provider = "postgres"
DBName = "app"
URL = "postgres://${env:DB_USER}:${env:DB_PASSWORD}@localhost:5432/app"
ReadOnly = true

[[database]]
Name = "Local MySQL"
Provider = "mysql"
URL = "mysql://root:password@localhost:3306/app"
```

### Environment variables

Use `${env:VAR_NAME}` inside config values:

```toml
[[database]]
Name = "Production"
Provider = "postgres"
URL = "postgres://${env:DB_USER}:${env:DB_PASSWORD}@localhost:5432/app"
```

Undefined environment variables resolve to an empty string.

### Connection helper commands

Connections can run commands before connecting. This is useful for SSH tunnels, port-forwarding, or dynamic secrets.

```toml
[[database]]
Name = "Bastion Postgres"
Provider = "postgres"
DBName = "app"
URL = "postgres://${user}:password@localhost:${port}/app"
Commands = [
  { Command = "ssh -tt remote-bastion -L ${port}:localhost:5432", WaitForPort = "${port}" },
  { Command = "whoami", SaveOutputTo = "user" }
]
```

Fields:

| Field | Description |
| --- | --- |
| `Command` | Shell command to run |
| `WaitForPort` | Wait until this port is open before connecting |
| `SaveOutputTo` | Save stdout to a variable usable as `${name}` |
| `Timeout` | Startup timeout in seconds |

`$SQL_EDITOR`, `$EDITOR`, and `$VISUAL` are used for editor integration; see [External editor](#external-editor).

### Example connection URLs

```text
postgres://user:pass@localhost/dbname
pg://user:pass@localhost/dbname?sslmode=disable
mysql://user:pass@localhost/dbname
mysql:/var/run/mysqld/mysqld.sock
sqlserver://user:pass@remote-host.com/dbname
mssql://user:pass@remote-host.com/instance/dbname
ms://user:pass@remote-host.com:port/instance/dbname?keepAlive=10
file:myfile.sqlite3?loc=auto
/path/to/sqlite/file/test.db
odbc+postgres://user:pass@localhost:port/dbname?option1=
```

## Keyboard workflow

Press `?` inside the app for the built-in help modal.

### Connect

1. Run `oh-my-lazysql`
2. Press `n` to add a connection, or select an existing one
3. Press `Enter` or `c` to connect
4. Use `Ctrl+P` from the main screen to return to the connection list

### Browse tables

1. Focus the tree with `H`
2. Expand items with `e` or `Enter`
3. Move with `j`/`k`, `Ctrl+D`/`Ctrl+U`, `g`, and `G`
4. Press `Enter` on a table to load records
5. Focus results with `L`

The tree sorts schemas and tables A-Z. When using prefix/scan movement, the selected row shows a highlighted prefix badge.

### Filter, sort, and paginate

- `/` focuses the table filter
- Enter applies the filter
- `Esc` clears/unfocuses the filter
- `K` sorts ascending by the selected column
- `J` sorts descending by the selected column
- `>` loads the next page
- `<` loads the previous page
- A spinner is shown while records are loading

### Follow related records

When a selected cell belongs to a foreign-key column, press `Enter` to jump to the referenced table and filter to the matching record.

Example:

```text
orders.user_id = 42  --Enter-->  users where id = 42
```

Press `[` to go back through relationship navigation history. If there is no relationship history, `[` behaves as the normal previous-tab key.

### Edit cells

- `c` starts text editing for the selected cell
- `Enter` commits the edit
- `Esc` cancels editing
- `Ctrl+A` selects the current edit text so the next typed character replaces it
- `Backspace` clears selected edit text
- `Ctrl+S` saves pending changes

Boolean columns (`bool`, `boolean`, `tinyint(1)`, `bit`) can be toggled directly with `Enter`:

```text
true -> false -> NULL -> true
```

For non-nullable booleans, `NULL` is skipped.

Enum columns cycle through known enum values with `Enter`. Nullable enum columns include `NULL` in the cycle.

### Insert, duplicate, delete

- `o` appends a new row
- `O` duplicates the selected row
- `d` marks the selected row for deletion
- `Ctrl+S` saves pending inserts, updates, and deletes

### SQL editor

- `Ctrl+E` toggles the SQL editor
- `Ctrl+R` executes the current SQL query
- After a `SELECT`, results appear under the editor
- `/` returns focus to the SQL editor from query results
- `Ctrl+O` opens the SQL editor content in an external editor

### Metadata tabs

Use number keys in the table view:

| Key | Tab |
| --- | --- |
| `1` | Records |
| `2` | Columns |
| `3` | Constraints |
| `4` | Foreign keys |
| `5` | Indexes |

### JSON viewer

- `z` opens the selected cell in the JSON viewer
- `Z` opens the selected row in the JSON viewer
- `w` toggles word wrap
- `y` copies JSON to the clipboard
- `z`/`Z` closes the viewer

If `enter_opens_json_viewer = true`, `Enter` opens JSON for normal cells. Foreign-key cells still use `Enter` for relationship navigation.

### Export CSV

From table results:

1. Open a table
2. Apply filters or sorting if needed
3. Press `E`
4. Choose current page or all records

From SQL query results:

1. Execute a query
2. Press `E`
3. Choose the output path

The default path is:

```text
~/Downloads/{database}_{table}_{timestamp}.csv
```

## Keybindings

Keybindings can be customized with `[keymap.<Group>]` sections in `config.toml` or `.lazysql.toml`.

```toml
[keymap.Home]
SwitchToEditorView = "i"
Quit = "Esc"

[keymap.Tree]
GotoTop = "t"
Search = "Ctrl-F"
```

For single-character keys, use the character directly (`"q"`, `"G"`, `"/"`). For special keys, use tcell key names (`"Enter"`, `"Esc"`, `"Ctrl-S"`). Group names are case-insensitive.

Available groups:

```text
Home, Connection, Tree, TreeFilter, Table, Editor, Sidebar,
QueryPreview, QueryHistory, JSONViewer
```

### Home

| Key | Action |
| --- | --- |
| `L` / `Ctrl+L` | Focus table |
| `H` / `Ctrl+H` | Focus tree |
| `Ctrl+E` | Toggle SQL editor |
| `Ctrl+S` | Save pending table changes |
| `Ctrl+P` | Return to connection list |
| `Ctrl+_` | Toggle query history |
| `T` | Toggle tree/sidebar |
| `[` | Back through relationship navigation history |
| `?` | Help |
| `q` | Quit |

### Connection list

| Key | Action |
| --- | --- |
| `n` | New connection |
| `e` | Edit connection |
| `d` | Delete connection |
| `c` / `Enter` | Connect |
| `q` | Quit |

### Tree

| Key | Action |
| --- | --- |
| `j` / `Down` | Move down |
| `k` / `Up` | Move up |
| `Ctrl+D` | Jump/page down |
| `Ctrl+U` | Jump/page up |
| `g` | Top |
| `G` | Bottom |
| `e` | Expand all/open |
| `Enter` | Open selected node/table |
| `/` | Search |
| `n` / `P` | Next found node |
| `N` / `p` | Previous found node |
| `c` | Collapse all |
| `R` | Refresh tree |

### Table/results

| Key | Action |
| --- | --- |
| `/` | Filter/search |
| `c` | Edit selected cell |
| `Enter` | Follow FK cell, toggle bool/enum cell, or commit edit depending on context |
| `d` | Delete selected row |
| `o` | Append row |
| `O` | Duplicate row |
| `Ctrl+S` | Save pending changes |
| `y` | Copy selected cell |
| `w` / `b` | Move next/previous cell |
| `0` / `$` | First/last cell |
| `K` / `J` | Sort ascending/descending |
| `R` | Refresh current table |
| `C` | Set value menu (`NULL`, empty, `DEFAULT`) |
| `[` | Go back through FK history, otherwise previous tab |
| `]` | Next tab |
| `{` / `}` | First/last tab |
| `>` / `<` | Next/previous page |
| `1`-`5` | Records, columns, constraints, foreign keys, indexes |
| `S` | Toggle sidebar |
| `s` | Focus sidebar |
| `z` / `Z` | JSON viewer for cell/row |
| `E` | Export CSV |
| `e` | Open selected cell in external editor |

### SQL editor

| Key | Action |
| --- | --- |
| `Ctrl+R` | Execute query |
| `Esc` | Unfocus editor |
| `Ctrl+O` | Open in external editor |

### Sidebar/detail editor

| Key | Action |
| --- | --- |
| `j` / `k` | Move field down/up |
| `g` / `G` | First/last field |
| `c` | Edit field |
| `Enter` | Commit field edit |
| `Esc` | Discard field edit |
| `C` | Set value menu |
| `y` | Copy value |
| `s` | Return focus to table |
| `S` | Toggle sidebar |

### Query preview/history

| Key | Action |
| --- | --- |
| `Ctrl+S` | Execute previewed queries |
| `s` | Save query in history |
| `d` | Delete query/history entry |
| `y` | Copy query |
| `/` | Search query history |
| `Ctrl+_` | Toggle query history |
| `[` / `]` | Previous/next history tab |
| `q` | Close |

## External editor

External editor support is available on Linux and macOS.

Resolution order:

- SQL editor: `$SQL_EDITOR` → `$EDITOR` → `$VISUAL` → `vi`
- Table cells: `$EDITOR` → `$VISUAL` → `vi`

## Clipboard support

OhMyLazySQL uses [atotto/clipboard](https://github.com/atotto/clipboard) for clipboard operations.

Platform notes:

- macOS: supported
- Windows: supported
- Linux/Unix: requires `xclip` or `xsel`

## Development

Requirements:

- Go 1.25+
- A terminal with good color support
- Nerd Font-compatible font recommended for icons

Common commands:

```bash
go test ./...
go build -o oh-my-lazysql .
./oh-my-lazysql
```

## Attribution

This fork is based on [jorgerojas26/lazysql](https://github.com/itisbryan/oh-my-lazysql), created by Jorge Rojas. Upstream LazySQL established the original TUI database client, driver interfaces, configuration model, and much of the application foundation.

This README describes the current behavior of this fork. If you are looking for upstream releases, upstream package-manager instructions, or upstream issue tracking, visit [jorgerojas26/lazysql](https://github.com/itisbryan/oh-my-lazysql).

Related projects and inspiration:

- [Lazygit](https://github.com/jesseduffield/lazygit)
- [Gobang](https://github.com/TaKO8Ki/gobang)
- [Mitzasql](https://github.com/vladbalmos/mitzasql)

## License

Distributed under the MIT License. See `LICENSE.txt` for details.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

[product-screenshot1]: images/lazysql-connection-selection.png
[product-screenshot2]: images/lazysql.png
