// cmd/export.go

package cmd

import (
    "fmt"
    "os"
    "strings"

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

        if hostname != "" {
            // Clean the hostname by removing any protocol and api/v3 if present
            hostname = strings.TrimPrefix(hostname, "http://")
            hostname = strings.TrimPrefix(hostname, "https://")
            hostname = strings.TrimSuffix(hostname, "/api/v3")
            hostname = strings.TrimSuffix(hostname, "/")
            hostname = fmt.Sprintf("https://%s/api/v3", hostname)
        }

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
            fmt.Printf("\n🔗 Using GitHub Enterprise Server: %s\n", hostname)
        } else {
            fmt.Println("\n📡 Using GitHub.com")
        }

        httpProxy := viper.GetString("HTTP_PROXY")
        httpsProxy := viper.GetString("HTTPS_PROXY")
        if httpProxy != "" || httpsProxy != "" {
            fmt.Println("🔄 Proxy: ✅ Configured\n")
        } else {
            fmt.Println("🔄 Proxy: ❌ Not configured\n")
        }

        if err := export.CreateCSVs(); err != nil {
            return fmt.Errorf("failed to export variables: %w", err)
        }

        return nil
    },
}

func init() {
    ExportCmd.Flags().StringP("organization", "o", "", "Organization to export (required)")
    ExportCmd.MarkFlagRequired("organization")

    ExportCmd.Flags().StringP("token", "t", "", "GitHub token (required)")
    ExportCmd.MarkFlagRequired("token")

    ExportCmd.Flags().StringP("file-prefix", "f", "", "Output filenames prefix")
    
    ExportCmd.Flags().StringP("hostname", "n", "", "GitHub Enterprise Server hostname URL (optional) Ex. https://github.example.com")
}