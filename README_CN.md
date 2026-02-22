[English](README.md) | [简体中文]

# Sleet CLI

> 为 [Sleet FiveM ORM](https://github.com/SleetCo/sleet) 提供代码生成与数据库工具支持。

## 功能

-   **`generate`** — 从 schema 文件生成 LuaLS 类型注解
-   **`sql`** — 从 schema 文件生成 MySQL `CREATE TABLE` SQL
-   **`pull`** — 从 MySQL/MariaDB 数据库反向生成 schema.lua

## 安装

### 从源码构建（需 Go 1.21+）

```bash
git clone https://github.com/SleetCo/sleet-orm-cli.git
cd sleet-orm-cli
go build -o sleet .
```

将生成的二进制加入 `PATH`，或直接运行 `./sleet`。

### Windows（构建脚本）

```batch
build_dev.bat
```

脚本会编译 `sleet.exe` 并复制到 `C:\Developer\bin`（可在脚本内修改目标路径）。

## 使用说明

### 生成 LLS 类型注解

在内嵌 Lua VM 中执行 `schema.lua`，拦截所有 `sl.table()` 调用，生成包含完整 LuaLS 类型推断链的 `---@meta` 文件：

```bash
sleet generate schema.lua
sleet generate schema.lua -o .sleet/types.lua
sleet generate schema.lua --stdout
```

默认输出：当前目录下的 `.sleet/types.lua`。

### 生成 SQL

从 schema 生成 MySQL `CREATE TABLE IF NOT EXISTS` 语句：

```bash
sleet sql server/schema.lua
sleet sql server/schema.lua -o database/init.sql
sleet sql server/schema.lua --stdout
```

### 从数据库拉取 schema

连接 MySQL/MariaDB，根据现有表结构生成 `schema.lua`：

```bash
sleet pull --db myserver
sleet pull --host 127.0.0.1 -u root -p s3cr3t --db myserver -o server/schema.lua
sleet pull --db myserver --stdout
```

| 参数       | 简写 | 默认值       | 说明                 |
| ---------- | ---- | ------------ | -------------------- |
| `--host`   |      | `127.0.0.1`  | 数据库主机           |
| `--port`   |      | `3306`       | 数据库端口           |
| `--user`   | `-u` | `root`       | 数据库用户           |
| `--pass`   | `-p` |              | 数据库密码           |
| `--db`     | `-d` |              | 数据库名（必填）     |
| `--out`    | `-o` | `schema.lua` | 输出文件路径         |
| `--stdout` |      |              | 输出到控制台而非文件 |

## 项目结构

```
cli/
├── cmd/           # Cobra 命令（root, generate, sql, pull）
├── internal/      # Loader、generators、puller、i18n、ui
├── main.go
├── go.mod
└── go.sum
```

## 许可证

参见父仓库的 [LICENSE](../LICENSE)。
