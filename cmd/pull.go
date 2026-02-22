package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SleetCo/sleet-orm-cli/internal/generators"
	"github.com/SleetCo/sleet-orm-cli/internal/i18n"
	"github.com/SleetCo/sleet-orm-cli/internal/puller"
	"github.com/SleetCo/sleet-orm-cli/internal/ui"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: shortPull(),
	Long:  longPull(),
	RunE: func(cmd *cobra.Command, args []string) error {
		host, _    := cmd.Flags().GetString("host")
		port, _    := cmd.Flags().GetInt("port")
		user, _    := cmd.Flags().GetString("user")
		pass, _    := cmd.Flags().GetString("pass")
		dbName, _  := cmd.Flags().GetString("db")
		outPath, _ := cmd.Flags().GetString("out")
		toStdout, _ := cmd.Flags().GetBool("stdout")

		ui.Step(i18n.PullConnecting(user, host, port, dbName))

		tables, err := puller.Pull(puller.Config{
			Host:   host,
			Port:   port,
			User:   user,
			Pass:   pass,
			DBName: dbName,
		})
		if err != nil {
			ui.Error(i18n.ErrPull(err))
			return fmt.Errorf("%w", err)
		}

		if len(tables) == 0 {
			ui.Info(i18n.NoDBTables())
			return nil
		}

		content := generators.GenerateSchemaLua(tables)

		// --stdout: print to terminal (useful for piping / inspection)
		if toStdout {
			fmt.Print(content)
			return nil
		}

		// Resolve output path: explicit -o flag, or default to schema.lua in CWD
		if outPath == "" {
			outPath = "schema.lua"
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			ui.Error(i18n.ErrMkdir(err))
			return fmt.Errorf("%w", err)
		}

		if err := os.WriteFile(outPath, []byte(content), 0o644); err != nil {
			ui.Error(i18n.ErrWrite(err))
			return fmt.Errorf("%w", err)
		}

		ui.Success(i18n.PullDone(len(tables), outPath))
		ui.Hint(i18n.HintGenerate())
		return nil
	},
}

func shortPull() string {
	if i18n.IsZh() {
		return "从数据库反向生成 schema.lua"
	}
	return "Introspect a database and generate a schema.lua"
}

func longPull() string {
	if i18n.IsZh() {
		return `连接到 MySQL/MariaDB 数据库, 读取 information_schema,
生成可直接用于 Sleet 的 schema.lua 文件

默认写入 <dbName>_schema.lua (当前目录), 使用 --stdout 输出到控制台

示例：
  sleet pull --db myserver
  sleet pull --host 127.0.0.1 -u root -p s3cr3t --db myserver -o server/schema.lua
  sleet pull --db myserver --stdout`
	}
	return `Connects to a live MySQL/MariaDB database, reads information_schema,
and generates a Sleet schema.lua that mirrors the existing table structure.

Writes to <dbName>_schema.lua in the current directory by default.
Use --stdout to print to the terminal instead (useful for piping).

Examples:
  sleet pull --db myserver
  sleet pull --host 127.0.0.1 -u root -p s3cr3t --db myserver -o server/schema.lua
  sleet pull --db myserver --stdout`
}

func init() {
	pullCmd.Flags().String("host", "127.0.0.1", "Database host")
	pullCmd.Flags().Int("port", 3306, "Database port")
	pullCmd.Flags().StringP("user", "u", "root", "Database user")
	pullCmd.Flags().StringP("pass", "p", "", "Database password")
	pullCmd.Flags().StringP("db", "d", "", "Database name (required)")
	pullCmd.Flags().StringP("out", "o", "", "Output file path (default: <dbName>_schema.lua)")
	pullCmd.Flags().Bool("stdout", false, "Print to stdout instead of writing a file")
	pullCmd.MarkFlagRequired("db") //nolint:errcheck
}
