package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SleetCo/sleet-orm-cli/internal/generators"
	"github.com/SleetCo/sleet-orm-cli/internal/i18n"
	"github.com/SleetCo/sleet-orm-cli/internal/loader"
	"github.com/SleetCo/sleet-orm-cli/internal/ui"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate [schema.lua]",
	Short: shortGenerate(),
	Long:  longGenerate(),
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		schemaPath := args[0]

		tables, err := loader.LoadSchema(schemaPath)
		if err != nil {
			ui.Error(i18n.ErrLoadSchema(err))
			return fmt.Errorf("%w", err)
		}

		if len(tables) == 0 {
			ui.Info(i18n.NoTables())
			return nil
		}

		content := generators.GenerateEmmyLua(tables)

		// --stdout: print to terminal
		toStdout, _ := cmd.Flags().GetBool("stdout")
		if toStdout {
			fmt.Print(content)
			ui.Hint(i18n.GenerateStdout())
			return nil
		}

		// Resolve output path — default to .sleet/types.lua in CWD (resource root),
		// regardless of where schema.lua lives.
		outPath, _ := cmd.Flags().GetString("out")
		if outPath == "" {
			outPath = filepath.Join(".sleet", "types.lua")
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			ui.Error(i18n.ErrMkdir(err))
			return fmt.Errorf("%w", err)
		}

		if err := os.WriteFile(outPath, []byte(content), 0o644); err != nil {
			ui.Error(i18n.ErrWrite(err))
			return fmt.Errorf("%w", err)
		}

		ui.Success(i18n.GenerateDone(len(tables), outPath))
		return nil
	},
}

func shortGenerate() string {
	if i18n.IsZh() {
		return "从 schema 文件生成 LLS 类型注解"
	}
	return "Generate LLS type annotations from a schema file"
}

func longGenerate() string {
	if i18n.IsZh() {
		return `在内嵌 Lua VM 中执行 schema.lua, 拦截所有 sl.table() 调用,
输出 ---@meta 文件, 包含完整的 LuaLS 类型注解链:

  XxxRecord        — SELECT 结果行形状 (字段类型 + 列描述)
  XxxTable         — ColumnDef<T> 字段的 schema 对象
  XxxSelectBuilder — 每张表专属构建器, execute() 返回 XxxRecord[]
  Sleet.table      — 每张表的 @overload, sl.table('x', ...) → XxxTable

默认输出到当前工作目录 (资源根目录) 的 .sleet/types.lua
使用 --stdout 输出到控制台

示例 (在资源根目录执行):
  sleet generate schema.lua
  sleet generate schema.lua -o .sleet/types.lua
  sleet generate schema.lua --stdout`
	}
	return `Executes your schema.lua in an embedded Lua VM, intercepts all
sl.table() calls, and generates a ---@meta file with the full LuaLS
type inference chain:

  XxxRecord        — row shape for SELECT results (typed fields + descriptions)
  XxxTable         — schema object with ColumnDef<T> fields
  XxxSelectBuilder — per-table builder; execute() returns XxxRecord[]
  Sleet.table      — @overload per table so sl.table('x', ...) → XxxTable

Writes to .sleet/types.lua in the current directory (resource root) by default.
Use --stdout to print to the terminal instead.

Examples (run from the resource root):
  sleet generate schema.lua
  sleet generate schema.lua -o .sleet/types.lua
  sleet generate schema.lua --stdout`
}

func init() {
	generateCmd.Flags().StringP("out", "o", "", "Output file path (default: .sleet/types.lua in CWD)")
	generateCmd.Flags().Bool("stdout", false, "Print to stdout instead of writing a file")
}
