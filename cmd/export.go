package cmd

import (
    "fmt"
    "os"

    "gh-migrate-variables/pkg/export"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var ExportCmd = &cobra.Command{
    Use:   "export",
    Short: "Creates a CSV file of the organization and repository variables",
    Long:  "Creates a CSV file of the organization and repository variables",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Get parameters
        organization := cmd.Flag("organization").Value.String()
        token := cmd.Flag("token").Value.String()
        filePrefix := cmd.Flag("file-prefix").Value.String()
        hostname := cmd.Flag("hostname").Value.String()

        if filePrefix == "" {
            filePrefix = organization
        }

        // Set ENV variables
        os.Setenv("GHMV_SOURCE_ORGANIZATION", organization)
        os.Setenv("GHMV_SOURCE_TOKEN", token)
        os.Setenv("GHMV_OUTPUT_FILE", filePrefix)
        os.Setenv("GHMV_SOURCE_HOSTNAME", hostname)

        // Bind ENV variables in Viper
        viper.BindEnv("SOURCE_ORGANIZATION", "GHMV_SOURCE_ORGANIZATION")
        viper.BindEnv("SOURCE_TOKEN", "GHMV_SOURCE_TOKEN")
        viper.BindEnv("OUTPUT_FILE", "GHMV_OUTPUT_FILE")
        viper.BindEnv("SOURCE_HOSTNAME", "GHMV_SOURCE_HOSTNAME")

        if hostname != "" {
            fmt.Printf("🌐 Using GitHub Enterprise Server: %s\n", hostname)
        } else {
            fmt.Println("🌐 Using GitHub.com")
        }

        if err := export.CreateCSVs(); err != nil {
            return fmt.Errorf("export failed: %w", err)
        }
        
        return nil
    },
}

func init() {
    rootCmd.AddCommand(ExportCmd)

    // Export command flags
    ExportCmd.Flags().StringP("organization", "o", "", "Organization to export (required)")
    ExportCmd.MarkFlagRequired("organization")

    ExportCmd.Flags().StringP("token", "t", "", "GitHub token (required)")
    ExportCmd.MarkFlagRequired("token")

    ExportCmd.Flags().StringP("file-prefix", "f", "", "Output filenames prefix")

    ExportCmd.Flags().StringP("hostname", "u", "", "GitHub Enterprise Server hostname URL (optional) Ex. https://github.example.com")
}