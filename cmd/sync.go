package cmd

import (
	"fmt"
	"os"
	"strings"

	"gh-migrate-variables/pkg/sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Syncs variables from a CSV file to a target organization",
	Long:  "Syncs variables from a CSV file to a target organization",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get parameters
		mappingFile := cmd.Flag("file-mapping").Value.String()
		targetOrg := cmd.Flag("target-organization").Value.String()
		targetToken := cmd.Flag("target-token").Value.String()
		hostname := cmd.Flag("hostname").Value.String()

		if hostname != "" {
			hostname = strings.TrimPrefix(hostname, "http://")
			hostname = strings.TrimPrefix(hostname, "https://")
			hostname = strings.TrimSuffix(hostname, "/api/v3")
			hostname = strings.TrimSuffix(hostname, "/")
			hostname = fmt.Sprintf("https://%s/api/v3", hostname)
		}

		// Set ENV variables
		os.Setenv("GHMV_MAPPING_FILE", mappingFile)
		os.Setenv("GHMV_TARGET_ORGANIZATION", targetOrg)
		os.Setenv("GHMV_TARGET_TOKEN", targetToken)
		os.Setenv("GHMV_TARGET_HOSTNAME", hostname)

		// Bind ENV variables in Viper
		viper.BindEnv("MAPPING_FILE", "GHMV_MAPPING_FILE")
		viper.BindEnv("TARGET_ORGANIZATION", "GHMV_TARGET_ORGANIZATION")
		viper.BindEnv("TARGET_TOKEN", "GHMV_TARGET_TOKEN")
		viper.BindEnv("TARGET_HOSTNAME", "GHMV_TARGET_HOSTNAME")

		if hostname != "" {
			fmt.Printf("\n🔗 Using GitHub Enterprise Server: %s\n", hostname)
		} else {
			fmt.Println("\n📡 Using GitHub.com")
		}

		httpProxy := viper.GetString("HTTP_PROXY")
		httpsProxy := viper.GetString("HTTPS_PROXY")
		if httpProxy != "" || httpsProxy != "" {
			fmt.Println("🔄 Proxy: ✅ Configured")
		} else {
			fmt.Println("🔄 Proxy: ❌ Not configured")
		}

		if err := sync.SyncVariables(); err != nil {
			return fmt.Errorf("failed to sync variables: %w", err)
		}

		return nil
	},
}

func init() {
	SyncCmd.Flags().StringP("file-mapping", "f", "", "CSV mapping file path to use for syncing variables (required)")
	SyncCmd.MarkFlagRequired("file-mapping")

	SyncCmd.Flags().StringP("target-organization", "o", "", "Target Organization to sync variables to (required)")
	SyncCmd.MarkFlagRequired("target-organization")

	SyncCmd.Flags().StringP("target-token", "t", "", "Target Organization GitHub token. Scopes: admin:org (required)")
	SyncCmd.MarkFlagRequired("target-token")

	SyncCmd.Flags().StringP("hostname", "n", "", "GitHub Enterprise Server hostname URL (optional) Ex. https://github.example.com")
}
