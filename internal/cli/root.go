package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "jd",
	Short: "Claude Code configuration manager",
	Version: Version,
	Long: `jd is a CLI tool for managing Claude Code configurations
including skills, commands, agents, and hooks.

Default scope: local (.claude) if present, otherwise global (~/.claude).

Subcommand aliases: skills(s), commands(c), agents(a), hooks(h), pkg(p), list(l)
Common subcommand aliases: list(l,ls), new(n,add,create), show(s,get,view), edit(e,update,modify), delete(d,rm,remove)

Use 'jd --help' for all available commands.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
