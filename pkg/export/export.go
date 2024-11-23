package export

import (
    "encoding/csv"
    "fmt"
    "os"

    "gh-migrate-variables/internal/api"

    "github.com/spf13/viper"
)

func CreateCSVs() error {
    // Validate environment variables
    organization := viper.GetString("SOURCE_ORGANIZATION")
    token := viper.GetString("SOURCE_TOKEN")
    filePrefix := viper.GetString("OUTPUT_FILE")
    hostname := viper.GetString("SOURCE_HOSTNAME")

    if organization == "" || token == "" || filePrefix == "" {
        return fmt.Errorf("missing required environment variables: SOURCE_ORGANIZATION, SOURCE_TOKEN, or OUTPUT_FILE")
    }

    var allVariables []map[string]string

    // Fetch organization variables
    fmt.Printf("Fetching organization variables for %s...\n", organization)
    orgVariables, err := api.GetOrgVariables(organization, token, hostname)
    if err != nil {
        fmt.Printf("Warning: Failed to fetch organization variables: %v\n", err)
    } else {
        fmt.Printf("Found %d organization variables\n", len(orgVariables))
        allVariables = append(allVariables, orgVariables...)
    }

    // Fetch repositories
    fmt.Printf("Fetching repository list for %s...\n", organization)
    repos, err := api.GetRepositories(organization, token, hostname)
    if err != nil {
        return fmt.Errorf("failed to fetch repositories: %w", err)
    }
    fmt.Printf("Found %d repositories\n", len(repos))

    // Process each repository
    var successful, failed int
    for _, repo := range repos {
        fmt.Printf("Processing repository %s...\n", repo)
        repoVariables, err := api.GetRepoVariables(organization, repo, token, hostname)
        if err != nil {
            fmt.Printf("Warning: Failed to fetch variables for repo %s: %v\n", repo, err)
            failed++
            continue
        }
        
        if len(repoVariables) > 0 {
            allVariables = append(allVariables, repoVariables...)
            fmt.Printf("Found %d variables in repository %s\n", len(repoVariables), repo)
            successful++
        } else {
            fmt.Printf("No variables found in repository %s\n", repo)
            successful++
        }
    }

    // Exit if no variables found
    if len(allVariables) == 0 {
        fmt.Println("\nNo variables found to export.")
        return nil
    }

    // Create and write to CSV file
    outputFile := filePrefix + "_variables.csv"
    file, err := os.Create(outputFile)
    if err != nil {
        return fmt.Errorf("cannot create file %s: %w", outputFile, err)
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Write header
    if err := writer.Write([]string{"Name", "Value", "Scope", "Visibility"}); err != nil {
        return fmt.Errorf("failed to write CSV header: %w", err)
    }

    // Write variables
    variablesWritten := 0
    for _, variable := range allVariables {
        if name, ok := variable["Name"]; ok && name != "" {
            value := variable["Value"]
            scope := variable["Scope"]
            visibility := variable["Visibility"]
            if err := writer.Write([]string{name, value, scope, visibility}); err != nil {
                return fmt.Errorf("failed to write variable to CSV: %w", err)
            }
            variablesWritten++
        }
    }

    // Print summary
    fmt.Printf("\n📊 Export Summary:\n")
    fmt.Printf("Total repositories found: %d\n", len(repos))
    fmt.Printf("✅ Successfully processed: %d repositories\n", successful)
    fmt.Printf("❌ Failed to process: %d repositories\n", failed)
    fmt.Printf("📝 Total variables exported: %d\n", variablesWritten)
    fmt.Printf("📁 Output file: %s\n", outputFile)

    if failed > 0 {
        fmt.Printf("\n🛑  Export completed with some failures. Some variables may not have been exported.\n")
        fmt.Printf("export completed with %d failed repositories", failed)
        os.Exit(1)
    }

    fmt.Println("\n✅ Export completed successfully!")
    return nil
}