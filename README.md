[English] | [简体中文](README_CN.md)

# Sleet CLI

> Code generation and database tooling for the [Sleet FiveM ORM](https://github.com/SleetCo/sleet).

## Features

-   **`generate`** — Generate LuaLS type annotations from schema files
-   **`sql`** — Generate MySQL `CREATE TABLE` SQL from schema files
-   **`pull`** — Reverse-engineer a MySQL/MariaDB database into a schema.lua

## Installation

### From source (Go 1.21+)

```bash
git clone https://github.com/SleetCo/sleet-orm-cli.git
cd sleet-orm-cli
go build -o sleet .
```

Add the binary to your `PATH`, or run `./sleet` directly.

### Windows (build script)

```batch
build_dev.bat
```

This builds `sleet.exe` and copies it to `C:\Developer\bin` (configurable in the script).

## Usage

### Generate LLS type annotations

Executes your `schema.lua` in an embedded Lua VM, intercepts all `sl.table()` calls, and generates a `---@meta` file with full LuaLS type inference:

```bash
sleet generate schema.lua
sleet generate schema.lua -o .sleet/types.lua
sleet generate schema.lua --stdout
```

Default output: `.sleet/types.lua` in the current directory.

### Generate SQL

Generates MySQL `CREATE TABLE IF NOT EXISTS` statements from your schema:

```bash
sleet sql server/schema.lua
sleet sql server/schema.lua -o database/init.sql
sleet sql server/schema.lua --stdout
```

### Pull schema from database

Connects to MySQL/MariaDB and generates a `schema.lua` from the existing database:

```bash
sleet pull --db myserver
sleet pull --host 127.0.0.1 -u root -p s3cr3t --db myserver -o server/schema.lua
sleet pull --db myserver --stdout
```

| Flag       | Short | Default      | Description                     |
| ---------- | ----- | ------------ | ------------------------------- |
| `--host`   |       | `127.0.0.1`  | Database host                   |
| `--port`   |       | `3306`       | Database port                   |
| `--user`   | `-u`  | `root`       | Database user                   |
| `--pass`   | `-p`  |              | Database password               |
| `--db`     | `-d`  |              | Database name (required)        |
| `--out`    | `-o`  | `schema.lua` | Output file path                |
| `--stdout` |       |              | Print to stdout instead of file |

## Project structure

```
cli/
├── cmd/           # Cobra commands (root, generate, sql, pull)
├── internal/      # Loader, generators, puller, i18n, ui
├── main.go
├── go.mod
└── go.sum
```

## License

See [LICENSE](../LICENSE) in the parent repository.
