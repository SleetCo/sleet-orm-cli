package cmd

import (
	"os"

	"github.com/SleetCo/sleet-orm-cli/internal/i18n"
	"github.com/SleetCo/sleet-orm-cli/internal/ui"
	"github.com/spf13/cobra"
)

const version = "v0.1.1"

var rootCmd = &cobra.Command{
	Use:     "sleet",
	Short:   shortRoot(),
	Long:    longRoot(),
	Version: version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Don't print banner for --help / --version / completion
		if cmd.Name() == "help" || cmd.Name() == "completion" {
			return
		}
		ui.Banner(version)
	},
}

// Execute is the entry point called by main.go.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
}

func shortRoot() string {
	if i18n.IsZh() {
		return "Sleet CLI — FiveM ORM 工具链"
	}
	return "Sleet CLI — tooling for the Sleet FiveM ORM"
}

func longRoot() string {
	if i18n.IsZh() {
		return `❄  Sleet CLI

为 Sleet FiveM ORM 提供代码生成和数据库工具支持

  sleet generate server/schema.lua          生成 LLS 类型注解
  sleet sql      server/schema.lua          生成 CREATE TABLE SQL
  sleet pull     --host 127.0.0.1 --db mydb  反向工程数据库`
	}
	return `❄  Sleet CLI

Code generation and database tooling for the Sleet FiveM ORM.

  sleet generate server/schema.lua          Generate LLS type annotations
  sleet sql      server/schema.lua          Generate CREATE TABLE SQL
  sleet pull     --host 127.0.0.1 --db mydb  Reverse-engineer a database`
}

func init() {
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(sqlCmd)
	rootCmd.AddCommand(pullCmd)
}
