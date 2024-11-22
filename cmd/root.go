package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "migrate-variables",
	Short: "gh cli extension to assist in the migration of variables between GitHub enterprises",
	Long:  "gh cli extension to assist in the migration of variables between GitHub enterprises",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Add root command flags
	rootCmd.PersistentFlags().String("http-proxy", "", "HTTP proxy (can also use HTTP_PROXY env var)")
	rootCmd.PersistentFlags().String("https-proxy", "", "HTTPS proxy (can also use HTTPS_PROXY env var)")
	rootCmd.PersistentFlags().String("no-proxy", "", "No proxy list (can also use NO_PROXY env var)")
	rootCmd.PersistentFlags().Int("retry-max", 3, "Maximum retry attempts")
	rootCmd.PersistentFlags().String("retry-delay", "1s", "Delay between retries")

	// Bind flags to viper
	viper.BindPFlag("HTTP_PROXY", rootCmd.PersistentFlags().Lookup("http-proxy"))
	viper.BindPFlag("HTTPS_PROXY", rootCmd.PersistentFlags().Lookup("https-proxy"))
	viper.BindPFlag("NO_PROXY", rootCmd.PersistentFlags().Lookup("no-proxy"))
	viper.BindPFlag("RETRY_MAX", rootCmd.PersistentFlags().Lookup("retry-max"))
	viper.BindPFlag("RETRY_DELAY", rootCmd.PersistentFlags().Lookup("retry-delay"))

	// Add subcommands
	rootCmd.AddCommand(ExportCmd)
	rootCmd.AddCommand(SyncCmd)

	// hide -h, --help from global/proxy flags
	rootCmd.Flags().BoolP("help", "h", false, "")
	rootCmd.Flags().Lookup("help").Hidden = true
}

func initConfig() {
	// Set up .env file reading first
	viper.SetConfigType("env")
	viper.SetConfigName(".env")
	viper.AddConfigPath(".")

	// Try to read the .env file first
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("\n🚩 Using: flags and/or environment variables")
		} else {
			fmt.Printf("Error reading .env file: %v\n", err)
		}
	} else {
		fmt.Printf("\n📄 Using: .env file for configuration")
	}

	// Set up environment handling
	viper.SetEnvPrefix("GHMV")
	viper.AutomaticEnv()

	// Bind environment variables
	viper.BindEnv("GHMV_SOURCE_ORGANIZATION")
	viper.BindEnv("GHMV_SOURCE_TOKEN")
	viper.BindEnv("GHMV_SOURCE_HOSTNAME")
	viper.BindEnv("GHMV_SOURCE_ORGANIZATION")
	viper.BindEnv("GHMV_SOURCE_TOKEN")
	viper.BindEnv("GHMV_SOURCE_HOSTNAME")
	viper.BindEnv("GHMV_CSV_FILE")
	viper.BindEnv("HTTP_PROXY")
	viper.BindEnv("HTTPS_PROXY")
	viper.BindEnv("NO_PROXY")
	viper.BindEnv("RETRY_MAX")
	viper.BindEnv("RETRY_DELAY")
}
