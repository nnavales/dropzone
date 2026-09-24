# Configuration reference

Config file: `~/.config/dropzone/config.yml` (`config.yaml` is accepted if present instead).
Create it with `dropzone init`. The daemon reloads it automatically when the file
changes; an invalid update is rejected and the previous config keeps running.

## `settings`

| Key                  | Default  | Description                                                                                                              |
| -------------------- | -------- | ------------------------------------------------------------------------------------------------------------------------ |
| `stable_for_seconds` | `2`      | Act on a file after its size and modification time stay unchanged for this long. Must be `>= 0` (`0` falls back to `2`). |
| `on_conflict`        | `rename` | What to do when an action's destination already exists: `skip` \| `overwrite` \| `rename`.                               |

## `zones`

```yaml
zones:
  - name: downloads
    path: ~/Downloads
    rules: [...]
```

| Key     | Description                                                                                                                                                                                    |
| ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `name`  | Label used in logs (optional).                                                                                                                                                                 |
| `path`  | Watched directory. Must be absolute. `~` and environment variables are expanded, the result is cleaned. Two zones cannot share the same path (compared after normalization, rejected at load). |
| `rules` | List of rules, evaluated top to bottom.                                                                                                                                                        |

Zones match a file by path prefix, top to bottom — the first match wins. For nested
zones, list the inner zone first.

## `rules`

```yaml
rules:
  - name: images
    on_conflict: overwrite # optional, overrides settings.on_conflict
    match:
      extensions: [jpg, png]
      glob: ["shot-*"]
    action:
      move: ~/Pictures
```

(`name` is optional on zones and rules; shown here for clarity.)

### `match`

| Key          | Description                                                                                                                                            |
| ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `extensions` | File extensions, case-insensitive, leading dot optional (`jpg` and `.jpg` are the same). Any entry may match (OR).                                     |
| `glob`       | [`filepath.Match`](https://pkg.go.dev/path/filepath#Match) patterns evaluated against the file name only, not the full path. Any entry may match (OR). |

`extensions` and `glob` combine with AND: when both are set, a file must satisfy
each kind. At least one of the two is required per rule.

### `action`

Exactly one action per rule:

| Action               | Target                | Description                                                                                           |
| -------------------- | --------------------- | ----------------------------------------------------------------------------------------------------- |
| `move: <dir>`        | destination directory | Move the file, keeping its name. Honors `on_conflict`.                                                |
| `copy: <dir>`        | destination directory | Copy the file, keeping its name. Honors `on_conflict`.                                                |
| `rename: <template>` | new file name         | Rename within the same directory. Supports placeholders. Honors `on_conflict`.                        |
| `run: <command>`     | shell command         | Execute via `sh -c` with a 30s timeout. Supports placeholders; shell operators work (`cmd1 && cmd2`). |
| `delete: true`       | —                     | Delete the file (`true` is required).                                                                 |

### Placeholders

Available in `move`, `copy`, `rename` targets and `run` commands:

| Placeholder   | Expands to                  | Example for `/dl/shot.png` |
| ------------- | --------------------------- | -------------------------- |
| `{file}`      | full source path            | `/dl/shot.png`             |
| `{filename}`  | file name                   | `shot.png`                 |
| `{basename}`  | file name without extension | `shot`                     |
| `{extension}` | extension, with dot         | `.png`                     |
| `{date}`      | current date `YYYY-MM-DD`   | `2026-09-24`               |
| `{timestamp}` | current unix time           | `1758741284`               |

### `on_conflict`

Applies to `move`, `copy`, `rename` (`run` and `delete` have no destination to conflict with):

| Value       | Behavior when the destination exists                                |
| ----------- | ------------------------------------------------------------------- |
| `skip`      | Do nothing; both files are kept.                                    |
| `overwrite` | Replace the destination.                                            |
| `rename`    | Write to a suffixed name instead (`photo-1.png`, `photo-2.png`, …). |

## Evaluation order

1. The first zone (top to bottom) whose path contains the file handles it.
2. Within the zone, the first rule (top to bottom) whose match succeeds runs; the rest are skipped.

Action outputs that land back in a watched directory are processed again like any
new file. Point outputs at a different directory or a non-matching name — a `move`
into the same zone under a still-matching name, or a `rename` that still matches,
loops forever.

## Full example

Note the ordering: specific rules before general ones — the first match wins
(`screenshots` must precede `images`, or `shot-*.png` files would match `images` first).

```yaml
settings:
  stable_for_seconds: 2
  on_conflict: rename

zones:
  - name: downloads
    path: ~/Downloads
    rules:
      - name: screenshots
        match:
          glob: ["shot-*"]
          extensions: [png]
        on_conflict: overwrite
        action:
          rename: "{date}-{filename}"

      - name: images
        match:
          extensions: [jpg, jpeg, png]
        action:
          move: ~/Pictures

      - name: invoices
        match:
          glob: ["invoice-*"]
        action:
          copy: ~/Documents/invoices

      - name: record logs
        match:
          extensions: [log]
        action:
          run: echo "{date} {filename}" >> ~/processed.log

      - name: temp files
        match:
          glob: ["*.tmp"]
        action:
          delete: true
```
