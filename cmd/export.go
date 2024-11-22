package cmd

import (
	"fmt"

	"github.com/mona-actions/gh-migrate-variables/pkg/export"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// exportCmd represents the export command
var ExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Exports organization and repository variables to CSV",
	Long:  "Exports organization and repository variables to CSV",
	Run: func(cmd *cobra.Command, args []string) {
		requiredConfigs := map[string]string{
			"GHMV_SOURCE_ORGANIZATION": "organization is required (provide via -o flag or GHMV_TARGET_ORGANIZATION env var)",
			"GHMV_SOURCE_TOKEN":        "token is required (provide via -t flag or GHMV_TARGET_TOKEN env var)",
		}

		for key, errMsg := range requiredConfigs {
			if viper.GetString(key) == "" {
				fmt.Printf("%s\n", errMsg)
				return
			}
		}

		ConfigureHostname("GHMV_SOURCE_HOSTNAME")

		if err := export.ExportVariables(); err != nil {
			fmt.Printf("failed to export variables: %v\n", err)
		}
	},
}

func init() {
	// Add flags to the ExportCmd
	ExportCmd.Flags().StringP("source-hostname", "n", "", "GitHub Enterprise Server hostname (optional) Ex. github.example.com")
	ExportCmd.Flags().StringP("source-organization", "o", "", "Organization to export (required)")
	ExportCmd.Flags().StringP("source-token", "t", "", "GitHub token (required)")

	// Bind flags to viper
	viper.BindPFlag("GHMV_SOURCE_HOSTNAME", ExportCmd.Flags().Lookup("source-hostname"))
	viper.BindPFlag("GHMV_SOURCE_ORGANIZATION", ExportCmd.Flags().Lookup("source-organization"))
	viper.BindPFlag("GHMV_SOURCE_TOKEN", ExportCmd.Flags().Lookup("source-token"))
}
