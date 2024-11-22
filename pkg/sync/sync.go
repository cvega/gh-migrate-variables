package sync

import (
	"encoding/csv"
	"fmt"
	"os"

	"gh-migrate-variables/internal/api"

	"github.com/spf13/viper"
)

// SyncVariables handles the syncing of variables from a CSV file to a target organization
func SyncVariables() error {
	mappingFile := viper.GetString("MAPPING_FILE")
	targetOrg := viper.GetString("TARGET_ORGANIZATION")
	targetToken := viper.GetString("TARGET_TOKEN")
	hostname := viper.GetString("TARGET_HOSTNAME")

	if mappingFile == "" || targetOrg == "" || targetToken == "" {
		return fmt.Errorf("missing required parameters: mapping file, target organization, or target token")
	}

	file, err := os.Open(mappingFile)
	if err != nil {
		return fmt.Errorf("cannot open file %s: %v", mappingFile, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("cannot read file %s: %v", mappingFile, err)
	}

	var stats struct {
		total     int
		succeeded int
		failed    int
		skipped   int
	}

	// Skip header row and process variables
	for _, record := range records[1:] {
		stats.total++

		if len(record) < 4 {
			fmt.Printf("Warning: record %v does not have enough columns. Skipping...\n", record)
			stats.skipped++
			continue
		}

		variableName := record[0]
		variableValue := record[1]
		scope := record[2]
		visibility := record[3]

		fmt.Printf("\nProcessing variable - Name: %s, Value: %s, Scope: %s, Visibility: %s\n",
			variableName, variableValue, scope, visibility)

		if scope == "organization" {
			err := api.CreateOrgVariable(targetOrg, variableName, variableValue, visibility, targetToken, hostname)
			if err != nil {
				fmt.Printf("❌ Error creating organization variable %s: %v\n", variableName, err)
				stats.failed++
			} else {
				fmt.Printf("✅ Successfully created organization variable: %s\n", variableName)
				stats.succeeded++
			}
		} else {
			err := api.CreateRepoVariable(targetOrg, scope, variableName, variableValue, visibility, targetToken, hostname)
			if err != nil {
				// Check if the error is due to missing repository
				if err.Error() == fmt.Sprintf("repository %s does not exist in organization %s", scope, targetOrg) {
					fmt.Printf("⚠️  Skipping variable %s: %v\n", variableName, err)
					stats.skipped++
				} else {
					fmt.Printf("❌ Error creating repository variable %s: %v\n", variableName, err)
					stats.failed++
				}
			} else {
				fmt.Printf("✅ Successfully created repository variable: %s in %s\n", variableName, scope)
				stats.succeeded++
			}
		}
	}

	fmt.Printf("\n📊 Sync Summary:\n")
	fmt.Printf("Total variables processed: %d\n", stats.total)
	fmt.Printf("✅ Successfully created: %d\n", stats.succeeded)
	fmt.Printf("❌ Failed: %d\n", stats.failed)
	fmt.Printf("🚧 Skipped: %d\n", stats.skipped)

	if stats.failed > 0 {
		fmt.Printf("\n🛑 sync completed with %d failed variables\n", stats.failed)
		os.Exit(1)
	}

	fmt.Println("\n✅ Sync completed successfully!")
	return nil
}
