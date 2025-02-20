/*
Copyright © 2024 Hinterland Software
Licensed under MIT License
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/hinterland-software/openv/internal/logging"
	"github.com/hinterland-software/openv/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var opServiceAuthToken string
var (
	verbose bool
	quiet   bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "openv",
	Short: "OpenV is a CLI tool to manage environment variables in 1Password",
	Long: `OpenV is a CLI tool to manage environment variables in 1Password.
It allows you to import environment variables from a .env file into 1Password.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is $HOME/.%s.yaml)", rootCmd.Use))
	rootCmd.PersistentFlags().StringVarP(&opServiceAuthToken, "op-token", "", "", "The 1Password service account token for authentication")
	cobra.CheckErr(viper.BindPFlag("op-token", rootCmd.PersistentFlags().Lookup("op-token")))
	viper.SetDefault("version", version.Info())

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", true, "Suppress all output except errors")

	// Initialize logger based on verbosity flags
	cobra.OnInitialize(initLogger)
}

func initLogger() {
	var level logging.LogLevel
	switch {
	case quiet:
		level = logging.Error
	case verbose:
		level = logging.Debug
	default:
		level = logging.Info
	}
	logging.InitLogger(level)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".openv" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(fmt.Sprintf(".%s", rootCmd.Use))
		viper.SetEnvPrefix(rootCmd.Use)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "📁 Using config file:", viper.ConfigFileUsed())
	}

}

// SetVersion sets the version of the application
func SetVersion(v string) {
	viper.Set("version", v)
}
