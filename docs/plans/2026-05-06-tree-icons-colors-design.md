# Tree Icons And Colors Design

## Goal

Make the database schema tree easier to scan by adding Nerd Font icons and semantic colors for database object types.

## Design

Use Nerd Font glyphs as prefixes for tree node labels:

- Database: `󰆼`
- Schema/system schema: `󰙅`
- Tables section: `󰓫`
- Table: `󰓱`
- Views section and view nodes: `󰈙`
- Functions section and function nodes: `󰊕`
- Procedures section and procedure nodes: `󰡱`

Apply semantic colors by node category:

- Database: blue/cyan
- Schema: muted purple/gray
- System schema: dim gray
- Section nodes: yellow
- Table nodes: primary/default text color
- View nodes: green
- Function nodes: magenta
- Procedure nodes: orange

## Constraints

Search and selection references must continue to use raw database object names, not icon-prefixed labels. Existing `SetReference` values remain the stable identifiers.

The selected node highlight must continue to work with the existing `tview` text color tag logic.

## Implementation Notes

Keep the change localized to `components/tree.go` by introducing small helpers for tree labels and colors, then using those helpers where nodes are created.

## Testing

Run `go test ./...` to verify the change does not break existing tree behavior or other components.
