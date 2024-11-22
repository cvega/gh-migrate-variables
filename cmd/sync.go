package cmd

import (
    "os"
    
    "gh-migrate-variables/pkg/sync"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var SyncCmd = &cobra.Command{
    Use:   "sync",
    Short: "Recreates variables from a CSV file to a target organization",
    Long:  "Sync variables from a previously exported CSV file to a target GitHub organization, including organization-level and repository-level variables.",
    Run: func(cmd *cobra.Command, args []string) {
        // Get parameters
        mappingFile := cmd.Flag("mapping-file").Value.String()
        targetOrg := cmd.Flag("target-organization").Value.String()
        targetToken := cmd.Flag("target-token").Value.String()

        // Set ENV variables
        os.Setenv("GHMV_MAPPING_FILE", mappingFile)
        os.Setenv("GHMV_TARGET_ORGANIZATION", targetOrg)
        os.Setenv("GHMV_TARGET_TOKEN", targetToken)

        // Bind ENV variables in Viper
        viper.BindEnv("MAPPING_FILE")
        viper.BindEnv("TARGET_ORGANIZATION")
        viper.BindEnv("TARGET_TOKEN")

        // Call sync function from the pkg/sync package
        sync.SyncVariables()
    },
}

func init() {
    rootCmd.AddCommand(SyncCmd)

    // Flags
    SyncCmd.Flags().StringP("mapping-file", "m", "", "Mapping file path to use for syncing variables (required)")
    SyncCmd.MarkFlagRequired("mapping-file")

    SyncCmd.Flags().StringP("target-organization", "t", "", "Target Organization to sync variables to (required)")
    SyncCmd.MarkFlagRequired("target-organization")

    SyncCmd.Flags().StringP("target-token", "b", "", "Target Organization GitHub token. Scopes: admin:org (required)")
    SyncCmd.MarkFlagRequired("target-token")
}
