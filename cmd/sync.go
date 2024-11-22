package cmd

import (
	"fmt"

	"github.com/mona-actions/gh-migrate-variables/pkg/sync"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync organization and repository variables from CSV",
	Long:  "Sync organization and repository variables from CSV",
	Run: func(cmd *cobra.Command, args []string) {
		requiredConfigs := map[string]string{
			"GHMV_TARGET_ORGANIZATION": "organization is required (provide via -o flag or GHMV_TARGET_ORGANIZATION env var)",
			"GHMV_TARGET_TOKEN":        "token is required (provide via -t flag or GHMV_TARGET_TOKEN env var)",
			"GHMV_CSV_FILE":            "CSV file is required",
		}

		for key, errMsg := range requiredConfigs {
			if viper.GetString(key) == "" {
				fmt.Printf("%s\n", errMsg)
				return
			}
		}

		ConfigureHostname("GHMV_TARGET_HOSTNAME")

		if err := sync.SyncVariables(); err != nil {
			fmt.Printf("failed to export variables: %v\n", err)
		}
	},
}

func init() {
	// Add flags to the SyncCmd
	SyncCmd.Flags().StringP("file", "f", "", "Input CSV file with variables to sync")
	SyncCmd.Flags().StringP("target-hostname", "n", "", "GitHub Enterprise Server hostname URL (optional) Ex. https://github.example.com")
	SyncCmd.Flags().StringP("target-organization", "o", "", "Organization to export (required)")
	SyncCmd.Flags().StringP("target-token", "t", "", "GitHub token (required)")

	// Bind flags to viper
	viper.BindPFlag("GHMV_TARGET_HOSTNAME", SyncCmd.Flags().Lookup("target-hostname"))
	viper.BindPFlag("GHMV_TARGET_ORGANIZATION", SyncCmd.Flags().Lookup("target-organization"))
	viper.BindPFlag("GHMV_TARGET_TOKEN", SyncCmd.Flags().Lookup("target-token"))
	viper.BindPFlag("GHMV_CSV_FILE", SyncCmd.Flags().Lookup("file"))
}
