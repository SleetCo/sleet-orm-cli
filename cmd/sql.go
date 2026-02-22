package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SleetCo/sleet-orm-cli/internal/generators"
	"github.com/SleetCo/sleet-orm-cli/internal/i18n"
	"github.com/SleetCo/sleet-orm-cli/internal/loader"
	"github.com/SleetCo/sleet-orm-cli/internal/ui"
	"github.com/spf13/cobra"
)

var sqlCmd = &cobra.Command{
	Use:   "sql [schema.lua]",
	Short: shortSQL(),
	Long:  longSQL(),
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

		content := generators.GenerateSQL(tables)

		// --stdout: print to terminal
		toStdout, _ := cmd.Flags().GetBool("stdout")
		if toStdout {
			fmt.Print(content)
			return nil
		}

		// Resolve output path
		outPath, _ := cmd.Flags().GetString("out")
		if outPath == "" {
			base := strings.TrimSuffix(filepath.Base(schemaPath), filepath.Ext(schemaPath))
			outPath = filepath.Join(filepath.Dir(schemaPath), base+".sql")
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			ui.Error(i18n.ErrMkdir(err))
			return fmt.Errorf("%w", err)
		}

		if err := os.WriteFile(outPath, []byte(content), 0o644); err != nil {
			ui.Error(i18n.ErrWrite(err))
			return fmt.Errorf("%w", err)
		}

		ui.Success(i18n.SQLDone(len(tables), outPath))
		return nil
	},
}

func shortSQL() string {
	if i18n.IsZh() {
		return "从 schema 文件生成 CREATE TABLE SQL"
	}
	return "Generate CREATE TABLE SQL from a schema file"
}

func longSQL() string {
	if i18n.IsZh() {
		return `在内嵌 Lua VM 中执行 schema.lua, 生成包含主键、
NOT NULL、默认值、外键约束的 MySQL CREATE TABLE IF NOT EXISTS 语句

默认写入 schema 同目录的 .sql 文件，使用 --stdout 输出到控制台。

示例：
  sleet sql server/schema.lua
  sleet sql server/schema.lua -o database/init.sql
  sleet sql server/schema.lua --stdout`
	}
	return `Executes your schema.lua in an embedded Lua VM and generates
MySQL CREATE TABLE IF NOT EXISTS statements, including primary keys,
NOT NULL constraints, default values, and foreign key references.

Writes to a .sql file next to the schema by default.
Use --stdout to print to the terminal instead.

Examples:
  sleet sql server/schema.lua
  sleet sql server/schema.lua -o database/init.sql
  sleet sql server/schema.lua --stdout`
}

func init() {
	sqlCmd.Flags().StringP("out", "o", "", "Output file path (default: <schema-name>.sql next to schema)")
	sqlCmd.Flags().Bool("stdout", false, "Print SQL to stdout instead of writing a file")
}
