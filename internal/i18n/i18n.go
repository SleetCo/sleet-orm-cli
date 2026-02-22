// Package i18n provides bilingual (zh/en) message lookup for sleet-cli.
// Language is auto-detected from env vars. Override with SLEET_LANG=zh or SLEET_LANG=en.
package i18n

import (
	"os"
	"runtime"
	"strings"
)

// ── Language detection ────────────────────────────────────────────────────────

var lang string

func init() {
	lang = detect()
}

func detect() string {
	// Manual override
	if v := os.Getenv("SLEET_LANG"); v != "" {
		if strings.HasPrefix(strings.ToLower(v), "zh") {
			return "zh"
		}
		return "en"
	}

	// Standard POSIX locale env vars
	for _, env := range []string{"LANG", "LANGUAGE", "LC_ALL", "LC_MESSAGES"} {
		if v := strings.ToLower(os.Getenv(env)); strings.Contains(v, "zh") {
			return "zh"
		}
	}

	// Windows: check UI language via LOCALAPPDATA path heuristic,
	//    or the ComSpec/APPDATA culture embedded in path (fragile but dependency-free).
	//    Prefer checking common Windows Chinese locale markers.
	if runtime.GOOS == "windows" {
		// Windows Terminal sets WT_SESSION; check APPDATA for \zh- path segment
		for _, env := range []string{"APPDATA", "USERPROFILE"} {
			if strings.Contains(strings.ToLower(os.Getenv(env)), `\zh`) {
				return "zh"
			}
		}
	}

	return "en"
}

// IsZh reports whether the detected language is Chinese.
func IsZh() bool { return lang == "zh" }

// ── Message catalog ───────────────────────────────────────────────────────────

type msgs struct{ zh, en string }

func t(m msgs) string {
	if lang == "zh" {
		return m.zh
	}
	return m.en
}

// ── generate ──────────────────────────────────────────────────────────────────

func GenerateDone(n int, path string) string {
	return t(msgs{
		zh: sf("已生成类型注解 (%d 张表)  → %s", n, path),
		en: sf("Generated type annotations (%d tables) → %s", n, path),
	})
}

func GenerateStdout() string {
	return t(msgs{zh: " (已输出至控制台) ", en: "(printed to stdout)"})
}

func NoTables() string {
	return t(msgs{zh: "未在 schema 文件中找到任何表定义", en: "No tables found in schema file."})
}

// ── sql ───────────────────────────────────────────────────────────────────────

func SQLDone(n int, path string) string {
	return t(msgs{
		zh: sf("已生成 SQL (%d 张表)  → %s", n, path),
		en: sf("Generated SQL (%d tables) → %s", n, path),
	})
}

// ── pull ──────────────────────────────────────────────────────────────────────

func PullConnecting(user, host string, port int, db string) string {
	return t(msgs{
		zh: sf("正在连接 %s@%s:%d/%s ...", user, host, port, db),
		en: sf("Connecting to %s@%s:%d/%s ...", user, host, port, db),
	})
}

func PullDone(n int, path string) string {
	return t(msgs{
		zh: sf("已生成 schema (%d 张表)  → %s", n, path),
		en: sf("Generated schema (%d tables) → %s", n, path),
	})
}

func NoDBTables() string {
	return t(msgs{zh: "数据库中未找到任何表", en: "No tables found in database."})
}

// ── errors ────────────────────────────────────────────────────────────────────

func ErrLoadSchema(err error) string {
	return t(msgs{
		zh: sf("加载 schema 失败: %v", err),
		en: sf("Failed to load schema: %v", err),
	})
}

func ErrMkdir(err error) string {
	return t(msgs{
		zh: sf("创建输出目录失败: %v", err),
		en: sf("Failed to create output directory: %v", err),
	})
}

func ErrWrite(err error) string {
	return t(msgs{
		zh: sf("写入文件失败: %v", err),
		en: sf("Failed to write output file: %v", err),
	})
}

func ErrPull(err error) string {
	return t(msgs{
		zh: sf("数据库拉取失败: %v", err),
		en: sf("Failed to pull from database: %v", err),
	})
}

// ── hints ─────────────────────────────────────────────────────────────────────

func HintGenerate() string {
	return t(msgs{
		zh: "提示: 运行 sleet generate schema.lua 后, LuaLS 可自动推断查询结果类型, 无需手写 ---@type",
		en: "Tip: After running sleet generate schema.lua, LuaLS infers query result types automatically — no manual ---@type needed.",
	})
}

func sf(format string, a ...any) string {
	// We just use fmt — the i18n package is internal-only.
	return sprintf(format, a...)
}
