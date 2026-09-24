# dropzone

Watch, match, and act on your files.

## Install

Requires Go. Works on Linux and macOS (`service` commands need Linux with systemd).

```sh
go install github.com/nnavales/dropzone/cmd/dropzone@latest
```

## Quickstart

```sh
dropzone init          # creates ~/.config/dropzone/config.yml
# edit ~/.config/dropzone/config.yml, then:
dropzone               # foreground (same as dropzone run)
```

## Configuration

```yaml
settings:
  stable_for_seconds: 2 # act after a file is unchanged for this long
  on_conflict: rename # skip | overwrite | rename (default)

zones:
  - name: downloads
    path: ~/Downloads
    rules:
      - name: images
        match:
          extensions: [jpg, png]
          glob: ["shot-*"]
        action:
          move: ~/Pictures
```

Full configuration reference (match semantics, actions, placeholders, conflicts): [`docs/configuration.md`](docs/configuration.md).

### Rule and zone order

- Rules are evaluated top to bottom; the **first matching rule wins**.
- Zones are matched by path prefix, top to bottom; the **first match wins**. For nested zones, list the inner zone first.
- Two zones cannot share the same path.

### Avoiding loops

Action outputs are processed like any new file. If an action produces a file that still matches the same rule in a watched directory, it may loop indefinitely. Use a different destination or a non-matching output name.

Config errors fail fast at startup with the file location. The daemon also reloads the config automatically when the file changes.

## Run as a service

Dropzone runs as a systemd user service and starts automatically when you log in.

```sh
dropzone service install
dropzone service status
dropzone service logs
dropzone service uninstall
```

## Commands

| Command                      | Description                        |
| ---------------------------- | ---------------------------------- |
| `dropzone init`              | Create the config file             |
| `dropzone run`               | Start the daemon in the foreground |
| `dropzone service <command>` | Manage the systemd user service    |
| `dropzone version`           | Print the version                  |
