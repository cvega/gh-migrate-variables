package cmd

import (
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
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

    // Add root command flags (not persistent)
    rootCmd.PersistentFlags().String("http-proxy", "", "HTTP proxy (can also use HTTP_PROXY env var)")
    rootCmd.PersistentFlags().String("https-proxy", "", "HTTPS proxy (can also use HTTPS_PROXY env var)")
    rootCmd.PersistentFlags().String("no-proxy", "", "No proxy list (can also use NO_PROXY env var)")

    // Bind flags to viper
    viper.BindPFlag("HTTP_PROXY", rootCmd.PersistentFlags().Lookup("http-proxy"))
    viper.BindPFlag("HTTPS_PROXY", rootCmd.PersistentFlags().Lookup("https-proxy"))
    viper.BindPFlag("NO_PROXY", rootCmd.PersistentFlags().Lookup("no-proxy"))
    
    // Add subcommands
    rootCmd.AddCommand(ExportCmd)
    rootCmd.AddCommand(SyncCmd)

    // hide -h, --help from global flags (shows up in proxy flags)
    rootCmd.Flags().BoolP("help", "h", false, "")
    rootCmd.Flags().Lookup("help").Hidden = true
}

func initConfig() {
    viper.SetEnvPrefix("GHMV")
    viper.BindEnv("HTTP_PROXY")
    viper.BindEnv("HTTPS_PROXY")
    viper.BindEnv("NO_PROXY")
    viper.AutomaticEnv()
}