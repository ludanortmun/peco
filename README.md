# peco

A small todo app with a CLI and a JSON REST API. Both use the same domain logic
and local SQLite database.

## Install

On macOS arm64 or Linux amd64, install both executables from the latest GitHub
release:

```sh
curl -fsSL https://raw.githubusercontent.com/ludanortmun/peco/main/install.sh | sh
```

The installer verifies the archive against the release's `SHA256SUMS` and installs
`peco` and `peco-server` into `~/.local/bin`. Add that directory to your shell
profile if it is not already on your `PATH`:

```sh
export PATH="$HOME/.local/bin:$PATH"
```

Set `PECO_INSTALL_DIR` to choose another destination:

```sh
curl -fsSL https://raw.githubusercontent.com/ludanortmun/peco/main/install.sh | PECO_INSTALL_DIR=/usr/local/bin sh
```

The destination must be writable by your user. Run the installer again to update
to the latest release. It requires `curl`, `tar`, `awk`, `mktemp`, and either
`sha256sum` or `shasum`.

## Build

Install the Go version specified in `go.mod` (or newer), then run:

```sh
go build -o bin/peco ./cmd/peco
go build -o bin/peco-server ./cmd/peco-server
```

Run `./bin/peco` in the examples below, or add `bin` to your `PATH`.
Release archives include both executables for macOS arm64 and Linux amd64,
plus a `SHA256SUMS` file alongside the archives.

## CLI

```sh
peco new "Buy groceries" --due 2026-12-01
peco new "Read a book"
peco list
peco edit 1 --content "Buy groceries and coffee" --due 2026-12-02
peco complete 1
peco list --include-done
```

Content containing spaces must be quoted. Due dates use `YYYY-MM-DD` at midnight
in your system local timezone. The CLI also displays due dates in that timezone;
the core, database, and REST API use UTC. Editing preserves fields whose flags
are omitted. IDs come from `new` or `list`. Completed items are hidden unless
`--include-done` is supplied.

Lists put pending items first, then the earliest due date (undated items last),
then the oldest creation time. All matching items are returned.

Use `peco --help` or `peco <command> --help` for command details.

## Database

The CLI and server create `peco/peco.db` beneath Go's `os.UserConfigDir()`:

- macOS: `~/Library/Application Support/peco/peco.db`
- Linux: `$XDG_CONFIG_HOME/peco/peco.db`, or `~/.config/peco/peco.db`
- Windows: `%AppData%\peco\peco.db`

The directory and database are created automatically. Both executables share the
database when run as the same user with the same config environment.

## REST API

```sh
peco-server                       # listens on 127.0.0.1:8080
peco-server --addr 127.0.0.1:9000  # choose another address
```

Open [Swagger UI](http://127.0.0.1:8080/docs/) to explore and call the API.
The [OpenAPI document](http://127.0.0.1:8080/openapi.json) is served by the same
binary. Swagger UI loads its assets from a CDN and requires internet access in
the browser. The API itself works offline.

| Method | Path | Behavior |
| --- | --- | --- |
| GET | `/todos` | List pending todos; add `?includeDone=true` to include completed items |
| GET | `/todos/{id}` | Get one todo |
| POST | `/todos` | Create a todo with `content`, optional `due`, and optional `state` |
| PUT | `/todos/{id}` | Update an existing todo using its full JSON representation |

`state` is `0` for pending and `1` for done. API timestamps use RFC3339 and are
returned in UTC (input offsets are converted to UTC); `due` may be `null`. Creation assigns `id` and `createdAt`; they cannot be supplied to
POST or changed through PUT. PUT requires `id`, `content`, `state`, and
`createdAt`; an omitted or null `due` removes the due date.

```sh
curl -X POST http://127.0.0.1:8080/todos \
  -H 'Content-Type: application/json' \
  -d '{"content":"Buy groceries","due":"2026-12-01T00:00:00Z"}'

curl http://127.0.0.1:8080/todos
curl http://127.0.0.1:8080/todos/1
```

To edit or complete a todo, save the GET response to a file, change `content`,
`due`, or `state`, and send it back. Preserve `id` and `createdAt`:

```sh
curl -fsS http://127.0.0.1:8080/todos/1 -o todo.json
# Edit todo.json; for example, set "state" to 1 to mark it complete.
curl -X PUT http://127.0.0.1:8080/todos/1 \
  -H 'Content-Type: application/json' --data-binary @todo.json
```

Successful GET and PUT requests return `200`; POST returns `201` with the saved
todo and a `Location` header. Invalid input or changed immutable fields returns
`400`; missing items return `404`. Errors have empty response bodies. PUT never
creates a missing todo.

## Development and releases

```sh
go test ./internal/core
go vet ./...
go build ./...
```

CI checks formatting, module files, vet, core tests with the race detector, and
builds on pushes and pull requests to `main`.

To release, push a `v*` tag pointing to a commit already on `main`:

```sh
git switch main
git pull --ff-only
git tag v0.1.0
git push origin v0.1.0
```

The release workflow verifies main ancestry, runs CI, builds both executables,
and publishes GitHub Release archives with checksums and generated notes.
