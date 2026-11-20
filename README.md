# peco

A small todo app with a CLI, a JSON REST API, and a stdio MCP server. All three
use the same domain logic and local SQLite database.

## Install

On macOS arm64 or Linux amd64, install all three executables from the latest GitHub
release:

```sh
curl -fsSL https://raw.githubusercontent.com/ludanortmun/peco/main/install.sh | sh
```

The installer verifies the archive against the release's `SHA256SUMS` and installs
`peco`, `peco-server`, and `peco-mcp` into `~/.local/bin`. Add that directory to your shell
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
go build -o bin/peco-mcp ./cmd/peco-mcp
```

Run `./bin/peco` in the examples below, or add `bin` to your `PATH`.
Release archives include all three executables for macOS arm64 and Linux amd64,
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

## MCP server

Run `peco-mcp` to serve MCP over stdin/stdout using the official
[MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk). Configure your MCP
host to launch it, for example:

```json
{
  "mcpServers": {
    "peco": {
      "command": "/absolute/path/to/peco-mcp"
    }
  }
}
```

Replace the command with the installed binary's absolute path. If upgrading from
v0.2.0-beta, replace the `peco` command and `["mcp"]` arguments with this standalone
command and no arguments. The host starts
and stops the process; no HTTP server is needed. Stdout is reserved for MCP
messages, and startup errors go to stderr.

The server calls the same core domain logic and uses the same database as the
CLI and REST API when launched with the same user and config environment.

| Tool | Arguments | Behavior |
| --- | --- | --- |
| `list_todos` | `include_done` (optional, defaults to false) | Return IDs, state, local due dates, and content in the CLI's sort order |
| `create_todo` | `content`, optional `due` | Create a pending TODO and return its ID |
| `edit_todo` | `todo_id`, optional `content`, optional `due` | Update an existing TODO, preserving fields when omitted or null |
| `complete_todo` | `todo_id` | Mark an existing TODO as done |

Results are plain text. Due dates use local `YYYY-MM-DD`, as in the CLI; empty or
invalid dates are rejected. Omitting or passing null for `due` preserves the
date on edit and creates an undated item on creation. These tools do not clear
an existing due date. Invalid input and missing TODOs return MCP tool errors.

## Database

The CLI, HTTP server, and MCP server create `peco/peco.db` beneath Go's `os.UserConfigDir()`:

- macOS: `~/Library/Application Support/peco/peco.db`
- Linux: `$XDG_CONFIG_HOME/peco/peco.db`, or `~/.config/peco/peco.db`
- Windows: `%AppData%\peco\peco.db`

The directory and database are created automatically. All three executables share the
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

The release workflow verifies main ancestry, runs CI, builds all three executables,
and publishes GitHub Release archives with checksums and generated notes.
