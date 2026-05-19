# Data Table UI Design

## Goal

Refresh the records table so it feels closer to a compact database client such as DataGrip or TablePlus while keeping the current Bubble Tea data flow and keyboard navigation.

## Scope

- Add a compact toolbar with active tab, row count, and selected cell position.
- Render the records grid inside a stronger bordered shell.
- Use a darker header with an accent on the active column.
- Add a dim row-number gutter.
- Use subtle alternating row treatment for readability.
- Keep the boxed active-cell highlight without solid-fill special values.
- Keep WHERE, pagination, and metadata tabs visual-only for this pass.

## Non-Goals

- No interactive WHERE editor in this pass.
- No clickable or switchable metadata tabs in this pass.
- No server-side column sort controls in this pass.
- No row edit/delete UI in this pass.

## Testing

- Add focused render tests for toolbar/grid markers where practical.
- Run `rtk go test ./...` and `rtk go build -o lazysql .`.
