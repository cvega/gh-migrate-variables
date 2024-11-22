package main

import (
    "github.com/spf13/cobra"
    "gh-migrate-variables/cmd"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "migrate-variables",
        Short: "A GitHub CLI extension for migrating org and repo variables",
    }

    // Add subcommands
    rootCmd.AddCommand(cmd.ExportCmd)
    rootCmd.AddCommand(cmd.SyncCmd)

    // Execute root command
    if err := rootCmd.Execute(); err != nil {
        panic(err)
    }
}

